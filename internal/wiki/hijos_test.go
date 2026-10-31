package wiki

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/backlog"
)

// E001/F002/S010: navegación por niveles. Cada página lista solo a sus hijos directos;
// la lista la genera aquas desde las fichas trabajo/ y el gate falla si se desfasa.

// ficha escribe una ficha en trabajo/activo (o trabajo/archivo si está cerrada o cancelada).
func ficha(t *testing.T, raiz string, f backlog.Ficha) {
	t.Helper()
	partes := strings.Split(f.ID, "/")
	f.Tipo = []string{"iniciativa", "feature", "slice"}[len(partes)-1]
	f.Padre = strings.Join(partes[:len(partes)-1], "/")
	if f.Prioridad == "" {
		f.Prioridad = "P1"
	}
	if f.Motivo == "" {
		f.Motivo = "m"
	}
	f.BloqueadoPor = []string{}
	dir := "trabajo/activo"
	if f.Estado == "CERRADO" || f.Estado == "CANCELADO" {
		dir = "trabajo/archivo"
	}
	b, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	escribir(t, raiz, dir+"/"+strings.ReplaceAll(f.ID, "/", "-")+".json", string(b)+"\n")
}

// conFichas: la wiki válida de wiki_test.go más fichas para E001 › F001 › S001 y hermanos sin página.
func conFichas(t *testing.T) string {
	t.Helper()
	raiz := wikiValida(t)
	reg := func(r map[string]string) *backlog.Registro {
		return &backlog.Registro{Fecha: "2026-10-21T10:00:00Z", Head: "abc", Resultado: r}
	}
	ficha(t, raiz, backlog.Ficha{ID: "E001", Titulo: "Gobierno", Estado: "EN_CURSO", Secuencia: 1})
	ficha(t, raiz, backlog.Ficha{ID: "E002", Titulo: "Otra | rara", Estado: "BORRADOR", Secuencia: 2})
	ficha(t, raiz, backlog.Ficha{ID: "E001/F001", Titulo: "Guardia", Estado: "CERRADO", Secuencia: 3})
	ficha(t, raiz, backlog.Ficha{ID: "E001/F002", Titulo: "Backlog", Estado: "BORRADOR", Secuencia: 4})
	ficha(t, raiz, backlog.Ficha{ID: "E001/F001/S001", Titulo: "Regla", Estado: "CERRADO", Secuencia: 5,
		Pruebas: []backlog.Prueba{{Nombre: "./x:TestA", Intencion: "bloquea fuera"}, {Nombre: "./x:TestB", Intencion: "deja pasar"}},
		Rojo:    reg(map[string]string{"./x:TestA": "falló"}), Verde: reg(nil)})
	ficha(t, raiz, backlog.Ficha{ID: "E001/F001/S002", Titulo: "Lista", Estado: "BORRADOR", Secuencia: 6})
	// Las pruebas que el bloque del slice nombra existen en el código (trazabilidad).
	escribir(t, raiz, "x/x_test.go", "package x\n\nimport \"testing\"\n\nfunc TestA(t *testing.T) {}\n\nfunc TestB(t *testing.T) {}\n")
	return raiz
}

func generar(t *testing.T, raiz string) []string {
	t.Helper()
	escritos, err := Generar(raiz)
	if err != nil {
		t.Fatal(err)
	}
	return escritos
}

const aviso = "<!-- GENERADO por aquas desde trabajo/: no editar a mano; `aquas compile` lo rehace. -->\n\n"

