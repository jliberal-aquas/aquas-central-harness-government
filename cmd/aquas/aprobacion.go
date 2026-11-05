package main

// E001/F010/S001: hook UserPromptSubmit que registra «apruebo <id>» del mensaje del humano.

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/backlog"
)

// aprobacionCLI: solo un evento UserPromptSubmit cuenta como mensaje del humano. PreToolUse, PostToolUse,
// Stop o cualquier otra entrada se ignoran. Nunca bloquea el mensaje: un fallo sale con exit 1 y stderr.
func aprobacionCLI(raiz string) {
	var e struct {
		Evento   string `json:"hook_event_name"`
		Prompt   string `json:"prompt"`
		Registro string `json:"transcript_path"`
	}
	datos, _ := io.ReadAll(os.Stdin)
	if err := json.Unmarshal(datos, &e); err != nil {
		fallar(1, fmt.Errorf("aprobacion: entrada del hook ilegible: %w", err))
	}
	if e.Evento != "UserPromptSubmit" {
		return
	}
	if backlog.EsAprobacionRetro(e.Prompt) { // E001/F018/S002 (decisión 1A): «apruebo retro» autoriza un retro close
		if err := backlog.AprobarRetro(raiz, e.Prompt, e.Registro, time.Now()); err != nil {
			fallar(1, fmt.Errorf("aprobacion retro: %w", err))
		}
		return
	}
	prompt := backlog.Resolver(e.Registro, e.Prompt) // E000/F013/S003: «sí» vale como la línea que el agente propuso
	id, ok := backlog.IDAprobado(prompt)
	if !ok {
		return
	}
	if _, _, err := backlog.Ver(raiz, id); err != nil {
		return // id sin ficha: no es una aprobación de nada
	}
	if _, err := backlog.Aprobar(raiz, id, prompt, e.Registro, time.Now()); err != nil {
		fallar(1, fmt.Errorf("aprobacion %s: %w", id, err))
	}
}
