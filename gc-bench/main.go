package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

type Data struct {
	field [256]byte
}

func main() {
	const N = 2000000
	alive := make([]*Data, 0, N/2)
	start := time.Now()

	// 50% объектов — долго живут:
	for i := 0; i < N/2; i++ {
		alive = append(alive, &Data{})
	}
	// 50% — коротко живут
	for i := 0; i < N/2; i++ {
		_ = &Data{} // не складируем — сразу могут быть собраны
	}

	// Дополнительно чередуем аллокации
	for i := 0; i < 100000; i++ {
		tmp := make([]*Data, 100)
		for j := range tmp {
			tmp[j] = &Data{}
		}
	}

	runtime.GC()
	stats := &runtime.MemStats{}
	runtime.ReadMemStats(stats)
	fmt.Printf("HeapAlloc = %v MiB\n", stats.HeapAlloc/1024/1024)
	fmt.Printf("TotalAlloc = %v MiB\n", stats.TotalAlloc/1024/1024)
	fmt.Printf("NumGC = %v\n", stats.NumGC)
	fmt.Printf("GC Pause total = %v ms\n", stats.PauseTotalNs/1_000_000)
	fmt.Printf("Elapsed = %v\n", time.Since(start).Seconds())

	// Чтобы alive не оптимизировался
	var sum byte
	for i := range alive {
		sum += alive[i].field[0]
	}
	fmt.Println("Check sum:", sum)

	// Дамп памяти:
	f, err := os.Create("heap.prof")
	if err == nil {
		pprof.WriteHeapProfile(f)
		f.Close()
		fmt.Println("Heap profile saved to heap.prof")
	}
}
