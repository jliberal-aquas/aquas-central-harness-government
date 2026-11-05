package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

var idsAgente atomic.Int64

// lineaAgente: texto del agente en el registro, con su id y usage (forma observada el 2026-11-04).
func lineaAgente(texto string) string {
	b, _ := json.Marshal(map[string]any{"type": "assistant", "timestamp": time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
		"message": map[string]any{"id": fmt.Sprintf("msg_agente_%d", idsAgente.Add(1)), "role": "assistant",
			"content": []any{map[string]any{"type": "text", "text": texto}}, "usage": map[string]int{"output_tokens": 1}}})
	return string(b)
}

// lineaEnCola: mensaje que el humano escribe mientras el agente trabaja; Claude Code lo guarda como adjunto
// queued_command (forma observada el 2026-11-04 en el registro de la sesión b84c39b9).
func lineaEnCola(texto string) string { return lineaEnColaDe("human", texto) }

// lineaEnColaDe: mensaje en cola con su origen; prompt es texto o bloques (con imagen, observado el 2026-11-03).
func lineaEnColaDe(origen string, prompt any) string {
	b, _ := json.Marshal(map[string]any{"type": "attachment", "timestamp": time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
		"attachment": map[string]any{"type": "queued_command", "prompt": prompt, "commandMode": "prompt",
			"origin": map[string]string{"kind": origen}}})
	return string(b)
}

// aprobacionCruda: la aprobación guardada en una ficha, tal cual; "" sin aprobación.
func aprobacionCruda(t *testing.T, raiz, ficha string) string {
	t.Helper()
	return string(fichaEnDisco(t, raiz, ficha)["aprobacion"])
}

// mensajeSinRegistro: entrada del hook UserPromptSubmit sin dejar el mensaje en el registro.
func mensajeSinRegistro(t *testing.T, prompt string) string {
	t.Helper()
	b, _ := json.Marshal(map[string]string{"session_id": "s1", "hook_event_name": "UserPromptSubmit", "prompt": prompt, "transcript_path": registroSesion(t)})
	return string(b)
}

// mensajeFeature: el mensaje guardado en la aprobación de la feature E001/F002; "" sin aprobación.
func mensajeFeature(t *testing.T, raiz string) string {
	t.Helper()
	b, hay := fichaEnDisco(t, raiz, "trabajo/activo/E001-F002.json")["aprobacion"]
	if !hay {
		return ""
	}
	var ap struct{ Mensaje string }
	if err := json.Unmarshal(b, &ap); err != nil {
		t.Fatal(err)
	}
	return ap.Mensaje
}

