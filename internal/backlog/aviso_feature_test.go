package backlog

import (
	"strings"
	"testing"
)

// TestAvisoPideAprobarFeature (E001/F015/X01, D000069): el aviso de falta de aprobación nombra la feature
// cuando tiene dos o más slices abiertos.
// Rojo cuando: con dos slices abiertos el aviso pide «apruebo <slice>».
// Verde cuando: con dos o más abiertos pide «apruebo <feature> con push» y con uno solo pide el slice.
func TestAvisoPideAprobarFeature(t *testing.T) {
	raiz := t.TempDir()
	nueva := func(id string) {
		t.Helper()
		if _, err := Nuevo(raiz, []string{"id=" + id, "titulo=ficha " + id, "prioridad=P1", "motivo=prueba",
			"que_hace=hace algo", "relacion=nueva", "prueba=./internal/x:TestX=valida algo|rojo cuando falla|verde cuando pasa"}); err != nil {
			t.Fatal(err)
		}
	}
	nueva("E009")
	nueva("E009/F001")
	nueva("E009/F001/S001")

	// Un solo slice abierto: pide el slice.
	s1, _, _ := Ver(raiz, "E009/F001/S001")
	if err := aprobada(raiz, s1); err == nil || !strings.Contains(err.Error(), "«apruebo E009/F001/S001»") {
		t.Fatalf("un solo slice abierto: debe pedir el slice; dio %v", err)
	}
	// Dos abiertos: pide la feature con push, para todos.
	nueva("E009/F001/S002")
	err := aprobada(raiz, s1)
	if err == nil || !strings.Contains(err.Error(), "«apruebo E009/F001 con push»") || strings.Contains(err.Error(), "«apruebo E009/F001/S001»") {
		t.Fatalf("dos slices abiertos: debe pedir la feature con push; dio %v", err)
	}
	// Un hermano cancelado no cuenta como abierto.
	if _, err := Estado(raiz, "E009/F001/S002", "CANCELADO"); err != nil {
		t.Fatal(err)
	}
	if err := aprobada(raiz, s1); err == nil || !strings.Contains(err.Error(), "«apruebo E009/F001/S001»") {
		t.Fatalf("hermano cancelado: vuelve a pedir el slice; dio %v", err)
	}
}
