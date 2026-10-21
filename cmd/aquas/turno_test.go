package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// E001/F002/S001: por el CLI real, el hook abre el turno con el literal, el cierre guarda la
// respuesta, nada de eso sale por stdout (no gasta contexto) y auditoria/ queda fuera de Git.
func TestTurnoCLIDePuntaAPunta(t *testing.T) {
	raiz := repoSano(t)
	os.WriteFile(filepath.Join(raiz, ".gitignore"), mustLeer(t, filepath.Join("..", "..", ".gitignore")), 0o644)
	s := aquas(t, raiz, `{"session_id":"s1","hook_event_name":"UserPromptSubmit","prompt":"haz X\n`+"```"+`y`+"```"+`"}`, "turno", "abrir")
	if s.codigo != 0 || s.out != "" {
		t.Fatalf("abrir: exit 0 sin stdout (el stdout entraría al contexto); dio %+v", s)
	}
	mes := time.Now().UTC().Format("2006-01")
	rel := filepath.Join("auditoria", "turnos", mes, "T000001.md")
	c := string(mustLeer(t, filepath.Join(raiz, rel)))
	if !strings.Contains(c, "````text\nhaz X\n```y```\n````") {
		t.Fatalf("el literal llega intacto desde el JSON del hook:\n%s", c)
	}
	s = aquas(t, raiz, `{"session_id":"s1","hook_event_name":"Stop","stop_hook_active":false,"last_assistant_message":"Hecho X."}`, "turno", "cerrar")
	if s.codigo != 0 || s.out != "" {
		t.Fatalf("cerrar: exit 0 sin stdout; dio %+v", s)
	}
	if c := string(mustLeer(t, filepath.Join(raiz, rel))); !strings.Contains(c, "```text\nHecho X.\n```") {
		t.Fatalf("la respuesta queda en el turno:\n%s", c)
	}
	if s := aquas(t, raiz, `no es json`, "turno", "abrir"); s.codigo != 1 || !strings.Contains(s.err, "turno abrir") {
		t.Errorf("entrada rota: exit 1 (no bloquea el mensaje) y lo dice por stderr; dio %+v", s)
	}
	out, err := exec.Command("git", "-C", raiz, "check-ignore", "-q", filepath.ToSlash(rel)).CombinedOutput()
	if err != nil {
		t.Errorf("auditoria/ fuera de Git (decisión del canónico): check-ignore %v %s", err, out)
	}
}

func mustLeer(t *testing.T, p string) []byte {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
