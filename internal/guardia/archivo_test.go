package guardia

import (
	"strings"
	"testing"
)

// E001/F005/S003: el archivo de memoria (registro/memoria/) solo lo escribe `aquas memoria archivar`.

func TestArchivoMemoriaSoloPorAquas(t *testing.T) {
	raiz := raizPrueba(t)
	for _, p := range []string{"registro/memoria/E001-F005-S003.json", "registro/memoria/nuevo.json"} {
		if m := RevisarArchivo(raiz, p); !strings.Contains(m, "aquas memoria archivar") {
			t.Errorf("Write/Edit a %s debe bloquearse; dio %q", p, m)
		}
	}
	for _, c := range []string{
		`echo {} > registro/memoria/E001-F005-S003.json`,
		`rm -rf registro/memoria`,
	} {
		if m := RevisarBash(raiz, raiz, c); !strings.Contains(m, "aquas memoria archivar") {
			t.Errorf("%q debe bloquearse; dio %q", c, m)
		}
	}
	for _, c := range []string{
		`go run ./cmd/aquas memoria archivar`,
		`cat registro/memoria/E001-F005-S003.json`,
		`echo x > registro/notas.txt`,
	} {
		if m := RevisarBash(raiz, raiz, c); m != "" {
			t.Errorf("%q debe pasar; dio %q", c, m)
		}
	}
}
