package wiki

import (
	"strings"
	"testing"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/backlog"
)

// TestTablaPruebasMuestraCuando: la tabla del slice muestra qué valida cada prueba, cuándo queda en rojo y cuándo en verde.
// Rojo cuando: la tabla omite alguna de esas columnas o pierde su texto.
// Verde cuando: las tres columnas salen con su texto; las pruebas viejas muestran «—».
func TestTablaPruebasMuestraCuando(t *testing.T) {
	f := backlog.Ficha{Pruebas: []backlog.Prueba{
		{Nombre: "./x:TestA", Intencion: "valida A", RojoCuando: "falta A", VerdeCuando: "hay A"},
		{Nombre: "./x:TestB", Intencion: "vieja"},
	}}
	got := tablaPruebas(f)
	for _, w := range []string{
		"| Prueba | Valida | Rojo cuando | Verde cuando | Rojo | Verde |",
		"| `TestA` | valida A | falta A | hay A | — | — |",
		"| `TestB` | vieja | — | — | — | — |",
	} {
		if !strings.Contains(got, w) {
			t.Errorf("falta %q en:\n%s", w, got)
		}
	}
}
