package backlog

import "testing"

// fichaCerrada: slice archivado con una prueba vieja (solo valida), rojo y verde.
func fichaCerrada(t *testing.T, raiz string) Ficha {
	f := Ficha{ID: "E001/F001/S001", Tipo: "slice", Titulo: "T", Estado: "CERRADO", Prioridad: "P1", Motivo: "m", Padre: "E001/F001",
		BloqueadoPor: []string{}, Secuencia: 1, Pruebas: []Prueba{{Nombre: "./x:TestA", Intencion: "vieja"}},
		Rojo: &Registro{Fecha: "2026-10-01T00:00:00Z", Resultado: map[string]string{"./x:TestA": "falló"}}, Verde: &Registro{Fecha: "2026-10-01T01:00:00Z"}}
	pisar(t, raiz, Archivo, f)
	return f
}

// TestDocumentarNoBorraRojoVerde: documentar cambia solo los textos de una prueba ya registrada, también en fichas cerradas.
// Rojo cuando: borra rojo o verde, rechaza la ficha cerrada, acepta un campo vacío o una prueba que la ficha no tiene.
// Verde cuando: guarda los tres textos y conserva rojo y verde; lo incompleto o ajeno falla.
func TestDocumentarNoBorraRojoVerde(t *testing.T) {
	raiz := t.TempDir()
	fichaCerrada(t, raiz)
	if _, err := Documentar(raiz, "E001/F001/S001", []string{"./x:TestA=valida|rojo"}); err == nil {
		t.Error("sin verde cuando: debió fallar")
	}
	if _, err := Documentar(raiz, "E001/F001/S001", []string{"./x:TestZ=v|r|v"}); err == nil {
		t.Error("prueba que la ficha no tiene: debió fallar")
	}
	g, err := Documentar(raiz, "E001/F001/S001", []string{"./x:TestA=valida A|cae si X|pasa si Y"})
	if err != nil {
		t.Fatal(err)
	}
	p := g.Pruebas[0]
	if p.Intencion != "valida A" || p.RojoCuando != "cae si X" || p.VerdeCuando != "pasa si Y" || g.Rojo == nil || g.Verde == nil {
		t.Errorf("textos o rojo/verde mal: %+v", g)
	}
	if h, archivada, _ := Ver(raiz, "E001/F001/S001"); !archivada || h.Pruebas[0].RojoCuando != "cae si X" || h.Rojo == nil {
		t.Errorf("no quedó guardado en el archivo: %+v", h)
	}
}

// TestDeudaCuentaPruebasSinCuando: el gate sabe cuántas pruebas faltan por completar.
// Rojo cuando: el conteo ignora las fichas archivadas o cuenta las ya completas.
// Verde cuando: cuenta 1 con la prueba vieja y 0 después de documentarla.
func TestDeudaCuentaPruebasSinCuando(t *testing.T) {
	raiz := t.TempDir()
	fichaCerrada(t, raiz)
	if n, err := Deuda(raiz); err != nil || n != 1 {
		t.Errorf("deuda con una prueba vieja: %d, %v", n, err)
	}
	if _, err := Documentar(raiz, "E001/F001/S001", []string{"./x:TestA=v|r|v"}); err != nil {
		t.Fatal(err)
	}
	if n, _ := Deuda(raiz); n != 0 {
		t.Errorf("deuda tras documentar: %d", n)
	}
}
