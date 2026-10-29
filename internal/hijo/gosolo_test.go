package hijo

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"
)

// TestGoSoloEnCentral (E002/F002/S005): la regla de Go como lenguaje único gobierna solo a la central.
// Rojo cuando: el CLAUDE.md del hijo, armado con la constitución real, contiene «lenguaje único».
// Verde cuando: el hijo no la trae, la central la conserva y sus principios van seguidos desde 1.
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
	// E001/F012/S001: la cuenta crece con la constitución; lo que se exige es que vayan seguidos desde 1 y que hijo y central coincidan.
	seguidos := func(got string) bool {
		ns := strings.Split(got, ",")
		for i, n := range ns {
			if n != strconv.Itoa(i+1) {
				return false
			}
		}
		return len(ns) >= 8
	}
	for nombre, txt := range map[string]string{"central": string(central), "hijo": string(hijo)} {
		if got := numeros(txt); !seguidos(got) {
			t.Errorf("%s: los principios deben ir seguidos desde 1 (al menos 8); van %s", nombre, got)
		}
	}
	if c, h := numeros(string(central)), numeros(string(hijo)); c != h {
		t.Errorf("hijo y central deben tener los mismos principios; central %s, hijo %s", c, h)
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
