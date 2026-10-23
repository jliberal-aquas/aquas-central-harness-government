package main

import (
	"strings"
	"testing"
)

// TestTablaBacklog: integral de E001/F007. aquas backlog tabla muestra cada ficha en una fila
// con su iniciativa, feature, slice y descripción corta, sin navegar entre archivos.
// Rojo cuando: el comando falla o una fila no trae iniciativa, feature, slice y título juntos.
// Verde cuando: sale 0, con encabezado y una fila por ficha que reúne las cuatro columnas.
func TestTablaBacklog(t *testing.T) {
	raiz := t.TempDir()
	for _, a := range [][]string{
		{"id=E009", "titulo=Iniciativa de prueba"},
		{"id=E009/F001", "titulo=Feature de prueba"},
		{"id=E009/F001/S001", "titulo=Slice de prueba"},
	} {
		args := append([]string{"backlog", "nuevo"}, a...)
		args = append(args, "prioridad=P2", "motivo=m", "relacion=nueva", "prueba=./p:TestA=a|r|v")
		if s := aquas(t, raiz, "", args...); s.codigo != 0 {
			t.Fatalf("aquas %v: exit %d\n%s%s", args, s.codigo, s.out, s.err)
		}
		if a[0] == "id=E009/F001" {
			rojoIntegral(t, raiz, []string{"nuevo", "id=E009/F001"})
		}
	}
	s := aquas(t, raiz, "", "backlog", "tabla")
	if s.codigo != 0 {
		t.Fatalf("aquas backlog tabla: exit %d; quiero 0\n%s%s", s.codigo, s.out, s.err)
	}
	lineas := strings.Split(s.out, "\n")
	for _, col := range []string{"Iniciativa", "Feature", "Slice", "Descripción"} {
		if !strings.Contains(lineas[0], col) {
			t.Errorf("encabezado sin %q: %q", col, lineas[0])
		}
	}
	fila := ""
	for _, l := range lineas {
		if strings.Contains(l, "Slice de prueba") {
			fila = l
		}
	}
	for _, v := range []string{"E009", "F001", "S001"} {
		if !strings.Contains(fila, v) {
			t.Errorf("fila del slice sin %q: %q", v, fila)
		}
	}
}
