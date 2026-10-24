package backlog

import (
	"testing"
	"time"
)

// TestRojoNoHeredaDeCancelado (D000032, E001/F002/S018): solo un slice CERRADO respalda el rojo heredado de su feature.
// Rojo cuando: un slice CANCELADO con rojo de la integral respalda el rojo de la feature.
// Verde cuando: con solo el CANCELADO se niega; con un CANCELADO y un CERRADO, el origen es el CERRADO.
func TestRojoNoHeredaDeCancelado(t *testing.T) {
	raiz := base(t)
	f, _, err := Ver(raiz, "E001/F002")
	if err != nil {
		t.Fatal(err)
	}
	f.Rojo = nil
	pisar(t, raiz, Activo, f)
	slice := func(id, estado string) Ficha {
		return Ficha{ID: id, Tipo: "slice", Titulo: "T", Estado: estado, Prioridad: "P1", Motivo: "m", Padre: "E001/F002", BloqueadoPor: []string{},
			Pruebas: []Prueba{{Nombre: "./p:TestFalla", Intencion: "i"}},
			Rojo:    &Registro{Fecha: "2026-10-22T17:33:24Z", Head: "2231b2b", Resultado: map[string]string{"./p:TestFalla": "falló en el slice"}}}
	}
	ahora := time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC)
	pisar(t, raiz, Archivo, slice("E001/F002/S008", "CANCELADO"))
	if g, err := Rojo(raiz, "E001/F002", ahora); err == nil {
		t.Fatalf("solo un slice CANCELADO: el rojo debió negarse; heredó de %s", g.Rojo.Origen)
	}
	if v, _, _ := Ver(raiz, "E001/F002"); v.Rojo != nil {
		t.Fatal("el rechazo no debió escribir rojo en la ficha")
	}
	pisar(t, raiz, Archivo, slice("E001/F002/S009", "CERRADO"))
	g, err := Rojo(raiz, "E001/F002", ahora)
	if err != nil {
		t.Fatalf("con un slice CERRADO: debió heredarse; dio %v", err)
	}
	if g.Rojo == nil || g.Rojo.Origen != "E001/F002/S009" {
		t.Errorf("el origen debió ser el CERRADO E001/F002/S009: %+v", g.Rojo)
	}
}
