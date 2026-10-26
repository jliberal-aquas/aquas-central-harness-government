package guardia

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// E001/F002/S007: la guardia aplica el alcance del slice en curso a Write/Edit y a Bash.
func TestGuardiaBloqueaFueraDelAlcance(t *testing.T) {
	raiz := raizPrueba(t)
	dir := filepath.Join(raiz, "trabajo", "activo")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	ficha := `{"id": "E001/F002/S007", "tipo": "slice", "titulo": "A", "estado": "EN_CURSO", "prioridad": "P1", "motivo": "m",
  "padre": "E001/F002", "bloqueado_por": [], "secuencia": 1, "alcance": ["internal/a/**"]}`
	if err := os.WriteFile(filepath.Join(dir, "E001-F002-S007.json"), []byte(ficha), 0o644); err != nil {
		t.Fatal(err)
	}
	if m := RevisarArchivo(raiz, "docs/y.md"); !strings.Contains(m, "fuera del alcance de E001/F002/S007") {
		t.Errorf("Write fuera del alcance: bloquea; motivo %q", m)
	}
	bash := `{"tool_name":"Bash","tool_input":{"command":"echo x > docs/y.md"}}`
	if m := Revisar(raiz, []byte(bash)); !strings.Contains(m, "fuera del alcance") {
		t.Errorf("Bash fuera del alcance: bloquea; motivo %q", m)
	}
	if m := RevisarArchivo(raiz, "internal/a/x.go"); m != "" {
		t.Errorf("dentro del alcance pasa; motivo %q", m)
	}
}
