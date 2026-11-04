package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// registroSesion: el registro de la sesión de la prueba (CLAUDE_CONFIG_DIR/projects/repo/sesion.jsonl); lo crea
// y fija CLAUDE_CONFIG_DIR y AQUAS_PRUEBA_REGISTRO la primera vez.
func registroSesion(t *testing.T) string {
	t.Helper()
	ruta := os.Getenv("AQUAS_PRUEBA_REGISTRO")
	if ruta == "" {
		dir := filepath.Join(t.TempDir(), "claude")
		ruta = filepath.Join(dir, "projects", "repo", "sesion.jsonl")
		if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
			t.Fatal(err)
		}
		t.Setenv("CLAUDE_CONFIG_DIR", dir)
		t.Setenv("AQUAS_PRUEBA_REGISTRO", ruta)
	}
	return ruta
}

// anexar suma líneas al final de un registro, como lo hace Claude Code.
func anexar(t *testing.T, ruta string, lineas ...string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
		t.Fatal(err)
	}
	fh, err := os.OpenFile(ruta, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = fh.Close() }()
	if _, err := fh.WriteString(strings.Join(lineas, "\n") + "\n"); err != nil {
		t.Fatal(err)
	}
}

var idsLlamada atomic.Int64

// lineaLlamada: línea de asistente con un bloque tool_use (forma observada el 2026-10-31).
func lineaLlamada(id, herramienta, comando string, cuando time.Time) string {
	b, _ := json.Marshal(map[string]any{"type": "assistant", "timestamp": cuando.UTC().Format("2006-01-02T15:04:05.000Z"),
		"message": map[string]any{"role": "assistant", "content": []any{map[string]any{"type": "tool_use", "id": id,
			"name": herramienta, "input": map[string]string{"command": comando}}}}})
	return string(b)
}

// lineaHumana: mensaje de usuario con origen humano (forma observada el 2026-10-31).
func lineaHumana(texto string) string {
	b, _ := json.Marshal(map[string]any{"type": "user", "message": map[string]string{"role": "user", "content": texto},
		"origin": map[string]string{"kind": "human"}, "timestamp": time.Now().UTC().Format("2006-01-02T15:04:05.000Z")})
	return string(b)
}

// llamar: como Claude Code, escribe la llamada en el registro (el del subagente si agente != "") y devuelve la
// entrada del hook PreToolUse. Con transcript "" no escribe nada.
func llamar(t *testing.T, agente, transcript, herramienta, comando string) string {
	t.Helper()
	if transcript != "" {
		destino := transcript
		if agente != "" {
			sesion := strings.TrimSuffix(filepath.Base(transcript), ".jsonl")
			destino = filepath.Join(filepath.Dir(transcript), sesion, "subagents", "agent-"+agente+".jsonl")
		}
		anexar(t, destino, lineaLlamada(fmt.Sprintf("toolu_%d", idsLlamada.Add(1)), herramienta, comando, time.Now()))
	}
	e := map[string]any{"session_id": "s1", "hook_event_name": "PreToolUse", "tool_name": herramienta,
		"transcript_path": transcript, "tool_input": map[string]string{"command": comando}}
	if agente != "" {
		e["agent_id"] = agente
	}
	b, _ := json.Marshal(e)
	return string(b)
}

// usoHerramienta: llamada Bash al registro dado; agente != "" es un subagente.
func usoHerramienta(t *testing.T, agente, transcript string) string {
	t.Helper()
	return llamar(t, agente, transcript, "Bash", "git status")
}

// usoComando: llamada del hilo principal con su comando, al registro de la sesión de la prueba.
func usoComando(t *testing.T, herramienta, comando string) string {
	t.Helper()
	return llamar(t, "", registroSesion(t), herramienta, comando)
}

