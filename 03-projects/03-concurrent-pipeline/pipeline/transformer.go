package pipeline

import (
	"context"
	"fmt"
	"sync"

	"github.com/jordi-nyxidiom/go-mastery/03-projects/03-concurrent-pipeline/model"
)

// taxRates define las tasas de impuesto por region.
// EU: 21%, US: 8%, LATAM: 16%, ASIA: 10%.
var taxRates = map[string]float64{
	"EU":    0.21,
	"US":    0.08,
	"LATAM": 0.16,
	"ASIA":  0.10,
}

// Transform arranca N workers que procesan registros de venta concurrentemente (fan-out).
// Cada worker lee del canal de entrada, calcula totales e impuestos, y envia el resultado
// al canal de salida. Usa sync.WaitGroup para coordinar los workers y cierra el canal
// de salida cuando todos terminan.
func Transform(ctx context.Context, in <-chan model.SaleRecord, out chan<- model.ProcessedRecord, workers int, errCh chan<- error) {
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for record := range in {
				select {
				case <-ctx.Done():
					sendError(errCh, fmt.Errorf("transformer worker %d: cancelado: %w", workerID, ctx.Err()))
					return
				default:
				}

				processed := processRecord(record)

				select {
				case <-ctx.Done():
					sendError(errCh, fmt.Errorf("transformer worker %d: cancelado: %w", workerID, ctx.Err()))
					return
				case out <- processed:
				}
			}
		}(i)
	}

	// Esperar a que todos los workers terminen y cerrar el canal de salida
	go func() {
		wg.Wait()
		close(out)
	}()
}

// processRecord aplica los calculos de impuestos a un registro de venta.
// Si la region no tiene una tasa definida, se aplica 0%.
func processRecord(record model.SaleRecord) model.ProcessedRecord {
	total := float64(record.Quantity) * record.Price

	rate, ok := taxRates[record.Region]
	if !ok {
		rate = 0.0
	}

	taxAmount := total * rate
	netTotal := total + taxAmount

	return model.ProcessedRecord{
		SaleRecord: record,
		Total:      total,
		TaxRate:    rate,
		TaxAmount:  taxAmount,
		NetTotal:   netTotal,
	}
}
