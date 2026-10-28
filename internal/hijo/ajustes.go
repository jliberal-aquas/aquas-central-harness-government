package hijo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// ComandoBuild (E002/F002/S006): lo que corre el hook PostToolUse del hijo tras cada edición.
const ComandoBuild = "aquas build ."

// AsegurarHook (E002/F002/S006) agrega a .claude/settings.json del hijo un hook PostToolUse
// que corre ComandoBuild. Conserva todo lo demás y no lo duplica. Un settings inválido
// no se toca: devuelve error.
func AsegurarHook(dir string) error {
	p := filepath.Join(dir, ".claude", "settings.json")
	s := map[string]any{}
	b, err := os.ReadFile(p)
	switch {
	case err == nil:
		if err := json.Unmarshal(b, &s); err != nil {
			return fmt.Errorf("%s inválido; corrígelo y repite: %w", p, err)
		}
	case !errors.Is(err, fs.ErrNotExist):
		return err
	}
	hooks, _ := s["hooks"].(map[string]any)
	if hooks == nil {
		if _, hay := s["hooks"]; hay {
			return fmt.Errorf("%s: «hooks» no es un objeto", p)
		}
		hooks = map[string]any{}
	}
	post, _ := hooks["PostToolUse"].([]any)
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
		"hooks":   []any{map[string]any{"type": "command", "command": ComandoBuild}},
	})
	s["hooks"] = hooks
	out, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, append(out, '\n'), 0o644)
}
