package entorno

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const estricto = `{"sandbox":{"enabled":true,"allowUnsandboxedCommands":false,"failIfUnavailable":true}}`

// TestAislamientoEstricto (E000/F008/S003, D000079): sin los tres ajustes estrictos de aislamiento de Bash en
// ajustes gestionados, el gate falla en macOS/Linux/WSL2; en Windows nativo solo hay aviso.
// Rojo cuando: sin sandbox estricto en Linux el gate pasa.
// Verde cuando: falla en Linux sin ajustes y solo avisa en Windows nativo.
func TestAislamientoEstricto(t *testing.T) {
	type caso struct {
		goos                    string
		gestionado, gestionadoD string // contenido; "" = sin archivo
		falla                   string // texto que debe tener el fallo; "" = pasa
	}
	casos := map[string]caso{
		"linux sin archivos":                  {goos: "linux", falla: "sandbox.enabled"},
		"linux gestionado estricto":           {goos: "linux", gestionado: estricto},
		"darwin gestionado estricto":          {goos: "darwin", gestionado: estricto},
		"darwin sin archivos":                 {goos: "darwin", falla: "sandbox.failIfUnavailable"},
		"linux JSON roto":                     {goos: "linux", gestionado: `{"sandbox":`, falla: "ilegible"},
		"linux falta una clave":               {goos: "linux", gestionado: `{"sandbox":{"enabled":true,"allowUnsandboxedCommands":false}}`, falla: "sandbox.failIfUnavailable"},
		"linux valor de otro tipo":            {goos: "linux", gestionado: `{"sandbox":{"enabled":"true","allowUnsandboxedCommands":false,"failIfUnavailable":true}}`, falla: "sandbox.enabled"},
		"linux solo en .d":                    {goos: "linux", gestionadoD: estricto},
		"linux gestionado.d posterior rebaja": {goos: "linux", gestionado: estricto, gestionadoD: `{"sandbox":{"allowUnsandboxedCommands":true}}`, falla: "sandbox.allowUnsandboxedCommands"},
		"windows sin archivos: solo aviso":    {goos: "windows"},
		"windows estricto: solo aviso":        {goos: "windows", gestionado: estricto},
	}
	for nombre, c := range casos {
		dir := t.TempDir()
		poner := func(rel, txt string) string {
			ruta := filepath.Join(dir, filepath.FromSlash(rel))
			if txt == "" {
				return ruta
			}
			if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(ruta, []byte(txt), 0o644); err != nil {
				t.Fatal(err)
			}
			return ruta
		}
		a := ArchivosAjustes{
			Gestionado:  poner("gestion/managed-settings.json", c.gestionado),
			GestionadoD: filepath.Join(dir, "gestion", "managed-settings.d"),
		}
		poner("gestion/managed-settings.d/10-base.json", `{"otro":1}`)
		poner("gestion/managed-settings.d/20-equipo.json", c.gestionadoD)
		// Ajustes de usuario y del repo estrictos no cuentan: el aislamiento solo sale de los gestionados.
		poner("usuario/settings.json", estricto)
		poner("repo/.claude/settings.local.json", estricto)
		fallo, aviso := Aislamiento(c.goos, a)
		switch {
		case c.goos == "windows":
			if fallo != "" || !strings.Contains(aviso, "D000079") {
				t.Errorf("%s: en Windows nativo no falla y avisa con D000079; dio fallo=%q aviso=%q", nombre, fallo, aviso)
			}
		case c.falla == "":
			if fallo != "" || aviso != "" {
				t.Errorf("%s: debió pasar sin aviso; dio fallo=%q aviso=%q", nombre, fallo, aviso)
			}
		default:
			if !strings.Contains(fallo, c.falla) || !strings.Contains(aviso, c.falla) {
				t.Errorf("%s: fallo y aviso nombran %s; dio fallo=%q aviso=%q", nombre, c.falla, fallo, aviso)
			}
		}
	}
}
