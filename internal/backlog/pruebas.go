package backlog

// E003/F001/S004: un hijo sin Go declara en pruebas.json cómo correr una prueba.

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// RutaPruebas: plantilla del comando de pruebas del repo; sin archivo, rojo y verde usan go test.
const RutaPruebas = "gobierno/config/pruebas.json"

// Con plantilla, una prueba es ruta:nombre; sin espacios ni signos de shell, porque va dentro del comando.
var reNombreLibre = regexp.MustCompile(`^[\w./-]+:[\w./:-]+$`)

const formatoLibre = "ruta:nombre=valida|rojo cuando|verde cuando"

// formato: cómo se escribe una prueba, según haya plantilla (libre) o go test.
func formato(libre bool) string {
	if libre {
		return formatoLibre
	}
	return formatoPrueba
}

// formatoDe: formato del repo; con pruebas.json roto muestra el de go test (el error sale al correr).
func formatoDe(raiz string) string {
	c, _ := plantilla(raiz)
	return formato(c != "")
}

// rutaPeligrosa (hallazgos A1.6 y A2.2): ruta o nombre que empieza con «-» (sería una opción del corredor),
// o ruta que sale del repo.
func rutaPeligrosa(prueba string) bool {
	ruta, n, _ := strings.Cut(prueba, ":")
	return strings.HasPrefix(ruta, "-") || strings.HasPrefix(n, "-") || fueraDelRepo(ruta)
}

// plantilla: el comando de pruebas.json, o "" si el repo no lo tiene (se usa go test).
func plantilla(raiz string) (string, error) {
	b, err := os.ReadFile(filepath.Join(raiz, filepath.FromSlash(RutaPruebas)))
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("%s: %v", RutaPruebas, err)
	}
	var p struct {
		Version int    `json:"version"`
		Comando string `json:"comando"`
	}
	switch {
	case json.Unmarshal(b, &p) != nil:
		return "", fmt.Errorf("%s no es JSON válido; ejemplo: {\"version\": 1, \"comando\": \"pytest {ruta}::{nombre}\"}", RutaPruebas)
	case p.Version != 1:
		return "", fmt.Errorf("%s: version %d; se espera 1", RutaPruebas, p.Version)
	case strings.TrimSpace(p.Comando) == "":
		return "", fmt.Errorf("%s: comando vacío", RutaPruebas)
	case !strings.Contains(p.Comando, "{nombre}"):
		return "", fmt.Errorf("%s: el comando no tiene {nombre}; ejemplo: pytest {ruta}::{nombre}", RutaPruebas)
	}
	return p.Comando, nil
}

// correrPlantilla corre la plantilla con sh, una prueba a la vez, en la raíz del repo.
// Salida 0 = pasa; 126 o 127 = el comando no corrió (no cuenta); otra = falla con sus últimas líneas.
func correrPlantilla(raiz, cmd string, ps []Prueba) (map[string]resultado, error) {
	out := map[string]resultado{}
	for _, p := range ps {
		ruta, n, _ := strings.Cut(p.Nombre, ":")
		c := exec.Command("sh", "-c", strings.NewReplacer("{ruta}", ruta, "{nombre}", n).Replace(cmd))
		c.Dir = raiz
		var o bytes.Buffer
		c.Stdout, c.Stderr = &o, &o
		r := resultado{accion: "pass"}
		if err := c.Run(); err != nil {
			var salio *exec.ExitError
			if !errors.As(err, &salio) {
				return nil, fmt.Errorf("no se pudo correr sh para %s: %v", RutaPruebas, err)
			}
			r = resultado{accion: "fail", salida: ultimas(o.String(), 3)}
			if k := salio.ExitCode(); k == 126 || k == 127 {
				r = resultado{noCorre: true, salida: fmt.Sprintf("salida %d: %s", k, r.salida)}
			}
		}
		out[p.Nombre] = r
	}
	return out, nil
}

// ultimas: las últimas n líneas no vacías, unidas en una.
func ultimas(s string, n int) string {
	var ls []string
	for _, l := range strings.Split(s, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			ls = append(ls, l)
		}
	}
	if len(ls) > n {
		ls = ls[len(ls)-n:]
	}
	return strings.Join(ls, " · ")
}
