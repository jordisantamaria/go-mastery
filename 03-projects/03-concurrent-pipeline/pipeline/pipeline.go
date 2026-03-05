// Package pipeline implementa un pipeline concurrente de procesamiento de datos de ventas.
//
// El pipeline tiene 4 etapas conectadas por canales:
//
//	CSV File -> [Reader] -> chan SaleRecord -> [Transformer x N] -> chan ProcessedRecord -> [Aggregator] -> Summary
//	                                                                                          |
//	                                                                                    [Writer] -> CSV Output
//
// Usa errgroup para coordinar las goroutines y soporta cancelacion graceful via context.
package pipeline

import (
	"context"
	"fmt"
	"time"

	"github.com/jordi-nyxidiom/go-mastery/03-projects/03-concurrent-pipeline/model"
)

// Config contiene la configuracion del pipeline.
type Config struct {
	InputPath  string // Ruta al archivo CSV de entrada
	OutputPath string // Ruta al archivo CSV de salida (vacio = solo resumen en stdout)
	Workers    int    // Numero de workers para la etapa de transformacion
	BufferSize int    // Tamano del buffer de los canales
}

// Run ejecuta el pipeline completo: lee CSV, transforma con N workers, agrega resultados
// y opcionalmente escribe la salida a CSV. Devuelve un resumen con las metricas de procesamiento.
//
// El pipeline soporta cancelacion graceful a traves del contexto. Si el contexto se cancela,
// todas las etapas terminan de forma ordenada y se devuelve el error correspondiente.
func Run(ctx context.Context, cfg Config) (*model.Summary, error) {
	start := time.Now()

	if cfg.Workers <= 0 {
		cfg.Workers = 1
	}
	if cfg.BufferSize <= 0 {
		cfg.BufferSize = 100
	}

	// Canales que conectan las etapas del pipeline
	saleRecords := make(chan model.SaleRecord, cfg.BufferSize)
	processedRecords := make(chan model.ProcessedRecord, cfg.BufferSize)
	errCh := make(chan error, cfg.Workers+2) // Buffer suficiente para errores de todas las etapas

	// Canal opcional para reenviar registros al writer
	var writerCh chan model.ProcessedRecord
	if cfg.OutputPath != "" {
		writerCh = make(chan model.ProcessedRecord, cfg.BufferSize)
	}

	// Etapa 1: Leer CSV y enviar registros al canal
	go ReadCSV(ctx, cfg.InputPath, saleRecords, errCh)

	// Etapa 2: Transformar registros con N workers (fan-out)
	// Transform cierra processedRecords cuando todos los workers terminan
	Transform(ctx, saleRecords, processedRecords, cfg.Workers, errCh)

	// Etapa 3 y 4: Agregar resultados y opcionalmente escribir CSV
	// Se ejecutan en la misma goroutine porque el aggregator alimenta al writer
	var summary *model.Summary
	var writeErr error

	// Si hay archivo de salida, el writer consume del canal writerCh
	if cfg.OutputPath != "" {
		writerDone := make(chan struct{})
		go func() {
			defer close(writerDone)
			writeErr = WriteCSV(cfg.OutputPath, writerCh)
		}()

		// El aggregator consume processedRecords y reenvia a writerCh
		summary = Aggregate(processedRecords, writerCh)

		// Esperar a que el writer termine
		<-writerDone
	} else {
		// Sin archivo de salida: solo agregar
		summary = Aggregate(processedRecords, nil)
	}

	summary.WorkersUsed = cfg.Workers
	summary.ProcessingTime = time.Since(start)

	// Recoger errores no fatales
	close(errCh)
	var firstErr error
	for err := range errCh {
		if firstErr == nil && err != nil {
			// Distinguir entre errores de cancelacion y errores reales
			if ctx.Err() != nil {
				firstErr = fmt.Errorf("pipeline cancelado: %w", ctx.Err())
			} else {
				firstErr = err
			}
		}
	}

	if writeErr != nil {
		return summary, writeErr
	}

	// Si el contexto fue cancelado, devolver el error de cancelacion
	// pero tambien devolver el resumen parcial
	if ctx.Err() != nil {
		return summary, fmt.Errorf("pipeline cancelado: %w", ctx.Err())
	}

	// Si el resumen tiene 0 registros y hubo un error, devolver el error
	if summary.TotalRecords == 0 && firstErr != nil {
		return summary, firstErr
	}

	return summary, nil
}
