package backlog

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// TestTablaEnOrden (E001/F007/S002): la tabla trae columnas fijas, solo fichas activas y el orden de Activas.
// Rojo cuando: falta una columna, sale una ficha archivada o el orden difiere del de siguiente.
// Verde cuando: encabezado fijo, una fila por ficha activa en el orden de Activas, celdas vacías con «—».
func TestTablaEnOrden(t *testing.T) {
	raiz := t.TempDir()
	if got := Tabla(nil); strings.Count(got, "\n") != 1 || !strings.Contains(got, "Iniciativa") {
		t.Fatalf("backlog vacío: solo el encabezado; dio %q", got)
	}
	nueva := func(id, prio string) {
		t.Helper()
		if _, err := Nuevo(raiz, []string{"id=" + id, "titulo=titulo " + id, "prioridad=" + prio, "motivo=prueba",
			"que_hace=hace algo", "relacion=nueva", "prueba=./internal/x:TestX=valida algo|rojo cuando falla|verde cuando pasa"}); err != nil {
			t.Fatal(err)
		}
	}
	nueva("E009", "P2")
	nueva("E009/F001", "P2")
	nueva("E009/F001/S001", "P3")
	nueva("E009/F001/S002", "P1")
	nueva("E009/F001/S003", "P1")
	if _, err := Estado(raiz, "E009/F001/S003", "CANCELADO"); err != nil {
		t.Fatal(err)
	}
	// Un slice EN_CURSO sale y va primero (se escribe directo: el paso real exige aprobación del humano).
	s1, _, err := Ver(raiz, "E009/F001/S001")
	if err != nil {
		t.Fatal(err)
	}
	s1.Estado = "EN_CURSO"
	if err := escribir(raiz, Activo, s1); err != nil {
		t.Fatal(err)
	}
	fs, err := Activas(raiz)
	if err != nil {
		t.Fatal(err)
	}
	lineas := strings.Split(strings.TrimSuffix(Tabla(fs), "\n"), "\n")
	campos := strings.Fields(lineas[0])
	if strings.Join(campos, " ") != "Iniciativa Feature Slice Estado Prior. Descripción" {
		t.Fatalf("encabezado: %q", lineas[0])
	}
	if len(lineas) != 1+len(fs) || len(fs) != 4 {
		t.Fatalf("quiero encabezado + 4 filas activas; dio %d fichas:\n%s", len(fs), strings.Join(lineas, "\n"))
	}
	col := func(l, s string) int { return utf8.RuneCountInString(l[:max(strings.Index(l, s), 0)]) }
	ancho := col(lineas[0], "Descripción")
	for i, f := range fs {
		l := lineas[i+1]
		if !strings.Contains(l, f.Titulo) || !strings.Contains(l, f.Estado) || !strings.Contains(l, f.Prioridad) {
			t.Errorf("fila %d no es %s: %q", i+1, f.ID, l)
		}
		if col(l, f.Titulo) != ancho {
			t.Errorf("fila %d desalineada: %q", i+1, l)
		}
		if strings.Contains(l, "S003") {
			t.Errorf("salió el slice cancelado: %q", l)
		}
	}
	// Celdas vacías: la feature no tiene slice; la iniciativa no tiene feature ni slice.
	for _, l := range lineas[1:] {
		c := strings.Fields(l)
		switch {
		case strings.Contains(l, "titulo E009/F001/"):
		case strings.Contains(l, "titulo E009/F001"):
			if c[0] != "E009" || c[1] != "F001" || c[2] != "—" {
				t.Errorf("fila de feature: %q", l)
			}
		case strings.Contains(l, "titulo E009"):
			if c[0] != "E009" || c[1] != "—" || c[2] != "—" {
				t.Errorf("fila de iniciativa: %q", l)
			}
		}
	}
	// Orden: el slice EN_CURSO (S001, P3) va primero; después el P1, como en siguiente.
	if !strings.Contains(lineas[1], "S001") || !strings.Contains(lineas[1], "EN_CURSO") || !strings.Contains(lineas[2], "S002") {
		t.Errorf("orden: S001 EN_CURSO primero, luego S002:\n%s", strings.Join(lineas, "\n"))
	}
}
