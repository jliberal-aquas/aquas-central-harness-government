package backlog

import (
	"os"
	"strings"
	"testing"
)

// TestBacklogSinRepetidos: integral de E001/F009; el backlog frena fichas repetidas y deja cambiar lo que aún no arranca.
// Rojo cuando: una ficha parecida nace sin relacion declarada o sin nombrar la parecida, se amplía algo cerrado, se edita algo en curso o la edición no guarda lo anterior.
// Verde cuando: las tres reglas se cumplen de punta a punta.
func TestBacklogSinRepetidos(t *testing.T) {
	raiz := base(t)
	nuevo(t, raiz, "id=E001/F002/S001", "titulo=Exportar el backlog a CSV", "prioridad=P1", "motivo=reportes", "relacion=nueva")
	pisar(t, raiz, Archivo, Ficha{ID: "E001/F003", Tipo: "feature", Titulo: "Guardia de rutas", Estado: "CERRADO", Prioridad: "P1",
		Motivo: "m", Padre: "E001", BloqueadoPor: []string{}, Secuencia: 9})
	prueba := "prueba=./p:TestFalla=falla a propósito|r|v"

	// 1. Una ficha parecida sin relacion se niega y nombra la parecida.
	_, err := Nuevo(raiz, []string{"id=E001/F002/S002", "titulo=Exportar backlog en CSV", "prioridad=P1", "motivo=reportes", prueba})
	if err == nil || !strings.Contains(err.Error(), "relacion") || !strings.Contains(err.Error(), "E001/F002/S001") {
		t.Errorf("parecida sin relacion: debió negarse nombrando E001/F002/S001; dio %v", err)
	}
	// 2. Ampliar algo cerrado se niega y remite a una feature nueva.
	_, err = Nuevo(raiz, []string{"id=E001/F002/S002", "titulo=Guardia también en Bash", "prioridad=P1", "motivo=m", "relacion=amplia:E001/F003", prueba})
	if err == nil || !strings.Contains(err.Error(), "feature nueva") {
		t.Errorf("ampliar lo cerrado: debió remitir a feature nueva; dio %v", err)
	}
	// 3. Con relacion declarada nace.
	if _, err := Nuevo(raiz, []string{"id=E001/F002/S002", "titulo=Exportar backlog en CSV", "prioridad=P1", "motivo=reportes", "relacion=amplia:E001/F002/S001", prueba}); err != nil {
		t.Fatalf("con relacion debió nacer: %v", err)
	}
	// 4. Editar antes de arrancar guarda lo anterior y el porqué.
	if _, err := Editar(raiz, "E001/F002/S002", []string{"titulo=Exportar backlog en CSV y JSON", "motivo=también JSON"}, hora); err != nil {
		t.Fatalf("editar en BORRADOR: %v", err)
	}
	b, err := os.ReadFile(activo(raiz, "E001-F002-S002.json"))
	if err != nil {
		t.Fatal(err)
	}
	if s := string(b); !strings.Contains(s, "Exportar backlog en CSV y JSON") || !strings.Contains(s, `"Exportar backlog en CSV"`) || !strings.Contains(s, "también JSON") {
		t.Errorf("la edición no guardó el valor nuevo, el anterior o el porqué:\n%s", s)
	}
	// 5. Editar algo en curso se niega.
	enCurso(t, raiz, "E001/F002/S001", "./p:TestFalla=falla|r|v")
	if _, err := Editar(raiz, "E001/F002/S001", []string{"titulo=otro", "motivo=x"}, hora); err == nil {
		t.Error("editar EN_CURSO: debió negarse")
	}
}
