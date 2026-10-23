package guardia

import (
	"strings"
	"testing"
)

// TestHeredocConBarraNEnGoSeFrena: un heredoc con «\\n» que edita un .go se frena y remite a Edit (D000027).
// Rojo cuando: el comando pasa, o se frena un heredoc sin «\\n», uno que no toca .go o un comando sin heredoc.
// Verde cuando: solo la mezcla heredoc + «\\n» + .go se frena, nombrando la herramienta Edit.
func TestHeredocConBarraNEnGoSeFrena(t *testing.T) {
	raiz := t.TempDir()
	malo := "python - <<'EOF'\np='internal/x/x.go'\ns=s.replace('a','m += \"\\\\n  - \"')\nEOF"
	if m := RevisarBash(raiz, raiz, malo); !strings.Contains(m, "Edit") {
		t.Errorf("heredoc con \\\\n sobre un .go: debió frenarse nombrando Edit; dio %q", m)
	}
	for _, bueno := range []string{
		"python - <<'EOF'\np='internal/x/x.go'\ns=s.replace('a','b')\nEOF", // sin \\n
		"python - <<'EOF'\np='docs/x.md'\ns=s.replace('a','\\\\n')\nEOF",   // no toca .go
		"printf 'a\\\\nb' > .tmp/x.txt && go test ./internal/x/x.go",       // sin heredoc
	} {
		if m := RevisarBash(raiz, raiz, bueno); m != "" {
			t.Errorf("%q: no debió frenarse; dio %q", bueno, m)
		}
	}
}
