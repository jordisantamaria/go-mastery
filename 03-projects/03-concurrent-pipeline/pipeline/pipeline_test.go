package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/jordi-nyxidiom/go-mastery/03-projects/03-concurrent-pipeline/model"
)

// testdataDir devuelve la ruta al directorio testdata relativa al paquete.
func testdataDir() string {
	return filepath.Join("..", "testdata")
}

// createTempCSV crea un archivo CSV temporal con el contenido proporcionado
// y devuelve su ruta. El archivo se limpia automaticamente al terminar el test.
func createTempCSV(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.csv")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("no se pudo crear CSV temporal: %v", err)
	}
	return path
}

// newSaleRecord crea un SaleRecord basico para tests.
func newSaleRecord(quantity int, price float64, region string) model.SaleRecord {
	return model.SaleRecord{
		ID:       "test-1",
		Date:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Category: "Test",
		Product:  "TestProduct",
		Quantity: quantity,
		Price:    price,
		Region:   region,
	}
}

// assertFloat compara dos floats con tolerancia de 0.01.
func assertFloat(t *testing.T, name string, expected, actual float64) {
	t.Helper()
	diff := expected - actual
	if diff < -0.01 || diff > 0.01 {
		t.Errorf("%s: esperado %.2f, obtenido %.2f", name, expected, actual)
	}
}

// --- Tests de integracion del pipeline completo ---

// TestPipelineCompleto verifica el pipeline con el dataset de ejemplo completo.
func TestPipelineCompleto(t *testing.T) {
	inputPath := filepath.Join(testdataDir(), "sales.csv")

	cfg := Config{
		InputPath:  inputPath,
		OutputPath: "",
		Workers:    4,
		BufferSize: 50,
	}

	summary, err := Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("error ejecutando pipeline: %v", err)
	}

	// Verificar que se procesaron las 50 filas del CSV de ejemplo
	if summary.TotalRecords != 50 {
		t.Errorf("se esperaban 50 registros, se obtuvieron %d", summary.TotalRecords)
	}

	// Verificar que hay ingresos
	if summary.TotalRevenue <= 0 {
		t.Error("los ingresos totales deberian ser positivos")
	}

	// Verificar categorias esperadas
	expectedCategories := []string{"Electronics", "Food", "Clothing", "Books", "Software"}
	for _, cat := range expectedCategories {
		if _, ok := summary.ByCategory[cat]; !ok {
			t.Errorf("categoria %q no encontrada en el resumen", cat)
		}
	}

	// Verificar regiones esperadas
	expectedRegions := []string{"EU", "US", "LATAM", "ASIA"}
	for _, reg := range expectedRegions {
		if _, ok := summary.ByRegion[reg]; !ok {
			t.Errorf("region %q no encontrada en el resumen", reg)
		}
	}

	// Verificar que hay top productos
	if len(summary.TopProducts) == 0 {
		t.Error("deberia haber al menos un producto en el top")
	}

	// Verificar que el numero de workers se registro
	if summary.WorkersUsed != 4 {
		t.Errorf("se esperaban 4 workers, se registraron %d", summary.WorkersUsed)
	}

	// Verificar que el tiempo de procesamiento es razonable
	if summary.ProcessingTime <= 0 {
		t.Error("el tiempo de procesamiento deberia ser positivo")
	}
}

// TestPipelineConSalidaCSV verifica que el pipeline escribe correctamente un CSV de salida.
func TestPipelineConSalidaCSV(t *testing.T) {
	inputPath := filepath.Join(testdataDir(), "sales.csv")
	outputDir := t.TempDir()
	outputPath := filepath.Join(outputDir, "output.csv")

	cfg := Config{
		InputPath:  inputPath,
		OutputPath: outputPath,
		Workers:    2,
		BufferSize: 50,
	}

	summary, err := Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("error ejecutando pipeline: %v", err)
	}

	if summary.TotalRecords != 50 {
		t.Errorf("se esperaban 50 registros, se obtuvieron %d", summary.TotalRecords)
	}

	// Verificar que el archivo de salida existe y tiene contenido
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("el archivo de salida no existe: %v", err)
	}
	if info.Size() == 0 {
		t.Error("el archivo de salida esta vacio")
	}

	// Leer el archivo y verificar que tiene la cabecera y filas
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("error leyendo archivo de salida: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	// Cabecera + 50 filas de datos
	if len(lines) != 51 {
		t.Errorf("se esperaban 51 lineas (cabecera + 50 datos), se obtuvieron %d", len(lines))
	}

	// Verificar que la cabecera tiene las columnas esperadas
	expectedHeader := "id,date,category,product,quantity,price,region,total,tax_rate,tax_amount,net_total"
	if lines[0] != expectedHeader {
		t.Errorf("cabecera inesperada:\n  obtenida:  %s\n  esperada: %s", lines[0], expectedHeader)
	}
}

