// Package herencia (E003/F001/S002) declara qué control de la central viaja a los hijos.
//
//	gobierno/config/herencia.json   una fila por control: control, destino (todos|central), motivo
//
// Controles reales: cada gancho de claude_settings.json (gancho:<orden>), cada archivo de .githooks
// (git:<archivo>) y cada chequeo de verify (gate:<chequeo>). Verify falla si no coinciden con las filas.
package herencia

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

// Ruta del archivo de herencia, relativa a la raíz.
const Ruta = "gobierno/config/herencia.json"

// rutaSettings: plantilla de ganchos de la central.
const rutaSettings = "gobierno/config/claude_settings.json"

// Chequeos de `aquas verify`, por nombre. Chequeo nuevo en verify → agregarlo aquí y su fila en herencia.json.
var Chequeos = []string{"calidad-go", "aislamiento", "derivados", "wiki", "backlog", "herencia", "guardia"}

// Destinos admitidos. Los perfiles por tipo de hijo entran cuando exista el primero.
var destinos = map[string]bool{"todos": true, "central": true}

// Fila de herencia.json.
type Fila struct {
	Control string `json:"control"`
	Destino string `json:"destino"`
	Motivo  string `json:"motivo"`
}

type archivo struct {
	Version   int    `json:"version"`
	Controles []Fila `json:"controles"`
}

// Controles devuelve la lista real de controles de la central: ganchos en su orden, .githooks y Chequeos.
func Controles(raiz string) ([]string, error) {
	b, err := os.ReadFile(filepath.Join(raiz, filepath.FromSlash(rutaSettings)))
	if err != nil {
		return nil, err
	}
	var s struct {
		Hooks []struct {
			Orden string `json:"orden"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, fmt.Errorf("%s: %w", rutaSettings, err)
	}
	var cs []string
	vistos := map[string]bool{}
	for _, h := range s.Hooks {
		if c := "gancho:" + h.Orden; !vistos[c] {
			vistos[c] = true
			cs = append(cs, c)
		}
	}
	es, err := os.ReadDir(filepath.Join(raiz, ".githooks"))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	var git []string
	for _, e := range es {
		if !e.IsDir() {
			git = append(git, "git:"+e.Name())
		}
	}
	sort.Strings(git)
	cs = append(cs, git...)
	for _, c := range Chequeos {
		cs = append(cs, "gate:"+c)
	}
	return cs, nil
}

// Leer devuelve las filas de herencia.json.
func Leer(raiz string) ([]Fila, error) {
	b, err := os.ReadFile(filepath.Join(raiz, filepath.FromSlash(Ruta)))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("falta %s: una fila por control (control, destino todos|central, motivo)", Ruta)
	}
	if err != nil {
		return nil, err
	}
	var a archivo
	if err := json.Unmarshal(b, &a); err != nil {
		return nil, fmt.Errorf("%s ilegible: %w", Ruta, err)
	}
	if a.Version != 1 {
		return nil, fmt.Errorf("%s: versión %d desconocida (se espera 1)", Ruta, a.Version)
	}
	return a.Controles, nil
}

// Validar devuelve los problemas de herencia.json frente a los controles reales; vacío = al día.
func Validar(raiz string) []string {
	controles, err := Controles(raiz)
	if err != nil {
		return []string{err.Error()}
	}
	fs, err := Leer(raiz)
	if err != nil {
		return []string{err.Error()}
	}
	existe := map[string]bool{}
	for _, c := range controles {
		existe[c] = true
	}
	var ps []string
	vistas := map[string]bool{}
	for _, f := range fs {
		switch {
		case vistas[f.Control]:
			ps = append(ps, fmt.Sprintf("%s: fila repetida", f.Control))
			continue
		case !existe[f.Control]:
			ps = append(ps, fmt.Sprintf("%s: el control no existe en la central; borra su fila", f.Control))
		}
		vistas[f.Control] = true
		if !destinos[f.Destino] {
			ps = append(ps, fmt.Sprintf("%s: destino «%s» desconocido (usa todos o central)", f.Control, f.Destino))
		}
		if strings.TrimSpace(f.Motivo) == "" {
			ps = append(ps, fmt.Sprintf("%s: falta el motivo", f.Control))
		}
	}
	for _, c := range controles {
		if !vistas[c] {
			ps = append(ps, fmt.Sprintf("%s sin fila en %s: agrega destino y motivo", c, Ruta))
		}
	}
	return ps
}

// ParaHijo (E003/F001/S005) devuelve los chequeos de verify que corren en un hijo: los que su copia
// sellada de herencia.json marca «todos». Un chequeo sin fila no corre y se informa como problema.
func ParaHijo(hijo string) (corren, problemas []string) {
	filas, err := Leer(filepath.Join(hijo, ".aquas", "central"))
	if err != nil {
		return nil, []string{err.Error()}
	}
	destino := map[string]string{}
	for _, f := range filas {
		destino[f.Control] = f.Destino
	}
	for _, c := range Chequeos {
		d, hay := destino["gate:"+c]
		switch {
		case !hay:
			problemas = append(problemas, fmt.Sprintf("gate:%s sin fila en la herencia sellada del hijo: no corre; actualiza la central (aquas update)", c))
		case d == "todos":
			corren = append(corren, c)
		}
	}
	return corren, problemas
}
