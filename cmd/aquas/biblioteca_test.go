package main

import (
	"strings"
	"testing"
)

// E001/F003/S001: `aquas biblioteca leer` lee el pedido JSON por stdin.
func TestBibliotecaCLI(t *testing.T) {
	raiz := t.TempDir()
	escribir(t, raiz, "a.txt", "uno\ndos\ntres\n")
	s := aquas(t, raiz, `{"rutas":["a.txt:2-3"]}`, "biblioteca", "leer")
	if s.codigo != 0 || !strings.Contains(s.out, "== a.txt · 2-3 de 3\n2\tdos\n3\ttres\n") || !strings.Contains(s.out, "== devueltos") {
		t.Fatalf("lectura por stdin: %+v", s)
	}
	if s := aquas(t, raiz, `no es json`, "biblioteca", "leer"); s.codigo != 64 {
		t.Fatalf("pedido ilegible: exit 64; dio %+v", s)
	}
}
