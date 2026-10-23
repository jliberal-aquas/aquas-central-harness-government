package entorno

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestBinarioViejoSeNombra: el gate avisa si el binario de aquas es más viejo que el código de producción (D000026).
// Rojo cuando: un .go de cmd/ o internal/ más nuevo que el binario no se señala, o se señala un _test.go o un binario al día.
// Verde cuando: nombra el .go más nuevo y pide reconstruir; con el binario al día o solo pruebas más nuevas, calla.
func TestBinarioViejoSeNombra(t *testing.T) {
	raiz := t.TempDir()
	poner := func(rel string, cuando time.Time) string {
		p := filepath.Join(raiz, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(p, cuando, cuando); err != nil {
			t.Fatal(err)
		}
		return p
	}
	t0 := time.Date(2026, 10, 23, 12, 0, 0, 0, time.UTC)
	exe := poner(".tmp/aq.exe", t0)
	poner("internal/x/x.go", t0.Add(-time.Hour))
	poner("internal/x/x_test.go", t0.Add(time.Hour))
	if m := BinarioViejo(raiz, exe); m != "" {
		t.Errorf("binario al día (solo una prueba más nueva): debió callar; dio %q", m)
	}
	poner("cmd/aquas/main.go", t0.Add(time.Minute))
	if m := BinarioViejo(raiz, exe); !strings.Contains(m, "cmd/aquas/main.go") || !strings.Contains(m, "go build") {
		t.Errorf("binario viejo: debió nombrar cmd/aquas/main.go y pedir go build; dio %q", m)
	}
}
