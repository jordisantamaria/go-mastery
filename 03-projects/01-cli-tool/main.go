// Punto de entrada de la aplicacion CLI de gestion de tareas.
// Configura las dependencias e invoca el dispatcher de comandos.
//
// Uso:
//
//	task add "titulo"     — crear nueva tarea
//	task list [--all]     — listar tareas
//	task done <id>        — completar tarea
//	task delete <id>      — eliminar tarea
//	task help             — mostrar ayuda
package main

import (
	"os"
	"path/filepath"

	"github.com/jordi-nyxidiom/go-mastery/03-projects/01-cli-tool/internal/cli"
	"github.com/jordi-nyxidiom/go-mastery/03-projects/01-cli-tool/internal/task"
)

// defaultFilename es el nombre del fichero JSON donde se almacenan las tareas.
const defaultFilename = ".tasks.json"

func main() {
	// Determinar la ruta del fichero de tareas.
	// Se guarda en el directorio home del usuario.
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	storePath := filepath.Join(home, defaultFilename)

	// Permitir sobreescribir la ruta con variable de entorno (util para testing).
	if envPath := os.Getenv("TASK_FILE"); envPath != "" {
		storePath = envPath
	}

	// Inicializar dependencias.
	store := task.NewJSONStore(storePath)
	app := &cli.App{
		Store:  store,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	// Ejecutar el comando. os.Args[0] es el nombre del programa, lo omitimos.
	code := app.Run(os.Args[1:])
	os.Exit(code)
}
