package main

import (
	"strings"
	"testing"
)

// TestVerifyHijoEstiloAvisa (E000/F013/S001): en un hijo y en la central, aquas verify muestra la oración larga como aviso, no como fallo.
// Rojo cuando: la oración larga sale entre los fallos del gate o no sale.
// Verde cuando: sale una sola vez, como «Aviso · wiki:», en el hijo y en la central.
func TestVerifyHijoEstiloAvisa(t *testing.T) {
	larga := "# Inicio\n\n" + strings.TrimSpace(strings.Repeat("sol ", 21)) + ".\n\n## Navegación\n\n- nada\n"
	dir := hijoNuevo(t)
	escribir(t, dir, "docs/README.md", larga)
	s := aquas(t, dir, "", "verify")
	if n := strings.Count(s.out+s.err, "oración de 21 palabras"); n != 1 {
		t.Fatalf("hijo: la oración larga debe salir una vez; salió %d: %+v", n, s)
	}
	if !strings.Contains(s.err, "Aviso · wiki: docs/README.md:3: oración de 21 palabras") {
		t.Errorf("hijo: la oración larga debe salir como aviso; dio %+v", s)
	}
	// Central (hallazgo A1.2): carpeta sin lock de hijo; el aviso sale antes de que otro chequeo corte.
	central := t.TempDir()
	escribir(t, central, "docs/README.md", larga)
	c := aquas(t, central, "", "verify")
	if n := strings.Count(c.out+c.err, "oración de 21 palabras"); n != 1 || !strings.Contains(c.err, "Aviso · wiki: docs/README.md:3: oración de 21 palabras") {
		t.Errorf("central: la oración larga debe salir una vez como aviso; dio %+v", c)
	}
}
