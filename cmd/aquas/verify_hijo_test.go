package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/herencia"
)

// hooksPathDe devuelve core.hooksPath del repo en dir ("" si no está fijado).
func hooksPathDe(dir string) string {
	out, _ := exec.Command("git", "-C", dir, "config", "--get", "core.hooksPath").Output()
	return strings.TrimSpace(string(out))
}

// hijoNuevo instala el gobierno en una carpeta con git y devuelve su ruta.
func hijoNuevo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	git(t, dir, "init", "-q")
	if s := aquas(t, dir, "", "install"); s.codigo != 0 {
		t.Fatalf("install: %+v", s)
	}
	return dir
}

// TestVerifyEnElHijo (E003/F001/S005): verify del hijo corre los chequeos marcados «todos» y el pre-commit lo exige.
// Rojo cuando: un chequeo todos no corre en el hijo o falta el pre-commit.
// Verde cuando: el gate del hijo falla ante una ficha rota y el pre-commit lo corre.
func TestVerifyEnElHijo(t *testing.T) {
	const preCommit = ".githooks/pre-commit"
	leer := func(dir, rel string) string {
		b, _ := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel)))
		return string(b)
	}

	// Ficha rota del hijo: verify falla y nombra backlog; sin la ficha rota, backlog no aparece.
	dir := hijoNuevo(t)
	base := aquas(t, dir, "", "verify")
	if strings.Contains(base.out+base.err, "backlog trabajo/") {
		t.Errorf("sin fichas rotas, verify del hijo no debe quejarse de backlog; dio %+v", base)
	}
	escribir(t, dir, "trabajo/activo/E001-F001-S001.json", "{roto")
	if s := aquas(t, dir, "", "verify"); s.codigo == 0 || !strings.Contains(s.out+s.err, "backlog trabajo/") {
		t.Errorf("verify del hijo debió fallar por la ficha rota y nombrar backlog; dio %+v", s)
	}

	// Chequeo marcado «central» (derivados, calidad-go) no corre en el hijo.
	escribir(t, dir, "roto.go", "package x\nfunc (")
	if s := aquas(t, dir, "", "verify"); strings.Contains(s.out+s.err, "derivados") || strings.Contains(s.out+s.err, "calidad Go") {
		t.Errorf("derivados y calidad-go son de la central; no deben correr en el hijo: %+v", s)
	}

	// Un chequeo sin fila en la herencia sellada del hijo no corre y se informa.
	sinFila := t.TempDir()
	escribir(t, sinFila, ".aquas/central/gobierno/config/herencia.json",
		`{"version":1,"controles":[{"control":"gate:wiki","destino":"todos","motivo":"m"}]}`)
	if corren, probs := herencia.ParaHijo(sinFila); len(corren) != 1 || corren[0] != "wiki" || len(probs) != len(herencia.Chequeos)-1 ||
		!strings.Contains(probs[0], "calidad-go") {
		t.Errorf("ParaHijo: debe correr solo wiki e informar los demás sin fila; dio %v %v", corren, probs)
	}

	// Hijo nuevo: pre-commit con aquas verify y core.hooksPath fijado.
	if !strings.Contains(leer(dir, preCommit), "aquas verify") {
		t.Errorf("install debió escribir %s con aquas verify; dio %q", preCommit, leer(dir, preCommit))
	}
	if h := hooksPathDe(dir); h != ".githooks" {
		t.Errorf("install debió fijar core.hooksPath=.githooks; dio %q", h)
	}

	// Hijo existente: update repone el pre-commit y el hooksPath; con el mismo contenido no falla.
	_ = os.Remove(filepath.Join(dir, filepath.FromSlash(preCommit)))
	_ = exec.Command("git", "-C", dir, "config", "--unset", "core.hooksPath").Run()
	escribir(t, dir, "roto.go", "package x\n")
	if s := aquas(t, dir, "", "update"); s.codigo != 0 {
		t.Fatalf("update en hijo existente: %+v", s)
	}
	if !strings.Contains(leer(dir, preCommit), "aquas verify") || hooksPathDe(dir) != ".githooks" {
		t.Errorf("update debió reponer el pre-commit y core.hooksPath; dio %q %q", leer(dir, preCommit), hooksPathDe(dir))
	}
	if s := aquas(t, dir, "", "update"); s.codigo != 0 {
		t.Errorf("update con el pre-commit ya correcto no debe fallar: %+v", s)
	}

	// Pre-commit propio distinto: update falla y pide unirlo a mano y repetir update; no lo pisa.
	propio := "#!/bin/sh\necho propio\n"
	escribir(t, dir, preCommit, propio)
	s := aquas(t, dir, "", "update")
	if s.codigo == 0 || !strings.Contains(s.err, "a mano") || !strings.Contains(s.err, "aquas update") {
		t.Errorf("update con pre-commit propio debe fallar pidiendo unirlo a mano y repetir aquas update; dio %+v", s)
	}
	if leer(dir, preCommit) != propio {
		t.Errorf("el pre-commit propio no se pisa; quedó %q", leer(dir, preCommit))
	}
	// Hallazgo A2.1: aquas verify que no es el último comando no vale (sh sale con el código del último).
	for _, malo := range []string{"#!/bin/sh\naquas verify\necho propio\n", "#!/bin/sh\nexit 0\naquas verify\nexit 0\n"} {
		escribir(t, dir, preCommit, malo)
		if s := aquas(t, dir, "", "update"); s.codigo == 0 || !strings.Contains(s.err, "último comando") {
			t.Errorf("update con aquas verify que no es el último comando debe fallar; dio %+v para %q", s, malo)
		}
		if s := aquas(t, dir, "", "verify"); !strings.Contains(s.out+s.err, "git:pre-commit") {
			t.Errorf("el gate debe rechazar un pre-commit cuyo último comando no es aquas verify; dio %+v", s)
		}
	}
	// Hallazgo A1.1 y A1.5: unido a mano (con CRLF de Windows), update lo acepta, no lo pisa y el gate no se queja.
	unido := "#!/bin/sh\r\necho propio\r\naquas verify\r\n"
	escribir(t, dir, preCommit, unido)
	if s := aquas(t, dir, "", "update"); s.codigo != 0 || leer(dir, preCommit) != unido {
		t.Errorf("update con el pre-commit propio ya unido debe pasar sin pisarlo; dio %+v %q", s, leer(dir, preCommit))
	}
	if s := aquas(t, dir, "", "verify"); strings.Contains(s.out+s.err, "git:pre-commit") {
		t.Errorf("con el pre-commit unido, el gate no debe quejarse del pre-commit; dio %+v", s)
	}

	// Install sobre carpeta con pre-commit propio: falla con el mismo pedido.
	conPropio := t.TempDir()
	git(t, conPropio, "init", "-q")
	escribir(t, conPropio, preCommit, propio)
	s = aquas(t, conPropio, "", "install")
	if s.codigo == 0 || !strings.Contains(s.err, "a mano") || !strings.Contains(s.err, "aquas update") {
		t.Errorf("install con pre-commit propio debe fallar pidiendo unirlo a mano y aquas update; dio %+v", s)
	}

	// core.hooksPath ya fijado a otra carpeta: error que pide unificarlo.
	otro := t.TempDir()
	git(t, otro, "init", "-q")
	git(t, otro, "config", "core.hooksPath", "ganchos")
	s = aquas(t, otro, "", "install")
	if s.codigo == 0 || !strings.Contains(s.err, "core.hooksPath") || !strings.Contains(s.err, "ganchos") {
		t.Errorf("install con core.hooksPath ajeno debe fallar nombrándolo; dio %+v", s)
	}
	if hooksPathDe(otro) != "ganchos" {
		t.Errorf("core.hooksPath ajeno no se pisa; quedó %q", hooksPathDe(otro))
	}

	// Sin .git: install inicia git (E003/F006/S001) y escribe el pre-commit.
	sinGit := t.TempDir()
	if s := aquas(t, sinGit, "", "install"); s.codigo != 0 {
		t.Errorf("install sin .git debe salir 0; dio %+v", s)
	}
	if _, err := os.Stat(filepath.Join(sinGit, filepath.FromSlash(preCommit))); err != nil {
		t.Errorf("install inicia git y escribe el pre-commit: %v", err)
	}
	// Hallazgos A1.2 y A1.3: si luego se borra el repo, el gate del hijo no da OK; nombra el pre-commit y el entorno.
	if err := os.RemoveAll(filepath.Join(sinGit, ".git")); err != nil {
		t.Fatal(err)
	}
	if s := aquas(t, sinGit, "", "verify"); s.codigo == 0 || !strings.Contains(s.out+s.err, "git:pre-commit") { // el entorno ya quedó declarado por install
		t.Errorf("verify en un hijo sin git debe fallar por el pre-commit; dio %+v", s)
	}
}
