package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestHookFrenaFalloSinDefecto (E001/F004/S002): el gancho Stop aplica el freno de fallo sin defecto.
// Rojo cuando: aquas propuesta sale 0 con un fallo sin defecto.
// Verde cuando: sale 2 con el aviso y sale 0 con No o con un defecto que existe.
func TestHookFrenaFalloSinDefecto(t *testing.T) {
	raiz := t.TempDir()
	escribir(t, raiz, "registro/defectos/D000001.json", `{"id":"D000001"}`)
	cierre := func(msg string) salida {
		t.Helper()
		b, _ := json.Marshal(map[string]string{"hook_event_name": "Stop", "last_assistant_message": msg})
		return aquas(t, raiz, string(b), "propuesta")
	}
	if s := cierre("- 🤦‍♂️ **¿Salió algo mal?:** la prueba tenía un error"); s.codigo != 2 || !strings.Contains(s.err, "Fallo sin defecto") {
		t.Errorf("fallo sin defecto: exit 2 con el aviso; dio %+v", s)
	}
	if s := cierre("- 🤦‍♂️ **¿Salió algo mal?:** No"); s.codigo != 0 {
		t.Errorf("No: pasa; dio %+v", s)
	}
	if s := cierre("- 🤦‍♂️ **¿Salió algo mal?:** la prueba tenía un error (D000001)"); s.codigo != 0 {
		t.Errorf("con defecto que existe: pasa; dio %+v", s)
	}
	// Choque con otro control: el freno de insights sigue frenando aunque el campo de fallos pase.
	if s := cierre("- 💡 **Insights:** algo. Estado: PROPUESTO\n- 🤦‍♂️ **¿Salió algo mal?:** No"); s.codigo != 2 || !strings.Contains(s.err, "Insight PROPUESTO") {
		t.Errorf("insight PROPUESTO sin defecto: sigue frenando; dio %+v", s)
	}
}
