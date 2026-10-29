package wiki

import (
	"strings"
	"testing"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/backlog"
)

// E001/F002/S011: la columna «Qué es» de la tabla de hijos muestra que_hace.

// Rojo cuando: la tabla muestra el título de un hijo que tiene que_hace (activo o archivado), no escapa la barra,
// no cae al título cuando falta que_hace, o el tope de backlog difiere del de la wiki.
// Verde cuando: cada fila trae que_hace de su ficha, o el título si no lo tiene, y los dos topes coinciden.
func TestHijosMuestranQueHace(t *testing.T) {
	if backlog.MaxPalabrasQueHace != MaxPalabrasOracion {
		t.Errorf("tope de que_hace %d != MaxPalabrasOracion %d", backlog.MaxPalabrasQueHace, MaxPalabrasOracion)
	}
	raiz := conFichas(t)
	escribir(t, raiz, "README.md", "- [E001](docs/E001/README.md)\n")
	ficha(t, raiz, backlog.Ficha{ID: "E001", Titulo: "Gobierno", QueHace: "frena pedidos sin plan", Estado: "EN_CURSO", Secuencia: 1})
	ficha(t, raiz, backlog.Ficha{ID: "E002", Titulo: "Otra | rara", QueHace: "separa a | b", Estado: "BORRADOR", Secuencia: 2})
	ficha(t, raiz, backlog.Ficha{ID: "E001/F001", Titulo: "Guardia", QueHace: "bloquea escrituras fuera del repo", Estado: "CERRADO", Secuencia: 3})
	// Hallazgo A1.1: que_hace en blanco (JSON escrito a mano) cuenta como ausente: muestra el título.
	ficha(t, raiz, backlog.Ficha{ID: "E001/F002", Titulo: "Backlog", QueHace: "   ", Estado: "BORRADOR", Secuencia: 4})
	generar(t, raiz) // F002 queda sin que_hace útil
	filas := map[string][]string{
		tInicio: {"| [E001](E001/README.md) | frena pedidos sin plan | En curso |\n", "| E002 | separa a \\| b | Planeado |\n"},
		tE001:   {"| [F001](F001/README.md) | bloquea escrituras fuera del repo | Hecho |\n", "| F002 | Backlog | Planeado |\n"},
	}
	for rel, quiere := range filas {
		c := leer(t, raiz, rel)
		for _, f := range quiere {
			if !strings.Contains(c, f) {
				t.Errorf("%s: falta la fila %q en:\n%s", rel, f, c)
			}
		}
	}
	if c := leer(t, raiz, tInicio); strings.Contains(c, "| Gobierno |") || strings.Contains(c, "Otra") {
		t.Errorf("con que_hace, la tabla no debe mostrar el título:\n%s", c)
	}
}
