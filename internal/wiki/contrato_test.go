package wiki

import (
	"testing"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/backlog"
)

// E001/F001/S004: sección «Contrato y evidencia», migas y trazabilidad de pruebas.

const tPruebaGo = "package x\n\nimport \"testing\"\n\nfunc TestNombrada(t *testing.T) {}\n"

func TestFallasContrato(t *testing.T) {
	casos := []struct {
		nombre string
		mutar  func(t *testing.T, raiz string)
		espera string
	}{
		{"S sin contrato", func(t *testing.T, r string) {
			reemplazar(t, r, tS001, "## Contrato y evidencia", "## Otra cosa")
		}, "docs/E001/F001/S001.md: falta sección «Contrato y evidencia»"},
		{"E sin contrato", func(t *testing.T, r string) {
			reemplazar(t, r, tE001, "## Contrato y evidencia", "## Otra cosa")
		}, "docs/E001/README.md: falta sección «Contrato y evidencia»"},
		{"contrato tras Navegación", func(t *testing.T, r string) {
			reemplazar(t, r, tF001, "## Contrato y evidencia", "## TMP")
			reemplazar(t, r, tF001, "## Navegación", "## Contrato y evidencia")
			reemplazar(t, r, tF001, "## TMP", "## Navegación")
		}, "docs/E001/F001/README.md: secciones fuera de orden"},
		{"S sin migas", func(t *testing.T, r string) {
			reemplazar(t, r, tS001, "[Inicio](../../README.md) › [E001](../README.md) › [F001](README.md) › S001", "")
		}, "docs/E001/F001/S001.md: migas"},
		{"migas con nivel equivocado", func(t *testing.T, r string) {
			reemplazar(t, r, tF001, "› F001", "› F002")
		}, "docs/E001/F001/README.md: migas"},
		{"prueba nombrada que no existe", func(t *testing.T, r string) {
			reemplazar(t, r, tS001, "Lo prometido y su prueba.", "Prueba: `TestFantasma`.")
		}, "docs/E001/F001/S001.md: prueba nombrada no existe en el código: TestFantasma"},
		{"prueba sin slice dueño", func(t *testing.T, r string) {
			escribir(t, r, "x/x_test.go", tPruebaGo)
		}, "x/x_test.go: prueba sin slice dueño: TestNombrada"},
		{"nombrada fuera de un slice no cuenta", func(t *testing.T, r string) {
			escribir(t, r, "x/x_test.go", tPruebaGo)
			reemplazar(t, r, tF001, "Lo prometido y su prueba.", "Prueba: `TestNombrada`.")
		}, "x/x_test.go: prueba sin slice dueño: TestNombrada"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			raiz := wikiValida(t)
			c.mutar(t, raiz)
			exigeProblema(t, raiz, c.espera)
		})
	}
}

// TestIntegralConDuenoEnSuFeature: la integral de una feature puede vivir nombrada solo en el README de su feature (D000025).
// Rojo cuando: el gate pide página de slice para una prueba que la ficha de la feature declara como suya.
// Verde cuando: nombrada en «Contrato y evidencia» de su feature, el gate queda limpio; una prueba ajena a la ficha sigue sin dueño (caso «nombrada fuera de un slice no cuenta»).
func TestIntegralConDuenoEnSuFeature(t *testing.T) {
	raiz := conFichas(t)
	escribir(t, raiz, "README.md", "- [E001](docs/E001/README.md)\n")
	ficha(t, raiz, backlog.Ficha{ID: "E001/F001", Titulo: "Guardia", Estado: "CERRADO", Secuencia: 3,
		Pruebas: []backlog.Prueba{{Nombre: "./x:TestNombrada", Intencion: "integral"}}})
	escribir(t, raiz, "x/x_test.go", "package x\n\nimport \"testing\"\n\nfunc TestA(t *testing.T) {}\n\nfunc TestB(t *testing.T) {}\n\nfunc TestNombrada(t *testing.T) {}\n")
	reemplazar(t, raiz, tF001, "Lo prometido y su prueba.", "Integral: `TestNombrada`.")
	generar(t, raiz)
	if p := verificar(t, raiz); len(p) > 0 {
		t.Fatalf("la integral nombrada en su feature debió tener dueño; problemas: %q", p)
	}
}

func TestPruebaConDuenoPasa(t *testing.T) {
	raiz := wikiValida(t)
	escribir(t, raiz, "x/x_test.go", tPruebaGo)
	escribir(t, raiz, ".tmp/y_test.go", "package y\n\nfunc TestOculta(t *testing.T) {}\n")
	reemplazar(t, raiz, tS001, "Lo prometido y su prueba.", "Prueba: `TestNombrada`.")
	if p := verificar(t, raiz); len(p) > 0 {
		t.Fatalf("se esperaba gate OK, problemas: %q", p)
	}
}

func TestInicioYRefSinContratoNiMigasPasan(t *testing.T) {
	raiz := wikiValida(t)
	escribir(t, raiz, "docs/ref/algo.md", tPagina("EJ-REF-algo", "", mermaid("flowchart LR"),
		"- ↑ Subir: [Inicio](../README.md)"))
	reemplazar(t, raiz, tInicio, "- Glosario:", "- Referencia: [Algo](ref/algo.md)\n- Glosario:")
	if p := verificar(t, raiz); len(p) > 0 {
		t.Fatalf("se esperaba gate OK, problemas: %q", p)
	}
}
