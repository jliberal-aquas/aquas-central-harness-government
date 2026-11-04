package main

import (
	"path/filepath"
	"testing"
)

// TestFrenosPropiosGobernados (integral E001/F011): el freno de este repo nunca bloquea lo que exige un cierre.
// Rojo cuando: memoria fijar se bloquea con el tope desbordado.
// Verde cuando: memoria fijar pasa y otro comando se frena.
// E001/F019/S002: el freno de 60 por turno salió; el tope desbordado es el de la feature («tope 60»).
func TestFrenosPropiosGobernados(t *testing.T) {
	raiz := repoFeature(t)
	if s := aquas(t, raiz, mensajeHumano(t, "apruebo E001/F002 tope 60"), "approval"); s.codigo != 0 {
		t.Fatalf("approval: %+v", s)
	}
	usar := func(comando string) salida {
		t.Helper()
		return aquas(t, raiz, usoComando(t, "Bash", comando), "budget", "tool")
	}
	for i := range 60 {
		if s := usar("git status"); s.codigo != 0 {
			t.Fatalf("dentro del tope; llamada %d dio %+v", i+1, s)
		}
	}
	if s := usar("aquas memoria fijar tarea=x"); s.codigo != 0 {
		t.Errorf("con el tope desbordado, memoria fijar pasa; dio %+v", s)
	}
	if s := usar("git status"); s.codigo != 2 {
		t.Errorf("con el tope desbordado, otro comando se frena; dio %+v", s)
	}
}

// TestFrenoDejaPasarMemoria (E001/F011/S002): memoria fijar suelta pasa sin contar; encadenada u otra orden, no.
// Rojo cuando: memoria fijar se frena con la feature desbordada.
// Verde cuando: memoria fijar suelta pasa y encadenada u otro comando se frena.
func TestFrenoDejaPasarMemoria(t *testing.T) {
	raiz := repoFeature(t)
	if s := aquas(t, raiz, "", "memoria", "fijar", "iniciativa=E001", "feature=F002"); s.codigo != 0 {
		t.Fatalf("memoria fijar: %+v", s)
	}
	turno := func(prompt string) {
		t.Helper()
		if s := aquas(t, raiz, mensajeHumano(t, prompt), "aprobacion"); s.codigo != 0 {
			t.Fatalf("aprobacion: %+v", s)
		}
	}
	usar := func(herramienta, comando string) salida {
		t.Helper()
		return aquas(t, raiz, usoComando(t, herramienta, comando), "presupuesto", "herramienta")
	}
	exentas := []string{
		"aquas memoria fijar tarea=x",
		"./.tmp/aquas.exe memoria fijar tarea=\"a b\" siguiente=y",
		`"` + filepath.ToSlash(filepath.Join(raiz, ".tmp", "aquas.exe")) + `" memoria fijar slice=`,
		`"./.tmp/aquas.exe" memoria fijar slice=`,
		"go run ./cmd/aquas memoria fijar tarea='a; b'",
		`aquas memoria fijar tarea="a; b | c"`,
		"aquas memory set tarea=x",            // E000/F006/S006 A1.2: el nombre nuevo también pasa
		"./.tmp/aquas.exe memory set tarea=x", // ídem con el binario del repo
	}
	frenadas := []struct{ herramienta, comando string }{
		{"Bash", "aquas memoria fijar tarea=x && git push"},
		{"Bash", "aquas memory set tarea=x && git push"},
		{"Bash", "aquas memoria fijar tarea=x; git push"},
		{"Bash", "aquas memoria fijar tarea=x | tee y"},
		{"Bash", "aquas memoria fijar tarea=x\ngit push"},
		{"Bash", "aquas memoria fijar tarea=`git push`"},
		{"Bash", "aquas memoria fijar tarea=$(git push)"},
		{"Bash", "aquas memoria fijar tarea=x & git push"},
		{"Bash", "aquas memoria fijar tarea=<(git push)"},
		{"Bash", "aquas memoria fijar tarea=>(git push)"},
		{"Bash", "aquas memoria fijar tarea=x > .claude/settings.json"},
		{"Bash", `aquas memoria fijar tarea="$(git push)"`},
		{"Bash", "./evil/aquas memoria fijar tarea=x"},
		{"Bash", `"../../evil/.tmp/aquas.exe" memoria fijar tarea=x`},
		{"Bash", `"C:/Users/x/otro/.tmp/aquas.exe" memoria fijar tarea=x`},
		{"Bash", "/.tmp/aquas memoria fijar tarea=x"},
		{"Bash", "FOO=1 aquas memoria fijar tarea=x"},
		{"Bash", "aquas memoria ver"},
		{"Bash", "echo aquas memoria fijar tarea=x"},
		{"Edit", "aquas memoria fijar tarea=x"},
	}

	// Tope de la feature desbordado: memoria fijar pasa y no gasta tope.
	turno("apruebo E001/F002 tope 3")
	for i := range 2 {
		if s := usar("Bash", "git status"); s.codigo != 0 {
			t.Fatalf("dentro del tope de la feature; llamada %d dio %+v", i+1, s)
		}
	}
	for _, c := range exentas {
		if s := usar("Bash", c); s.codigo != 0 {
			t.Errorf("dentro del tope: %q pasa; dio %+v", c, s)
		}
	}
	if s := usar("Bash", "git status"); s.codigo != 0 {
		t.Errorf("la llamada 3 de la feature pasa: memoria fijar no gastó tope; dio %+v", s)
	}
	for _, c := range exentas {
		if s := usar("Bash", c); s.codigo != 0 {
			t.Errorf("feature desbordada: %q pasa; dio %+v", c, s)
		}
	}
	for _, c := range frenadas {
		if s := usar(c.herramienta, c.comando); s.codigo != 2 {
			t.Errorf("feature desbordada: %s %q se frena; dio %+v", c.herramienta, c.comando, s)
		}
	}

	// E001/F019/S002: sin freno de 60 por turno; la exención ante el tope extremo del turno la prueba S004.
}
