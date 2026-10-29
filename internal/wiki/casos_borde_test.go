package wiki

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/backlog"
)

// E001/F012: la página de un slice activo trae la tabla «Casos borde» con las 7 categorías.

const tS002 = "docs/E001/F001/S002.md"

// filaCB: una fila de la tabla «Casos borde».
type filaCB struct{ categoria, aplica, texto string }

// filasCompletas: las 7 categorías, cada una bien llenada.
func filasCompletas() []filaCB {
	var f []filaCB
	for i, c := range CategoriasCasosBorde {
		if i%2 == 0 {
			f = append(f, filaCB{c, "sí", "lo cubre `TestA`"})
		} else {
			f = append(f, filaCB{c, "no", "no hay tal cosa aquí"})
		}
	}
	return f
}

func tablaCB(filas []filaCB) string {
	s := "**Casos borde.**\n\n| Categoría | Aplica | Prueba o motivo |\n|---|---|---|\n"
	for _, f := range filas {
		s += "| " + f.categoria + " | " + f.aplica + " | " + f.texto + " |\n"
	}
	return s
}

// conSliceActivo: conFichas más una página para el slice activo E001/F001/S002 con la tabla dada ("" = sin tabla).
func conSliceActivo(t *testing.T, estado, tabla string) string {
	t.Helper()
	raiz := conFichas(t)
	escribir(t, raiz, "README.md", "- [E001](docs/E001/README.md)\n")
	ficha(t, raiz, backlog.Ficha{ID: "E001/F001/S002", Titulo: "Lista", Estado: estado, Secuencia: 6})
	escribir(t, raiz, tS002, tPagina("EJ-E001-F001-S002", "continúa: [EJ-E001-F001](README.md#EJ-E001-F001)\n\n",
		mermaid("sequenceDiagram"), "- ↑ Subir: [F001](README.md)"))
	if tabla != "" {
		reemplazar(t, raiz, tS002, "Problema → solución.", "Problema → solución.\n\n"+tabla)
	}
	generar(t, raiz)
	return raiz
}

// TestPrevencionActivaGobernada (integral E001/F012): todo contrato nuevo trae casos borde.
// Rojo cuando: un contrato sin tabla de casos borde pasa el gate.
// Verde cuando: el gate lo rechaza y el contrato completo pasa.
func TestPrevencionActivaGobernada(t *testing.T) {
	exigeProblema(t, conSliceActivo(t, "BORRADOR", ""), tS002+": falta la tabla «Casos borde»")
	if p := verificar(t, conSliceActivo(t, "BORRADOR", tablaCB(filasCompletas()))); len(p) > 0 {
		t.Errorf("el contrato completo debió pasar; problemas: %q", p)
	}
}

