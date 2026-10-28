package backlog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestAprobacionDeFeatureCubreSusSlices (E001/F015/S001): la aprobación de la feature cubre los slices de su foto.
// Rojo cuando: un slice de la foto exige apruebo propio, avanza con una prueba aprobada cambiada o quitada, o avanza un slice nuevo.
// Verde cuando: agregar pruebas o cambiar alcance mantiene la aprobación y cambiar o quitar una prueba aprobada la invalida.
func TestAprobacionDeFeatureCubreSusSlices(t *testing.T) {
	raiz := modulo(t)
	for _, rel := range []string{"docs/E007/README.md", "docs/E007/F001/README.md", "docs/E007/F001/S001.md", "docs/E007/F001/S002.md", "docs/E007/F002/README.md"} {
		ruta := filepath.Join(raiz, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(ruta, []byte("# x\n\n## Para qué existe\n\nContrato.\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	nuevo(t, raiz, "id=E007", "titulo=I", "prioridad=P1", "motivo=m")
	nuevo(t, raiz, "id=E007/F001", "titulo=F", "prioridad=P1", "motivo=m")
	nuevo(t, raiz, "id=E007/F002", "titulo=G", "prioridad=P1", "motivo=m")
	nuevo(t, raiz, "id=E007/F001/S001", "titulo=A", "prioridad=P1", "motivo=m", "prueba=./p:TestA=a|r|v")
	nuevo(t, raiz, "id=E007/F001/S002", "titulo=B", "prioridad=P1", "motivo=m", "prueba=./p:TestB=b|r|v")
	revisar := func(id string) error {
		t.Helper()
		f, _, err := Ver(raiz, id)
		if err != nil {
			t.Fatal(err)
		}
		return aprobada(raiz, f)
	}
	pruebas := func(id string, lista ...string) {
		t.Helper()
		if _, err := Pruebas(raiz, id, lista); err != nil {
			t.Fatal(err)
		}
	}
	if revisar("E007/F001/S001") == nil {
		t.Fatal("sin aprobación de nadie, el slice no avanza")
	}
	if _, err := Aprobar(raiz, "E007", "apruebo E007", hora); err == nil || !strings.Contains(err.Error(), "se aprueban slices o features") {
		t.Error("una iniciativa no se aprueba")
	}
	if _, err := Aprobar(raiz, "E007/F002", "apruebo E007/F002", hora); err == nil {
		t.Error("una feature sin slices no se aprueba")
	}
	f, err := Aprobar(raiz, "E007/F001", "apruebo E007/F001", hora)
	if err != nil {
		t.Fatal(err)
	}
	if f.Aprobacion.Push {
		t.Error("sin «con push» la aprobación no lleva push")
	}
	if f, err = Aprobar(raiz, "E007/F001", "apruebo E007/F001 tope 50 con push", hora); err != nil || !f.Aprobacion.Push {
		t.Fatalf("«con push» después de «tope N» también vale; dio %v", err)
	}
	if f, err = Aprobar(raiz, "E007/F001", "apruebo E007/F001 con push\nsigue", hora); err != nil || !f.Aprobacion.Push {
		t.Fatalf("«con push» queda en la aprobación; dio %v", err)
	}
	for _, id := range []string{"E007/F001/S001", "E007/F001/S002"} {
		if err := revisar(id); err != nil {
			t.Errorf("%s está en la foto: avanza sin apruebo propio; dio %v", id, err)
		}
	}
	pruebas("E007/F001/S001", "./p:TestA=a|r|v", "./p:TestC=c|r|v")
	if _, err := Alcance(raiz, "E007/F001/S001", []string{"p/**"}); err != nil {
		t.Fatal(err)
	}
	if err := revisar("E007/F001/S001"); err != nil {
		t.Errorf("agregar una prueba o cambiar el alcance mantiene la aprobación; dio %v", err)
	}
	pruebas("E007/F001/S001", "./p:TestA=otra intención|r|v", "./p:TestC=c|r|v")
	if err := revisar("E007/F001/S001"); err == nil || !strings.Contains(err.Error(), "apruebo E007/F001") {
		t.Errorf("cambiar una prueba aprobada pide «apruebo E007/F001»; dio %v", err)
	}
	pruebas("E007/F001/S002", "./p:TestC=c|r|v")
	if err := revisar("E007/F001/S002"); err == nil || !strings.Contains(err.Error(), "apruebo E007/F001") {
		t.Errorf("quitar una prueba aprobada pide «apruebo E007/F001»; dio %v", err)
	}
	if _, err := Aprobar(raiz, "E007/F001/S002", "apruebo E007/F001/S002", hora); err != nil {
		t.Fatal(err)
	}
	if err := revisar("E007/F001/S002"); err != nil {
		t.Errorf("la aprobación propia vigente vale aunque la foto ya no cuadre; dio %v", err)
	}
	nuevo(t, raiz, "id=E007/F001/S003", "titulo=C", "prioridad=P1", "motivo=m")
	if err := revisar("E007/F001/S003"); err == nil {
		t.Error("un slice creado después de la aprobación no avanza")
	}
	if f, err = Aprobar(raiz, "E007/F001", "apruebo E007/F001\ncon push", hora); err != nil || f.Aprobacion.Push {
		t.Errorf("«con push» fuera de la primera línea no cuenta; dio push=%v, %v", f.Aprobacion != nil && f.Aprobacion.Push, err)
	}
	if err := revisar("E007/F001/S003"); err != nil {
		t.Errorf("tras aprobar otra vez, S003 está en la foto; dio %v", err)
	}
	ruta := filepath.Join(raiz, "docs", "E007", "F001", "README.md")
	if err := os.WriteFile(ruta, []byte("# x\n\n## Para qué existe\n\nOtro contrato.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := revisar("E007/F001/S003"); err == nil || !strings.Contains(err.Error(), "apruebo E007/F001") {
		t.Errorf("cambiar el contrato de la feature pide «apruebo E007/F001»; dio %v", err)
	}
}
