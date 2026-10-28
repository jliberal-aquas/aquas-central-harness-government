package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// cuentaFeature: archivo y llamadas de la cuenta de feature más reciente.
func cuentaFeature(t *testing.T, raiz string) (string, int) {
	t.Helper()
	ms, _ := filepath.Glob(filepath.Join(raiz, ".tmp", "presupuesto-cuentas", "feature-*.log"))
	if len(ms) == 0 {
		t.Fatal("no hay cuenta de feature")
	}
	ultima := ms[0]
	for _, m := range ms {
		a, _ := os.Stat(m)
		b, _ := os.Stat(ultima)
		if a.ModTime().After(b.ModTime()) {
			ultima = m
		}
	}
	b, _ := os.ReadFile(ultima)
	return ultima, strings.Count(string(b), "\n")
}

// E001/F015/S002: aquas presupuesto, el freno de llamadas por feature y por turno, sobre el binario real.

// usoHerramienta: entrada del hook PreToolUse; agente != "" es una llamada de un subagente.
func usoHerramienta(agente, transcript string) string {
	e := map[string]string{"session_id": "s1", "hook_event_name": "PreToolUse", "tool_name": "Bash", "transcript_path": transcript}
	if agente != "" {
		e["agent_id"] = agente
	}
	b, _ := json.Marshal(e)
	return string(b)
}

