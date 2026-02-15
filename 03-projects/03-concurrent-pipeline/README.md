# Proyecto 3: Concurrent Data Pipeline

Pipeline de procesamiento de datos concurrente. Demuestra dominio de goroutines, channels y patrones de concurrencia.

## Que hace
Procesa un dataset grande (CSV/JSON) en paralelo:
1. **Reader** — lee datos del archivo en chunks
2. **Transformer** — N workers procesan en paralelo (fan-out)
3. **Aggregator** — recoge resultados (fan-in)
4. **Writer** — escribe el output

## Patrones de concurrencia
- Fan-out / Fan-in
- Worker pool con numero configurable de workers
- Rate limiting
- Graceful shutdown con context
- Error propagation con errgroup

## Lo que aprenderas
- Diseno de pipelines con channels
- Control de concurrencia
- Graceful shutdown
- Benchmarking y profiling

> En progreso — semanas 7-8 del roadmap.
