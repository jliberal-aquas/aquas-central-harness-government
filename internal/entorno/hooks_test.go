package entorno

import (
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