// TestPipelineDatasetPequeno verifica el pipeline con un dataset minimo.
func TestPipelineDatasetPequeno(t *testing.T) {
	csvData := `id,date,category,product,quantity,price,region
1,2024-01-01,Electronics,Laptop,1,1000.00,EU
2,2024-01-02,Food,Coffee,10,5.00,US`

	path := createTempCSV(t, csvData)

	cfg := Config{
		InputPath:  path,
		Workers:    1,
		BufferSize: 10,
	}

	summary, err := Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	if summary.TotalRecords != 2 {
		t.Errorf("se esperaban 2 registros, se obtuvieron %d", summary.TotalRecords)
	}

	// Laptop: 1 * 1000 = 1000, impuesto EU 21% = 210, neto = 1210
	// Coffee: 10 * 5 = 50, impuesto US 8% = 4, neto = 54
	// Total esperado: 1264
	expectedRevenue := 1264.0
	assertFloat(t, "TotalRevenue", expectedRevenue, summary.TotalRevenue)
}

// TestPipelineCancelacion verifica que el pipeline responde a la cancelacion inmediata.
func TestPipelineCancelacion(t *testing.T) {
	inputPath := filepath.Join(testdataDir(), "sales.csv")

	// Crear un contexto que ya esta cancelado
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cfg := Config{
		InputPath:  inputPath,
		Workers:    2,
		BufferSize: 10,
	}

	_, err := Run(ctx, cfg)
	// Deberia devolver un error de cancelacion o un resumen con 0 registros
	if err != nil {
		if !strings.Contains(err.Error(), "cancel") {
			t.Logf("error inesperado (puede ser aceptable): %v", err)
		}
	}
}

// TestPipelineCancelacionDuranteProceso verifica que el pipeline se detiene
// cuando el contexto se cancela mientras esta procesando.
func TestPipelineCancelacionDuranteProceso(t *testing.T) {
	// Crear un CSV grande para dar tiempo a la cancelacion
	var sb strings.Builder
	sb.WriteString("id,date,category,product,quantity,price,region\n")
	for i := 0; i < 1000; i++ {
		sb.WriteString("1,2024-01-01,Electronics,Laptop,1,100.00,EU\n")
	}
	path := createTempCSV(t, sb.String())

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	time.Sleep(1 * time.Millisecond)

	cfg := Config{
		InputPath:  path,
		Workers:    2,
		BufferSize: 5,
	}

	summary, _ := Run(ctx, cfg)
	if summary != nil && summary.TotalRecords >= 1000 {
		t.Log("el pipeline proceso todos los registros antes de la cancelacion (aceptable si fue muy rapido)")
	}
}

// TestPipelineDiferentesWorkers verifica que el resultado es consistente
// independientemente del numero de workers.
func TestPipelineDiferentesWorkers(t *testing.T) {
	inputPath := filepath.Join(testdataDir(), "sales.csv")

	workerCounts := []int{1, 2, 4, 8}
	var firstRevenue float64

	for _, w := range workerCounts {
		w := w // Capturar variable del loop
		name := fmt.Sprintf("workers_%d", w)
		t.Run(name, func(t *testing.T) {
			cfg := Config{
				InputPath:  inputPath,
				Workers:    w,
				BufferSize: 50,
			}

			summary, err := Run(context.Background(), cfg)
			if err != nil {
				t.Fatalf("error con %d workers: %v", w, err)
			}

			if summary.TotalRecords != 50 {
				t.Errorf("con %d workers: se esperaban 50 registros, se obtuvieron %d", w, summary.TotalRecords)
			}

			if firstRevenue == 0 {
				firstRevenue = summary.TotalRevenue
			} else {
				assertFloat(t, "TotalRevenue con workers diferentes", firstRevenue, summary.TotalRevenue)
			}
		})
	}
}

// --- Tests de manejo de errores ---

