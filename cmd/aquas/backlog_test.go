package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// E001/F002/S002: los cinco comandos de `aquas backlog` sobre el binario real.
func TestBacklogCLI(t *testing.T) {
	raiz := t.TempDir()
	paso := func(want string, args ...string) {
		t.Helper()
		s := aquas(t, raiz, "", append([]string{"backlog"}, args...)...)
		if s.codigo != 0 || !strings.Contains(s.out, want) {
			t.Fatalf("backlog %v: se esperaba exit 0 con %q; dio %+v", args, want, s)
		}
	}
	// E000/F005/S003: cada paso se comprueba leyendo la ficha en disco, no solo la salida.
	campo := func(rel, k, quiere string) {
		t.Helper()
		if v := strings.Trim(string(fichaEnDisco(t, raiz, rel)[k]), `"`); v != quiere {
			t.Errorf("%s: %s = %q; se esperaba %q", rel, k, v, quiere)
		}
	}
	const s2 = "trabajo/activo/E001-F002-S002.json"
	paso("E001", "nuevo", "id=E001", "titulo=Gobierno", "prioridad=P1", "motivo=raíz", "relacion=nueva", "prueba=./p:TestA=a|r|v")
	campo("trabajo/activo/E001.json", "titulo", "Gobierno")
	paso("E001/F002", "nuevo", "id=E001/F002", "titulo=Backlog", "prioridad=P1", "motivo=orden", "relacion=nueva", "prueba=./p:TestA=a|r|v")
	campo("trabajo/activo/E001-F002.json", "padre", "E001")
	rojoIntegral(t, raiz, []string{"nuevo", "id=E001/F002"})
	paso("E001/F002/S002", "nuevo", "id=E001/F002/S002", "titulo=Backlog en JSON", "prioridad=P2", "motivo=m", "relacion=nueva", "prueba=./p:TestA=a|r|v")
	campo(s2, "estado", "BORRADOR")
	campo(s2, "prioridad", "P2")
	paso("E001/F002/S002 ·", "siguiente")
	paso("P0", "prioridad", "E001/F002/S002", "P0", "motivo=urgente")
	campo(s2, "prioridad", "P0")
	campo(s2, "motivo", "urgente")
	aprobarCLI(t, raiz, "E001/F002/S002")
	paso("EN_CURSO", "estado", "E001/F002/S002", "EN_CURSO")
	campo(s2, "estado", "EN_CURSO")
	paso("E001/F002/S002", "ver")
	paso("urgente", "ver", "E001/F002/S002")
	paso("archivada", "estado", "E001/F002/S002", "CANCELADO") // CERRADO exige rojo y verde: TestCerrarSliceDePuntaAPunta
	campo("trabajo/archivo/E001-F002-S002.json", "estado", "CANCELADO")
	if _, err := os.Stat(filepath.Join(raiz, filepath.FromSlash(s2))); !os.IsNotExist(err) {
		t.Errorf("la ficha archivada ya no está en trabajo/activo/: %v", err)
	}
	if s := aquas(t, raiz, "", "backlog", "siguiente"); s.codigo != 0 || !strings.Contains(s.out, "definir E001/F002") {
		t.Fatalf("sin slices pendientes: exit 0 y pide definir la feature (E001/F002/S008); dio %+v", s)
	}
	for _, args := range [][]string{{"backlog"}, {"backlog", "volar"}, {"backlog", "estado", "E001"}, {"backlog", "prioridad", "E001", "P1"}} {
		if s := aquas(t, raiz, "", args...); s.codigo != 64 {
			t.Errorf("%v: uso malo = exit 64; dio %+v", args, s)
		}
	}
	if s := aquas(t, raiz, "", "backlog", "nuevo", "id=E1"); s.codigo != 1 || !strings.Contains(s.err, "id") {
		t.Errorf("nuevo inválido: exit 1 con motivo; dio %+v", s)
	}
}
