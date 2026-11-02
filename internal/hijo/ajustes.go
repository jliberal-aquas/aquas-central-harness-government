package hijo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ComandoBuild (E002/F002/S006): lo que corre el hook PostToolUse del hijo tras cada edición.
// Usa la raíz del proyecto que da Claude Code, no la carpeta actual; corre en bash, como el hook central.
const ComandoBuild = `aquas build "$CLAUDE_PROJECT_DIR"`

// gancho: una entrada de hooks del settings del hijo.
type gancho struct {
	evento, matcher, shell, comando string
}

// AsegurarHook (E002/F002/S006, E003/F001/S003) agrega a .claude/settings.json del hijo el hook
// PostToolUse de ComandoBuild y cada gancho de la central que viaja (ganchosQueViajan). Conserva
// todo lo demás y no duplica. Un settings con forma inesperada no se toca: devuelve error.
// Al reescribir, el orden de claves queda alfabético.
func AsegurarHook(dir string) error {
	viajan, err := ganchosQueViajan(dir)
	if err != nil {
		return err
	}
	gs := append([]gancho{{"PostToolUse", "Edit|Write", "bash", ComandoBuild}}, viajan...)
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
	// Primero se revisa la forma de cada evento existente: un error no deja nada a medias.
	for ev, v := range hooks {
		if l, _ := v.([]any); l == nil && v != nil {
			return fmt.Errorf("%s: «hooks.%s» no es una lista", p, ev)
		}
	}
	// Hallazgos A1.3 y A2.d: un gancho de aquas que ya no viaja, o con otro filtro o shell, se quita.
	cambio := podar(hooks, gs)
	for _, g := range gs {
		lista, _ := hooks[g.evento].([]any)
		if tiene(lista, g) {
			continue
		}
		grupo := map[string]any{"hooks": []any{map[string]any{"type": "command", "shell": g.shell, "command": g.comando}}}
		if g.matcher != "" {
			grupo["matcher"] = g.matcher
		}
		hooks[g.evento] = append(lista, grupo)
		cambio = true
	}
	if !cambio {
		return nil
	}
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

// clave identifica un gancho por evento, filtro, shell y comando.
func clave(evento string, grupo, h map[string]any) string {
	m, _ := grupo["matcher"].(string)
	sh, _ := h["shell"].(string)
	c, _ := h["command"].(string)
	return evento + "\x00" + m + "\x00" + sh + "\x00" + c
}

// podar quita de hooks cada gancho puesto por aquas (PrefijoGancho) que no está en gs tal cual
// (evento, filtro, shell y comando). Solo se borra un grupo o un evento que quedó vacío por la poda;
// lo propio del hijo, vacío o no, no se toca (hallazgo A2.b). Devuelve si cambió algo.
func podar(hooks map[string]any, gs []gancho) bool {
	quiere := map[string]bool{}
	for _, g := range gs {
		quiere[g.evento+"\x00"+g.matcher+"\x00"+g.shell+"\x00"+g.comando] = true
	}
	cambio := false
	for ev, v := range hooks {
		lista, _ := v.([]any)
		var nueva []any
		podado := false
		for _, g := range lista {
			gm, _ := g.(map[string]any)
			hs, ok := gm["hooks"].([]any)
			if !ok {
				nueva = append(nueva, g)
				continue
			}
			var quedan []any
			for _, h := range hs {
				hm, _ := h.(map[string]any)
				c, _ := hm["command"].(string)
				if strings.HasPrefix(c, PrefijoGancho) && !quiere[clave(ev, gm, hm)] {
					continue
				}
				quedan = append(quedan, h)
			}
			if len(quedan) == len(hs) {
				nueva = append(nueva, g)
				continue
			}
			podado = true
			if len(quedan) > 0 {
				gm["hooks"] = quedan
				nueva = append(nueva, gm)
			}
		}
		if !podado {
			continue
		}
		cambio = true
		if len(nueva) == 0 {
			delete(hooks, ev)
		} else {
			hooks[ev] = nueva
		}
	}
	return cambio
}

// tiene dice si algún grupo de la lista ya tiene el gancho g tal cual.
func tiene(lista []any, g gancho) bool {
	want := g.evento + "\x00" + g.matcher + "\x00" + g.shell + "\x00" + g.comando
	for _, gr := range lista {
		gm, _ := gr.(map[string]any)
		hs, _ := gm["hooks"].([]any)
		for _, h := range hs {
			if hm, _ := h.(map[string]any); hm != nil && clave(g.evento, gm, hm) == want {
				return true
			}
		}
	}
	return false
}
