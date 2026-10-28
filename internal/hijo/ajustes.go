package hijo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// ComandoBuild (E002/F002/S006): lo que corre el hook PostToolUse del hijo tras cada edición.
// Usa la raíz del proyecto que da Claude Code, no la carpeta actual; corre en bash, como el hook central.
const ComandoBuild = `aquas build "$CLAUDE_PROJECT_DIR"`

// AsegurarHook (E002/F002/S006) agrega a .claude/settings.json del hijo un hook PostToolUse
// que corre ComandoBuild. Conserva todo lo demás y no lo duplica. Un settings con forma
// inesperada no se toca: devuelve error. Al reescribir, el orden de claves queda alfabético.
func AsegurarHook(dir string) error {
	p := filepath.Join(dir, ".claude", "settings.json")
	var s map[string]any
	b, err := os.ReadFile(p)
	switch {
	case err == nil:
		d := json.NewDecoder(bytes.NewReader(b))
		d.UseNumber() // conserva enteros grandes tal cual
		var v any
		if err := d.Decode(&v); err != nil {
			return fmt.Errorf("%s inválido; corrígelo: %w", p, err)
		}
		if s, _ = v.(map[string]any); s == nil {
			return fmt.Errorf("%s: debe ser un objeto JSON", p)
		}
	case errors.Is(err, fs.ErrNotExist):
		s = map[string]any{}
	default:
		return err
	}
	hooks := map[string]any{}
	if v, hay := s["hooks"]; hay {
		if hooks, _ = v.(map[string]any); hooks == nil {
			return fmt.Errorf("%s: «hooks» no es un objeto", p)
		}
	}
	var post []any
	if v, hay := hooks["PostToolUse"]; hay {
		if post, _ = v.([]any); post == nil && v != nil {
			return fmt.Errorf("%s: «hooks.PostToolUse» no es una lista", p)
		}
	}
	for _, g := range post {
		gm, _ := g.(map[string]any)
		hs, _ := gm["hooks"].([]any)
		for _, h := range hs {
			if hm, _ := h.(map[string]any); hm["command"] == ComandoBuild {
				return nil
			}
		}
	}
	hooks["PostToolUse"] = append(post, map[string]any{
		"matcher": "Edit|Write",
		"hooks":   []any{map[string]any{"type": "command", "shell": "bash", "command": ComandoBuild}},
	})
	s["hooks"] = hooks
	var out bytes.Buffer
	e := json.NewEncoder(&out)
	e.SetEscapeHTML(false) // «&&» y «<» quedan legibles
	e.SetIndent("", "  ")
	if err := e.Encode(s); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, out.Bytes(), 0o644)
}
