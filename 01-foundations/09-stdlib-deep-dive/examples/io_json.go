// Ejemplo: I/O y JSON — lectura/escritura de archivos, composicion de Readers,
// JSON con struct tags, custom MarshalJSON, Encoder/Decoder, MultiReader, TeeReader.
//
// Ejecutar: go run ./01-foundations/09-stdlib-deep-dive/examples/io_json.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// =============================================================================
// Modelo con struct tags y custom MarshalJSON
// =============================================================================

// Priority es un tipo personalizado que se serializa como string en JSON.
type Priority int

const (
	PriorityLow    Priority = iota // 0
	PriorityMedium                 // 1
	PriorityHigh                   // 2
)

// MarshalJSON convierte Priority a su representacion string en JSON.
// En vez de serializar como numero (0, 1, 2), se serializa como "low", "medium", "high".
func (p Priority) MarshalJSON() ([]byte, error) {
	var s string
	switch p {
	case PriorityLow:
		s = "low"
	case PriorityMedium:
		s = "medium"
	case PriorityHigh:
		s = "high"
	default:
		s = "unknown"
	}
	return json.Marshal(s)
}

// UnmarshalJSON convierte el string JSON de vuelta a Priority.
func (p *Priority) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	switch s {
	case "low":
		*p = PriorityLow
	case "medium":
		*p = PriorityMedium
	case "high":
		*p = PriorityHigh
	default:
		return fmt.Errorf("prioridad desconocida: %s", s)
	}
	return nil
}

// Task demuestra struct tags de JSON:
// - `json:"name"` renombra el campo
// - `json:"done,omitempty"` omite el campo si es false (zero value)
// - `json:"-"` ignora el campo completamente
type Task struct {
	ID       int      `json:"id"`
	Title    string   `json:"title"`
	Done     bool     `json:"done,omitempty"`     // omitempty: omite si false
	Priority Priority `json:"priority"`           // usa custom MarshalJSON
	Notes    string   `json:"notes,omitempty"`    // omitempty: omite si string vacio
	Internal string   `json:"-"`                  // siempre ignorado en JSON
}

func main() {
	demoJSONBasics()
	demoFileIO()
	demoEncoderDecoder()
	demoMultiReader()
	demoTeeReader()
	demoCopyBetweenFiles()
}

// =============================================================================
// JSON basico: Marshal, Unmarshal, MarshalIndent
// =============================================================================

func demoJSONBasics() {
	fmt.Println("=== JSON Basico ===")

	tasks := []Task{
		{ID: 1, Title: "Aprender Go", Done: true, Priority: PriorityHigh, Notes: "stdlib first", Internal: "secreto"},
		{ID: 2, Title: "Escribir tests", Done: false, Priority: PriorityMedium, Internal: "otro secreto"},
		{ID: 3, Title: "Deploy", Priority: PriorityLow},
	}

	// Marshal con indentacion (pretty print)
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		fmt.Println("Error marshal:", err)
		return
	}
	fmt.Println("JSON generado:")
	fmt.Println(string(data))
	// Observar:
	// - Priority se serializa como "high", "medium", "low" (custom MarshalJSON)
	// - "done" se omite cuando es false (omitempty)
	// - "notes" se omite cuando es string vacio (omitempty)
	// - "Internal" nunca aparece (json:"-")

	// Unmarshal de vuelta
	var decoded []Task
	if err := json.Unmarshal(data, &decoded); err != nil {
		fmt.Println("Error unmarshal:", err)
		return
	}
	fmt.Printf("\nDecodificado: %d tareas, primera = %+v\n\n", len(decoded), decoded[0])
}

// =============================================================================
// File I/O: os.Create, os.Open, lectura/escritura
// =============================================================================

func demoFileIO() {
	fmt.Println("=== File I/O ===")

	// Crear directorio temporal para no ensuciar el proyecto
	tmpDir, err := os.MkdirTemp("", "go-mastery-io-*")
	if err != nil {
		fmt.Println("Error creando tmpdir:", err)
		return
	}
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "saludo.txt")

	// Escribir archivo con os.Create + WriteString
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creando archivo:", err)
		return
	}
	file.WriteString("Hola desde Go!\n")
	file.WriteString("Linea dos.\n")
	file.WriteString("Linea tres.\n")
	file.Close() // cerramos manualmente aqui para poder leerlo despues

	// Leer archivo completo con os.ReadFile (la forma mas simple)
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error leyendo:", err)
		return
	}
	fmt.Printf("Contenido del archivo:\n%s\n", content)

	// Leer archivo con os.Open + io.ReadAll (mas control)
	file2, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error abriendo:", err)
		return
	}
	defer file2.Close()

	data, err := io.ReadAll(file2)
	if err != nil {
		fmt.Println("Error leyendo:", err)
		return
	}
	fmt.Printf("Leido con io.ReadAll: %d bytes\n\n", len(data))
}

