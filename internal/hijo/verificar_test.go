package hijo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestVerificarEnHijo: el gate de un hijo revisa su central sellada y su CLAUDE.md armado (E002/F001/S002/X02).
// Rojo cuando: un hijo recién armado no pasa, o pasa con su CLAUDE.md editado a mano o su central alterada.
// Verde cuando: recién armado no da hallazgos; CLAUDE.md desfasado pide aquas build; central alterada se nombra.
func TestVerificarEnHijo(t *testing.T) {
	dir := instalarConHuecos(t, map[string]string{"mision": "- Landing de Aquas.\n"})
	if err := Build(dir); err != nil {
		t.Fatal(err)
	}
	if m := Verificar(dir); len(m) != 0 {
		t.Fatalf("hijo recién armado: debió pasar; dio %v", m)
	}
	claude := filepath.Join(dir, "CLAUDE.md")
	if err := os.WriteFile(claude, []byte("editado a mano\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if m := strings.Join(Verificar(dir), "\n"); !strings.Contains(m, "aquas build") {
		t.Errorf("CLAUDE.md editado a mano: debió pedir aquas build; dio %q", m)
	}
	if err := Build(dir); err != nil {
		t.Fatal(err)
	}
	sellado := filepath.Join(dir, Carpeta, "central", "gobierno", "canonico", "constitucion.md")
	_ = os.Chmod(sellado, 0o644)
	if err := os.WriteFile(sellado, []byte("alterada\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if m := strings.Join(Verificar(dir), "\n"); !strings.Contains(m, "constitucion.md") {
		t.Errorf("central alterada: debió nombrar el archivo; dio %q", m)
	}
}
