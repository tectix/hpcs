package main

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	fmt.Println("Simple load test - 10 connections, 10 seconds")
	
	var ops int64
	var wg sync.WaitGroup
	
	start := time.Now()
	
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			conn, err := net.Dial("tcp4", "localhost:6379")
			if err != nil {
				fmt.Printf("Worker %d failed to connect: %v\n", id, err)
				return
			}
			defer conn.Close()
			
			end := time.Now().Add(10 * time.Second)
			counter := 0
			
			for time.Now().Before(end) {
				counter++
				
				cmd := "*1\r\n$4\r\nPING\r\n"
				
				_, err := conn.Write([]byte(cmd))
				if err != nil {
					continue
				}
				
				buf := make([]byte, 64)
				_, err = conn.Read(buf)
				if err != nil {
					continue
				}
				
				atomic.AddInt64(&ops, 1)
			}
			
			fmt.Printf("Worker %d completed %d ops\n", id, counter)
		}(i)
	}
	
	wg.Wait()
	
	elapsed := time.Since(start)
	totalOps := atomic.LoadInt64(&ops)
	opsPerSec := float64(totalOps) / elapsed.Seconds()
	
	fmt.Printf("\nResults:\n")
	fmt.Printf("Total operations: %d\n", totalOps)
	fmt.Printf("Duration: %v\n", elapsed)
	fmt.Printf("Ops/sec: %.2f\n", opsPerSec)
}