// TestContratoExigeCasosBorde (E001/F012/S001): la tabla trae las 7 categorías, bien llenadas.
// Rojo cuando: falta una categoría y el gate da OK.
// Verde cuando: el gate nombra la categoría que falta.
func TestContratoExigeCasosBorde(t *testing.T) {
	cambiar := func(i int, f filaCB) []filaCB {
		filas := filasCompletas()
		filas[i] = f
		return filas
	}
	ultima := len(CategoriasCasosBorde) - 1
	casos := []struct {
		nombre, estado string
		filas          []filaCB // nil = sin tabla
		espera         []string
	}{
		{"sin tabla en BORRADOR", "BORRADOR", nil, []string{"falta la tabla «Casos borde»"}},
		{"sin tabla en EN_CURSO", "EN_CURSO", nil, []string{"falta la tabla «Casos borde»"}},
		{"sin tabla en BLOQUEADO", "BLOQUEADO", nil, []string{"falta la tabla «Casos borde»"}},
		{"falta una categoría", "BORRADOR", filasCompletas()[:ultima],
			[]string{"casos borde: falta la categoría «" + CategoriasCasosBorde[ultima] + "»"}},
		{"categoría mal escrita", "BORRADOR", cambiar(4, filaCB{"choque con otros controles", "sí", "`TestA`"}),
			[]string{"casos borde: categoría desconocida «choque con otros controles»",
				"casos borde: falta la categoría «" + CategoriasCasosBorde[4] + "»"}},
		{"categoría repetida", "BORRADOR", append(filasCompletas(), filasCompletas()[0]),
			[]string{"casos borde: categoría repetida «" + CategoriasCasosBorde[0] + "»"}},
		{"aplica raro", "BORRADOR", cambiar(1, filaCB{CategoriasCasosBorde[1], "quizás", "x"}),
			[]string{"casos borde: «" + CategoriasCasosBorde[1] + "»: Aplica debe ser sí o no"}},
		{"sí sin prueba", "BORRADOR", cambiar(0, filaCB{CategoriasCasosBorde[0], "sí", "lo cubre algo"}),
			[]string{"casos borde: «" + CategoriasCasosBorde[0] + "»: con sí nombra una prueba"}},
		{"no sin motivo", "BORRADOR", cambiar(1, filaCB{CategoriasCasosBorde[1], "no", " "}),
			[]string{"casos borde: «" + CategoriasCasosBorde[1] + "»: con no explica por qué"}},
		{"prueba de la tabla que no existe", "BORRADOR", cambiar(0, filaCB{CategoriasCasosBorde[0], "sí", "`TestFantasma`"}),
			[]string{"prueba nombrada no existe en el código: TestFantasma"}},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			tabla := ""
			if c.filas != nil {
				tabla = tablaCB(c.filas)
			}
			raiz := conSliceActivo(t, c.estado, tabla)
			for _, e := range c.espera {
				exigeProblema(t, raiz, tS002+": "+e)
			}
		})
	}

	// Hallazgos A1: la marca escondida no cuenta; encabezado exacto; «—» no es motivo.
	completa := tablaCB(filasCompletas())
	for _, c := range []struct{ nombre, tabla, espera string }{
		{"tabla dentro de un bloque de código", "```\n" + completa + "```\n", "falta la tabla «Casos borde»"},
		{"tabla dentro de un comentario HTML", "<!--\n" + completa + "-->\n", "falta la tabla «Casos borde»"},
		{"encabezado distinto", strings.Replace(completa, "| Categoría | Aplica | Prueba o motivo |", "| x | y | z |", 1),
			"casos borde: el encabezado debe ser «Categoría | Aplica | Prueba o motivo»"},
		{"barra escapada queda dentro de la celda", tablaCB(cambiar(0, filaCB{"cada tipo \\| de ficha", "sí", "`TestA`"})),
			"casos borde: categoría desconocida «cada tipo | de ficha»"},
		{"no con raya de campo vacío", tablaCB(cambiar(1, filaCB{CategoriasCasosBorde[1], "no", "—"})),
			"casos borde: «" + CategoriasCasosBorde[1] + "»: con no explica por qué"},
	} {
		t.Run(c.nombre, func(t *testing.T) {
			exigeProblema(t, conSliceActivo(t, "BORRADOR", c.tabla), tS002+": "+c.espera)
		})
	}

	t.Run("plantilla en código y enlace en la categoría pasan", func(t *testing.T) {
		filas := cambiar(1, filaCB{"cada [estado](README.md) de origen", "no", "un motivo"})
		tabla := "```\n" + tablaCB(nil) + "```\n\n" + tablaCB(filas)
		if p := verificar(t, conSliceActivo(t, "BORRADOR", tabla)); len(p) > 0 {
			t.Errorf("se esperaba gate limpio; problemas: %q", p)
		}
	})

	t.Run("las categorías del código son las de la regla 04", func(t *testing.T) {
		b, err := os.ReadFile(filepath.Join("..", "..", "gobierno", "canonico", "reglas", "04_CASOS_BORDE.md"))
		if err != nil {
			t.Fatal(err)
		}
		var regla []string
		for _, l := range strings.Split(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n") {
			if strings.HasPrefix(l, "  - ") {
				regla = append(regla, strings.TrimRight(strings.TrimPrefix(l, "  - "), ";."))
			}
		}
		if strings.Join(regla, "\n") != strings.Join(CategoriasCasosBorde, "\n") {
			t.Errorf("la regla 04 y CategoriasCasosBorde difieren:\nregla  %q\ncódigo %q", regla, CategoriasCasosBorde)
		}
	})

	t.Run("feature activa sin tabla pasa", func(t *testing.T) {
		raiz := conSliceActivo(t, "BORRADOR", completa)
		ficha(t, raiz, backlog.Ficha{ID: "E001/F002", Titulo: "Backlog", Estado: "EN_CURSO", Secuencia: 4})
		escribir(t, raiz, "docs/E001/F002/README.md", tPagina("EJ-E001-F002", "continúa: [EJ-E001](../README.md#EJ-E001)\n\n",
			mermaid("flowchart LR")+"\n"+mermaid("stateDiagram-v2"), "- ↑ Subir: [E001](../README.md)"))
		generar(t, raiz)
		if p := verificar(t, raiz); len(p) > 0 {
			t.Errorf("la página de feature activa no exige tabla; problemas: %q", p)
		}
	})

	t.Run("completa pasa; feature y slice archivado sin tabla también", func(t *testing.T) {
		// conFichas: S001 está CERRADO (archivado) y su página no trae tabla; F001 y E001 tampoco.
		if p := verificar(t, conSliceActivo(t, "EN_CURSO", tablaCB(filasCompletas()))); len(p) > 0 {
			t.Errorf("se esperaba gate limpio; problemas: %q", p)
		}
		if p := strings.Join(verificar(t, conFichas(t)), " | "); strings.Contains(p, "Casos borde") {
			t.Errorf("slice archivado, feature e iniciativa no exigen tabla; problemas: %q", p)
		}
	})
}
