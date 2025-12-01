package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

func main() {
	const N = 5e8
	arr := make([]*byte, 0, 1024)

	start := time.Now()
	for i := 0; i < int(N); i++ {
		b := new(byte)
		arr = append(arr, b)
		if len(arr) > 1024 {
			arr = arr[1:]
		}
		if i%(N/10) == 0 {
			fmt.Printf("Allocated: %d\n", i)
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

	// Сохраняем текущий дамп куча:
	f, err := os.Create("heap.prof")
	if err != nil {
		fmt.Println("Ошибка создания heap.prof:", err)
		return
	}
	defer f.Close()
	if err := pprof.WriteHeapProfile(f); err != nil {
		fmt.Println("Ошибка записи дампа:", err)
		return
	}
	fmt.Println("Heap profile saved to heap.prof")
}
