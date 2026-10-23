package backlog

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestArregloRegistraSusPruebas: un arreglo registra sus pruebas y pasa rojo y verde con su propio ID, aunque su slice esté cerrado.
// Rojo cuando: pruebas, rojo o verde rechazan el ID padre/Xnn, guardan en la ficha padre, o el verde pasa sin rojo o con la prueba fallando.
// Verde cuando: el arreglo guarda su prueba, su rojo y su verde; la ficha padre conserva los suyos.
func TestArregloRegistraSusPruebas(t *testing.T) {
	raiz := modulo(t)
	fichaCerrada(t, raiz)
	if _, err := Arreglar(raiz, "E001/F001/S001", "t", "m", "", hora); err != nil {
		t.Fatal(err)
	}
	id := "E001/F001/S001/X01"
	if _, err := Pruebas(raiz, id, []string{"./p:TestBandera=pide el arreglo|cae sin arreglo|pasa con arreglo"}); err != nil {
		t.Fatalf("pruebas del arreglo: %v", err)
	}
	if _, err := Verde(raiz, id, hora); err == nil {
		t.Error("verde sin rojo: debió fallar")
	}
	if _, err := Rojo(raiz, id, hora); err != nil {
		t.Fatalf("rojo del arreglo: %v", err)
	}
	if _, err := Verde(raiz, id, hora); err == nil {
		t.Error("verde con la prueba fallando: debió fallar")
	}
	if err := os.WriteFile(filepath.Join(raiz, "p", "arreglado"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Verde(raiz, id, hora.Add(time.Hour)); err != nil {
		t.Fatalf("verde del arreglo: %v", err)
	}
	f, _, err := Ver(raiz, "E001/F001/S001")
	if err != nil {
		t.Fatal(err)
	}
	a := f.Arreglos[0]
	if len(a.Pruebas) != 1 || a.Rojo == nil || a.Rojo.Resultado["./p:TestBandera"] == "" || a.Verde == nil {
		t.Errorf("arreglo sin prueba, rojo o verde: %+v", a)
	}
	if len(f.Pruebas) != 1 || f.Pruebas[0].Nombre != "./x:TestA" || f.Rojo == nil || f.Rojo.Fecha != "2026-10-01T00:00:00Z" {
		t.Errorf("la ficha padre cambió: %+v", f)
	}
}
