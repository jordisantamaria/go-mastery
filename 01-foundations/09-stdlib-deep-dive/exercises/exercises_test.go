package exercises

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

// =============================================================================
// Test 1: CountWords
// =============================================================================

func TestCountWords(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  map[string]int
	}{
		{
			name:  "palabras simples",
			input: "hello world hello",
			want:  map[string]int{"hello": 2, "world": 1},
		},
		{
			name:  "case insensitive",
			input: "Go go GO gO",
			want:  map[string]int{"go": 4},
		},
		{
			name:  "con newlines y tabs",
			input: "one\ttwo\nthree\none",
			want:  map[string]int{"one": 2, "two": 1, "three": 1},
		},
		{
			name:  "string vacio",
			input: "",
			want:  map[string]int{},
		},
		{
			name:  "una palabra",
			input: "solo",
			want:  map[string]int{"solo": 1},
		},
		{
			name:  "multiples espacios",
			input: "  hello   world  ",
			want:  map[string]int{"hello": 1, "world": 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CountWords(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("CountWords() error = %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("CountWords() returned %d entries, want %d\ngot:  %v\nwant: %v",
					len(got), len(tt.want), got, tt.want)
			}
			for word, wantCount := range tt.want {
				if gotCount, ok := got[word]; !ok {
					t.Errorf("palabra %q no encontrada en resultado", word)
				} else if gotCount != wantCount {
					t.Errorf("CountWords()[%q] = %d, want %d", word, gotCount, wantCount)
				}
			}
		})
	}
}

// =============================================================================
// Test 2: LineCount
// =============================================================================

