package backlog

import (
	"fmt"
	"strings"
	"text/tabwriter"
)

// Tabla (E001/F007/S002): las fichas en columnas alineadas, en el orden recibido (el de Activas).
// Celda sin valor = «—»; la descripción es el título. Escribe en memoria: no hay error que propagar.
func Tabla(fs []Ficha) string {
	var b strings.Builder
	w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "Iniciativa\tFeature\tSlice\tEstado\tPrior.\tDescripción")
	for _, f := range fs {
		p := strings.Split(f.ID, "/")
		for len(p) < 3 {
			p = append(p, "—")
		}
		_, _ = fmt.Fprintln(w, strings.Join([]string{p[0], p[1], p[2], f.Estado, f.Prioridad, f.Titulo}, "\t"))
	}
	_ = w.Flush()
	return b.String()
}
