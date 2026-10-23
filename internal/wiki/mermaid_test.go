package wiki

import (
	"strings"
	"testing"
)

// TestMermaidSinPuntoYComa: un «;» dentro de un bloque mermaid hace caer el gate (D000021).
// Rojo cuando: el gate calla ante un «;» dentro de un bloque mermaid, o avisa por un «;» en la prosa.
// Verde cuando: avisa nombrando archivo y línea del «;» en el diagrama, y calla con el «;» fuera del diagrama.
func TestMermaidSinPuntoYComa(t *testing.T) {
	c := "# T\n\nprosa; con punto y coma\n\n```mermaid\nsequenceDiagram\n    A->>B: uno; dos\n```\n"
	ps := revisarMermaid("docs/x.md", c)
	if len(ps) != 1 || !strings.Contains(ps[0], "docs/x.md:7") {
		t.Errorf("se esperaba un aviso en docs/x.md:7: %v", ps)
	}
	if ps := revisarMermaid("docs/x.md", strings.ReplaceAll(c, "uno; dos", "uno, dos")); len(ps) != 0 {
		t.Errorf("sin «;» en el diagrama no debe avisar: %v", ps)
	}
}
