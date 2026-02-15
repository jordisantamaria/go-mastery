# Proyecto 1: CLI Task Manager

CLI para gestionar tareas desde la terminal. Demuestra Go idiomatico, testing y packaging.

## Stack
- [Cobra](https://github.com/spf13/cobra) — framework CLI
- [BoltDB](https://github.com/etcd-io/bbolt) — base de datos embebida

## Features
- `task add "Comprar leche"` — Anyadir tarea
- `task list` — Listar tareas pendientes
- `task done <id>` — Marcar como completada
- `task delete <id>` — Eliminar tarea

## Lo que aprenderas
- Estructura de proyecto Go
- Manejo de argumentos y flags
- Persistencia con BoltDB
- Testing de CLIs
- Build y distribucion con `go build`

> En progreso — semanas 1-2 del roadmap.
