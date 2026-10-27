package main

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
)

// ultimoMensaje (D000053): SubagentStop no trae last_assistant_message en esta versión de Claude Code
// (entrada real observada 2026-10-27). El cierre se lee del transcript del
// subagente: el último texto del asistente o el mensaje de su SubagentHandback. "" si no hay ninguno.
func ultimoMensaje(ruta string) string {
	f, err := os.Open(ruta)
	if err != nil {
		return ""
	}
	defer func() { _ = f.Close() }()
	var ultimo string
	r := bufio.NewReader(f)
	for {
		linea, err := r.ReadBytes('\n')
		if len(linea) > 0 {
			var l struct {
				Message struct {
					Role    string          `json:"role"`
					Content json.RawMessage `json:"content"`
				} `json:"message"`
			}
			if json.Unmarshal(linea, &l) == nil && l.Message.Role == "assistant" {
				var crudas []json.RawMessage
				if json.Unmarshal(l.Message.Content, &crudas) == nil {
					for _, cruda := range crudas { // cada parte por separado: una rara no descarta las demás
						var p struct {
							Type  string `json:"type"`
							Text  string `json:"text"`
							Name  string `json:"name"`
							Input struct {
								Message string `json:"message"`
							} `json:"input"`
						}
						if json.Unmarshal(cruda, &p) != nil {
							continue
						}
						switch {
						case p.Type == "text" && strings.TrimSpace(p.Text) != "":
							ultimo = p.Text
						case p.Type == "tool_use" && p.Name == "SubagentHandback" && p.Input.Message != "":
							ultimo = p.Input.Message
						}
					}
				}
			}
		}
		if err != nil {
			return ultimo
		}
	}
}
