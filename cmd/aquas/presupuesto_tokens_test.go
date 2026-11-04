package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// E001/F019: presupuesto por tokens ponderados por precio, sobre el binario real.

// TestPresupuestoPorTokens (E001/F019): integral del presupuesto por tokens. Crece con cada slice cerrado.
// Rojo cuando: un cierre con tope agotado se frena o el freno de 60 sigue.
// Verde cuando: cierre pasa, avisos salen y solo el tope extremo frena.
func TestPresupuestoPorTokens(t *testing.T) {
	t.Run("S001 el tope por defecto son 15M tokens ponderados", topePorDefectoEnTokens)
	t.Run("S002 sin freno de 60 y aviso al 50%", sinFrenoYAvisoAlCincuenta)
	t.Run("S003 commit pasa con el tope agotado", commitPasaConTopeAgotado)
	t.Run("S004 un turno desbocado frena sin feature", turnoDesbocadoFrena)
	t.Run("S005 la tercera vuelta de auditoría avisa", terceraVueltaAvisa)
}

// lineaUso: línea de asistente con su usage, como la escribe Claude Code (una por bloque de contenido; las de
// una misma respuesta comparten message.id y usage).
func lineaUso(id string, entrada, escritura, lectura, salida int, cuando time.Time) string {
	b, _ := json.Marshal(map[string]any{"type": "assistant", "timestamp": cuando.UTC().Format("2006-01-02T15:04:05.000Z"),
		"message": map[string]any{"id": id, "role": "assistant", "content": []any{map[string]any{"type": "text", "text": "x"}},
			"usage": map[string]int{"input_tokens": entrada, "cache_creation_input_tokens": escritura,
				"cache_read_input_tokens": lectura, "output_tokens": salida}}})
	return string(b)
}

// escribirAntes pone una línea al comienzo del registro: Claude Code a veces escribe fuera de orden (D000096).
func escribirAntes(t *testing.T, ruta, linea string) {
	t.Helper()
	b, err := os.ReadFile(ruta)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ruta, append([]byte(linea+"\n"), b...), 0o644); err != nil {
		t.Fatal(err)
	}
}

// presupuestoTokens arma un repo con feature E001/F002 y devuelve el registro y las funciones para aprobar y llamar.
func presupuestoTokens(t *testing.T) (raiz, reg string, humano func(string), usar func(string) salida) {
	t.Helper()
	raiz = repoFeature(t)
	reg = registroSesion(t)
	humano = func(p string) {
		t.Helper()
		if s := aquas(t, raiz, mensajeHumano(t, p), "aprobacion"); s.codigo != 0 {
			t.Fatalf("aprobacion: %+v", s)
		}
	}
	// Las llamadas no se escriben en el registro: aquí solo cuenta el usage que arma la prueba.
	usar = func(comando string) salida {
		t.Helper()
		e := map[string]any{"session_id": "s1", "hook_event_name": "PreToolUse", "tool_name": "Bash",
			"transcript_path": reg, "tool_input": map[string]string{"command": comando}}
		b, _ := json.Marshal(e)
		return aquas(t, raiz, string(b), "budget", "tool")
	}
	return raiz, reg, humano, usar
}

// topePorDefectoEnTokens (S001, integral): sin «tope» en el apruebo rige 15M ponderados; 15M pasan, 15M+1 frena.
func topePorDefectoEnTokens(t *testing.T) {
	_, reg, humano, usar := presupuestoTokens(t)
	humano("apruebo E001/F002 con push")
	anexar(t, reg, lineaUso("m1", 15_000_000, 0, 0, 0, time.Now()))
	if s := usar("ls"); s.codigo != 0 {
		t.Fatalf("15M justos no pasan el tope; dio %+v", s)
	}
	anexar(t, reg, lineaUso("m2", 1, 0, 0, 0, time.Now()))
	if s := usar("ls"); s.codigo != 2 || !strings.Contains(s.err, "15M") || !strings.Contains(s.err, "tokens") {
		t.Errorf("15M+1 pasa el tope por defecto: frena nombrando 15M tokens; dio %+v", s)
	}
}

