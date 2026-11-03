package wiki

// E001/F017/S006: la página del slice EN_CURSO no puede ser más vieja que el código que el slice está cambiando.

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/backlog"
)

// ahora: hora de lo que cambió sin commit; gana a cualquier commit.
const ahora = math.MaxInt64

// RevisarAlDia: un problema si la hora del código vivo del slice EN_CURSO es más nueva que la última edición del
// texto propio de su página. Sin git, sin commits o sin slice en curso no revisa nada.
func RevisarAlDia(raiz string) []string {
	if _, err := gitAlDia(raiz, "rev-parse", "HEAD"); err != nil {
		return nil
	}
	id, hay, err := backlog.SliceEnCurso(raiz)
	if err != nil || !hay {
		return nil
	}
	codigo, err := backlog.CodigoVivo(raiz, id)
	if err != nil || len(codigo) == 0 {
		return nil
	}
	pagina := "docs/" + id + ".md"
	hc, archivo := horaCodigo(raiz, id, codigo)
	if hc <= horaPagina(raiz, pagina) {
		return nil
	}
	return []string{fmt.Sprintf("%s: el código de %s (%s) cambió después de la última edición de la página; actualiza su texto (la tabla que reescribe `aquas compile` no cuenta) (E001/F017/S006)", pagina, id, archivo)}
}

// horaCodigo: «ahora» si algún archivo tiene cambios sin commit; si no, el último commit que nombra el ID y lo toca.
func horaCodigo(raiz, id string, codigo []string) (int64, string) {
	if out, _ := gitAlDia(raiz, append([]string{"status", "--porcelain", "--"}, codigo...)...); strings.TrimSpace(out) != "" {
		for _, a := range codigo {
			if o, _ := gitAlDia(raiz, "status", "--porcelain", "--", a); strings.TrimSpace(o) != "" {
				return ahora, a
			}
		}
	}
	var hora int64
	archivo := ""
	for _, a := range codigo {
		out, _ := gitAlDia(raiz, "log", "-1", "--format=%ct", "-E", "--grep="+regexp.QuoteMeta(id)+"([^0-9]|$)", "--", a)
		if h, err := strconv.ParseInt(strings.TrimSpace(out), 10, 64); err == nil && h > hora {
			hora, archivo = h, a
		}
	}
	return hora, archivo
}

// horaPagina: «ahora» si su texto propio cambió sin commit; si no, el commit más nuevo que cambió ese texto; 0 si nunca.
func horaPagina(raiz, pagina string) int64 {
	b, err := os.ReadFile(filepath.Join(raiz, filepath.FromSlash(pagina)))
	if err != nil {
		return 0
	}
	cabeza, _ := gitAlDia(raiz, "show", "HEAD:"+pagina)
	if textoPropio(string(b)) != textoPropio(cabeza) {
		return ahora
	}
	out, _ := gitAlDia(raiz, "log", "--format=%H %ct", "--", pagina)
	for _, l := range strings.Split(strings.TrimSpace(out), "\n") {
		h, ct, ok := strings.Cut(l, " ")
		if !ok {
			continue
		}
		nuevo, _ := gitAlDia(raiz, "show", h+":"+pagina)
		viejo, _ := gitAlDia(raiz, "show", h+"^:"+pagina) // primer commit o página nueva: viejo vacío
		if textoPropio(nuevo) != textoPropio(viejo) {
			t, _ := strconv.ParseInt(ct, 10, 64)
			return t
		}
	}
	return 0
}

// textoPropio: la página sin los bloques GENERADO y con fin de línea LF.
func textoPropio(c string) string {
	c = strings.ReplaceAll(c, "\r\n", "\n")
	for {
		ini, fin := bloqueActual(c)
		if ini < 0 || fin < 0 {
			return c
		}
		c = c[:ini] + c[fin:]
	}
}

// gitAlDia corre git en la raíz y devuelve su salida.
func gitAlDia(raiz string, args ...string) (string, error) {
	out, err := exec.Command("git", append([]string{"-C", raiz, "-c", "core.quotepath=off"}, args...)...).Output()
	return string(out), err
}
