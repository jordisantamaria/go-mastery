# CLI Task Manager

Task manager from the terminal, written in pure Go (stdlib only). A portfolio project that demonstrates mastery of CLI development with Go.

## Features

- Create, list, complete, and delete tasks
- Persistence in a JSON file (no external dependencies)
- Formatted table-style output
- Clean, idiomatic, and well-tested code

## Build and Run

```bash
# Build the binary
go build -o task .

# Or run directly
go run . <command>
```

## Usage

```bash
# Add tasks
./task add "Buy milk"
./task add "Study Go"

# List pending tasks
./task list

# List all (including completed)
./task list --all

# Mark as completed
./task done 1

# Delete a task
./task delete 2

# Show help
./task help
```

## Example Session

```
$ ./task add "Buy milk"
✓ Task #1 created: "Buy milk"

$ ./task add "Study Go"
✓ Task #2 created: "Study Go"

$ ./task list
  ID   | Status     | Created    | Title
  ---  | ---------- | ---------- | ------
  1    | pending    | 2024-01-15 | Buy milk
  2    | pending    | 2024-01-15 | Study Go

$ ./task done 1
✓ Task #1 completed

$ ./task list --all
  ID   | Status     | Created    | Title
  ---  | ---------- | ---------- | ------
  1    | completed  | 2024-01-15 | Buy milk
  2    | pending    | 2024-01-15 | Study Go

$ ./task delete 1
✓ Task #1 deleted
```

## Architecture

```
01-cli-tool/
├── go.mod                  # Independent Go module (no external dependencies)
├── main.go                 # Entry point: configures dependencies and runs
├── README.md
├── internal/
│   ├── task/
│   │   ├── task.go         # Task model and Status type
│   │   ├── store.go        # Store interface + JSONStore implementation
│   │   └── store_test.go   # Store tests (persistence, edge cases)
│   └── cli/
│       ├── cli.go          # Command dispatcher and output formatting
│       └── cli_test.go     # CLI tests (parsing, integration)
└── testdata/               # Test fixtures (reserved)
```

### Design Decisions

- **Stdlib only**: no Cobra, BoltDB, or any external dependency is used. Everything is solved with `flag`, `encoding/json`, `os`, and `fmt`. This demonstrates deep knowledge of the standard library.
- **Store interface**: the store is defined as an interface, which allows changing the implementation (for example to SQLite) without touching the command logic.
- **Dependency injection**: `cli.App` receives `Store`, `Out`, and `ErrOut` as fields, making the code fully testable without complex mocks.
- **`internal/`**: internal packages are not importable from outside the module, respecting Go's encapsulation.
- **Mutex in JSONStore**: ensures safety for concurrent access to the file.
- **`TASK_FILE` environment variable**: allows configuring the data file path, useful for tests and different environments.

## Go Patterns Demonstrated

| Pattern | Where It Is Applied |
|---|---|
| Interfaces | `task.Store` as the store contract |
| Dependency injection | `cli.App` receives its dependencies |
| `internal/` packages | Module-level encapsulation |
| Table-driven tests | Parameterized tests in `cli_test.go` and `store_test.go` |
| Subtests (`t.Run`) | Hierarchical test organization |
| `t.TempDir()` | Temporary directories that are automatically cleaned up |
| `io.Writer` | Output abstraction for testing |
| `flag.NewFlagSet` | Per-subcommand flag parsing |
| Error wrapping (`%w`) | Error chains with context |
| Sentinel errors | `ErrTaskNotFound`, `ErrEmptyTitle` |
| Mutex (`sync.Mutex`) | Safe concurrency in the store |
| JSON marshaling | Struct tags for serialization |
| Methods with receiver | `Task.StatusLabel()` |

## Tests

```bash
# Run all tests
go test ./...

# With detail
go test -v ./...

# With coverage
go test -cover ./...
```

Tests cover:
- **Store**: create, list, complete, delete tasks; persistence between instances; ID auto-increment; errors (empty title, non-existent ID, task already completed)
- **CLI**: parsing of each subcommand; flags (`--all`); handling of invalid arguments; formatted output; correct exit codes

## Storage

Tasks are saved in `~/.tasks.json` by default. You can change the path with the `TASK_FILE` environment variable:

```bash
TASK_FILE=/tmp/my-tasks.json ./task list
```

The file format is human-readable JSON:

```json
{
  "next_id": 3,
  "tasks": [
    {
      "id": 1,
      "title": "Buy milk",
      "status": "done",
      "created_at": "2024-01-15T10:30:00Z",
      "done_at": "2024-01-15T14:00:00Z"
    },
    {
      "id": 2,
      "title": "Study Go",
      "status": "pending",
      "created_at": "2024-01-15T10:31:00Z"
    }
  ]
}
```
