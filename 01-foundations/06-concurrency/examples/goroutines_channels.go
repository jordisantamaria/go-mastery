package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	// =============================================
	// GOROUTINES basicas + WaitGroup
	// =============================================

	fmt.Println("=== Goroutines + WaitGroup ===")
	var wg sync.WaitGroup

	for i := 1; i <= 5; i++ {
		wg.Add(1) // ANTES de lanzar la goroutine
		go func(id int) {
			defer wg.Done()
			time.Sleep(time.Duration(id*100) * time.Millisecond)
			fmt.Printf("  Worker %d done\n", id)
		}(i)
	}

	wg.Wait() // espera a que todas terminen
	fmt.Println("  All workers done!")

	// =============================================
	// UNBUFFERED CHANNEL — sincronizacion
	// =============================================

	fmt.Println("\n=== Unbuffered Channel ===")
	ch := make(chan string)

	go func() {
		time.Sleep(500 * time.Millisecond)
		ch <- "hello from goroutine" // bloquea hasta que alguien reciba
	}()

	msg := <-ch // bloquea hasta que alguien envie
	fmt.Println(" ", msg)

	// =============================================
	// BUFFERED CHANNEL — producer/consumer
	// =============================================

	fmt.Println("\n=== Buffered Channel ===")
	tasks := make(chan int, 3)

	// Producer
	go func() {
		for i := 1; i <= 5; i++ {
			fmt.Printf("  Producing task %d\n", i)
			tasks <- i
		}
		close(tasks)
	}()

	// Consumer — range itera hasta que el channel se cierra
	for task := range tasks {
		fmt.Printf("  Consumed task %d\n", task)
		time.Sleep(200 * time.Millisecond)
	}

	// =============================================
	// CHANNEL DIRECTIONS
	// =============================================

	fmt.Println("\n=== Channel Directions ===")
	numbers := make(chan int)
	doubled := make(chan int)

	go produce(numbers)   // solo envia
	go transform(numbers, doubled) // recibe y envia
	consume(doubled)      // solo recibe

	// =============================================
	// done SIGNAL con chan struct{}
	// =============================================

	fmt.Println("\n=== Done Signal ===")
	done := make(chan struct{})

	go func() {
		fmt.Println("  Background task running...")
		time.Sleep(300 * time.Millisecond)
		fmt.Println("  Background task finished!")
		close(done) // senal de "terminado"
	}()

	<-done // espera la senal
	fmt.Println("  Main received done signal")
}

func produce(out chan<- int) {
	for i := 1; i <= 3; i++ {
		out <- i
	}
	close(out)
}

func transform(in <-chan int, out chan<- int) {
	for n := range in {
		out <- n * 2
	}
	close(out)
}

func consume(in <-chan int) {
	for n := range in {
		fmt.Printf("  Got %d\n", n)
	}
}
