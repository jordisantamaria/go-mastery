package main

import (
	"fmt"
	"sync"
	"time"
)

// =============================================
// SAFE COUNTER con Mutex
// =============================================

type SafeCounter struct {
	mu    sync.Mutex
	count int
}

func (c *SafeCounter) Increment() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count++
}

func (c *SafeCounter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}

// =============================================
// SAFE CACHE con RWMutex
// =============================================

type SafeCache struct {
	mu   sync.RWMutex
	data map[string]int
}

func NewSafeCache() *SafeCache {
	return &SafeCache{data: make(map[string]int)}
}

func (c *SafeCache) Get(key string) (int, bool) {
	c.mu.RLock() // multiples readers
	defer c.mu.RUnlock()
	val, ok := c.data[key]
	return val, ok
}

func (c *SafeCache) Set(key string, value int) {
	c.mu.Lock() // exclusive writer
	defer c.mu.Unlock()
	c.data[key] = value
}

func main() {
	// =============================================
	// SELECT — multiplexar channels
	// =============================================

	fmt.Println("=== Select ===")
	ch1 := make(chan string)
	ch2 := make(chan string)

	go func() {
		time.Sleep(200 * time.Millisecond)
		ch1 <- "message from ch1"
	}()

	go func() {
		time.Sleep(100 * time.Millisecond)
		ch2 <- "message from ch2"
	}()

	// Recibir de quien llegue primero
	for i := 0; i < 2; i++ {
		select {
		case msg := <-ch1:
			fmt.Println(" ", msg)
		case msg := <-ch2:
			fmt.Println(" ", msg)
		}
	}

	// =============================================
	// SELECT con TIMEOUT
	// =============================================

	fmt.Println("\n=== Select + Timeout ===")
	slow := make(chan string)

	go func() {
		time.Sleep(2 * time.Second) // simular operacion lenta
		slow <- "result"
	}()

	select {
	case result := <-slow:
		fmt.Println("  Got:", result)
	case <-time.After(500 * time.Millisecond):
		fmt.Println("  Timeout! Operacion demasiado lenta")
	}

	// =============================================
	// SELECT con default (non-blocking)
	// =============================================

	fmt.Println("\n=== Select + Default ===")
	ch := make(chan int, 1)

	// Non-blocking send
	select {
	case ch <- 42:
		fmt.Println("  Sent 42")
	default:
		fmt.Println("  Channel full, skipping")
	}

	// Non-blocking receive
	select {
	case val := <-ch:
		fmt.Println("  Received:", val)
	default:
		fmt.Println("  No value ready")
	}

	// =============================================
	// MUTEX — SafeCounter
	// =============================================

	fmt.Println("\n=== Mutex Counter ===")
	counter := &SafeCounter{}
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Increment()
		}()
	}

	wg.Wait()
	fmt.Printf("  Counter value: %d (expected 1000)\n", counter.Value())

	// =============================================
	// RWMutex — SafeCache
	// =============================================

	fmt.Println("\n=== RWMutex Cache ===")
	cache := NewSafeCache()
	var wg2 sync.WaitGroup

	// 10 writers
	for i := 0; i < 10; i++ {
		wg2.Add(1)
		go func(id int) {
			defer wg2.Done()
			key := fmt.Sprintf("key-%d", id)
			cache.Set(key, id*10)
		}(i)
	}

	// 100 readers
	for i := 0; i < 100; i++ {
		wg2.Add(1)
		go func(id int) {
			defer wg2.Done()
			key := fmt.Sprintf("key-%d", id%10)
			cache.Get(key) // multiples readers concurrentes
		}(i)
	}

	wg2.Wait()
	fmt.Println("  Cache operations completed")

	// =============================================
	// sync.Once — inicializacion unica
	// =============================================

	fmt.Println("\n=== sync.Once ===")
	var once sync.Once
	var wg3 sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg3.Add(1)
		go func(id int) {
			defer wg3.Done()
			once.Do(func() {
				fmt.Printf("  Initialized by goroutine %d\n", id)
			})
			fmt.Printf("  Goroutine %d continuing\n", id)
		}(i)
	}

	wg3.Wait()
}
