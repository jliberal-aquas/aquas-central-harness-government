package main

import (
	"strings"
	"testing"
)

// E001/F002/S009: cerrar un slice exige que su memoria esté archivada, apunte o no la memoria a él (X01).
// Orden correcto: memoria fijar → memoria archivar → backlog estado … CERRADO.
func TestCerrarExigeMemoriaArchivada(t *testing.T) {
	raiz := t.TempDir()
	for _, args := range [][]string{
		{"nuevo", "id=E001", "titulo=G", "prioridad=P1", "motivo=m", "relacion=nueva", "que_hace=hace algo", "prueba=./p:TestA=a|r|v"},
		{"nuevo", "id=E001/F002", "titulo=B", "prioridad=P1", "motivo=m", "relacion=nueva", "que_hace=hace algo", "prueba=./p:TestA=a|r|v"},
		{"nuevo", "id=E001/F002/S003", "titulo=R", "prioridad=P1", "motivo=m", "relacion=nueva", "que_hace=hace algo", "prueba=./p:TestA=a|r|v"},
		{"nuevo", "id=E001/F002/S004", "titulo=Q", "prioridad=P1", "motivo=m", "relacion=nueva", "que_hace=hace algo", "prueba=./p:TestA=a|r|v"},
		{"estado", "E001/F002/S003", "EN_CURSO"},
	} {
		if args[0] == "estado" {
			aprobarCLI(t, raiz, args[1])
		}
		if s := aquas(t, raiz, "", append([]string{"backlog"}, args...)...); s.codigo != 0 {
			t.Fatalf("backlog %v: %+v", args, s)
		}
		rojoIntegral(t, raiz, args)
	}
	fijar := func(pares ...string) {
		t.Helper()
		if s := aquas(t, raiz, "", append([]string{"memoria", "fijar"}, pares...)...); s.codigo != 0 {
			t.Fatalf("memoria fijar %v: %+v", pares, s)
		}
	}
	cerrar := func(id string) salida {
		t.Helper()
		sellarCLI(t, raiz, id) // desde S004 cerrar exige el sello marcado
		return aquas(t, raiz, "", "backlog", "estado", id, "CERRADO")
	}

	// 1 · Memoria en el slice, sin archivar: CERRADO se niega por la memoria y dice cómo salir.
	fijar("iniciativa=E001", "feature=F002", "slice=S003", "tarea=t")
	if s := cerrar("E001/F002/S003"); s.codigo != 1 || !strings.Contains(s.err, "memoria apunta a E001/F002/S003") ||
		!strings.Contains(s.err, "aquas memoria archivar") {
		t.Errorf("memoria sin archivar: CERRADO da exit 1, nombra el slice y pide archivar; dio %+v", s)
	}

	// 2 · Cancelar no exige memoria archivada (decisión: el bloqueo es solo para CERRADO).
	if s := aquas(t, raiz, "", "backlog", "estado", "E001/F002/S003", "CANCELADO"); s.codigo != 0 {
		t.Fatalf("CANCELADO con memoria sin archivar no se bloquea; dio %+v", s)
	}
	aprobarCLI(t, raiz, "E001/F002/S004")
	if s := aquas(t, raiz, "", "backlog", "estado", "E001/F002/S004", "EN_CURSO"); s.codigo != 0 {
		t.Fatalf("EN_CURSO S004: %+v", s)
	}

	// 3 · Memoria en otro foco y sin archivo del slice: también se bloquea (E001/F002/S009/X01, D000010).
	fijar("slice=", "tarea=otra")
	if s := cerrar("E001/F002/S004"); s.codigo != 1 || !strings.Contains(s.err, "E001/F002/S004 sin memoria archivada") {
		t.Errorf("memoria en otro foco: CERRADO exige el archivo de memoria del slice; dio %+v", s)
	}

	// 4 · Memoria archivada: tampoco se bloquea por memoria.
	fijar("slice=S004", "tarea=t")
	if s := aquas(t, raiz, "", "memoria", "archivar"); s.codigo != 0 {
		t.Fatalf("memoria archivar: %+v", s)
	}
	fijar("slice=S004", "tarea=t2") // la memoria vuelve a apuntar al slice, ya archivado
	if s := cerrar("E001/F002/S004"); strings.Contains(s.err, "memoria apunta") || !strings.Contains(s.err, "rojo") {
		t.Errorf("memoria archivada: el cierre no se bloquea por memoria; dio %+v", s)
	}
}