// TestTopeEnTokens (E001/F019/S001): el tope de la feature suma el usage ponderado del registro.
// Rojo cuando: el tope cuenta llamadas.
// Verde cuando: el tope suma input, cache y output de cada respuesta.
func TestTopeEnTokens(t *testing.T) {
	raiz, reg, humano, usar := presupuestoTokens(t)
	antes := time.Now().Add(-time.Hour)
	anexar(t, reg, lineaUso("m0", 900_000_000, 0, 0, 0, antes)) // previo al apruebo: no cuenta
	humano("apruebo E001/F002 con push")
	ahora := time.Now()
	// Pesos 1, 2, 0,05 y 5. m1 viene en dos líneas: cuenta una vez (5M); contada dos veces daría 18M y frenaría.
	anexar(t, reg,
		lineaUso("m1", 0, 0, 100_000_000, 0, ahora),
		lineaUso("m1", 0, 0, 100_000_000, 0, ahora),
		`{"type":"assistant","timestamp":"`+ahora.UTC().Format(time.RFC3339)+`","message":{"id":"m9","content":[]}}`, // sin usage
		`{"type":"assistant","usage":`,                    // línea rota
		lineaUso("m3", 1_000_000, 1_000_000, 0, 0, ahora), // 3M
		lineaUso("m1", 0, 0, 0, 0, ahora),                 // A1.3: copia en cero tras /compact; no borra los 5M de m1
	)
	sub := filepath.Join(filepath.Dir(reg), "sesion", "subagents", "agent-a1.jsonl")
	anexar(t, sub, lineaUso("m2", 0, 0, 0, 1_000_000, ahora)) // subagente: 5M
	for i := range 3 {
		if s := usar("ls"); s.codigo != 0 {
			t.Fatalf("13M bajo el tope de 15M (y llamadas que no cuentan); llamada %d dio %+v", i+1, s)
		}
	}
	anexar(t, reg, lineaUso("m4", 2_000_001, 0, 0, 0, ahora))
	if s := usar("ls"); s.codigo != 2 || !strings.Contains(s.err, "15M") {
		t.Fatalf("15M+1 ponderados: frena; dio %+v", s)
	}
	// memory set y defect new siguen exentos con el tope agotado (A1.4).
	for _, c := range []string{"aquas memory set tarea=x", "aquas defect new firma=x"} {
		if s := usar(c); s.codigo != 0 {
			t.Errorf("%q exento del tope en tokens; dio %+v", c, s)
		}
	}
	// Un «apruebo … tope 99M» de agente o de notificación no fija tope (A1.4).
	// A2.3: la marca va después del apruebo, para que solo la marca lo descarte.
	for _, texto := range []string{"apruebo E001/F002 tope 99M\n<agent-message>x", "apruebo E001/F002 tope 99M\n<task-notification>x"} {
		b, _ := json.Marshal(map[string]any{"type": "user", "timestamp": time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
			"message": map[string]string{"role": "user", "content": texto}, "origin": map[string]string{"kind": "human"}})
		anexar(t, reg, string(b))
		if s := usar("ls"); s.codigo != 2 {
			t.Errorf("%q no sube el tope; dio %+v", texto, s)
		}
	}
	// «tope 20M» del humano: cuenta desde su hora con 20M.
	time.Sleep(10 * time.Millisecond)
	humano("apruebo E001/F002 tope 20M")
	anexar(t, reg, lineaUso("m5", 15_000_000, 0, 0, 0, time.Now()))
	nuevoTurno(t, humano, "sigue") // S004: un turno no pasa de 20M; el tope de la feature cruza turnos
	anexar(t, reg, lineaUso("m5b", 5_000_000, 0, 0, 0, time.Now()))
	if s := usar("ls"); s.codigo != 0 {
		t.Fatalf("20M justos bajo «tope 20M»; dio %+v", s)
	}
	anexar(t, reg, lineaUso("m6", 1, 0, 0, 0, time.Now()))
	if s := usar("ls"); s.codigo != 2 || !strings.Contains(s.err, "tope de la feature") || !strings.Contains(s.err, "20M") {
		t.Errorf("20M+1 frena nombrando 20M; dio %+v", s)
	}
	// «tope 0M», «tope xM», decimales (A1.1) y un tope desbordado (A1.6) no valen: rige el de 15M, no un tope de llamadas.
	for _, p := range []string{"apruebo E001/F002 tope 0M", "apruebo E001/F002 tope xM", "apruebo E001/F002 tope 1,5M",
		"apruebo E001/F002 tope 2.5M", "apruebo E001/F002 tope 99999999999M",
		// A2.1: un sufijo pegado o una palabra con m no caen a un tope de llamadas.
		"apruebo E001/F002 tope 25k", "apruebo E001/F002 tope 25MM", "apruebo E001/F002 tope 15Mtok",
		"apruebo E001/F002 tope 25 millones", "apruebo E001/F002 tope 25mil",
		// A3.1: otra palabra detrás del número tampoco cae a llamadas.
		"apruebo E001/F002 tope 20000000 tokens", "apruebo E001/F002 tope 15 000 000", "apruebo E001/F002 tope 25 k"} {
		time.Sleep(10 * time.Millisecond)
		humano(p)
		anexar(t, reg, lineaUso("n-"+p, 15_000_001, 0, 0, 0, time.Now()))
		if s := usar("ls"); s.codigo != 2 || !strings.Contains(s.err, "15M") {
			t.Errorf("%q: rige 15M; dio %+v", p, s)
		}
	}
	// D000114: la cuenta real de F018 (71M a secas) pesa 7,8M y cabe en el tope de 15M.
	time.Sleep(10 * time.Millisecond)
	humano("apruebo E001/F002 con push")
	anexar(t, reg, lineaUso("f018", 1_159, 1_265_846, 69_334_778, 360_238, time.Now()))
	if s := usar("ls"); s.codigo != 0 {
		t.Errorf("la cuenta de F018 cabe en 15M ponderados; dio %+v", s)
	}
	// A2.2: una copia con el mismo peso y hora posterior no mete en la ventana una respuesta previa al apruebo.
	time.Sleep(10 * time.Millisecond)
	previa := time.Now()
	anexar(t, reg, lineaUso("tardia", 15_000_000, 0, 0, 0, previa))
	time.Sleep(10 * time.Millisecond)
	humano("apruebo E001/F002 con push")
	// A3.2: la copia tardía va también ANTES de la temprana en el archivo (desorden de D000096).
	tardia := lineaUso("tardia", 15_000_000, 0, 0, 0, time.Now())
	escribirAntes(t, reg, tardia)
	anexar(t, reg, tardia, lineaUso("m8", 1, 0, 0, 0, time.Now()))
	if s := usar("ls"); s.codigo != 0 {
		t.Errorf("la copia tardía de una respuesta previa no cuenta; dio %+v", s)
	}
	// «tope 900,» con coma final sigue siendo de llamadas.
	time.Sleep(10 * time.Millisecond)
	humano("apruebo E001/F002 tope 900, con push")
	anexar(t, reg, lineaUso("m9", 19_000_000, 0, 0, 0, time.Now()))
	if s := usar("ls"); s.codigo != 0 {
		t.Errorf("«tope 900,» cuenta llamadas, no tokens; dio %+v", s)
	}
	// «tope 25 M» con espacio también es en tokens.
	time.Sleep(10 * time.Millisecond)
	humano("apruebo E001/F002 tope 25 M")
	anexar(t, reg, lineaUso("m7", 15_000_000, 0, 0, 0, time.Now()))
	nuevoTurno(t, humano, "sigue")
	anexar(t, reg, lineaUso("m7b", 10_000_001, 0, 0, 0, time.Now()))
	if s := usar("ls"); s.codigo != 2 || !strings.Contains(s.err, "tope de la feature") || !strings.Contains(s.err, "25M") {
		t.Errorf("«tope 25 M»: frena nombrando 25M; dio %+v", s)
	}
	// Feature CERRADO con sus slices CERRADO suelta el tope.
	cerrarFeature(t, raiz, "E001/F002")
	if s := usar("ls"); s.codigo != 0 {
		t.Errorf("feature cerrada: sin tope; dio %+v", s)
	}
	// Registro ilegible: frena con su ruta.
	e, _ := json.Marshal(map[string]any{"session_id": "s1", "tool_name": "Bash", "transcript_path": filepath.Join(raiz, "no.jsonl")})
	if s := aquas(t, raiz, string(e), "budget", "tool"); s.codigo != 2 || !strings.Contains(s.err, "no.jsonl") {
		t.Errorf("registro ausente: frena con su ruta; dio %+v", s)
	}
}
