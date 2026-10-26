package hijo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestAvisoDeVersion (E002/F002/S003): el hijo avisa cuando su lock y el ejecutable difieren en versión.
// Rojo cuando: calla con versiones distintas, avisa con versiones iguales o con devel, o confunde cuál quedó atrás.
// Verde cuando: lock viejo pide aquas update, ejecutable viejo pide go install, y versiones iguales o devel callan.
func TestAvisoDeVersion(t *testing.T) {
	const (
		vieja = "v0.0.0-20261025173306-1a31e1bd98fa"
		nueva = "v0.0.0-20261026002129-e5f9d7c6e1a0"
	)
	hijoCon := func(lock string) string {
		dir := t.TempDir()
		if lock == "" {
			return dir
		}
		if err := os.MkdirAll(filepath.Join(dir, Carpeta), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, Carpeta, "lock.json"), []byte(lock), 0o644); err != nil {
			t.Fatal(err)
		}
		return dir
	}
	conVersion := func(v string) string { return hijoCon(`{"version":"` + v + `"}`) }

	casos := []struct {
		nombre, dir, binario, quiero string
	}{
		{"criterio 1: lock viejo", conVersion(vieja), nueva, "corre aquas update"},
		{"criterio 2: ejecutable viejo", conVersion(nueva), vieja, "actualiza aquas con go install"},
		{"criterio 3: versiones iguales", conVersion(nueva), nueva, ""},
		{"criterio 3: ejecutable devel", conVersion(vieja), "devel", ""},
		{"borde: +dirty compara solo la fecha", conVersion(vieja), nueva + "+dirty", "corre aquas update"},
		{"borde: misma fecha con +dirty", conVersion(nueva), nueva + "+dirty", ""},
		{"borde: carpeta sin lock", hijoCon(""), nueva, ""},
		{"borde: lock ilegible", hijoCon("{no es json"), nueva, ""},
		{"borde: versión con etiqueta", conVersion("v1.2.0"), nueva, ""},
		{"borde: pseudo-versión tras etiqueta", conVersion("v1.2.1-0.20261025173306-1a31e1bd98fa"), nueva, "corre aquas update"},
	}
	for _, c := range casos {
		got := AvisoVersion(c.dir, c.binario)
		switch {
		case c.quiero == "" && got != "":
			t.Errorf("%s: no debió avisar; avisó %q", c.nombre, got)
		case c.quiero != "" && !strings.Contains(got, c.quiero):
			t.Errorf("%s: quiero aviso con %q; obtuve %q", c.nombre, c.quiero, got)
		}
	}
}
