package wiki

// Contrato y evidencia (E001/F001/S004): las páginas E/F/S llevan la sección
// «Contrato y evidencia» antes de Navegación y migas «Inicio › E › F › S» bajo el
// título. Trazabilidad: toda prueba nombrada (`TestX`) en esa sección de un slice
// existe en algún *_test.go del repo, y toda `func TestX` del repo la nombra algún slice.

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const secContrato = "Contrato y evidencia"

var (
	rePruebaNombrada = regexp.MustCompile("`(Test[A-Za-z0-9_]+)`")
	reFuncTest       = regexp.MustCompile(`(?m)^func (Test[A-Za-z0-9_]+)\(`)
)

// seccionesDe: orden fijo de secciones según el nivel de la página.
func seccionesDe(n nivel) []string {
	if n != nE && n != nF && n != nS {
		return Secciones
	}
	s := append([]string{}, Secciones[:len(Secciones)-1]...)
	return append(s, secContrato, secNav)
}

// migasEsperadas: migas «Inicio › E001 › F001», cada nivel con enlace relativo a la página.
func migasEsperadas(pg *pagina) string {
	var ids []string
	for _, re := range []*regexp.Regexp{reS, reF, reE} {
		if m := re.FindStringSubmatch(pg.rel); m != nil {
			ids = m[1:]
			break
		}
	}
	if ids == nil {
		return ""
	}
	// Profundidad de la carpeta de la página bajo docs/: E=1, F=2, S=2 (vive en la carpeta de F).
	prof := min(len(ids), 2)
	partes := []string{"[Inicio](" + strings.Repeat("../", prof) + "README.md)"}
	for k := 1; k < len(ids); k++ { // ancestro de nivel k vive a profundidad k
		partes = append(partes, "["+ids[k-1]+"]("+strings.Repeat("../", prof-k)+"README.md)")
	}
	return strings.Join(append(partes, ids[len(ids)-1]), " › ")
}

// revisarMigas: aviso si las migas de pan no coinciden con la ruta de la página.
func revisarMigas(pg *pagina) []string {
	esperadas := migasEsperadas(pg)
	if esperadas == "" || pg.migas == esperadas {
		return nil
	}
	return []string{fmt.Sprintf("%s: migas ausentes o distintas; la primera línea bajo el título debe ser: %s", pg.rel, esperadas)}
}

// revisarTrazabilidad cruza las pruebas nombradas en los slices con las func Test del repo.
func revisarTrazabilidad(raiz string, paginas map[string]*pagina, orden []string) ([]string, error) {
	nombradas := map[string]bool{}
	var ps []string
	enCodigo, err := pruebasDelRepo(raiz)
	if err != nil {
		return nil, err
	}
	for _, rel := range orden {
		pg := paginas[rel]
		if pg.nivel != nS {
			continue
		}
		for _, n := range pg.pruebas {
			nombradas[n] = true
			if enCodigo[n] == "" {
				ps = append(ps, fmt.Sprintf("%s: prueba nombrada no existe en el código: %s", rel, n))
			}
		}
	}
	var sinDueno []string
	for n, archivo := range enCodigo {
		if !nombradas[n] {
			sinDueno = append(sinDueno, fmt.Sprintf("%s: prueba sin slice dueño: %s (nómbrala en «%s» de su slice)", archivo, n, secContrato))
		}
	}
	sort.Strings(sinDueno)
	return append(ps, sinDueno...), nil
}

// pruebasDelRepo: nombre de cada func Test → archivo; salta carpetas ocultas (.git, .tmp).
func pruebasDelRepo(raiz string) (map[string]string, error) {
	pruebas := map[string]string{}
	err := filepath.WalkDir(raiz, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if p != raiz && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(p, "_test.go") {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		for _, m := range reFuncTest.FindAllStringSubmatch(string(b), -1) {
			pruebas[m[1]] = relativa(raiz, p)
		}
		return nil
	})
	return pruebas, err
}
