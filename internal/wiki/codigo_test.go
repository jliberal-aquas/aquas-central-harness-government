package wiki

import (
	"strings"
	"testing"
)

// TestFuncionSinComentarioCae: el gate exige comentario a cada función y el índice lista cada una con su comentario.
// Rojo cuando: una función sin comentario no hace caer el gate, o el índice omite una función o su comentario.
// Verde cuando: la función sin comentario cae nombrada; con comentario el gate calla y el índice la lista.
func TestFuncionSinComentarioCae(t *testing.T) {
	raiz := t.TempDir()
	escribir(t, raiz, "internal/x/x.go", "package x\n\n// Suma devuelve a más b.\nfunc Suma(a, b int) int { return a + b }\n\nfunc resta(a, b int) int { return a - b }\n")
	escribir(t, raiz, "internal/x/x_test.go", "package x\n\nfunc ayuda() {}\n")
	ps, err := revisarCodigo(raiz)
	if err != nil {
		t.Fatal(err)
	}
	if len(ps) != 1 || !strings.Contains(ps[0], "internal/x/x.go") || !strings.Contains(ps[0], "resta") {
		t.Errorf("se esperaba un aviso que nombre resta en internal/x/x.go: %v", ps)
	}
	idx, err := indiceCodigo(raiz)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(idx, "`Suma`") || !strings.Contains(idx, "Suma devuelve a más b.") || strings.Contains(idx, "ayuda") {
		t.Errorf("índice sin Suma o su comentario, o con funciones de prueba:\n%s", idx)
	}
	escribir(t, raiz, "internal/x/x.go", "package x\n\n// Suma devuelve a más b.\nfunc Suma(a, b int) int { return a + b }\n\n// resta devuelve a menos b.\nfunc resta(a, b int) int { return a - b }\n")
	if ps, _ := revisarCodigo(raiz); len(ps) != 0 {
		t.Errorf("todas comentadas: %v", ps)
	}
	if idx, _ := indiceCodigo(raiz); !strings.Contains(idx, "resta devuelve a menos b.") {
		t.Errorf("índice sin resta:\n%s", idx)
	}
}
