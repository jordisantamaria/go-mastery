package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func main() {
	// =============================================
	// PIPELINE pattern
	// =============================================

	fmt.Println("=== Pipeline ===")

	// generate -> double -> filter (> 10) -> print
	nums := generate(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
	doubled := mapChan(nums, func(n int) int { return n * 2 })
	big := filterChan(doubled, func(n int) bool { return n > 10 })

	for v := range big {
		fmt.Printf("  %d\n", v) // 12, 14, 16, 18, 20
	}

	// =============================================
	// WORKER POOL pattern
	// =============================================

	fmt.Println("\n=== Worker Pool ===")
	const numWorkers = 3
	const numJobs = 10

	jobs := make(chan int, numJobs)
	results := make(chan string, numJobs)

	// Lanzar workers
	var wg sync.WaitGroup
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for job := range jobs {
				time.Sleep(50 * time.Millisecond) // simular trabajo
				results <- fmt.Sprintf("Worker %d processed job %d", id, job)
			}
		}(w)
	}

	// Enviar jobs
	for j := 1; j <= numJobs; j++ {
		jobs <- j
	}
	close(jobs)

	// Cerrar results cuando todos los workers terminen
	go func() {
		wg.Wait()
		close(results)
	}()

	// Recoger resultados
	for r := range results {
		fmt.Println(" ", r)
	}

	// =============================================
	// FAN-OUT / FAN-IN
	// =============================================

	fmt.Println("\n=== Fan-out / Fan-in ===")
	input := generate(1, 2, 3, 4, 5, 6, 7, 8)

	// Fan-out: distribuir a 3 workers
	w1 := squareWorker(input)
	w2 := squareWorker(input)
	w3 := squareWorker(input)

	// Fan-in: combinar resultados
	merged := fanIn(w1, w2, w3)

	for v := range merged {
		fmt.Printf("  squared: %d\n", v)
	}

	// =============================================
	// CONTEXT con timeout
	// =============================================

	fmt.Println("\n=== Context Timeout ===")

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	result := make(chan string, 1)
	go func() {
		// Simular trabajo que puede tardar
		time.Sleep(200 * time.Millisecond)
		result <- "operation completed"
	}()

	select {
	case r := <-result:
		fmt.Println(" ", r)
	case <-ctx.Done():
		fmt.Println("  Cancelled:", ctx.Err())
	}

	// =============================================
	// CONTEXT con cancelacion manual
	// =============================================

	fmt.Println("\n=== Context Cancellation ===")
	ctx2, cancel2 := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx2.Done():
				fmt.Println("  Background worker stopped:", ctx2.Err())
				return
			default:
				fmt.Println("  Background worker running...")
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	time.Sleep(350 * time.Millisecond)
	cancel2() // cancelar la goroutine
	time.Sleep(50 * time.Millisecond)

	// =============================================
	// RATE LIMITING con ticker
	// =============================================

	fmt.Println("\n=== Rate Limiting ===")
	requests := make(chan int, 5)
	for i := 1; i <= 5; i++ {
		requests <- i
	}
	close(requests)

	// Limitar a 1 request cada 100ms
	limiter := time.NewTicker(100 * time.Millisecond)
	defer limiter.Stop()

	for req := range requests {
		<-limiter.C // esperar el tick
		fmt.Printf("  Request %d processed at %s\n", req, time.Now().Format("15:04:05.000"))
	}
}

// --- Pipeline helpers ---

func generate(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		for _, n := range nums {
			out <- n
		}
		close(out)
	}()
	return out
}

func mapChan(in <-chan int, fn func(int) int) <-chan int {
	out := make(chan int)
	go func() {
		for n := range in {
			out <- fn(n)
		}
		close(out)
	}()
	return out
}

func filterChan(in <-chan int, pred func(int) bool) <-chan int {
	out := make(chan int)
	go func() {
		for n := range in {
			if pred(n) {
				out <- n
			}
		}
		close(out)
	}()
	return out
}

// --- Fan-out/Fan-in helpers ---

func squareWorker(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		for n := range in {
			out <- n * n
		}
		close(out)
	}()
	return out
}

func fanIn(channels ...<-chan int) <-chan int {
	var wg sync.WaitGroup
	merged := make(chan int)

	for _, ch := range channels {
		wg.Add(1)
		go func(c <-chan int) {
			defer wg.Done()
			for v := range c {
				merged <- v
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(merged)
	}()

	return merged
}
