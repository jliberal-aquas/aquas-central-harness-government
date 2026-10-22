package wiki

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var reCodigoEnLinea = regexp.MustCompile("`[^`]*`")

// Enlazar pone el enlace al glosario en la primera aparición en prosa de cada término (E000/F004/S003).
// Usa la misma detección que el gate (faltantes): lo que arregla es justo lo que el gate pide.
// No toca bloques de código, código en línea, títulos, tablas, comentarios, navegación ni enlaces.
func Enlazar(raiz, rel string) (int, error) {
	g := &pagina{rel: glosario}
	if err := analizar(raiz, g); err != nil {
		return 0, err
	}
	ts := terminos(g)
	hechos := map[string]bool{}
	for n := 0; ; n++ {
		pg := &pagina{rel: rel}
		if err := analizar(raiz, pg); err != nil {
			return n, err
		}
		fs := faltantes(pg, ts)
		if len(fs) == 0 {
			return n, nil
		}
		f := fs[0]
		if hechos[f.t.nombre] { // ya lo envolvimos y sigue faltando: enlace equivocado previo
			return n, fmt.Errorf("%s: «%s» ya tiene un enlace que no va al glosario; corrígelo a mano", rel, f.forma)
		}
		hechos[f.t.nombre] = true
		ok, err := envolver(raiz, rel, f)
		if err != nil {
			return n, err
		}
		if !ok {
			return n, fmt.Errorf("%s: no hallé «%s» en prosa para enlazar", rel, f.forma)
		}
	}
}

// envolver enlaza la primera aparición en prosa del término en el archivo.
func envolver(raiz, rel string, f falta) (bool, error) {
	ruta := filepath.Join(raiz, filepath.FromSlash(rel))
	b, err := os.ReadFile(ruta)
	if err != nil {
		return false, err
	}
	lineas := strings.Split(string(b), "\n")
	cerca, seccion := "", ""
	for i, l := range lineas {
		t := strings.TrimSpace(l)
		if cerca != "" {
			if strings.HasPrefix(t, cerca) && strings.Trim(t, cerca[:1]) == "" {
				cerca = ""
			}
			continue
		}
		if strings.HasPrefix(t, "```") || strings.HasPrefix(t, "~~~") {
			cerca = t[:3]
			continue
		}
		if m := reTitulo.FindStringSubmatch(t); m != nil {
			if len(m[1]) == 2 {
				seccion = m[2]
			}
			continue
		}
		if seccion == secNav || strings.HasPrefix(t, "|") || strings.HasPrefix(t, "<!--") || strings.HasPrefix(t, "continúa:") {
			continue
		}
		var tapado [][]int // código en línea y enlaces enteros: no son prosa enlazable
		tapado = append(tapado, reCodigoEnLinea.FindAllStringIndex(l, -1)...)
		tapado = append(tapado, reEnlaceTD.FindAllStringIndex(l, -1)...)
		for _, m := range f.t.re.FindAllStringSubmatchIndex(l, -1) {
			ini, fin := m[2], m[3]
			libre := true
			for _, x := range tapado {
				if ini < x[1] && fin > x[0] {
					libre = false
				}
			}
			if !libre {
				continue
			}
			lineas[i] = l[:ini] + "[" + l[ini:fin] + "](" + relDesde(rel, glosario) + "#" + f.t.ancla + ")" + l[fin:]
			return true, os.WriteFile(ruta, []byte(strings.Join(lineas, "\n")), 0o644)
		}
	}
	return false, nil
}
