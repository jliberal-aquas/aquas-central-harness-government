// Paquete wiki: gate determinista de la wiki pedagógica en docs/.
//
// Árbol: docs/README.md (inicio) → docs/E001/README.md → docs/E001/F001/README.md → docs/E001/F001/S001.md.
// Cada página: secciones fijas en orden, diagrama Mermaid según nivel, ejemplo con ancla EJ-<id>
// encadenado al padre («continúa:»), navegación ↑ Subir / Abajo, enlaces y anclas válidos,
// alcanzable desde el inicio, prosa legible (legibilidad.go) y términos enlazados al glosario (glosario.go).
// Referencia (S003): docs/ref/<nombre>.md documenta lo existente; ancla EJ-REF-<nombre>, cualquier diagrama
// Mermaid, ↑ Subir al inicio, enlazada desde el inicio; sin «continúa:».
package wiki

import (
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// Secciones fijas, en orden, de toda página de la wiki.
var Secciones = []string{
	"En una frase", "Para qué existe", "Analogía", "Diagrama",
	"Ejemplo de punta a punta", "Paso a paso", "Cómo se usa", "Navegación",
}

const (
	secDiagrama = "Diagrama"
	secEjemplo  = "Ejemplo de punta a punta"
	secNav      = "Navegación"
	inicio      = "docs/README.md"
)

type nivel int

const (
	nInicio nivel = iota
	nE
	nF
	nS
	nGlosario
	nRef
)

var (
	reE       = regexp.MustCompile(`^docs/(E\d{3})/README\.md$`)
	reF       = regexp.MustCompile(`^docs/(E\d{3})/(F\d{3})/README\.md$`)
	reS       = regexp.MustCompile(`^docs/(E\d{3})/(F\d{3})/(S\d{3})\.md$`)
	reRef     = regexp.MustCompile(`^docs/ref/([a-z0-9-]+)\.md$`)
	reEnlace  = regexp.MustCompile(`!?\[[^\]]*\]\(([^)\s]+)(?:\s+"[^"]*")?\)`)
	reCodigo  = regexp.MustCompile("`[^`]*`")
	reTitulo  = regexp.MustCompile(`^(#{1,6})\s+(.*?)\s*#*\s*$`)
	reAncla   = regexp.MustCompile(`<a\s+(?:id|name)="([^"]+)"`)
	reEsquema = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*:`)
)

type enlace struct {
	seccion, linea, destino string
}

type seccion struct {
	titulo    string
	contenido bool     // tiene alguna línea no vacía
	diagramas []string // tipo de cada bloque mermaid (primera palabra)
	lineas    []string
}

type pagina struct {
	rel, padre, ej string
	nivel          nivel
	secciones      []seccion
	anclas         map[string]bool
	enlaces        []enlace
	prosa          []string // líneas medibles por legibilidad; "" = corte de párrafo
	migas          string   // primera línea no vacía bajo el título H1
	pruebas        []string // pruebas nombradas en «Contrato y evidencia»
}

// Verificar devuelve los problemas de la wiki bajo raiz/docs; vacío = gate OK.
func Verificar(raiz string) ([]string, error) {
	if _, err := os.Stat(filepath.Join(raiz, "docs", "README.md")); errors.Is(err, fs.ErrNotExist) {
		return []string{"falta docs/README.md (inicio de la wiki)"}, nil
	}
	var problemas []string
	if _, err := os.Stat(filepath.Join(raiz, filepath.FromSlash(glosario))); errors.Is(err, fs.ErrNotExist) {
		problemas = append(problemas, "falta "+glosario+" (glosario de términos técnicos)")
	}
	paginas := map[string]*pagina{}
	var orden []string
	err := filepath.WalkDir(filepath.Join(raiz, "docs"), func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".md") {
			return err
		}
		rel := relativa(raiz, p)
		pg := clasificar(rel)
		if pg == nil {
			problemas = append(problemas, rel+": fuera del árbol (válidas: docs/README.md, docs/glosario.md, docs/E000/README.md, docs/E000/F000/README.md, docs/E000/F000/S000.md, docs/ref/nombre-en-minusculas.md)")
			return nil
		}
		if err := analizar(raiz, pg); err != nil {
			return err
		}
		paginas[rel] = pg
		orden = append(orden, rel)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(orden)

	anclas := map[string]map[string]bool{}
	for rel, pg := range paginas {
		anclas[rel] = pg.anclas
	}
	hijos := map[string][]string{}
	for _, rel := range orden {
		if pg := paginas[rel]; pg.padre != "" {
			hijos[pg.padre] = append(hijos[pg.padre], rel)
		}
	}
	var ts []termino
	if g := paginas[glosario]; g != nil {
		ts = terminos(g)
	}
	for _, rel := range orden {
		pg := paginas[rel]
		if pg.nivel == nGlosario {
			problemas = append(problemas, revisarGlosario(pg)...)
			problemas = append(problemas, revisarEnlaces(raiz, pg, anclas)...)
			problemas = append(problemas, revisarLegibilidad(pg)...)
			continue
		}
		problemas = append(problemas, revisarSecciones(pg)...)
		problemas = append(problemas, revisarMigas(pg)...)
		problemas = append(problemas, revisarDiagramas(pg)...)
		problemas = append(problemas, revisarEjemplo(pg, paginas)...)
		problemas = append(problemas, revisarNavegacion(pg, hijos[rel])...)
		problemas = append(problemas, revisarEnlaces(raiz, pg, anclas)...)
		problemas = append(problemas, revisarLegibilidad(pg)...)
		problemas = append(problemas, revisarTerminos(pg, ts)...)
	}
	problemas = append(problemas, revisarAlcance(paginas, orden)...)
	traza, err := revisarTrazabilidad(raiz, paginas, orden)
	if err != nil {
		return nil, err
	}
	problemas = append(problemas, traza...)
	return problemas, nil
}

func relativa(raiz, p string) string {
	r, _ := filepath.Rel(raiz, p)
	return filepath.ToSlash(r)
}

// clasificar asigna nivel, padre e id de ejemplo según la ruta; nil = fuera del árbol.
func clasificar(rel string) *pagina {
	if rel == inicio {
		return &pagina{rel: rel, nivel: nInicio, ej: "EJ-INICIO"}
	}
	if rel == glosario {
		return &pagina{rel: rel, nivel: nGlosario}
	}
	if m := reRef.FindStringSubmatch(rel); m != nil {
		return &pagina{rel: rel, nivel: nRef, padre: inicio, ej: "EJ-REF-" + m[1]}
	}
	if m := reE.FindStringSubmatch(rel); m != nil {
		return &pagina{rel: rel, nivel: nE, padre: inicio, ej: "EJ-" + m[1]}
	}
	if m := reF.FindStringSubmatch(rel); m != nil {
		return &pagina{rel: rel, nivel: nF, padre: "docs/" + m[1] + "/README.md", ej: "EJ-" + m[1] + "-" + m[2]}
	}
	if m := reS.FindStringSubmatch(rel); m != nil {
		return &pagina{rel: rel, nivel: nS, padre: "docs/" + m[1] + "/" + m[2] + "/README.md", ej: "EJ-" + m[1] + "-" + m[2] + "-" + m[3]}
	}
	return nil
}

// analizar lee la página: secciones H2, diagramas, anclas (id HTML y títulos) y enlaces fuera de código.
func analizar(raiz string, pg *pagina) error {
	b, err := os.ReadFile(filepath.Join(raiz, filepath.FromSlash(pg.rel)))
	if err != nil {
		return err
	}
	pg.anclas = map[string]bool{}
	slugs := map[string]int{}
	actual := -1
	cerca, mermaid, primera := "", false, false
	vistaPrimera := false
	for _, l := range strings.Split(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n") {
		t := strings.TrimSpace(l)
		if cerca != "" { // dentro de bloque de código
			if strings.HasPrefix(t, cerca) && strings.Trim(t, cerca[:1]) == "" {
				cerca = ""
			} else if mermaid && primera && t != "" && actual >= 0 {
				pg.secciones[actual].diagramas = append(pg.secciones[actual].diagramas, strings.Fields(t)[0])
				primera = false
			}
			if actual >= 0 && t != "" {
				pg.secciones[actual].contenido = true
			}
			continue
		}
		if strings.HasPrefix(t, "```") || strings.HasPrefix(t, "~~~") {
			pg.prosa = append(pg.prosa, "")
			cerca = t[:3]
			mermaid = strings.TrimSpace(strings.Trim(t, "`~")) == "mermaid"
			primera = true
			if actual >= 0 {
				pg.secciones[actual].contenido = true
			}
			continue
		}
		if m := reTitulo.FindStringSubmatch(t); m != nil {
			s := slug(m[2])
			if n := slugs[s]; n > 0 {
				pg.anclas[fmt.Sprintf("%s-%d", s, n)] = true
			} else {
				pg.anclas[s] = true
			}
			slugs[s]++
			if len(m[1]) == 2 {
				pg.secciones = append(pg.secciones, seccion{titulo: m[2]})
				actual = len(pg.secciones) - 1
			}
			pg.prosa = append(pg.prosa, "")
			if len(m[1]) == 2 {
				continue
			}
		}
		for _, a := range reAncla.FindAllStringSubmatch(t, -1) {
			pg.anclas[a[1]] = true
		}
		if actual < 0 && t != "" && !reTitulo.MatchString(t) && !vistaPrimera {
			pg.migas, vistaPrimera = t, true
		}
		if actual >= 0 && pg.secciones[actual].titulo == secContrato {
			for _, m := range rePruebaNombrada.FindAllStringSubmatch(t, -1) {
				pg.pruebas = append(pg.pruebas, m[1])
			}
		}
		limpia := reCodigo.ReplaceAllString(t, "")
		titulo := ""
		if actual >= 0 {
			titulo = pg.secciones[actual].titulo
			if t != "" {
				pg.secciones[actual].contenido = true
				pg.secciones[actual].lineas = append(pg.secciones[actual].lineas, limpia)
			}
		}
		if reTitulo.MatchString(t) || titulo == secNav || strings.HasPrefix(t, "|") || strings.HasPrefix(t, "continúa:") {
			pg.prosa = append(pg.prosa, "")
		} else {
			pg.prosa = append(pg.prosa, limpia)
		}
		for _, e := range reEnlace.FindAllStringSubmatch(limpia, -1) {
			pg.enlaces = append(pg.enlaces, enlace{seccion: titulo, linea: limpia, destino: e[1]})
		}
	}
	return nil
}

