package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// usoComando: entrada del hook PreToolUse para una herramienta con su comando.
func usoComando(herramienta, comando string) string {
	b, _ := json.Marshal(map[string]any{"session_id": "s1", "hook_event_name": "PreToolUse", "tool_name": herramienta,
		"tool_input": map[string]string{"command": comando}})
	return string(b)
}

// TestFrenosPropiosGobernados (integral E001/F011): el freno de este repo nunca bloquea lo que exige un cierre.
// Rojo cuando: memoria fijar se bloquea con el turno desbordado.
// Verde cuando: memoria fijar pasa y otro comando se frena.
func TestFrenosPropiosGobernados(t *testing.T) {
	raiz := t.TempDir()
	usar := func(comando string) salida {
		t.Helper()
		return aquas(t, raiz, usoComando("Bash", comando), "presupuesto", "herramienta")
	}
	for i := range 60 {
		if s := usar("git status"); s.codigo != 0 {
			t.Fatalf("dentro del freno del turno; llamada %d dio %+v", i+1, s)
		}
	}
	if s := usar("aquas memoria fijar tarea=x"); s.codigo != 0 {
		t.Errorf("con el turno desbordado, memoria fijar pasa; dio %+v", s)
	}
	if s := usar("git status"); s.codigo != 2 {
		t.Errorf("con el turno desbordado, otro comando se frena; dio %+v", s)
	}
}

// TestFrenoDejaPasarMemoria (E001/F011/S002): memoria fijar suelta pasa sin contar; encadenada u otra orden, no.
// Rojo cuando: memoria fijar se frena con el turno o la feature desbordados.
// Verde cuando: memoria fijar suelta pasa y encadenada u otro comando se frena.
func TestFrenoDejaPasarMemoria(t *testing.T) {
	raiz := repoFeature(t)
	if s := aquas(t, raiz, "", "memoria", "fijar", "iniciativa=E001", "feature=F002"); s.codigo != 0 {
		t.Fatalf("memoria fijar: %+v", s)
	}
	turno := func(prompt string) {
		t.Helper()
		for _, orden := range [][]string{{"aprobacion"}, {"presupuesto", "prompt"}} {
			if s := aquas(t, raiz, mensajeHumano(prompt), orden...); s.codigo != 0 {
				t.Fatalf("%v: %+v", orden, s)
			}
		}
	}
	usar := func(herramienta, comando string) salida {
		t.Helper()
		return aquas(t, raiz, usoComando(herramienta, comando), "presupuesto", "herramienta")
	}
	exentas := []string{
		"aquas memoria fijar tarea=x",
		"./.tmp/aquas.exe memoria fijar tarea=\"a b\" siguiente=y",
		`"` + filepath.ToSlash(filepath.Join(raiz, ".tmp", "aquas.exe")) + `" memoria fijar slice=`,
		`"./.tmp/aquas.exe" memoria fijar slice=`,
		"go run ./cmd/aquas memoria fijar tarea='a; b'",
		`aquas memoria fijar tarea="a; b | c"`,
	}
	frenadas := []struct{ herramienta, comando string }{
		{"Bash", "aquas memoria fijar tarea=x && git push"},
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
	turno("apruebo E001/F002 tope 2")
	for i := range 2 {
		if s := usar("Bash", "git status"); s.codigo != 0 {
			t.Fatalf("dentro del tope de la feature; llamada %d dio %+v", i+1, s)
		}
	}
	_, antes := cuentaFeature(t, raiz)
	for _, c := range exentas {
		if s := usar("Bash", c); s.codigo != 0 {
			t.Errorf("feature desbordada: %q pasa; dio %+v", c, s)
		}
	}
	if _, despues := cuentaFeature(t, raiz); despues != antes {
		t.Errorf("memoria fijar no gasta tope de la feature: %d → %d", antes, despues)
	}
	for _, c := range frenadas {
		if s := usar(c.herramienta, c.comando); s.codigo != 2 {
			t.Errorf("feature desbordada: %s %q se frena; dio %+v", c.herramienta, c.comando, s)
		}
	}

	// Turno desbordado: memoria fijar pasa y no cuenta en el turno.
	time.Sleep(1100 * time.Millisecond) // la fecha de aprobación va en segundos
	turno("apruebo E001/F002 tope 1000")
	for i := range 59 {
		if s := usar("Bash", "git status"); s.codigo != 0 {
			t.Fatalf("dentro del freno del turno; llamada %d dio %+v", i+1, s)
		}
	}
	for _, c := range exentas {
		if s := usar("Bash", c); s.codigo != 0 {
			t.Errorf("memoria fijar no cuenta en el turno: %q pasa; dio %+v", c, s)
		}
	}
	if s := usar("Bash", "git status"); s.codigo != 0 {
		t.Errorf("la llamada 60 del turno pasa: memoria fijar no contó; dio %+v", s)
	}
	for _, c := range exentas {
		if s := usar("Bash", c); s.codigo != 0 {
			t.Errorf("turno desbordado: %q pasa; dio %+v", c, s)
		}
	}
	if s := usar("Bash", "git status"); s.codigo != 2 || !strings.Contains(s.err, "turno") {
		t.Errorf("turno desbordado: otro comando se frena; dio %+v", s)
	}
}
