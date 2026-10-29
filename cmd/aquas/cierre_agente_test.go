package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// lineasRegistro: líneas de .tmp/hooks/cierre-agente.jsonl bajo raiz (E001/F004/S005).
func lineasRegistro(t *testing.T, raiz string) []map[string]string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(raiz, ".tmp", "hooks", "cierre-agente.jsonl"))
	if err != nil {
		t.Fatalf("no hay registro del gancho: %v", err)
	}
	var ls []map[string]string
	for _, l := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		var m map[string]string
		if err := json.Unmarshal([]byte(l), &m); err != nil {
			t.Fatalf("línea ilegible %q: %v", l, err)
		}
		ls = append(ls, m)
	}
	return ls
}

// TestCierreAgenteTipoDesconocido (E001/F004/S005): un tipo de agente que el gancho no conoce frena y se nombra.
// Rojo cuando: el gancho sale 0 con agent_type desconocido.
// Verde cuando: sale 2 y nombra el tipo; con stop_hook_active deja pasar.
func TestCierreAgenteTipoDesconocido(t *testing.T) {
	raiz := t.TempDir()
	for _, tipo := range []string{"otro:auditor", ""} {
		e := `{"agent_type":"` + tipo + `","last_assistant_message":"id: E001-F004-S005-A1\nveredicto: APROBADO\nhallazgos: —\nno_revisado: —","stop_hook_active":false}`
		if s := aquas(t, raiz, e, "cierre-agente"); s.codigo != 2 || !strings.Contains(s.err, "tipo de agente desconocido «"+tipo+"»") {
			t.Errorf("tipo %q: exit 2 nombrando el tipo; dio %+v", tipo, s)
		}
	}
	if s := aquas(t, raiz, `{"agent_type":"otro:auditor","last_assistant_message":"x","stop_hook_active":true}`, "cierre-agente"); s.codigo != 0 {
		t.Errorf("con stop_hook_active deja pasar; dio %+v", s)
	}
	ls := lineasRegistro(t, raiz)
	if len(ls) != 3 || ls[0]["tipo"] != "otro:auditor" || ls[0]["resultado"] != "tipo desconocido" || ls[2]["resultado"] != "tipo desconocido" {
		t.Errorf("cada llamada con tipo desconocido deja su línea; quedó %v", ls)
	}
	for _, tipo := range []string{"worker-sonnet", "worker-haiku", "auditor"} { // los conocidos siguen como antes
		if s := aquas(t, raiz, `{"agent_type":"`+tipo+`","last_assistant_message":"prosa","stop_hook_active":true}`, "cierre-agente"); s.codigo != 0 {
			t.Errorf("%s con stop_hook_active: exit 0; dio %+v", tipo, s)
		}
	}
}

