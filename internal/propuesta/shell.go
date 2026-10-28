package propuesta

import (
	"fmt"
	"strings"
)

// ajenos: lenguajes de bloque que no corren en el shell declarado del humano (E001/F016/S001).
var ajenos = map[string][]string{
	"powershell": {"bash", "sh"},
	"bash":       {"powershell"},
}

// RevisarShell (E001/F016/S001): "" si ningún bloque de código marcado es de un shell ajeno al declarado;
// si no, el motivo para frenar el cierre. Sin shell declarado no revisa. No mira comandos sueltos ni bloques sin marca.
// Lee las cercas igual que Revisar (cercaDe): ``` o ~~~, y solo cierra una cerca del mismo carácter, igual o más larga, sola.
func RevisarShell(msg, shell string) string {
	if shell == "" {
		return ""
	}
	prohibidos, ok := ajenos[shell]
	if !ok { // un error de escritura no apaga la revisión en silencio; el agente puede corregir el archivo
		return fmt.Sprintf("shell_humano %q inválido en gobierno/config/entorno.local.json: usa powershell o bash.", shell)
	}
	var abierta byte
	largo := 0
	for _, l := range strings.Split(strings.ReplaceAll(msg, "\r\n", "\n"), "\n") {
		t := strings.TrimSpace(l)
		c, n := cercaDe(t)
		if abierta != 0 {
			if c == abierta && n >= largo && strings.Trim(t, string(c)) == "" {
				abierta = 0
			}
			continue
		}
		if n == 0 {
			continue
		}
		abierta, largo = c, n
		lang, _, _ := strings.Cut(strings.ToLower(strings.TrimSpace(t[n:])), " ")
		for _, p := range prohibidos {
			if lang == p {
				return fmt.Sprintf("Bloque de código %s para el humano, pero su shell declarado es %s (gobierno/config/entorno.local.json): reescríbelo en %s.", lang, shell, shell)
			}
		}
	}
	return ""
}
