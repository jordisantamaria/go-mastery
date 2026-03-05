# Proyecto 3: Pipeline Concurrente de Datos

Pipeline de procesamiento de datos de ventas que demuestra dominio de patrones de concurrencia en Go. Procesa archivos CSV en paralelo a traves de etapas conectadas por canales, calcula impuestos por region y genera reportes agregados.

## Arquitectura del Pipeline

```
                              fan-out (N workers)
                             ┌─── Worker 1 ───┐
CSV ─── [Reader] ──► chan ──├─── Worker 2 ───├──► chan ──► [Aggregator] ──► Summary
                             ├─── Worker 3 ───┤                │
                             └─── Worker N ───┘                │
                                                         [Writer] ──► CSV
```

### Flujo de datos

1. **Reader** (`reader.go`) — Lee el CSV linea por linea y envia `SaleRecord` a un canal. Salta filas mal formadas sin detener el pipeline.
2. **Transformer** (`transformer.go`) — N workers concurrentes consumen registros del canal (fan-out), calculan totales e impuestos, y envian `ProcessedRecord` al siguiente canal. Coordinados con `sync.WaitGroup`.
3. **Aggregator** (`aggregator.go`) — Recibe todos los registros procesados (fan-in) y construye un `Summary` con totales por categoria, region y top productos.
4. **Writer** (`writer.go`) — Escribe los registros procesados a un CSV de salida, o imprime el resumen por stdout.

### Canales y coordinacion

```
saleRecords (buffered)      processedRecords (buffered)     writerCh (buffered)
    Reader ──────────────► Transform ──────────────────► Aggregator ─────────► Writer
         close(out)              close(out) via WaitGroup       close(out)
```

- Cada etapa cierra su canal de salida cuando termina
- Los errores se reportan por un canal dedicado (no bloqueante)
- Soporta cancelacion graceful via `context.Context`
- Senales SIGINT/SIGTERM disparan el shutdown ordenado

## Estructura del proyecto

```
03-concurrent-pipeline/
├── go.mod                    # Modulo Go (stdlib only)
├── main.go                   # Punto de entrada CLI
├── pipeline/
│   ├── pipeline.go           # Orquestador del pipeline
│   ├── pipeline_test.go      # Tests de integracion y benchmarks
│   ├── reader.go             # Etapa de lectura CSV
│   ├── transformer.go        # Worker pool con calculo de impuestos
│   ├── aggregator.go         # Fan-in y agregacion
│   └── writer.go             # Salida CSV o stdout
├── model/
│   └── record.go             # Tipos de datos
└── testdata/
    └── sales.csv             # Dataset de ejemplo (50 registros)
```

## Compilar y ejecutar

```bash
# Compilar
go build -o pipeline .

# Ejecutar con resumen en stdout
go run main.go -input testdata/sales.csv

# Ejecutar con salida CSV y 8 workers
go run main.go -input testdata/sales.csv -output resultado.csv -workers 8

# Con buffer personalizado
go run main.go -input testdata/sales.csv -workers 4 -buffer 200
```

### Flags disponibles

| Flag       | Descripcion                              | Valor por defecto   |
|------------|------------------------------------------|---------------------|
| `-input`   | Ruta al CSV de entrada (obligatorio)     | -                   |
| `-output`  | Ruta al CSV de salida (opcional)         | "" (solo stdout)    |
| `-workers` | Numero de workers concurrentes           | `runtime.NumCPU()`  |
| `-buffer`  | Tamano del buffer de canales             | 100                 |

## Ejemplo de salida

```
Iniciando pipeline con 8 workers...
Archivo de entrada: testdata/sales.csv

============================================================
          RESUMEN DEL PIPELINE DE VENTAS
============================================================
  Registros procesados:   50
  Ingresos totales:       $45832.47
  Workers utilizados:     8
  Tiempo de proceso:      1.234ms
------------------------------------------------------------
  INGRESOS POR CATEGORIA:
    Electronics          $15234.56
    Food                 $3456.78
    Software             $2345.67
    Clothing             $8901.23
    Books                $4567.89
------------------------------------------------------------
  INGRESOS POR REGION:
    EU                   $18234.56
    US                   $12345.67
    LATAM                $8901.23
    ASIA                 $6351.01
------------------------------------------------------------
  TOP PRODUCTOS POR INGRESOS:
     1. Laptop                Qty: 2      $2419.98
     2. Headphones            Qty: 55     $9982.50
     ...
============================================================
```

## Tests

```bash
# Ejecutar todos los tests
go test ./...

# Tests con verbose
go test -v ./pipeline/

# Solo tests de integracion
go test -v -run TestPipeline ./pipeline/

# Solo tests de impuestos
go test -v -run TestTransformador ./pipeline/
```

## Benchmarks

```bash
# Ejecutar todos los benchmarks
go test -bench=. -benchmem ./pipeline/

# Comparar 1 worker vs N workers
go test -bench=BenchmarkPipeline -benchmem -count=5 ./pipeline/

# Con benchstat (instalar: go install golang.org/x/perf/cmd/benchstat@latest)
go test -bench=BenchmarkPipeline1Worker -benchmem -count=10 ./pipeline/ > 1worker.txt
go test -bench=BenchmarkPipelineNWorkers -benchmem -count=10 ./pipeline/ > nworkers.txt
benchstat 1worker.txt nworkers.txt
```

### Ejemplo de resultados de benchmark

```
BenchmarkPipeline1Worker-8     10000    105234 ns/op    45678 B/op    234 allocs/op
BenchmarkPipeline2Workers-8    12000     89456 ns/op    48901 B/op    267 allocs/op
BenchmarkPipeline4Workers-8    15000     72345 ns/op    52345 B/op    312 allocs/op
BenchmarkPipelineNWorkers-8    18000     65432 ns/op    56789 B/op    345 allocs/op
```

## Patrones de concurrencia demostrados

| Patron                    | Donde se usa                                       |
|---------------------------|----------------------------------------------------|
| **Fan-out**               | `transformer.go` — N workers leen del mismo canal  |
| **Fan-in**                | `aggregator.go` — un consumidor recoge resultados  |
| **Worker Pool**           | `transformer.go` — pool configurable de goroutines |
| **Pipeline**              | `pipeline.go` — etapas conectadas por canales      |
| **Context Cancellation**  | Todas las etapas respetan `ctx.Done()`             |
| **Graceful Shutdown**     | `main.go` — SIGINT/SIGTERM cancela el contexto     |
| **WaitGroup**             | `transformer.go` — coordina cierre de workers      |
| **Buffered Channels**     | Todos los canales usan buffer configurable         |
| **Non-blocking Send**     | `reader.go` — envio de errores sin bloqueo         |

## Tasas de impuesto por region

| Region | Tasa |
|--------|------|
| EU     | 21%  |
| US     | 8%   |
| LATAM  | 16%  |
| ASIA   | 10%  |

## Dependencias

Solo stdlib de Go: `encoding/csv`, `context`, `sync`, `os`, `flag`, `os/signal`, `time`, `fmt`, `sort`, `strconv`, `io`, `strings`.
