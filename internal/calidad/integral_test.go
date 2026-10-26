package calidad

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestCanarioFallaYRepoPasa (E001/F013, integral): el gate de calidad detecta código malo y deja pasar el limpio.
// Rojo cuando: algún linter bloqueante no salta en el canario, o el repo limpio tiene hallazgos.
// Verde cuando: el canario dispara los seis linters y el repo sale limpio.
func TestCanarioFallaYRepoPasa(t *testing.T) {
	raiz, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	config := filepath.Join(raiz, "gobierno", "config", "golangci.yml")
	h, err := Revisar(filepath.Join(raiz, "internal", "calidad", "testdata", "canario"), config)
	if err != nil {
		t.Fatalf("el canario debió revisarse: %v", err)
	}
	todo := strings.Join(h, "\n")
	for _, l := range []string{"errcheck", "gofmt", "unused", "ineffassign", "govet", "staticcheck"} {
		if !strings.Contains(todo, "("+l+")") {
			t.Errorf("el canario no disparó %s; hallazgos:\n%s", l, todo)
		}
	}
	if h, err := Revisar(raiz, config); err != nil || len(h) > 0 {
		t.Errorf("el repo debió salir limpio: err %v, %d hallazgos", err, len(h))
	}
}
