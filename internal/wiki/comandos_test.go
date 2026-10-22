package wiki

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestComandoCitadoExiste: un comando citado que el CLI no tiene hace caer el gate;
// los que sí existen, o los que no van entre comillas invertidas, pasan.
func TestComandoCitadoExiste(t *testing.T) {
	raiz := t.TempDir()
	poner := func(rel, txt string) {
		t.Helper()
		p := filepath.Join(raiz, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(txt), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	poner("cmd/aquas/main.go", "package main\n\nfunc main() {\n\tswitch x {\n\tcase \"turno\":\n\t\tif y == \"abrir\" {}\n\t}\n}\n")
	poner("internal/otro/otro.go", "package otro\n\nvar v = \"cerrar\"\n")
	poner("docs/a.md", "Abre con `aquas turno abrir T1`, cierra con `aquas turno cerrar`.\n"+
		"Viejo: `aquas turno refinar x=1`.\n")
	poner("gobierno/canonico/metodologias/m.md", "Cancelado: `aquas ver E001`. En construcción: aquas install, sin comillas.\n")

	ps, err := RevisarComandos(raiz)
	if err != nil {
		t.Fatal(err)
	}
	todo := strings.Join(ps, "\n")
	if len(ps) != 2 || !strings.Contains(todo, "docs/a.md") || !strings.Contains(todo, "aquas turno refinar") ||
		!strings.Contains(todo, "gobierno/canonico/metodologias/m.md") || !strings.Contains(todo, "aquas ver") {
		t.Fatalf("quiero 2 problemas (turno refinar en docs/a.md, ver en m.md); obtuve %d:\n%s", len(ps), todo)
	}
}
