package backlog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRojoSubtestGuardaMotivo (D000033, E001/F002/S019): si una prueba falla solo en un subtest, el motivo trae esa línea.
// Rojo cuando: el fallo solo en el subtest deja el motivo vacío.
// Verde cuando: el motivo trae la línea del subtest y el resultado sigue siendo falla.
func TestRojoSubtestGuardaMotivo(t *testing.T) {
	raiz := t.TempDir()
	for rel, txt := range map[string]string{
		"go.mod": "module juguete\n\ngo 1.26\n",
		"p/p_test.go": "package p\n\nimport \"testing\"\n\n" +
			"func TestPadre(t *testing.T) {\n\tt.Run(\"pasa\", func(t *testing.T) {})\n" +
			"\tt.Run(\"sub\", func(t *testing.T) { t.Run(\"hondo\", func(t *testing.T) { t.Fatal(\"falla el nieto\") }) })\n}\n",
	} {
		ruta := filepath.Join(raiz, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(ruta, []byte(txt), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	rs, err := correr(raiz, []Prueba{{Nombre: "./p:TestPadre", Intencion: "i"}})
	if err != nil {
		t.Fatal(err)
	}
	r := rs["./p:TestPadre"]
	if r.accion != "fail" {
		t.Errorf("el padre debió fallar; dio %q (un subtest que pasa no cambia eso)", r.accion)
	}
	if !strings.Contains(r.salida, "falla el nieto") {
		t.Fatalf("el motivo debió traer la línea del subtest Padre/sub/hondo; dio %q", r.salida)
	}
}
