package wiki

import (
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"
)

// Glosario (S002): docs/glosario.md define cada término técnico como «## término»,
// con definición, línea «Ejemplo:» y, opcional, «Variantes: a, b» (plurales, formas).
// En cada página, la primera aparición en prosa de un término debe enlazar a su entrada:
// [término](…/glosario.md#término). Términos no registrados no se detectan: criterio humano.

const glosario = "docs/glosario.md"

var reEnlaceTD = regexp.MustCompile(`\[([^\]]*)\]\(([^)\s]+)(?:\s+"[^"]*")?\)`)

type termino struct {
	nombre, ancla string
	re            *regexp.Regexp
}

// terminos extrae las entradas del glosario con sus formas (nombre + variantes).
func terminos(pg *pagina) []termino {
	var ts []termino
	for _, s := range pg.secciones {
		formas := []string{s.titulo}
		for _, l := range s.lineas {
			if v, ok := strings.CutPrefix(l, "Variantes:"); ok {
				for _, f := range strings.Split(v, ",") {
					if f = strings.TrimSpace(f); f != "" {
						formas = append(formas, f)
					}
				}
			}
		}
		sort.Slice(formas, func(i, j int) bool { return len(formas[i]) > len(formas[j]) })
		for i, f := range formas {
			formas[i] = regexp.QuoteMeta(f)
		}
		// borde de palabra: «-» une palabras (pre-commit ≠ commit); «.» delante es extensión (wiki.go ≠ Go)
		re := regexp.MustCompile(`(?i)(?:^|[^\p{L}\p{N}_.-])(` + strings.Join(formas, "|") + `)(?:$|[^\p{L}\p{N}_-])`)
		ts = append(ts, termino{nombre: s.titulo, ancla: slug(s.titulo), re: re})
	}
	return ts
}

func revisarGlosario(pg *pagina) []string {
	var ps []string
	for _, s := range pg.secciones {
		definicion, ejemplo := false, false
		for _, l := range s.lineas {
			switch {
			case strings.HasPrefix(l, "Ejemplo:"):
				ejemplo = true
			case strings.HasPrefix(l, "Variantes:"):
			case strings.TrimSpace(l) != "":
				definicion = true
			}
		}
		if !definicion {
			ps = append(ps, fmt.Sprintf("%s: entrada «%s» sin definición", pg.rel, s.titulo))
		}
		if !ejemplo {
			ps = append(ps, fmt.Sprintf("%s: entrada «%s» sin línea «Ejemplo:»", pg.rel, s.titulo))
		}
	}
	return ps
}

// revisarTerminos exige que la primera aparición en prosa de cada término enlace a su entrada.
func revisarTerminos(pg *pagina, ts []termino) []string {
	var ps []string
	for _, f := range faltantes(pg, ts) {
		rel := relDesde(pg.rel, glosario)
		ps = append(ps, fmt.Sprintf("%s: primera aparición de «%s» sin enlace al glosario (escribe [%s](%s#%s))",
			pg.rel, f.t.nombre, f.forma, rel, f.t.ancla))
	}
	return ps
}

// falta: término cuya primera aparición en prosa no enlaza al glosario; forma = el texto tal cual aparece.
type falta struct {
	t     termino
	forma string
}

// faltantes: la detección que comparten el gate (revisarTerminos) y el arreglo (Enlazar).
func faltantes(pg *pagina, ts []termino) []falta {
	var fs []falta
	visto := make([]bool, len(ts))
	for _, l := range pg.prosa {
		if l == "" {
			continue
		}
		enlaces := reEnlaceTD.FindAllStringSubmatchIndex(l, -1)
		for i, t := range ts {
			if visto[i] {
				continue
			}
			for _, m := range t.re.FindAllStringSubmatchIndex(l, -1) {
				ini, fin := m[2], m[3]
				enDestino, ok := false, false
				for _, e := range enlaces {
					if ini >= e[4] && fin <= e[5] {
						enDestino = true // dentro de la URL de un enlace: no es prosa
					}
					if ini >= e[2] && fin <= e[3] {
						d, f := resolver(pg.rel, l[e[4]:e[5]])
						ok = d == glosario && f == t.ancla
					}
				}
				if enDestino {
					continue
				}
				visto[i] = true
				if !ok {
					fs = append(fs, falta{t, l[ini:fin]})
				}
				break
			}
		}
	}
	return fs
}

// relDesde da la ruta relativa de destino vista desde la carpeta de la página.
func relDesde(desde, destino string) string {
	dir := strings.Split(path.Dir(desde), "/")
	dst := strings.Split(destino, "/")
	i := 0
	for i < len(dir) && i < len(dst)-1 && dir[i] == dst[i] {
		i++
	}
	return strings.Repeat("../", len(dir)-i) + strings.Join(dst[i:], "/")
}
