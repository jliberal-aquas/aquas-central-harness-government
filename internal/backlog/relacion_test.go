package backlog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNuevoExigeRelacion: backlog nuevo lista las parecidas y exige relacion; ampliar o corregir lo cerrado se niega.
// Rojo cuando: crea la ficha sin relacion, no nombra la ficha o el defecto parecido, acepta ampliar o corregir algo cerrado, o no guarda la relación.
// Verde cuando: exige relacion con las parecidas a la vista, niega lo cerrado con su salida y guarda la relación declarada.
func TestNuevoExigeRelacion(t *testing.T) {
	raiz := base(t)
	nuevo(t, raiz, "id=E001/F002/S001", "titulo=Exportar el backlog a CSV", "prioridad=P1", "motivo=m")
	pisar(t, raiz, Archivo, Ficha{ID: "E001/F003", Tipo: "feature", Titulo: "Guardia de rutas", Estado: "CERRADO", Prioridad: "P1",
		Motivo: "m", Padre: "E001", BloqueadoPor: []string{}, Secuencia: 9})
	if err := os.MkdirAll(filepath.Join(raiz, "registro", "defectos"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(raiz, "registro", "defectos", "D000007.json"), []byte(`{"id":"D000007","que":"exportar a CSV pierde acentos"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	ok := []string{"id=E001/F002/S002", "titulo=Exportar backlog en CSV", "prioridad=P1", "motivo=m", "que_hace=hace algo", "prueba=./p:TestFalla=f|r|v"}
	casos := []struct {
		nombre, relacion string
		dice             []string
	}{
		{"sin relacion", "", []string{"relacion=", "E001/F002/S001", "D000007"}},
		{"valor inválido", "relacion=quizas", []string{"relacion"}},
		{"amplia inexistente", "relacion=amplia:E001/F002/S099", []string{"no existe"}},
		{"amplia cerrada", "relacion=amplia:E001/F003", []string{"feature nueva"}},
		{"corrige cerrada", "relacion=corrige:E001/F003", []string{"backlog fix E001/F003"}},
	}
	for _, c := range casos {
		campos := append([]string{}, ok...)
		if c.relacion != "" {
			campos = append(campos, c.relacion)
		}
		_, err := Nuevo(raiz, campos)
		for _, d := range c.dice {
			if err == nil || !strings.Contains(err.Error(), d) {
				t.Errorf("%s: se esperaba error con %q; dio %v", c.nombre, d, err)
			}
		}
	}
	f, err := Nuevo(raiz, append(ok, "relacion=amplia:E001/F002/S001"))
	if err != nil {
		t.Fatalf("con relacion debió nacer: %v", err)
	}
	if g := leerFicha(t, activo(raiz, "E001-F002-S002.json")); f.Relacion != "amplia:E001/F002/S001" || g.Relacion != f.Relacion {
		t.Errorf("relación no guardada: %q / %q", f.Relacion, g.Relacion)
	}
}
