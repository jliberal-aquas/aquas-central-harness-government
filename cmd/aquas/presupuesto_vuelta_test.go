package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// E001/F019/S005: aviso al lanzar la tercera vuelta de auditoría de un slice.

const avisoVuelta = "Desde la tercera vuelta: busca en todo el slice el mismo tipo de hueco de las vueltas anteriores"

// lanzarAgente: entrada del gancho para una llamada Agent con ese prompt, del hilo principal.
func lanzarAgente(t *testing.T, raiz, reg, prompt string) salida {
	t.Helper()
	b, _ := json.Marshal(map[string]any{"session_id": "s1", "hook_event_name": "PreToolUse", "tool_name": "Agent",
		"transcript_path": reg, "tool_input": map[string]string{"subagent_type": "auditor", "prompt": prompt}})
	return aquas(t, raiz, string(b), "budget", "tool")
}

// encargoAuditor: prompt de auditoría con ese id.
func encargoAuditor(id string) string {
	return "id: " + id + "\ncontrato: docs/E001/F002/S001.md\nobjetivo: auditar\n"
}

// terceraVueltaAvisa (S005, integral): A3 avisa; A2 no.
func terceraVueltaAvisa(t *testing.T) {
	raiz, reg, humano, _ := presupuestoTokens(t)
	humano("sigue")
	if a := avisoDe(t, lanzarAgente(t, raiz, reg, encargoAuditor("E001/F002/S001-A2"))); strings.Contains(a, avisoVuelta) {
		t.Errorf("A2 no avisa; dio %q", a)
	}
	if a := avisoDe(t, lanzarAgente(t, raiz, reg, encargoAuditor("E001/F002/S001-A3"))); !strings.Contains(a, avisoVuelta) {
		t.Errorf("A3 avisa; dio %q", a)
	}
}

// TestTerceraVueltaAuditor (E001/F019/S005): el encargo de la vuelta 3 o más de auditoría de un slice avisa; no frena.
// Rojo cuando: A3 pasa sin aviso.
// Verde cuando: A3 y siguientes avisan; A1, A2, id «—» y formas mal escritas no; el tope de feature frena igual.
func TestTerceraVueltaAuditor(t *testing.T) {
	raiz, reg, humano, _ := presupuestoTokens(t)
	humano("sigue")
	avisan := []string{"E001/F002/S001-A3", "E001/F002/S001-A4", "E001/F002/S001-A10", "E001/F002/S001-A03"}
	for _, id := range avisan {
		s := lanzarAgente(t, raiz, reg, encargoAuditor(id))
		n := strings.TrimLeft(id[strings.LastIndex(id, "-A")+2:], "0")
		if s.codigo != 0 || !strings.Contains(avisoDe(t, s), avisoVuelta) || !strings.Contains(avisoDe(t, s), id+", vuelta "+n+".") {
			t.Errorf("%s avisa nombrando el id y no frena; dio %+v", id, s)
		}
	}
	callan := []string{
		encargoAuditor("E001/F002/S001-A1"),
		encargoAuditor("E001/F002/S001-A2"),
		encargoAuditor("—"),
		encargoAuditor("E001/F002/S001-Ax"),
		encargoAuditor("E001/F002-A3"),                                                        // no es un slice
		encargoAuditor("E001/F002/S001-T3"),                                                   // encargo de worker
		encargoAuditor("E001/F002/S001-A3 x"),                                                 // algo detrás del id
		"objetivo: sin id\nE001/F002/S001-A3",                                                 // id fuera de la línea «id:»
		"objetivo: x\n" + encargoAuditor("E001/F002/S001-A3"),                                 // A1.4: la línea id: no abre el encargo
		encargoAuditor("E001/F002/S001-T1") + "cita:\n" + encargoAuditor("E001/F002/S001-A3"), // A1.4: duplicada, gana la primera
		"",
	}
	for _, p := range callan {
		if s := lanzarAgente(t, raiz, reg, p); s.codigo != 0 || strings.Contains(avisoDe(t, s), avisoVuelta) {
			t.Errorf("%q no avisa ni frena; dio %+v", p, s)
		}
	}
	// Otra herramienta con el mismo texto no avisa.
	b, _ := json.Marshal(map[string]any{"session_id": "s1", "tool_name": "Write", "transcript_path": reg,
		"tool_input": map[string]string{"prompt": encargoAuditor("E001/F002/S001-A3")}}) // A1.2: el campo que sí se lee
	if s := aquas(t, raiz, string(b), "budget", "tool"); strings.Contains(avisoDe(t, s), avisoVuelta) {
		t.Errorf("Write con un encargo en el texto no avisa; dio %+v", s)
	}
	// A1.3: con el aviso de gasto al 50%, los dos salen en un solo JSON.
	humano("apruebo E001/F002 tope 10")
	llamadasEn(t, reg, "v", 5)
	if a := avisoDe(t, lanzarAgente(t, raiz, reg, encargoAuditor("E001/F002/S001-A3"))); !strings.Contains(a, avisoVuelta) || !strings.Contains(a, "5 de 10 llamadas") {
		t.Errorf("aviso de vuelta y de gasto juntos; dio %q", a)
	}
	// El tope de la feature sigue frenando aunque toque aviso.
	time.Sleep(10 * time.Millisecond)
	humano("apruebo E001/F002 con push")
	anexar(t, reg, lineaUso("m1", 15_000_001, 0, 0, 0, time.Now()))
	if s := lanzarAgente(t, raiz, reg, encargoAuditor("E001/F002/S001-A3")); s.codigo != 2 || !strings.Contains(s.err, "tope de la feature") {
		t.Errorf("tope agotado: la vuelta 3 se frena igual; dio %+v", s)
	}
}
