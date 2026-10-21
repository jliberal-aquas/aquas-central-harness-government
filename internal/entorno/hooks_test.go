package entorno

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// E001/F005/S002: settings.json trae los hooks de la memoria caliente.

const plantillaConEventos = `{
  "agent": "orchestrator",
  "hooks": [
    {"evento": "SessionStart", "matcher": "startup|resume|clear|compact", "orden": "memoria inicio"},
    {"evento": "Stop", "orden": "memoria cierre"}
  ],
  "por_so": {"windows": {"shell": "bash", "herramientas_shell": ["Bash"]}}
}`

func TestSettingsEventoConMatcherPropio(t *testing.T) {
	b, err := generarSettings([]byte(plantillaConEventos), "windows")
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for _, quiero := range []string{`"SessionStart"`, `"matcher": "startup|resume|clear|compact"`,
		`aquas.sh\" memoria inicio`, `"Stop"`, `aquas.sh\" memoria cierre`} {
		if !strings.Contains(s, quiero) {
			t.Errorf("falta %s en:\n%s", quiero, s)
		}
	}
	if strings.Contains(s, `"matcher": ""`) {
		t.Errorf("Stop no lleva matcher vacío:\n%s", s)
	}
}

func TestSettingsRealTraeHooksDeMemoria(t *testing.T) {
	fuente, err := os.ReadFile(filepath.Join("..", "..", filepath.FromSlash(RutaPlantilla)))
	if err != nil {
		t.Fatal(err)
	}
	b, err := generarSettings(fuente, "windows")
	if err != nil {
		t.Fatal(err)
	}
	for _, quiero := range []string{`"matcher": "startup|resume|clear|compact"`, `memoria inicio`, `memoria cierre`} {
		if !strings.Contains(string(b), quiero) {
			t.Errorf("la plantilla real debe declarar %s", quiero)
		}
	}
}

// E000/F005/S004: la plantilla real trae la guardia antes de escribir y el recompilado después.
// Se lee el JSON generado: cada hook en su evento, con su matcher exacto y su orden.
func TestSettingsRealTraeGuardiaYHook(t *testing.T) {
	fuente, err := os.ReadFile(filepath.Join("..", "..", filepath.FromSlash(RutaPlantilla)))
	if err != nil {
		t.Fatal(err)
	}
	type hook struct {
		Matcher string
		Hooks   []struct{ Command string }
	}
	for _, c := range []struct{ so, preMatcher string }{
		{"windows", "Write|Edit|MultiEdit|NotebookEdit|Bash|PowerShell"},
		{"linux", "Write|Edit|MultiEdit|NotebookEdit|Bash"},
	} {
		b, err := generarSettings(fuente, c.so)
		if err != nil {
			t.Fatal(err)
		}
		var s struct{ Hooks map[string][]hook }
		if err := json.Unmarshal(b, &s); err != nil {
			t.Fatal(err)
		}
		tiene := func(evento, matcher, sufijo string) {
			t.Helper()
			for _, h := range s.Hooks[evento] {
				if h.Matcher == matcher && len(h.Hooks) == 1 && strings.HasSuffix(h.Hooks[0].Command, `aquas.sh" `+sufijo) {
					return
				}
			}
			t.Errorf("%s: falta %s con matcher %q y orden %q; hay %+v", c.so, evento, matcher, sufijo, s.Hooks[evento])
		}
		tiene("PreToolUse", c.preMatcher, "guardia")
		tiene("PostToolUse", "Edit|Write", "hook")
	}
}

// E001/F006/S002: el cierre de worker y auditor lo revisa un hook SubagentStop.
func TestSettingsRealTraeHookDeCierreDeAgente(t *testing.T) {
	fuente, err := os.ReadFile(filepath.Join("..", "..", filepath.FromSlash(RutaPlantilla)))
	if err != nil {
		t.Fatal(err)
	}
	b, err := generarSettings(fuente, "windows")
	if err != nil {
		t.Fatal(err)
	}
	for _, quiero := range []string{`"SubagentStop"`, `"matcher": "worker-sonnet|worker-haiku|auditor"`, `cierre-agente`} {
		if !strings.Contains(string(b), quiero) {
			t.Errorf("la plantilla real debe declarar %s", quiero)
		}
	}
}