// TestTopeDesdeElRegistro (E000/F008/S005, D000077): el tope de la feature y el turno se cuentan en el
// registro de Claude Code; borrar .tmp/presupuesto-cuentas no cambia nada.
// Rojo cuando: borrar .tmp/presupuesto-cuentas reinicia la cuenta.
// Verde cuando: el tope y el turno salen del registro y borrar .tmp no cambia nada.
func TestTopeDesdeElRegistro(t *testing.T) {
	raiz := repoFeature(t)
	if s := aquas(t, raiz, "", "memory", "set", "iniciativa=E001", "feature=F002"); s.codigo != 0 {
		t.Fatalf("memory set: %+v", s)
	}
	reg := registroSesion(t)
	humano := func(prompt string) {
		t.Helper()
		if s := aquas(t, raiz, mensajeHumano(t, prompt), "approval"); s.codigo != 0 {
			t.Fatalf("approval: %+v", s)
		}
	}
	borrar := func() {
		t.Helper()
		if err := os.RemoveAll(filepath.Join(raiz, ".tmp", "presupuesto-cuentas")); err != nil {
			t.Fatal(err)
		}
	}
	usar := func(agente string) salida {
		t.Helper()
		return aquas(t, raiz, usoHerramienta(t, agente, reg), "budget", "tool")
	}

	// Turno: las llamadas desde el último mensaje humano salen del registro; borrar .tmp no lo reinicia.
	// E001/F019/S002: el freno de 60 salió; la cuenta del turno queda para el tope extremo (S004).
	turnoDe := func() int {
		t.Helper()
		e, err := escanear(raiz, reg)
		if err != nil {
			t.Fatal(err)
		}
		return len(e.turno)
	}
	humano("sigue")
	for i := range 61 {
		if s := usar(""); s.codigo != 0 {
			t.Fatalf("sin freno de 60; llamada %d dio %+v", i+1, s)
		}
		if i == 30 {
			borrar()
		}
	}
	borrar()
	if n := turnoDe(); n != 61 {
		t.Errorf("borrar .tmp no reinicia el turno: 61 llamadas; contó %d", n)
	}
	if s := usar("a1"); s.codigo != 0 || turnoDe() != 61 {
		t.Errorf("un subagente no cuenta en el turno del hilo; dio %+v", s)
	}

	// Quién habla: un aviso de agente o un resultado de herramienta no abren turno; el humano sí.
	resultado, _ := json.Marshal(map[string]any{"type": "user", "origin": map[string]string{"kind": "human"},
		"message": map[string]any{"role": "user", "content": []map[string]string{{"type": "tool_result", "content": "sigue"}}}})
	anexar(t, reg, lineaHumana("<task-notification>agente terminó</task-notification>"), string(resultado))
	if n := turnoDe(); n != 61 {
		t.Errorf("un aviso de agente o un resultado de herramienta no reinician el turno; contó %d", n)
	}
	humano("sigue")

	// Entrada duplicada o rota: la misma llamada repetida cuenta una vez; una línea rota se ignora.
	repetida := lineaLlamada("toolu_repetida", "Bash", "git status", time.Now())
	for range 70 {
		anexar(t, reg, repetida)
	}
	anexar(t, reg, `{"type":"assistant","message":{"content":[{"type":"tool_use","id":"rota`)
	if n := turnoDe(); n != 1 {
		t.Errorf("una llamada repetida 70 veces cuenta una vez y la rota se ignora; contó %d", n)
	}
	if s := usar(""); s.codigo != 0 {
		t.Errorf("una llamada repetida 70 veces cuenta una vez; dio %+v", s)
	}

	// Tope de la feature: cuenta desde la aprobación, con subagentes; lo de antes no cuenta.
	for range 10 {
		anexar(t, reg, lineaLlamada(fmt.Sprintf("toolu_vieja_%d", idsLlamada.Add(1)), "Bash", "ls", time.Now().Add(-time.Hour)))
	}
	time.Sleep(1100 * time.Millisecond) // la fecha de aprobación va en segundos
	humano("apruebo E001/F002 tope 5")
	for i, agente := range []string{"", "a1", "a1", "", ""} {
		if s := usar(agente); s.codigo != 0 {
			t.Fatalf("dentro del tope de la feature; llamada %d dio %+v", i+1, s)
		}
	}
	borrar()
	for _, agente := range []string{"a1", ""} {
		if s := usar(agente); s.codigo != 2 || !strings.Contains(s.err, "tope de la feature") {
			t.Errorf("borrar .tmp no reinicia la feature: la llamada 6 (%q) se bloquea; dio %+v", agente, s)
		}
	}
	if s := aquas(t, raiz, usoComando(t, "Bash", "aquas memory set tarea=x"), "budget", "tool"); s.codigo != 0 {
		t.Errorf("memory set sigue exento con la feature desbordada; dio %+v", s)
	}

	// D000082: cambiar el foco, editar la aprobación de la ficha o aprobar un slice no cambia el tope.
	if s := aquas(t, raiz, "", "memory", "set", "feature=F099", "slice="); s.codigo != 0 {
		t.Fatalf("memory set: %+v", s)
	}
	ficha := filepath.Join(raiz, "trabajo", "activo", "E001-F002.json")
	b, err := os.ReadFile(ficha)
	if err != nil {
		t.Fatal(err)
	}
	var f map[string]any
	if err := json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	f["aprobacion"] = map[string]any{"fecha": time.Now().UTC().Format(time.RFC3339), "mensaje": "apruebo E001/F002 tope 5000", "registro": reg}
	b, _ = json.Marshal(f)
	if err := os.WriteFile(ficha, b, 0o644); err != nil {
		t.Fatal(err)
	}
	humano("apruebo E001/F002/S003")
	if s := usar(""); s.codigo != 2 || !strings.Contains(s.err, "tope de la feature") {
		t.Errorf("foco cambiado, ficha editada y apruebo de slice: el tope sigue; dio %+v", s)
	}

	// Fallo a mitad de camino: un apruebo nuevo, en otra sesión del proyecto, cuenta desde su hora;
	// memory set no suma.
	time.Sleep(10 * time.Millisecond)
	anexar(t, filepath.Join(filepath.Dir(reg), "otra.jsonl"), lineaHumana("apruebo E001/F002 tope 3"))
	for range 5 {
		if s := aquas(t, raiz, usoComando(t, "Bash", "aquas memory set tarea=x"), "budget", "tool"); s.codigo != 0 {
			t.Fatalf("memory set exento: %+v", s)
		}
	}
	for i := range 3 {
		if s := usar(""); s.codigo != 0 {
			t.Errorf("apruebo nuevo en otra sesión: la llamada %d pasa (memory set no sumó); dio %+v", i+1, s)
		}
	}
	if s := usar(""); s.codigo != 2 || !strings.Contains(s.err, "tope de la feature") {
		t.Errorf("la llamada 4 pasa el tope 3; dio %+v", s)
	}

	// Registro ausente o ilegible: bloquea y lo nombra; no cuenta como cero.
	for _, tr := range []string{"", filepath.Join(filepath.Dir(reg), "no-existe.jsonl")} {
		e, _ := json.Marshal(map[string]any{"session_id": "s1", "tool_name": "Bash", "transcript_path": tr,
			"tool_input": map[string]string{"command": "git status"}})
		if s := aquas(t, raiz, string(e), "budget", "tool"); s.codigo != 2 || !strings.Contains(s.err, "registro") {
			t.Errorf("registro %q: exit 2 nombrando el registro; dio %+v", tr, s)
		}
	}
	if s := aquas(t, raiz, `{"tool_name":"Bash"}`, "budget", "prompt"); s.codigo == 0 {
		t.Errorf("budget prompt se retiró: ya no es una orden válida; dio %+v", s)
	}
}
