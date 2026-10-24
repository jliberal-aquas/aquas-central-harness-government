package wiki

import (
	"strings"
	"testing"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/backlog"
)

// E001/F010/S001: una prueba nombrada que aún no existe en el código se acepta solo mientras su ficha no tiene rojo.

// Rojo cuando: el gate acepta una prueba inexistente aunque su ficha ya tenga rojo o esté archivada, o rechaza la prevista sin rojo.
// Verde cuando: ficha sin rojo → aceptada; ficha con rojo → rechazada; ficha archivada → rechazada.
func TestPruebaPrevistaSoloSinRojo(t *testing.T) {
	const falta = "prueba nombrada no existe en el código: TestFutura"
	casos := []struct {
		nombre   string
		estado   string
		rojo     bool
		rechazar bool
	}{
		{"ficha sin rojo", "EN_CURSO", false, false},
		{"ficha con rojo", "EN_CURSO", true, true},
		{"ficha archivada", "CERRADO", false, true},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			raiz := wikiValida(t)
			f := backlog.Ficha{ID: "E001/F001/S001", Titulo: "Regla", Estado: c.estado, Secuencia: 5,
				Pruebas: []backlog.Prueba{{Nombre: "./x:TestFutura", Intencion: "aún no escrita"}}}
			if c.rojo {
				f.Rojo = &backlog.Registro{Fecha: "2026-10-21T10:00:00Z", Head: "abc", Resultado: map[string]string{"./x:TestFutura": "falló"}}
			}
			ficha(t, raiz, f)
			reemplazar(t, raiz, tS001, "Lo prometido y su prueba.", "Prueba: `TestFutura`.")
			generar(t, raiz)
			hay := false
			for _, p := range verificar(t, raiz) {
				hay = hay || strings.Contains(p, falta)
			}
			if hay != c.rechazar {
				t.Errorf("%s: rechazar=%v pero el gate dio %v (problemas: %q)", c.nombre, c.rechazar, hay, verificar(t, raiz))
			}
		})
	}
}
