package backlog

import (
	"regexp"
	"sort"
	"strings"
)

// codigoTocado: .go de producción en cmd/ e internal/ que tocaron los commits del slice y, si sinCommit, los cambios pendientes (E001/F008/S005).
func codigoTocado(raiz, id string, sinCommit bool) ([]string, error) {
	log, err := gitLineas(raiz, "log", "--format=", "--name-only", "-E", "--grep="+regexp.QuoteMeta(id)+"([^0-9]|$)")
	if err != nil {
		return nil, err
	}
	var st []string
	if sinCommit {
		if st, err = gitLineas(raiz, "status", "--porcelain", "-uall"); err != nil {
			return nil, err
		}
	}
	for _, l := range st {
		_, l, _ = strings.Cut(l, " ") // gitLineas recorta: « M x» llega como «M x»
		if _, nuevo, ok := strings.Cut(l, " -> "); ok {
			l = nuevo
		}
		log = append(log, strings.Trim(l, `"`))
	}
	return soloProduccion(log), nil
}

// soloProduccion: .go de cmd/ e internal/ sin pruebas y archivos de configuración (settings, .githooks, gobierno/config sin sellos), sin repetir y ordenados.
func soloProduccion(log []string) []string {
	vistos := map[string]bool{}
	var xs []string
	for _, p := range log {
		goProd := (strings.HasPrefix(p, "cmd/") || strings.HasPrefix(p, "internal/")) && strings.HasSuffix(p, ".go") && !strings.HasSuffix(p, "_test.go")
		config := p == ".claude/settings.json" || strings.HasPrefix(p, ".claude/hooks/") || strings.HasPrefix(p, ".githooks/") || // D000022: la configuración también es código tocado
			(strings.HasPrefix(p, "gobierno/config/") && p != "gobierno/config/sellos.json")
		prod := goProd || config
		if prod && !vistos[p] {
			vistos[p] = true
			xs = append(xs, p)
		}
	}
	sort.Strings(xs)
	return xs
}

// Codigo recalcula y guarda el código tocado de un slice desde los commits que nombran su ID y los commits dados.
func Codigo(raiz, id string, commits []string) (Ficha, error) {
	f, archivada, err := Ver(raiz, id)
	if err != nil {
		return Ficha{}, err
	}
	if f.Codigo, err = codigoTocado(raiz, id, false); err != nil {
		return Ficha{}, err
	}
	for _, c := range commits { // D000022: commits que no nombran el ID
		xs, err := gitLineas(raiz, "show", "--format=", "--name-only", c)
		if err != nil {
			return Ficha{}, err
		}
		f.Codigo = soloProduccion(append(f.Codigo, xs...))
	}
	return f, escribir(raiz, dirDe(archivada), f)
}
