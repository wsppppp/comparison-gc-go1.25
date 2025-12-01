package main

import (
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"
)

type Node struct {
	children []*Node
	payload  [512]byte
}

func worker(wg *sync.WaitGroup, heavy, cross int, keeper *[]*Node) {
	defer wg.Done()

	// 1. Много короткоживущих объектов
	for i := 0; i < heavy; i++ {
		_ = &Node{}
	}

	// 2. Граф с циклическими/перекрестными ссылками
	nodes := make([]*Node, cross)
	for i := 0; i < cross; i++ {
		nodes[i] = &Node{}
	}
	for i := 0; i < cross; i++ {
		// Добавляем случайные перекрестные ссылки и циклы
		nodes[i].children = append(nodes[i].children, nodes[rand.Intn(cross)])
	}

	// 3. Долго живущие объекты (запоминаем часть из них)
	for i := 0; i < cross/100; i++ {
		idx := rand.Intn(cross)
		*keeper = append(*keeper, nodes[idx])
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	const GOROUTINES = 40
	const SHORT_LIVED = 2_000_000
	const GRAPH_SIZE = 120_000

	var wg sync.WaitGroup
	keeper := make([]*Node, 0, GOROUTINES*GRAPH_SIZE/100)

	start := time.Now()
	for i := 0; i < GOROUTINES; i++ {
		wg.Add(1)
		go worker(&wg, SHORT_LIVED, GRAPH_SIZE, &keeper)
	}
	wg.Wait()
	runtime.GC()

	stats := &runtime.MemStats{}
	runtime.ReadMemStats(stats)
	fmt.Printf("HeapAlloc = %v MiB\n", stats.HeapAlloc/1024/1024)
	fmt.Printf("TotalAlloc = %v MiB\n", stats.TotalAlloc/1024/1024)
	fmt.Printf("NumGC = %v\n", stats.NumGC)
	fmt.Printf("GC Pause total = %v ms\n", stats.PauseTotalNs/1_000_000)
	fmt.Printf("Elapsed = %v\n", time.Since(start).Seconds())
	fmt.Printf("Long-living keeper: %d nodes\n", len(keeper))

	// Итоговый дамп кучи
	f, err := os.Create("heap.prof")
	if err == nil {
		pprof.WriteHeapProfile(f)
		f.Close()
		fmt.Println("Heap profile saved to heap.prof")
	}
}