// TestPresupuestoPorFeature (E001/F015/S002).
// Rojo cuando: se pasa del tope de la feature o del freno por turno sin bloquear.
// Verde cuando: bloquea al pasar cualquiera de los dos topes y deja pasar dentro de ellos.
func TestPresupuestoPorFeature(t *testing.T) {
	cfg, err := os.ReadFile(filepath.Join("..", "..", "gobierno", "config", "claude_settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`{"evento": "UserPromptSubmit", "orden": "presupuesto prompt"}`, `{"evento": "PreToolUse", "orden": "presupuesto herramienta"}`} {
		if !strings.Contains(string(cfg), want) {
			t.Errorf("claude_settings.json no conecta el hook %s", want)
		}
	}
	raiz := repoFeature(t)
	if s := aquas(t, raiz, "", "memoria", "fijar", "iniciativa=E001", "feature=F002"); s.codigo != 0 {
		t.Fatalf("memoria fijar: %+v", s)
	}
	turno := func(prompt string) {
		t.Helper()
		if s := aquas(t, raiz, mensajeHumano(prompt), "aprobacion"); s.codigo != 0 {
			t.Fatalf("aprobacion: %+v", s)
		}
		if s := aquas(t, raiz, mensajeHumano(prompt), "presupuesto", "prompt"); s.codigo != 0 {
			t.Fatalf("presupuesto prompt: %+v", s)
		}
	}
	usar := func(agente string) salida {
		t.Helper()
		return aquas(t, raiz, usoHerramienta(agente, ""), "presupuesto", "herramienta")
	}
	turno("sigue")
	for i := range 5 {
		if s := usar(""); s.codigo != 0 {
			t.Fatalf("sin aprobación de la feature solo rige el turno; llamada %d dio %+v", i+1, s)
		}
	}
	turno("apruebo E001/F002 tope 5")
	for i, agente := range []string{"", "", "a1", "a1", "a1"} {
		if s := usar(agente); s.codigo != 0 {
			t.Fatalf("dentro del tope de la feature; llamada %d dio %+v", i+1, s)
		}
	}
	if s := usar("a1"); s.codigo != 2 || !strings.Contains(s.err, "tope de la feature") {
		t.Errorf("la llamada 6 de un subagente pasa el tope 5: bloqueo; dio %+v", s)
	}
	time.Sleep(1100 * time.Millisecond) // la fecha de aprobación va en segundos
	turno("apruebo E001/F002 tope 5")
	if s := usar(""); s.codigo != 0 {
		t.Errorf("la aprobación nueva cuenta desde cero; dio %+v", s)
	}
	time.Sleep(1100 * time.Millisecond)
	turno("apruebo E001/F002\ntope 2")
	for i := range 3 {
		if s := usar(""); s.codigo != 0 {
			t.Fatalf("«tope» fuera de la primera línea no cuenta: rige 300; llamada %d dio %+v", i+1, s)
		}
	}
	ruta, _ := cuentaFeature(t, raiz)
	if err := os.WriteFile(ruta, []byte(strings.Repeat(".\n", 299)), 0o644); err != nil {
		t.Fatal(err)
	}
	if s := usar("a1"); s.codigo != 0 {
		t.Errorf("sin «tope N» la llamada 300 pasa; dio %+v", s)
	}
	if s := usar("a1"); s.codigo != 2 || !strings.Contains(s.err, "300") {
		t.Errorf("sin «tope N» la llamada 301 se bloquea; dio %+v", s)
	}
	time.Sleep(1100 * time.Millisecond)
	turno("apruebo E001/F002 tope 1000")
	for i := range 60 {
		if s := usar(""); s.codigo != 0 {
			t.Fatalf("la aprobación nueva cuenta desde cero; llamada %d dio %+v", i+1, s)
		}
	}
	if s := usar("a2"); s.codigo != 0 {
		t.Errorf("un subagente no cuenta en el freno del turno; dio %+v", s)
	}
	_, antes := cuentaFeature(t, raiz)
	if s := usar(""); s.codigo != 2 || !strings.Contains(s.err, "turno") {
		t.Errorf("la llamada 61 del turno: bloqueo; dio %+v", s)
	}
	if _, despues := cuentaFeature(t, raiz); despues != antes {
		t.Errorf("la llamada que frena el turno no gasta tope de la feature: %d → %d", antes, despues)
	}
	aquas(t, raiz, mensajeHumano("<task-notification>agente terminó</task-notification>"), "presupuesto", "prompt")
	if s := usar(""); s.codigo != 2 {
		t.Errorf("un aviso de agente no reinicia el turno; dio %+v", s)
	}
	if s := aquas(t, raiz, "no json", "presupuesto", "herramienta"); s.codigo != 1 {
		t.Errorf("entrada ilegible: código 1, sin bloquear; dio %+v", s)
	}
	turno("sigue")
	if s := usar(""); s.codigo != 0 {
		t.Errorf("un mensaje nuevo del humano reinicia el turno; dio %+v", s)
	}
}

// TestPresupuestoAvisaContexto (E001/F015/S002).
// Rojo cuando: el contexto pasa el umbral sin aviso o avisa por debajo.
// Verde cuando: avisa solo desde el umbral y pide compact.
func TestPresupuestoAvisaContexto(t *testing.T) {
	raiz := repoFeature(t)
	tr := filepath.Join(raiz, "transcript.jsonl")
	medida := func(tokens int) string {
		return `{"type":"assistant","message":{"usage":{"input_tokens":10,"cache_read_input_tokens":` + itoa(tokens-10) + `}}}` + "\n"
	}
	usar := func() salida {
		t.Helper()
		s := aquas(t, raiz, usoHerramienta("", tr), "presupuesto", "herramienta")
		if s.codigo != 0 {
			t.Fatalf("el aviso de contexto no bloquea; dio %+v", s)
		}
		return s
	}
	aquas(t, raiz, mensajeHumano("sigue"), "presupuesto", "prompt")
	escribir(t, raiz, "transcript.jsonl", medida(199_999))
	if s := usar(); strings.Contains(s.out, "compact") {
		t.Errorf("por debajo de 200000 no avisa; dio %+v", s)
	}
	escribir(t, raiz, "transcript.jsonl", medida(199_999)+medida(200_000))
	if s := usar(); !strings.Contains(s.out, "/compact") || !strings.Contains(s.out, "200000") {
		t.Errorf("desde 200000 avisa con la cifra y pide /compact; dio %+v", s)
	}
	if s := usar(); strings.Contains(s.out, "compact") {
		t.Errorf("avisa una vez por turno; dio %+v", s)
	}
	aquas(t, raiz, mensajeHumano("sigue"), "presupuesto", "prompt")
	if s := usar(); !strings.Contains(s.out, "/compact") {
		t.Errorf("cada turno nuevo vuelve a avisar; dio %+v", s)
	}
	aquas(t, raiz, mensajeHumano("sigue"), "presupuesto", "prompt")
	escribir(t, raiz, "transcript.jsonl", medida(250_000)+`{"type":"system","subtype":"compact_boundary"}`+"\n")
	if s := usar(); strings.Contains(s.out, "compact") {
		t.Errorf("tras compactar la medida vieja no cuenta; dio %+v", s)
	}
}

func itoa(n int) string {
	b, _ := json.Marshal(n)
	return string(b)
}