func TestHijosGeneradosDesdeFichas(t *testing.T) {
	raiz := conFichas(t)
	escribir(t, raiz, "README.md", "- [E001](docs/E001/README.md)\n")
	if e := generar(t, raiz); len(e) != 4 {
		t.Errorf("primera generación escribe inicio, E001, F001 y S001; escribió %q", e)
	}
	quiere := map[string]string{
		tInicio: aviso + "| Hijo | Qué es | Estado |\n|---|---|---|\n" +
			"| [E001](E001/README.md) | Gobierno | En curso |\n" +
			"| E002 | Otra \\| rara | Planeado |\n\n<!-- /GENERADO -->",
		tE001: aviso + "| Hijo | Qué es | Estado |\n|---|---|---|\n" +
			"| [F001](F001/README.md) | Guardia | Hecho |\n" +
			"| F002 | Backlog | Planeado |\n\n<!-- /GENERADO -->",
		tF001: aviso + "| Hijo | Qué es | Estado |\n|---|---|---|\n" +
			"| [S001](S001.md) | Regla | Hecho |\n" +
			"| S002 | Lista | Planeado |\n\n<!-- /GENERADO -->",
		tS001: aviso + "| Prueba | Valida | Rojo cuando | Verde cuando | Rojo | Verde |\n|---|---|---|---|---|---|\n" +
			"| `TestA` | bloquea fuera | — | — | 2026-10-21 | 2026-10-21 |\n" +
			"| `TestB` | deja pasar | — | — | — | 2026-10-21 |\n" +
			"\n### Código\n\nSin código registrado: sus cambios no tocaron código ni configuración, o no nombran el ID (`aquas backlog code E001/F001/S001 <commit>…` los asigna).\n" + // D000022
			"\n<!-- /GENERADO -->",
	}
	seccionDe := map[string]string{tInicio: "## Navegación", tE001: "## Navegación", tF001: "## Navegación", tS001: "## Contrato y evidencia"}
	for rel, bloque := range quiere {
		c := leer(t, raiz, rel)
		if !strings.Contains(c, seccionDe[rel]+"\n\n"+bloque+"\n") {
			t.Errorf("%s: falta el bloque justo bajo %q:\n%s\n--- página:\n%s", rel, seccionDe[rel], bloque, c)
		}
		if strings.Count(c, "<!-- GENERADO") != 1 {
			t.Errorf("%s: un solo bloque GENERADO", rel)
		}
	}
	if p := verificar(t, raiz); len(p) > 0 {
		t.Errorf("tras generar, gate OK; problemas: %q", p)
	}
	if e := generar(t, raiz); len(e) != 0 {
		t.Errorf("segunda generación sin cambios no escribe nada; escribió %q", e)
	}
}

func TestHijosDesfasadosFallan(t *testing.T) {
	casos := []struct {
		nombre string
		mutar  func(t *testing.T, raiz string)
		espera string
	}{
		{"ficha cambió de estado", func(t *testing.T, r string) {
			ficha(t, r, backlog.Ficha{ID: "E001/F002", Titulo: "Backlog", Estado: "LISTO", Secuencia: 4})
			ficha(t, r, backlog.Ficha{ID: "E002", Titulo: "Otra | rara", Estado: "BLOQUEADO", Secuencia: 2})
		}, "docs/README.md: hijos desfasados con trabajo/ (corre `aquas compile`)"},
		{"fila editada a mano", func(t *testing.T, r string) {
			reemplazar(t, r, tE001, "| Guardia | Hecho |", "| Guardia | En curso |")
		}, "docs/E001/README.md: hijos desfasados con trabajo/ (corre `aquas compile`)"},
		{"pruebas del slice editadas", func(t *testing.T, r string) {
			reemplazar(t, r, tS001, "| — | 2026-10-21 |", "| 2026-10-21 | 2026-10-21 |")
		}, "docs/E001/F001/S001.md: hijos desfasados con trabajo/ (corre `aquas compile`)"},
		{"bloque borrado", func(t *testing.T, r string) {
			c := leer(t, r, tF001)
			i, j := strings.Index(c, "<!-- GENERADO"), strings.Index(c, "<!-- /GENERADO -->")
			if i < 0 || j < 0 {
				t.Fatalf("no hay bloque GENERADO que borrar:\n%s", c)
			}
			escribir(t, r, tF001, c[:i]+c[j+len("<!-- /GENERADO -->"):])
		}, "docs/E001/F001/README.md: falta el bloque GENERADO de hijos (corre `aquas compile`)"},
		{"bloque sin cierre", func(t *testing.T, r string) {
			reemplazar(t, r, tE001, "<!-- /GENERADO -->", "")
		}, "docs/E001/README.md: bloque GENERADO sin cierre <!-- /GENERADO -->"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			raiz := conFichas(t)
			generar(t, raiz)
			c.mutar(t, raiz)
			exigeProblema(t, raiz, c.espera)
		})
	}
	// El estado del ejemplo: la iniciativa nueva bloqueada se ve en el inicio al regenerar.
	raiz := conFichas(t)
	ficha(t, raiz, backlog.Ficha{ID: "E002", Titulo: "Otra | rara", Estado: "BLOQUEADO", Secuencia: 2})
	generar(t, raiz)
	if c := leer(t, raiz, tInicio); !strings.Contains(c, "| E002 | Otra \\| rara | Bloqueado |\n") {
		t.Errorf("BLOQUEADO se muestra «Bloqueado»:\n%s", c)
	}
}

