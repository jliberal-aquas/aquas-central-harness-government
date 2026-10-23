package backlog

import (
	"strings"
	"testing"
)

// TestPruebaExigeValidaRojoVerde: toda prueba nueva dice qué valida, cuándo queda en rojo y cuándo en verde.
// Rojo cuando: se acepta una prueba sin «rojo cuando» o sin «verde cuando», o se pierden al guardarla.
// Verde cuando: las incompletas fallan y la completa guarda sus tres campos; el gate señala la que trae uno solo.
func TestPruebaExigeValidaRojoVerde(t *testing.T) {
	for _, mala := range []string{"./x:TestA=valida", "./x:TestA=valida|rojo", "./x:TestA=valida||verde"} {
		if _, err := listaPruebas([]string{mala}); err == nil {
			t.Errorf("%q: debió fallar por campo faltante", mala)
		}
	}
	ps, err := listaPruebas([]string{"./x:TestA=valida algo|rojo si X|verde si Y"})
	if err != nil {
		t.Fatal(err)
	}
	if p := ps[0]; p.Intencion != "valida algo" || p.RojoCuando != "rojo si X" || p.VerdeCuando != "verde si Y" {
		t.Errorf("campos mal guardados: %+v", p)
	}
	if m := problemaPrueba(Prueba{Nombre: "./x:TestA", Intencion: "v", RojoCuando: "r"}); m == "" {
		t.Error("el gate debió señalar que falta «verde cuando»")
	}
}

// TestPruebaConPartesDeMasCae: una prueba con más de tres partes separadas por | se rechaza (D000024).
// Rojo cuando: el texto sobrante se pega a «verde cuando» y la prueba se registra sin aviso.
// Verde cuando: cuatro partes caen con un mensaje que dice cuántas hay; tres partes pasan.
func TestPruebaConPartesDeMasCae(t *testing.T) {
	_, err := listaPruebas([]string{"./x:TestA=exige relacion=nueva|amplia|corrige|verde"})
	if err == nil || !strings.Contains(err.Error(), "4 partes") {
		t.Errorf("cuatro partes: debió caer diciendo cuántas hay; dio %v", err)
	}
	if _, err := listaPruebas([]string{"./x:TestA=valida|rojo|verde"}); err != nil {
		t.Errorf("tres partes: debió pasar; dio %v", err)
	}
}
