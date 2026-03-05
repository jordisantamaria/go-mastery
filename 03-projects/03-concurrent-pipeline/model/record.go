// Package model define los tipos de datos utilizados en el pipeline de procesamiento.
// Incluye registros de venta, registros procesados con impuestos, y resumenes agregados.
package model

import "time"

// SaleRecord representa una fila del CSV de ventas sin procesar.
type SaleRecord struct {
	ID       string
	Date     time.Time
	Category string
	Product  string
	Quantity int
	Price    float64
	Region   string
}

// ProcessedRecord es un registro de venta con los calculos de impuestos aplicados.
type ProcessedRecord struct {
	SaleRecord
	Total     float64 // Quantity * Price
	TaxRate   float64 // Tasa de impuesto segun la region
	TaxAmount float64 // Total * TaxRate
	NetTotal  float64 // Total + TaxAmount
}

// Summary contiene el resumen agregado de todos los registros procesados.
type Summary struct {
	TotalRecords   int
	TotalRevenue   float64
	ByCategory     map[string]float64
	ByRegion       map[string]float64
	TopProducts    []ProductSummary
	ProcessingTime time.Duration
	WorkersUsed    int
}

// ProductSummary agrupa la cantidad vendida e ingresos por producto.
type ProductSummary struct {
	Product  string
	Quantity int
	Revenue  float64
}
