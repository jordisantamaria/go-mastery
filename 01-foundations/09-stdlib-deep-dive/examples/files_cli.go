// Ejemplo: flag, slog, filepath.Walk, time, strings.Builder, strings.Cut.
//
// Ejecutar: go run ./01-foundations/09-stdlib-deep-dive/examples/files_cli.go
//
// Con flags:
//   go run ./01-foundations/09-stdlib-deep-dive/examples/files_cli.go -dir=. -ext=.go -format=json
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	demoFlags()
	demoSlog()
	demoFilepathWalk()
	demoTime()
	demoStringsBuilder()
	demoStringsCut()
}

// =============================================================================
// flag — parsing de argumentos CLI
// =============================================================================

func demoFlags() {
	fmt.Println("=== flag package ===")

	// Crear un FlagSet separado para no interferir con el global
	// (en una app real usarias flag.String, flag.Parse directamente)
	fs := flag.NewFlagSet("demo", flag.ContinueOnError)

	dir := fs.String("dir", ".", "directorio a escanear")
	ext := fs.String("ext", ".go", "extension a buscar")
	format := fs.String("format", "text", "formato de log (text/json)")
	verbose := fs.Bool("verbose", false, "activar modo verbose")

	// Custom usage
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Uso: files_cli [opciones]\n\nOpciones:\n")
		fs.PrintDefaults()
	}

	// Parsear os.Args (ignorar errores en la demo)
	_ = fs.Parse(os.Args[1:])

	fmt.Printf("  dir=%q ext=%q format=%q verbose=%t\n", *dir, *ext, *format, *verbose)
	fmt.Printf("  args restantes: %v\n\n", fs.Args())
}

// =============================================================================
// log/slog — structured logging
// =============================================================================

func demoSlog() {
	fmt.Println("=== slog (Structured Logging) ===")

	// --- Text Handler (formato key=value, bueno para desarrollo) ---
	fmt.Println("-- Text Handler --")
	textLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug, // mostrar todos los niveles
	}))

	textLogger.Debug("mensaje debug", "modulo", "auth")
	textLogger.Info("usuario logueado", "userID", 42, "email", "alice@example.com")
	textLogger.Warn("disco casi lleno", "usage_percent", 92.5)
	textLogger.Error("conexion fallida", "host", "db.example.com", "retries", 3)

	// --- JSON Handler (formato JSON, bueno para produccion/log aggregation) ---
	fmt.Println("\n-- JSON Handler --")
	jsonLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo, // solo INFO y superior
	}))

	jsonLogger.Info("request procesado",
		"method", "POST",
		"path", "/api/users",
		"status", 201,
		"duration_ms", 42,
	)

	// --- With: logger con campos fijos ---
	fmt.Println("\n-- slog.With (campos fijos) --")
	reqLogger := textLogger.With("requestID", "req-abc-123", "traceID", "trace-xyz")
	reqLogger.Info("procesando request")
	reqLogger.Info("request completado", "status", 200)

	// --- Group: agrupar campos relacionados ---
	fmt.Println("\n-- slog.Group --")
	jsonLogger.Info("api call",
		slog.Group("request",
			slog.String("method", "GET"),
			slog.String("path", "/api/users"),
			slog.String("ip", "192.168.1.100"),
		),
		slog.Group("response",
			slog.Int("status", 200),
			slog.Int("bytes", 1024),
			slog.Duration("latency", 15*time.Millisecond),
		),
	)

	fmt.Println()
}

// =============================================================================
// filepath.Walk — recorrer directorios
// =============================================================================

