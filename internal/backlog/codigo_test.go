package backlog

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
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

// TestCodigoIDEnTodoElMensaje: el ID cuenta en cualquier parte del mensaje y se pueden sumar commits que no lo nombran (D000022).
// Rojo cuando: ignora un commit con el ID al final, toma un ID más largo que empieza igual, o no suma un commit explícito.
// Verde cuando: toma el commit con el ID al final y el explícito con su configuración (sin sellos), y descarta el ID más largo.
func TestCodigoIDEnTodoElMensaje(t *testing.T) {
	raiz := t.TempDir()
	git := func(args ...string) string {
		t.Helper()
		out, err := exec.Command("git", append([]string{"-C", raiz, "-c", "user.name=t", "-c", "user.email=t@t"}, args...)...).CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	poner := func(rel string) {
		p := filepath.Join(raiz, filepath.FromSlash(rel))
		os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte("package a"), 0o644); err != nil {
			t.Fatal(err)
		}
		git("add", "-A")
	}
	git("init", "-q")
	poner("internal/a/a.go")
	git("commit", "-q", "-m", "Guardia: algo (E001/F001/S001)")
	poner("internal/largo/l.go")
	git("commit", "-q", "-m", "E001/F001/S0010: otro")
	poner("internal/b/b.go")
	poner(".claude/settings.json")
	poner(".claude/hooks/x.sh")
	poner("gobierno/config/sellos.json")
	git("commit", "-q", "-m", "sin ID")
	hash := git("rev-parse", "HEAD")
	f := Ficha{ID: "E001/F001/S001", Tipo: "slice", Titulo: "T", Estado: "CERRADO", Prioridad: "P1", Motivo: "m", Padre: "E001/F001", BloqueadoPor: []string{}, Secuencia: 1}
	pisar(t, raiz, Archivo, f)
	g, err := Codigo(raiz, "E001/F001/S001", []string{hash})
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{".claude/hooks/x.sh", ".claude/settings.json", "internal/a/a.go", "internal/b/b.go"}; !reflect.DeepEqual(g.Codigo, want) {
		t.Errorf("Codigo = %v, se esperaba %v", g.Codigo, want)
	}
}
