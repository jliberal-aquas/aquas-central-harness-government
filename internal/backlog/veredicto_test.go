package backlog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// cierreAuditor: el CIERRE del auditor para ese slice con ese veredicto.
func cierreAuditor(id, veredicto string) string {
	return fmt.Sprintf("id: %s-A1\nveredicto: %s\nhallazgos: —\nno_revisado: —", strings.ReplaceAll(id, "/", "-"), veredicto)
}

// lineaAuditor: una línea de asistente del registro de un subagente (forma observada 2026-10-31): texto o
// mensaje de SubagentHandback, con attributionAgent y timestamp.
func lineaAuditor(t *testing.T, texto, agente string, cuando time.Time, handback bool) string {
	t.Helper()
	parte := map[string]any{"type": "text", "text": texto}
	if handback {
		parte = map[string]any{"type": "tool_use", "name": "SubagentHandback", "input": map[string]string{"message": texto}}
	}
	l := map[string]any{"type": "assistant", "isSidechain": true, "timestamp": cuando.UTC().Format("2006-01-02T15:04:05.000Z"),
		"message": map[string]any{"role": "assistant", "content": []any{parte}}}
	if agente != "" {
		l["attributionAgent"] = agente
	}
	b, err := json.Marshal(l)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// registroAuditor: deja un .jsonl de subagente en el CLAUDE_CONFIG_DIR de la prueba (el mismo de la aprobación)
// y devuelve su ruta.
func registroAuditor(t *testing.T, lineas ...string) string {
	t.Helper()
	return jsonlEn(t, filepath.Join(dirClaude(t), "projects", "repo", "sesion", "subagents"), "agent-*.jsonl", lineas)
}

// auditoriaRespaldada: un APROBADO del auditor para id, con hora fecha (RFC 3339) y su registro de auditor.
func auditoriaRespaldada(t *testing.T, id, fecha string) *Auditoria {
	t.Helper()
	cuando, err := time.Parse(time.RFC3339, fecha)
	if err != nil {
		t.Fatal(err)
	}
	reg := registroAuditor(t, lineaAuditor(t, cierreAuditor(id, "APROBADO"), "auditor", cuando, false))
	return &Auditoria{Fecha: fecha, Veredicto: "APROBADO", Registro: reg}
}

// TestVeredictoConRegistroDelAuditor (E000/F008/S004, D000077): un slice cierra solo si el registro del auditor
// trae su APROBADO, escrito por el auditor, después del verde, y sin un veredicto posterior que lo contradiga.
// Rojo cuando: un APROBADO sin registro deja cerrar.
// Verde cuando: solo cierra con el APROBADO del auditor después del verde.
func TestVeredictoConRegistroDelAuditor(t *testing.T) {
	despues := hora.Add(2 * time.Hour) // el verde de sliceConVerde es hora+1h
	falla := map[string]func(t *testing.T, raiz, id string) string{
		"sin registro": func(t *testing.T, raiz, id string) string { return "" },
		"archivo ausente": func(t *testing.T, raiz, id string) string {
			return filepath.Join(filepath.Dir(registroAuditor(t)), "no.jsonl")
		},
		"fuera de projects": func(t *testing.T, raiz, id string) string {
			otra := filepath.Join(t.TempDir(), "agent.jsonl")
			_ = os.WriteFile(otra, []byte(lineaAuditor(t, cierreAuditor(id, "APROBADO"), "auditor", despues, false)+"\n"), 0o644)
			return otra
		},
		"dentro del repo": func(t *testing.T, raiz, id string) string {
			ruta := filepath.Join(raiz, ".tmp", "agent.jsonl")
			_ = os.MkdirAll(filepath.Dir(ruta), 0o755)
			_ = os.WriteFile(ruta, []byte(lineaAuditor(t, cierreAuditor(id, "APROBADO"), "auditor", despues, false)+"\n"), 0o644)
			return ruta
		},
		"línea rota": func(t *testing.T, raiz, id string) string {
			return registroAuditor(t, `{"type":"assistant","attributionAgent":"auditor","message":{"content":[{"type":"text","text":"id: `)
		},
		"otro id": func(t *testing.T, raiz, id string) string {
			return registroAuditor(t, lineaAuditor(t, cierreAuditor("E001/F002/S009", "APROBADO"), "auditor", despues, false))
		},
		"lo escribió el orquestador": func(t *testing.T, raiz, id string) string {
			return registroAuditor(t, lineaAuditor(t, cierreAuditor(id, "APROBADO"), "", despues, false))
		},
		"otro tipo de agente": func(t *testing.T, raiz, id string) string {
			return registroAuditor(t, lineaAuditor(t, cierreAuditor(id, "APROBADO"), "worker-sonnet", despues, false))
		},
		"línea de usuario (el encargo)": func(t *testing.T, raiz, id string) string {
			l := strings.Replace(lineaAuditor(t, cierreAuditor(id, "APROBADO"), "auditor", despues, false), `"type":"assistant"`, `"type":"user"`, 1)
			return registroAuditor(t, l)
		},
		"APROBADO anterior al verde": func(t *testing.T, raiz, id string) string {
			return registroAuditor(t, lineaAuditor(t, cierreAuditor(id, "APROBADO"), "auditor", hora.Add(30*time.Minute), false))
		},
		"APROBADO y luego CON_HALLAZGOS": func(t *testing.T, raiz, id string) string {
			return registroAuditor(t,
				lineaAuditor(t, cierreAuditor(id, "APROBADO"), "auditor", despues, false),
				lineaAuditor(t, cierreAuditor(id, "CON_HALLAZGOS"), "auditor", despues.Add(time.Minute), false))
		},
	}
	for nombre, reg := range falla {
		raiz, id := sliceConVerde(t)
		if err := AnotarAuditoria(raiz, id, "APROBADO", reg(t, raiz, id), despues); err != nil {
			t.Fatalf("%s: %v", nombre, err)
		}
		if _, err := Estado(raiz, id, "CERRADO"); err == nil || !strings.Contains(err.Error(), "registro del auditor") {
			t.Errorf("%s: CERRADO se rechaza nombrando el registro del auditor; dio %v", nombre, err)
		}
	}

	// Choque con la huella (A1, hallazgo 2): registro válido, pero el APROBADO se dio sobre otro código.
	{
		raiz, id := sliceConVerde(t)
		reg := registroAuditor(t, lineaAuditor(t, cierreAuditor(id, "APROBADO"), "auditor", despues, false))
		if err := AnotarAuditoria(raiz, id, "APROBADO", reg, despues); err != nil {
			t.Fatal(err)
		}
		f := leerFicha(t, activo(raiz, nombre(id)))
		otra := map[string]string{"p/p.go": "otra"}
		f.Auditoria.Huella = &otra
		pisar(t, raiz, Activo, f)
		if _, err := Estado(raiz, id, "CERRADO"); err == nil || !strings.Contains(err.Error(), "otro código") {
			t.Errorf("registro válido sobre otro código: la huella sigue mandando; dio %v", err)
		}
	}

	// Pasa: APROBADO del auditor después del verde, como texto o como SubagentHandback, tras un CON_HALLAZGOS previo.
	for nombre, handback := range map[string]bool{"texto": false, "SubagentHandback": true} {
		raiz, id := sliceConVerde(t)
		reg := registroAuditor(t, `{roto`,
			lineaAuditor(t, cierreAuditor(id, "CON_HALLAZGOS"), "auditor", despues, handback),
			lineaAuditor(t, cierreAuditor(id, "APROBADO"), "auditor", despues.Add(time.Minute), handback))
		if err := AnotarAuditoria(raiz, id, "APROBADO", reg, despues.Add(time.Minute)); err != nil {
			t.Fatal(err)
		}
		if _, err := Estado(raiz, id, "CERRADO"); err != nil {
			t.Errorf("%s: con el APROBADO del auditor en su registro, cierra; dio %v", nombre, err)
		}
	}
}
