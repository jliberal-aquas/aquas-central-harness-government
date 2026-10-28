package hijo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

// TestGoSoloEnCentral (E002/F002/S005): la regla de Go como lenguaje único gobierna solo a la central.
// Rojo cuando: el CLAUDE.md del hijo, armado con la constitución real, contiene «lenguaje único».
// Verde cuando: el hijo no la trae, la central la conserva y sus principios van del 1 al 8.
func TestGoSoloEnCentral(t *testing.T) {
	raiz := filepath.Join("..", "..")
	b, err := os.ReadFile(filepath.Join(raiz, "gobierno", "canonico", "constitucion.md"))
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(t.TempDir(), "hijo")
	if err := Install(dir, fstest.MapFS{"gobierno/canonico/constitucion.md": {Data: b}}, "v1.0.0"); err != nil {
		t.Fatal(err)
	}
	if err := Listo(dir); err != nil {
		t.Fatal(err)
	}
	hijo, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(hijo), "lenguaje único") {
		t.Error("el CLAUDE.md del hijo no debe traer la regla de Go como lenguaje único")
	}
	central, err := os.ReadFile(filepath.Join(raiz, "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(central), "lenguaje único") {
		t.Error("el CLAUDE.md de la central debe conservar la regla de Go")
	}
	for _, txt := range []string{string(central), string(hijo)} {
		if !strings.Contains(txt, "\n8. ") || strings.Contains(txt, "\n9. ") {
			t.Error("los principios deben ir del 1 al 8")
		}
	}
}
