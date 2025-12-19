package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"sync"
	"sync/atomic"
	"time"
)

type Node struct {
	children []*Node
	payload  [512]byte
}

/* =======================
   LONG-LIVING OBJECTS
   ======================= */

var (
	globalRoots []*Node
	rootsMu     sync.Mutex
)

func buildLongLivingGraph(nodes, fanout int) {
	local := make([]*Node, nodes)
	for i := 0; i < nodes; i++ {
		local[i] = &Node{}
	}
	for i := 0; i < nodes; i++ {
		for j := 0; j < fanout; j++ {
			idx := (i + j*17) % nodes
			local[i].children = append(local[i].children, local[idx])
		}
	}
	rootsMu.Lock()
	globalRoots = append(globalRoots, local...)
	rootsMu.Unlock()
}

/* =======================
   METRICS
   ======================= */

type Metrics struct {
	mu   sync.Mutex
	file *os.File
}

func NewMetrics(path string) *Metrics {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(f, "ts_ms,heap_mb,total_mb,num_gc,pause_ms")
	return &Metrics{file: f}
}

func (m *Metrics) Collect() {
	var s runtime.MemStats
	runtime.ReadMemStats(&s)

	m.mu.Lock()
	defer m.mu.Unlock()

	fmt.Fprintf(
		m.file,
		"%d,%.2f,%.2f,%d,%.2f\n",
		time.Now().UnixMilli(),
		float64(s.HeapAlloc)/1024/1024,
		float64(s.TotalAlloc)/1024/1024,
		s.NumGC,
		float64(s.PauseTotalNs)/1e6,
	)
}

func (m *Metrics) Close() {
	m.file.Close()
}

/* =======================
   WORKER
   ======================= */

func worker(wg *sync.WaitGroup, iters int, counter *int64) {
	defer wg.Done()

	// mid-lived objects
	tmp := make([]*Node, 0, 1024)

	for i := 0; i < iters; i++ {
		n := &Node{} // short-lived

		if i%1000 == 0 {
			tmp = append(tmp, n)
		}
		if i%50000 == 0 && len(tmp) > 0 {
			tmp = tmp[:len(tmp)/2]
		}

		if i%10000 == 0 {
			atomic.AddInt64(counter, 10000)
		}
	}
}

/* =======================
   MAIN
   ======================= */

func main() {
	// ===== MEMORY LIMIT =====
	debug.SetMemoryLimit(1 << 30) // 1 GiB heap cap

	const (
		GOROUTINES = 40
		ITERS      = 5_000_000
	)

	os.MkdirAll("logs", 0755)
	os.MkdirAll("profiles", 0755)

	runID := os.Getenv("RUN_ID")
	if runID == "" {
		runID = "unknown"
	}

	metrics := NewMetrics("logs/run_" + runID + ".csv")
	defer metrics.Close()

	// ===== LONG-LIVING PRESSURE =====
	buildLongLivingGraph(1_500_000, 4)

	// build more during execution
	go func() {
		time.Sleep(3 * time.Second)
		buildLongLivingGraph(1_500_000, 4)
	}()

	// ===== METRICS TICKER =====
	stop := make(chan struct{})
	go func() {
		t := time.NewTicker(50 * time.Millisecond)
		for {
			select {
			case <-t.C:
				metrics.Collect()
			case <-stop:
				t.Stop()
				return
			}
		}
	}()

	start := time.Now()

	var wg sync.WaitGroup
	var counter int64

	for i := 0; i < GOROUTINES; i++ {
		wg.Add(1)
		go worker(&wg, ITERS, &counter)
	}

	wg.Wait()
	close(stop)

	runtime.GC()
	metrics.Collect()

	elapsed := time.Since(start)

	var s runtime.MemStats
	runtime.ReadMemStats(&s)

	fmt.Println("=== FINAL ===")
	fmt.Printf("Elapsed = %.2fs\n", elapsed.Seconds())
	fmt.Printf("HeapAlloc = %.1f MiB\n", float64(s.HeapAlloc)/1024/1024)
	fmt.Printf("TotalAlloc = %.1f MiB\n", float64(s.TotalAlloc)/1024/1024)
	fmt.Printf("NumGC = %d\n", s.NumGC)
	fmt.Printf("GC Pause total = %.2f ms\n", float64(s.PauseTotalNs)/1e6)

	// ===== HEAP PROFILE =====
	f, _ := os.Create("profiles/heap_" + runID + ".prof")
	pprof.WriteHeapProfile(f)
	f.Close()
}
