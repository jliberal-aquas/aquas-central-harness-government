package memoria

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// E001/F005/S002: huella de Git, hook de inicio de sesión y hook de cierre de turno.

func git(t *testing.T, raiz string, args ...string) {
	t.Helper()
	c := exec.Command("git", append([]string{"-C", raiz, "-c", "user.name=t", "-c", "user.email=t@t", "-c", "commit.gpgsign=false"}, args...)...)
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// repoGit: repositorio propio en una carpeta temporal, con un commit.
func repoGit(t *testing.T) string {
	t.Helper()
	raiz := t.TempDir()
	git(t, raiz, "init", "-q")
	escribirArchivo(t, raiz, "a.txt", "uno\n")
	git(t, raiz, "add", "a.txt")
	git(t, raiz, "commit", "-q", "-m", "inicio")
	return raiz
}

func escribirArchivo(t *testing.T, raiz, rel, contenido string) {
	t.Helper()
	p := filepath.Join(raiz, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(contenido), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestHuellaCambiaConElEstadoDeGit(t *testing.T) {
	raiz := repoGit(t)
	h1 := HuellaGit(raiz)
	if h1 == "" {
		t.Fatal("repo propio debe dar huella")
	}
	fijar(t, raiz, "iniciativa=E001", "tarea=x") // escribe .aquas/: no cuenta
	if h := HuellaGit(raiz); h != h1 {
		t.Fatalf("escribir la memoria no debe cambiar la huella")
	}
	escribirArchivo(t, raiz, "a.txt", "dos\n")
	h2 := HuellaGit(raiz)
	if h2 == h1 {
		t.Fatal("un archivo modificado debe cambiar la huella")
	}
	escribirArchivo(t, raiz, "a.txt", "tres\n")
	if h := HuellaGit(raiz); h == h2 {
		t.Fatal("otro cambio al mismo archivo también debe cambiar la huella")
	}
	git(t, raiz, "commit", "-q", "-am", "dos")
	if h := HuellaGit(raiz); h == h1 || h == h2 {
		t.Fatal("un commit nuevo debe cambiar la huella")
	}
}

func TestHuellaFueraDeGitVacia(t *testing.T) {
	if h := HuellaGit(t.TempDir()); h != "" {
		t.Fatalf("carpeta que no es raíz de un repo: huella vacía; dio %q", h)
	}
}

func TestFijarGuardaLaHuella(t *testing.T) {
	raiz := repoGit(t)
	m := fijar(t, raiz, "iniciativa=E001")
	// E000/F005/S003: valor calculado aparte. Repo limpio (.aquas excluida) → huella = sha256(salida de rev-parse HEAD).
	head, err := exec.Command("git", "-C", raiz, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	suma := sha256.Sum256(head)
	if quiere := hex.EncodeToString(suma[:]); m.HuellaGit != quiere {
		t.Fatalf("fijar debe guardar la huella del estado de git: %q; se esperaba %q", m.HuellaGit, quiere)
	}
}

func TestCierre(t *testing.T) {
	casos := []struct {
		nombre string
		armar  func(t *testing.T, raiz string)
		activo bool
		espera string // "" = deja cerrar
	}{
		{"sin memoria", func(t *testing.T, r string) { escribirArchivo(t, r, "a.txt", "cambio\n") }, false, ""},
		{"foco vacío", func(t *testing.T, r string) {
			fijar(t, r, "siguiente=nada")
			escribirArchivo(t, r, "a.txt", "cambio\n")
		}, false, ""},
		{"al día", func(t *testing.T, r string) {
			escribirArchivo(t, r, "a.txt", "cambio\n")
			fijar(t, r, "iniciativa=E001", "tarea=x")
		}, false, ""},
		{"vieja", func(t *testing.T, r string) {
			fijar(t, r, "iniciativa=E001", "tarea=x")
			escribirArchivo(t, r, "a.txt", "cambio\n")
		}, false, "memoria caliente vieja"},
		{"vieja pero ya se pidió", func(t *testing.T, r string) {
			fijar(t, r, "iniciativa=E001", "tarea=x")
			escribirArchivo(t, r, "a.txt", "cambio\n")
		}, true, ""},
		{"sello roto", func(t *testing.T, r string) {
			fijar(t, r, "iniciativa=E001", "tarea=original")
			p := filepath.Join(r, Ruta)
			b, _ := os.ReadFile(p)
			if err := os.WriteFile(p, []byte(strings.Replace(string(b), "original", "trucada", 1)), 0o644); err != nil {
				t.Fatal(err)
			}
		}, false, "el sello no coincide"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			raiz := repoGit(t)
			c.armar(t, raiz)
			got := Cierre(raiz, c.activo)
			if c.espera == "" && got != "" {
				t.Fatalf("debe dejar cerrar; bloqueó: %s", got)
			}
			if c.espera != "" && !strings.Contains(got, c.espera) {
				t.Fatalf("debe bloquear con %q; dio %q", c.espera, got)
			}
		})
	}
}

func TestCierreDiceComoArreglar(t *testing.T) {
	raiz := repoGit(t)
	fijar(t, raiz, "iniciativa=E001", "tarea=x")
	escribirArchivo(t, raiz, "b.txt", "nuevo\n")
	if got := Cierre(raiz, false); !strings.Contains(got, "aquas memoria fijar") {
		t.Fatalf("el bloqueo debe decir el comando que lo resuelve: %q", got)
	}
}

func TestInicio(t *testing.T) {
	raiz := repoGit(t)
	if got := Inicio(raiz); !strings.Contains(got, "sin memoria caliente") || !strings.Contains(got, "aquas memoria fijar") {
		t.Fatalf("sin memoria: aviso con el comando; dio %q", got)
	}
	fijar(t, raiz, "iniciativa=E001", "feature=F005", "tarea=hooks")
	if got := Inicio(raiz); !strings.Contains(got, "MEMORIA CALIENTE") || !strings.Contains(got, "tarea: hooks") {
		t.Fatalf("con memoria: debe mostrarla; dio %q", got)
	}
	p := filepath.Join(raiz, Ruta)
	b, _ := os.ReadFile(p)
	if err := os.WriteFile(p, []byte(strings.Replace(string(b), "hooks", "trucada", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := Inicio(raiz); !strings.Contains(got, "el sello no coincide") {
		t.Fatalf("sello roto: debe avisar; dio %q", got)
	}
}

// El sello cubre los bytes escritos, no la estructura actual: agregar un campo
// (como huella_git en S002) no debe hacer pasar una memoria vieja por editada a mano.
func TestSelloSobreviveCampoNuevo(t *testing.T) {
	raiz := t.TempDir()
	sinSello := "{\n  \"foco\": {\n    \"iniciativa\": \"E001\",\n    \"feature\": \"\",\n    \"slice\": \"\",\n    \"tarea\": \"vieja\"\n  },\n" +
		"  \"siguiente\": \"\",\n  \"pendientes\": null,\n  \"decisiones\": null,\n  \"actualizado\": \"2026-10-20T12:00:00Z\",\n  \"sello\": \"\"\n}\n"
	s := sha256.Sum256([]byte(sinSello))
	escribirArchivo(t, raiz, Ruta, strings.Replace(sinSello, `"sello": ""`, `"sello": "`+hex.EncodeToString(s[:])+`"`, 1))
	m, err := Leer(raiz)
	if err != nil {
		t.Fatalf("memoria escrita antes del campo nuevo debe leerse: %v", err)
	}
	if m.Foco.Tarea != "vieja" {
		t.Fatalf("tarea: %q", m.Foco.Tarea)
	}
}
