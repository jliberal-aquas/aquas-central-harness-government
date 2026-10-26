package calidad

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestLintBloquea (E001/F013/S001): el gate falla ante código sin formato o con error ignorado.
// Rojo cuando: el canario pasa, la herramienta ausente o con otra versión no da error, o el repo tiene hallazgos.
// Verde cuando: el canario da errcheck y gofmt con su archivo, ausente y versión distinta fallan, y el repo sale limpio.
func TestLintBloquea(t *testing.T) {
	raiz, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	config := filepath.Join(raiz, "gobierno", "config", "golangci.yml")
	canario := filepath.Join(raiz, "internal", "calidad", "testdata", "canario")

	h, err := Revisar(canario, config)
	if err != nil {
		t.Fatalf("el canario debió revisarse: %v", err)
	}
	todo := strings.Join(h, "\n")
	for _, quiero := range []string{"errcheck.go", "(errcheck)", "sinformato.go", "(gofmt)"} {
		if !strings.Contains(todo, quiero) {
			t.Errorf("el canario no reportó %q; hallazgos:\n%s", quiero, todo)
		}
	}

	if err := chequearVersion("golangci-lint has version 2.14.0 built with go1.26.2"); err == nil {
		t.Error("la versión 2.14.0 debió fallar")
	}
	if err := chequearVersion("golangci-lint has version " + Version + " built with go1.26.2"); err != nil {
		t.Errorf("la versión fija debió pasar: %v", err)
	}

	if h, err := Revisar(raiz, config); err != nil || len(h) > 0 {
		t.Errorf("el repo debió salir limpio: err %v, %d hallazgos:\n%s", err, len(h), strings.Join(h, "\n"))
	}

	t.Run("errcheck nombra archivo y linter en la misma línea", func(t *testing.T) {
		for _, l := range h {
			if strings.Contains(l, "errcheck.go:") && strings.HasSuffix(l, "(errcheck)") {
				return
			}
		}
		t.Errorf("ninguna línea junta errcheck.go:N y (errcheck):\n%s", todo)
	})

	t.Run("ausente lo dice", func(t *testing.T) {
		t.Setenv("PATH", t.TempDir())
		if _, err := Revisar(canario, config); err == nil || !strings.Contains(err.Error(), "herramienta ausente") {
			t.Errorf("quiero «herramienta ausente»; err %v", err)
		}
	})

	dirFalso := construir(t, "./testdata/falso", "golangci-lint")
	t.Run("Revisar usa la versión", func(t *testing.T) {
		t.Setenv("PATH", dirFalso)
		t.Setenv("FALSO_VERSION", "2.14.0")
		if _, err := Revisar(canario, config); err == nil || !strings.Contains(err.Error(), "debe ser la versión "+Version) {
			t.Errorf("con 2.14.0 quiero «debe ser la versión %s»; err %v", Version, err)
		}
	})
	t.Run("caída nunca es limpio", func(t *testing.T) {
		t.Setenv("PATH", dirFalso)
		t.Setenv("FALSO_VERSION", Version)
		linea := "a.go:1:1: algo (errcheck)\n"
		for _, c := range []struct {
			exit, salida string
			falla        bool
		}{{"2", "", true}, {"1", "", true}, {"3", linea, true}, {"1", linea, false}, {"0", "", false}} {
			t.Setenv("FALSO_EXIT", c.exit)
			t.Setenv("FALSO_SALIDA", c.salida)
			hs, err := Revisar(canario, config)
			if (err != nil) != c.falla {
				t.Errorf("exit %s salida %q: err %v, quiero falla=%v (hallazgos %v)", c.exit, c.salida, err, c.falla, hs)
			}
		}
	})

	t.Run("verificar sale 1 solo por calidad Go", func(t *testing.T) {
		aq := filepath.Join(construir(t, "../../cmd/aquas", "aquas"), nombreExe("aquas"))
		copia := copiarRepo(t, raiz)
		// la guardia exige que la raíz declarada sea la real: la copia declara la suya.
		entorno := `{"version": 1, "so": "` + runtime.GOOS + `", "raiz": "` + filepath.ToSlash(copia) + `"}`
		if err := os.WriteFile(filepath.Join(copia, "gobierno", "config", "entorno.local.json"), []byte(entorno), 0o644); err != nil {
			t.Fatal(err)
		}
		f := filepath.Join(copia, "internal", "calidad", "calidad.go")
		fi, err := os.Stat(f)
		if err != nil {
			t.Fatal(err)
		}
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		correr := func(entorno ...string) (string, int) {
			cmd := exec.Command(aq, "verificar")
			cmd.Dir = copia
			cmd.Env = append(os.Environ(), entorno...)
			out, err := cmd.CombinedOutput()
			var ee *exec.ExitError
			if errors.As(err, &ee) {
				return string(out), ee.ExitCode()
			}
			if err != nil {
				t.Fatal(err)
			}
			return string(out), 0
		}
		if out, codigo := correr(); codigo != 0 || !strings.Contains(out, "GATE OK") {
			t.Fatalf("la copia sin romper debió dar GATE OK para aislar la calidad; exit %d, salida:\n%s", codigo, out)
		}
		if err := os.WriteFile(f, []byte(strings.Replace(string(b), "const Version", "const   Version", 1)), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(f, fi.ModTime(), fi.ModTime()); err != nil {
			t.Fatal(err)
		}
		out, codigo := correr()
		ruta := filepath.Join("internal", "calidad", "calidad.go") + ":"
		if codigo != 1 || !strings.Contains(out, "GATE FALLIDO · calidad Go:") || !strings.Contains(out, ruta) ||
			strings.Contains(out, ".."+string(filepath.Separator)) || strings.Contains(out, "GATE OK") {
			t.Errorf("solo calidad Go falla: quiero exit 1, «calidad Go» y %q relativo a la raíz; exit %d, salida:\n%s", ruta, codigo, out)
		}
		out, codigo = correr("PATH="+dirFalso+string(os.PathListSeparator)+os.Getenv("PATH"), "FALSO_VERSION=2.14.0")
		if codigo != 1 || !strings.Contains(out, "debe ser la versión "+Version) || strings.Contains(out, "GATE OK") {
			t.Errorf("con golangci-lint 2.14.0 verificar debió salir 1 y decirlo; exit %d, salida:\n%s", codigo, out)
		}
	})

	t.Run("RevisarCentral sin go.mod calla e ilegible avisa", func(t *testing.T) {
		if hs := RevisarCentral(t.TempDir()); hs != nil {
			t.Errorf("sin go.mod no hay nada que revisar; hallazgos %v", hs)
		}
		if hs := RevisarCentral("ilegible\x00raiz"); len(hs) != 1 || !strings.Contains(hs[0], "go.mod ilegible") {
			t.Errorf("go.mod ilegible debió volver como hallazgo; %v", hs)
		}
	})
}

// construir compila pkg como ejecutable nombre en una carpeta temporal y devuelve la carpeta.
func construir(t *testing.T, pkg, nombre string) string {
	t.Helper()
	dir := t.TempDir()
	if out, err := exec.Command("go", "build", "-o", filepath.Join(dir, nombreExe(nombre)), pkg).CombinedOutput(); err != nil {
		t.Fatalf("go build %s: %v\n%s", pkg, err, out)
	}
	return dir
}

// nombreExe agrega .exe en Windows.
func nombreExe(n string) string {
	if runtime.GOOS == "windows" {
		return n + ".exe"
	}
	return n
}

// copiarRepo copia el árbol de trabajo de raiz, sin .git ni .tmp, a una carpeta temporal.
func copiarRepo(t *testing.T, raiz string) string {
	t.Helper()
	dst := t.TempDir()
	err := filepath.WalkDir(raiz, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(raiz, p)
		if err != nil {
			return err
		}
		if d.IsDir() {
			if rel == ".git" || rel == ".tmp" {
				return filepath.SkipDir
			}
			return os.MkdirAll(filepath.Join(dst, rel), 0o755)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dst, rel), b, 0o644); err != nil {
			return err
		}
		// conserva la fecha: la guardia rechaza fuentes más nuevas que el binario de aquas.
		return os.Chtimes(filepath.Join(dst, rel), info.ModTime(), info.ModTime())
	})
	if err != nil {
		t.Fatal(err)
	}
	return dst
}
