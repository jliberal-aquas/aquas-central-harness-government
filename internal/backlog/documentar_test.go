package backlog

import (
	"strings"
	"testing"
)

// fichaCerrada: slice archivado con una prueba vieja (solo valida), rojo y verde.
func fichaCerrada(t *testing.T, raiz string) Ficha {
	f := Ficha{ID: "E001/F001/S001", Tipo: "slice", Titulo: "T", Estado: "CERRADO", Prioridad: "P1", Motivo: "m", Padre: "E001/F001",
		BloqueadoPor: []string{}, Secuencia: 1, Pruebas: []Prueba{{Nombre: "./x:TestA", Intencion: "vieja"}},
		Rojo: &Registro{Fecha: "2026-10-01T00:00:00Z", Resultado: map[string]string{"./x:TestA": "falló"}}, Verde: &Registro{Fecha: "2026-10-01T01:00:00Z"}}
	pisar(t, raiz, Archivo, f)
	return f
}

// conArreglo: cuelga de la ficha cerrada un arreglo con una prueba vieja.
func conArreglo(t *testing.T, raiz string) {
	f, _, _ := Ver(raiz, "E001/F001/S001")
	f.Arreglos = []Arreglo{{ID: "E001/F001/S001/X01", Titulo: "t", Motivo: "m", Fecha: "2026-10-02", Pruebas: []Prueba{{Nombre: "./x:TestB", Intencion: "vieja"}}}}
	pisar(t, raiz, Archivo, f)
}

// TestDocumentarNoBorraRojoVerde: documentar cambia solo los textos de una prueba ya registrada, también en fichas cerradas y en arreglos.
// Rojo cuando: borra rojo o verde, rechaza la ficha cerrada o el ID de un arreglo, acepta un campo vacío o una prueba ajena.
// Verde cuando: guarda los tres textos, en la ficha y en su arreglo, y conserva rojo y verde; lo incompleto o ajeno falla.
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
	conArreglo(t, raiz)
	if _, err := Documentar(raiz, "E001/F001/S001/X01", []string{"./x:TestB=v|r|v"}); err != nil {
		t.Fatalf("documentar en arreglo: %v", err)
	}
	if h, _, _ := Ver(raiz, "E001/F001/S001"); h.Arreglos[0].Pruebas[0].VerdeCuando != "v" || h.Rojo == nil {
		t.Errorf("arreglo no guardado o rojo perdido: %+v", h)
	}
}

// TestGateExigeCuandoATodas: el gate exige rojo cuando y verde cuando a toda prueba, vieja o de un arreglo.
// Rojo cuando: el gate calla ante una prueba vieja o de arreglo sin los dos textos.
// Verde cuando: el gate da un aviso por cada prueba incompleta y calla cuando todas están completas.
func TestGateExigeCuandoATodas(t *testing.T) {
	raiz := t.TempDir()
	fichaCerrada(t, raiz)
	conArreglo(t, raiz)
	avisos := func() (n int) {
		for _, m := range Validar(raiz) {
			if strings.Contains(m, "rojo cuando") {
				n++
			}
		}
		return n
	}
	if n := avisos(); n != 2 {
		t.Errorf("prueba vieja y de arreglo incompletas: %d avisos, se esperaban 2: %v", n, Validar(raiz))
	}
	f, _, _ := Ver(raiz, "E001/F001/S001")
	f.Pruebas[0].RojoCuando, f.Pruebas[0].VerdeCuando = "r", "v"
	f.Arreglos[0].Pruebas[0].RojoCuando, f.Arreglos[0].Pruebas[0].VerdeCuando = "r", "v"
	pisar(t, raiz, Archivo, f)
	if n := avisos(); n != 0 {
		t.Errorf("todas completas: %d avisos: %v", n, Validar(raiz))
	}
}
