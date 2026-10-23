package backlog

import (
	"regexp"
	"sort"
	"strings"
)

// codigoTocado: .go de producción en cmd/ e internal/ que tocaron los commits del slice y, si sinCommit, los cambios pendientes (E001/F008/S005).
func codigoTocado(raiz, id string, sinCommit bool) ([]string, error) {
	log, err := gitLineas(raiz, "log", "--format=", "--name-only", "-E", "--grep=^"+regexp.QuoteMeta(id)+"[:/]")
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
	vistos := map[string]bool{}
	var xs []string
	for _, p := range log {
		prod := (strings.HasPrefix(p, "cmd/") || strings.HasPrefix(p, "internal/")) && strings.HasSuffix(p, ".go") && !strings.HasSuffix(p, "_test.go")
		if prod && !vistos[p] {
			vistos[p] = true
			xs = append(xs, p)
		}
	}
	sort.Strings(xs)
	return xs, nil
}

// Codigo recalcula y guarda el código tocado de un slice desde sus commits; sirve para slices ya cerrados.
func Codigo(raiz, id string) (Ficha, error) {
	f, archivada, err := Ver(raiz, id)
	if err != nil {
		return Ficha{}, err
	}
	if f.Codigo, err = codigoTocado(raiz, id, false); err != nil {
		return Ficha{}, err
	}
	return f, escribir(raiz, dirDe(archivada), f)
}
