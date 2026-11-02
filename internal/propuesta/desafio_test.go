package propuesta

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// linea: una línea del registro de Claude Code (forma observada el 2026-11-02).
func linea(tipo, origen, hora string, contenido any) string {
	m := map[string]any{"type": tipo, "timestamp": hora, "message": map[string]any{"role": tipo, "content": contenido}}
	if origen != "" {
		m["origin"] = map[string]string{"kind": origen}
	}
	b, _ := json.Marshal(m)
	return string(b)
}

func registro(t *testing.T, lineas ...string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "sesion.jsonl")
	if err := os.WriteFile(p, []byte(strings.Join(lineas, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// TestRevisarDesafio (E000/F010/S001): ante el último mensaje humano del registro, un pedido nuevo exige el bloque
// DESAFÍO completo; un mensaje operativo, un turno de corrección o un registro sin mensaje humano no se revisan.
// Rojo cuando: un pedido nuevo sin bloque pasa o un «dale» frena.
// Verde cuando: solo el pedido nuevo sin bloque completo frena.
func TestRevisarDesafio(t *testing.T) {
	const h1, h2 = "2026-11-02T10:00:00Z", "2026-11-02T11:00:00Z"
	nuevo := linea("user", "human", h1, "quiero etiquetas en el backlog")
	completo := "Resumen.\n\n**DESAFÍO**\n- Supuestos: a\n- Contraargumentos: sin objeción real\n- Riesgos: b\n- Recomendación: c\n"
	sinBloque := "Listo, agregué etiquetas."

	// Pedido nuevo sin bloque: frena y nombra las cuatro partes.
	m := RevisarDesafio(registro(t, nuevo), sinBloque, false)
	for _, p := range []string{"DESAFÍO", "Supuestos", "Contraargumentos", "Riesgos", "Recomendación"} {
		if !strings.Contains(m, p) {
			t.Errorf("pedido nuevo sin bloque: debe frenar nombrando %s; dio %q", p, m)
		}
	}
	// Bloque completo, con «sin objeción real»: pasa.
	if m := RevisarDesafio(registro(t, nuevo), completo, false); m != "" {
		t.Errorf("bloque completo: pasa; dio %q", m)
	}
	// Bloque con una palabra de menos: frena nombrando solo esa.
	if m, _, _ := strings.Cut(RevisarDesafio(registro(t, nuevo), strings.Replace(completo, "Riesgos", "Peligros", 1), false), ". Agrega"); !strings.Contains(m, "Riesgos") || strings.Contains(m, "Supuestos") {
		t.Errorf("falta Riesgos: frena nombrando solo Riesgos; dio %q", m)
	}
	// Las partes antes de la línea DESAFÍO no cuentan.
	if m := RevisarDesafio(registro(t, nuevo), "Supuestos, Contraargumentos, Riesgos, Recomendación.\nDESAFÍO\n", false); m == "" {
		t.Error("partes antes de la línea DESAFÍO: frena")
	}
	// Operativos: pasan sin bloque.
	for _, op := range []string{"dale", "Sí", "A, dale, sigue", "ok, aplica", "apruebo E000/F010 con push", "continúa", "Apruebo", "apruebo.", "Apruebo, continua con lo que sigue y lo demás"} {
		if m := RevisarDesafio(registro(t, linea("user", "human", h1, op)), sinBloque, false); m != "" {
			t.Errorf("operativo %q: pasa; dio %q", op, m)
		}
	}
	// Más de 6 palabras con «dale»: es pedido nuevo.
	if m := RevisarDesafio(registro(t, linea("user", "human", h1, "Si, dale. Aparte, quiero que como directriz me desafíes")), sinBloque, false); m == "" {
		t.Error("mensaje largo con «dale»: es pedido nuevo y frena")
	}
	// Choque con otro control: el turno que corrige un freno del cierre no se vuelve a frenar.
	if m := RevisarDesafio(registro(t, nuevo), sinBloque, true); m != "" {
		t.Errorf("stop_hook_active: no se revisa; dio %q", m)
	}
	// Entrada ausente o vacía: sin registro, registro vacío o mensaje humano vacío → no se revisa.
	if m := RevisarDesafio("", sinBloque, false); m != "" {
		t.Errorf("sin registro: no se revisa; dio %q", m)
	}
	if m := RevisarDesafio(filepath.Join(t.TempDir(), "no.jsonl"), sinBloque, false); m != "" {
		t.Errorf("registro ausente: no se revisa; dio %q", m)
	}
	if m := RevisarDesafio(registro(t, linea("user", "human", h1, "   ")), sinBloque, false); m != "" {
		t.Errorf("mensaje humano vacío: no se revisa; dio %q", m)
	}
	// Quién habla: avisos de subagente, notificaciones, frenos del cierre y resultados de herramienta no cuentan.
	otros := []string{
		linea("user", "peer", h2, "Another Claude session sent a message:\n<agent-message from=\"a1\">\n[Subagent hand-back] x"),
		linea("user", "task-notification", h2, "<task-notification> listo"),
		linea("user", "", h2, "Stop hook feedback: falta algo"),
		linea("user", "", h2, []map[string]string{{"type": "tool_result", "content": "salida"}}),
		linea("user", "human", h2, "<task-notification> x </task-notification>"),
		linea("assistant", "", h2, []map[string]string{{"type": "text", "text": "dale"}}),
	}
	// Hallazgo A1.1: un turno que abre un agente (aviso de subagente o de tarea) no se revisa contra el pedido viejo.
	for _, o := range otros[:2] {
		if m := RevisarDesafio(registro(t, nuevo, o), sinBloque, false); m != "" {
			t.Errorf("turno abierto por un agente tras un pedido nuevo: no se revisa; dio %q", m)
		}
	}
	// A2.1: un aviso de agente sin marca también abre un turno que no es del humano: manda el origen.
	if m := RevisarDesafio(registro(t, nuevo, linea("user", "peer", h2, "un aviso sin marca")), sinBloque, false); m != "" {
		t.Errorf("aviso peer sin marca: el turno no es del humano; dio %q", m)
	}
	if m := RevisarDesafio(registro(t, nuevo, otros[4]), sinBloque, false); m != "" {
		t.Errorf("mensaje con marca de agente abre un turno que no es del humano; dio %q", m)
	}
	// Resultados de herramienta, avisos de ganchos y texto del asistente no abren turno: el pedido nuevo sigue mandando.
	if m := RevisarDesafio(registro(t, nuevo, otros[2], otros[3], otros[5]), sinBloque, false); m == "" {
		t.Error("líneas sin origen no tapan el pedido nuevo del humano: frena")
	}
	// Hallazgo A1.8: la línea DESAFÍO puede ir en viñeta.
	if m := RevisarDesafio(registro(t, nuevo), strings.Replace(completo, "**DESAFÍO**", "- **DESAFÍO**", 1), false); m != "" {
		t.Errorf("«- **DESAFÍO**» en viñeta: pasa; dio %q", m)
	}
	if m := RevisarDesafio(registro(t, otros...), sinBloque, false); m != "" {
		t.Errorf("registro sin mensaje humano: no se revisa; dio %q", m)
	}
	// Mensaje humano en bloques de texto: cuenta.
	if m := RevisarDesafio(registro(t, linea("user", "human", h1, []map[string]string{{"type": "text", "text": "dale"}})), sinBloque, false); m != "" {
		t.Errorf("«dale» en bloques de texto: pasa; dio %q", m)
	}
	// D000096: el último es el de hora más reciente, aunque Claude Code lo haya escrito antes en el archivo.
	if m := RevisarDesafio(registro(t, linea("user", "human", h2, "dale"), nuevo), sinBloque, false); m != "" {
		t.Errorf("el «dale» de hora más reciente manda aunque esté antes en el archivo; dio %q", m)
	}
	// Fallo a mitad de camino: una línea rota se salta, no tumba la revisión.
	if m := RevisarDesafio(registro(t, nuevo, `{"type":"user","origin":{"kind":"human"`), sinBloque, false); m == "" {
		t.Error("línea rota: se salta y el pedido nuevo frena")
	}
}
