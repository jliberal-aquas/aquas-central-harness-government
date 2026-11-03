package wiki

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestPaginaAlDiaConCodigo (E001/F017/S006): el código vivo del slice EN_CURSO no puede ser más nuevo que la
// última edición del texto propio de su página (sin los bloques GENERADO).
// Rojo cuando: el código cambia y la página no.
// Verde cuando: la página se actualiza después del código.
func TestPaginaAlDiaConCodigo(t *testing.T) {
	raiz := t.TempDir()
	n := 0
	git := func(args ...string) {
		t.Helper()
		n++
		c := exec.Command("git", append([]string{"-C", raiz, "-c", "user.name=t", "-c", "user.email=t@t", "-c", "core.autocrlf=false"}, args...)...)
		fecha := "2026-01-01T00:00:" + strings.Repeat("0", 2-len(itoa(n))) + itoa(n) + "Z"
		c.Env = append(os.Environ(), "GIT_COMMITTER_DATE="+fecha, "GIT_AUTHOR_DATE="+fecha)
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	poner := func(rel, texto string) {
		t.Helper()
		p := filepath.Join(raiz, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(texto), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	ficha := func(carpeta, estado string) {
		t.Helper()
		for _, c := range []string{"activo", "archivo"} {
			_ = os.Remove(filepath.Join(raiz, "trabajo", c, "E001-F002-S001.json"))
		}
		poner("trabajo/"+carpeta+"/E001-F002-S001.json", `{"id": "E001/F002/S001", "tipo": "slice", "titulo": "A", "estado": "`+estado+`", "prioridad": "P1", "motivo": "m", "padre": "E001/F002", "bloqueado_por": [], "secuencia": 1}`)
	}
	pagina := "docs/E001/F002/S001.md"
	texto := func(propio, generado string) string {
		return "# S001\n\n" + propio + "\n\n" + marcaIni + "\n" + generado + "\n" + marcaFin + "\n"
	}
	revisar := func(caso string, falla bool) {
		t.Helper()
		ps := RevisarAlDia(raiz)
		if falla && (len(ps) != 1 || !strings.Contains(ps[0], pagina)) || !falla && len(ps) != 0 {
			t.Errorf("%s: falla=%v; dio %v", caso, falla, ps)
		}
	}

	// Fallo a mitad de camino: sin git, y con git sin commits, no se revisa.
	poner("trabajo/activo/E001-F002.json", `{"id": "E001/F002", "tipo": "feature", "titulo": "F", "estado": "EN_CURSO", "prioridad": "P1", "motivo": "m", "padre": "E001", "bloqueado_por": [], "secuencia": 1}`)
	ficha("activo", "EN_CURSO")
	poner(pagina, texto("contrato v1", "tabla v1"))
	poner("internal/x/x.go", "package x\n")
	revisar("repo sin git", false)
	git("init", "-q")
	revisar("repo sin commits", false)
	git("add", "-A")
	git("commit", "-q", "-m", "base")
	revisar("sin cambios", false)

	poner("internal/x/x_test.go", "package x\n")
	revisar("código solo en pruebas no cuenta", false)
	poner("internal/x/x.go", "package x\n\nvar A = 1\n")
	revisar("código sin commit y página sin tocar", true)
	poner(pagina, texto("contrato v1", "tabla v2"))
	revisar("compile solo reescribió la tabla generada", true)
	poner(pagina, texto("contrato v2", "tabla v2"))
	revisar("página editada sin commit", false)
	git("add", "-A")
	git("commit", "-q", "-m", "E001/F002/S001: código y página")
	revisar("mismo commit", false)

	poner("internal/x/x.go", "package x\n\nvar A = 2\n")
	git("add", "-A")
	git("commit", "-q", "-m", "E001/F002/S001: más código")
	revisar("commit de código más nuevo que la página", true)
	poner(pagina, texto("contrato v2", "tabla v3"))
	git("add", "-A")
	git("commit", "-q", "-m", "E001/F002/S001: compile")
	revisar("commit que solo cambia la tabla generada", true)
	poner(pagina, texto("contrato v3", "tabla v3"))
	git("add", "-A")
	git("commit", "-q", "-m", "docs: página al día")
	revisar("commit de la página después del código", false)

	poner("internal/x/x.go", "package x\n\nvar A = 3\n")
	ficha("activo", "BORRADOR")
	revisar("slice BORRADOR; la feature en curso no se revisa", false)
	ficha("archivo", "CERRADO")
	revisar("slice archivado", false)
	ficha("activo", "EN_CURSO")
	revisar("de vuelta EN_CURSO con código sin commit", true)
}

// itoa: entero corto a texto, sin strconv para no ensuciar la prueba.
func itoa(i int) string {
	if i < 10 {
		return string(rune('0' + i))
	}
	return itoa(i/10) + string(rune('0'+i%10))
}
