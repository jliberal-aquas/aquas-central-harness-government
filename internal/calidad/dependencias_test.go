package calidad

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

// TestImportNoPermitido (E001/F013/S002): un import interno fuera de la tabla, de código o de prueba, y un paquete sin fila hacen fallar el gate.
// Rojo cuando: un import prohibido, uno solo de prueba o un paquete sin fila pasan.
// Verde cuando: cada uno falla nombrando arista o paquete y el repo de hoy pasa.
func TestImportNoPermitido(t *testing.T) {
	raiz, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	deps := filepath.Join(raiz, "internal", "calidad", "testdata", "deps")
	if h, err := RevisarDependencias(deps, filepath.Join(deps, "tabla_ok.json")); err != nil || len(h) > 0 {
		t.Errorf("la tabla completa debió pasar: err %v, hallazgos %v", err, h)
	}
	h, err := RevisarDependencias(deps, filepath.Join(deps, "tabla_mala.json"))
	if err != nil {
		t.Fatalf("la tabla mala debió revisarse: %v", err)
	}
	todo := strings.Join(h, "\n")
	for _, quiero := range []string{
		"internal/a → internal/b no está en la tabla", // criterio 1: import de código
		"internal/c → internal/b no está en la tabla", // criterio 2: import solo de prueba
		"internal/d sin fila en la tabla",             // criterio 3: paquete nuevo
		"internal/fantasma no existe",                 // caso borde: fila de un paquete que ya no está
		"internal/e → internal/b no está en la tabla", // import de prueba externa (package e_test); e → e no cuenta
	} {
		if !strings.Contains(todo, quiero) {
			t.Errorf("falta %q; hallazgos:\n%s", quiero, todo)
		}
	}
	if _, err := RevisarDependencias(deps, filepath.Join(deps, "no_existe.json")); err == nil {
		t.Error("tabla ausente debió fallar")
	}
	if strings.Contains(todo, "internal/e → internal/e") {
		t.Error("el propio paquete no es una arista")
	}
	// el orden de un mapa cambia por corrida: 20 corridas ordenadas por azar son casi imposibles.
	for i := 0; i < 20 && !t.Failed(); i++ {
		otra, err := RevisarDependencias(deps, filepath.Join(deps, "tabla_mala.json"))
		if err != nil || !sort.StringsAreSorted(otra) {
			t.Errorf("hallazgos sin orden fijo (corrida %d): err %v, %v", i, err, otra)
		}
	}
	if _, err := RevisarDependencias(deps, filepath.Join(deps, "tabla_rota.json")); err == nil {
		t.Error("tabla ilegible debió fallar")
	}
	roto := t.TempDir() // cae dentro del repo (TMP=.tmp): un go.mod roto obliga a go list a fallar
	if err := os.WriteFile(filepath.Join(roto, "go.mod"), []byte("esto no es un go.mod\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := RevisarDependencias(roto, filepath.Join(deps, "tabla_ok.json")); err == nil {
		t.Error("go list con go.mod roto debió fallar")
	}
	t.Run("RevisarCentral sin tabla avisa", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module x\n\ngo 1.26\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if hs := RevisarCentral(dir); !strings.Contains(strings.Join(hs, "\n"), "dependencias.json") {
			t.Errorf("sin tabla debió volver un hallazgo que la nombre; %v", hs)
		}
	})
	tablaRepo := filepath.Join(raiz, "gobierno", "config", "dependencias.json")
	if h, err := RevisarDependencias(raiz, tablaRepo); err != nil || len(h) > 0 {
		t.Errorf("criterio 4: el repo de hoy debió pasar: err %v, hallazgos %v", err, h)
	}

	t.Run("verificar falla por una arista", func(t *testing.T) {
		aq := filepath.Join(construir(t, "../../cmd/aquas", "aquas"), nombreExe("aquas"))
		copia := copiarRepo(t, raiz)
		entorno := `{"version": 1, "so": "` + runtimeGOOS() + `", "raiz": "` + filepath.ToSlash(copia) + `"}`
		if err := os.WriteFile(filepath.Join(copia, "gobierno", "config", "entorno.local.json"), []byte(entorno), 0o644); err != nil {
			t.Fatal(err)
		}
		if out, codigo := correrAquas(t, aq, copia); codigo != 0 || !strings.Contains(out, "GATE OK") {
			t.Fatalf("la copia sin tocar debió dar GATE OK para aislar las dependencias; exit %d:\n%s", codigo, out)
		}
		f := filepath.Join(copia, "gobierno", "config", "dependencias.json")
		fi, err := os.Stat(f)
		if err != nil {
			t.Fatal(err)
		}
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		var m map[string][]string
		if err := json.Unmarshal(b, &m); err != nil {
			t.Fatal(err)
		}
		m["internal/backlog"] = []string{} // quita la arista real internal/backlog → internal/compilador
		nuevo, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(f, nuevo, 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(f, fi.ModTime(), fi.ModTime()); err != nil {
			t.Fatal(err)
		}
		out, codigo := correrAquas(t, aq, copia)
		if codigo != 1 || !strings.Contains(out, "internal/backlog → internal/compilador no está en la tabla") || strings.Contains(out, "GATE OK") {
			t.Errorf("verificar debió salir 1 nombrando la arista; exit %d:\n%s", codigo, out)
		}
	})
}

// correrAquas corre `aquas verificar` en dir y devuelve la salida y el código de salida.
func correrAquas(t *testing.T, aq, dir string) (string, int) {
	t.Helper()
	cmd := exec.Command(aq, "verificar")
	cmd.Dir = dir
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

// runtimeGOOS devuelve el sistema operativo para el entorno de la copia.
func runtimeGOOS() string { return runtime.GOOS }
