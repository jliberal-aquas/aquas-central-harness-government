package backlog

import (
	"strings"
	"testing"
)

// TestPruebasNoPisaArregloAjeno: backlog pruebas se niega a cambiar la prueba que un arreglo ya declara (D000028).
// Rojo cuando: una prueba distinta reemplaza en silencio la del arreglo, o se niega reescribir la misma prueba.
// Verde cuando: otra prueba se rechaza nombrando la que ya tiene; la misma prueba con otro texto se acepta.
func TestPruebasNoPisaArregloAjeno(t *testing.T) {
	raiz := base(t)
	pisar(t, raiz, Archivo, Ficha{ID: "E001/F002/S009", Tipo: "slice", Titulo: "T", Estado: "CERRADO", Prioridad: "P1",
		Motivo: "m", Padre: "E001/F002", BloqueadoPor: []string{},
		Pruebas: []Prueba{{Nombre: "./p:TestSlice", Intencion: "i"}},
		Arreglos: []Arreglo{{ID: "E001/F002/S009/X01", Titulo: "t", Motivo: "m", Fecha: "2026-10-23T00:00:00Z",
			Pruebas: []Prueba{{Nombre: "./p:TestVieja", Intencion: "i"}}}}})
	const x = "E001/F002/S009/X01"
	if _, err := Pruebas(raiz, x, []string{"./p:TestNueva=a|r|v"}); err == nil || !strings.Contains(err.Error(), "TestVieja") {
		t.Errorf("otra prueba sobre un arreglo con prueba: debió negarse nombrando TestVieja; dio %v", err)
	}
	if _, err := Pruebas(raiz, x, []string{"./p:TestVieja=otro texto|r|v"}); err != nil {
		t.Errorf("misma prueba con otro texto: debió aceptarse; dio %v", err)
	}
}
