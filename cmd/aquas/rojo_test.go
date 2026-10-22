package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// E001/F002/S003 sobre el binario real: memoria según el backlog; rojo, verde y cierre de punta a punta.

func fichaEnDisco(t *testing.T, raiz, rel string) map[string]json.RawMessage {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(raiz, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("leer %s: %v", rel, err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("JSON de %s: %v", rel, err)
	}
	return m
}

func TestMemoriaSegunBacklog(t *testing.T) {
	raiz := t.TempDir()
	for _, args := range [][]string{
		{"nuevo", "id=E001", "titulo=G", "prioridad=P1", "motivo=m", "prueba=./p:TestA=a|r|v"},
		{"nuevo", "id=E001/F002", "titulo=B", "prioridad=P1", "motivo=m", "prueba=./p:TestA=a|r|v"},
		{"nuevo", "id=E001/F002/S003", "titulo=R", "prioridad=P1", "motivo=m", "prueba=./p:TestA=a|r|v"},
	} {
		if s := aquas(t, raiz, "", append([]string{"backlog"}, args...)...); s.codigo != 0 {
			t.Fatalf("backlog %v: %+v", args, s)
		}
	}
	foco := []string{"memoria", "fijar", "iniciativa=E001", "feature=F002", "slice=S003", "tarea=t"}
	if s := aquas(t, raiz, "", foco...); s.codigo != 1 || !strings.Contains(s.err, "en curso") {
		t.Errorf("sin slice en curso, la memoria no fija S003: exit 1 y lo dice; dio %+v", s)
	}
	if s := aquas(t, raiz, "", "backlog", "estado", "E001/F002/S003", "EN_CURSO"); s.codigo != 0 {
		t.Fatalf("EN_CURSO: %+v", s)
	}
	otro := []string{"memoria", "fijar", "iniciativa=E001", "feature=F002", "slice=S004", "tarea=t"}
	if s := aquas(t, raiz, "", otro...); s.codigo != 1 || !strings.Contains(s.err, "E001/F002/S003") {
		t.Errorf("foco distinto del slice en curso: exit 1 y nombra el que está en curso; dio %+v", s)
	}
	if s := aquas(t, raiz, "", foco...); s.codigo != 0 {
		t.Errorf("foco igual al slice en curso: exit 0; dio %+v", s)
	}
	// Hallazgo del auditor: cerrado el slice sin archivar la memoria, el error dice cómo salir.
	if s := aquas(t, raiz, "", "backlog", "estado", "E001/F002/S003", "CANCELADO"); s.codigo != 0 {
		t.Fatalf("CANCELADO: %+v", s)
	}
	if s := aquas(t, raiz, "", "memoria", "fijar", "tarea=otra"); s.codigo != 1 || !strings.Contains(s.err, "memoria archivar") || !strings.Contains(s.err, "antes de cerrar") || !strings.Contains(s.err, "memoria fijar slice=") {
		t.Errorf("slice archivado en la memoria: exit 1 y dice cómo vaciarlo; dio %+v", s)
	}
	if s := aquas(t, raiz, "", "memoria", "fijar", "slice=", "tarea=otra"); s.codigo != 0 {
		t.Errorf("con el slice vaciado, la memoria fija: exit 0; dio %+v", s)
	}
}

// Prueba integral de F002: un slice nace con su prueba, ve su rojo, su verde, y cierra al archivo.
func TestCerrarSliceDePuntaAPunta(t *testing.T) {
	raiz := t.TempDir()
	escribir(t, raiz, "go.mod", "module juguete\n\ngo 1.26\n")
	escribir(t, raiz, "p/p_test.go", "package p\n\nimport (\n\t\"os\"\n\t\"testing\"\n)\n\n"+
		"func TestBandera(t *testing.T) {\n\tif _, err := os.Stat(\"arreglado\"); err != nil {\n\t\tt.Fatal(\"falta el arreglo\")\n\t}\n}\n")
	paso := func(codigo int, want string, args ...string) {
		t.Helper()
		s := aquas(t, raiz, "", append([]string{"backlog"}, args...)...)
		if s.codigo != codigo || !strings.Contains(s.out+s.err, want) {
			t.Fatalf("backlog %v: se esperaba exit %d con %q; dio %+v", args, codigo, want, s)
		}
	}
	paso(0, "E001", "nuevo", "id=E001", "titulo=G", "prioridad=P1", "motivo=m", "prueba=./p:TestBandera=pide el arreglo|r|v")
	paso(0, "E001/F002", "nuevo", "id=E001/F002", "titulo=B", "prioridad=P1", "motivo=m", "prueba=./p:TestBandera=pide el arreglo|r|v")
	paso(0, "E001/F002/S003", "nuevo", "id=E001/F002/S003", "titulo=R", "prioridad=P1", "motivo=m", "prueba=./p:TestOtra=otra|r|v")
	paso(0, "TestBandera · pide el arreglo", "pruebas", "E001/F002/S003", "./p:TestBandera=pide el arreglo|r|v")
	paso(0, "EN_CURSO", "estado", "E001/F002/S003", "EN_CURSO")
	paso(1, "sin rojo", "estado", "E001/F002/S003", "CERRADO")
	paso(1, "rojo", "verde", "E001/F002/S003")
	paso(0, "rojo: p_test.go", "rojo", "E001/F002/S003")
	escribir(t, raiz, "p/arreglado", "")
	paso(0, "· verde 20", "verde", "E001/F002/S003")
	sellarCLI(t, raiz, "E001/F002/S003") // S004: el cierre exige el sello marcado
	paso(0, "archivada", "estado", "E001/F002/S003", "CERRADO")
	f := fichaEnDisco(t, raiz, "trabajo/archivo/E001-F002-S003.json")
	if !strings.Contains(string(f["rojo"]), "falta el arreglo") || !strings.Contains(string(f["verde"]), "fecha") {
		t.Errorf("la ficha archivada guarda el rojo con su motivo y el verde: rojo %s verde %s", f["rojo"], f["verde"])
	}
	for _, args := range [][]string{{"backlog", "rojo"}, {"backlog", "verde"}, {"backlog", "pruebas", "E001/F002/S003"}} {
		if s := aquas(t, raiz, "", args...); s.codigo != 64 {
			t.Errorf("%v: uso malo = exit 64; dio %+v", args, s)
		}
	}
}
