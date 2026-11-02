package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
	"time"
)

// humanaEn: mensaje humano con hora fija.
func humanaEn(texto string, cuando time.Time) string {
	b, _ := json.Marshal(map[string]any{"type": "user", "message": map[string]string{"role": "user", "content": texto},
		"origin": map[string]string{"kind": "human"}, "timestamp": cuando.UTC().Format("2006-01-02T15:04:05.000Z")})
	return string(b)
}

// TestFrenoTurnoPorHora (E000/F010/X01, D000096): Claude Code escribe a veces el mensaje humano nuevo antes de
// líneas más viejas. El turno del freno y el último apruebo de feature se eligen por hora, no por posición.
// Rojo cuando: cuenta llamadas de un turno viejo o toma un apruebo viejo.
// Verde cuando: cuenta solo las llamadas desde el mensaje humano más reciente y toma el apruebo más reciente.
func TestFrenoTurnoPorHora(t *testing.T) {
	raiz := t.TempDir()
	t0 := time.Date(2026, 11, 1, 23, 0, 0, 0, time.UTC)
	t1, t2, t3 := t0.Add(time.Hour), t0.Add(12*time.Hour), t0.Add(13*time.Hour)
	ruta := filepath.Join(t.TempDir(), "sesion.jsonl")
	lineas := []string{humanaEn("apruebo E000/F002", t2), humanaEn("apruebo E000/F001", t0)}
	for i := 0; i < 70; i++ { // turno viejo, escrito después del mensaje nuevo
		lineas = append(lineas, lineaLlamada(fmt.Sprintf("viejo_%d", i), "Bash", "git status", t1))
	}
	for i := 0; i < 5; i++ { // turno abierto
		lineas = append(lineas, lineaLlamada(fmt.Sprintf("nuevo_%d", i), "Bash", "git status", t3))
	}
	anexar(t, ruta, lineas...)
	e, err := escanear(raiz, ruta)
	if err != nil {
		t.Fatal(err)
	}
	if n := len(e.turno); n != 5 {
		t.Errorf("el turno abierto tiene 5 llamadas; contó %d", n)
	}
	if e.apruebo == nil || e.apruebo.feature != "E000/F002" {
		t.Errorf("el apruebo vigente es el de hora más reciente (E000/F002); dio %+v", e.apruebo)
	}
	if len(e.llamadas) != 75 {
		t.Errorf("el tope de la feature sigue viendo todas las llamadas (75); vio %d", len(e.llamadas))
	}
}
