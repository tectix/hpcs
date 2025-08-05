package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type Stats struct {
	Operations int64
	Errors     int64
	StartTime  time.Time
}

func (s *Stats) RecordOp() {
	atomic.AddInt64(&s.Operations, 1)
}

func (s *Stats) RecordError() {
	atomic.AddInt64(&s.Errors, 1)
}

func (s *Stats) Report() {
	elapsed := time.Since(s.StartTime)
	ops := atomic.LoadInt64(&s.Operations)
	errors := atomic.LoadInt64(&s.Errors)
	
	opsPerSec := float64(ops) / elapsed.Seconds()
	
	fmt.Printf("Results:\n")
	fmt.Printf("  Operations: %d\n", ops)
	fmt.Printf("  Errors: %d\n", errors)
	fmt.Printf("  Duration: %v\n", elapsed)
	fmt.Printf("  Ops/sec: %.2f\n", opsPerSec)
	fmt.Printf("  Success rate: %.2f%%\n", float64(ops-errors)/float64(ops)*100)
}

func worker(addr string, stats *Stats, duration time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()
	
	conn, err := net.Dial("tcp4", addr)
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		return
	}
	defer conn.Close()
	
	endTime := time.Now().Add(duration)
	counter := 0
	
	for time.Now().Before(endTime) {
		counter++
		
		var cmd string
		if counter%2 == 0 {
			cmd = fmt.Sprintf("*3\r\n$3\r\nSET\r\n$7\r\nkey_%d\r\n$8\r\nvalue_%d\r\n", counter, counter)
		} else {
			cmd = fmt.Sprintf("*2\r\n$3\r\nGET\r\n$7\r\nkey_%d\r\n", counter-1)
		}
		
		_, err := conn.Write([]byte(cmd))
		if err != nil {
			stats.RecordError()
			continue
		}
		
		response := make([]byte, 1024)
		_, err = conn.Read(response)
		if err != nil {
			stats.RecordError()
			continue
		}
		
		stats.RecordOp()
	}
}

func main() {
	var (
		addr        = flag.String("addr", "localhost:6379", "Server address")
		connections = flag.Int("connections", 10, "Number of concurrent connections")
		duration    = flag.Duration("duration", 10*time.Second, "Test duration")
	)
	flag.Parse()
	
	fmt.Printf("Load testing %s with %d connections for %v\n", *addr, *connections, *duration)
	
	stats := &Stats{StartTime: time.Now()}
	var wg sync.WaitGroup
	
	for i := 0; i < *connections; i++ {
		wg.Add(1)
		go worker(*addr, stats, *duration, &wg)
	}
	
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				ops := atomic.LoadInt64(&stats.Operations)
				elapsed := time.Since(stats.StartTime)
				opsPerSec := float64(ops) / elapsed.Seconds()
				fmt.Printf("Progress: %d ops, %.2f ops/sec\n", ops, opsPerSec)
			}
		}
	}()
	
	wg.Wait()
	
	stats.Report()
	
	elapsed := time.Since(stats.StartTime)
	ops := atomic.LoadInt64(&stats.Operations)
	opsPerSec := float64(ops) / elapsed.Seconds()
	
	if opsPerSec >= 100000 {
		fmt.Printf("SUCCESS: Achieved target of 100K+ ops/sec!\n")
		os.Exit(0)
	} else {
		fmt.Printf("Target not reached. Need %.2f more ops/sec\n", 100000-opsPerSec)
		os.Exit(1)
	}
}