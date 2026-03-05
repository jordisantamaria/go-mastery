// Tests para el dispatcher CLI — validacion del parseo de comandos y salida formateada.
// Demuestra: bytes.Buffer como io.Writer, table-driven tests, tests de integracion ligeros.
package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jordi-nyxidiom/go-mastery/03-projects/01-cli-tool/internal/task"
)

// newTestApp crea una App con un store temporal y buffers para capturar la salida.
func newTestApp(t *testing.T) (*App, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	dir := t.TempDir()
	store := task.NewJSONStore(filepath.Join(dir, "tasks.json"))
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	return &App{Store: store, Out: out, ErrOut: errOut}, out, errOut
}

func TestRunAdd(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantCode int
		wantOut  string
		wantErr  string
	}{
		{
			name:     "anadir tarea valida",
			args:     []string{"add", "Comprar leche"},
			wantCode: 0,
			wantOut:  "Tarea #1 creada",
		},
		{
			name:     "anadir tarea con multiples palabras sin comillas",
			args:     []string{"add", "Estudiar", "Go", "concurrencia"},
			wantCode: 0,
			wantOut:  "Estudiar Go concurrencia",
		},
		{
			name:     "add sin argumentos",
			args:     []string{"add"},
			wantCode: 1,
			wantErr:  "se requiere un titulo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, out, errOut := newTestApp(t)
			code := app.Run(tt.args)

			if code != tt.wantCode {
				t.Errorf("codigo de salida: esperado %d, obtenido %d", tt.wantCode, code)
			}

			if tt.wantOut != "" && !strings.Contains(out.String(), tt.wantOut) {
				t.Errorf("salida esperada contiene '%s', obtenida: '%s'", tt.wantOut, out.String())
			}

			if tt.wantErr != "" && !strings.Contains(errOut.String(), tt.wantErr) {
				t.Errorf("error esperado contiene '%s', obtenido: '%s'", tt.wantErr, errOut.String())
			}
		})
	}
}

func TestRunList(t *testing.T) {
	t.Run("lista vacia", func(t *testing.T) {
		app, out, _ := newTestApp(t)
		code := app.Run([]string{"list"})

		if code != 0 {
			t.Fatalf("codigo de salida esperado 0, obtenido %d", code)
		}
		if !strings.Contains(out.String(), "No hay tareas pendientes") {
			t.Errorf("salida esperada: mensaje de lista vacia, obtenida: '%s'", out.String())
		}
	})

	t.Run("listar tareas pendientes", func(t *testing.T) {
		app, out, _ := newTestApp(t)
		app.Run([]string{"add", "Tarea 1"})
		app.Run([]string{"add", "Tarea 2"})
		out.Reset()

		code := app.Run([]string{"list"})
		if code != 0 {
			t.Fatalf("codigo de salida esperado 0, obtenido %d", code)
		}

		output := out.String()
		if !strings.Contains(output, "Tarea 1") || !strings.Contains(output, "Tarea 2") {
			t.Errorf("la salida deberia contener ambas tareas: '%s'", output)
		}
		if !strings.Contains(output, "pendiente") {
			t.Errorf("la salida deberia mostrar estado 'pendiente': '%s'", output)
		}
	})

	t.Run("listar con --all muestra completadas", func(t *testing.T) {
		app, out, _ := newTestApp(t)
		app.Run([]string{"add", "Tarea pendiente"})
		app.Run([]string{"add", "Tarea completada"})
		app.Run([]string{"done", "2"})
		out.Reset()

		code := app.Run([]string{"list", "--all"})
		if code != 0 {
			t.Fatalf("codigo de salida esperado 0, obtenido %d", code)
		}

		output := out.String()
		if !strings.Contains(output, "pendiente") {
			t.Errorf("la salida deberia contener 'pendiente': '%s'", output)
		}
		if !strings.Contains(output, "completada") {
			t.Errorf("la salida deberia contener 'completada': '%s'", output)
		}
	})

	t.Run("listar sin --all oculta completadas", func(t *testing.T) {
		app, out, _ := newTestApp(t)
		app.Run([]string{"add", "Tarea pendiente"})
		app.Run([]string{"add", "Tarea a completar"})
		app.Run([]string{"done", "2"})
		out.Reset()

		code := app.Run([]string{"list"})
		if code != 0 {
			t.Fatalf("codigo de salida esperado 0, obtenido %d", code)
		}

		output := out.String()
		if !strings.Contains(output, "Tarea pendiente") {
			t.Errorf("la salida deberia contener la tarea pendiente: '%s'", output)
		}
		if strings.Contains(output, "Tarea a completar") {
			t.Errorf("la salida no deberia contener la tarea completada: '%s'", output)
		}
	})
}

