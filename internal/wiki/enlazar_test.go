package wiki

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnlazaGlosarioSolo(t *testing.T) {
	raiz := t.TempDir()
	escribirArchivo := func(rel, s string) {
		p := filepath.Join(raiz, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	escribirArchivo(glosario, "# Glosario\n\n## agente\n\nQuien trabaja.\n\nEjemplo: el worker.\n\n## wiki\n\nDocs.\n\nEjemplo: docs/.\n")
	const rel = "docs/E000/F001/S001.md"
	antes := "# S001 · agente\n\n## Diagrama\n\n```mermaid\nflowchart LR\n  A[\"agente\"]\n```\n\n| agente | wiki |\n|---|---|\n\n" +
		"Usa `agente` en código, pero el agente y la wiki van enlazados. Otro agente no.\n"
	despues := "# S001 · agente\n\n## Diagrama\n\n```mermaid\nflowchart LR\n  A[\"agente\"]\n```\n\n| agente | wiki |\n|---|---|\n\n" +
		"Usa `agente` en código, pero el [agente](../../glosario.md#agente) y la [wiki](../../glosario.md#wiki) van enlazados. Otro agente no.\n"
	escribirArchivo(rel, antes)

	n, err := Enlazar(raiz, rel)
	if err != nil || n != 2 {
		t.Fatalf("primera corrida: %d enlaces, %v; esperaba 2", n, err)
	}
	b, _ := os.ReadFile(filepath.Join(raiz, filepath.FromSlash(rel)))
	if string(b) != despues {
		t.Fatalf("página\n got %q\nwant %q", b, despues)
	}
	if n, err := Enlazar(raiz, rel); err != nil || n != 0 {
		t.Fatalf("segunda corrida: %d enlaces, %v; esperaba 0 (idempotente)", n, err)
	}
}
