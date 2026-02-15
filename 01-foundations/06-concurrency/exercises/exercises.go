// Package exercises contiene funciones para practicar concurrencia en Go.
// Ejecuta los tests con: go test -race ./01-foundations/06-concurrency/exercises/...
// IMPORTANTE: usa -race para detectar race conditions!
package exercises

import (
	"context"
	"sync"
)

// --- Ejercicio 1: ConcurrentSum ---
// Suma todos los numeros del slice usando N goroutines.
// Divide el slice en N partes iguales (o casi iguales) y suma cada parte en una goroutine.
// Usa channels para recoger los resultados parciales.
// ConcurrentSum([]int{1,2,3,4,5,6,7,8,9,10}, 3) -> 55
func ConcurrentSum(nums []int, numWorkers int) int {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 2: SafeMap ---
// Implementa un map thread-safe usando sync.RWMutex.
// Debe soportar Set, Get, Delete, y Len de forma concurrente.

type SafeMap struct {
	// TODO: definir campos (mu sync.RWMutex, data map[string]int)
}

func NewSafeMap() *SafeMap {
	// TODO: implementar
	panic("not implemented")
}

func (m *SafeMap) Set(key string, value int) {
	// TODO: implementar con Lock
	panic("not implemented")
}

func (m *SafeMap) Get(key string) (int, bool) {
	// TODO: implementar con RLock
	panic("not implemented")
}

func (m *SafeMap) Delete(key string) {
	// TODO: implementar con Lock
	panic("not implemented")
}

func (m *SafeMap) Len() int {
	// TODO: implementar con RLock
	panic("not implemented")
}

// --- Ejercicio 3: Pipeline ---
// Implementa una pipeline de 3 stages usando channels:
// 1. Generate: envia los numeros del slice al channel de salida
// 2. Square: lee de input, eleva al cuadrado, y envia al output
// 3. FilterEven: lee de input, solo deja pasar los pares
// RunPipeline debe componer las 3 stages y devolver los resultados como slice.
// RunPipeline([]int{1,2,3,4,5}) -> [4, 16] (1->1 impar, 2->4 par, 3->9 impar, 4->16 par, 5->25 impar)

func Generate(nums []int) <-chan int {
	// TODO: implementar — enviar nums al channel y cerrarlo
	panic("not implemented")
}

func Square(in <-chan int) <-chan int {
	// TODO: implementar — leer, elevar al cuadrado, enviar
	panic("not implemented")
}

func FilterEven(in <-chan int) <-chan int {
	// TODO: implementar — solo dejar pasar numeros pares
	panic("not implemented")
}

func RunPipeline(nums []int) []int {
	// TODO: componer Generate -> Square -> FilterEven -> recoger resultados
	panic("not implemented")
}

// --- Ejercicio 4: WorkerPool ---
// Implementa un worker pool que procesa jobs en paralelo.
// - numWorkers goroutines leen del channel de jobs
// - Cada worker aplica la funcion process al job y envia el resultado
// - Los resultados se recogen en un slice (el orden no importa)

func WorkerPool(jobs []int, numWorkers int, process func(int) int) []int {
	// TODO: implementar
	// Pista: usa WaitGroup para saber cuando cerrar el channel de resultados
	panic("not implemented")
}

// --- Ejercicio 5: FirstResult ---
// Lanza N funciones concurrentemente y devuelve el resultado de la PRIMERA que termine.
// Las demas se deben cancelar via context.
// Pista: usa un buffered channel de capacidad 1.

func FirstResult(ctx context.Context, fns ...func(context.Context) (string, error)) (string, error) {
	// TODO: implementar
	panic("not implemented")
}

// --- Ejercicio 6: Semaphore ---
// Implementa un semaphore usando un buffered channel.
// Limita la concurrencia a maxConcurrent goroutines ejecutandose a la vez.
// ProcessWithLimit debe procesar todos los items pero con un maximo de maxConcurrent en paralelo.

func ProcessWithLimit(items []int, maxConcurrent int, process func(int) int) []int {
	// TODO: implementar
	// Pista: usa un chan struct{} de capacidad maxConcurrent como semaforo
	panic("not implemented")
}

// --- Ejercicio 7: Merge ---
// Recibe multiples channels de solo lectura y los fusiona en un solo channel.
// El channel de salida se cierra cuando TODOS los channels de entrada se cierran.
// (Este es el patron fan-in)

func Merge(channels ...<-chan int) <-chan int {
	// TODO: implementar
	// Pista: usa sync.WaitGroup para saber cuando cerrar el channel de salida
	_ = sync.WaitGroup{}
	panic("not implemented")
}
