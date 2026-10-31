// Package entorno declara dónde corre el harness: sistema operativo y raíz del repo.
//
//	gobierno/config/entorno.local.json   local, no versionado; lo escribe `aquas init`
//	gobierno/config/claude_settings.json canónico versionado → .claude/settings.json (local)
//
// La guardia bloquea todo si el entorno falta o no coincide con la realidad:
// mover el repo de carpeta o cambiar de SO exige volver a correr `aquas init`.
package entorno

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	RutaEntorno   = "gobierno/config/entorno.local.json"
	RutaPlantilla = "gobierno/config/claude_settings.json"
	RutaSettings  = ".claude/settings.json"
	version       = 1
)

// Entorno es el contenido de entorno.local.json.
type Entorno struct {
	Version int    `json:"version"`
	SO      string `json:"so"`
	Raiz    string `json:"raiz"`
	// ShellHumano (E001/F016/S001): shell donde el humano corre los comandos que le da el agente: powershell o bash.
	ShellHumano string `json:"shell_humano,omitempty"`
}

// Leer carga el entorno declarado.
func Leer(raiz string) (Entorno, error) {
	var e Entorno
	b, err := os.ReadFile(filepath.Join(raiz, RutaEntorno))
	if err != nil {
		return e, err
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&e); err != nil {
		return e, fmt.Errorf("%s ilegible: %w", RutaEntorno, err)
	}
	if e.Version != version {
		return e, fmt.Errorf("%s: version %d, se esperaba %d", RutaEntorno, e.Version, version)
	}
	return e, nil
}

// Validar compara el entorno declarado con el real; "" = coincide, si no, el motivo del bloqueo.
func Validar(raiz string) string {
	const arreglo = " → corre `go run ./cmd/aquas init` en la raíz del repo"
	e, err := Leer(raiz)
	if os.IsNotExist(err) {
		return "falta " + RutaEntorno + arreglo
	}
	if err != nil {
		return err.Error() + arreglo
	}
	if e.SO != runtime.GOOS {
		return fmt.Sprintf("sistema operativo declarado %q ≠ real %q%s", e.SO, runtime.GOOS, arreglo)
	}
	if !mismaRuta(e.Raiz, raiz) {
		return fmt.Sprintf("raíz declarada %q ≠ real %q (el repo no se mueve de carpeta)%s", e.Raiz, filepath.ToSlash(raiz), arreglo)
	}
	return ""
}

// Iniciar detecta SO y raíz (git) y escribe entorno.local.json; exige correr desde la raíz.
func Iniciar(raiz string) (Entorno, error) {
	out, err := exec.Command("git", "-C", raiz, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return Entorno{}, fmt.Errorf("no se pudo leer la raíz con git: %w", err)
	}
	top := strings.TrimSpace(string(out))
	if !mismaRuta(top, raiz) {
		return Entorno{}, fmt.Errorf("corre `aquas init` desde la raíz del repo (%s), no desde %s", top, raiz)
	}
	e := Entorno{Version: version, SO: runtime.GOOS, Raiz: filepath.ToSlash(filepath.Clean(top))}
	b, _ := json.MarshalIndent(e, "", "  ")
	if err := os.WriteFile(filepath.Join(raiz, RutaEntorno), append(b, '\n'), 0o644); err != nil {
		return e, err
	}
	return e, nil
}

// mismaRuta compara dos rutas limpias; en Windows sin distinguir mayúsculas.
func mismaRuta(a, b string) bool {
	a = filepath.ToSlash(filepath.Clean(a))
	b = filepath.ToSlash(filepath.Clean(b))
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}

// --- settings.json ---

type plantilla struct {
	Agent string `json:"agent"`
	Hooks []struct {
		Evento       string   `json:"evento"`
		Herramientas []string `json:"herramientas"`
		Matcher      string   `json:"matcher"` // eventos sin herramienta (SessionStart)
		ConShell     bool     `json:"con_shell"`
		Orden        string   `json:"orden"`
	} `json:"hooks"`
	PorSO map[string]struct {
		Shell             string   `json:"shell"`
		HerramientasShell []string `json:"herramientas_shell"`
	} `json:"por_so"`
}

type hookCmd struct {
	Type    string `json:"type"`
	Command string `json:"command"`
	Shell   string `json:"shell"`
}

type grupo struct {
	Matcher string    `json:"matcher,omitempty"`
	Hooks   []hookCmd `json:"hooks"`
}

type settings struct {
	Schema string             `json:"$schema"`
	Agent  string             `json:"agent"`
	Hooks  map[string][]grupo `json:"hooks"`
}

// generarSettings arma settings.json para un SO desde la plantilla canónica. Bytes estables.
func generarSettings(fuente []byte, so string) ([]byte, error) {
	var p plantilla
	dec := json.NewDecoder(bytes.NewReader(fuente))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		return nil, fmt.Errorf("%s ilegible: %w", RutaPlantilla, err)
	}
	cfg, ok := p.PorSO[so]
	if !ok || cfg.Shell == "" || len(cfg.HerramientasShell) == 0 {
		return nil, fmt.Errorf("%s: sin shell declarado para el SO %q", RutaPlantilla, so)
	}
	s := settings{Schema: "https://json.schemastore.org/claude-code-settings.json", Agent: p.Agent, Hooks: map[string][]grupo{}}
	for _, h := range p.Hooks {
		herr := append([]string{}, h.Herramientas...)
		if h.ConShell {
			herr = append(herr, cfg.HerramientasShell...)
		}
		matcher := strings.Join(herr, "|")
		if h.Matcher != "" {
			matcher = h.Matcher
		}
		s.Hooks[h.Evento] = append(s.Hooks[h.Evento], grupo{
			Matcher: matcher,
			Hooks: []hookCmd{{
				Type:    "command",
				Command: `sh "${CLAUDE_PROJECT_DIR}/.claude/hooks/aquas.sh" ` + h.Orden,
				Shell:   cfg.Shell,
			}},
		})
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(s); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Settings genera el settings.json esperado para el entorno declarado.
func Settings(raiz string) ([]byte, error) {
	e, err := Leer(raiz)
	if err != nil {
		return nil, err
	}
	fuente, err := os.ReadFile(filepath.Join(raiz, RutaPlantilla))
	if err != nil {
		return nil, err
	}
	return generarSettings(fuente, e.SO)
}

// CompilarSettings escribe .claude/settings.json; devuelve true si cambió.
func CompilarSettings(raiz string) (bool, error) {
	b, err := Settings(raiz)
	if err != nil {
		return false, err
	}
	ruta := filepath.Join(raiz, RutaSettings)
	if actual, err := os.ReadFile(ruta); err == nil && bytes.Equal(actual, b) {
		return false, nil
	}
	return true, os.WriteFile(ruta, b, 0o644)
}

// VerificarSettings: "" si .claude/settings.json coincide con lo regenerado.
func VerificarSettings(raiz string) string {
	b, err := Settings(raiz)
	if err != nil {
		return err.Error()
	}
	actual, err := os.ReadFile(filepath.Join(raiz, RutaSettings))
	if err != nil || !bytes.Equal(actual, b) {
		return RutaSettings + " desfasado o editado a mano → go run ./cmd/aquas compile"
	}
	return ""
}
