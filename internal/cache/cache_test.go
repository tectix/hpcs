package cache

import (
	"strconv"
	"testing"
	"time"
)

func BenchmarkCacheGet(b *testing.B) {
	cache := New(1024 * 1024 * 1024) // 1GB
	
	for i := 0; i < 10000; i++ {
		key := "key" + strconv.Itoa(i)
		value := []byte("value" + strconv.Itoa(i))
		cache.Set(key, value, 0)
	}
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := "key" + strconv.Itoa(i%10000)
			cache.Get(key)
			i++
		}
	})
}

func BenchmarkCacheSet(b *testing.B) {
	cache := New(1024 * 1024 * 1024) // 1GB
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := "key" + strconv.Itoa(i)
			value := []byte("value" + strconv.Itoa(i))
			cache.Set(key, value, 0)
			i++
		}
	})
}

func BenchmarkCacheMixed(b *testing.B) {
	cache := New(1024 * 1024 * 1024) // 1GB
	
	for i := 0; i < 1000; i++ {
		key := "key" + strconv.Itoa(i)
		value := []byte("value" + strconv.Itoa(i))
		cache.Set(key, value, 0)
	}
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%5 == 0 { // 20% writes, 80% reads
				key := "key" + strconv.Itoa(i)
				value := []byte("value" + strconv.Itoa(i))
				cache.Set(key, value, 0)
			} else {
				key := "key" + strconv.Itoa(i%1000)
				cache.Get(key)
			}
			i++
		}
	})
}

func BenchmarkCacheSetWithTTL(b *testing.B) {
	cache := New(1024 * 1024 * 1024) // 1GB
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := "key" + strconv.Itoa(i)
			value := []byte("value" + strconv.Itoa(i))
			cache.Set(key, value, time.Hour)
			i++
		}
	})
}

func TestCacheBasicOperations(t *testing.T) {
	cache := New(1024)
	
	cache.Set("key1", []byte("value1"), 0)
	value, exists := cache.Get("key1")
	if !exists || string(value) != "value1" {
		t.Errorf("Expected value1, got %s", string(value))
	}
	
	deleted := cache.Delete("key1")
	if !deleted {
		t.Error("Expected key to be deleted")
	}
	
	_, exists = cache.Get("key1")
	if exists {
		t.Error("Key should not exist after deletion")
	}
}

func TestCacheTTL(t *testing.T) {
	cache := New(1024)
	
	cache.Set("expiring", []byte("value"), 100*time.Millisecond)
	
	_, exists := cache.Get("expiring")
	if !exists {
		t.Error("Key should exist immediately after setting")
	}
	
	time.Sleep(150 * time.Millisecond)
	
	_, exists = cache.Get("expiring")
	if exists {
		t.Error("Key should not exist after TTL expiration")
	}
}

func TestCacheConcurrency(t *testing.T) {
	cache := New(1024 * 1024)
	
	done := make(chan bool, 100)
	
	for i := 0; i < 100; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				key := "key" + strconv.Itoa(id) + "_" + strconv.Itoa(j)
				value := []byte("value" + strconv.Itoa(id) + "_" + strconv.Itoa(j))
				
				cache.Set(key, value, 0)
				
				retrieved, exists := cache.Get(key)
				if !exists || string(retrieved) != string(value) {
					t.Errorf("Concurrent operation failed for %s", key)
				}
			}
			done <- true
		}(i)
	}
	
	for i := 0; i < 100; i++ {
		<-done
	}
}