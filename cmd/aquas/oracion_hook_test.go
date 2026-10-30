package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

// TestHookAvisaOracionLarga (E000/F006/S004): al guardar una página de docs/, el hook avisa la línea
// de cada oración de más de 20 palabras, sin esperar al gate.
// Rojo cuando: se guarda docs/ con una oración larga y el hook no avisa su línea.
// Verde cuando: avisa la línea donde empieza la oración larga y calla si todas son cortas.
func TestHookAvisaOracionLarga(t *testing.T) {
	raiz := t.TempDir()
	evento := func(rel string) string {
		b, _ := json.Marshal(map[string]any{"tool_input": map[string]string{"file_path": filepath.Join(raiz, filepath.FromSlash(rel))}})
		return string(b)
	}
	larga := strings.TrimSpace(strings.Repeat("sol ", 21))
	pagina := "# Título\n\n" + // 1-2
		"Una corta.\n\n" + // 3-4
		"```\n" + larga + "\n```\n\n" + // 5-8: código, no es prosa
		"| " + larga + " |\n\n" + // 9-10: tabla, no es prosa
		"Corta primero. " + strings.Repeat("luna ", 12) + "\n" + // 11: la larga empieza aquí
		strings.Repeat("mar ", 10) + "fin.\n\n" + // 12: y sigue aquí
		"- " + larga + ".\n" // 14: ítem de lista largo
	escribir(t, raiz, "docs/E000/F001/S001.md", pagina)
	s := aquas(t, raiz, evento("docs/E000/F001/S001.md"), "hook")
	if s.codigo != 2 {
		t.Fatalf("oración larga: el hook debe salir 2 para que Claude lo vea; exit %d, stderr %q", s.codigo, s.err)
	}
	for _, want := range []string{"docs/E000/F001/S001.md:11: oración de 23 palabras", "docs/E000/F001/S001.md:14: oración de 21 palabras"} {
		if !strings.Contains(s.err, want) {
			t.Errorf("falta %q en el aviso:\n%s", want, s.err)
		}
	}
	if n := strings.Count(s.err, "oración de"); n != 2 {
		t.Errorf("código y tablas no son prosa: esperaba 2 avisos, hubo %d:\n%s", n, s.err)
	}

	// Hallazgos A1.1 y A1.4: la larga no es la primera del párrafo y empieza en otra línea;
	// en el mismo guardado el hook también enlaza el glosario; salen los dos avisos.
	escribir(t, raiz, "docs/glosario.md", "# Glosario\n\n## agente\n\nQuien trabaja.\n\nEjemplo: el worker.\n")
	escribir(t, raiz, "docs/E000/F001/S003.md", "# Título\n\nUn agente corto.\n"+larga+".\n")
	s = aquas(t, raiz, evento("docs/E000/F001/S003.md"), "hook")
	if s.codigo != 2 || !strings.Contains(s.err, "S003.md:4: oración de 21 palabras") || !strings.Contains(s.err, "1 enlaces al glosario") {
		t.Errorf("enlace y oración larga en otra línea del párrafo: exit %d, stderr %q", s.codigo, s.err)
	}
	// Hallazgo A1.3: el glosario también se mide al guardarlo.
	escribir(t, raiz, "docs/glosario.md", "# Glosario\n\n## agente\n\n"+larga+".\n")
	if s := aquas(t, raiz, evento("docs/glosario.md"), "hook"); s.codigo != 2 || !strings.Contains(s.err, "glosario.md:5: oración de 21 palabras") {
		t.Errorf("glosario con oración larga: exit %d, stderr %q", s.codigo, s.err)
	}

	// Todas cortas: el hook calla.
	escribir(t, raiz, "docs/E000/F001/S002.md", "# Título\n\nUna corta.\n\nOtra corta también.\n")
	if s := aquas(t, raiz, evento("docs/E000/F001/S002.md"), "hook"); s.codigo != 0 || strings.Contains(s.err, "oración") {
		t.Errorf("todas cortas: el hook no debe avisar; exit %d, stderr %q", s.codigo, s.err)
	}
	// Fuera de docs/: no se mide.
	escribir(t, raiz, "notas.md", larga+".\n")
	if s := aquas(t, raiz, evento("notas.md"), "hook"); strings.Contains(s.err, "oración") {
		t.Errorf("fuera de docs/ no se mide: %q", s.err)
	}
}
