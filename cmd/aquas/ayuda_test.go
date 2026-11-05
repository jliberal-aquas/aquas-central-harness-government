package main

import (
	"os/exec"
	"strings"
	"testing"
)

// TestAyudaNoEjecuta (E000/F015/S002): --help y -h muestran el uso y no ejecutan la orden.
// Rojo cuando: una orden con --help escribe archivos.
// Verde cuando: solo muestra el uso y el repo queda igual.
func TestAyudaNoEjecuta(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init", "-q")
	if s := aquas(t, dir, "", "install"); s.codigo != 0 {
		t.Fatalf("install: %+v", s)
	}
	git(t, dir, "add", "-A")
	git(t, dir, "-c", "user.email=a@b", "-c", "user.name=a", "-c", "core.hooksPath=/dev/null", "commit", "-q", "-m", "base")
	estado := func() string {
		out, err := exec.Command("git", "-C", dir, "status", "--porcelain", "--untracked-files=all").Output()
		if err != nil {
			t.Fatal(err)
		}
		return string(out)
	}
	antes := estado()
	for _, args := range [][]string{
		{"compile", "--help"},
		{"backlog", "new", "-h"},
		{"backlog", "new", "id=E001", "--help"},
		{"memory", "set", "-h", "-h"},
	} {
		s := aquas(t, dir, "", args...)
		if s.codigo != 0 || !strings.Contains(s.out, "uso: aquas") || !strings.Contains(s.out, "no se ejecutó") {
			t.Errorf("%v: debió mostrar el uso y salir 0; dio %+v", args, s)
		}
		if d := estado(); d != antes {
			t.Errorf("%v: la ayuda cambió el repo:\n%s", args, d)
		}
	}
	// Un valor que contiene help no es pedir ayuda.
	if s := aquas(t, dir, "", "backlog", "new", "motivo=--help"); strings.Contains(s.out, "ayuda") {
		t.Errorf("motivo=--help no es pedir ayuda; dio %+v", s)
	}
}
