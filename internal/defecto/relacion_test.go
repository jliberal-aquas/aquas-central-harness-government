package defecto

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDefectoNuevoExigeRelacion: defecto nuevo lista defectos y fichas parecidos y exige relacion.
// Rojo cuando: registra el defecto sin relacion, no nombra el defecto o la ficha parecida, acepta repetir un defecto existente o apuntar a algo que no existe.
// Verde cuando: exige relacion con las parecidas a la vista, remite el repetido a ocurre y guarda la relación declarada.
func TestDefectoNuevoExigeRelacion(t *testing.T) {
	raiz := t.TempDir()
	base := map[string]string{"firma": "csv-acentos", "que": "exportar a CSV pierde acentos", "causa": "c", "arreglo": "a", "cercioro": "x", "relacion": "nuevo"}
	if _, err := Nuevo(raiz, base); err != nil {
		t.Fatalf("relacion=nuevo debió registrar: %v", err)
	}
	dir := filepath.Join(raiz, "trabajo", "activo")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(map[string]any{"id": "E001/F002/S001", "tipo": "slice", "titulo": "Exportar el backlog a CSV", "estado": "BORRADOR",
		"prioridad": "P1", "motivo": "m", "padre": "E001/F002", "bloqueado_por": []string{}, "secuencia": 1})
	if err := os.WriteFile(filepath.Join(dir, "E001-F002-S001.json"), b, 0o644); err != nil {
		t.Fatal(err)
	}
	otro := func(rel string) map[string]string {
		c := map[string]string{"firma": "csv-comas", "que": "exportar CSV rompe comas", "causa": "c", "arreglo": "a", "cercioro": "x"}
		if rel != "" {
			c["relacion"] = rel
		}
		return c
	}
	for _, c := range []struct {
		nombre, rel string
		dice        []string
	}{
		{"sin relacion", "", []string{"relacion=", "D000001", "E001/F002/S001"}},
		{"valor inválido", "quizas", []string{"relacion"}},
		{"repite existente", "repite:D000001", []string{"defect occur D000001"}},
		{"repite inexistente", "repite:D000099", []string{"no existe"}},
		{"ficha inexistente", "E001/F002/S099", []string{"no existe"}},
	} {
		_, err := Nuevo(raiz, otro(c.rel))
		for _, d := range c.dice {
			if err == nil || !strings.Contains(err.Error(), d) {
				t.Errorf("%s: se esperaba error con %q; dio %v", c.nombre, d, err)
			}
		}
	}
	d, err := Nuevo(raiz, otro("E001/F002/S001"))
	if err != nil {
		t.Fatalf("con relacion a una ficha debió registrar: %v", err)
	}
	if g, err := leerUno(raiz, d.ID); err != nil || d.Relacion != "E001/F002/S001" || g.Relacion != d.Relacion {
		t.Errorf("relación no guardada: %q / %+v %v", d.Relacion, g, err)
	}
}