// TestAprobacionCorta (E000/F013/S003): una respuesta corta aprueba la única línea «apruebo …» que el agente
// propuso en su último turno; un apruebo escrito en cola también vale.
// Rojo cuando: pide repetir la línea.
// Verde cuando: la ficha queda aprobada con el id, el push y el tope propuestos.
func TestAprobacionCorta(t *testing.T) {
	const propuesta = "apruebo E001/F002 con push tope 2M"
	t.Run("sí aprueba la propuesta con push y tope", func(t *testing.T) {
		raiz, reg, _, usar := presupuestoTokens(t)
		anexar(t, reg,
			lineaAgente("DECISIÓN: falta tu aprobación.\n\n`"+propuesta+"`\n\nResponde «sí»."),
			lineaAgente("Corrijo el cierre: listo.")) // la corrección del gancho de cierre no borra la propuesta
		if s := aquas(t, raiz, mensajeHumano(t, "sí"), "aprobacion"); s.codigo != 0 {
			t.Fatalf("«sí»: %+v", s)
		}
		if m := mensajeFeature(t, raiz); m != propuesta {
			t.Fatalf("«sí» aprueba la línea propuesta; mensaje guardado %q", m)
		}
		if ap := aprobacionFeature(t, raiz); !ap.Push {
			t.Error("«con push» de la propuesta queda en la ficha")
		}
		if s := aquas(t, raiz, "", "backlog", "estado", "E001/F002/S003", "EN_CURSO"); s.codigo != 0 {
			t.Fatalf("el slice arranca con la aprobación de la feature dada con «sí»: %+v", s)
		}
		anexar(t, reg, lineaUso("m1", 2_000_001, 0, 0, 0, time.Now()))
		if s := usar("ls"); s.codigo != 2 || !strings.Contains(s.err, "2M") {
			t.Errorf("el tope propuesto (2M) llega al presupuesto; dio %+v", s)
		}
	})
	t.Run("variantes cortas", func(t *testing.T) {
		for _, r := range []string{"si", "Dale.", "ejecútalo", "ejecuta", "apruebo", "OK!"} {
			t.Run(r, func(t *testing.T) { // registro propio por caso
				raiz, reg, _, _ := presupuestoTokens(t)
				anexar(t, reg, lineaAgente(propuesta))
				aquas(t, raiz, mensajeHumano(t, r), "aprobacion")
				if m := mensajeFeature(t, raiz); m != propuesta {
					t.Errorf("aprueba la propuesta; mensaje guardado %q", m)
				}
			})
		}
	})
	t.Run("no aprueba", func(t *testing.T) {
		casos := []struct{ nombre, agente, respuesta string }{
			{"sin propuesta", "Listo, sigo.", "sí"},
			{"dos propuestas", "apruebo E001/F002\n\napruebo E001/F002/S003", "dale"},
			{"respuesta larga", propuesta, "sí, pero cambia el tope"},
			{"sí de agente", propuesta, "<agent-message from=\"x\">sí</agent-message>"},
			{"push ofrecido como opción", "🧙‍♂️ **¿Qué necesito de ti?:** apruebo E001/F002, o dime si lo prefieres con push", "sí"},
			{"dos opciones en una línea", "`apruebo E001/F002` o `apruebo E001/F002 con push tope 9M`", "sí"},
		}
		for _, c := range casos {
			t.Run(c.nombre, func(t *testing.T) { // registro propio por caso
				raiz, reg, _, _ := presupuestoTokens(t)
				anexar(t, reg, lineaAgente(c.agente))
				aquas(t, raiz, mensajeHumano(t, c.respuesta), "aprobacion")
				if m := mensajeFeature(t, raiz); m != "" {
					t.Errorf("no aprueba nada; guardó %q", m)
				}
				if a := aprobacionCruda(t, raiz, "trabajo/activo/E001-F002-S003.json"); a != "" {
					t.Errorf("tampoco aprueba el slice; guardó %s", a)
				}
			})
		}
		raiz, reg, _, _ := presupuestoTokens(t)
		anexar(t, reg, lineaAgente(propuesta))
		aquas(t, raiz, `{"hook_event_name":"PostToolUse","prompt":"sí","transcript_path":"`+strings.ReplaceAll(reg, `\`, `\\`)+`"}`, "aprobacion")
		if m := mensajeFeature(t, raiz); m != "" {
			t.Errorf("una herramienta no aprueba; guardó %q", m)
		}
	})
	t.Run("sí a otra pregunta no vuelve a aprobar", func(t *testing.T) {
		for _, escrito := range []bool{false, true} { // el prompt aún no está en el registro, o ya está (A1, hallazgo 1)
			t.Setenv("AQUAS_PRUEBA_REGISTRO", "") // registro propio por caso
			raiz, reg, _, _ := presupuestoTokens(t)
			anexar(t, reg, lineaAgente(propuesta))
			aquas(t, raiz, mensajeHumano(t, "sí"), "aprobacion")
			antes := aprobacionCruda(t, raiz, "trabajo/activo/E001-F002.json")
			if antes == "" {
				t.Fatal("el primer «sí» aprueba la propuesta")
			}
			escribir(t, raiz, "docs/E001/F002/README.md", "# F002\n\n## Para qué existe\n\nContrato cambiado por el agente.\n")
			anexar(t, reg, lineaAgente("Ajusté el contrato. ¿Sigo?"))
			entrada := mensajeSinRegistro(t, "sí")
			if escrito {
				entrada = mensajeHumano(t, "sí")
			}
			aquas(t, raiz, entrada, "aprobacion")
			if despues := aprobacionCruda(t, raiz, "trabajo/activo/E001-F002.json"); despues != antes {
				t.Errorf("escrito=%v: un «sí» sin propuesta nueva no vuelve a aprobar; antes %s, después %s", escrito, antes, despues)
			}
		}
	})
	t.Run("propuesta en el campo de pedido del cierre", func(t *testing.T) {
		raiz, reg, _, _ := presupuestoTokens(t)
		anexar(t, reg, lineaAgente("🧙‍♂️ **¿Qué necesito de ti?:** `"+propuesta+"`.\n🎯 **Próximo movimiento:** seguir."))
		aquas(t, raiz, mensajeHumano(t, "dale"), "aprobacion")
		if m := mensajeFeature(t, raiz); m != propuesta {
			t.Errorf("«dale» aprueba la propuesta del campo de pedido; guardó %q", m)
		}
	})
	t.Run("sí aprueba un slice", func(t *testing.T) {
		raiz, reg, _, _ := presupuestoTokens(t)
		anexar(t, reg, lineaAgente("apruebo E001/F002/S003"))
		aquas(t, raiz, mensajeHumano(t, "sí"), "aprobacion")
		if s := aquas(t, raiz, "", "backlog", "state", "E001/F002/S003", "EN_CURSO"); s.codigo != 0 {
			t.Errorf("el slice aprobado con «sí» arranca; dio %+v", s)
		}
		if mensajeFeature(t, raiz) != "" {
			t.Error("aprobar el slice no aprueba la feature")
		}
	})
	t.Run("en cola de otro origen no respalda", func(t *testing.T) {
		raiz, reg, _, usar := presupuestoTokens(t)
		otro := strings.Replace(lineaEnColaDe("task-notification", "apruebo E001/F002 con push tope 1M"), `"attachment":`, `"usage":{},"attachment":`, 1)
		anexar(t, reg, otro, lineaEnColaDe("task-notification", propuesta))
		aquas(t, raiz, mensajeSinRegistro(t, propuesta), "aprobacion")
		if s := aquas(t, raiz, "", "backlog", "state", "E001/F002/S003", "EN_CURSO"); s.codigo == 0 {
			t.Error("un mensaje en cola que no es del humano no respalda la aprobación")
		}
		anexar(t, reg, lineaUso("m1", 1_500_000, 0, 0, 0, time.Now()))
		if s := usar("ls"); s.codigo != 0 { // A2, hallazgo 3: su «tope 1M» no llega al presupuesto
			t.Errorf("el tope de un mensaje en cola de otro origen no rige; dio %+v", s)
		}
	})
	t.Run("en cola con imagen vale", func(t *testing.T) {
		raiz, reg, _, _ := presupuestoTokens(t)
		anexar(t, reg, lineaEnColaDe("human", []any{map[string]string{"type": "text", "text": propuesta}, map[string]string{"type": "image"}}))
		aquas(t, raiz, mensajeSinRegistro(t, propuesta), "aprobacion")
		if s := aquas(t, raiz, "", "backlog", "state", "E001/F002/S003", "EN_CURSO"); s.codigo != 0 {
			t.Errorf("el apruebo en cola con imagen respalda la aprobación; dio %+v", s)
		}
	})
	t.Run("apruebo en cola vale", func(t *testing.T) {
		raiz, reg, _, usar := presupuestoTokens(t)
		anexar(t, reg, lineaEnCola(propuesta))
		if s := aquas(t, raiz, mensajeSinRegistro(t, propuesta), "aprobacion"); s.codigo != 0 {
			t.Fatalf("apruebo en cola: %+v", s)
		}
		if s := aquas(t, raiz, "", "backlog", "estado", "E001/F002/S003", "EN_CURSO"); s.codigo != 0 {
			t.Fatalf("el apruebo en cola respalda la aprobación: %+v", s)
		}
		anexar(t, reg, lineaUso("m1", 2_000_001, 0, 0, 0, time.Now()))
		if s := usar("ls"); s.codigo != 2 || !strings.Contains(s.err, "2M") {
			t.Errorf("el tope del apruebo en cola llega al presupuesto; dio %+v", s)
		}
	})
	t.Run("registro sin el mensaje pide la línea", func(t *testing.T) {
		raiz, reg, _, _ := presupuestoTokens(t)
		anexar(t, reg, lineaAgente(propuesta))
		aquas(t, raiz, mensajeSinRegistro(t, "sí"), "aprobacion") // el «sí» no quedó en el registro
		if s := aquas(t, raiz, "", "backlog", "estado", "E001/F002/S003", "EN_CURSO"); s.codigo == 0 || !strings.Contains(s.err, "apruebo E001/F002") {
			t.Errorf("sin el «sí» en el registro: rechaza y pide «apruebo E001/F002»; dio %+v", s)
		}
	})
}
