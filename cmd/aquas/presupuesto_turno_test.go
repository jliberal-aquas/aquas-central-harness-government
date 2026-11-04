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

// E001/F019/S004: tope extremo por turno, con o sin feature aprobada.

// llamadasEn anexa n llamadas Bash al registro con la hora actual.
func llamadasEn(t *testing.T, ruta, prefijo string, n int) {
	t.Helper()
	lineas := make([]string, n)
	ahora := time.Now()
	for i := range n {
		lineas[i] = lineaLlamada(fmt.Sprintf("toolu_%s%d", prefijo, i), "Bash", "ls", ahora)
	}
	anexar(t, ruta, lineas...)
}

// subRegistro: el registro de un subagente de la sesión de la prueba.
func subRegistro(reg, agente string) string {
	return filepath.Join(filepath.Dir(reg), "sesion", "subagents", "agent-"+agente+".jsonl")
}

// frenaElTurno: la salida es el freno extremo del turno.
func frenaElTurno(s salida) bool {
	return s.codigo == 2 && strings.Contains(s.err, "informa al humano y espera")
}

// nuevoTurno: mensaje humano que abre turno, separado en el tiempo de lo anterior.
func nuevoTurno(t *testing.T, humano func(string), texto string) {
	t.Helper()
	time.Sleep(10 * time.Millisecond)
	humano(texto)
	time.Sleep(10 * time.Millisecond)
}

// turnoDesbocadoFrena (S004, integral): sin feature aprobada, la llamada 1001 del turno frena.
func turnoDesbocadoFrena(t *testing.T) {
	_, reg, humano, usar := presupuestoTokens(t)
	nuevoTurno(t, humano, "sigue")
	llamadasEn(t, reg, "d", 1000)
	if s := usar("ls"); s.codigo != 0 {
		t.Fatalf("1000 llamadas del turno pasan; dio %+v", s)
	}
	llamadasEn(t, reg, "e", 1)
	if s := usar("ls"); !frenaElTurno(s) {
		t.Errorf("la llamada 1001 del turno frena; dio %+v", s)
	}
}