func TestRunDone(t *testing.T) {
	tests := []struct {
		name     string
		setup    [][]string // comandos previos para preparar el estado
		args     []string
		wantCode int
		wantOut  string
		wantErr  string
	}{
		{
			name:     "completar tarea existente",
			setup:    [][]string{{"add", "Mi tarea"}},
			args:     []string{"done", "1"},
			wantCode: 0,
			wantOut:  "Tarea #1 completada",
		},
		{
			name:     "completar sin ID",
			args:     []string{"done"},
			wantCode: 1,
			wantErr:  "se requiere el ID",
		},
		{
			name:     "completar con ID invalido",
			args:     []string{"done", "abc"},
			wantCode: 1,
			wantErr:  "no es un ID valido",
		},
		{
			name:     "completar tarea inexistente",
			args:     []string{"done", "999"},
			wantCode: 1,
			wantErr:  "tarea no encontrada",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, out, errOut := newTestApp(t)

			for _, cmd := range tt.setup {
				app.Run(cmd)
			}
			out.Reset()
			errOut.Reset()

			code := app.Run(tt.args)

			if code != tt.wantCode {
				t.Errorf("codigo de salida: esperado %d, obtenido %d", tt.wantCode, code)
			}

			if tt.wantOut != "" && !strings.Contains(out.String(), tt.wantOut) {
				t.Errorf("salida esperada contiene '%s', obtenida: '%s'", tt.wantOut, out.String())
			}

			if tt.wantErr != "" && !strings.Contains(errOut.String(), tt.wantErr) {
				t.Errorf("error esperado contiene '%s', obtenido: '%s'", tt.wantErr, errOut.String())
			}
		})
	}
}

func TestRunDelete(t *testing.T) {
	tests := []struct {
		name     string
		setup    [][]string
		args     []string
		wantCode int
		wantOut  string
		wantErr  string
	}{
		{
			name:     "eliminar tarea existente",
			setup:    [][]string{{"add", "Tarea temporal"}},
			args:     []string{"delete", "1"},
			wantCode: 0,
			wantOut:  "Tarea #1 eliminada",
		},
		{
			name:     "eliminar sin ID",
			args:     []string{"delete"},
			wantCode: 1,
			wantErr:  "se requiere el ID",
		},
		{
			name:     "eliminar con ID invalido",
			args:     []string{"delete", "xyz"},
			wantCode: 1,
			wantErr:  "no es un ID valido",
		},
		{
			name:     "eliminar tarea inexistente",
			args:     []string{"delete", "999"},
			wantCode: 1,
			wantErr:  "tarea no encontrada",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, out, errOut := newTestApp(t)

			for _, cmd := range tt.setup {
				app.Run(cmd)
			}
			out.Reset()
			errOut.Reset()

			code := app.Run(tt.args)

			if code != tt.wantCode {
				t.Errorf("codigo de salida: esperado %d, obtenido %d", tt.wantCode, code)
			}

			if tt.wantOut != "" && !strings.Contains(out.String(), tt.wantOut) {
				t.Errorf("salida esperada contiene '%s', obtenida: '%s'", tt.wantOut, out.String())
			}

			if tt.wantErr != "" && !strings.Contains(errOut.String(), tt.wantErr) {
				t.Errorf("error esperado contiene '%s', obtenido: '%s'", tt.wantErr, errOut.String())
			}
		})
	}
}

func TestRunHelp(t *testing.T) {
	variants := []string{"help", "--help", "-h"}

	for _, cmd := range variants {
		t.Run(cmd, func(t *testing.T) {
			app, out, _ := newTestApp(t)
			code := app.Run([]string{cmd})

			if code != 0 {
				t.Errorf("codigo de salida esperado 0 para '%s', obtenido %d", cmd, code)
			}

			output := out.String()
			if !strings.Contains(output, "Gestor de tareas CLI") {
				t.Errorf("la ayuda deberia contener el titulo: '%s'", output)
			}
			if !strings.Contains(output, "add") || !strings.Contains(output, "list") {
				t.Errorf("la ayuda deberia listar los comandos: '%s'", output)
			}
		})
	}
}

func TestRunNoArgs(t *testing.T) {
	app, _, _ := newTestApp(t)
	code := app.Run([]string{})

	if code != 1 {
		t.Errorf("sin argumentos deberia devolver codigo 1, obtenido %d", code)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	app, _, errOut := newTestApp(t)
	code := app.Run([]string{"foo"})

	if code != 1 {
		t.Errorf("comando desconocido deberia devolver codigo 1, obtenido %d", code)
	}
	if !strings.Contains(errOut.String(), "comando desconocido") {
		t.Errorf("deberia indicar comando desconocido: '%s'", errOut.String())
	}
}
