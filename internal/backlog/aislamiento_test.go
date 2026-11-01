package backlog

import (
	"os/exec"
	"strings"
	"testing"
)

// TestGitNoSubeALaCentral (E000/F009/S001, D000050): git en una carpeta temporal de prueba no ve el repo que la
// contiene, y codigoTocado acepta un repo propio sin commits.
// Rojo cuando: git rev-parse en la carpeta temporal da la central.
// Verde cuando: da «not a git repository» y codigoTocado acepta repo sin commits.
func TestGitNoSubeALaCentral(t *testing.T) {
	raiz := t.TempDir()
	if out, err := exec.Command("git", "-C", raiz, "rev-parse", "--show-toplevel").CombinedOutput(); err == nil {
		t.Errorf("carpeta temporal sin git init: git no debe hallar un repo; halló %s", strings.TrimSpace(string(out)))
	}
	if xs, err := codigoTocado(raiz, "E001/F001/S001", false); err == nil {
		t.Errorf("fuera de un repo codigoTocado da error, no una lista; dio %v", xs)
	}
	if out, err := exec.Command("git", "-C", raiz, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	xs, err := codigoTocado(raiz, "E001/F001/S001", true)
	if err != nil || len(xs) != 0 {
		t.Errorf("repo propio sin commits: codigoTocado da lista vacía sin error; dio %v, %v", xs, err)
	}
}
