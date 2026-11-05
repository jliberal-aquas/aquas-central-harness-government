package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/wiki"
)

// arbolWiki: inicio, glosario y páginas padre E001 y E001/F002 tomadas de la wiki real (E000 y E000/F014),
// para que el gate wiki analice la página del slice dentro de un árbol válido.
func arbolWiki(t *testing.T, raiz string) {
	t.Helper()
	real := filepath.Join("..", "..", "docs")
	copiar := func(desde, hacia string, cambiar func(string) string) {
		b, err := os.ReadFile(filepath.Join(real, filepath.FromSlash(desde)))
		if err != nil {
			t.Fatal(err)
		}
		escribir(t, raiz, hacia, cambiar(string(b)))
	}
	igual := func(s string) string { return s }
	aE001 := strings.NewReplacer("E000/F014", "E001/F002", "EJ-E000-F014", "EJ-E001-F002", "F014", "F002", "E000", "E001").Replace
	copiar("README.md", "docs/README.md", igual)
	copiar("glosario.md", "docs/glosario.md", igual)
	copiar("E000/README.md", "docs/E001/README.md", aE001)
	copiar("E000/F014/README.md", "docs/E001/F002/README.md", aE001)
}

// TestSliceNaceConContrato (E000/F014/S001, D000125): backlog new de un slice crea su página de contrato.
// Rojo cuando: backlog new deja un slice sin página.
// Verde cuando: la página existe con «Para qué existe», pasa el gate wiki, no pisa una escrita y el apruebo se registra.
func TestSliceNaceConContrato(t *testing.T) {
	nuevo := func(t *testing.T, raiz, id string) salida {
		t.Helper()
		return aquas(t, raiz, "", "backlog", "nuevo", "id="+id, "titulo=T "+id, "prioridad=P1", "motivo=el hijo lo pidió",
			"relacion=nueva", "que_hace=valida algo", "prueba=./p:TestBandera=pide el arreglo|r|v")
	}
	padres := func(t *testing.T, raiz string) {
		t.Helper()
		escribir(t, raiz, "go.mod", "module juguete\n\ngo 1.26\n")
		escribir(t, raiz, "p/p_test.go", "package p\n\nimport \"testing\"\n\nfunc TestBandera(t *testing.T) { t.Fatal(\"falta\") }\n")
		for _, id := range []string{"E001", "E001/F002"} {
			if s := nuevo(t, raiz, id); s.codigo != 0 {
				t.Fatalf("new %s: %+v", id, s)
			}
		}
	}
	existe := func(raiz, rel string) bool {
		_, err := os.Stat(filepath.Join(raiz, filepath.FromSlash(rel)))
		return err == nil
	}

	t.Run("slice nace con página; feature e iniciativa no", func(t *testing.T) {
		raiz := t.TempDir()
		padres(t, raiz)
		if existe(raiz, "docs/E001/README.md") || existe(raiz, "docs/E001/F002/README.md") {
			t.Fatal("feature e iniciativa no cambian: no crean página")
		}
		if s := nuevo(t, raiz, "E001/F002/S001"); s.codigo != 0 {
			t.Fatalf("new slice: %+v", s)
		}
		b, err := os.ReadFile(filepath.Join(raiz, "docs", "E001", "F002", "S001.md"))
		if err != nil || !strings.Contains(string(b), "## Para qué existe") || !strings.Contains(string(b), "el hijo lo pidió") {
			t.Fatalf("la página debe nacer con «Para qué existe» y el motivo; err=%v\n%s", err, b)
		}
		arbolWiki(t, raiz)
		if _, err := wiki.Generar(raiz); err != nil { // lo que hace `aquas compile` tras crear fichas
			t.Fatal(err)
		}
		docs, _, err := wiki.VerificarConAvisos(raiz)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range docs {
			if strings.Contains(d, "docs/E001/F002/S001.md") {
				t.Errorf("el gate wiki debe aceptar la página nacida: %s", d)
			}
		}
		if s := aquas(t, raiz, "", "backlog", "alcance", "E001/F002/S001", "p/**"); s.codigo != 0 {
			t.Fatalf("alcance: %+v", s)
		}
		for _, id := range []string{"E001/F002/S001", "E001/F002"} { // por el slice y por su feature (A2)
			s := aquas(t, raiz, mensajeHumano(t, "apruebo "+id), "aprobacion")
			if !strings.Contains(s.out+s.err, "docs/E001/F002/S001.md tiene filas PENDIENTE:") {
				t.Errorf("apruebo %s con filas PENDIENTE: el rechazo nombra la página; dio %+v", id, s)
			}
			for _, f := range []string{"E001-F002-S001", "E001-F002"} {
				if _, hay := fichaEnDisco(t, raiz, "trabajo/activo/"+f+".json")["aprobacion"]; hay {
					t.Fatalf("apruebo %s con filas PENDIENTE no se registra en %s (hallazgo A1)", id, f)
				}
			}
		}
		escribir(t, raiz, "docs/E001/F002/S009.md", "# S009\n\n## Para qué existe\n\nLa regla habla de filas `PENDIENTE:` sin tener ninguna.\n\n| Regla | Valor |\n|---|---|\n| apruebo | rechaza filas PENDIENTE: sin decidir |\n")
		if s := nuevo(t, raiz, "E001/F002/S009"); s.codigo != 0 {
			t.Fatalf("new S009: %+v", s)
		}
		if s := aquas(t, raiz, mensajeHumano(t, "apruebo E001/F002/S009"), "aprobacion"); s.codigo != 0 {
			t.Fatalf("apruebo S009: %+v", s)
		}
		if _, hay := fichaEnDisco(t, raiz, "trabajo/activo/E001-F002-S009.json")["aprobacion"]; !hay {
			t.Error("mencionar la marca en prosa o dentro de una celda no frena el apruebo (A2)")
		}
		contratoDePrueba(t, raiz, "docs/E001/F002/S001.md", "E001/F002/S001")
		if s := aquas(t, raiz, mensajeHumano(t, "apruebo E001/F002/S001"), "aprobacion"); s.codigo != 0 {
			t.Fatalf("apruebo: %+v", s)
		}
		if _, hay := fichaEnDisco(t, raiz, "trabajo/activo/E001-F002-S001.json")["aprobacion"]; !hay {
			t.Fatal("con el contrato completo, el apruebo queda registrado")
		}
	})

	t.Run("página ya escrita no se pisa", func(t *testing.T) {
		raiz := t.TempDir()
		padres(t, raiz)
		escribir(t, raiz, "docs/E001/F002/S001.md", paginaS003)
		if s := nuevo(t, raiz, "E001/F002/S001"); s.codigo != 0 {
			t.Fatalf("new slice: %+v", s)
		}
		if b, _ := os.ReadFile(filepath.Join(raiz, "docs", "E001", "F002", "S001.md")); string(b) != paginaS003 {
			t.Fatalf("la página escrita no se toca; quedó:\n%s", b)
		}
	})

	t.Run("sin poder escribir la página no nace la ficha", func(t *testing.T) {
		raiz := t.TempDir()
		padres(t, raiz)
		escribir(t, raiz, "docs/E001/F002", "un archivo donde va la carpeta") // MkdirAll falla
		if s := nuevo(t, raiz, "E001/F002/S001"); s.codigo == 0 || !strings.Contains(s.err, "página") {
			t.Fatalf("debe fallar nombrando la página; dio %+v", s)
		}
		if existe(raiz, "trabajo/activo/E001-F002-S001.json") {
			t.Fatal("sin página, la ficha no se crea")
		}
	})

	t.Run("en un hijo instalado también nace", func(t *testing.T) {
		raiz := t.TempDir()
		if s := aquas(t, raiz, "", "install"); s.codigo != 0 {
			t.Fatalf("install: %+v", s)
		}
		padres(t, raiz)
		if s := nuevo(t, raiz, "E001/F002/S001"); s.codigo != 0 {
			t.Fatalf("new slice en hijo: %+v", s)
		}
		if !existe(raiz, "docs/E001/F002/S001.md") {
			t.Fatal("en el hijo el slice también nace con página")
		}
	})
}
