package pipeline

import (
	"sort"

	"github.com/jordi-nyxidiom/go-mastery/03-projects/03-concurrent-pipeline/model"
)

// Aggregate recibe registros procesados del canal de entrada (fan-in) y construye
// un resumen agregado. Acumula totales por categoria, por region, y calcula los
// productos mas vendidos por ingresos. Esta funcion es segura para un solo consumidor;
// la concurrencia se maneja en las etapas anteriores.
func Aggregate(in <-chan model.ProcessedRecord, records chan<- model.ProcessedRecord) *model.Summary {
	summary := &model.Summary{
		ByCategory: make(map[string]float64),
		ByRegion:   make(map[string]float64),
	}

	// Mapa temporal para agrupar productos
	productMap := make(map[string]*model.ProductSummary)

	for rec := range in {
		summary.TotalRecords++
		summary.TotalRevenue += rec.NetTotal
		summary.ByCategory[rec.Category] += rec.NetTotal
		summary.ByRegion[rec.Region] += rec.NetTotal

		// Acumular datos por producto
		ps, ok := productMap[rec.Product]
		if !ok {
			ps = &model.ProductSummary{Product: rec.Product}
			productMap[rec.Product] = ps
		}
		ps.Quantity += rec.Quantity
		ps.Revenue += rec.NetTotal

		// Reenviar el registro para que el writer lo pueda escribir
		if records != nil {
			records <- rec
		}
	}

	// Cerrar el canal de registros para que el writer sepa que no hay mas datos
	if records != nil {
		close(records)
	}

	// Convertir el mapa de productos a slice y ordenar por ingresos (descendente)
	products := make([]model.ProductSummary, 0, len(productMap))
	for _, ps := range productMap {
		products = append(products, *ps)
	}
	sort.Slice(products, func(i, j int) bool {
		return products[i].Revenue > products[j].Revenue
	})

	// Quedarnos con los top 10 productos (o menos si hay menos)
	if len(products) > 10 {
		products = products[:10]
	}
	summary.TopProducts = products

	return summary
}
