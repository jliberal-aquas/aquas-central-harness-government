package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// E001/F015/S002: aquas presupuesto, el freno de llamadas por feature y por turno, sobre el binario real.
// E000/F008/S005: las llamadas se cuentan en el registro de la sesión; usoHerramienta las escribe ahí.

// TestPresupuestoPorFeature (E001/F015/S002).
// Rojo cuando: se pasa del tope de la feature o del freno por turno sin bloquear.
// Verde cuando: bloquea al pasar cualquiera de los dos topes y deja pasar dentro de ellos.
func TestPresupuestoPorFeature(t *testing.T) {
	cfg, err := os.ReadFile(filepath.Join("..", "..", "gobierno", "config", "claude_settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(cfg), `{"evento": "PreToolUse", "orden": "budget tool"}`) {
		t.Errorf("claude_settings.json no conecta el hook budget tool")
	}
	if strings.Contains(string(cfg), "budget prompt") {
		t.Errorf("claude_settings.json conecta budget prompt, que se retiró (E000/F008/S005)")
	}
	raiz := repoFeature(t)
	if s := aquas(t, raiz, "", "memoria", "fijar", "iniciativa=E001", "feature=F002"); s.codigo != 0 {
		t.Fatalf("memoria fijar: %+v", s)
	}
	turno := func(prompt string) {
		t.Helper()
		if s := aquas(t, raiz, mensajeHumano(t, prompt), "aprobacion"); s.codigo != 0 {
			t.Fatalf("aprobacion: %+v", s)
		}
	}
	usar := func(agente string) salida {
		t.Helper()
		return aquas(t, raiz, usoHerramienta(t, agente, registroSesion(t)), "presupuesto", "herramienta")
	}
	turno("sigue")
	for i := range 5 {
		if s := usar(""); s.codigo != 0 {
			t.Fatalf("sin aprobación de la feature solo rige el turno; llamada %d dio %+v", i+1, s)
		}
	}
	time.Sleep(1100 * time.Millisecond) // la fecha de aprobación va en segundos: lo anterior no debe caer en ella
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
			t.Fatalf("«tope» fuera de la primera línea no cuenta: rige el por defecto; llamada %d dio %+v", i+1, s)
		}
	}
	sub := filepath.Join(filepath.Dir(registroSesion(t)), "sesion", "subagents", "agent-a1.jsonl")
	for i := range 296 { // 3 llamadas ya hechas desde la aprobación + 296 = 299
		anexar(t, sub, lineaLlamada(fmt.Sprintf("toolu_relleno_%d", i), "Bash", "ls", time.Now()))
	}
	if s := usar("a1"); s.codigo != 0 {
		t.Errorf("sin «tope N» la llamada 300 pasa; dio %+v", s)
	}
	if s := usar("a1"); s.codigo != 0 { // E001/F019/S001: sin «tope» rige 15M tokens ponderados, no 300 llamadas
		t.Errorf("sin «tope N» la llamada 301 pasa: el tope por defecto es en tokens; dio %+v", s)
	}
	// El freno del turno (sin apruebo, y su cesión con apruebo vigente) lo prueba TestFrenoTurnoCedeConFeatureAprobada (E001/F017/S009).
	if s := aquas(t, raiz, "no json", "presupuesto", "herramienta"); s.codigo != 1 {
		t.Errorf("entrada ilegible: código 1, sin bloquear; dio %+v", s)
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
		s := aquas(t, raiz, usoHerramienta(t, "", tr), "presupuesto", "herramienta")
		if s.codigo != 0 {
			t.Fatalf("el aviso de contexto no bloquea; dio %+v", s)
		}
		return s
	}
	anexar(t, tr, lineaHumana("sigue"))
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
	anexar(t, tr, lineaHumana("sigue"))
	if s := usar(); !strings.Contains(s.out, "/compact") {
		t.Errorf("cada turno nuevo vuelve a avisar; dio %+v", s)
	}
	anexar(t, tr, lineaHumana("sigue"))
	escribir(t, raiz, "transcript.jsonl", medida(250_000)+`{"type":"system","subtype":"compact_boundary"}`+"\n")
	if s := usar(); strings.Contains(s.out, "compact") {
		t.Errorf("tras compactar la medida vieja no cuenta; dio %+v", s)
	}
}

func itoa(n int) string {
	b, _ := json.Marshal(n)
	return string(b)
}
