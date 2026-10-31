package entorno

// E000/F008/S003 (D000079): comprueba que Claude Code encierre los comandos Bash del agente en el aislamiento
// del sistema operativo. Doc oficial consultada el 2026-10-31: code.claude.com/docs/en/sandboxing, /settings
// y /managed-settings. Solo cuentan los ajustes gestionados: con ellos Claude Code ignora las listas del repo
// (sandbox.excludedCommands, sandbox.filesystem.allowWrite) que el agente podría editar. Opción 1 del humano.
// Límite: no se leen --settings de línea de comandos, el registro de Windows (HKLM/HKCU
// SOFTWARE\Policies\ClaudeCode) ni los ajustes remotos (remote-settings.json).

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ArchivosAjustes: rutas de los ajustes gestionados de Claude Code.
type ArchivosAjustes struct {
	Gestionado  string // managed-settings.json del sistema
	GestionadoD string // carpeta managed-settings.d junto a él; sus *.json se aplican en orden alfabético
}

// ArchivosReales: las rutas de esta máquina según el sistema operativo.
func ArchivosReales(goos string) ArchivosAjustes {
	sistema := "/etc/claude-code"
	switch goos {
	case "darwin":
		sistema = "/Library/Application Support/ClaudeCode"
	case "windows":
		sistema = `C:\Program Files\ClaudeCode`
	}
	return ArchivosAjustes{
		Gestionado:  filepath.Join(sistema, "managed-settings.json"),
		GestionadoD: filepath.Join(sistema, "managed-settings.d"),
	}
}

// estrictas: las tres claves de `sandbox` y el valor que exige el aislamiento estricto.
var estrictas = []struct {
	clave  string
	quiere bool
}{{"enabled", true}, {"allowUnsandboxedCommands", false}, {"failIfUnavailable", true}}

// sandbox: el objeto `sandbox` de un archivo de ajustes; archivo ausente = vacío.
func sandbox(ruta string) (map[string]any, error) {
	b, err := os.ReadFile(ruta)
	if errors.Is(err, fs.ErrNotExist) {
		return map[string]any{}, nil
	}
	if err != nil {
		return nil, err
	}
	var a struct {
		Sandbox map[string]any `json:"sandbox"`
	}
	if err := json.Unmarshal(b, &a); err != nil {
		return nil, fmt.Errorf("%s ilegible: %w", ruta, err)
	}
	if a.Sandbox == nil {
		a.Sandbox = map[string]any{}
	}
	return a.Sandbox, nil
}

// Aislamiento: fallo para el gate y aviso para el arranque; "" = nada que decir. En Windows nativo no hay
// aislamiento: nunca falla y siempre avisa. En el resto, las tres claves salen de los ajustes gestionados
// (managed-settings.json y luego managed-settings.d en orden); si falta alguna, falla y el arranque lo avisa.
func Aislamiento(goos string, a ArchivosAjustes) (fallo, aviso string) {
	if goos == "windows" {
		return "", "AVISO aislamiento: en Windows nativo Claude Code corre Bash sin aislamiento; los frenos contra el agente son probabilísticos (D000079). Para frenos deterministas, usa WSL2 con los ajustes sandbox estrictos en ajustes gestionados."
	}
	gestion := map[string]any{}
	rutas := []string{a.Gestionado}
	if d, err := filepath.Glob(filepath.Join(a.GestionadoD, "*.json")); err == nil {
		sort.Strings(d)
		rutas = append(rutas, d...)
	}
	for _, r := range rutas {
		s, err := sandbox(r)
		if err != nil {
			m := "aislamiento de Bash: " + err.Error()
			return m, "AVISO " + m
		}
		for k, v := range s {
			gestion[k] = v
		}
	}
	var malas []string
	for _, e := range estrictas {
		if gestion[e.clave] != e.quiere {
			malas = append(malas, fmt.Sprintf("sandbox.%s=%v", e.clave, e.quiere))
		}
	}
	if len(malas) == 0 {
		return "", ""
	}
	m := fmt.Sprintf("aislamiento de Bash no estricto (D000079): falta %s en ajustes gestionados (%s); los del usuario o del repo no cuentan", strings.Join(malas, ", "), a.Gestionado)
	return m, "AVISO " + m
}
