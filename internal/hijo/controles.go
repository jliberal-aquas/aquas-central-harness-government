package hijo

// E003/F001/S003: los controles de la central que viajan al hijo (herencia.json, destino «todos»).

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/compilador"
	"github.com/jliberal-aquas/aquas-central-harness-government/internal/herencia"
)

// rutaPlantilla: ganchos de la central, con su evento y su filtro.
const rutaPlantilla = "gobierno/config/claude_settings.json"

// PrefijoGancho: inicio de todo gancho que aquas pone en el hijo; update reconoce los suyos por él.
// Falla cerrada, como .claude/hooks/aquas.sh en la central: sin carpeta o sin aquas en el PATH sale 2
// (bloquea); un 1 o un 127 dejarían pasar todo.
const PrefijoGancho = `cd "$CLAUDE_PROJECT_DIR" || exit 2; command -v aquas >/dev/null || { echo "aquas: no está en el PATH; el gancho bloquea (falla cerrada)" >&2; exit 2; }; aquas `

// ComandoGancho arma el comando del gancho de la central `orden` en el hijo.
func ComandoGancho(orden string) string { return PrefijoGancho + orden }

// existe dice si la central instalada trae rel; una central sin él (antigua o de prueba) no lo exige.
func existe(central, rel string) (bool, error) {
	_, err := os.Stat(filepath.Join(central, filepath.FromSlash(rel)))
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return err == nil, err
}

// ganchosQueViajan lee la central instalada y devuelve cada gancho con fila «todos» en herencia.json,
// con el mismo evento y filtro que en la central; corre `aquas <orden>` desde la raíz del hijo.
func ganchosQueViajan(dir string) ([]gancho, error) {
	central := filepath.Join(dir, Carpeta, "central")
	if hay, err := existe(central, herencia.Ruta); !hay {
		return nil, err
	}
	filas, err := herencia.Leer(central)
	if err != nil {
		return nil, err
	}
	viaja := map[string]bool{}
	for _, f := range filas {
		viaja[f.Control] = f.Destino == "todos"
	}
	b, err := os.ReadFile(filepath.Join(central, filepath.FromSlash(rutaPlantilla)))
	if err != nil {
		return nil, err
	}
	var p struct {
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
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, fmt.Errorf("%s ilegible: %w", rutaPlantilla, err)
	}
	so, ok := p.PorSO[runtime.GOOS]
	if !ok || so.Shell == "" {
		return nil, fmt.Errorf("%s: sin shell declarado para el SO %q", rutaPlantilla, runtime.GOOS)
	}
	var gs []gancho
	for _, h := range p.Hooks {
		if !viaja["gancho:"+h.Orden] {
			continue
		}
		herr := append([]string{}, h.Herramientas...)
		if h.ConShell {
			herr = append(herr, so.HerramientasShell...)
		}
		m := strings.Join(herr, "|")
		if h.Matcher != "" {
			m = h.Matcher
		}
		gs = append(gs, gancho{h.Evento, m, so.Shell, ComandoGancho(h.Orden)})
	}
	return gs, nil
}

// agentes devuelve los .claude/agents/*.md compilados desde la central instalada (ruta → contenido).
func agentes(dir string) (map[string][]byte, error) {
	central := filepath.Join(dir, Carpeta, "central")
	if hay, err := existe(central, compilador.RutaAgentes); !hay {
		return nil, err
	}
	salidas, err := compilador.Planificar(central)
	if err != nil {
		return nil, err
	}
	out := map[string][]byte{}
	for _, s := range salidas {
		if strings.HasPrefix(s.Ruta, compilador.DirAgentes+"/") {
			out[s.Ruta] = s.Contenido
		}
	}
	return out, nil
}
