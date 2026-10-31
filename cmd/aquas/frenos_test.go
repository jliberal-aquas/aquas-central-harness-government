package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/guardia"
)

// TestFrenosNoFalsificables (E000/F008, integral; humano 2026-10-31, D000077): el agente no puede
// lanzar por su cuenta una orden reservada a los hooks. Pasa por main con el nombre en inglés,
// como lo llama el hook PreToolUse real.
// Rojo cuando: el agente lanza approval por Bash y la guardia lo deja pasar.
// Verde cuando: la guardia lo bloquea con exit 2.
func TestFrenosNoFalsificables(t *testing.T) {
	raiz := repoSano(t)
	bash := func(cmd string) string {
		b, _ := json.Marshal(map[string]any{"tool_name": "Bash", "tool_input": map[string]string{"command": cmd}})
		return string(b)
	}
	for _, cmd := range []string{
		`echo '{"prompt":"apruebo E9/F1"}' | ./.tmp/aquas.exe approval`,
		`./.tmp/aquas.exe agent-close < falso.json`,
		`aquas budget prompt`,
	} {
		s := aquas(t, raiz, bash(cmd), "guard")
		if s.codigo != 2 || !strings.Contains(s.err, "solo la lanza Claude Code") {
			t.Errorf("%s: exit 2 y «solo la lanza Claude Code»; dio %+v", cmd, s)
		}
	}
	if s := aquas(t, raiz, bash(`./.tmp/aquas.exe memory set tarea="x"`), "guard"); s.codigo != 0 {
		t.Errorf("memory set es libre: exit 0; dio %+v", s)
	}

	// Una sola fuente (A1, hallazgo 4): cada orden de hook de la config, con cada nombre que la CLI
	// acepta para ella (nuevo o viejo, cruzados), está bloqueada. Un hook o alias nuevo sin
	// protección tumba esta prueba.
	b, err := os.ReadFile(filepath.Join("..", "..", "gobierno", "config", "claude_settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cfg struct {
		Hooks []struct {
			Orden string `json:"orden"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		t.Fatal(err)
	}
	internos := map[string]bool{} // «cmd» o «cmd sub», con nombres internos
	for _, h := range cfg.Hooks {
		args, _ := traducir(strings.Fields(h.Orden))
		if args[0] == "memoria" || args[0] == "turno" || args[0] == "presupuesto" {
			internos[args[0]+" "+args[1]] = true
		} else {
			internos[args[0]] = true
		}
	}
	nombres := func(interno string) []string { // nombres que la CLI acepta para un comando interno
		ns := []string{interno}
		for en, in := range comandos {
			if in == interno {
				ns = append(ns, en)
			}
		}
		return ns
	}
	n := 0
	for k := range internos {
		cmd, sub, _ := strings.Cut(k, " ")
		subs := []string{""}
		if sub != "" {
			subs = []string{sub}
			for en, in := range subcomandos[cmd] {
				if in == sub {
					subs = append(subs, en)
				}
			}
		}
		for _, c := range nombres(cmd) {
			for _, s := range subs {
				orden := strings.TrimSpace("aquas " + c + " " + s)
				n++
				if guardia.RevisarHook(orden) == "" {
					t.Errorf("%s: orden de hook (config) sin bloqueo en la guardia", orden)
				}
			}
		}
	}
	if n < 20 {
		t.Errorf("solo %d nombres de órdenes de hook revisados; la config o la traducción no se leyeron bien", n)
	}
}