func TestLineCount(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"dos lineas con newline final", "hello\nworld\n", 2},
		{"dos lineas sin newline final", "hello\nworld", 2},
		{"una linea", "hello", 1},
		{"vacio", "", 0},
		{"solo newlines", "\n\n\n", 3},
		{"linea vacia al final", "a\nb\n\n", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LineCount(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("LineCount() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("LineCount(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

// =============================================================================
// Test 3: ToJSON
// =============================================================================

func TestToJSON(t *testing.T) {
	t.Run("map simple", func(t *testing.T) {
		input := map[string]int{"x": 1, "y": 2}
		got, err := ToJSON(input)
		if err != nil {
			t.Fatalf("ToJSON() error = %v", err)
		}
		// Verificar que es JSON valido
		var decoded map[string]int
		if err := json.Unmarshal([]byte(got), &decoded); err != nil {
			t.Fatalf("resultado no es JSON valido: %v\ngot: %s", err, got)
		}
		if decoded["x"] != 1 || decoded["y"] != 2 {
			t.Errorf("valores incorrectos: %v", decoded)
		}
		// Verificar que tiene indentacion
		if !strings.Contains(got, "\n") {
			t.Error("ToJSON deberia producir JSON indentado (con newlines)")
		}
	})

	t.Run("struct", func(t *testing.T) {
		type Point struct {
			X int `json:"x"`
			Y int `json:"y"`
		}
		got, err := ToJSON(Point{X: 10, Y: 20})
		if err != nil {
			t.Fatalf("ToJSON() error = %v", err)
		}
		if !strings.Contains(got, `"x": 10`) {
			t.Errorf("ToJSON(Point) no contiene expected field:\n%s", got)
		}
	})

	t.Run("slice", func(t *testing.T) {
		got, err := ToJSON([]string{"a", "b"})
		if err != nil {
			t.Fatalf("ToJSON() error = %v", err)
		}
		var decoded []string
		if err := json.Unmarshal([]byte(got), &decoded); err != nil {
			t.Fatalf("resultado no es JSON valido: %v", err)
		}
		if len(decoded) != 2 || decoded[0] != "a" || decoded[1] != "b" {
			t.Errorf("decoded = %v, want [a b]", decoded)
		}
	})
}

// =============================================================================
// Test 4: FromJSON
// =============================================================================

func TestFromJSON(t *testing.T) {
	t.Run("map", func(t *testing.T) {
		got, err := FromJSON[map[string]int](`{"a": 1, "b": 2}`)
		if err != nil {
			t.Fatalf("FromJSON() error = %v", err)
		}
		if got["a"] != 1 || got["b"] != 2 {
			t.Errorf("FromJSON() = %v, want map[a:1 b:2]", got)
		}
	})

	t.Run("slice", func(t *testing.T) {
		got, err := FromJSON[[]string](`["hello", "world"]`)
		if err != nil {
			t.Fatalf("FromJSON() error = %v", err)
		}
		if len(got) != 2 || got[0] != "hello" || got[1] != "world" {
			t.Errorf("FromJSON() = %v, want [hello world]", got)
		}
	})

	t.Run("int", func(t *testing.T) {
		got, err := FromJSON[int](`42`)
		if err != nil {
			t.Fatalf("FromJSON() error = %v", err)
		}
		if got != 42 {
			t.Errorf("FromJSON() = %d, want 42", got)
		}
	})

	t.Run("error en JSON invalido", func(t *testing.T) {
		_, err := FromJSON[int](`"not a number"`)
		if err == nil {
			t.Error("FromJSON() deberia devolver error para tipo incompatible")
		}
	})

	t.Run("error en sintaxis", func(t *testing.T) {
		_, err := FromJSON[map[string]int](`{invalid}`)
		if err == nil {
			t.Error("FromJSON() deberia devolver error para JSON invalido")
		}
	})
}

// =============================================================================
// Test 5: CopyFile
// =============================================================================

func TestCopyFile(t *testing.T) {
	t.Run("copia exitosa", func(t *testing.T) {
		tmpDir := t.TempDir()
		srcPath := filepath.Join(tmpDir, "source.txt")
		dstPath := filepath.Join(tmpDir, "dest.txt")

		content := "contenido del archivo para copiar\nsegunda linea\n"
		os.WriteFile(srcPath, []byte(content), 0644)

		n, err := CopyFile(srcPath, dstPath)
		if err != nil {
			t.Fatalf("CopyFile() error = %v", err)
		}
		if n != int64(len(content)) {
			t.Errorf("CopyFile() = %d bytes, want %d", n, len(content))
		}

		// Verificar contenido
		got, err := os.ReadFile(dstPath)
		if err != nil {
			t.Fatalf("error leyendo destino: %v", err)
		}
		if string(got) != content {
			t.Errorf("contenido copiado = %q, want %q", string(got), content)
		}
	})

	t.Run("source no existe", func(t *testing.T) {
		tmpDir := t.TempDir()
		_, err := CopyFile(
			filepath.Join(tmpDir, "noexiste.txt"),
			filepath.Join(tmpDir, "dest.txt"),
		)
		if err == nil {
			t.Error("CopyFile() deberia devolver error si source no existe")
		}
	})

	t.Run("archivo vacio", func(t *testing.T) {
		tmpDir := t.TempDir()
		srcPath := filepath.Join(tmpDir, "empty.txt")
		dstPath := filepath.Join(tmpDir, "dest.txt")

		os.WriteFile(srcPath, []byte{}, 0644)

		n, err := CopyFile(srcPath, dstPath)
		if err != nil {
			t.Fatalf("CopyFile() error = %v", err)
		}
		if n != 0 {
			t.Errorf("CopyFile(empty) = %d bytes, want 0", n)
		}
	})
}

// =============================================================================
// Test 6: FindFiles
// =============================================================================

func TestFindFiles(t *testing.T) {
	t.Run("encontrar archivos .txt", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Crear estructura
		os.MkdirAll(filepath.Join(tmpDir, "sub"), 0755)
		os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("a"), 0644)
		os.WriteFile(filepath.Join(tmpDir, "b.go"), []byte("b"), 0644)
		os.WriteFile(filepath.Join(tmpDir, "sub", "c.txt"), []byte("c"), 0644)
		os.WriteFile(filepath.Join(tmpDir, "sub", "d.md"), []byte("d"), 0644)

		got, err := FindFiles(tmpDir, ".txt")
		if err != nil {
			t.Fatalf("FindFiles() error = %v", err)
		}

		sort.Strings(got)
		want := []string{"a.txt", filepath.Join("sub", "c.txt")}
		sort.Strings(want)

		if len(got) != len(want) {
			t.Fatalf("FindFiles() returned %d files, want %d\ngot:  %v\nwant: %v",
				len(got), len(want), got, want)
		}
		for i, g := range got {
			if g != want[i] {
				t.Errorf("FindFiles()[%d] = %q, want %q", i, g, want[i])
			}
		}
	})

	t.Run("sin coincidencias", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.WriteFile(filepath.Join(tmpDir, "a.go"), []byte("a"), 0644)

		got, err := FindFiles(tmpDir, ".txt")
		if err != nil {
			t.Fatalf("FindFiles() error = %v", err)
		}
		if len(got) != 0 {
			t.Errorf("FindFiles() returned %v, want empty", got)
		}
	})

	t.Run("directorio vacio", func(t *testing.T) {
		tmpDir := t.TempDir()

		got, err := FindFiles(tmpDir, ".go")
		if err != nil {
			t.Fatalf("FindFiles() error = %v", err)
		}
		if len(got) != 0 {
			t.Errorf("FindFiles(empty dir) returned %v, want empty", got)
		}
	})
}

// =============================================================================
// Test 7: HTTPGet
// =============================================================================

func TestHTTPGet(t *testing.T) {
	t.Run("request exitoso", func(t *testing.T) {
		// Crear servidor de prueba con httptest
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("metodo = %s, want GET", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"message":"hello"}`)
		}))
		defer server.Close()

		body, err := HTTPGet(server.URL, 5*time.Second)
		if err != nil {
			t.Fatalf("HTTPGet() error = %v", err)
		}

		if !strings.Contains(string(body), "hello") {
			t.Errorf("HTTPGet() body = %q, want contener 'hello'", string(body))
		}
	})

	t.Run("status no-200", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		_, err := HTTPGet(server.URL, 5*time.Second)
		if err == nil {
			t.Error("HTTPGet() deberia devolver error para status 404")
		}
		if !strings.Contains(err.Error(), "404") {
			t.Errorf("error deberia contener '404', got: %v", err)
		}
	})

	t.Run("timeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(500 * time.Millisecond) // simular server lento
			fmt.Fprint(w, "too late")
		}))
		defer server.Close()

		_, err := HTTPGet(server.URL, 50*time.Millisecond) // timeout muy corto
		if err == nil {
			t.Error("HTTPGet() deberia devolver error por timeout")
		}
	})
}

// =============================================================================
// Test 8: MultiMerge
// =============================================================================

func TestMultiMerge(t *testing.T) {
	t.Run("merge multiples readers", func(t *testing.T) {
		r1 := strings.NewReader("hello ")
		r2 := strings.NewReader("world ")
		r3 := strings.NewReader("!")

		merged := MultiMerge(r1, r2, r3)

		data, err := io.ReadAll(merged)
		if err != nil {
			t.Fatalf("ReadAll() error = %v", err)
		}

		got := string(data)
		want := "hello world !"
		if got != want {
			t.Errorf("MultiMerge() = %q, want %q", got, want)
		}
	})

	t.Run("un solo reader", func(t *testing.T) {
		r := strings.NewReader("solo")
		merged := MultiMerge(r)

		data, err := io.ReadAll(merged)
		if err != nil {
			t.Fatalf("ReadAll() error = %v", err)
		}

		if string(data) != "solo" {
			t.Errorf("MultiMerge(solo) = %q, want 'solo'", string(data))
		}
	})

	t.Run("sin readers", func(t *testing.T) {
		merged := MultiMerge()

		data, err := io.ReadAll(merged)
		if err != nil {
			t.Fatalf("ReadAll() error = %v", err)
		}

		if len(data) != 0 {
			t.Errorf("MultiMerge() con 0 readers = %q, want vacio", string(data))
		}
	})

	t.Run("readers con contenido largo", func(t *testing.T) {
		part1 := strings.Repeat("a", 1000)
		part2 := strings.Repeat("b", 2000)

		merged := MultiMerge(
			strings.NewReader(part1),
			strings.NewReader(part2),
		)

		data, err := io.ReadAll(merged)
		if err != nil {
			t.Fatalf("ReadAll() error = %v", err)
		}

		if len(data) != 3000 {
			t.Errorf("MultiMerge() largo = %d bytes, want 3000", len(data))
		}
	})
}
