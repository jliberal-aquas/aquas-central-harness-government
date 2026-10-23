package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestHijoInstaladoPasaSuGate: integral de E002/F001. Un hijo recién instalado en carpeta vacía
// arma su CLAUDE.md y pasa su propio gate sin errores.
// Rojo cuando: install, build (con los huecos llenos) o verificar dentro del hijo sale con código distinto de 0.
// Verde cuando: los tres pasos salen con 0 y verificar corre en la carpeta del hijo.
func TestHijoInstaladoPasaSuGate(t *testing.T) {
	base := t.TempDir()
	hijo := filepath.Join(base, "hijo")
	correr := func(paso ...string) {
		t.Helper()
		if s := aquas(t, base, "", paso...); s.codigo != 0 {
			t.Fatalf("aquas %v: exit %d; quiero 0\n%s%s", paso, s.codigo, s.out, s.err)
		}
	}
	correr("install", hijo)
	// los huecos son lo único que el hijo escribe a mano (E002/F001/S002)
	huecos := filepath.Join(hijo, "gobierno", "canonico", "huecos")
	if err := os.MkdirAll(huecos, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, h := range []string{"titulo", "mision", "mapa"} {
		if err := os.WriteFile(filepath.Join(huecos, h+".md"), []byte("Hijo de prueba: "+h+".\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	correr("build", hijo)
	if s := aquas(t, hijo, "", "verificar"); s.codigo != 0 {
		t.Fatalf("aquas verificar en el hijo: exit %d; quiero 0\n%s%s", s.codigo, s.out, s.err)
	}
}
