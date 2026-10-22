package wiki

import (
	"strings"
	"testing"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/backlog"
)

// TestTablaArreglos: la página de un slice lista sus arreglos con qué arreglan, por qué, sus pruebas y la fecha.
// Rojo cuando: falta la sección «Arreglos», una columna o el slice que el arreglo absorbió.
// Verde cuando: la tabla sale completa; una ficha sin arreglos no muestra la sección.
func TestTablaArreglos(t *testing.T) {
	f := backlog.Ficha{Tipo: "slice", Arreglos: []backlog.Arreglo{{ID: "E002/F001/S002/X01", Titulo: "Título del hijo", Motivo: "defecto de S002",
		Fecha: "2026-10-22T10:00:00Z", Antes: "E002/F001/S005", Pruebas: []backlog.Prueba{{Nombre: "./x:TestT", Intencion: "t"}}}}}
	want := "### Arreglos\n\n| Arreglo | Qué arregla | Por qué | Pruebas | Fecha |\n|---|---|---|---|---|\n" +
		"| X01 | Título del hijo (antes S005) | defecto de S002 | `TestT` | 2026-10-22 |\n"
	if got := tablaArreglos(f); !strings.Contains(got, want) {
		t.Errorf("falta:\n%s\nen:\n%s", want, got)
	}
	if got := tablaArreglos(backlog.Ficha{}); got != "" {
		t.Errorf("sin arreglos no hay sección; salió %q", got)
	}
}
