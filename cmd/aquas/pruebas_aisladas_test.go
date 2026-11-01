package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// paquetesSinLimite: carpetas con pruebas bajo raiz (sin testdata) sin un TestMain que use el literal
// "GIT_CEILING_DIRECTORIES"; git puede correr directo o desde el código que la prueba llama. Lee el código Go, no el texto: un string que lo nombre no cuenta.
func paquetesSinLimite(t *testing.T, raiz string) []string {
	t.Helper()
	conPruebas, conLimite := map[string]bool{}, map[string]bool{}
	err := filepath.WalkDir(raiz, func(ruta string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && (d.Name() == ".tmp" || d.Name() == ".git" || d.Name() == "testdata") {
			return filepath.SkipDir
		}
		if d.IsDir() || !strings.HasSuffix(ruta, "_test.go") {
			return nil
		}
		archivo, err := parser.ParseFile(token.NewFileSet(), ruta, nil, 0)
		if err != nil {
			return err
		}
		dir := filepath.Dir(ruta)
		conPruebas[dir] = true // A1 hallazgo 1: git también corre por código de producción; todo paquete con pruebas lleva el límite
		ast.Inspect(archivo, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.FuncDecl: // TestMain con el literal "GIT_CEILING_DIRECTORIES"
				if x.Name.Name == "TestMain" && x.Body != nil {
					ast.Inspect(x.Body, func(m ast.Node) bool {
						if lit, ok := m.(*ast.BasicLit); ok && lit.Value == `"GIT_CEILING_DIRECTORIES"` {
							conLimite[dir] = true
						}
						return true
					})
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	var faltan []string
	for dir := range conPruebas {
		if !conLimite[dir] {
			rel, _ := filepath.Rel(raiz, dir)
			faltan = append(faltan, filepath.ToSlash(rel))
		}
	}
	sort.Strings(faltan)
	return faltan
}

// TestPruebasNoVenLaCentral (E000/F009, D000050): todo paquete con pruebas fija GIT_CEILING_DIRECTORIES en su
// TestMain, para que git, directo o desde el código probado, no suba desde .tmp al repo de la central.
// Rojo cuando: un paquete con pruebas no fija el límite.
// Verde cuando: todos lo fijan; en la muestra se detecta el paquete sin límite y no el que solo lo nombra en un texto.
func TestPruebasNoVenLaCentral(t *testing.T) {
	muestra := t.TempDir()
	git := "package x\n\nimport \"os/exec\"\n\nfunc f() { _ = exec.Command(\"git\", \"log\") }\n"
	escribir(t, muestra, "a/a_test.go", git)
	escribir(t, muestra, "b/b_test.go", git)
	escribir(t, muestra, "b/main_test.go", "package x\n\nimport (\n\t\"os\"\n\t\"testing\"\n)\n\nfunc TestMain(m *testing.M) {\n\t_ = os.Setenv(\"GIT_CEILING_DIRECTORIES\", os.TempDir())\n\tos.Exit(m.Run())\n}\n")
	escribir(t, muestra, "c/c_test.go", git)
	escribir(t, muestra, "c/texto_test.go", "package x\n\nfunc TestMain() { _ = \"func TestMain( GIT_CEILING_DIRECTORIES\" }\n")
	escribir(t, muestra, "d/d_test.go", "package x\n")
	escribir(t, muestra, "d/testdata/e/e_test.go", "package x\n")
	if got := paquetesSinLimite(t, muestra); strings.Join(got, ",") != "a,c,d" {
		t.Fatalf("muestra: «a», «c» (solo lo nombra en un texto) y «d» (git indirecto) van sin límite; testdata no cuenta; dio %v", got)
	}
	raiz, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	if got := paquetesSinLimite(t, raiz); len(got) > 0 {
		t.Errorf("paquetes con pruebas sin GIT_CEILING_DIRECTORIES en su TestMain: %v", got)
	}
}
