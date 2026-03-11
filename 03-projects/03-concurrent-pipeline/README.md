# Project 3: Concurrent Data Pipeline

Data processing pipeline for sales data that demonstrates mastery of concurrency patterns in Go. Processes CSV files in parallel through stages connected by channels, calculates taxes by region, and generates aggregated reports.

## Pipeline Architecture

```
                              fan-out (N workers)
                             ┌─── Worker 1 ───┐
CSV ─── [Reader] ──► chan ──├─── Worker 2 ───├──► chan ──► [Aggregator] ──► Summary
                             ├─── Worker 3 ───┤                │
                             └─── Worker N ───┘                │
                                                         [Writer] ──► CSV
```

### Data Flow

1. **Reader** (`reader.go`) — Reads the CSV line by line and sends `SaleRecord` to a channel. Skips malformed rows without stopping the pipeline.
2. **Transformer** (`transformer.go`) — N concurrent workers consume records from the channel (fan-out), calculate totals and taxes, and send `ProcessedRecord` to the next channel. Coordinated with `sync.WaitGroup`.
3. **Aggregator** (`aggregator.go`) — Receives all processed records (fan-in) and builds a `Summary` with totals by category, region, and top products.
4. **Writer** (`writer.go`) — Writes the processed records to an output CSV, or prints the summary to stdout.

### Channels and Coordination

```
saleRecords (buffered)      processedRecords (buffered)     writerCh (buffered)
    Reader ──────────────► Transform ──────────────────► Aggregator ─────────► Writer
         close(out)              close(out) via WaitGroup       close(out)
```

- Each stage closes its output channel when it finishes
- Errors are reported through a dedicated channel (non-blocking)
- Supports graceful cancellation via `context.Context`
- SIGINT/SIGTERM signals trigger orderly shutdown

## Project Structure

```
03-concurrent-pipeline/
├── go.mod                    # Go module (stdlib only)
├── main.go                   # CLI entry point
├── pipeline/
│   ├── pipeline.go           # Pipeline orchestrator
│   ├── pipeline_test.go      # Integration tests and benchmarks
│   ├── reader.go             # CSV reading stage
│   ├── transformer.go        # Worker pool with tax calculation
│   ├── aggregator.go         # Fan-in and aggregation
│   └── writer.go             # CSV or stdout output
├── model/
│   └── record.go             # Data types
└── testdata/
    └── sales.csv             # Example dataset (50 records)
```

## Build and Run

```bash
# Build
go build -o pipeline .

# Run with summary on stdout
go run main.go -input testdata/sales.csv

# Run with CSV output and 8 workers
go run main.go -input testdata/sales.csv -output result.csv -workers 8

# With custom buffer
go run main.go -input testdata/sales.csv -workers 4 -buffer 200
```

### Available Flags

| Flag       | Description                              | Default             |
|------------|------------------------------------------|---------------------|
| `-input`   | Path to the input CSV (required)         | -                   |
| `-output`  | Path to the output CSV (optional)        | "" (stdout only)    |
| `-workers` | Number of concurrent workers             | `runtime.NumCPU()`  |
| `-buffer`  | Channel buffer size                      | 100                 |

## Example Output

```
Starting pipeline with 8 workers...
Input file: testdata/sales.csv

============================================================
          SALES PIPELINE SUMMARY
============================================================
  Records processed:    50
  Total revenue:        $45832.47
  Workers used:         8
  Processing time:      1.234ms
------------------------------------------------------------
  REVENUE BY CATEGORY:
    Electronics          $15234.56
    Food                 $3456.78
    Software             $2345.67
    Clothing             $8901.23
    Books                $4567.89
------------------------------------------------------------
  REVENUE BY REGION:
    EU                   $18234.56
    US                   $12345.67
    LATAM                $8901.23
    ASIA                 $6351.01
------------------------------------------------------------
  TOP PRODUCTS BY REVENUE:
     1. Laptop                Qty: 2      $2419.98
     2. Headphones            Qty: 55     $9982.50
     ...
============================================================
```

## Tests

```bash
# Run all tests
go test ./...

# Tests with verbose
go test -v ./pipeline/

# Only integration tests
go test -v -run TestPipeline ./pipeline/

# Only tax tests
go test -v -run TestTransformador ./pipeline/
```

## Benchmarks

```bash
# Run all benchmarks
go test -bench=. -benchmem ./pipeline/

# Compare 1 worker vs N workers
go test -bench=BenchmarkPipeline -benchmem -count=5 ./pipeline/

# With benchstat (install: go install golang.org/x/perf/cmd/benchstat@latest)
go test -bench=BenchmarkPipeline1Worker -benchmem -count=10 ./pipeline/ > 1worker.txt
go test -bench=BenchmarkPipelineNWorkers -benchmem -count=10 ./pipeline/ > nworkers.txt
benchstat 1worker.txt nworkers.txt
```

### Example Benchmark Results

```
BenchmarkPipeline1Worker-8     10000    105234 ns/op    45678 B/op    234 allocs/op
BenchmarkPipeline2Workers-8    12000     89456 ns/op    48901 B/op    267 allocs/op
BenchmarkPipeline4Workers-8    15000     72345 ns/op    52345 B/op    312 allocs/op
BenchmarkPipelineNWorkers-8    18000     65432 ns/op    56789 B/op    345 allocs/op
```

## Concurrency Patterns Demonstrated

| Pattern                   | Where It Is Used                                   |
|---------------------------|----------------------------------------------------|
| **Fan-out**               | `transformer.go` — N workers read from the same channel  |
| **Fan-in**                | `aggregator.go` — one consumer collects results    |
| **Worker Pool**           | `transformer.go` — configurable goroutine pool     |
| **Pipeline**              | `pipeline.go` — stages connected by channels       |
| **Context Cancellation**  | All stages respect `ctx.Done()`                    |
| **Graceful Shutdown**     | `main.go` — SIGINT/SIGTERM cancels the context     |
| **WaitGroup**             | `transformer.go` — coordinates worker shutdown     |
| **Buffered Channels**     | All channels use configurable buffers              |
| **Non-blocking Send**     | `reader.go` — non-blocking error sending           |

## Tax Rates by Region

| Region | Rate |
|--------|------|
| EU     | 21%  |
| US     | 8%   |
| LATAM  | 16%  |
| ASIA   | 10%  |

## Dependencies

Only Go's stdlib: `encoding/csv`, `context`, `sync`, `os`, `flag`, `os/signal`, `time`, `fmt`, `sort`, `strconv`, `io`, `strings`.
