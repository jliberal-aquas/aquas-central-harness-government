package guardia

import (
	"path"
	"strings"
	"testing"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/backlog"
	mem "github.com/jliberal-aquas/aquas-central-harness-government/internal/memoria"
)

// E000/F005/S005: las carpetas protegidas salen de los paquetes que las usan, no de una copia.
// Si memoria o backlog mueven su carpeta, la guardia protege la nueva sin que nadie la toque.
func TestCarpetasProtegidasSalenDeSusPaquetes(t *testing.T) {
	raiz := raizPrueba(t)
	for _, c := range []struct{ dir, motivo string }{
		{path.Dir(mem.Ruta), "memoria caliente: "},
		{mem.Archivo, "archivo de memoria: "},
		{backlog.Activo, "backlog: "},
		{backlog.Archivo, "backlog: "},
	} {
		p := c.dir + "/x.json"
		if m := RevisarArchivo(raiz, p); !strings.Contains(m, c.motivo+p) {
			t.Errorf("Write a %s: se esperaba %q; motivo %q", p, c.motivo+p, m)
		}
		bash := `{"tool_name":"Bash","tool_input":{"command":"echo {} > ` + p + `"}}`
		if m := Revisar(raiz, []byte(bash)); !strings.Contains(m, c.motivo) {
			t.Errorf("Bash hacia %s: se esperaba %q; motivo %q", p, c.motivo, m)
		}
	}
}
