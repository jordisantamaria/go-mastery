package pipeline

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/jordi-nyxidiom/go-mastery/03-projects/03-concurrent-pipeline/model"
)

// ReadCSV lee un archivo CSV linea por linea y envia cada registro al canal de salida.
// Respeta la cancelacion del contexto y reporta errores por el canal de errores.
// Cierra el canal de salida cuando termina de leer o se cancela el contexto.
func ReadCSV(ctx context.Context, filePath string, out chan<- model.SaleRecord, errCh chan<- error) {
	defer close(out)

	f, err := os.Open(filePath)
	if err != nil {
		sendError(errCh, fmt.Errorf("reader: no se pudo abrir el archivo %s: %w", filePath, err))
		return
	}
	defer f.Close()

	r := csv.NewReader(f)
	// Permitir filas con numero variable de campos para poder manejar errores
	// de parsing nosotros mismos en lugar de que el csv.Reader falle
	r.FieldsPerRecord = -1

	// Leer y descartar la cabecera del CSV
	header, err := r.Read()
	if err != nil {
		sendError(errCh, fmt.Errorf("reader: error leyendo cabecera: %w", err))
		return
	}

	// Validar que la cabecera tenga las columnas esperadas
	if len(header) < 7 {
		sendError(errCh, fmt.Errorf("reader: cabecera invalida, se esperan 7 columnas, se encontraron %d", len(header)))
		return
	}

	lineNum := 1 // La cabecera es la linea 1
	for {
		select {
		case <-ctx.Done():
			sendError(errCh, fmt.Errorf("reader: cancelado: %w", ctx.Err()))
			return
		default:
		}

		row, err := r.Read()
		if err == io.EOF {
			return
		}
		if err != nil {
			sendError(errCh, fmt.Errorf("reader: error en linea %d: %w", lineNum+1, err))
			return
		}
		lineNum++

		record, err := parseRow(row, lineNum)
		if err != nil {
			sendError(errCh, fmt.Errorf("reader: linea %d: %w", lineNum, err))
			continue // Saltar filas mal formadas en lugar de detener el pipeline
		}

		select {
		case <-ctx.Done():
			sendError(errCh, fmt.Errorf("reader: cancelado: %w", ctx.Err()))
			return
		case out <- record:
		}
	}
}

// sendError envia un error al canal de forma no bloqueante.
// Si el canal esta lleno, descarta el error para evitar deadlocks.
func sendError(errCh chan<- error, err error) {
	select {
	case errCh <- err:
	default:
		// Canal lleno, descartar error para evitar bloqueo
	}
}

// parseRow convierte una fila CSV en un SaleRecord.
// Espera exactamente 7 campos: id, date, category, product, quantity, price, region.
func parseRow(row []string, lineNum int) (model.SaleRecord, error) {
	if len(row) < 7 {
		return model.SaleRecord{}, fmt.Errorf("se esperan 7 campos, se encontraron %d", len(row))
	}

	date, err := time.Parse("2006-01-02", row[1])
	if err != nil {
		return model.SaleRecord{}, fmt.Errorf("fecha invalida %q: %w", row[1], err)
	}

	quantity, err := strconv.Atoi(row[4])
	if err != nil {
		return model.SaleRecord{}, fmt.Errorf("cantidad invalida %q: %w", row[4], err)
	}

	price, err := strconv.ParseFloat(row[5], 64)
	if err != nil {
		return model.SaleRecord{}, fmt.Errorf("precio invalido %q: %w", row[5], err)
	}

	return model.SaleRecord{
		ID:       row[0],
		Date:     date,
		Category: row[2],
		Product:  row[3],
		Quantity: quantity,
		Price:    price,
		Region:   row[6],
	}, nil
}
