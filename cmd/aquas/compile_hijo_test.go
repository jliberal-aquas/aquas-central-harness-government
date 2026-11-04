package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCompileEnHijo (E003/F006/X01, D000121): en un hijo, aquas compile rehace los bloques GENERADO de la wiki sin la config de la central.
// Rojo cuando: compile falla en el hijo o el gate sigue pidiendo el bloque GENERADO tras compile.
// Verde cuando: compile sale 0 en el hijo y el gate ya no reclama el bloque.
func TestCompileEnHijo(t *testing.T) {
	dir := hijoNuevo(t)
	escribir(t, dir, "gobierno/config/pruebas.json", `{"version": 1, "comando": "npx vitest run {ruta} -t {nombre}"}`)
	if s := aquas(t, dir, "", "backlog", "new", "id=E001", "titulo=POC", "prioridad=P0", "motivo=m", "relacion=nueva",
		"que_hace=hace algo", "prueba=e2e/poc.test.ts:aprendiz_completa=valida la ruta|r|v"); s.codigo != 0 {
		t.Fatalf("backlog new en el hijo: %+v", s)
	}
	escribir(t, dir, "docs/README.md", "# Inicio\n\n## Navegación\n\n- nada\n")
	if s := aquas(t, dir, "", "verify"); !strings.Contains(s.out+s.err, "bloque GENERADO") {
		t.Fatalf("sin compile, el gate del hijo debe pedir el bloque GENERADO; dio %+v", s)
	}
	if err := os.Remove(filepath.Join(dir, "CLAUDE.md")); err != nil {
		t.Fatal(err)
	}
	if s := aquas(t, dir, "", "compile"); s.codigo != 0 {
		t.Fatalf("compile en el hijo debe salir 0; dio %+v", s)
	}
	if !existe(dir, "CLAUDE.md") { // hallazgo A2.2: compile arma CLAUDE.md como build
		t.Error("compile en el hijo debe rehacer CLAUDE.md")
	}
	if s := aquas(t, dir, "", "verify"); strings.Contains(s.out+s.err, "bloque GENERADO") || strings.Contains(s.out+s.err, "desfasados") {
		t.Errorf("tras compile, el gate del hijo no debe reclamar el bloque; dio %+v", s)
	}
	if s := aquas(t, dir, "", "compile"); s.codigo != 0 {
		t.Errorf("compile repetido en el hijo debe salir 0; dio %+v", s)
	}
}
