// Package cli implementa el dispatcher de comandos y el formateo de salida.
// Conecta los subcomandos del usuario con las operaciones del almacen de tareas.
// Demuestra: flag package, formateo de tablas, patron Command.
package cli

import (
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/jordi-nyxidiom/go-mastery/03-projects/01-cli-tool/internal/task"
)

// App contiene la configuracion y dependencias de la aplicacion CLI.
// Inyectamos Store como interfaz para facilitar el testing.
type App struct {
	Store  task.Store
	Out    io.Writer // Salida configurable (stdout en produccion, buffer en tests).
	ErrOut io.Writer // Salida de errores configurable.
}

// Run procesa los argumentos de la linea de comandos y ejecuta el subcomando correspondiente.
// Devuelve un codigo de salida: 0 para exito, 1 para error.
func (a *App) Run(args []string) int {
	if len(args) < 1 {
		a.printUsage()
		return 1
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "add":
		return a.runAdd(subArgs)
	case "list":
		return a.runList(subArgs)
	case "done":
		return a.runDone(subArgs)
	case "delete":
		return a.runDelete(subArgs)
	case "help", "--help", "-h":
		a.printUsage()
		return 0
	default:
		fmt.Fprintf(a.ErrOut, "Error: comando desconocido '%s'\n\n", subcommand)
		a.printUsage()
		return 1
	}
}

// runAdd anade una nueva tarea con el titulo proporcionado.
func (a *App) runAdd(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(a.ErrOut, "Error: se requiere un titulo para la tarea")
		fmt.Fprintln(a.ErrOut, "Uso: task add \"titulo de la tarea\"")
		return 1
	}

	// Unir todos los argumentos como titulo (permite escribir sin comillas).
	title := strings.Join(args, " ")

	t, err := a.Store.Add(title)
	if err != nil {
		fmt.Fprintf(a.ErrOut, "Error: %v\n", err)
		return 1
	}

	fmt.Fprintf(a.Out, "\u2713 Tarea #%d creada: \"%s\"\n", t.ID, t.Title)
	return 0
}

// runList muestra las tareas en formato tabla.
func (a *App) runList(args []string) int {
	// Parsear flags del subcomando 'list'.
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(a.ErrOut)
	showAll := fs.Bool("all", false, "Mostrar tambien las tareas completadas")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	tasks, err := a.Store.List(*showAll)
	if err != nil {
		fmt.Fprintf(a.ErrOut, "Error: %v\n", err)
		return 1
	}

	if len(tasks) == 0 {
		if *showAll {
			fmt.Fprintln(a.Out, "No hay tareas registradas.")
		} else {
			fmt.Fprintln(a.Out, "No hay tareas pendientes.")
		}
		return 0
	}

	a.printTable(tasks)
	return 0
}

// runDone marca una tarea como completada por su ID.
func (a *App) runDone(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(a.ErrOut, "Error: se requiere el ID de la tarea")
		fmt.Fprintln(a.ErrOut, "Uso: task done <id>")
		return 1
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(a.ErrOut, "Error: '%s' no es un ID valido\n", args[0])
		return 1
	}

	if err := a.Store.Complete(id); err != nil {
		fmt.Fprintf(a.ErrOut, "Error: %v\n", err)
		return 1
	}

	fmt.Fprintf(a.Out, "\u2713 Tarea #%d completada\n", id)
	return 0
}

// runDelete elimina una tarea por su ID.
func (a *App) runDelete(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(a.ErrOut, "Error: se requiere el ID de la tarea")
		fmt.Fprintln(a.ErrOut, "Uso: task delete <id>")
		return 1
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(a.ErrOut, "Error: '%s' no es un ID valido\n", args[0])
		return 1
	}

	if err := a.Store.Delete(id); err != nil {
		fmt.Fprintf(a.ErrOut, "Error: %v\n", err)
		return 1
	}

	fmt.Fprintf(a.Out, "\u2713 Tarea #%d eliminada\n", id)
	return 0
}

// printTable renderiza las tareas en formato tabla alineado.
func (a *App) printTable(tasks []task.Task) {
	// Calcular ancho maximo del titulo para alinear columnas.
	maxTitle := len("Titulo")
	for _, t := range tasks {
		if len(t.Title) > maxTitle {
			maxTitle = len(t.Title)
		}
	}

	// Cabecera.
	header := fmt.Sprintf("  %-4s | %-10s | %-10s | %s", "ID", "Estado", "Creada", "Titulo")
	separator := fmt.Sprintf("  %-4s | %-10s | %-10s | %s", "---", "----------", "----------", "------")

	fmt.Fprintln(a.Out, header)
	fmt.Fprintln(a.Out, separator)

	// Filas de datos.
	for _, t := range tasks {
		date := t.CreatedAt.Format("2006-01-02")
		fmt.Fprintf(a.Out, "  %-4d | %-10s | %-10s | %s\n", t.ID, t.StatusLabel(), date, t.Title)
	}
}

// printUsage muestra la ayuda de uso de la aplicacion.
func (a *App) printUsage() {
	usage := `Gestor de tareas CLI — escrito en Go

Uso:
  task <comando> [argumentos]

Comandos:
  add <titulo>     Anadir una nueva tarea
  list [--all]     Listar tareas (pendientes por defecto, --all para todas)
  done <id>        Marcar una tarea como completada
  delete <id>      Eliminar una tarea
  help             Mostrar esta ayuda

Ejemplos:
  task add "Comprar leche"
  task list
  task list --all
  task done 1
  task delete 2`

	fmt.Fprintln(a.Out, usage)
}
