package wiki

import (
	"strings"
	"testing"
)

// TestPortadasListanIniciativas: cada iniciativa del backlog tiene enlace en README.md y en docs/README.md.
// Rojo cuando: una iniciativa falta en una portada, o aparece solo como texto sin enlace, y el gate no lo dice.
// Verde cuando: el gate nombra la portada y la iniciativa que faltan; con todas enlazadas no reporta nada.
func TestPortadasListanIniciativas(t *testing.T) {
	ids := []string{"E000", "E001"}
	readme := "- [E000 Mantenimiento](docs/E000/README.md)\n- E001 Gobierno, sin enlace\n"
	wiki := "| [E000](E000/README.md) |\n| [E001](E001/README.md) |\n"
	ps := faltanIniciativas(ids, readme, wiki)
	if len(ps) != 1 || !strings.Contains(ps[0], "README.md") || !strings.Contains(ps[0], "E001") {
		t.Errorf("se esperaba solo E001 sin enlace en README.md; salió %q", ps)
	}
	if ps := faltanIniciativas(ids, readme, "| [E000](E000/README.md) |\n"); len(ps) != 2 {
		t.Errorf("se esperaba también E001 ausente de docs/README.md; salió %q", ps)
	}
	if ps := faltanIniciativas(ids, readme+"- [E001](docs/E001/README.md)\n", wiki); len(ps) != 0 {
		t.Errorf("todas enlazadas: no debía reportar; salió %q", ps)
	}
}
