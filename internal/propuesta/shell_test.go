package propuesta

import (
	"strings"
	"testing"
)

// TestShellDelHumano (E001/F016/S001): la revisión frena bloques marcados de otro shell y deja pasar los demás.
// Rojo cuando: un bloque ajeno pasa o uno propio, sin marca o sin shell declarado frena.
// Verde cuando: solo frena el bloque marcado de otro shell.
func TestShellDelHumano(t *testing.T) {
	bloque := func(lang, cmd string) string { return "Texto.\n\n```" + lang + "\n" + cmd + "\n```\n\nFin." }
	frena := []struct{ shell, msg string }{
		{"powershell", bloque("bash", "ls")},
		{"powershell", bloque("sh", "ls")},
		{"powershell", bloque("BASH", "ls")},
		{"bash", bloque("powershell", "Get-ChildItem")},
		{"powershell", bloque("powershell", "dir") + "\n" + bloque("bash", "ls")},
		{"powershell", "~~~bash\nls\n~~~\n"},
		{"powershell", "````bash\nls\n````\n"},
		{"powershell", "```text\n```bash dentro de otro bloque\n```\n\n```bash\nls\n```\n"},
	}
	if m := RevisarShell("Listo.", "pwsh"); !strings.Contains(m, "shell_humano") {
		t.Errorf("shell_humano inválido: frena avisando; dio %q", m)
	}
	for _, c := range frena {
		if m := RevisarShell(c.msg, c.shell); m == "" || !strings.Contains(m, c.shell) {
			t.Errorf("shell %s, mensaje %q: debe frenar nombrando el shell; dio %q", c.shell, c.msg, m)
		}
	}
	pasa := []struct{ shell, msg string }{
		{"powershell", bloque("powershell", "dir")},
		{"bash", bloque("bash", "ls")},
		{"bash", bloque("sh", "ls")},
		{"powershell", bloque("", "ls")},
		{"powershell", bloque("go", "fmt.Println()")},
		{"", bloque("bash", "ls")},
		{"powershell", "Corre `ls -la` en la terminal."},
		{"powershell", "```text\n```bash dentro de otro bloque\n```\n"},
		{"powershell", "~~~md\n```bash\nls\n```\n~~~\n"},
		{"powershell", "~~~md\n```\n```bash\nls\n```\n~~~\n"},
		{"powershell", "````md\n```\n```bash\nls\n```\n````\n"},
		{"powershell", "Texto con ```bash en medio de la línea."},
	}
	for _, c := range pasa {
		if m := RevisarShell(c.msg, c.shell); m != "" {
			t.Errorf("shell %q, mensaje %q: debe pasar; dio %q", c.shell, c.msg, m)
		}
	}
}
