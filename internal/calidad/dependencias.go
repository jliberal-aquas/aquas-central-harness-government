package calidad

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

// RevisarDependencias (E001/F013/S002) compara los imports internos de dir, de código y de prueba,
// con la tabla JSON {paquete: [imports internos permitidos]}; devuelve un hallazgo por violación.
// Tabla ausente o ilegible, o `go list` caído = error; nunca cuenta como limpio.
func RevisarDependencias(dir, tabla string) ([]string, error) {
	b, err := os.ReadFile(tabla)
	if err != nil {
		return nil, fmt.Errorf("tabla de dependencias ilegible: %w", err)
	}
	var permitidas map[string][]string
	if err := json.Unmarshal(b, &permitidas); err != nil {
		return nil, fmt.Errorf("tabla de dependencias inválida %s: %w", tabla, err)
	}
	cmd := exec.Command("go", "list", "-json", "./...")
	cmd.Dir = dir
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("go list falló: %w: %s%s", err, out.String(), errb.String())
	}
	type paquete struct {
		ImportPath                         string
		Imports, TestImports, XTestImports []string
		Module                             struct{ Path string }
	}
	reales := map[string]map[string]bool{}
	dec := json.NewDecoder(&out)
	for dec.More() {
		var p paquete
		if err := dec.Decode(&p); err != nil {
			return nil, fmt.Errorf("salida de go list ilegible: %w", err)
		}
		mod := p.Module.Path
		clave := rel(p.ImportPath, mod)
		imps := reales[clave]
		if imps == nil {
			imps = map[string]bool{}
			reales[clave] = imps
		}
		for _, lista := range [][]string{p.Imports, p.TestImports, p.XTestImports} {
			for _, i := range lista {
				if strings.HasPrefix(i, mod+"/") && i != p.ImportPath {
					imps[rel(i, mod)] = true
				}
			}
		}
	}
	var hs []string
	for pkg, imps := range reales {
		fila, ok := permitidas[pkg]
		if !ok {
			hs = append(hs, pkg+" sin fila en la tabla")
			continue
		}
		ok2 := map[string]bool{}
		for _, f := range fila {
			ok2[f] = true
		}
		for i := range imps {
			if !ok2[i] {
				hs = append(hs, pkg+" → "+i+" no está en la tabla")
			}
		}
	}
	for fila := range permitidas {
		if _, ok := reales[fila]; !ok {
			hs = append(hs, fila+" no existe")
		}
	}
	sort.Strings(hs)
	return hs, nil
}

// rel devuelve la ruta de import relativa al módulo ("." para la raíz).
func rel(imp, mod string) string {
	if imp == mod {
		return "."
	}
	return strings.TrimPrefix(imp, mod+"/")
}
