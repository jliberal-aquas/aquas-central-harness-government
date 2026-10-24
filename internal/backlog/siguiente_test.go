package backlog

import (
	"strings"
	"testing"
)

// E001/F002/S008 (regla del humano, 2026-10-21): definir antes de ejecutar. Una iniciativa o
// feature activa sin hijos pendientes sale antes que cualquier slice libre, sin importar
// prioridades: iniciativas primero, luego features, cada grupo por orden de creación, aunque
// estén bloqueadas. El slice en curso no se interrumpe.
func TestSiguienteExigeDefinirAntes(t *testing.T) {
	raiz := base(t)
	pide := func(want string) {
		t.Helper()
		f, hay, err := Siguiente(raiz)
		if err != nil || !hay || f.ID != want {
			t.Fatalf("siguiente = %q (%v, %v); se esperaba %s", f.ID, hay, err, want)
		}
	}
	nuevo(t, raiz, "id=E001/F002/S001", "titulo=T", "prioridad=P0", "motivo=m")
	pide("E001/F002/S001") // todo definido → el slice

	nuevo(t, raiz, "id=E001/F003", "titulo=T", "prioridad=P3", "motivo=m")
	pide("E001/F003") // feature sin slices gana a un slice P0

	nuevo(t, raiz, "id=E002", "titulo=T", "prioridad=P3", "motivo=m")
	pide("E002") // iniciativa sin features va antes que la feature, aunque sea más nueva

	nuevo(t, raiz, "id=E002/F001", "titulo=T", "prioridad=P0", "motivo=m", "bloqueado_por=E001/F002")
	pide("E001/F003") // E002 ya tiene feature; entre features, la más vieja primero

	nuevo(t, raiz, "id=E001/F003/S001", "titulo=T", "prioridad=P3", "motivo=m")
	pide("E002/F001") // E001/F003 ya tiene slice; la bloqueada también exige definirse

	if m := Definir(Ficha{ID: "E002", Tipo: "iniciativa"}); !strings.Contains(m, "definir E002") || !strings.Contains(m, "features") || !strings.Contains(m, "reprioriza") {
		t.Errorf("mensaje de iniciativa: %q", m)
	}
	if m := Definir(Ficha{ID: "E002/F001", Tipo: "feature"}); !strings.Contains(m, "definir E002/F001") || !strings.Contains(m, "slices") {
		t.Errorf("mensaje de feature: %q", m)
	}
	if m := Definir(Ficha{ID: "E001/F002/S001", Tipo: "slice"}); m != "" {
		t.Errorf("un slice no se define: %q", m)
	}

	if _, err := arrancar(t, raiz, "E001/F002/S001"); err != nil {
		t.Fatal(err)
	}
	pide("E001/F002/S001") // el slice en curso no se interrumpe
}
