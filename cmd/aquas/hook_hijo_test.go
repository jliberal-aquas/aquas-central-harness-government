package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/hijo"
)

// TestHookDelHijoCorreBuild (E002/F002/S006): el comando real del hook, en bash y desde otra carpeta,
// rearma el CLAUDE.md del hijo; si falla, sale 2 para que Claude Code muestre el error al modelo.
func TestHookDelHijoCorreBuild(t *testing.T) {
	dir := t.TempDir()
	if s := aquas(t, dir, "", "install"); s.codigo != 0 {
		t.Fatalf("install: exit %d\n%s%s", s.codigo, s.out, s.err)
	}
	b, err := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	var s struct {
		Hooks struct {
			PostToolUse []struct {
				Hooks []struct{ Shell, Command string } `json:"hooks"`
			} `json:"PostToolUse"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(b, &s); err != nil {
		t.Fatal(err)
	}
	var cmd string
	for _, g := range s.Hooks.PostToolUse {
		for _, h := range g.Hooks {
			if h.Command == hijo.ComandoBuild && h.Shell == "bash" {
				cmd = h.Command
			}
		}
	}
	if cmd == "" {
		t.Fatalf("install no dejó el hook de build en bash:\n%s", b)
	}
	// aquas en el PATH: copia del binario de prueba.
	bin := t.TempDir()
	nombre := "aquas"
	if runtime.GOOS == "windows" {
		nombre += ".exe"
	}
	exe, err := os.ReadFile(os.Args[0])
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bin, nombre), exe, 0o755); err != nil {
		t.Fatal(err)
	}
	correr := func(proyecto string) (int, string) {
		c := exec.Command("bash", "-c", cmd)
		c.Dir = t.TempDir() // otra carpeta: el hook no depende de la actual
		c.Env = append(os.Environ(), "AQUAS_COMO_CLI=1", "CLAUDE_PROJECT_DIR="+proyecto,
			"PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"))
		out, err := c.CombinedOutput()
		if x, ok := err.(*exec.ExitError); ok {
			return x.ExitCode(), string(out)
		} else if err != nil {
			t.Fatal(err)
		}
		return 0, string(out)
	}
	for n, txt := range map[string]string{"titulo": "# CONSTITUCIÓN · prueba\n", "mision": "- Landing.\n", "mapa": "- `web/` → sitio.\n"} {
		if err := os.WriteFile(filepath.Join(dir, hijo.CarpetaHuecos, n+".md"), []byte(txt), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if m := hijo.Verificar(dir); len(m) == 0 {
		t.Fatal("antes del hook, verificar debió fallar por CLAUDE.md desfasado")
	}
	if c, out := correr(dir); c != 0 {
		t.Fatalf("hook en el hijo: exit %d\n%s", c, out)
	}
	if m := hijo.Verificar(dir); len(m) != 0 {
		t.Errorf("tras el hook, verificar debió pasar; dio %v", m)
	}
	if c, out := correr(filepath.Join(dir, "no-existe")); c != 2 || !strings.Contains(out, "aquas:") {
		t.Errorf("build fallido: se esperaba exit 2 con el error; dio %d\n%s", c, out)
	}
}