func TestInicioMuestraSoloIniciativas(t *testing.T) {
	raiz := conFichas(t)
	generar(t, raiz)
	if c := leer(t, raiz, tInicio); strings.Contains(c, "F001") || strings.Contains(c, "S001") {
		t.Errorf("el bloque del inicio no nombra features ni slices:\n%s", c)
	}
	casos := []struct{ rel, viejo, nuevo, espera string }{
		{tInicio, "Algo simple.", "Ver [F001](E001/F001/README.md).",
			"docs/README.md: enlace salta niveles → docs/E001/F001/README.md (cada página enlaza solo hasta sus hijos directos)"},
		{tInicio, "Algo simple.", "Ver [S001](E001/F001/S001.md#ej).",
			"docs/README.md: enlace salta niveles → docs/E001/F001/S001.md (cada página enlaza solo hasta sus hijos directos)"},
		{tE001, "Algo simple.", "Ver [S001](F001/S001.md).",
			"docs/E001/README.md: enlace salta niveles → docs/E001/F001/S001.md (cada página enlaza solo hasta sus hijos directos)"},
	}
	for _, c := range casos {
		r := conFichas(t)
		generar(t, r)
		reemplazar(t, r, c.rel, c.viejo, c.nuevo)
		exigeProblema(t, r, c.espera)
	}
	// Hacia arriba o al lado sí: el slice enlaza al inicio y a su hermano.
	r := conFichas(t)
	generar(t, r)
	escribir(t, r, "docs/E001/F001/S002.md", strings.ReplaceAll(leer(t, r, tS001), "S001", "S002"))
	reemplazar(t, r, tS001, "Algo simple.", "Ver [Inicio](../../README.md) y [S002](S002.md).")
	for _, p := range verificar(t, r) {
		if strings.Contains(p, "salta niveles") {
			t.Errorf("subir o ir al hermano no salta niveles: %s", p)
		}
	}
}

func TestEntidadArrancadaSinPagina(t *testing.T) {
	for _, c := range []struct {
		estado, espera string
	}{
		{"EN_CURSO", "docs/E001/F002/README.md: falta la página de E001/F002 (ficha EN_CURSO)"},
		{"VERIFICANDO", "docs/E001/F002/README.md: falta la página de E001/F002 (ficha VERIFICANDO)"},
		{"CERRADO", "docs/E001/F002/README.md: falta la página de E001/F002 (ficha CERRADO)"},
	} {
		raiz := conFichas(t)
		if err := os.Remove(filepath.Join(raiz, "trabajo", "activo", "E001-F002.json")); err != nil {
			t.Fatal(err)
		}
		ficha(t, raiz, backlog.Ficha{ID: "E001/F002", Titulo: "Backlog", Estado: c.estado, Secuencia: 4})
		generar(t, raiz)
		exigeProblema(t, raiz, c.espera)
	}
	for _, estado := range []string{"BORRADOR", "LISTO", "BLOQUEADO", "CANCELADO"} {
		raiz := conFichas(t)
		if err := os.Remove(filepath.Join(raiz, "trabajo", "activo", "E001-F002.json")); err != nil {
			t.Fatal(err)
		}
		ficha(t, raiz, backlog.Ficha{ID: "E001/F002", Titulo: "Backlog", Estado: estado, Secuencia: 4})
		generar(t, raiz)
		for _, p := range verificar(t, raiz) {
			if strings.Contains(p, "falta la página") {
				t.Errorf("%s no ha arrancado: no exige página; dio %s", estado, p)
			}
		}
	}
}
