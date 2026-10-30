package backlog

import (
	"os"
	"path/filepath"
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

	if m := Definir(raiz, Ficha{ID: "E002", Tipo: "iniciativa"}); !strings.Contains(m, "definir E002") || !strings.Contains(m, "features") || !strings.Contains(m, "reprioriza") {
		t.Errorf("mensaje de iniciativa: %q", m)
	}
	if m := Definir(raiz, Ficha{ID: "E002/F001", Tipo: "feature"}); !strings.Contains(m, "definir E002/F001") || !strings.Contains(m, "slices") {
		t.Errorf("mensaje de feature: %q", m)
	}
	if m := Definir(raiz, Ficha{ID: "E001/F002/S001", Tipo: "slice"}); m != "" {
		t.Errorf("un slice no se define: %q", m)
	}

	if _, err := arrancar(t, raiz, "E001/F002/S001"); err != nil {
		t.Fatal(err)
	}
	pide("E001/F002/S001") // el slice en curso no se interrumpe
}

// TestSiguientePideCerrarConHijosArchivados (E001/F002/S008/X01, D000064): una feature o iniciativa
// activa sin hijos activos y con algún hijo CERRADO en el archivo no está «sin definir»: siguiente pide cerrarla.
// Rojo cuando: el mensaje dice «definir» para la feature con todos sus slices archivados.
// Verde cuando: dice «cerrar» con esos hijos y sigue diciendo «definir» sin ningún hijo.
func TestSiguientePideCerrarConHijosArchivados(t *testing.T) {
	raiz := base(t)
	nuevo(t, raiz, "id=E001/F003", "titulo=T", "prioridad=P1", "motivo=m")
	s := nuevo(t, raiz, "id=E001/F003/S001", "titulo=T", "prioridad=P1", "motivo=m")
	s.Estado = "CERRADO"
	if err := escribir(raiz, Archivo, s); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(raiz, filepath.FromSlash(Activo), nombre(s.ID))); err != nil {
		t.Fatal(err)
	}
	f, _, err := Ver(raiz, "E001/F003")
	if err != nil {
		t.Fatal(err)
	}
	if m := Definir(raiz, f); !strings.Contains(m, "cerrar E001/F003") || strings.Contains(m, "definir") {
		t.Errorf("feature con sus slices archivados: pide cerrarla; dio %q", m)
	}
	if m := Definir(raiz, Ficha{ID: "E001/F004", Tipo: "feature"}); !strings.Contains(m, "definir E001/F004") {
		t.Errorf("feature sin ningún hijo: pide definirla; dio %q", m)
	}
	c := Ficha{ID: "E001/F005/S001", Tipo: "slice", Padre: "E001/F005", Estado: "CANCELADO", Titulo: "T", Prioridad: "P1", Motivo: "m"}
	if err := escribir(raiz, Archivo, c); err != nil {
		t.Fatal(err)
	}
	if m := Definir(raiz, Ficha{ID: "E001/F005", Tipo: "feature"}); !strings.Contains(m, "definir E001/F005") {
		t.Errorf("feature con solo slices cancelados: pide definirla; dio %q", m)
	}
	g := Ficha{ID: "E009/F001", Tipo: "feature", Padre: "E009", Estado: "CERRADO", Titulo: "T", Prioridad: "P1", Motivo: "m"}
	if err := escribir(raiz, Archivo, g); err != nil {
		t.Fatal(err)
	}
	if m := Definir(raiz, Ficha{ID: "E009", Tipo: "iniciativa"}); !strings.Contains(m, "cerrar E009") || !strings.Contains(m, "features") {
		t.Errorf("iniciativa con sus features cerradas: pide cerrarla; dio %q", m)
	}
}

// TestSiguienteNoExigeDefinirReincidenciasVacia (E000/F006/S003, decisión del humano 2026-10-30):
// la feature de reincidencias nace vacía y se llena sola con `aquas defecto`; vacía no traba siguiente.
// Rojo cuando: siguiente responde E000/F007 con la feature vacía.
// Verde cuando: la salta vacía y sigue pidiendo definir otras features sin slices.
func TestSiguienteNoExigeDefinirReincidenciasVacia(t *testing.T) {
	raiz := base(t)
	nuevo(t, raiz, "id=E001/F002/S001", "titulo=T", "prioridad=P0", "motivo=m")
	nuevo(t, raiz, "id=E000", "titulo=Mantenimiento", "prioridad=P1", "motivo=m")
	nuevo(t, raiz, "id=E000/F001", "titulo=Otra", "prioridad=P1", "motivo=m")
	nuevo(t, raiz, "id=E000/F001/S001", "titulo=T", "prioridad=P1", "motivo=m")
	nuevo(t, raiz, "id="+FeatureReincidencias, "titulo=Reincidencias", "prioridad=P1", "motivo=m")
	if f, hay, err := Siguiente(raiz); err != nil || !hay || f.ID != "E001/F002/S001" {
		t.Fatalf("reincidencias vacía no se exige definir: siguiente = %q (%v, %v)", f.ID, hay, err)
	}
	// Hallazgo A1.1: con sus slices CERRADOS en el archivo tampoco se pide cerrarla; es un
	// contenedor permanente (cerrada, `aquas defecto` ya no podría crear slices en ella).
	s := nuevo(t, raiz, "id="+FeatureReincidencias+"/S001", "titulo=T", "prioridad=P1", "motivo=m")
	s.Estado = "CERRADO"
	if err := escribir(raiz, Archivo, s); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(raiz, filepath.FromSlash(Activo), nombre(s.ID))); err != nil {
		t.Fatal(err)
	}
	if f, _, _ := Siguiente(raiz); f.ID != "E001/F002/S001" {
		t.Fatalf("reincidencias con slices archivados no se pide cerrar: siguiente = %q", f.ID)
	}
	// Hallazgo A1.3: otra feature vacía, con el mismo padre y la misma prioridad, sí se exige:
	// la excepción es solo por el ID.
	nuevo(t, raiz, "id=E000/F008", "titulo=T", "prioridad=P1", "motivo=m")
	if f, _, _ := Siguiente(raiz); f.ID != "E000/F008" {
		t.Errorf("otra feature vacía debe pedirse: siguiente = %q", f.ID)
	}
}
