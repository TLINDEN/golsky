package main

import "runtime"

// returns current memory usage in MB
func GetMem() float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return float64(m.Alloc) / 1024 / 1024
}
