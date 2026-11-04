package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// contextoGancho: el additionalContext que el gancho PostToolUse entrega a Claude por stdout ("" si no hay).
func contextoGancho(t *testing.T, out string) string {
	t.Helper()
	if strings.TrimSpace(out) == "" {
		return ""
	}
	var r struct {
		HookSpecificOutput struct {
			HookEventName     string `json:"hookEventName"`
			AdditionalContext string `json:"additionalContext"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal([]byte(out), &r); err != nil {
		t.Fatalf("stdout del gancho no es JSON: %v\n%s", err, out)
	}
	if r.HookSpecificOutput.HookEventName != "PostToolUse" {
		t.Fatalf("hookEventName = %q; esperado PostToolUse", r.HookSpecificOutput.HookEventName)
	}
	return r.HookSpecificOutput.AdditionalContext
}

// TestGanchoOracionAvisa (E000/F013/S002): al guardar en docs/, oración larga y enlaces agregados llegan como aviso, sin frenar.
// Rojo cuando: el gancho sale con código 2.
// Verde cuando: sale 0 y el aviso va en additionalContext; sin nada que avisar, calla.
func TestGanchoOracionAvisa(t *testing.T) {
	raiz := t.TempDir()
	evento := func(rel string) string {
		b, _ := json.Marshal(map[string]any{"hook_event_name": "PostToolUse", "tool_input": map[string]string{"file_path": filepath.Join(raiz, filepath.FromSlash(rel))}})
		return string(b)
	}
	larga := strings.TrimSpace(strings.Repeat("sol ", 21))
	escribir(t, raiz, "docs/glosario.md", "# Glosario\n\n## agente\n\nQuien trabaja.\n\nEjemplo: el worker.\n")

	// Página de slice: enlace agregado y oración larga en el mismo guardado.
	escribir(t, raiz, "docs/E000/F001/S001.md", "# Título\n\nUn agente corto.\n"+larga+".\n")
	s := aquas(t, raiz, evento("docs/E000/F001/S001.md"), "hook")
	if s.codigo != 0 {
		t.Fatalf("el gancho no debe frenar; exit %d, stderr %q", s.codigo, s.err)
	}
	c := contextoGancho(t, s.out)
	if !strings.Contains(c, "S001.md:4: oración de 21 palabras") || !strings.Contains(c, "1 enlaces al glosario") {
		t.Errorf("aviso de oración larga y enlace agregado: %q", c)
	}
	if b, _ := os.ReadFile(filepath.Join(raiz, "docs", "E000", "F001", "S001.md")); !strings.Contains(string(b), "[agente](../../glosario.md#agente)") {
		t.Error("el enlazado del glosario debe seguir agregando el enlace")
	}

	// Glosario: se mide al guardarlo, también sin frenar.
	escribir(t, raiz, "docs/glosario.md", "# Glosario\n\n## agente\n\n"+larga+".\n")
	s = aquas(t, raiz, evento("docs/glosario.md"), "hook")
	if s.codigo != 0 || !strings.Contains(contextoGancho(t, s.out), "glosario.md:5: oración de 21 palabras") {
		t.Errorf("glosario con oración larga: exit %d, out %q", s.codigo, s.out)
	}

	// Sin nada que avisar: sale 0 y no escribe nada.
	escribir(t, raiz, "docs/E000/F001/S002.md", "# Título\n\nUna corta.\n")
	if s := aquas(t, raiz, evento("docs/E000/F001/S002.md"), "hook"); s.codigo != 0 || strings.TrimSpace(s.out+s.err) != "" {
		t.Errorf("página corta: el gancho debe callar; %+v", s)
	}
}
