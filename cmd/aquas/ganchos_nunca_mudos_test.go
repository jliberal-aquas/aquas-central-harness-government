package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGanchosNuncaMudos (E000/F014/S002, D000125): ningún gancho falla en silencio.
// Claude Code no muestra el exit 1 ni el stderr de un exit 0 (code.claude.com/docs/en/hooks, 2026-11-04).
// Rojo cuando: un gancho sale con exit 1, o con exit 0 y stderr sin aviso en stdout.
// Verde cuando: cada gancho con entrada rota avisa (exit 0 con systemMessage) o frena con exit 2 y su causa.
func TestGanchosNuncaMudos(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "gobierno", "config", "claude_settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cfg struct {
		Hooks []struct {
			Evento string `json:"evento"`
			Orden  string `json:"orden"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(b, &cfg); err != nil || len(cfg.Hooks) == 0 {
		t.Fatalf("configuración de ganchos ilegible o vacía: %v", err)
	}
	raiz := repoContrato(t)
	for _, h := range cfg.Hooks {
		args, _ := traducir(strings.Fields(h.Orden))
		if ev, ok := eventoGancho(args); !ok || ev != h.Evento {
			t.Errorf("%s: el gancho nuevo debe estar en ganchos con su evento %s; dio %q, %v", h.Orden, h.Evento, ev, ok)
		}
		for _, entrada := range []string{"", "{", "no es json"} {
			s := aquas(t, raiz, entrada, args...)
			switch {
			case s.codigo == 2 && strings.TrimSpace(s.err) == "":
				t.Errorf("%s con %q: frena (exit 2) sin decir la causa", h.Orden, entrada)
			case s.codigo != 0 && s.codigo != 2:
				t.Errorf("%s con %q: exit %d es mudo para Claude Code; dio %+v", h.Orden, entrada, s.codigo, s)
			case s.codigo == 0 && hayFallo(s.err) && !strings.Contains(s.out, `"systemMessage"`):
				t.Errorf("%s con %q: exit 0 con stderr y sin aviso en stdout: nadie lo ve; dio %+v", h.Orden, entrada, s)
			}
		}
	}

	// La aprobación que no se puede guardar avisa al agente con la causa, en su contexto.
	if err := os.Remove(filepath.Join(raiz, "docs", "E001", "F002", "S003.md")); err != nil {
		t.Fatal(err)
	}
	s := aquas(t, raiz, mensajeHumano(t, "apruebo E001/F002/S003"), "approval")
	var out struct {
		Sistema string `json:"systemMessage"`
		Hook    struct {
			Evento   string `json:"hookEventName"`
			Contexto string `json:"additionalContext"`
		} `json:"hookSpecificOutput"`
	}
	if s.codigo != 0 || json.Unmarshal([]byte(s.out), &out) != nil {
		t.Fatalf("aprobación sin página: exit 0 con JSON en stdout; dio %+v", s)
	}
	for _, m := range []string{out.Sistema, out.Hook.Contexto} {
		if !strings.Contains(m, "aprobación NO registrada") || !strings.Contains(m, "falta la página del contrato docs/E001/F002/S003.md") {
			t.Errorf("el aviso nombra que no se registró y la causa; dio %q", m)
		}
	}
	if out.Hook.Evento != "UserPromptSubmit" {
		t.Errorf("additionalContext exige hookEventName UserPromptSubmit; dio %q", out.Hook.Evento)
	}

	// Aprobación de feature que falla (sin la página del slice) y ficha ilegible al aprobar: avisan igual (A1.2).
	for _, c := range []struct{ prompt, romper, causa string }{
		{"apruebo E001/F002", "", "falta la página del contrato docs/E001/F002/S003.md"},
		{"apruebo E001/F002/S003", "trabajo/activo/E001-F002-S003.json", "E001-F002-S003.json"},
	} {
		if c.romper != "" {
			escribir(t, raiz, c.romper, "{roto")
		}
		s := aquas(t, raiz, mensajeHumano(t, c.prompt), "approval")
		if s.codigo != 0 || !strings.Contains(s.out, "aprobación NO registrada") || !strings.Contains(s.out, c.causa) {
			t.Errorf("%s que falla: exit 0 con «aprobación NO registrada» y la causa %q; dio %+v", c.prompt, c.causa, s)
		}
	}

	// Un freno real (cierre de auditor sin sus campos) atraviesa el envoltorio con exit 2, sin volverse aviso (A1.2).
	if s := aquas(t, raiz, cierreAuditor(t, "auditor", "sin campos", false), "agent-close"); s.codigo != 2 ||
		strings.TrimSpace(s.err) == "" || strings.Contains(s.out, `"systemMessage"`) {
		t.Errorf("freno con exit 2: pasa tal cual con su causa; dio %+v", s)
	}

	// Exit 0 con stderr (ficha rota al anotar el veredicto): el aviso viaja como systemMessage.
	escribir(t, raiz, "trabajo/activo/E001-F014-S002.json", "{roto")
	cuerpo := "id: E001-F014-S002-A1\nveredicto: APROBADO\nhallazgos: —\nno_revisado: —\n"
	if s := aquas(t, raiz, cierreAuditor(t, "auditor", cuerpo, true), "agent-close"); s.codigo != 0 ||
		!strings.Contains(s.out, `"systemMessage"`) || !strings.Contains(s.out, "no quedó anotado") {
		t.Errorf("exit 0 con stderr: el aviso llega en systemMessage; dio %+v", s)
	}

	// La guardia falla cerrada: su fallo frena con exit 2 y la causa; los demás avisan con exit 0.
	if c, o, e := fallaGancho("guardia", errors.New("rota")); c != 2 || o != "" || !strings.Contains(e, "rota") {
		t.Errorf("guardia que falla: exit 2 con la causa en stderr; dio %d %q %q", c, o, e)
	}
	if c, o, _ := fallaGancho("propuesta", errors.New("rota")); c != 0 || !strings.Contains(o, `"systemMessage"`) {
		t.Errorf("gancho Stop que falla: exit 0 con systemMessage; dio %d %q", c, o)
	}
}

// hayFallo: stderr con algo más que el aviso de nombre viejo (no es un fallo del gancho).
func hayFallo(stderr string) bool {
	for _, l := range strings.Split(stderr, "\n") {
		if l = strings.TrimSpace(l); l != "" && !strings.Contains(l, avisoNombreViejo) {
			return true
		}
	}
	return false
}
