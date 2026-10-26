package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/hijo"
)

// TestAvisoEnElArranque (E002/F002/S003): el arranque avisa en los comandos del hijo, menos install y update.
// Rojo cuando: calla en un comando del hijo, avisa en install o update, o el aviso cambia el código de salida.
// Verde cuando: avisa en verificar, calla en install y update, y la salida es la misma con y sin lock.
func TestAvisoEnElArranque(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, hijo.Carpeta), 0o755); err != nil {
		t.Fatal(err)
	}
	lock := `{"version":"v0.0.0-20261025173306-1a31e1bd98fa"}`
	if err := os.WriteFile(filepath.Join(dir, hijo.Carpeta, "lock.json"), []byte(lock), 0o644); err != nil {
		t.Fatal(err)
	}
	nueva := "v0.0.0-20261026002129-e5f9d7c6e1a0"
	for _, c := range []struct {
		comando string
		avisa   bool
	}{{"verificar", true}, {"backlog", true}, {"install", false}, {"update", false}} {
		var w bytes.Buffer
		avisarVersion(&w, c.comando, dir, nueva)
		if got := strings.Contains(w.String(), "corre aquas update"); got != c.avisa {
			t.Errorf("%s: avisa=%v; quiero %v (%q)", c.comando, got, c.avisa, w.String())
		}
	}

	// Cable y criterio 4 por el arranque real (D000043): el binario de prueba es devel,
	// así que AQUAS_VERSION fija una versión con fecha para que el aviso compare.
	t.Setenv("AQUAS_VERSION", nueva)
	sinLock := t.TempDir()
	con, sin := aquas(t, dir, "", "limpiar"), aquas(t, sinLock, "", "limpiar")
	if !strings.Contains(con.err, "corre aquas update") {
		t.Errorf("el arranque no avisó en el hijo con lock viejo; stderr %q", con.err)
	}
	if strings.Contains(sin.err, "aviso:") {
		t.Errorf("sin lock no debió avisar; stderr %q", sin.err)
	}
	if con.codigo != sin.codigo {
		t.Errorf("el aviso cambió el código de salida: con %d, sin %d", con.codigo, sin.codigo)
	}
}
