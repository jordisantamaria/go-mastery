package pipeline

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jordi-nyxidiom/go-mastery/03-projects/03-concurrent-pipeline/model"
)

// WriteCSV escribe los registros procesados a un archivo CSV de salida.
// Lee del canal de entrada hasta que se cierra y escribe cada registro como una fila.
func WriteCSV(outputPath string, in <-chan model.ProcessedRecord) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("writer: no se pudo crear archivo %s: %w", outputPath, err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	// Escribir cabecera del CSV de salida
	header := []string{
		"id", "date", "category", "product", "quantity", "price",
		"region", "total", "tax_rate", "tax_amount", "net_total",
	}
	if err := w.Write(header); err != nil {
		return fmt.Errorf("writer: error escribiendo cabecera: %w", err)
	}

	for rec := range in {
		row := []string{
			rec.ID,
			rec.Date.Format("2006-01-02"),
			rec.Category,
			rec.Product,
			strconv.Itoa(rec.Quantity),
			strconv.FormatFloat(rec.Price, 'f', 2, 64),
			rec.Region,
			strconv.FormatFloat(rec.Total, 'f', 2, 64),
			strconv.FormatFloat(rec.TaxRate, 'f', 2, 64),
			strconv.FormatFloat(rec.TaxAmount, 'f', 2, 64),
			strconv.FormatFloat(rec.NetTotal, 'f', 2, 64),
		}
		if err := w.Write(row); err != nil {
			return fmt.Errorf("writer: error escribiendo fila %s: %w", rec.ID, err)
		}
	}

	return nil
}

// PrintSummary imprime el resumen del pipeline en stdout con formato legible.
func PrintSummary(summary *model.Summary) {
	separator := strings.Repeat("=", 60)
	thinSep := strings.Repeat("-", 60)

	fmt.Println(separator)
	fmt.Println("          RESUMEN DEL PIPELINE DE VENTAS")
	fmt.Println(separator)
	fmt.Printf("  Registros procesados:   %d\n", summary.TotalRecords)
	fmt.Printf("  Ingresos totales:       $%.2f\n", summary.TotalRevenue)
	fmt.Printf("  Workers utilizados:     %d\n", summary.WorkersUsed)
	fmt.Printf("  Tiempo de proceso:      %s\n", summary.ProcessingTime)
	fmt.Println(thinSep)

	// Ingresos por categoria
	fmt.Println("  INGRESOS POR CATEGORIA:")
	for cat, revenue := range summary.ByCategory {
		fmt.Printf("    %-20s $%.2f\n", cat, revenue)
	}
	fmt.Println(thinSep)

	// Ingresos por region
	fmt.Println("  INGRESOS POR REGION:")
	for region, revenue := range summary.ByRegion {
		fmt.Printf("    %-20s $%.2f\n", region, revenue)
	}
	fmt.Println(thinSep)

	// Top productos
	fmt.Println("  TOP PRODUCTOS POR INGRESOS:")
	for i, p := range summary.TopProducts {
		fmt.Printf("    %2d. %-20s  Qty: %-6d  $%.2f\n", i+1, p.Product, p.Quantity, p.Revenue)
	}
	fmt.Println(separator)
}
