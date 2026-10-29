package wiki

// E001/F012/S001 (D000031): la página de un slice activo trae la tabla «Casos borde» con las 7 categorías.
// Que estén es determinista; que estén bien pensadas lo juzga el auditor.

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/backlog"
)

// CategoriasCasosBorde: nombres exactos de las filas, en el orden de la regla 04_CASOS_BORDE.md.
var CategoriasCasosBorde = []string{
	"cada tipo de ficha",
	"cada estado de origen",
	"padre e hijo, activo y archivado",
	"entrada ausente, vacía, mal escrita o duplicada",
	"choque con otro control",
	"quién habla: humano, agente o herramienta",
	"fallo a mitad de camino",
}

const (
	marcaCasosBorde      = "**Casos borde.**"
	encabezadoCasosBorde = "Categoría | Aplica | Prueba o motivo"
)

var reEnlaceMd = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)

// revisarCasosBorde: problemas de la tabla «Casos borde» de la página rel; nada si no es un slice con ficha activa.
// enCodigo: func Test del repo; una prueba de la tabla debe existir o estar prevista, como las del contrato.
func revisarCasosBorde(raiz, rel string, enCodigo map[string]string) []string {
	m := reS.FindStringSubmatch(rel)
	if m == nil {
		return nil
	}
	if _, archivada, err := backlog.Ver(raiz, m[1]+"/"+m[2]+"/"+m[3]); err != nil || archivada {
		return nil
	}
	c, err := os.ReadFile(filepath.Join(raiz, filepath.FromSlash(rel)))
	if err != nil {
		return []string{fmt.Sprintf("%s: casos borde: no se pudo leer: %v", rel, err)}
	}
	encabezado, filas, hay := filasCasosBorde(string(c))
	if !hay {
		return []string{fmt.Sprintf("%s: falta la tabla «Casos borde» (párrafo %s y una fila por cada una de las %d categorías; regla 04_CASOS_BORDE.md)", rel, marcaCasosBorde, len(CategoriasCasosBorde))}
	}
	var ps []string
	mal := func(formato string, a ...any) { ps = append(ps, rel+": casos borde: "+fmt.Sprintf(formato, a...)) }
	if encabezado != encabezadoCasosBorde {
		mal("el encabezado debe ser «%s»; dice «%s»", encabezadoCasosBorde, encabezado)
	}
	conocida := map[string]bool{}
	for _, cat := range CategoriasCasosBorde {
		conocida[cat] = true
	}
	vista := map[string]bool{}
	for _, f := range filas {
		cat, aplica, texto := f[0], f[1], f[2]
		switch {
		case !conocida[cat]:
			mal("categoría desconocida «%s»", cat)
			continue
		case vista[cat]:
			mal("categoría repetida «%s»", cat)
			continue
		}
		vista[cat] = true
		nombres := rePruebaNombrada.FindAllStringSubmatch(texto, -1)
		switch aplica {
		case "sí":
			if len(nombres) == 0 {
				mal("«%s»: con sí nombra una prueba entre acentos graves", cat)
			}
		case "no":
			if strings.Trim(texto, " —-") == "" { // «—» es la marca de campo vacío

				mal("«%s»: con no explica por qué", cat)
			}
		default:
			mal("«%s»: Aplica debe ser sí o no; dice «%s»", cat, aplica)
		}
		for _, n := range nombres {
			if enCodigo[n[1]] == "" && !prevista(raiz, rel, n[1]) {
				ps = append(ps, fmt.Sprintf("%s: prueba nombrada no existe en el código: %s", rel, n[1]))
			}
		}
	}
	for _, cat := range CategoriasCasosBorde {
		if !vista[cat] {
			mal("falta la categoría «%s»", cat)
		}
	}
	return ps
}

// filasCasosBorde: encabezado y celdas [categoría, aplica, texto] de la tabla que sigue a la marca; false si no hay marca.
// La marca cuenta solo fuera de bloques de código y de comentarios HTML: ahí no se ve al leer la página.
// Los enlaces Markdown se reducen a su texto: quien escribe la tabla puede enlazar una palabra de la categoría.
func filasCasosBorde(contenido string) (string, [][3]string, bool) {
	lineas := strings.Split(strings.ReplaceAll(contenido, "\r\n", "\n"), "\n")
	i := lineaDeMarca(lineas)
	if i < 0 {
		return "", nil, false
	}
	for i++; i < len(lineas) && strings.TrimSpace(lineas[i]) == ""; i++ {
	}
	var encabezado string
	var filas [][3]string
	for k := 0; i < len(lineas) && strings.HasPrefix(strings.TrimSpace(lineas[i]), "|"); i, k = i+1, k+1 {
		celdas := celdasTabla(reEnlaceMd.ReplaceAllString(lineas[i], "$1"))
		switch k {
		case 0:
			encabezado = strings.Join(celdas, " | ")
			continue
		case 1: // separador
			continue
		}
		for len(celdas) < 3 {
			celdas = append(celdas, "")
		}
		filas = append(filas, [3]string{celdas[0], celdas[1], strings.Join(celdas[2:], " | ")})
	}
	return encabezado, filas, true
}

// celdasTabla parte una fila Markdown por «|» sin escapar y recorta cada celda.
func celdasTabla(linea string) []string {
	linea = strings.TrimSpace(linea)
	linea = strings.TrimSuffix(strings.TrimPrefix(linea, "|"), "|")
	var celdas []string
	var b strings.Builder
	for j := 0; j < len(linea); j++ {
		switch {
		case linea[j] == '\\' && j+1 < len(linea) && linea[j+1] == '|':
			b.WriteByte('|')
			j++
		case linea[j] == '|':
			celdas = append(celdas, strings.TrimSpace(b.String()))
			b.Reset()
		default:
			b.WriteByte(linea[j])
		}
	}
	return append(celdas, strings.TrimSpace(b.String()))
}

// lineaDeMarca: índice de la marca fuera de bloques de código y comentarios HTML; -1 si no está.
func lineaDeMarca(lineas []string) int {
	cerca, comentario := "", false
	for i, l := range lineas {
		l = strings.TrimSpace(l)
		switch {
		case comentario:
			comentario = !strings.Contains(l, "-->")
		case cerca != "":
			if strings.HasPrefix(l, cerca) {
				cerca = ""
			}
		case strings.HasPrefix(l, "```") || strings.HasPrefix(l, "~~~"):
			cerca = l[:3]
		case strings.HasPrefix(l, "<!--"):
			comentario = !strings.Contains(l, "-->")
		case l == marcaCasosBorde:
			return i
		}
	}
	return -1
}
