package backlog

import (
	"testing"
	"time"
)

// TestArregloCuelgaDeSlice: un arreglo vive bajo su slice con ID padre/X01, no como slice nuevo.
// Rojo cuando: se acepta un arreglo sin padre o sin título, el ID no sigue X01, X02…, el slice absorbido sigue vivo o el gate no ve un ID fuera de secuencia.
// Verde cuando: el arreglo queda en la ficha del slice con su ID, hereda las pruebas del slice absorbido y este desaparece.
func TestArregloCuelgaDeSlice(t *testing.T) {
	raiz := t.TempDir()
	pisar(t, raiz, Archivo, Ficha{ID: "E002/F001/S002", Tipo: "slice", Titulo: "Huecos", Estado: "CERRADO", Prioridad: "P1", Motivo: "m", Padre: "E002/F001", BloqueadoPor: []string{}, Secuencia: 1})
	pisar(t, raiz, Archivo, Ficha{ID: "E002/F001/S005", Tipo: "slice", Titulo: "Título del hijo", Estado: "CERRADO", Prioridad: "P1", Motivo: "m", Padre: "E002/F001", BloqueadoPor: []string{}, Secuencia: 1,
		Pruebas: []Prueba{{Nombre: "./x:TestT", Intencion: "t"}}})
	ahora := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	if _, err := Arreglar(raiz, "E002/F001/S009", "t", "m", "", ahora); err == nil {
		t.Error("arreglo sin padre: debió fallar")
	}
	if _, err := Arreglar(raiz, "E002/F001/S002", "", "m", "", ahora); err == nil {
		t.Error("arreglo sin título: debió fallar")
	}
	a, err := Arreglar(raiz, "E002/F001/S002", "Título del hijo", "era un defecto de S002", "E002/F001/S005", ahora)
	if err != nil {
		t.Fatal(err)
	}
	if a.ID != "E002/F001/S002/X01" || a.Antes != "E002/F001/S005" || len(a.Pruebas) != 1 {
		t.Errorf("arreglo mal armado: %+v", a)
	}
	if _, _, err := Ver(raiz, "E002/F001/S005"); err == nil {
		t.Error("el slice absorbido debió desaparecer")
	}
	if b, _ := Arreglar(raiz, "E002/F001/S002", "otro", "m", "", ahora); b.ID != "E002/F001/S002/X02" {
		t.Errorf("segundo arreglo: %q", b.ID)
	}
	g, _, _ := Ver(raiz, "E002/F001/S002")
	if len(g.Arreglos) != 2 {
		t.Fatalf("la ficha no guardó los arreglos: %+v", g.Arreglos)
	}
	g.Arreglos[0].ID = "E002/F001/S002/X07"
	if len(problemasArreglos("r", g)) == 0 {
		t.Error("el gate debió señalar el ID fuera de secuencia")
	}
}
