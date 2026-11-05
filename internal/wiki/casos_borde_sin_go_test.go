package wiki

import (
	"strings"
	"testing"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/backlog"
)

// TestCasosBordeSinGo (E000/F015/S001): la tabla de casos borde acepta nombres ruta:nombre que la ficha declara,
// en un repo con pruebas.json.
// Rojo cuando: un nombre ruta:nombre de la ficha se rechaza.
// Verde cuando: se acepta, y un nombre inventado o mal formado se sigue rechazando.
func TestCasosBordeSinGo(t *testing.T) {
	const nom = "src/a.test.ts:rechaza_vacio"
	armar := func(t *testing.T, conJSON, conRojo bool, cita string) string {
		t.Helper()
		filas := filasCompletas()
		filas[0] = filaCB{CategoriasCasosBorde[0], "sí", "lo cubre `" + cita + "`"}
		raiz := conSliceActivo(t, "EN_CURSO", tablaCB(filas))
		f := backlog.Ficha{ID: "E001/F001/S002", Titulo: "Lista", Estado: "EN_CURSO", Secuencia: 6,
			Pruebas: []backlog.Prueba{{Nombre: nom, Intencion: "valida"}}}
		if conRojo {
			f.Rojo = &backlog.Registro{Fecha: "2026-11-05T10:00:00Z", Head: "abc", Resultado: map[string]string{nom: "falló"}}
		}
		ficha(t, raiz, f)
		if conJSON {
			escribir(t, raiz, backlog.RutaPruebas, `{"version": 1, "comando": "node {ruta} {nombre}"}`)
		}
		return raiz
	}
	limpio := func(t *testing.T, raiz string) {
		t.Helper()
		for _, p := range verificar(t, raiz) {
			if strings.Contains(p, nom) || strings.Contains(p, "con sí nombra una prueba") {
				t.Errorf("el nombre ruta:nombre de la ficha debió aceptarse; problema: %s", p)
			}
		}
	}

	t.Run("ficha sin rojo acepta su nombre", func(t *testing.T) { limpio(t, armar(t, true, false, nom)) })
	t.Run("ficha con rojo acepta su nombre", func(t *testing.T) { limpio(t, armar(t, true, true, nom)) })
	t.Run("nombre inventado se rechaza", func(t *testing.T) {
		exigeProblema(t, armar(t, true, true, "src/a.test.ts:inventada"), "prueba nombrada no existe en el código: src/a.test.ts:inventada")
	})
	t.Run("nombre mal formado no cuenta como prueba", func(t *testing.T) {
		exigeProblema(t, armar(t, true, true, "src/a.test.ts:"), "con sí nombra una prueba")
	})
	t.Run("sin pruebas.json solo valen Test de Go", func(t *testing.T) {
		exigeProblema(t, armar(t, false, true, nom), "con sí nombra una prueba")
	})
	t.Run("ficha ilegible no acepta el nombre", func(t *testing.T) {
		raiz := armar(t, true, true, nom)
		escribir(t, raiz, "trabajo/activo/E001-F001-S002.json", "{")
		if declarada(raiz, tS002, nom) {
			t.Error("con la ficha ilegible el nombre no debió aceptarse")
		}
		if _, err := Verificar(raiz); err == nil {
			t.Error("con la ficha ilegible el gate debió frenar")
		}
	})
}