// slug al estilo GitHub/VS Code: minúsculas, letras/dígitos/-/_ se conservan, espacio → guion.
func slug(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_':
			b.WriteRune(r)
		case r == ' ':
			b.WriteRune('-')
		}
	}
	return b.String()
}

func (pg *pagina) seccion(titulo string) *seccion {
	for i := range pg.secciones {
		if pg.secciones[i].titulo == titulo {
			return &pg.secciones[i]
		}
	}
	return nil
}

func revisarSecciones(pg *pagina) []string {
	var ps []string
	ultimo, desorden := -1, false
	orden := seccionesDe(pg.nivel)
	for _, req := range orden {
		idx := -1
		for i, s := range pg.secciones {
			if s.titulo == req {
				idx = i
				break
			}
		}
		switch {
		case idx < 0:
			ps = append(ps, fmt.Sprintf("%s: falta sección «%s»", pg.rel, req))
			continue
		case !pg.secciones[idx].contenido:
			ps = append(ps, fmt.Sprintf("%s: sección vacía «%s»", pg.rel, req))
		}
		if idx < ultimo {
			desorden = true
		}
		ultimo = idx
	}
	if desorden {
		ps = append(ps, fmt.Sprintf("%s: secciones fuera de orden (orden fijo: %s)", pg.rel, strings.Join(orden, " · ")))
	}
	return ps
}

