package main

// E001/F010/S001: hook UserPromptSubmit que registra «apruebo <id>» del mensaje del humano.

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/backlog"
)

// reApruebo: la línea empieza (tras espacios) por «apruebo <id>». «no apruebo …» no cuenta.
var reApruebo = regexp.MustCompile(`(?i)^\s*apruebo\s+(E\d{3}(?:/F\d{3}(?:/S\d{3})?)?)\b`)

// marcasDeAgente: un prompt con alguna de estas marcas es un mensaje de agente, no del humano.
var marcasDeAgente = []string{"<agent-message", "[Subagent hand-back]", "<task-notification>"}

// idAprobado: id de la PRIMERA línea no vacía si es «apruebo <id>»; una línea que termina en «?» es pregunta.
// Todo prompt con marcas de mensaje de agente se descarta.
func idAprobado(prompt string) (string, bool) {
	for _, m := range marcasDeAgente {
		if strings.Contains(prompt, m) {
			return "", false
		}
	}
	for _, l := range strings.Split(strings.ReplaceAll(prompt, "\r\n", "\n"), "\n") {
		if strings.TrimSpace(l) == "" {
			continue
		}
		if strings.HasSuffix(strings.TrimSpace(l), "?") {
			return "", false
		}
		if m := reApruebo.FindStringSubmatch(l); m != nil {
			return m[1], true
		}
		return "", false
	}
	return "", false
}

// aprobacionCLI: solo un evento UserPromptSubmit cuenta como mensaje del humano. PreToolUse, PostToolUse,
// Stop o cualquier otra entrada se ignoran. Nunca bloquea el mensaje: un fallo sale con exit 1 y stderr.
func aprobacionCLI(raiz string) {
	var e struct {
		Evento string `json:"hook_event_name"`
		Prompt string `json:"prompt"`
	}
	datos, _ := io.ReadAll(os.Stdin)
	if err := json.Unmarshal(datos, &e); err != nil {
		fallar(1, fmt.Errorf("aprobacion: entrada del hook ilegible: %w", err))
	}
	if e.Evento != "UserPromptSubmit" {
		return
	}
	id, ok := idAprobado(e.Prompt)
	if !ok {
		return
	}
	if _, _, err := backlog.Ver(raiz, id); err != nil {
		return // id sin ficha: no es una aprobación de nada
	}
	if _, err := backlog.Aprobar(raiz, id, e.Prompt, time.Now()); err != nil {
		fallar(1, fmt.Errorf("aprobacion %s: %w", id, err))
	}
}