func demoFilepathWalk() {
	fmt.Println("=== filepath.WalkDir ===")

	// Crear directorio temporal con estructura de ejemplo
	tmpDir, err := os.MkdirTemp("", "go-mastery-walk-*")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer os.RemoveAll(tmpDir)

	// Crear estructura de archivos
	dirs := []string{"src", "src/pkg", "docs", "build"}
	for _, d := range dirs {
		os.MkdirAll(filepath.Join(tmpDir, d), 0755)
	}
	files := map[string]string{
		"src/main.go":     "package main",
		"src/pkg/utils.go": "package pkg",
		"docs/readme.md":  "# Readme",
		"docs/guide.md":   "# Guide",
		"build/output.bin": "binary",
		"go.mod":          "module example",
	}
	for name, content := range files {
		os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644)
	}

	// WalkDir: recorrer y filtrar archivos .go
	fmt.Println("Archivos .go encontrados:")
	err = filepath.WalkDir(tmpDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // manejar error de acceso
		}
		// Saltar directorio "build"
		if d.IsDir() && d.Name() == "build" {
			return filepath.SkipDir
		}
		if !d.IsDir() && filepath.Ext(path) == ".go" {
			// Mostrar path relativo al tmpDir para legibilidad
			rel, _ := filepath.Rel(tmpDir, path)
			info, _ := d.Info()
			fmt.Printf("  %s (%d bytes)\n", rel, info.Size())
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error en WalkDir:", err)
	}

	// Contar archivos por extension
	fmt.Println("\nConteo por extension:")
	counts := make(map[string]int)
	filepath.WalkDir(tmpDir, func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			ext := filepath.Ext(path)
			if ext == "" {
				ext = "(sin extension)"
			}
			counts[ext]++
		}
		return nil
	})
	for ext, count := range counts {
		fmt.Printf("  %s: %d archivos\n", ext, count)
	}
	fmt.Println()
}

// =============================================================================
// time — parsing, formatting, duraciones, timers
// =============================================================================

func demoTime() {
	fmt.Println("=== time package ===")

	now := time.Now()
	fmt.Printf("Ahora: %v\n", now)

	// --- Formatting con el reference time de Go ---
	// Reference time: Mon Jan 2 15:04:05 MST 2006
	fmt.Println("\n-- Formatos de fecha --")
	fmt.Printf("  ISO 8601:    %s\n", now.Format("2006-01-02"))
	fmt.Printf("  Datetime:    %s\n", now.Format("2006-01-02 15:04:05"))
	fmt.Printf("  RFC3339:     %s\n", now.Format(time.RFC3339))
	fmt.Printf("  Espanol:     %s\n", now.Format("02/01/2006 15:04"))
	fmt.Printf("  Kitchen:     %s\n", now.Format(time.Kitchen))

	// --- Parsing ---
	fmt.Println("\n-- Parsing de fechas --")
	t1, err := time.Parse("2006-01-02", "2024-06-15")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("  Parsed: %v\n", t1)
	}

	t2, err := time.Parse(time.RFC3339, "2024-01-15T10:30:00Z")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("  Parsed RFC3339: %v\n", t2)
	}

	// --- Duraciones ---
	fmt.Println("\n-- Duraciones --")
	d1 := 2*time.Hour + 30*time.Minute + 15*time.Second
	fmt.Printf("  Duracion: %v\n", d1)
	fmt.Printf("  En segundos: %.0f\n", d1.Seconds())
	fmt.Printf("  En minutos: %.1f\n", d1.Minutes())

	d2, _ := time.ParseDuration("1h45m30s")
	fmt.Printf("  Parsed: %v\n", d2)

	// --- Operaciones con tiempo ---
	fmt.Println("\n-- Operaciones --")
	future := now.Add(72 * time.Hour)
	fmt.Printf("  En 72 horas: %s\n", future.Format("2006-01-02 15:04"))

	diff := future.Sub(now)
	fmt.Printf("  Diferencia: %v\n", diff)

	past := now.Add(-24 * time.Hour)
	fmt.Printf("  Hace 24h: %s\n", past.Format("2006-01-02 15:04"))
	fmt.Printf("  now.Before(future): %t\n", now.Before(future))
	fmt.Printf("  now.After(past): %t\n", now.After(past))

	// --- Timer (se dispara una vez) ---
	fmt.Println("\n-- Timer (una vez) --")
	timer := time.NewTimer(100 * time.Millisecond)
	start := time.Now()
	<-timer.C // esperar
	fmt.Printf("  Timer disparo despues de %v\n", time.Since(start))

	// --- Medir duracion de una operacion ---
	fmt.Println("\n-- Medir tiempo --")
	start = time.Now()
	// Simular trabajo
	sum := 0
	for i := 0; i < 1_000_000; i++ {
		sum += i
	}
	elapsed := time.Since(start)
	fmt.Printf("  Operacion tomo: %v (resultado: %d)\n\n", elapsed, sum)
}