// =============================================================================
// json.Encoder y json.Decoder — streaming JSON a/desde archivos
// =============================================================================

func demoEncoderDecoder() {
	fmt.Println("=== JSON Encoder/Decoder (streaming a archivo) ===")

	tmpDir, err := os.MkdirTemp("", "go-mastery-json-*")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer os.RemoveAll(tmpDir)

	jsonPath := filepath.Join(tmpDir, "tasks.json")

	// --- Encoder: escribir JSON directamente a un archivo ---
	tasks := []Task{
		{ID: 1, Title: "Leer docs", Priority: PriorityHigh, Done: true},
		{ID: 2, Title: "Practicar", Priority: PriorityMedium},
	}

	outFile, err := os.Create(jsonPath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	encoder := json.NewEncoder(outFile)
	encoder.SetIndent("", "  ") // indentacion bonita
	if err := encoder.Encode(tasks); err != nil {
		fmt.Println("Error encoding:", err)
	}
	outFile.Close()

	fmt.Println("Archivo JSON escrito:", jsonPath)

	// --- Decoder: leer JSON directamente de un archivo ---
	inFile, err := os.Open(jsonPath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer inFile.Close()

	var loaded []Task
	decoder := json.NewDecoder(inFile)
	if err := decoder.Decode(&loaded); err != nil {
		fmt.Println("Error decoding:", err)
		return
	}

	fmt.Printf("Tareas cargadas del archivo: %d\n", len(loaded))
	for _, t := range loaded {
		fmt.Printf("  [%d] %s (priority=%d, done=%t)\n", t.ID, t.Title, t.Priority, t.Done)
	}
	fmt.Println()
}

// =============================================================================
// io.MultiReader — concatenar multiples Readers
// =============================================================================

func demoMultiReader() {
	fmt.Println("=== io.MultiReader ===")

	// Combinar 3 readers en uno solo
	header := strings.NewReader("=== INICIO DEL REPORTE ===\n")
	body := strings.NewReader("Datos del reporte...\nLinea 1\nLinea 2\n")
	footer := strings.NewReader("=== FIN DEL REPORTE ===\n")

	// MultiReader los concatena secuencialmente
	combined := io.MultiReader(header, body, footer)

	// Copiar todo a stdout
	fmt.Println("Output de MultiReader:")
	io.Copy(os.Stdout, combined)
	fmt.Println()
}

// =============================================================================
// io.TeeReader — leer y copiar simultaneamente
// =============================================================================

func demoTeeReader() {
	fmt.Println("=== io.TeeReader ===")

	// Simular un body de HTTP response
	original := strings.NewReader("datos importantes del response body")

	// TeeReader: al leer de tee, los bytes tambien se escriben a buf
	var buf bytes.Buffer
	tee := io.TeeReader(original, &buf)

	// Leer del tee (esto tambien llena buf)
	content, err := io.ReadAll(tee)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Leido directamente: %q\n", string(content))
	fmt.Printf("Copia en buffer:    %q\n", buf.String())
	fmt.Printf("Son iguales: %t\n\n", string(content) == buf.String())
}

// =============================================================================
// io.Copy entre archivos
// =============================================================================

func demoCopyBetweenFiles() {
	fmt.Println("=== io.Copy entre archivos ===")

	tmpDir, err := os.MkdirTemp("", "go-mastery-copy-*")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer os.RemoveAll(tmpDir)

	// Crear archivo fuente
	srcPath := filepath.Join(tmpDir, "original.txt")
	os.WriteFile(srcPath, []byte("Contenido original para copiar.\nSegunda linea.\n"), 0644)

	dstPath := filepath.Join(tmpDir, "copia.txt")

	// Abrir fuente y destino
	src, err := os.Open(srcPath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer src.Close()

	dst, err := os.Create(dstPath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer dst.Close()

	// io.Copy: streaming eficiente sin cargar todo en memoria
	n, err := io.Copy(dst, src)
	if err != nil {
		fmt.Println("Error copiando:", err)
		return
	}

	fmt.Printf("Copiados %d bytes de %s a %s\n", n, filepath.Base(srcPath), filepath.Base(dstPath))

	// Verificar
	copied, _ := os.ReadFile(dstPath)
	fmt.Printf("Contenido de la copia: %q\n", string(copied))
}
