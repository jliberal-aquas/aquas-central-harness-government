package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const mainBueno = "package main\n\nimport \"fmt\"\n\nfunc main() { fmt.Println(\"ok\") }\n"
const mainRoto = "package main\n\nfunc main() { no compila }\n"

// TestHookFallaCerrado: si aquas no compila, el hook no deja pasar a ciegas:
// usa la última versión que compiló o, sin ella, bloquea con exit 2.
func TestHookFallaCerrado(t *testing.T) {
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Fatal("sin sh en el PATH:", err)
	}
	guion, err := filepath.Abs("../../.claude/hooks/aquas.sh")
	if err != nil {
		t.Fatal(err)
	}
	proj := t.TempDir()
	escribir := func(rel, txt string) {
		t.Helper()
		p := filepath.Join(proj, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(txt), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	correr := func() (int, string, string) {
		t.Helper()
		c := exec.Command(sh, guion, "hook")
		c.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj)
		var out, errb strings.Builder
		c.Stdout, c.Stderr = &out, &errb
		err := c.Run()
		var ee *exec.ExitError
		switch {
		case err == nil:
			return 0, out.String(), errb.String()
		case errors.As(err, &ee):
			return ee.ExitCode(), out.String(), errb.String()
		}
		t.Fatal(err)
		return 0, "", ""
	}
	escribir("go.mod", "module x\n\ngo 1.26\n")

	escribir("cmd/aquas/main.go", mainRoto)
	if c, _, e := correr(); c != 2 {
		t.Fatalf("roto y sin versión anterior: exit %d, quiero 2 (bloquea); stderr: %s", c, e)
	}

	escribir("cmd/aquas/main.go", mainBueno)
	if c, o, e := correr(); c != 0 || strings.TrimSpace(o) != "ok" {
		t.Fatalf("compila: exit %d, salida %q; stderr: %s", c, o, e)
	}

	escribir("cmd/aquas/main.go", mainRoto)
	c, o, e := correr()
	if c != 0 || strings.TrimSpace(o) != "ok" {
		t.Fatalf("roto con versión anterior: exit %d, salida %q, quiero la última que compiló; stderr: %s", c, o, e)
	}
	if !strings.Contains(e, "no compila") {
		t.Fatalf("roto con versión anterior: no avisa; stderr: %q", e)
	}
}
