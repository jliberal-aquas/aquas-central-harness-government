package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const binPreCommit = "./.tmp/aquas-precommit.exe"

// problemaPreCommit: por qué el script del pre-commit no cumple D000034 ("" si cumple).
// Ignora comentarios: un `go build` comentado no construye nada.
func problemaPreCommit(script string) string {
	var cod []string
	for _, l := range strings.Split(script, "\n") {
		if l = strings.TrimSpace(l); l != "" && !strings.HasPrefix(l, "#") {
			cod = append(cod, l)
		}
	}
	tiene := func(s string) bool {
		for _, l := range cod {
			if l == s {
				return true
			}
		}
		return false
	}
	for _, l := range cod {
		if strings.Contains(l, "go run") {
			return "usa `go run`: reutiliza un ejecutable de la caché de Go con fecha vieja"
		}
	}
	if !tiene("go build -o " + binPreCommit + " ./cmd/aquas || exit 1") {
		return "debe construir con `go build -o " + binPreCommit + " ./cmd/aquas || exit 1` (el fallo de compilación bloquea)"
	}
	if !tiene("exec " + binPreCommit + " verificar") {
		return "debe hacer `exec " + binPreCommit + " verificar`"
	}
	return ""
}

// TestPreCommitNoUsaGoRun (D000034, E000/F006/S008): el pre-commit construye aquas a .tmp y corre ese binario.
// Rojo cuando: el script contiene `go run` (ejecutable de la caché con fecha vieja → falso «binario viejo»).
// Verde cuando: el script hace `go build -o ./.tmp/… || exit 1` y luego `exec` de ese binario.
func TestPreCommitNoUsaGoRun(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", ".githooks", "pre-commit"))
	if err != nil {
		t.Fatal(err)
	}
	if p := problemaPreCommit(string(b)); p != "" {
		t.Fatal("el pre-commit " + p)
	}
	// La revisión no se deja engañar por comentarios ni por quitar `|| exit 1`.
	for nombre, malo := range map[string]string{
		"comentado": "# go build -o " + binPreCommit + " ./cmd/aquas || exit 1\n# exec " + binPreCommit + " verificar\n",
		"sin exit":  "go build -o " + binPreCommit + " ./cmd/aquas\nexec " + binPreCommit + " verificar\n",
		"go run":    "go build -o " + binPreCommit + " ./cmd/aquas || exit 1\nexec " + binPreCommit + " verificar\ngo run ./cmd/aquas verificar\n",
	} {
		if problemaPreCommit(malo) == "" {
			t.Errorf("script %s debió rechazarse", nombre)
		}
	}
}
