package guardia

import (
	"strings"
	"testing"
)

// TestBashDestinoPorOpcion (E000/F001/S001/X02, D000119): cp, mv, install y ln con el destino en una opción (-t DIR, -tDIR, --target-directory=DIR).
// Rojo cuando: alguna de esas órdenes escribe fuera del repo sin bloqueo, o una hacia dentro se bloquea.
// Verde cuando: hacia fuera se bloquean todas y hacia dentro pasan todas.
func TestBashDestinoPorOpcion(t *testing.T) {
	raiz := raizPrueba(t)
	for _, c := range []string{
		`cp -t ~/fuera docs/a.md`,
		`cp -t../fuera docs/a.md`,
		`cp --target-directory=/tmp docs/a.md`,
		`cp --target-directory /tmp docs/a.md`,
		`cp -r -t /c/Users/x docs`,
		`mv --target-directory=../fuera docs/a.md`,
		`install --target-directory=/tmp docs/a.md`,
		`ln -s --target-directory=~/x docs/a.md`,
		`cp -rt /c/fuera docs`,         // hallazgo A1.3: t dentro de un grupo de letras
		`cp -Zt /c/fuera docs/a.md`,    // hallazgo A1.2: -Z no toma valor
		`cp --targ=/c/fuera docs/a.md`, // hallazgo A1.1: abreviatura GNU de --target-directory
		`cp --t /c/fuera docs/a.md`,
		`mv --target=/c/fuera docs/a.md`,
		`install --t=/c/fuera docs/a.md`,
		`ln -s --target /c/fuera docs/a.md`,
		`install -m755 -t /c/fuera docs/a.md`,
		`cp -S.txt docs/a.md /c/fuera/x`, // hallazgo A2.2: la t de .txt es el sufijo de -S, no -t
	} {
		if m := RevisarBash(raiz, raiz, c); !strings.Contains(m, "fuera del repositorio") {
			t.Errorf("%q escribe fuera: debe bloquearse; dio %q", c, m)
		}
	}
	for _, c := range []string{
		`cp -t .tmp docs/a.md`,
		`cp -t.tmp docs/a.md`,
		`cp --target-directory=.tmp docs/a.md`,
		`cp -t .tmp ~/.claude/x.md`, // copiar desde fuera hacia dentro
		`mv --target-directory=docs .tmp/a.md`,
		`cp -S t docs/a.md .tmp/b.md`, // la t es el sufijo de -S, no un destino
		`install -otom docs/a.md .tmp/b.md`,
	} {
		if m := RevisarBash(raiz, raiz, c); m != "" {
			t.Errorf("%q escribe dentro: debe pasar; dio %q", c, m)
		}
	}
}
