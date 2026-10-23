package wiki

import (
	"fmt"
	"strings"
)

// revisarMermaid: avisa cada «;» dentro de un bloque mermaid; Mermaid lo toma como fin de sentencia (D000021).
func revisarMermaid(rel, c string) []string {
	var ps []string
	dentro := false
	for i, l := range strings.Split(strings.ReplaceAll(c, "\r\n", "\n"), "\n") {
		t := strings.TrimSpace(l)
		switch {
		case !dentro && t == "```mermaid":
			dentro = true
		case dentro && strings.HasPrefix(t, "```"):
			dentro = false
		case dentro && strings.Contains(l, ";"):
			ps = append(ps, fmt.Sprintf("%s:%d: «;» dentro de un diagrama Mermaid corta la sentencia y GitHub no lo dibuja: usa «,» (D000021)", rel, i+1))
		}
	}
	return ps
}
