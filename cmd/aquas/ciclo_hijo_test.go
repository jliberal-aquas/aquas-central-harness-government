package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCicloDeVidaDelHijo (integral E002): un hijo cumple su ciclo completo: install, build, update y report.
// Rojo cuando: algún paso sale con error o no deja su efecto visible.
// Verde cuando: los cuatro pasos salen 0 sobre un hijo de prueba y dejan su efecto.
func TestCicloDeVidaDelHijo(t *testing.T) {
	dir := t.TempDir()
	leer := func(rel string) string {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	paso := func(args ...string) {
		t.Helper()
		if s := aquas(t, dir, "", args...); s.codigo != 0 {
			t.Fatalf("%s: exit %d\n%s%s", args[0], s.codigo, s.out, s.err)
		}
	}

	// install: deja la central con su lock y el CLAUDE.md armado.
	paso("install")
	if !strings.Contains(leer("CLAUDE.md"), "PENDIENTE") {
		t.Error("install: CLAUDE.md debió armarse con los huecos PENDIENTE")
	}
	// E001/F007/S003: install por la CLI deja /backlog con el aquas instalado.
	if s := leer(".claude/skills/backlog/SKILL.md"); !strings.Contains(s, "aquas backlog tabla") || strings.Contains(s, "aquas.sh") {
		t.Errorf("install: /backlog del hijo debió correr el aquas instalado:\n%s", s)
	}

	// build: un hueco escrito entra al CLAUDE.md.
	mision := "- Misión de la prueba del ciclo.\n"
	if err := os.WriteFile(filepath.Join(dir, "gobierno", "canonico", "huecos", "mision.md"), []byte(mision), 0o644); err != nil {
		t.Fatal(err)
	}
	paso("build", dir)
	if !strings.Contains(leer("CLAUDE.md"), "Misión de la prueba del ciclo.") {
		t.Error("build: CLAUDE.md debió llevar el hueco escrito")
	}

	// update: trae la central del binario (el lock deja la versión vieja) sin pisar el hueco.
	var l map[string]any
	if err := json.Unmarshal([]byte(leer(".aquas/lock.json")), &l); err != nil {
		t.Fatal(err)
	}
	l["version"] = "vieja-de-la-prueba"
	b, _ := json.Marshal(l)
	if err := os.WriteFile(filepath.Join(dir, ".aquas", "lock.json"), b, 0o644); err != nil {
		t.Fatal(err)
	}
	paso("update")
	if strings.Contains(leer(".aquas/lock.json"), "vieja-de-la-prueba") {
		t.Error("update: el lock debió pasar a la versión del binario")
	}
	if !strings.Contains(leer("CLAUDE.md"), "Misión de la prueba del ciclo.") {
		t.Error("update: no debió pisar el hueco escrito")
	}

	// report: deja un borrador de issue con lo que pasó, sin enviarlo.
	paso("report", dir, "falla del ciclo", "correr la prueba")
	if !strings.Contains(leer(".aquas/reporte.md"), "falla del ciclo") {
		t.Error("report: el borrador debió contar qué pasó")
	}
}
