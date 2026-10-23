package backlog

import (
	"strings"
	"testing"
	"time"
)

// TestFeatureHeredaRojoDeSuSlice: si la integral de una feature ya no puede caer, la feature toma el rojo
// que un slice cerrado suyo registró para esa misma prueba, y nombra de dónde vino (E001/F002/S003/X02).
// Rojo cuando: el rojo se niega pese al slice cerrado con esa prueba en rojo, o se hereda de un slice
// con otra prueba, o el rojo heredado no nombra el slice de origen.
// Verde cuando: sin slice que la respalde se niega; con él, la feature guarda ese rojo y su origen.
func TestFeatureHeredaRojoDeSuSlice(t *testing.T) {
	raiz := base(t)
	f, _, err := Ver(raiz, "E001/F002")
	if err != nil {
		t.Fatal(err)
	}
	f.Rojo = nil
	pisar(t, raiz, Activo, f)
	cerrado := func(id, prueba string) Ficha {
		return Ficha{ID: id, Tipo: "slice", Titulo: "T", Estado: "CERRADO", Prioridad: "P1", Motivo: "m", Padre: "E001/F002", BloqueadoPor: []string{},
			Pruebas: []Prueba{{Nombre: prueba, Intencion: "i"}},
			Rojo:    &Registro{Fecha: "2026-10-22T17:33:24Z", Head: "2231b2b", Resultado: map[string]string{prueba: "falló en el slice"}}}
	}
	ahora := time.Date(2026, 10, 23, 0, 0, 0, 0, time.UTC)
	pisar(t, raiz, Archivo, cerrado("E001/F002/S008", "./p:TestOtra"))
	if _, err := Rojo(raiz, "E001/F002", ahora); err == nil {
		t.Fatal("solo un slice con otra prueba: el rojo debió negarse")
	}
	pisar(t, raiz, Archivo, cerrado("E001/F002/S009", "./p:TestFalla"))
	g, err := Rojo(raiz, "E001/F002", ahora)
	if err != nil {
		t.Fatalf("slice cerrado con esa prueba en rojo: debió heredarse; dio %v", err)
	}
	if g.Rojo == nil || g.Rojo.Origen != "E001/F002/S009" || g.Rojo.Head != "2231b2b" ||
		!strings.Contains(g.Rojo.Resultado["./p:TestFalla"], "falló en el slice") {
		t.Errorf("rojo heredado incompleto: %+v", g.Rojo)
	}
}
