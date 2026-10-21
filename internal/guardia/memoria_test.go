package guardia

import (
	"path/filepath"
	"strings"
	"testing"
)

// E001/F005/S001: la memoria caliente (.aquas/) solo la escribe `aquas memoria fijar`.

func TestMemoriaSoloPorAquas(t *testing.T) {
	raiz := raizPrueba(t)
	for _, p := range []string{".aquas/memoria.json", filepath.Join(raiz, ".aquas", "memoria.json"), ".aquas/otra.json"} {
		if m := RevisarArchivo(raiz, p); !strings.Contains(m, "memoria caliente") {
			t.Errorf("Write/Edit a %s debe bloquearse; dio %q", p, m)
		}
	}
	for _, c := range []string{
		`echo {} > .aquas/memoria.json`,
		`rm -rf .aquas`,
		`sed -i 's/a/b/' .aquas/memoria.json`,
		`cd .aquas && touch memoria.json`,
	} {
		if m := RevisarBash(raiz, raiz, c); !strings.Contains(m, "memoria caliente") {
			t.Errorf("%q debe bloquearse; dio %q", c, m)
		}
	}
	for _, c := range []string{
		`go run ./cmd/aquas memoria fijar tarea=x`,
		`cat .aquas/memoria.json`,
		`echo x > .aquasnota.txt`,
	} {
		if m := RevisarBash(raiz, raiz, c); m != "" {
			t.Errorf("%q debe pasar; dio %q", c, m)
		}
	}
}
