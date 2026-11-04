package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// E001/F019/S003: con el tope de la feature agotado, las órdenes de cierre pasan; todo lo demás se frena.

// topeAgotado: repo con E001/F002 aprobada y 15M+1 ponderados gastados.
func topeAgotado(t *testing.T) (raiz, reg string, usar func(string) salida) {
	t.Helper()
	raiz, reg, humano, usar := presupuestoTokens(t)
	humano("apruebo E001/F002 con push")
	anexar(t, reg, lineaUso("m1", 15_000_001, 0, 0, 0, time.Now()))
	return raiz, reg, usar
}

// commitPasaConTopeAgotado (S003, integral): git commit pasa con el tope agotado; leer un archivo, no.
func commitPasaConTopeAgotado(t *testing.T) {
	_, _, usar := topeAgotado(t)
	if s := usar(`git commit -m "cierre"`); s.codigo != 0 {
		t.Errorf("git commit pasa con el tope agotado; dio %+v", s)
	}
	if s := usar("cat x"); s.codigo != 2 {
		t.Errorf("otra orden se frena con el tope agotado; dio %+v", s)
	}
}

// TestCierreConTopeAgotado (E001/F019/S003): las órdenes de cierre pasan el tope agotado de la feature.
// Rojo cuando: commit o push se frenan con el tope agotado.
// Verde cuando: commit, push, verify, seal, memory y state pasan; otra orden se frena.
func TestCierreConTopeAgotado(t *testing.T) {
	raiz, reg, usar := topeAgotado(t)
	raizBash := filepath.ToSlash(raiz)
	pasan := []string{
		"git add -A",
		"git add cmd/aquas/presupuesto.go docs/E001/F019/S003.md",
		`git commit -m "E001/F019/S003: cierre"`,
		`git commit -q -m 'cierre; con punto y coma entre comillas'`,
		"git push",
		"git push -q origin main",
		"aquas verify",
		"./.tmp/aquas.exe verify",
		"go run ./cmd/aquas verify",
		"aquas backlog seal E001/F002/S001",
		`aquas backlog seal E001/F002/S001 ficha:E001 sin_cambio motivo="no cambia"`,
		"aquas backlog state E001/F002/S001 CERRADO",
		"aquas backlog green E001/F002/S001",
		"aquas memory set tarea=x",
		"aquas memory archive",
		`cd "` + raizBash + `" && git push`,
		"cd '" + raizBash + "' && aquas verify",
	}
	for _, c := range pasan {
		if s := usar(c); s.codigo != 0 {
			t.Errorf("%q pasa con el tope agotado; dio %+v", c, s)
		}
	}
	frenadas := []string{
		"",
		"git commitx -m x",
		"git status",
		"git push --force; rm -rf x",
		"git push && ls",
		"git push | tee x",
		"git push\nls",
		"git commit -m $(ls)",
		"git commit -m \"$(ls)\"",
		"git commit -F - <<'FIN'",
		"git add x > y",
		"aquas backlog red E001/F002/S001",
		"aquas verifyx",
		"aquas memory show",
		"echo git push",
		"FOO=1 git push",
		"cd /otro && git push",
		"cd " + raizBash + " && cd /otro && git push",
		"cat x",
		// A1.1: push con una opción que corre un comando, entera, abreviada o entre comillas.
		"git push --receive-pack='echo X; exit 1' .",
		"git push --receive-pack=./x.sh .",
		"git push --rec=./x.sh .",
		"git push --exec=./x.sh .",
		`git push "--exec=./x.sh" .`,
		"git push -o x origin",
		// A2.1–A2.3: llaves, ruta local, push forzado o borrado de rama.
		"git push {--receive-pack=./x.sh,origin}",
		"git push origin {--exec=./x.sh,main}",
		"git push ../otro",
		"git push otro main",
		"git push origin +HEAD:main",
		"git push origin :main",
		"git push origin main:otra",
		"git push origin main extra",
		"git push origin ma*",
		// A3.1: push forzado, borrado o espejo por opción.
		"git push --force origin main",
		"git push -f",
		"git push origin --delete main",
		"git push --mirror",
		"git push --prune origin",
		"git push origin main --force",
	}
	for _, c := range frenadas {
		if s := usar(c); s.codigo != 2 || !strings.Contains(s.err, "tope de la feature") {
			t.Errorf("%q se frena con el tope agotado; dio %+v", c, s)
		}
	}
	// Otra herramienta con el mismo texto no pasa: la lista es de órdenes Bash.
	if s := aquas(t, raiz, llamar(t, "", reg, "Edit", "git push"), "budget", "tool"); s.codigo != 2 {
		t.Errorf("Edit con texto «git push» se frena; dio %+v", s)
	}
	// A1.2: un registro de subagente ilegible frena igual a una orden de cierre (no es el tope agotado).
	roto := filepath.Join(filepath.Dir(reg), "sesion", "subagents", "agent-roto.jsonl")
	if err := os.MkdirAll(roto, 0o755); err != nil { // una carpeta con nombre de registro no se puede leer
		t.Fatal(err)
	}
	if s := usar("git push"); s.codigo != 2 || !strings.Contains(s.err, "agent-roto.jsonl") {
		t.Errorf("registro ilegible: git push se frena nombrándolo; dio %+v", s)
	}
	if err := os.Remove(roto); err != nil {
		t.Fatal(err)
	}
	// Bajo el tope todo pasa: un apruebo nuevo cuenta desde cero.
	_, _, humano, usar2 := presupuestoTokens(t)
	humano("apruebo E001/F002 con push")
	if s := usar2("cat x"); s.codigo != 0 {
		t.Errorf("bajo el tope, cualquier orden pasa; dio %+v", s)
	}
}
