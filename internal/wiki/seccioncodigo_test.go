package wiki

import (
	"strings"
	"testing"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/backlog"
)

// TestSliceEnlazaSuCodigo: la página del slice lista su código: archivo enlazado, función, parámetros y qué hace.
// Rojo cuando: falta la sección Código, el enlace al archivo, una función, sus parámetros o su comentario, o un archivo borrado queda como enlace roto.
// Verde cuando: cada archivo del slice aparece enlazado con sus funciones, parámetros y comentario; el borrado aparece sin enlace.
func TestSliceEnlazaSuCodigo(t *testing.T) {
	raiz := t.TempDir()
	escribir(t, raiz, "internal/x/x.go", "package x\n\n// Suma devuelve a más b.\nfunc Suma(a, b int) int { return a + b }\n")
	f := backlog.Ficha{ID: "E001/F001/S001", Tipo: "slice", Codigo: []string{"internal/x/x.go", "internal/y/borrado.go"}}
	s, err := tablaCodigo(raiz, f)
	if err != nil {
		t.Fatal(err)
	}
	for _, q := range []string{"### Código", "[internal/x/x.go](../../../internal/x/x.go)", "`Suma`", "(a, b int) int", "Suma devuelve a más b."} {
		if !strings.Contains(s, q) {
			t.Errorf("falta %q en:\n%s", q, s)
		}
	}
	if strings.Contains(s, "](../../../internal/y/borrado.go)") || !strings.Contains(s, "internal/y/borrado.go") {
		t.Errorf("el archivo borrado debe aparecer sin enlace:\n%s", s)
	}
}
