package main

import (
	"strings"
	"testing"
)

// E001/F006/S002: `aquas cierre-agente` lee los campos que Claude Code manda a SubagentStop
// (agent_type, last_assistant_message, stop_hook_active; doc oficial de hooks, 2026-10-21).
// JSON ilegible → deja cerrar: un fallo de entrada no traba a los agentes.
func TestCierreAgenteCLI(t *testing.T) {
	raiz := t.TempDir()
	malo := `{"agent_type":"worker-haiku","last_assistant_message":"Listo, terminé.","stop_hook_active":false}`
	if s := aquas(t, raiz, malo, "cierre-agente"); s.codigo != 2 || !strings.Contains(s.err, "cierre inválido") {
		t.Fatalf("cierre en prosa: exit 2 con aviso por stderr; dio %+v", s)
	}
	bueno := `{"agent_type":"auditor","last_assistant_message":"veredicto: APROBADO\nhallazgos: —\nno_revisado: —","stop_hook_active":false}`
	if s := aquas(t, raiz, bueno, "cierre-agente"); s.codigo != 0 {
		t.Fatalf("cierre completo: exit 0; dio %+v", s)
	}
	if s := aquas(t, raiz, `{"agent_type":"worker-haiku","last_assistant_message":"prosa","stop_hook_active":true}`, "cierre-agente"); s.codigo != 0 {
		t.Fatalf("ya pedido una vez: exit 0; dio %+v", s)
	}
	if s := aquas(t, raiz, `no es json`, "cierre-agente"); s.codigo != 0 {
		t.Fatalf("JSON ilegible: deja cerrar; dio %+v", s)
	}
}
