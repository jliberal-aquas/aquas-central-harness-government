package hijo

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"testing/fstest"
)

// centralReal arma la central como la embebe el binario: gobierno/canonico y la config que viaja.
func centralReal(t *testing.T) fstest.MapFS {
	t.Helper()
	m := fstest.MapFS{}
	raiz := filepath.Join("..", "..")
	err := filepath.WalkDir(filepath.Join(raiz, "gobierno", "canonico"), func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := os.ReadFile(p)
		rel, _ := filepath.Rel(raiz, p)
		m[filepath.ToSlash(rel)] = &fstest.MapFile{Data: b}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []string{"agentes.json", "claude_settings.json", "documentos.json", "herencia.json"} {
		b, err := os.ReadFile(filepath.Join(raiz, "gobierno", "config", c))
		if err != nil {
			t.Fatal(err)
		}
		m["gobierno/config/"+c] = &fstest.MapFile{Data: b}
	}
	return m
}

// ganchosHijo devuelve cada gancho del settings del hijo como «evento|matcher|shell|comando».
func ganchosHijo(t *testing.T, dir string) []string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	var s struct {
		Hooks map[string][]struct {
			Matcher string `json:"matcher"`
			Hooks   []struct {
				Shell   string `json:"shell"`
				Command string `json:"command"`
			} `json:"hooks"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(b, &s); err != nil {
		t.Fatal(err)
	}
	var out []string
	for ev, gs := range s.Hooks {
		for _, g := range gs {
			for _, h := range g.Hooks {
				out = append(out, ev+"|"+g.Matcher+"|"+h.Shell+"|"+h.Command)
			}
		}
	}
	return out
}

// esperados: los ganchos que viajan según la config real, calculados aparte de la implementación.
func esperados(t *testing.T) (viajan, quedan []string) {
	t.Helper()
	cfg := filepath.Join("..", "..", "gobierno", "config")
	var s struct {
		Hooks []struct {
			Evento       string   `json:"evento"`
			Herramientas []string `json:"herramientas"`
			Matcher      string   `json:"matcher"`
			ConShell     bool     `json:"con_shell"`
			Orden        string   `json:"orden"`
		} `json:"hooks"`
		PorSO map[string]struct {
			Shell             string   `json:"shell"`
			HerramientasShell []string `json:"herramientas_shell"`
		} `json:"por_so"`
	}
	var h struct {
		Controles []struct{ Control, Destino string } `json:"controles"`
	}
	for p, v := range map[string]any{"claude_settings.json": &s, "herencia.json": &h} {
		b, err := os.ReadFile(filepath.Join(cfg, p))
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(b, v); err != nil {
			t.Fatal(err)
		}
	}
	destino := map[string]string{}
	for _, f := range h.Controles {
		destino[f.Control] = f.Destino
	}
	so := s.PorSO[runtime.GOOS]
	for _, g := range s.Hooks {
		m := strings.Join(g.Herramientas, "|")
		if g.ConShell {
			m = strings.Join(append(append([]string{}, g.Herramientas...), so.HerramientasShell...), "|")
		}
		if g.Matcher != "" {
			m = g.Matcher
		}
		linea := g.Evento + "|" + m + "|" + so.Shell + "|" + ComandoGancho(g.Orden)
		if destino["gancho:"+g.Orden] == "todos" {
			viajan = append(viajan, linea)
		} else {
			quedan = append(quedan, "aquas "+g.Orden)
		}
	}
	return viajan, quedan
}

// TestInstallPoneControlesQueViajan (E003/F001/S003): install y update ponen en el hijo cada gancho «todos» y los agentes.
// Rojo cuando: falta un gancho todos o un agente en el hijo, se duplica, se pierde lo propio o se toca un settings roto.
// Verde cuando: el hijo tiene cada gancho todos una vez, sin los de la central, con lo suyo y los agentes compilados.
func TestInstallPoneControlesQueViajan(t *testing.T) {
	viajan, quedan := esperados(t)
	if len(viajan) == 0 || len(quedan) == 0 {
		t.Fatalf("la config real debe tener ganchos que viajan y que se quedan; viajan %v, quedan %v", viajan, quedan)
	}
	central := centralReal(t)
	propio := "PreToolUse|Bash|bash|./formato.sh"
	dir := carpeta(t, map[string]string{".claude/settings.json": `{"model": "x", "hooks": {"PreToolUse": [
		{"matcher": "Bash", "hooks": [{"type": "command", "shell": "bash", "command": "./formato.sh"}]}]}}`})

	revisar := func(caso string) {
		t.Helper()
		gs := ganchosHijo(t, dir)
		for _, v := range append(viajan, propio) {
			if n := contar(gs, v); n != 1 {
				t.Errorf("%s: %q debe estar una vez; está %d. Ganchos: %v", caso, v, n, gs)
			}
		}
		build := 0
		for _, g := range gs {
			if strings.HasSuffix(g, "|"+ComandoBuild) {
				build++
			}
			for _, q := range quedan {
				if strings.HasSuffix(g, " "+q) || strings.HasSuffix(g, "|"+q) {
					t.Errorf("%s: %q es de la central y no debe instalarse: %s", caso, q, g)
				}
			}
		}
		if build != 1 {
			t.Errorf("%s: el gancho de build debe seguir una vez; está %d", caso, build)
		}
		for _, a := range []string{"orchestrator", "worker-sonnet", "worker-haiku", "auditor"} {
			want, err := os.ReadFile(filepath.Join("..", "..", ".claude", "agents", a+".md"))
			if err != nil {
				t.Fatal(err)
			}
			got, err := os.ReadFile(filepath.Join(dir, ".claude", "agents", a+".md"))
			if err != nil || string(got) != string(want) {
				t.Errorf("%s: .claude/agents/%s.md del hijo debe ser el compilado de la central (err %v)", caso, a, err)
			}
		}
	}

	// Hijo nuevo con un settings y un gancho propios.
	if err := Install(dir, central, "v1.0.0"); err != nil {
		t.Fatal(err)
	}
	if err := Listo(dir); err != nil {
		t.Fatal(err)
	}
	revisar("install")
	// Hijo ya instalado: update no duplica.
	if _, err := Update(dir, central, "v1.0.1"); err != nil {
		t.Fatal(err)
	}
	if err := Listo(dir); err != nil {
		t.Fatal(err)
	}
	revisar("update")

	// Hallazgo A1.3: un gancho de aquas que ya no viaja (fila pasada a central) se quita; lo propio del mismo grupo queda.
	// A2.b: un grupo propio vacío no se toca. A2.d: un gancho que viaja con otro filtro se reemplaza.
	viejo, _ := json.Marshal(ComandoGancho("ya-no-viaja"))
	guard, _ := json.Marshal(ComandoGancho("guard"))
	if err := os.WriteFile(filepath.Join(dir, ".claude", "settings.json"), []byte(`{"hooks": {"Stop": [
		{"hooks": [{"type": "command", "shell": "bash", "command": `+string(viejo)+`},
		           {"type": "command", "shell": "bash", "command": "./mio.sh"}]}],
		"Notification": [{"matcher": "vacio", "hooks": []}],
		"PreToolUse": [{"matcher": "FiltroViejo", "hooks": [{"type": "command", "shell": "bash", "command": `+string(guard)+`}]}]}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Listo(dir); err != nil {
		t.Fatal(err)
	}
	gs := ganchosHijo(t, dir)
	if contar(gs, "Stop||bash|./mio.sh") != 1 {
		t.Errorf("poda: el gancho propio debe quedar una vez; ganchos %v", gs)
	}
	for _, g := range gs {
		if strings.Contains(g, "ya-no-viaja") || strings.Contains(g, "FiltroViejo") {
			t.Errorf("poda: el gancho de aquas que ya no viaja o con filtro viejo debió quitarse: %s", g)
		}
	}
	if b, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json")); !strings.Contains(string(b), `"matcher": "vacio"`) {
		t.Errorf("poda: el grupo propio vacío de Notification debe quedar; settings %s", b)
	}
	for _, v := range viajan {
		if n := contar(gs, v); n != 1 {
			t.Errorf("poda: %q debe estar una vez; está %d", v, n)
		}
	}

	// Hallazgo A1.2: sin aquas en el PATH o sin carpeta, el gancho bloquea (exit 2), no deja pasar.
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Fatalf("falla cerrada: hace falta sh para probar el comando del gancho: %v", err)
	}
	// A2.a: en «carpeta inexistente» hay un aquas falso que deja pasar (exit 0): solo el `cd … || exit 2` puede frenar.
	falso := t.TempDir()
	if err := os.WriteFile(filepath.Join(falso, "aquas"), []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	// A3.1: cada caso prueba su rama. El control (carpeta válida + aquas falso → 0) demuestra que sh halla el falso y entra a la carpeta.
	correr := func(path, proyecto string) (int, string) {
		cmd := exec.Command(sh, "-c", ComandoGancho("guard"))
		cmd.Env = []string{"PATH=" + path, "CLAUDE_PROJECT_DIR=" + proyecto}
		out, err := cmd.CombinedOutput()
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return ee.ExitCode(), string(out)
		}
		if err != nil {
			t.Fatalf("falla cerrada: %v", err)
		}
		return 0, string(out)
	}
	if c, out := correr(falso, dir); c != 0 {
		t.Errorf("falla cerrada, control: con carpeta válida y aquas en el PATH debe salir 0; dio %d %s", c, out)
	}
	if c, out := correr(t.TempDir(), dir); c != 2 || !strings.Contains(out, "no está en el PATH") {
		t.Errorf("falla cerrada, sin aquas en el PATH: debe salir 2 por esa rama; dio %d %s", c, out)
	}
	if c, out := correr(falso, filepath.Join(dir, "no-existe")); c != 2 || strings.Contains(out, "no está en el PATH") {
		t.Errorf("falla cerrada, carpeta inexistente: debe salir 2 por el cd; dio %d %s", c, out)
	}

	// Sin settings previo: install lo crea con los ganchos.
	dir = carpeta(t, nil)
	if err := Install(dir, central, "v1.0.0"); err != nil {
		t.Fatal(err)
	}
	if err := Listo(dir); err != nil {
		t.Fatal(err)
	}
	for _, v := range viajan {
		if n := contar(ganchosHijo(t, dir), v); n != 1 {
			t.Errorf("sin settings previo: %q debe estar una vez; está %d", v, n)
		}
	}

	// Settings roto tras copiar la central: error que pide update y el settings queda intacto.
	for _, roto := range []string{`{"hooks": `, `{"hooks": {"PreToolUse": {}}}`, `[]`} {
		dir = carpeta(t, map[string]string{".claude/settings.json": roto})
		if err := Install(dir, central, "v1.0.0"); err != nil {
			t.Fatal(err)
		}
		err := Listo(dir)
		if err == nil {
			t.Errorf("settings %q: Listo debe fallar", roto)
			continue
		}
		if m := AMedias(err).Error(); !strings.Contains(m, "aquas update") || !strings.Contains(m, "settings.json") {
			t.Errorf("settings %q: el error debe nombrar settings.json y pedir aquas update; dio %q", roto, m)
		}
		if b, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json")); string(b) != roto {
			t.Errorf("settings %q: no debe tocarse; quedó %q", roto, b)
		}
	}
}
