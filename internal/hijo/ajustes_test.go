package hijo

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// comandosPostToolUse devuelve los hooks PostToolUse del settings del hijo como «matcher → type shell command».
func comandosPostToolUse(t *testing.T, dir string) (map[string]any, []string) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	if err != nil {
		t.Fatalf("el hijo debe tener .claude/settings.json: %v", err)
	}
	var s struct {
		Hooks struct {
			PostToolUse []struct {
				Matcher string `json:"matcher"`
				Hooks   []struct {
					Type    string `json:"type"`
					Shell   string `json:"shell"`
					Command string `json:"command"`
				} `json:"hooks"`
			} `json:"PostToolUse"`
		} `json:"hooks"`
	}
	var todo map[string]any
	if err := json.Unmarshal(b, &s); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &todo); err != nil {
		t.Fatal(err)
	}
	var cs []string
	for _, g := range s.Hooks.PostToolUse {
		for _, h := range g.Hooks {
			cs = append(cs, g.Matcher+" → "+h.Type+" "+h.Shell+" "+h.Command)
		}
	}
	return todo, cs
}

func contar(xs []string, x string) int {
	n := 0
	for _, s := range xs {
		if s == x {
			n++
		}
	}
	return n
}

// TestHijoRecompilaAlEditarHueco (E002/F002/S006): install y update dejan un hook PostToolUse con aquas build.
// Rojo cuando: el hijo queda sin ese hook o pierde su settings previo.
// Verde cuando: el hook existe una vez, lo previo se conserva y tras editar un hueco el CLAUDE.md queda al día.
func TestHijoRecompilaAlEditarHueco(t *testing.T) {
	const build = "Edit|Write → command bash " + ComandoBuild
	// Recién instalado: el hook existe, de tipo command y en bash.
	dir := instalarConHuecos(t, map[string]string{"mision": "- Landing.\n"})
	if err := Listo(dir); err != nil {
		t.Fatal(err)
	}
	if _, cs := comandosPostToolUse(t, dir); contar(cs, build) != 1 {
		t.Errorf("recién instalado: se esperaba el hook %q una vez; hay %q", build, cs)
	}
	// Settings previo: se conserva lo suyo, sin escapar «&&» ni «<» ni redondear enteros, y sin duplicar.
	previo := `{"n":12345678901234567890,"permissions":{"allow":["Bash(ls)"]},"hooks":{"PostToolUse":[{"matcher":"Bash","hooks":[{"type":"command","command":"a && b <c>"}]}]}}`
	p := filepath.Join(dir, ".claude", "settings.json")
	if err := os.WriteFile(p, []byte(previo), 0o644); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := Listo(dir); err != nil {
			t.Fatal(err)
		}
	}
	todo, cs := comandosPostToolUse(t, dir)
	if contar(cs, build) != 1 || contar(cs, "Bash → command  a && b <c>") != 1 {
		t.Errorf("con settings previo: se esperaba el propio y el de build una vez cada uno; hay %q", cs)
	}
	if b, _ := json.Marshal(todo["permissions"]); string(b) != `{"allow":["Bash(ls)"]}` {
		t.Errorf("permissions previos perdidos: %s", b)
	}
	if b, _ := os.ReadFile(p); !strings.Contains(string(b), `"a && b <c>"`) || !strings.Contains(string(b), "12345678901234567890") {
		t.Errorf("la reescritura alteró texto o números previos:\n%s", b)
	}
	// Forma inesperada: falla sin tocar el archivo.
	for _, malo := range []string{"{roto", "null", `{"hooks":[]}`, `{"hooks":{"PostToolUse":{}}}`} {
		if err := os.WriteFile(p, []byte(malo), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := Listo(dir); err == nil {
			t.Errorf("settings %s: Listo debió fallar", malo)
		}
		if b, _ := os.ReadFile(p); string(b) != malo {
			t.Errorf("settings %s no debe tocarse; quedó %q", malo, b)
		}
	}
	// Tras editar un hueco, el build que corre el hook deja el CLAUDE.md al día.
	if err := os.WriteFile(filepath.Join(dir, CarpetaHuecos, "mision.md"), []byte("- Otra misión.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if m := Verificar(dir); len(m) == 0 {
		t.Fatal("tras editar el hueco y antes del build, verificar debió fallar")
	}
	if err := Build(dir); err != nil {
		t.Fatal(err)
	}
	if m := Verificar(dir); len(m) != 0 {
		t.Errorf("tras el build del hook, verificar debió pasar; dio %v", m)
	}
}
