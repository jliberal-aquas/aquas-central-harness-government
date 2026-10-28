package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestHookFrenaShellAjeno (E001/F016, integral): el hook de cierre frena un bloque de otro shell según el entorno.
// Rojo cuando: un bloque bash pasa con shell_humano powershell.
// Verde cuando: frena el bloque ajeno y deja pasar el del shell declarado.
func TestHookFrenaShellAjeno(t *testing.T) {
	raiz := t.TempDir()
	escribir(t, raiz, "gobierno/config/entorno.local.json", `{"version": 1, "so": "windows", "raiz": "x", "shell_humano": "powershell"}`)
	cierre := func(msg string) salida {
		t.Helper()
		b, _ := json.Marshal(map[string]string{"hook_event_name": "Stop", "last_assistant_message": msg})
		return aquas(t, raiz, string(b), "propuesta")
	}
	if s := cierre("Corre esto:\n\n```bash\nls -la\n```\n"); s.codigo != 2 || !strings.Contains(s.err, "powershell") {
		t.Errorf("bloque bash con shell_humano powershell: frena y nombra el shell; dio %+v", s)
	}
	if s := cierre("Corre esto:\n\n```powershell\nGet-ChildItem\n```\n"); s.codigo != 0 {
		t.Errorf("bloque del shell declarado: pasa; dio %+v", s)
	}
	escribir(t, raiz, "gobierno/config/entorno.local.json", `{"version": 2, "so": "windows", "raiz": "x", "shell_humano": "powershell"}`)
	if s := cierre("Corre esto:\n\n```bash\nls -la\n```\n"); s.codigo != 0 {
		t.Errorf("entorno ilegible: el cierre no frena (la guardia ya lo frena); dio %+v", s)
	}
	escribir(t, raiz, "gobierno/config/entorno.local.json", `{"version": 1, "so": "windows", "raiz": "x", "shell_humano": "pwsh"}`)
	if s := cierre("Listo."); s.codigo != 2 || !strings.Contains(s.err, "shell_humano") {
		t.Errorf("shell_humano inválido: frena avisando, no apaga la revisión en silencio; dio %+v", s)
	}
}