// =============================================================================
// strings.Builder — construccion eficiente de strings
// =============================================================================

func demoStringsBuilder() {
	fmt.Println("=== strings.Builder ===")

	// Builder es mucho mas eficiente que concatenar con + en un loop
	var b strings.Builder

	// Escribir diferentes tipos de contenido
	b.WriteString("Usuarios activos:\n")
	users := []struct {
		Name  string
		Email string
	}{
		{"Alice", "alice@example.com"},
		{"Bob", "bob@example.com"},
		{"Charlie", "charlie@example.com"},
	}

	for i, u := range users {
		// fmt.Fprintf tambien funciona con Builder (implementa io.Writer)
		fmt.Fprintf(&b, "  %d. %s <%s>\n", i+1, u.Name, u.Email)
	}

	b.WriteString(fmt.Sprintf("Total: %d usuarios\n", len(users)))

	// Obtener el string final (solo una allocation)
	result := b.String()
	fmt.Println(result)

	// Comparar con el approach ineficiente (NO hacer esto en produccion):
	// s := ""
	// for _, u := range users {
	//     s += fmt.Sprintf("  %s <%s>\n", u.Name, u.Email) // nueva allocation cada vez!
	// }
}

// =============================================================================
// strings.Cut — dividir string por separador (Go 1.18+)
// =============================================================================

func demoStringsCut() {
	fmt.Println("=== strings.Cut (Go 1.18+) ===")

	// strings.Cut divide por la PRIMERA ocurrencia del separador
	// Mas limpio que strings.SplitN para el caso comun de 2 partes

	examples := []struct {
		input string
		sep   string
	}{
		{"host:8080", ":"},
		{"user@example.com", "@"},
		{"key=value=extra", "="},  // solo divide en la primera "="
		{"no-separator", ":"},     // found sera false
		{"Content-Type: application/json", ": "},
	}

	for _, ex := range examples {
		before, after, found := strings.Cut(ex.input, ex.sep)
		fmt.Printf("  Cut(%q, %q) -> before=%q, after=%q, found=%t\n",
			ex.input, ex.sep, before, after, found)
	}

	// Caso de uso practico: parsear headers HTTP
	fmt.Println("\n-- Caso practico: parsear headers --")
	rawHeaders := "Content-Type: application/json\nAuthorization: Bearer token123\nX-Custom: value"

	for _, line := range strings.Split(rawHeaders, "\n") {
		key, value, ok := strings.Cut(line, ": ")
		if ok {
			fmt.Printf("  Header[%q] = %q\n", key, value)
		}
	}

	// Otro caso: parsear KEY=VALUE de variables de entorno
	fmt.Println("\n-- Caso practico: parsear env vars --")
	envLines := []string{
		"DATABASE_URL=postgres://localhost:5432/mydb",
		"DEBUG=true",
		"EMPTY=",
		"INVALID_LINE",
	}
	for _, line := range envLines {
		key, value, found := strings.Cut(line, "=")
		if found {
			fmt.Printf("  %s = %q\n", key, value)
		} else {
			fmt.Printf("  (sin '='): %q\n", line)
		}
	}
}
