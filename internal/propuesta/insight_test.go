package propuesta

import "testing"

// TestInsightPropuestoExigeDefecto: un insight PROPUESTO en el cierre nombra un defecto D000NNN.
// Rojo cuando: el freno deja pasar un insight PROPUESTO sin número de defecto, o frena uno que lo nombra o que está implementado.
// Verde cuando: frena el PROPUESTO sin defecto y deja pasar el que nombra D000NNN y los implementados.
func TestInsightPropuestoExigeDefecto(t *testing.T) {
	casos := []struct {
		msg    string
		frena  bool
		motivo string
	}{
		{"💡 **Insights:** Aprendimos X. Para evitarlo, Y. Estado: PROPUESTO; evidencia: pendiente.", true, "propuesto sin defecto"},
		{"💡 **Insights:** Aprendimos X. Para evitarlo, Y. Estado: PROPUESTO; evidencia: D000003.", false, "propuesto con defecto"},
		{"💡 **Insights:** Aprendimos X. Estado: IMPLEMENTADO Y PROBADO; evidencia: TestA.", false, "implementado"},
		{"💡 **Insights:** Sin insight nuevo", false, "sin insight"},
		{"texto que menciona PROPUESTO fuera del campo", false, "fuera del campo"},
		// E001/F002/S016/X01 (D000009): la palabra citada no es el estado.
		{"💡 **Insights:** los insights «PROPUESTO» se perdían. Estado: IMPLEMENTADO Y PROBADO; evidencia: TestA.", false, "implementado que cita la palabra"},
		{"💡 **Insights:** Aprendimos X. Estado: propuesto; evidencia: pendiente.", true, "propuesto en minúsculas sin defecto"},
	}
	for _, c := range casos {
		if got := RevisarInsight(c.msg) != ""; got != c.frena {
			t.Errorf("%s: frena=%v, se esperaba %v", c.motivo, got, c.frena)
		}
	}
}
