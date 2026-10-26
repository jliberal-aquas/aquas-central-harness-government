// Package calidad corre los linters del gate sobre el código Go de la central (E001/F013/S001).
package calidad

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// Version es la versión de golangci-lint que exige el gate.
const Version = "2.15.0"

// hallazgo reconoce una línea de hallazgo: «archivo:línea:col: mensaje (linter)».
var hallazgo = regexp.MustCompile(`^\S+:\d+(:\d+)?: .+ \([a-z0-9_-]+\)$`)

// Revisar corre golangci-lint con config sobre dir y devuelve un hallazgo por línea.
// Herramienta ausente, versión distinta o caída = error; nunca cuenta como limpio.
func Revisar(dir, config string) ([]string, error) {
	bin, err := exec.LookPath("golangci-lint")
	if err != nil {
		return nil, fmt.Errorf("herramienta ausente: golangci-lint %s no está en el PATH: %w", Version, err)
	}
	v, err := exec.Command(bin, "version").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("golangci-lint version falló: %w: %s", err, strings.TrimSpace(string(v)))
	}
	if err := chequearVersion(string(v)); err != nil {
		return nil, err
	}
	cmd := exec.Command(bin, "run", "-c", config, "--max-issues-per-linter", "0", "--max-same-issues", "0", "./...")
	cmd.Dir = dir
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	err = cmd.Run()
	var hs []string
	for _, l := range strings.Split(out.String(), "\n") {
		l = strings.TrimSpace(l)
		if hallazgo.MatchString(l) {
			hs = append(hs, l)
		}
	}
	if err == nil {
		return hs, nil
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) && ee.ExitCode() == 1 && len(hs) > 0 {
		return hs, nil
	}
	return nil, fmt.Errorf("golangci-lint run falló sin hallazgos legibles: %w: %s%s", err, out.String(), errb.String())
}

// chequearVersion falla si la salida de `golangci-lint version` no es la versión fija.
func chequearVersion(salida string) error {
	if !strings.Contains(salida, "version "+Version+" ") {
		return fmt.Errorf("golangci-lint debe ser la versión %s; salida: %s", Version, strings.TrimSpace(salida))
	}
	return nil
}

// RevisarCentral corre la calidad Go de raiz con la config del gate; sin go.mod no hay código que revisar.
// Un error de go.mod o de la herramienta vuelve como hallazgo: nunca cuenta como limpio.
func RevisarCentral(raiz string) []string {
	if _, err := os.Stat(filepath.Join(raiz, "go.mod")); errors.Is(err, fs.ErrNotExist) {
		return nil
	} else if err != nil {
		return []string{"go.mod ilegible: " + err.Error()}
	}
	hs, err := Revisar(raiz, filepath.Join(raiz, "gobierno", "config", "golangci.yml"))
	if err != nil {
		hs = []string{err.Error()}
	}
	dh, err := RevisarDependencias(raiz, filepath.Join(raiz, "gobierno", "config", "dependencias.json"))
	if err != nil {
		dh = []string{err.Error()}
	}
	return append(hs, dh...)
}
