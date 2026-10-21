package guardia

import (
	"strings"
	"testing"
)

// E001/F002/S002: las fichas del backlog (trabajo/) solo las escribe `aquas backlog`.

func TestTrabajoSoloPorAquas(t *testing.T) {
	raiz := raizPrueba(t)
	for _, p := range []string{"trabajo/activo/E001-F002-S002.json", "trabajo/archivo/E001.json", "trabajo/nuevo.json"} {
		if m := RevisarArchivo(raiz, p); !strings.Contains(m, "aquas backlog") {
			t.Errorf("Write/Edit a %s debe bloquearse; dio %q", p, m)
		}
	}
	for _, c := range []string{
		`echo {} > trabajo/activo/E001.json`,
		`rm -rf trabajo`,
		`mv trabajo/activo/E001.json trabajo/archivo/`,
	} {
		if m := RevisarBash(raiz, raiz, c); !strings.Contains(m, "aquas backlog") {
			t.Errorf("%q debe bloquearse; dio %q", c, m)
		}
	}
	for _, c := range []string{
		`go run ./cmd/aquas backlog estado E001 CERRADO`,
		`cat trabajo/activo/E001.json`,
		`ls trabajo/archivo`,
	} {
		if m := RevisarBash(raiz, raiz, c); m != "" {
			t.Errorf("%q debe pasar; dio %q", c, m)
		}
	}
}
