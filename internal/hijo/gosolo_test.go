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
	for nombre, txt := range map[string]string{"central": string(central), "hijo": string(hijo)} {
		if got := numeros(txt); got != "1,2,3,4,5,6,7,8" {
			t.Errorf("%s: los principios deben ir del 1 al 8 seguidos; van %s", nombre, got)
		}
	}
}

// numeros devuelve los números de la lista de «## Principios», en orden.
func numeros(txt string) string {
	_, resto, _ := strings.Cut(txt, "## Principios\n")
	resto, _, _ = strings.Cut(resto, "\n## ")
	var ns []string
	for _, l := range strings.Split(resto, "\n") {
		if n, _, ok := strings.Cut(l, ". "); ok && n != "" && strings.Trim(n, "0123456789") == "" {
			ns = append(ns, n)
		}
	}
	return strings.Join(ns, ",")
}