// TestCierreAgenteRegistra (E001/F004/S005): cada llamada del gancho deja una línea con lo recibido y lo hecho.
// Rojo cuando: no queda registro en .tmp.
// Verde cuando: cada llamada deja tipo, transcript, primera línea y resultado; sin registro posible, el gancho responde igual.
func TestCierreAgenteRegistra(t *testing.T) {
	raiz := t.TempDir()
	fichaSlice(t, raiz, "E001/F014/S002", "EN_CURSO")
	casos := []struct{ entrada, resultado string }{
		{`no es json`, "entrada ilegible"},
		{`{"agent_type":"worker-haiku","last_assistant_message":"","stop_hook_active":false}`, "rechazado: cierre inválido: cierre vacío"},
		{`{"agent_type":"worker-haiku","last_assistant_message":"estado: HECHO\narchivos: —\nevidencia: x\nbloqueo: —\nfuera_de_alcance: —","agent_transcript_path":"t.jsonl","stop_hook_active":false}`, "aceptado"},
		{`{"agent_type":"auditor","last_assistant_message":"id: E001-F014-S002-A1\nveredicto: APROBADO\nhallazgos: —\nno_revisado: —","stop_hook_active":false}`, "anotado E001/F014/S002 APROBADO"},
		// Hallazgo A1.2: el veredicto de un tramo sin ficha en curso no se anota; una sola línea con el motivo.
		{`{"agent_type":"auditor","last_assistant_message":"id: E009-F001-S001-A1\nveredicto: APROBADO\nhallazgos: —\nno_revisado: —","stop_hook_active":false}`, "rechazado: no se anota el veredicto de E009/F001/S001"},
		// Hallazgo A2.1: con stop_hook_active deja salir, pero igual una sola línea y nunca «anotado».
		{`{"agent_type":"auditor","last_assistant_message":"id: E009-F001-S001-A1\nveredicto: APROBADO\nhallazgos: —\nno_revisado: —","stop_hook_active":true}`, "rechazado: no se anota el veredicto de E009/F001/S001"},
		// Hallazgo A1.1: auditor sin veredicto anotable (id: — o cierre malo con stop_hook_active) no figura como aceptado.
		{`{"agent_type":"auditor","last_assistant_message":"id: —\nveredicto: APROBADO\nhallazgos: —\nno_revisado: —","stop_hook_active":false}`, "sin anotar"},
		{`{"agent_type":"auditor","last_assistant_message":"Listo, aprobado.","stop_hook_active":true}`, "sin anotar"},
	}
	codigos := []int{0, 2, 0, 0, 2, 0, 0, 0}
	for i, c := range casos {
		if s := aquas(t, raiz, c.entrada, "cierre-agente"); s.codigo != codigos[i] {
			t.Errorf("llamada %d: exit %d; dio %+v", i, codigos[i], s)
		}
	}
	ls := lineasRegistro(t, raiz)
	if len(ls) != len(casos) {
		t.Fatalf("una línea por llamada: %d llamadas, %d líneas: %v", len(casos), len(ls), ls)
	}
	for i, c := range casos {
		if !strings.HasPrefix(ls[i]["resultado"], c.resultado) || ls[i]["fecha"] == "" {
			t.Errorf("llamada %d: resultado %q con fecha; quedó %v", i, c.resultado, ls[i])
		}
	}
	if ls[2]["tipo"] != "worker-haiku" || ls[2]["transcript"] != "t.jsonl" || ls[2]["primera"] != "estado: HECHO" {
		t.Errorf("la línea trae tipo, transcript y primera línea; quedó %v", ls[2])
	}

	// Fallo a mitad de camino: .tmp/hooks no se puede crear; el gancho responde igual.
	otra := t.TempDir()
	if err := os.MkdirAll(filepath.Join(otra, ".tmp"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(otra, ".tmp", "hooks"), []byte("no soy carpeta"), 0o644); err != nil {
		t.Fatal(err)
	}
	if s := aquas(t, otra, `{"agent_type":"worker-haiku","last_assistant_message":"prosa","stop_hook_active":false}`, "cierre-agente"); s.codigo != 2 || !strings.Contains(s.err, "cierre inválido") {
		t.Errorf("sin registro posible, el gancho responde igual (exit 2 por prosa); dio %+v", s)
	}
}

// E001/F006/S002: `aquas cierre-agente` lee los campos que Claude Code manda a SubagentStop
// (agent_type, last_assistant_message, stop_hook_active; doc oficial de hooks, 2026-10-21).
// JSON ilegible → deja cerrar: un fallo de entrada no traba a los agentes.
func TestCierreAgenteCLI(t *testing.T) {
	raiz := t.TempDir()
	malo := `{"agent_type":"worker-haiku","last_assistant_message":"Listo, terminé.","stop_hook_active":false}`
	if s := aquas(t, raiz, malo, "cierre-agente"); s.codigo != 2 || !strings.Contains(s.err, "cierre inválido") {
		t.Fatalf("cierre en prosa: exit 2 con aviso por stderr; dio %+v", s)
	}
	fichaSlice(t, raiz, "E001/F014/S002", "EN_CURSO") // E001/F014/S002: el cierre del auditor se anota en la ficha de su tramo
	bueno := `{"agent_type":"auditor","last_assistant_message":"id: E001-F014-S002-A1\nveredicto: APROBADO\nhallazgos: —\nno_revisado: —","stop_hook_active":false}`
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
