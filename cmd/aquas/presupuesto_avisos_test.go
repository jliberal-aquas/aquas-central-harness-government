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

// E001/F019/S002: sin freno de 60 llamadas por turno; avisos al 50% y al 80% del tope de la feature.

// avisoDe: el contexto adicional que el gancho entrega; falla si la salida no es un solo JSON.
func avisoDe(t *testing.T, s salida) string {
	t.Helper()
	if strings.TrimSpace(s.out) == "" {
		return ""
	}
	var r struct {
		H struct {
			Contexto string `json:"additionalContext"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal([]byte(s.out), &r); err != nil {
		t.Fatalf("la salida del gancho no es un solo JSON: %v; dio %q", err, s.out)
	}
	return r.H.Contexto
}

// llamadaSubagente: entrada del gancho para una llamada de subagente, sin escribirla en el registro.
func llamadaSubagente(t *testing.T, raiz, reg string) salida {
	t.Helper()
	b, _ := json.Marshal(map[string]any{"session_id": "s1", "agent_id": "a1", "hook_event_name": "PreToolUse",
		"tool_name": "Bash", "transcript_path": reg, "tool_input": map[string]string{"command": "ls"}})
	return aquas(t, raiz, string(b), "budget", "tool")
}

// sinFrenoYAvisoAlCincuenta (S002, integral): 61 llamadas del turno pasan y el 50% del tope avisa.
func sinFrenoYAvisoAlCincuenta(t *testing.T) {
	_, reg, humano, usar := presupuestoTokens(t)
	humano("sigue")
	for i := range 61 {
		anexar(t, reg, lineaLlamada(fmt.Sprintf("toolu_i%d", i), "Bash", "ls", time.Now()))
	}
	if s := usar("ls"); s.codigo != 0 {
		t.Fatalf("la llamada 61 del turno pasa: ya no hay freno de 60; dio %+v", s)
	}
	humano("apruebo E001/F002 con push")
	anexar(t, reg, lineaUso("m1", 7_500_000, 0, 0, 0, time.Now()))
	if s := usar("ls"); s.codigo != 0 || !strings.Contains(avisoDe(t, s), "50%") {
		t.Errorf("7,5M de 15M: aviso del 50%%; dio %+v", s)
	}
}

// TestAvisosDeGasto (E001/F019/S002): avisos al 50 y 80 por ciento, sin freno de 60.
// Rojo cuando: la llamada 61 se frena o no hay aviso.
// Verde cuando: 61 llamadas pasan y el aviso sale una vez en cada umbral.
func TestAvisosDeGasto(t *testing.T) {
	raiz, reg, humano, usar := presupuestoTokens(t)
	// Sin apruebo de feature: 61 llamadas del turno pasan y no hay aviso de gasto.
	humano("sigue")
	for i := range 61 {
		anexar(t, reg, lineaLlamada(fmt.Sprintf("toolu_t%d", i), "Bash", "ls", time.Now()))
	}
	s := usar("ls")
	if s.codigo != 0 {
		t.Fatalf("sin apruebo la llamada 61 del turno pasa; dio %+v", s)
	}
	if a := avisoDe(t, s); strings.Contains(a, "del tope") {
		t.Errorf("sin feature aprobada no hay aviso de gasto; dio %q", a)
	}

	humano("apruebo E001/F002 con push")
	anexar(t, reg, lineaUso("m1", 7_499_999, 0, 0, 0, time.Now()))
	if a := avisoDe(t, usar("ls")); strings.Contains(a, "%") {
		t.Errorf("bajo el 50%% no avisa; dio %q", a)
	}
	// Turno nuevo: el aviso de contexto (200k) y el del 50% salen juntos en un solo JSON.
	humano("sigue")
	anexar(t, reg, lineaUso("m2", 0, 0, 200_000, 0, time.Now())) // 10 mil ponderados; contexto medido 200k
	s = usar("ls")
	if a := avisoDe(t, s); s.codigo != 0 || !strings.Contains(a, "50%") || !strings.Contains(a, "E001/F002") ||
		!strings.Contains(a, "7,5M") || !strings.Contains(a, "15M") || !strings.Contains(a, "/compact") {
		t.Errorf("7,5M de 15M: aviso del 50%% con feature, gasto y tope, junto al de contexto; dio %+v", s)
	}
	if a := avisoDe(t, usar("ls")); strings.Contains(a, "%") {
		t.Errorf("el aviso del 50%% sale una vez; dio %q", a)
	}
	// Uso de subagente cuenta; su llamada no recibe avisos (ni el de gasto ni el de contexto, A1.3): los recibe la siguiente del hilo.
	humano("sigue")
	sub := filepath.Join(filepath.Dir(reg), "sesion", "subagents", "agent-a1.jsonl")
	anexar(t, sub, lineaUso("m3", 0, 0, 0, 900_000, time.Now())) // 4,5M ponderados → 12,0M
	if a := avisoDe(t, llamadaSubagente(t, raiz, reg)); a != "" {
		t.Errorf("el subagente no recibe avisos; dio %q", a)
	}
	if a := avisoDe(t, usar("ls")); !strings.Contains(a, "80%") || !strings.Contains(a, "12,0M de 15M") || !strings.Contains(a, "/compact") {
		t.Errorf("12M de 15M: aviso del 80%% en el hilo; dio %q", a)
	}
	if a := avisoDe(t, usar("ls")); strings.Contains(a, "%") {
		t.Errorf("el aviso del 80%% sale una vez; dio %q", a)
	}
	// Marca borrada o corrupta: el aviso se repite y no frena.
	dir := filepath.Join(raiz, ".tmp", "presupuesto-cuentas")
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
	if s := usar("ls"); s.codigo != 0 || !strings.Contains(avisoDe(t, s), "80%") {
		t.Errorf("sin marca el aviso se repite sin frenar; dio %+v", s)
	}
	marcas, _ := filepath.Glob(filepath.Join(dir, "gasto-*"))
	if len(marcas) != 1 {
		t.Fatalf("una marca de gasto en %s; hay %v", dir, marcas)
	}
	if err := os.WriteFile(marcas[0], []byte("\x00basura"), 0o644); err != nil {
		t.Fatal(err)
	}
	if s := usar("ls"); s.codigo != 0 || !strings.Contains(avisoDe(t, s), "80%") {
		t.Errorf("marca corrupta: el aviso se repite sin frenar; dio %+v", s)
	}
	// Sin carpeta escribible (la ruta es un archivo): el aviso sale igual.
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
	escribir(t, raiz, ".tmp/presupuesto-cuentas", "x")
	if s := usar("ls"); s.codigo != 0 || !strings.Contains(avisoDe(t, s), "80%") {
		t.Errorf("sin .tmp escribible el aviso sale igual; dio %+v", s)
	}
	if err := os.Remove(dir); err != nil {
		t.Fatal(err)
	}

	// Un apruebo nuevo reinicia los avisos: cuenta desde su hora.
	time.Sleep(10 * time.Millisecond)
	humano("apruebo E001/F002 con push")
	anexar(t, reg, lineaUso("m4", 8_000_000, 0, 0, 0, time.Now()))
	if a := avisoDe(t, usar("ls")); !strings.Contains(a, "50%") || strings.Contains(a, "80%") {
		t.Errorf("apruebo nuevo: 8M de 15M vuelve a avisar el 50%%; dio %q", a)
	}
	// Tope en llamadas («tope 10»): 5 llamadas avisan el 50% en llamadas.
	time.Sleep(10 * time.Millisecond)
	humano("apruebo E001/F002 tope 10")
	for i := range 5 {
		anexar(t, reg, lineaLlamada(fmt.Sprintf("toolu_c%d", i), "Bash", "ls", time.Now()))
	}
	if a := avisoDe(t, usar("ls")); !strings.Contains(a, "50%") || !strings.Contains(a, "5 de 10 llamadas") {
		t.Errorf("tope 10 con 5 llamadas: aviso del 50%% en llamadas; dio %q", a)
	}
	// Feature cerrada: sin tope y sin avisos.
	for i := range 4 {
		anexar(t, reg, lineaLlamada(fmt.Sprintf("toolu_d%d", i), "Bash", "ls", time.Now()))
	}
	cerrarFeature(t, raiz, "E001/F002")
	if s := usar("ls"); s.codigo != 0 || strings.Contains(avisoDe(t, s), "%") {
		t.Errorf("feature cerrada: sin aviso; dio %+v", s)
	}
}
