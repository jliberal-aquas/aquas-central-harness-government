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
	paso("E001", "nuevo", "id=E001", "titulo=Gobierno", "prioridad=P1", "motivo=raíz", "prueba=./p:TestA=a")
	paso("E001/F002", "nuevo", "id=E001/F002", "titulo=Backlog", "prioridad=P1", "motivo=orden", "prueba=./p:TestA=a")
	paso("E001/F002/S002", "nuevo", "id=E001/F002/S002", "titulo=Backlog en JSON", "prioridad=P2", "motivo=m", "prueba=./p:TestA=a")
	paso("E001/F002/S002", "siguiente")
	paso("P0", "prioridad", "E001/F002/S002", "P0", "motivo=urgente")
	paso("EN_CURSO", "estado", "E001/F002/S002", "EN_CURSO")
	paso("E001/F002/S002", "ver")
	paso("urgente", "ver", "E001/F002/S002")
	paso("archivada", "estado", "E001/F002/S002", "CANCELADO") // CERRADO exige rojo y verde: TestCerrarSliceDePuntaAPunta
	if _, err := os.Stat(filepath.Join(raiz, "trabajo", "archivo", "E001-F002-S002.json")); err != nil {
		t.Fatalf("la ficha cerrada vive en trabajo/archivo/: %v", err)
	}
	if s := aquas(t, raiz, "", "backlog", "siguiente"); s.codigo != 0 || !strings.Contains(s.out, "nada") {
		t.Fatalf("sin slices pendientes: exit 0 y lo dice; dio %+v", s)
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
