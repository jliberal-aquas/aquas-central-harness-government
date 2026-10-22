package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// E000/F004/S004: el hook PostToolUse enlaza el glosario al guardar una página de docs/.
func TestHookEnlazaAlGuardar(t *testing.T) {
	raiz := t.TempDir()
	escribir(t, raiz, "docs/glosario.md", "# Glosario\n\n## agente\n\nQuien trabaja.\n\nEjemplo: el worker.\n")
	const pagina = "Un agente trabaja.\n"
	escribir(t, raiz, "docs/E000/F001/S001.md", pagina)
	escribir(t, raiz, "notas.md", pagina)
	evento := func(rel string) string {
		b, _ := json.Marshal(map[string]any{"tool_input": map[string]string{"file_path": filepath.Join(raiz, filepath.FromSlash(rel))}})
		return string(b)
	}
	leer := func(rel string) string {
		b, err := os.ReadFile(filepath.Join(raiz, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}

	s := aquas(t, raiz, evento("docs/E000/F001/S001.md"), "hook")
	if got, want := leer("docs/E000/F001/S001.md"), "Un [agente](../../glosario.md#agente) trabaja.\n"; got != want {
		t.Fatalf("página de docs/ sin enlazar\n got %q\nwant %q\nstderr %s", got, want, s.err)
	}
	if !strings.Contains(s.err, "1 enlaces al glosario") {
		t.Errorf("el hook debe avisar lo que cambió; stderr %q", s.err)
	}

	aquas(t, raiz, evento("notas.md"), "hook")
	if got := leer("notas.md"); got != pagina {
		t.Fatalf("fuera de docs/ no debe tocar nada: %q", got)
	}
}
