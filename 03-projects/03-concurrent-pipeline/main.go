// Punto de entrada CLI para el pipeline concurrente de procesamiento de datos de ventas.
//
// Uso:
//
//	go run main.go -input testdata/sales.csv
//	go run main.go -input testdata/sales.csv -output resultado.csv -workers 8
//
// Flags:
//
//	-input    Ruta al archivo CSV de entrada (obligatorio)
//	-output   Ruta al archivo CSV de salida (opcional, si se omite solo muestra resumen)
//	-workers  Numero de workers concurrentes (por defecto: numero de CPUs)
//	-buffer   Tamano del buffer de los canales (por defecto: 100)
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/jordi-nyxidiom/go-mastery/03-projects/03-concurrent-pipeline/pipeline"
)

func main() {
	// Definir flags de la CLI
	inputPath := flag.String("input", "", "Ruta al archivo CSV de entrada (obligatorio)")
	outputPath := flag.String("output", "", "Ruta al archivo CSV de salida (opcional)")
	workers := flag.Int("workers", runtime.NumCPU(), "Numero de workers concurrentes")
	bufferSize := flag.Int("buffer", 100, "Tamano del buffer de los canales")
	flag.Parse()

	if *inputPath == "" {
		fmt.Fprintln(os.Stderr, "Error: se requiere el flag -input con la ruta al archivo CSV")
		fmt.Fprintln(os.Stderr, "Uso: go run main.go -input testdata/sales.csv [-output salida.csv] [-workers N] [-buffer N]")
		os.Exit(1)
	}

	// Configurar contexto con cancelacion por SIGINT/SIGTERM para shutdown graceful
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		fmt.Fprintf(os.Stderr, "\nSenal recibida: %v. Cerrando pipeline...\n", sig)
		cancel()
	}()

	// Configurar y ejecutar el pipeline
	cfg := pipeline.Config{
		InputPath:  *inputPath,
		OutputPath: *outputPath,
		Workers:    *workers,
		BufferSize: *bufferSize,
	}

	fmt.Printf("Iniciando pipeline con %d workers...\n", cfg.Workers)
	fmt.Printf("Archivo de entrada: %s\n", cfg.InputPath)
	if cfg.OutputPath != "" {
		fmt.Printf("Archivo de salida: %s\n", cfg.OutputPath)
	}
	fmt.Println()

	summary, err := pipeline.Run(ctx, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error en pipeline: %v\n", err)
		// Si hay resumen parcial, mostrarlo de todas formas
		if summary != nil && summary.TotalRecords > 0 {
			fmt.Println("\nResumen parcial:")
			pipeline.PrintSummary(summary)
		}
		os.Exit(1)
	}

	pipeline.PrintSummary(summary)

	if cfg.OutputPath != "" {
		fmt.Printf("\nResultados escritos en: %s\n", cfg.OutputPath)
	}
}
