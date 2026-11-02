package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDesafioEnPedidoNuevo (E000/F010, integral): el gancho de cierre real lee el último mensaje humano del registro
// y frena un pedido nuevo cuya respuesta no trae DESAFÍO; un «dale» pasa sin bloque.
// Rojo cuando: el pedido nuevo sin bloque pasa o el «dale» frena.
// Verde cuando: solo frena el pedido nuevo sin bloque completo.
func TestDesafioEnPedidoNuevo(t *testing.T) {
	raiz := t.TempDir()
	reg := filepath.Join(t.TempDir(), "sesion.jsonl")
	humano := func(texto string) {
		t.Helper()
		b, _ := json.Marshal(map[string]any{"type": "user", "timestamp": "2026-11-02T12:00:00Z", "origin": map[string]string{"kind": "human"},
			"message": map[string]any{"role": "user", "content": texto}})
		if err := os.WriteFile(reg, append(b, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	activo := false
	cierre := func(msg string) salida {
		t.Helper()
		b, _ := json.Marshal(map[string]any{"hook_event_name": "Stop", "last_assistant_message": msg, "transcript_path": reg, "stop_hook_active": activo})
		return aquas(t, raiz, string(b), "proposal")
	}
	humano("quiero que el backlog tenga etiquetas")
	if s := cierre("Listo, agregué etiquetas."); s.codigo != 2 || !strings.Contains(s.err, "DESAFÍO") {
		t.Errorf("pedido nuevo sin bloque: frena nombrando DESAFÍO; dio %+v", s)
	}
	// Hallazgo A1.3: con otra falla del cierre (PROPUESTA sin diagrama), el primer freno nombra DESAFÍO.
	if s := cierre("PROPUESTA: hacer X."); s.codigo != 2 || !strings.Contains(s.err, "DESAFÍO") {
		t.Errorf("dos fallas: el primer freno debe ser DESAFÍO, que el turno de corrección ya no revisa; dio %+v", s)
	}
	// Hallazgo A1.4: el turno de corrección (stop_hook_active) no se revisa por CLI.
	activo = true
	if s := cierre("Corregido."); s.codigo != 0 {
		t.Errorf("stop_hook_active: no se revisa; dio %+v", s)
	}
	activo = false
	bloque := "DESAFÍO\n- Supuestos: x\n- Contraargumentos: sin objeción real\n- Riesgos: y\n- Recomendación: z\n"
	if s := cierre(bloque); s.codigo != 0 {
		t.Errorf("pedido nuevo con bloque completo: pasa; dio %+v", s)
	}
	humano("dale")
	if s := cierre("Hecho."); s.codigo != 0 {
		t.Errorf("«dale»: pasa sin bloque; dio %+v", s)
	}
}
