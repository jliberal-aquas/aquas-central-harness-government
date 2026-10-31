package entorno

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func escribirEntorno(t *testing.T, raiz, contenido string) {
	t.Helper()
	dir := filepath.Join(raiz, "gobierno", "config")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "entorno.local.json"), []byte(contenido), 0o644); err != nil {
		t.Fatal(err)
	}
}

func entornoJSON(so, raiz string) string {
	return `{"version":1,"so":"` + so + `","raiz":"` + filepath.ToSlash(raiz) + `"}`
}

func TestValidarSinArchivoBloquea(t *testing.T) {
	if m := Validar(t.TempDir()); !strings.Contains(m, "aquas init") {
		t.Fatalf("sin entorno.local.json debe pedir `aquas init`; obtuve %q", m)
	}
}

func TestValidarSODistintoBloquea(t *testing.T) {
	raiz := t.TempDir()
	escribirEntorno(t, raiz, entornoJSON("plan9", raiz))
	if m := Validar(raiz); !strings.Contains(m, "sistema operativo") {
		t.Fatalf("SO declarado ≠ real debe bloquear; obtuve %q", m)
	}
}

func TestValidarRaizDistintaBloquea(t *testing.T) {
	raiz := t.TempDir()
	escribirEntorno(t, raiz, entornoJSON(runtime.GOOS, filepath.Join(raiz, "otro")))
	if m := Validar(raiz); !strings.Contains(m, "raíz") {
		t.Fatalf("raíz declarada ≠ real debe bloquear; obtuve %q", m)
	}
}

func TestValidarArchivoRotoBloquea(t *testing.T) {
	raiz := t.TempDir()
	escribirEntorno(t, raiz, `{roto`)
	// E000/F005/S002: bloquea porque el archivo está roto, y dice cómo regenerarlo.
	if m := Validar(raiz); !strings.Contains(m, RutaEntorno+" ilegible") || !strings.Contains(m, "aquas init") {
		t.Fatalf("entorno ilegible debe bloquear por ilegible y pedir `aquas init`; obtuve %q", m)
	}
}

func TestValidarCorrectoPasa(t *testing.T) {
	raiz := t.TempDir()
	escribirEntorno(t, raiz, entornoJSON(runtime.GOOS, raiz))
	if m := Validar(raiz); m != "" {
		t.Fatalf("entorno correcto debe pasar; obtuve %q", m)
	}
}

const plantillaPrueba = `{
  "agent": "orchestrator",
  "hooks": [
    {"evento": "PreToolUse", "herramientas": ["Write", "Edit"], "con_shell": true, "orden": "guardia"},
    {"evento": "PostToolUse", "herramientas": ["Edit", "Write"], "orden": "hook"}
  ],
  "por_so": {
    "windows": {"shell": "bash", "herramientas_shell": ["Bash", "PowerShell"]},
    "linux": {"shell": "bash", "herramientas_shell": ["Bash"]}
  }
}`

func TestSettingsWindowsIncluyePowerShell(t *testing.T) {
	b, err := generarSettings([]byte(plantillaPrueba), "windows")
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for _, quiero := range []string{`"matcher": "Write|Edit|Bash|PowerShell"`, `"shell": "bash"`, `"agent": "orchestrator"`, `aquas.sh\" guardia`} {
		if !strings.Contains(s, quiero) {
			t.Errorf("falta %s en:\n%s", quiero, s)
		}
	}
}

func TestSettingsLinuxSinPowerShell(t *testing.T) {
	b, err := generarSettings([]byte(plantillaPrueba), "linux")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "PowerShell") {
		t.Fatalf("Linux no debe vigilar PowerShell:\n%s", b)
	}
}

func TestSettingsSOSinConfigFalla(t *testing.T) {
	// E000/F005/S002: el error nombra el SO que no tiene configuración.
	_, err := generarSettings([]byte(plantillaPrueba), "plan9")
	if err == nil || !strings.Contains(err.Error(), `sin shell declarado para el SO "plan9"`) {
		t.Fatalf("SO sin entrada en por_so debe fallar nombrando el SO, no inventar shell; dio %v", err)
	}
}
