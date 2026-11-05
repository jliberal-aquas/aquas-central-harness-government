package backlog

import (
	"strings"
	"testing"
)

// TestSelloSinMenciones (E000/F013/S004): el sello omite menciones e importaciones que vienen de archivos .go.
// Rojo cuando: el sello pide confirmar una prueba o un código Go que solo nombra o importa una ruta tocada.
// Verde cuando: esas menciones no salen y las páginas, los canónicos y las fichas padre siguen en la lista.
// repoSello trae, en su commit base, cmd/x/main.go (importa internal/a), internal/a/extra_test.go y
// gobierno/canonico/x.md (los dos nombran internal/a/a.go, la ruta tocada).
func TestSelloSinMenciones(t *testing.T) {
	raiz := repoSello(t)
	m := claves(t, raiz, "E001/F002/S004")
	for k := range m {
		if strings.HasPrefix(k, "importa:") {
			t.Errorf("importación desde un .go no se pide: %s", k)
		}
		if fuente, _, ok := strings.Cut(strings.TrimPrefix(k, "nombra:"), ">"); ok && strings.HasPrefix(k, "nombra:") && strings.HasSuffix(fuente, ".go") {
			t.Errorf("mención desde un .go no se pide: %s", k)
		}
	}
	for _, k := range []string{"nombra:docs/ref.md>internal/a/a.go", "nombra:gobierno/canonico/x.md>internal/a/a.go",
		"ficha:E001/F002", "ficha:E001", "pagina:docs/E001/F002/S004.md"} {
		if _, ok := m[k]; !ok {
			t.Errorf("falta %s: páginas, canónicos y fichas padre se piden como hoy; hay %v", k, m)
		}
	}
}
