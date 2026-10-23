package backlog

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

// TestCerrarGuardaCodigo: el código tocado de un slice sale de sus commits y de los cambios sin commit.
// Rojo cuando: omite un archivo de producción tocado, o incluye pruebas, docs o archivos de commits ajenos.
// Arreglo: un archivo modificado sin commit perdía su primera letra.
// Verde cuando: devuelve solo los .go de producción de cmd/ e internal/ tocados por el slice, ordenados.
func TestCerrarGuardaCodigo(t *testing.T) {
	raiz := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		c := exec.Command("git", append([]string{"-C", raiz, "-c", "user.name=t", "-c", "user.email=t@t"}, args...)...)
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	poner := func(rel string) {
		p := filepath.Join(raiz, filepath.FromSlash(rel))
		os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte("package a\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	git("init", "-q")
	poner("internal/ajeno/z.go")
	git("add", "-A")
	git("commit", "-q", "-m", "E001/F001/S002: otro slice")
	poner("internal/a/a.go")
	poner("internal/a/a_test.go")
	poner("docs/x.md")
	git("add", "-A")
	git("commit", "-q", "-m", "E001/F001/S001: el slice")
	poner("cmd/c/main.go")
	if err := os.WriteFile(filepath.Join(raiz, "internal", "ajeno", "z.go"), []byte("package b"), 0o644); err != nil { // modificado sin commit
		t.Fatal(err)
	}
	got, err := codigoTocado(raiz, "E001/F001/S001", true)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"cmd/c/main.go", "internal/a/a.go", "internal/ajeno/z.go"}; !reflect.DeepEqual(got, want) {
		t.Errorf("codigoTocado = %v, se esperaba %v", got, want)
	}
}