func revisarDiagramas(pg *pagina) []string {
	s := pg.seccion(secDiagrama)
	if s == nil {
		return nil // ya reportado como sección faltante
	}
	tiene := func(tipos ...string) bool {
		for _, d := range s.diagramas {
			for _, t := range tipos {
				if d == t {
					return true
				}
			}
		}
		return false
	}
	var ps []string
	if pg.nivel == nRef {
		if len(s.diagramas) == 0 {
			ps = append(ps, pg.rel+": falta diagrama Mermaid en «Diagrama» (cualquier tipo)")
		}
		return ps
	}
	if pg.nivel != nS && !tiene("flowchart", "graph") {
		ps = append(ps, pg.rel+": falta diagrama flowchart en «Diagrama» (flujo del nivel)")
	}
	if pg.nivel == nF && !tiene("stateDiagram", "stateDiagram-v2") {
		ps = append(ps, pg.rel+": falta diagrama stateDiagram en «Diagrama» (estados de la feature)")
	}
	if pg.nivel == nS && !tiene("sequenceDiagram", "erDiagram", "classDiagram") {
		ps = append(ps, pg.rel+": falta diagrama de datos o secuencia en «Diagrama» (sequenceDiagram, erDiagram o classDiagram)")
	}
	return ps
}

func revisarEjemplo(pg *pagina, paginas map[string]*pagina) []string {
	s := pg.seccion(secEjemplo)
	if s == nil {
		return nil
	}
	var ps []string
	if !strings.Contains(strings.Join(s.lineas, "\n"), `id="`+pg.ej+`"`) {
		ps = append(ps, fmt.Sprintf(`%s: falta ancla %s en «%s» (escribe <a id="%s"></a>)`, pg.rel, pg.ej, secEjemplo, pg.ej))
	}
	if pg.nivel != nF && pg.nivel != nS {
		return ps
	}
	ejPadre := "EJ-?"
	if p := paginas[pg.padre]; p != nil {
		ejPadre = p.ej
	}
	for _, l := range s.lineas {
		if !strings.HasPrefix(l, "continúa:") {
			continue
		}
		m := reEnlace.FindStringSubmatch(l)
		if m != nil {
			destino, frag := resolver(pg.rel, m[1])
			if destino == pg.padre && frag == ejPadre {
				return ps
			}
		}
		return append(ps, fmt.Sprintf("%s: continúa debe apuntar a %s en %s", pg.rel, ejPadre, pg.padre))
	}
	return append(ps, fmt.Sprintf("%s: falta «continúa:» hacia %s (%s)", pg.rel, ejPadre, pg.padre))
}

