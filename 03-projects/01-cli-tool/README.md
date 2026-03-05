# CLI Task Manager

Gestor de tareas desde la terminal, escrito en Go puro (solo stdlib). Proyecto de portfolio que demuestra dominio del desarrollo CLI con Go.

## Caracteristicas

- Crear, listar, completar y eliminar tareas
- Persistencia en fichero JSON (sin dependencias externas)
- Salida formateada tipo tabla
- Mensajes en espanol para el usuario
- Codigo limpio, idiomatico y bien testeado

## Compilar y ejecutar

```bash
# Compilar el binario
go build -o task .

# O ejecutar directamente
go run . <comando>
```

## Uso

```bash
# Anadir tareas
./task add "Comprar leche"
./task add "Estudiar Go"

# Listar tareas pendientes
./task list

# Listar todas (incluidas completadas)
./task list --all

# Marcar como completada
./task done 1

# Eliminar una tarea
./task delete 2

# Ver ayuda
./task help
```

## Ejemplo de sesion

```
$ ./task add "Comprar leche"
✓ Tarea #1 creada: "Comprar leche"

$ ./task add "Estudiar Go"
✓ Tarea #2 creada: "Estudiar Go"

$ ./task list
  ID   | Estado     | Creada     | Titulo
  ---  | ---------- | ---------- | ------
  1    | pendiente  | 2024-01-15 | Comprar leche
  2    | pendiente  | 2024-01-15 | Estudiar Go

$ ./task done 1
✓ Tarea #1 completada

$ ./task list --all
  ID   | Estado     | Creada     | Titulo
  ---  | ---------- | ---------- | ------
  1    | completada | 2024-01-15 | Comprar leche
  2    | pendiente  | 2024-01-15 | Estudiar Go

$ ./task delete 1
✓ Tarea #1 eliminada
```

## Arquitectura

```
01-cli-tool/
├── go.mod                  # Modulo Go independiente (sin dependencias externas)
├── main.go                 # Punto de entrada: configura dependencias y ejecuta
├── README.md
├── internal/
│   ├── task/
│   │   ├── task.go         # Modelo Task y tipo Status
│   │   ├── store.go        # Interfaz Store + implementacion JSONStore
│   │   └── store_test.go   # Tests del almacen (persistencia, edge cases)
│   └── cli/
│       ├── cli.go          # Dispatcher de comandos y formateo de salida
│       └── cli_test.go     # Tests del CLI (parseo, integracion)
└── testdata/               # Fixtures para tests (reservado)
```

### Decisiones de diseno

- **Solo stdlib**: no se usa Cobra, BoltDB ni ninguna dependencia externa. Todo se resuelve con `flag`, `encoding/json`, `os` y `fmt`. Esto demuestra conocimiento profundo de la biblioteca estandar.
- **Interfaz Store**: el almacen se define como interfaz, lo que permite cambiar la implementacion (por ejemplo a SQLite) sin tocar la logica de comandos.
- **Inyeccion de dependencias**: `cli.App` recibe `Store`, `Out` y `ErrOut` como campos, haciendo el codigo completamente testable sin mocks complejos.
- **`internal/`**: los paquetes internos no son importables desde fuera del modulo, respetando la encapsulacion de Go.
- **Mutex en JSONStore**: garantiza seguridad en accesos concurrentes al fichero.
- **Variable de entorno `TASK_FILE`**: permite configurar la ruta del fichero de datos, util para tests y entornos diferentes.

## Patrones de Go demostrados

| Patron | Donde se aplica |
|---|---|
| Interfaces | `task.Store` como contrato del almacen |
| Inyeccion de dependencias | `cli.App` recibe sus dependencias |
| `internal/` packages | Encapsulacion a nivel de modulo |
| Table-driven tests | Tests parametrizados en `cli_test.go` y `store_test.go` |
| Subtests (`t.Run`) | Organizacion jerarquica de tests |
| `t.TempDir()` | Directorios temporales que se limpian automaticamente |
| `io.Writer` | Abstraccion de salida para testing |
| `flag.NewFlagSet` | Parseo de flags por subcomando |
| Error wrapping (`%w`) | Cadenas de errores con contexto |
| Sentinel errors | `ErrTaskNotFound`, `ErrEmptyTitle` |
| Mutex (`sync.Mutex`) | Concurrencia segura en el almacen |
| JSON marshaling | Tags struct para serializacion |
| Metodos con receiver | `Task.StatusLabel()` |

## Tests

```bash
# Ejecutar todos los tests
go test ./...

# Con detalle
go test -v ./...

# Con cobertura
go test -cover ./...
```

Los tests cubren:
- **Store**: crear, listar, completar, eliminar tareas; persistencia entre instancias; auto-incremento de IDs; errores (titulo vacio, ID inexistente, tarea ya completada)
- **CLI**: parseo de cada subcomando; flags (`--all`); manejo de argumentos invalidos; salida formateada; codigos de salida correctos

## Almacenamiento

Las tareas se guardan en `~/.tasks.json` por defecto. Puedes cambiar la ruta con la variable de entorno `TASK_FILE`:

```bash
TASK_FILE=/tmp/mis-tareas.json ./task list
```

El formato del fichero es JSON legible:

```json
{
  "next_id": 3,
  "tasks": [
    {
      "id": 1,
      "title": "Comprar leche",
      "status": "done",
      "created_at": "2024-01-15T10:30:00Z",
      "done_at": "2024-01-15T14:00:00Z"
    },
    {
      "id": 2,
      "title": "Estudiar Go",
      "status": "pending",
      "created_at": "2024-01-15T10:31:00Z"
    }
  ]
}
```