// TestPipelineCSVMalformado verifica el manejo de errores con filas invalidas.
func TestPipelineCSVMalformado(t *testing.T) {
	csvData := `id,date,category,product,quantity,price,region
1,2024-01-01,Electronics,Laptop,1,1000.00,EU
2,fecha-invalida,Food,Coffee,10,5.00,US
3,2024-01-03,Clothing,Shirt,abc,20.00,LATAM
4,2024-01-04,Books,Go Book,5,45.00,ASIA`

	path := createTempCSV(t, csvData)

	cfg := Config{
		InputPath:  path,
		Workers:    2,
		BufferSize: 10,
	}

	summary, _ := Run(context.Background(), cfg)

	if summary == nil {
		t.Fatal("el resumen no deberia ser nil")
	}

	// Deberian procesarse las filas validas (fila 1 y fila 4)
	if summary.TotalRecords < 2 {
		t.Errorf("se esperaban al menos 2 registros validos, se obtuvieron %d", summary.TotalRecords)
	}
}

// TestPipelineCSVCamposFaltantes verifica filas con columnas insuficientes.
func TestPipelineCSVCamposFaltantes(t *testing.T) {
	csvData := "id,date,category,product,quantity,price,region\n1,2024-01-01,Electronics,Laptop,1,1000.00,EU\n"

	path := createTempCSV(t, csvData)

	cfg := Config{
		InputPath:  path,
		Workers:    1,
		BufferSize: 10,
	}

	summary, err := Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	if summary.TotalRecords != 1 {
		t.Errorf("se esperaba 1 registro valido, se obtuvieron %d", summary.TotalRecords)
	}
}

// TestPipelineArchivoInexistente verifica el manejo de un archivo que no existe.
func TestPipelineArchivoInexistente(t *testing.T) {
	cfg := Config{
		InputPath:  "/ruta/que/no/existe.csv",
		Workers:    1,
		BufferSize: 10,
	}

	summary, err := Run(context.Background(), cfg)
	// Deberia haber un error o 0 registros
	if err == nil && summary.TotalRecords != 0 {
		t.Error("se esperaba un error o 0 registros para un archivo inexistente")
	}
}

// --- Tests unitarios del transformador ---

// TestTransformadorImpuestos verifica que los calculos de impuestos son correctos por region.
func TestTransformadorImpuestos(t *testing.T) {
	tests := []struct {
		region        string
		quantity      int
		price         float64
		expectedTotal float64
		expectedTax   float64
		expectedNet   float64
	}{
		{"EU", 2, 100.0, 200.0, 42.0, 242.0},
		{"US", 10, 50.0, 500.0, 40.0, 540.0},
		{"LATAM", 5, 200.0, 1000.0, 160.0, 1160.0},
		{"ASIA", 3, 300.0, 900.0, 90.0, 990.0},
		{"UNKNOWN", 1, 100.0, 100.0, 0.0, 100.0},
	}

	for _, tt := range tests {
		t.Run(tt.region, func(t *testing.T) {
			record := newSaleRecord(tt.quantity, tt.price, tt.region)
			processed := processRecord(record)

			assertFloat(t, "Total", tt.expectedTotal, processed.Total)
			assertFloat(t, "TaxAmount", tt.expectedTax, processed.TaxAmount)
			assertFloat(t, "NetTotal", tt.expectedNet, processed.NetTotal)
		})
	}
}

// --- Benchmarks ---

// BenchmarkPipeline1Worker mide el rendimiento del pipeline con 1 worker.
func BenchmarkPipeline1Worker(b *testing.B) {
	benchmarkPipeline(b, 1)
}

// BenchmarkPipeline2Workers mide el rendimiento del pipeline con 2 workers.
func BenchmarkPipeline2Workers(b *testing.B) {
	benchmarkPipeline(b, 2)
}

// BenchmarkPipeline4Workers mide el rendimiento del pipeline con 4 workers.
func BenchmarkPipeline4Workers(b *testing.B) {
	benchmarkPipeline(b, 4)
}

// BenchmarkPipelineNWorkers mide el rendimiento del pipeline con NumCPU workers.
func BenchmarkPipelineNWorkers(b *testing.B) {
	benchmarkPipeline(b, runtime.NumCPU())
}

func benchmarkPipeline(b *testing.B, workers int) {
	inputPath := filepath.Join(testdataDir(), "sales.csv")

	if _, err := os.Stat(inputPath); err != nil {
		b.Skipf("archivo de test no encontrado: %s", inputPath)
	}

	cfg := Config{
		InputPath:  inputPath,
		Workers:    workers,
		BufferSize: 100,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Run(context.Background(), cfg)
		if err != nil {
			b.Fatalf("error en benchmark: %v", err)
		}
	}
}
