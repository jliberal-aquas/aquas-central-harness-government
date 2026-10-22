package wiki

// E001/F008/S003: las dos portadas (README.md y docs/README.md) listan todas las iniciativas con enlace.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/backlog"
)

// faltanIniciativas: problemas del gate si alguna iniciativa no está enlazada en una portada.
// README.md enlaza docs/<id>/README.md; docs/README.md enlaza <id>/README.md.
func faltanIniciativas(ids []string, readme, inicioWiki string) []string {
	var ps []string
	for _, id := range ids {
		if !strings.Contains(readme, "](docs/"+id+"/README.md)") {
			ps = append(ps, fmt.Sprintf("README.md: falta enlace a la iniciativa %s (docs/%s/README.md)", id, id))
		}
		if !strings.Contains(inicioWiki, "]("+id+"/README.md)") {
			ps = append(ps, fmt.Sprintf("%s: falta enlace a la iniciativa %s (%s/README.md)", inicio, id, id))
		}
	}
	return ps
}

// revisarPortadas: lee las iniciativas del backlog y las dos portadas; solo exige las iniciativas que ya tienen página.
func revisarPortadas(raiz string) ([]string, error) {
	fichas, err := backlog.Todas(raiz)
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, f := range fichas {
		if f.Tipo == "iniciativa" && f.Estado != "CANCELADO" && existe(raiz, paginaDe(f.ID)) {
			ids = append(ids, f.ID)
		}
	}
	if len(ids) == 0 {
		return nil, nil
	}
	leer := func(rel string) string {
		b, _ := os.ReadFile(filepath.Join(raiz, filepath.FromSlash(rel)))
		return string(b)
	}
	return faltanIniciativas(ids, leer("README.md"), leer(inicio)), nil
}
