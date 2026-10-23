package wiki

import (
	"strings"
	"testing"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/backlog"
)

// TestDocumentacionVerificable: integral de E001/F008; de la ficha a la página, un slice cerrado muestra todo lo que lo hace verificable.
// Rojo cuando: la página generada omite qué valida una prueba, cuándo cae o pasa, su rojo y verde, el código con funciones y parámetros, o un arreglo con sus pruebas, su rojo y su verde; o el gate de la wiki no queda limpio.
// Verde cuando: la página trae las tres partes completas y el gate no reporta problemas.
func TestDocumentacionVerificable(t *testing.T) {
	raiz := conFichas(t)
	escribir(t, raiz, "README.md", "- [E001](docs/E001/README.md)\n")
	escribir(t, raiz, "internal/x/x.go", "package x\n\n// Suma devuelve a más b.\nfunc Suma(a, b int) int { return a + b }\n")
	escribir(t, raiz, "x/x_test.go", "package x\n\nimport \"testing\"\n\nfunc TestA(t *testing.T) {}\n\nfunc TestB(t *testing.T) {}\n\nfunc TestC(t *testing.T) {}\n")
	reg := func(fecha string, r map[string]string) *backlog.Registro {
		return &backlog.Registro{Fecha: fecha, Head: "abc", Resultado: r}
	}
	ficha(t, raiz, backlog.Ficha{ID: "E001/F001/S001", Titulo: "Regla", Estado: "CERRADO", Secuencia: 5,
		Pruebas: []backlog.Prueba{
			{Nombre: "./x:TestA", Intencion: "bloquea fuera", RojoCuando: "deja escribir fuera", VerdeCuando: "niega fuera"},
			{Nombre: "./x:TestB", Intencion: "deja pasar", RojoCuando: "niega dentro", VerdeCuando: "acepta dentro"}},
		Rojo:   reg("2026-10-21T10:00:00Z", map[string]string{"./x:TestA": "falló", "./x:TestB": "falló"}),
		Verde:  reg("2026-10-21T11:00:00Z", nil),
		Codigo: []string{"internal/x/x.go"},
		Arreglos: []backlog.Arreglo{{ID: "E001/F001/S001/X01", Titulo: "ruta con espacios", Motivo: "D000099", Fecha: "2026-10-22T09:00:00Z",
			Pruebas: []backlog.Prueba{{Nombre: "./x:TestC", Intencion: "acepta espacios", RojoCuando: "rompe", VerdeCuando: "pasa"}},
			Rojo:    reg("2026-10-22T09:30:00Z", map[string]string{"./x:TestC": "falló"}),
			Verde:   reg("2026-10-23T10:00:00Z", nil)}}})
	generar(t, raiz)
	c := leer(t, raiz, tS001)
	for _, q := range []string{
		"| `TestA` | bloquea fuera | deja escribir fuera | niega fuera | 2026-10-21 | 2026-10-21 |", // prueba con valida, rojo/verde cuando y fechas
		"### Código", "[internal/x/x.go](../../../internal/x/x.go)", "`Suma`", "(a, b int) int", "Suma devuelve a más b.", // código
		"### Arreglos", "X01", "ruta con espacios", "D000099", "`TestC`", // arreglo y su prueba
		"2026-10-22 | 2026-10-23", // rojo y verde del arreglo (D000020)
	} {
		if !strings.Contains(c, q) {
			t.Errorf("la página del slice no muestra %q:\n%s", q, c)
		}
	}
	if p := verificar(t, raiz); len(p) > 0 {
		t.Errorf("gate de la wiki con problemas: %q", p)
	}
}
