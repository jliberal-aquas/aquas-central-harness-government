package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Páginas que cuentan la historia del problema; no son instrucciones.
var excepcionesGoRunVerificar = map[string]bool{
	"docs/E000/F006/S008.md": true,
	"docs/E000/F006/S010.md": true,
}

// TestInstruccionesSinGoRunVerificar (D000034, E000/F006/S010): canónicos y wiki no mandan `go run … verificar`.
// Rojo cuando: algún canónico o página contiene el comando (da el falso «binario viejo»).
// Verde cuando: ninguno lo contiene, salvo las excepciones.
func TestInstruccionesSinGoRunVerificar(t *testing.T) {
	raiz := filepath.Join("..", "..")
	malo := "go run ./cmd/aquas verificar"
	for _, dir := range []string{"README.md", "gobierno/canonico", "docs"} {
		err := filepath.WalkDir(filepath.Join(raiz, dir), func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(p, ".md") {
				return err
			}
			rel, _ := filepath.Rel(raiz, p)
			rel = filepath.ToSlash(rel)
			if excepcionesGoRunVerificar[rel] {
				return nil
			}
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			for i, l := range strings.Split(string(b), "\n") {
				if strings.Contains(l, malo) {
					t.Errorf("%s:%d manda `%s`; usa `go build -o .tmp/aquas.exe ./cmd/aquas && ./.tmp/aquas.exe verificar`", rel, i+1, malo)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}
