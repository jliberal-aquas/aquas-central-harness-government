package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// E000/F014 (D000125): integral de los tres huecos por los que un hijo perdió aprobaciones y trabajó por fuera de aquas.
// Rojo cuando: un slice nace sin página, el gancho de aprobación falla mudo (exit 1) o un commit de código pasa sin slice EN_CURSO.
// Verde cuando: la página nace con la ficha y el apruebo se registra; el fallo llega al agente con su causa; el commit
// fuera de alcance se rechaza nombrando la ruta.
func TestControlesSinHuecos(t *testing.T) {
	raiz := t.TempDir()
	escribir(t, raiz, "go.mod", "module juguete\n\ngo 1.26\n")
	escribir(t, raiz, "p/p_test.go", "package p\n\nimport \"testing\"\n\nfunc TestBandera(t *testing.T) { t.Fatal(\"falta el arreglo\") }\n")
	for _, args := range [][]string{
		{"nuevo", "id=E001", "titulo=G", "prioridad=P1", "motivo=m", "relacion=nueva", "que_hace=hace algo", "prueba=./p:TestBandera=pide el arreglo|r|v"},
		{"nuevo", "id=E001/F002", "titulo=B", "prioridad=P1", "motivo=m", "relacion=nueva", "que_hace=hace algo", "prueba=./p:TestBandera=pide el arreglo|r|v"},
		{"nuevo", "id=E001/F002/S001", "titulo=R", "prioridad=P1", "motivo=el hijo lo pidió", "relacion=nueva", "que_hace=valida algo", "prueba=./p:TestBandera=pide el arreglo|r|v"},
	} {
		if s := aquas(t, raiz, "", append([]string{"backlog"}, args...)...); s.codigo != 0 {
			t.Fatalf("backlog %v: %+v", args, s)
		}
		rojoIntegral(t, raiz, args)
	}
	if s := aquas(t, raiz, "", "backlog", "alcance", "E001/F002/S001", "p/**"); s.codigo != 0 {
		t.Fatalf("alcance: %+v", s)
	}
	pagina := filepath.Join(raiz, "docs", "E001", "F002", "S001.md")

	// Hueco 1 (S001): la página del contrato nace con la ficha.
	b, err := os.ReadFile(pagina)
	if err != nil || !strings.Contains(string(b), "## Para qué existe") {
		t.Fatalf("hueco 1: backlog new dejó el slice sin página de contrato (err=%v)", err)
	}

	// Hueco 2 (S002): sin página, el apruebo no se registra y el agente lo ve con la causa; nunca exit 1.
	if err := os.Rename(pagina, pagina+".bak"); err != nil {
		t.Fatal(err)
	}
	s := aquas(t, raiz, mensajeHumano(t, "apruebo E001/F002/S001"), "aprobacion")
	if s.codigo != 0 || !strings.Contains(s.out, "aprobación NO registrada") || !strings.Contains(s.out, "página") {
		t.Fatalf("hueco 2: el fallo de aprobación debe salir 0 con aviso al agente y su causa; dio %+v", s)
	}
	if err := os.Rename(pagina+".bak", pagina); err != nil {
		t.Fatal(err)
	}
	contratoDePrueba(t, raiz, "docs/E001/F002/S001.md", "E001/F002/S001") // el agente completa las filas PENDIENTE
	if s := aquas(t, raiz, mensajeHumano(t, "apruebo E001/F002/S001"), "aprobacion"); s.codigo != 0 {
		t.Fatalf("apruebo con página: %+v", s)
	}
	if _, hay := fichaEnDisco(t, raiz, "trabajo/activo/E001-F002-S001.json")["aprobacion"]; !hay {
		t.Fatal("hueco 1: con la página nacida, el apruebo debe quedar registrado")
	}

	// Hueco 3 (S003): un commit con código fuera de un slice EN_CURSO se rechaza nombrando la ruta.
	git(t, raiz, "init", "-q")
	escribir(t, raiz, "src/validez.ts", "export const x = 1\n")
	git(t, raiz, "add", "src/validez.ts")
	if s := aquas(t, raiz, "", "verificar"); s.codigo == 0 || !strings.Contains(s.out+s.err, "src/validez.ts: ningún slice EN_CURSO lo cubre") {
		t.Fatalf("hueco 3: código preparado sin slice EN_CURSO debe frenar nombrando la ruta; dio %+v", s)
	}
}
