package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRetrospectivaSinReincidencia (E001/F017): cada familia de la retrospectiva del 2026-11-02 tiene su control,
// con un caso que debe fallar y uno que debe pasar. Crece con cada slice cerrado.
// Rojo cuando: una familia con slice cerrado vuelve a pasar sin freno.
// Verde cuando: el control de cada slice cerrado frena su caso malo y deja pasar el bueno.
func TestRetrospectivaSinReincidencia(t *testing.T) {
	t.Run("S001 freno atado al trabajo activo", TestFrenoSigueAlTrabajoActivo)
	t.Run("S002 no pedir apruebo ya cubierto", TestNoPideAprueboCubierto)
	t.Run("S003 turnos separan humano de agente", turnosSeparanOrigen)
}

// turnosSeparanOrigen (S003) sobre el binario: el hook de turnos guarda el origen de cada mensaje.
func turnosSeparanOrigen(t *testing.T) {
	raiz := t.TempDir()
	for _, c := range []struct{ sesion, prompt, origen string }{
		{"h", "dale, sigue", "humano"},
		{"a", "<agent-message from=\"x\">informe</agent-message>", "agente"},
	} {
		b, _ := json.Marshal(map[string]string{"session_id": c.sesion, "prompt": c.prompt})
		if s := aquas(t, raiz, string(b), "turn", "open"); s.codigo != 0 {
			t.Fatalf("turn open: %+v", s)
		}
	}
	turnos, _ := filepath.Glob(filepath.Join(raiz, "auditoria", "turnos", "*", "T*.md"))
	var origenes []string
	for _, f := range turnos {
		b, _ := os.ReadFile(f)
		for _, o := range []string{"humano", "agente"} {
			if strings.Contains(string(b), "- Origen: "+o+"\n") {
				origenes = append(origenes, o)
			}
		}
	}
	if strings.Join(origenes, ",") != "humano,agente" {
		t.Errorf("los dos turnos deben quedar con origen humano y agente, en ese orden; dio %v", origenes)
	}
}
