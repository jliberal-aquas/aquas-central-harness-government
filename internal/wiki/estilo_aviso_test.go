package wiki

import (
	"strings"
	"testing"
)

// largo: oración de 21 palabras.
var largo = strings.TrimSpace(strings.Repeat("sol ", 21)) + "."

// contiene: algún elemento de xs contiene sub.
func contiene(xs []string, sub string) bool {
	for _, x := range xs {
		if strings.Contains(x, sub) {
			return true
		}
	}
	return false
}

// TestEstiloSoloAvisa (integral E000/F013): el estilo de la wiki avisa y no frena el gate.
// Rojo cuando: una oración larga o una mención sin enlace al glosario hace caer el gate.
// Verde cuando: la wiki con esos dos estilos pasa y cada uno sale como aviso.
func TestEstiloSoloAvisa(t *testing.T) {
	raiz := wikiValida(t)
	reemplazar(t, raiz, tS001, "Se lee.", largo+" La zeta manda.")
	p, a, err := VerificarConAvisos(raiz)
	if err != nil {
		t.Fatal(err)
	}
	if len(p) != 0 {
		t.Errorf("el estilo no debe frenar; frenó con %q", p)
	}
	if !contiene(a, "oración de 21 palabras") || !contiene(a, "«zeta»") {
		t.Errorf("oración larga y término sin enlace deben salir como aviso; avisos %q", a)
	}
}

// TestOracionLargaAvisa (E000/F013/S001): oración larga y primera mención sin enlace avisan con ruta; el resto de la wiki sigue frenando.
// Rojo cuando: alguno de los dos estilos frena, no avisa, o un fallo de otra regla deja de frenar.
// Verde cuando: los dos avisan en páginas de slice, feature e iniciativa, una vez por término, y otra regla sigue frenando.
func TestOracionLargaAvisa(t *testing.T) {
	raiz := wikiValida(t)
	for _, rel := range []string{tS001, tF001, tE001} { // cada tipo de página
		reemplazar(t, raiz, rel, "Se lee.", largo+" La zeta manda. Otra zeta más.")
	}
	p, a, err := VerificarConAvisos(raiz)
	if err != nil {
		t.Fatal(err)
	}
	if len(p) != 0 {
		t.Errorf("el estilo no debe frenar; frenó con %q", p)
	}
	for _, rel := range []string{tS001, tF001, tE001} {
		larga := false // hallazgo A1.3: el aviso de oración larga de esta página, no cualquier aviso con su ruta
		for _, x := range a {
			larga = larga || (strings.HasPrefix(x, rel+":") && strings.Contains(x, ": oración de 21 palabras"))
		}
		if !larga || !contiene(a, rel+": primera aparición de «zeta»") {
			t.Errorf("%s: debe avisar oración larga con línea y término sin enlace; avisos %q", rel, a)
		}
	}
	n := 0
	for _, x := range a {
		if strings.Contains(x, tS001+": primera aparición de «zeta»") {
			n++
		}
	}
	if n != 1 {
		t.Errorf("término repetido: un solo aviso por página; hubo %d", n)
	}
	// Verificar (lo que usa el gate) no devuelve los avisos.
	if v, err := Verificar(raiz); err != nil || len(v) != 0 {
		t.Errorf("Verificar no debe devolver avisos de estilo; dio %q %v", v, err)
	}
	// Choque con otro control: otra regla de la misma página sigue frenando.
	reemplazar(t, raiz, tS001, "## Cómo se usa", "## Cómo se usaba")
	if p, a, err := VerificarConAvisos(raiz); err != nil || !contiene(p, "Cómo se usa") || !contiene(a, tS001+": primera aparición de «zeta»") {
		t.Errorf("una sección faltante debe seguir frenando junto al aviso de estilo; problemas %q avisos %q err %v", p, a, err)
	}
	// Página vacía: frena por sus secciones y no deja avisos de estilo propios.
	escribir(t, raiz, tS001, "")
	p, a, err = VerificarConAvisos(raiz)
	if err != nil || len(p) == 0 || contiene(a, tS001+":") {
		t.Errorf("página vacía debe frenar por secciones y sin aviso de estilo propio; problemas %q avisos %q err %v", p, a, err)
	}
}

// exigeAviso: el texto aparece entre los avisos y la wiki no tiene problemas.
func exigeAviso(t *testing.T, raiz, sub string) {
	t.Helper()
	p, a, err := VerificarConAvisos(raiz)
	if err != nil {
		t.Fatal(err)
	}
	if len(p) != 0 {
		t.Fatalf("estilo no debe frenar; problemas: %q", p)
	}
	for _, x := range a {
		if strings.Contains(x, sub) {
			return
		}
	}
	t.Fatalf("se esperaba aviso con %q; hubo: %q", sub, a)
}