func revisarNavegacion(pg *pagina, hijos []string) []string {
	var ps []string
	enlazados := map[string]bool{}
	subir := false
	for _, e := range pg.enlaces {
		if e.seccion != secNav {
			continue
		}
		destino, _ := resolver(pg.rel, e.destino)
		enlazados[destino] = true
		if destino == pg.padre && strings.Contains(e.linea, "↑ Subir") {
			subir = true
		}
	}
	if pg.padre != "" && !subir {
		ps = append(ps, fmt.Sprintf("%s: Navegación sin ↑ Subir hacia %s", pg.rel, pg.padre))
	}
	for _, h := range hijos {
		if !enlazados[h] {
			ps = append(ps, fmt.Sprintf("%s: Navegación no enlaza %s", pg.rel, h))
		}
	}
	return ps
}

func revisarEnlaces(raiz string, pg *pagina, anclas map[string]map[string]bool) []string {
	var ps []string
	for _, e := range pg.enlaces {
		if reEsquema.MatchString(e.destino) {
			continue // externo (https:, mailto:, …)
		}
		destino, frag := resolver(pg.rel, e.destino)
		info, err := os.Stat(filepath.Join(raiz, filepath.FromSlash(destino)))
		if err != nil {
			ps = append(ps, fmt.Sprintf("%s: enlace roto → %s", pg.rel, e.destino))
			continue
		}
		if frag == "" || info.IsDir() || !strings.HasSuffix(destino, ".md") {
			continue
		}
		a, ok := anclas[destino]
		if !ok { // archivo fuera de la wiki: extraer anclas bajo demanda
			ext := &pagina{rel: destino}
			if err := analizar(raiz, ext); err == nil {
				a = ext.anclas
			}
			anclas[destino] = a
		}
		if !a[frag] {
			ps = append(ps, fmt.Sprintf("%s: ancla inexistente → %s", pg.rel, e.destino))
		}
	}
	return ps
}

// revisarAlcance: toda página del árbol debe alcanzarse desde docs/README.md siguiendo enlaces.
func revisarAlcance(paginas map[string]*pagina, orden []string) []string {
	visto := map[string]bool{inicio: true}
	cola := []string{inicio}
	for len(cola) > 0 {
		pg := paginas[cola[0]]
		cola = cola[1:]
		if pg == nil {
			continue
		}
		for _, e := range pg.enlaces {
			d, _ := resolver(pg.rel, e.destino)
			if paginas[d] != nil && !visto[d] {
				visto[d] = true
				cola = append(cola, d)
			}
		}
	}
	var ps []string
	for _, rel := range orden {
		if !visto[rel] {
			ps = append(ps, rel+": inalcanzable desde "+inicio)
		}
	}
	return ps
}

// resolver convierte un destino relativo a la página en ruta relativa a la raíz + fragmento.
func resolver(desde, destino string) (string, string) {
	p, frag, _ := strings.Cut(destino, "#")
	if u, err := url.PathUnescape(p); err == nil {
		p = u
	}
	if u, err := url.PathUnescape(frag); err == nil {
		frag = u
	}
	if p == "" {
		return desde, frag
	}
	return path.Clean(path.Join(path.Dir(desde), p)), frag
}
