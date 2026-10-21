package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// E001/F005/S003: arranque en frío. Una sesión nueva, sin historial de chat, recupera el foco
// solo con lo que el hook de inicio imprime. Prueba integral de la feature F005.
//
// El binario de prueba hace de `aquas`: con AQUAS_COMO_CLI=1 corre main() en la carpeta dada.

func TestMain(m *testing.M) {
	if os.Getenv("AQUAS_COMO_CLI") == "1" {
		main()
		os.Exit(0)
	}
	os.Exit(m.Run())
}

type salida struct {
	out, err string
	codigo   int
}

func aquas(t *testing.T, raiz, stdin string, args ...string) salida {
	t.Helper()
	c := exec.Command(os.Args[0], args...)
	c.Dir = raiz
	c.Env = append(os.Environ(), "AQUAS_COMO_CLI=1")
	c.Stdin = strings.NewReader(stdin)
	var o, e bytes.Buffer
	c.Stdout, c.Stderr = &o, &e
	err := c.Run()
	s := salida{o.String(), e.String(), 0}
	if x, ok := err.(*exec.ExitError); ok {
		s.codigo = x.ExitCode()
	} else if err != nil {
		t.Fatal(err)
	}
	return s
}

func git(t *testing.T, raiz string, args ...string) {
	t.Helper()
	c := exec.Command("git", append([]string{"-C", raiz, "-c", "user.name=t", "-c", "user.email=t@t", "-c", "commit.gpgsign=false"}, args...)...)
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func escribir(t *testing.T, raiz, rel, contenido string) {
	t.Helper()
	p := filepath.Join(raiz, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(contenido), 0o644); err != nil {
		t.Fatal(err)
	}
}

// ordenDeInicio: la orden que la plantilla real de settings da al hook SessionStart de una
// sesión nueva (matcher con «startup»). Si alguien quita el hook, la prueba falla.
func ordenDeInicio(t *testing.T) []string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "gobierno", "config", "claude_settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	var p struct {
		Hooks []struct{ Evento, Matcher, Orden string }
	}
	if err := json.Unmarshal(b, &p); err != nil {
		t.Fatal(err)
	}
	for _, h := range p.Hooks {
		if h.Evento == "SessionStart" && strings.Contains("|"+h.Matcher+"|", "|startup|") {
			return strings.Fields(h.Orden)
		}
	}
	t.Fatal("la plantilla de settings no tiene hook SessionStart para una sesión nueva")
	return nil
}

func TestArranqueEnFrio(t *testing.T) {
	raiz := t.TempDir()
	git(t, raiz, "init", "-q")
	escribir(t, raiz, "a.txt", "uno\n")
	git(t, raiz, "add", "a.txt")
	git(t, raiz, "commit", "-q", "-m", "inicio")

	// Sesión 1: fija el foco, trabaja y el cierre exige actualizar.
	if s := aquas(t, raiz, "", "memoria", "fijar", "iniciativa=E001", "feature=F005", "slice=S003",
		"tarea=archivo al cerrar", "siguiente=probar arranque en frío", "pendiente=limpiar .tmp"); s.codigo != 0 {
		t.Fatalf("fijar: %+v", s)
	}
	escribir(t, raiz, "a.txt", "dos\n")
	if s := aquas(t, raiz, `{"stop_hook_active":false}`, "memoria", "cierre"); s.codigo != 2 || !strings.Contains(s.err, "memoria caliente vieja") {
		t.Fatalf("cierre con memoria vieja debe dar exit 2: %+v", s)
	}
	if s := aquas(t, raiz, "", "memoria", "fijar", "siguiente=probar arranque en frío"); s.codigo != 0 {
		t.Fatalf("fijar: %+v", s)
	}
	if s := aquas(t, raiz, `{"stop_hook_active":false}`, "memoria", "cierre"); s.codigo != 0 {
		t.Fatalf("cierre al día debe dar exit 0: %+v", s)
	}

	// Sesión 2, en frío: solo la orden del hook, sin nada de la sesión 1.
	inicio := aquas(t, raiz, `{"source":"startup"}`, ordenDeInicio(t)...)
	if inicio.codigo != 0 {
		t.Fatalf("el hook de inicio nunca bloquea: %+v", inicio)
	}
	for _, w := range []string{"foco: E001 › F005 › S003", "tarea: archivo al cerrar",
		"siguiente: probar arranque en frío", "- limpiar .tmp", "slice: docs/E001/F005/S003.md"} {
		if !strings.Contains(inicio.out, w) {
			t.Errorf("la sesión nueva debe recibir %q; recibió:\n%s", w, inicio.out)
		}
	}

	// Cierre del slice: archiva y la sesión siguiente arranca en la feature, sin slice.
	arch := aquas(t, raiz, "", "memoria", "archivar")
	if arch.codigo != 0 || !strings.Contains(arch.out, "registro/memoria/E001-F005-S003.json") {
		t.Fatalf("archivar: %+v", arch)
	}
	if _, err := os.Stat(filepath.Join(raiz, "registro", "memoria", "E001-F005-S003.json")); err != nil {
		t.Fatalf("falta el archivo: %v", err)
	}
	siguiente := aquas(t, raiz, `{"source":"startup"}`, ordenDeInicio(t)...)
	if !strings.Contains(siguiente.out, "foco: E001 › F005\n") || !strings.Contains(siguiente.out, "- limpiar .tmp") {
		t.Fatalf("tras archivar, la sesión nueva arranca en E001 › F005 con sus pendientes:\n%s", siguiente.out)
	}
}