// TestTopeExtremo (E001/F019/S004): un turno se frena al pasar 1000 llamadas o 20M tokens ponderados.
// Rojo cuando: el turno pasa de 1000 llamadas o 20M tokens sin freno.
// Verde cuando: la llamada 1001 o el token 20M frenan, con o sin feature, contando subagentes.
func TestTopeExtremo(t *testing.T) {
	raiz, reg, humano, usar := presupuestoTokens(t)

	// Sin apruebo: 1000 llamadas pasan; la 1001 frena y dice lo gastado.
	nuevoTurno(t, humano, "sigue")
	llamadasEn(t, reg, "a", 1000)
	if s := usar("ls"); s.codigo != 0 {
		t.Fatalf("1000 llamadas pasan; dio %+v", s)
	}
	llamadasEn(t, reg, "b", 1)
	s := usar("ls")
	if !frenaElTurno(s) || !strings.Contains(s.err, "1001 llamadas") {
		t.Fatalf("la llamada 1001 frena nombrando lo gastado; dio %+v", s)
	}
	// La llamada del subagente también se frena.
	if s := llamadaSubagente(t, raiz, reg); !frenaElTurno(s) {
		t.Errorf("subagente frenado por el tope del turno; dio %+v", s)
	}
	// A1.2: con transcript_path en el registro del propio subagente, sube al de su sesión y frena igual.
	sub := subRegistro(reg, "propio")
	anexar(t, sub, lineaLlamada("toolu_propio", "Bash", "ls", time.Now()))
	if s := aquas(t, raiz, llamar(t, "", sub, "Bash", "ls"), "budget", "tool"); !frenaElTurno(s) {
		t.Errorf("registro de subagente: cuenta el turno de su sesión; dio %+v", s)
	}
	// Las órdenes de cierre de S003 no lo saltan; memory set y defect new|occur sí.
	for _, c := range []string{"git push", `git commit -m "x"`, "aquas verify", "aquas backlog state E001/F002/S001 CERRADO"} {
		if s := usar(c); !frenaElTurno(s) {
			t.Errorf("%q se frena con el turno agotado; dio %+v", c, s)
		}
	}
	for _, c := range []string{"aquas memory set tarea=x", "aquas defect new firma=x", "aquas defect occur D000001"} {
		if s := usar(c); s.codigo != 0 {
			t.Errorf("%q exento del tope del turno; dio %+v", c, s)
		}
	}
	// Un mensaje humano sin hora, una notificación o un mensaje de agente no abren turno.
	sinHora, _ := json.Marshal(map[string]any{"type": "user", "message": map[string]string{"role": "user", "content": "sigue"},
		"origin": map[string]string{"kind": "human"}})
	anexar(t, reg, string(sinHora))
	for _, texto := range []string{"x\n<task-notification>y", "x\n<agent-message>y"} {
		b, _ := json.Marshal(map[string]any{"type": "user", "timestamp": time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
			"message": map[string]string{"role": "user", "content": texto}, "origin": map[string]string{"kind": "human"}})
		anexar(t, reg, string(b))
	}
	if s := usar("ls"); !frenaElTurno(s) {
		t.Errorf("sin hora, notificación o mensaje de agente no abren turno; dio %+v", s)
	}

	// Un mensaje humano abre turno nuevo: cuenta desde cero.
	nuevoTurno(t, humano, "sigue")
	if s := usar("ls"); s.codigo != 0 {
		t.Fatalf("turno nuevo pasa; dio %+v", s)
	}

	// Subagentes del turno suman llamadas: 500 del hilo y 501 de dos subagentes.
	llamadasEn(t, reg, "c", 500)
	llamadasEn(t, subRegistro(reg, "s1"), "s", 250)
	llamadasEn(t, subRegistro(reg, "s2"), "u", 250)
	if s := usar("ls"); s.codigo != 0 {
		t.Fatalf("1000 llamadas con subagentes pasan; dio %+v", s)
	}
	llamadasEn(t, subRegistro(reg, "s2"), "v", 1)
	if s := usar("ls"); !frenaElTurno(s) {
		t.Errorf("1001 llamadas con subagentes frenan; dio %+v", s)
	}

	// Tokens: 20M ponderados pasan, 20M+1 frenan; un message.id repetido cuenta una vez; el subagente suma.
	nuevoTurno(t, humano, "sigue")
	ahora := time.Now()
	anexar(t, reg, lineaUso("t1", 10_000_000, 0, 0, 0, ahora), lineaUso("t1", 10_000_000, 0, 0, 0, ahora))
	anexar(t, subRegistro(reg, "s3"), lineaUso("t2", 0, 0, 0, 2_000_000, ahora)) // salida ×5: 10M
	if s := usar("ls"); s.codigo != 0 {
		t.Fatalf("20M ponderados del turno pasan; dio %+v", s)
	}
	anexar(t, reg, lineaUso("t3", 1, 0, 0, 0, time.Now()))
	if s := usar("ls"); !frenaElTurno(s) || !strings.Contains(s.err, "20M") {
		t.Errorf("20M+1 ponderados frenan nombrando 20M; dio %+v", s)
	}

	// Con feature aprobada y su tope lejos, el tope del turno frena igual.
	nuevoTurno(t, humano, "apruebo E001/F002 tope 5000")
	llamadasEn(t, reg, "f", 1001)
	if s := usar("ls"); !frenaElTurno(s) {
		t.Errorf("feature abierta: el tope del turno frena igual; dio %+v", s)
	}
	// A1.1: con el tope de la feature también agotado, las órdenes de cierre de S003 no saltan el del turno.
	nuevoTurno(t, humano, "apruebo E001/F002 tope 5")
	llamadasEn(t, reg, "g", 1001)
	for _, c := range []string{"git push", `git commit -m "x"`} {
		if s := usar(c); !frenaElTurno(s) {
			t.Errorf("%q: tope de feature y de turno agotados; frena el del turno; dio %+v", c, s)
		}
	}
	// Con la feature cerrada, también.
	cerrarFeature(t, raiz, "E001/F002")
	if s := usar("ls"); !frenaElTurno(s) {
		t.Errorf("feature cerrada: el tope del turno frena igual; dio %+v", s)
	}

	// A1.3: una sesión sin mensaje humano es un solo turno: 1001 llamadas frenan.
	otra := filepath.Join(filepath.Dir(reg), "otra.jsonl")
	llamadasEn(t, otra, "h", 1001)
	if s := aquas(t, raiz, llamar(t, "", otra, "Bash", "ls"), "budget", "tool"); !frenaElTurno(s) {
		t.Errorf("sesión sin mensaje humano: frena al pasar 1000; dio %+v", s)
	}
	if err := os.Remove(otra); err != nil {
		t.Fatal(err)
	}

	// Un registro de subagente ilegible frena con su ruta, sin feature aprobada.
	nuevoTurno(t, humano, "sigue")
	roto := subRegistro(reg, "roto")
	if err := os.MkdirAll(roto, 0o755); err != nil { // una carpeta con nombre de registro no se puede leer
		t.Fatal(err)
	}
	if s := usar("ls"); s.codigo != 2 || !strings.Contains(s.err, "agent-roto.jsonl") {
		t.Errorf("registro ilegible: frena nombrándolo; dio %+v", s)
	}
}
