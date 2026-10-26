package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// E000/F005/S004: el CLI cableado de punta a punta, sobre el binario y la config reales.

// Inicio mínimo que pasa el gate de la wiki: secciones fijas, diagrama y ejemplo con ancla.
const inicioMinimo = `# Inicio

## En una frase

Un repo de prueba.

## Para qué existe

Para ver el gate.

## Analogía

Como una caja.

## Diagrama

` + "```mermaid\nflowchart LR\n    A --> B\n```" + `

## Ejemplo de punta a punta

<a id="EJ-INICIO"></a>

Se corre el gate.

## Paso a paso

1. Uno.

## Cómo se usa

Se lee.

## Navegación

- Glosario: [palabras](glosario.md)
`

// repoSano: repo Git propio con la config real copiada, entorno declarado, compilado y una wiki mínima.
func repoSano(t *testing.T) string {
	t.Helper()
	raiz := t.TempDir()
	if out, err := exec.Command("git", "-C", raiz, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	copiarGobierno(t, raiz)
	if err := os.MkdirAll(filepath.Join(raiz, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(raiz, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(raiz, "docs", "README.md"), []byte(inicioMinimo), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(raiz, "docs", "glosario.md"), []byte("# Glosario\n\nPalabras técnicas.\n\n## zeta\n\nUna letra.\n\nEjemplo: la zeta va al final.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, orden := range [][]string{{"iniciar"}, {"compilar"}} {
		if s := aquas(t, raiz, "", orden...); s.codigo != 0 {
			t.Fatalf("%v: %+v", orden, s)
		}
	}
	return raiz
}

func TestVerificarDePuntaAPunta(t *testing.T) {
	raiz := repoSano(t)
	if s := aquas(t, raiz, "", "verificar"); s.codigo != 0 || !strings.Contains(s.out, "GATE OK") {
		t.Fatalf("repo sano: exit 0 y GATE OK; dio %+v", s)
	}
	claude := filepath.Join(raiz, "CLAUDE.md")
	b, _ := os.ReadFile(claude)
	if err := os.WriteFile(claude, append(b, []byte("\nregla metida a mano\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	s := aquas(t, raiz, "", "verificar")
	if s.codigo != 1 || !strings.Contains(s.err, "CLAUDE.md: desfasado o editado a mano") {
		t.Fatalf("derivado editado: exit 1 y lo nombra; dio %+v", s)
	}
	if s := aquas(t, raiz, "", "compilar"); s.codigo != 0 {
		t.Fatalf("compilar: %+v", s)
	}
	if s := aquas(t, raiz, "", "verificar"); s.codigo != 0 {
		t.Fatalf("tras compilar vuelve a exit 0; dio %+v", s)
	}
	inicio := filepath.Join(raiz, "docs", "README.md")
	if err := os.WriteFile(inicio, []byte(strings.Replace(inicioMinimo, "## Paso a paso", "## Otra cosa", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	if s := aquas(t, raiz, "", "verificar"); s.codigo != 1 || !strings.Contains(s.err, "docs/README.md: falta sección «Paso a paso»") {
		t.Fatalf("wiki rota: exit 1 y nombra la sección; dio %+v", s)
	}
}

func TestGuardiaCLIBloquea(t *testing.T) {
	raiz := repoSano(t)
	s := aquas(t, raiz, `{"tool_name":"Write","tool_input":{"file_path":"../fuera.txt"}}`, "guardia")
	if s.codigo != 2 || !strings.Contains(s.err, "bloqueada: ../fuera.txt") {
		t.Fatalf("escritura fuera: exit 2 con la ruta en el motivo; dio %+v", s)
	}
	s = aquas(t, raiz, `{"tool_name":"Bash","tool_input":{"command":"echo x > ../fuera.txt"}}`, "guardia")
	if s.codigo != 2 || !strings.Contains(s.err, "bloqueada: ../fuera.txt") {
		t.Fatalf("redirección fuera por Bash: exit 2; dio %+v", s)
	}
	if s := aquas(t, raiz, `{"tool_name":"Write","tool_input":{"file_path":"docs/a.md"}}`, "guardia"); s.codigo != 0 {
		t.Fatalf("escritura dentro: exit 0; dio %+v", s)
	}
}

func TestGuardiaCLISinEntornoBloquea(t *testing.T) {
	dentro := `{"tool_name":"Write","tool_input":{"file_path":"docs/a.md"}}`
	raiz := repoSano(t)
	if err := os.Remove(filepath.Join(raiz, "gobierno", "config", "entorno.local.json")); err != nil {
		t.Fatal(err)
	}
	s := aquas(t, raiz, dentro, "guardia")
	if s.codigo != 2 || !strings.Contains(s.err, "entorno no verificado, se bloquea todo") || !strings.Contains(s.err, "aquas iniciar") {
		t.Fatalf("sin entorno: hasta lo de dentro se bloquea (exit 2) y dice cómo arreglarlo; dio %+v", s)
	}
	raiz = repoSano(t)
	p := filepath.Join(raiz, "gobierno", "config", "entorno.local.json")
	b, _ := os.ReadFile(p)
	if err := os.WriteFile(p, []byte(strings.Replace(string(b), filepath.ToSlash(raiz), "C:/otra/carpeta", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	if s := aquas(t, raiz, dentro, "guardia"); s.codigo != 2 || !strings.Contains(s.err, "entorno no verificado") {
		t.Fatalf("entorno de otra carpeta: exit 2; dio %+v", s)
	}
}
