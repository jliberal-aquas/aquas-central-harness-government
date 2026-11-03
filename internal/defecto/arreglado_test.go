package defecto

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDefectoArreglado (E000/F007/S003, D000109): `aquas defect fix` pasa un defecto a arreglado con su prueba real.
// Rojo cuando: el arreglo no se puede actualizar o acepta una prueba inexistente.
// Verde cuando: fix guarda arreglo y prueba real y rechaza la inexistente.
func TestDefectoArreglado(t *testing.T) {
	raiz := t.TempDir()
	semilla(t, raiz)
	ficha(t, raiz, "activo", "E001/F002/S001", "BORRADOR")
	pendiente := campos("con-pendiente")
	pendiente["prueba"] = "pendiente:E001/F002/S001"
	d, err := Nuevo(raiz, pendiente)
	if err != nil {
		t.Fatal(err)
	}
	escribirArchivo(t, raiz, Carpeta+"/D000002.json", `{"id":"D000002","firma":"viejo","que":"q","causa":"c","arreglo":"a","como_me_cercioro":"x","ocurrencias":[]}`)

	malos := []struct {
		caso, id, arreglo, prueba, dice string
	}{
		{"id mal escrito", "D12", "hecho", "./p:TestP", "formato D000001"},
		{"defecto inexistente", "D000099", "hecho", "./p:TestP", "D000099"},
		{"arreglo vacío", d.ID, "  ", "./p:TestP", "arreglo"},
		{"prueba ausente", d.ID, "hecho", "", "falta el campo prueba"},
		{"prueba inexistente", d.ID, "hecho", "./p:TestNoHay", "no existe"},
		{"pendiente no vale al arreglar", d.ID, "hecho", "pendiente:E001/F002/S001", "prueba real"},
	}
	antes, _ := os.ReadFile(filepath.Join(raiz, Carpeta, d.ID+".json"))
	for _, c := range malos {
		if _, err := Arreglar(raiz, c.id, c.arreglo, c.prueba); err == nil || !strings.Contains(err.Error(), c.dice) {
			t.Errorf("%s: debía fallar mencionando %q; dio %v", c.caso, c.dice, err)
		}
	}
	// Fallo a mitad de camino: un rechazo no cambia el archivo.
	if despues, _ := os.ReadFile(filepath.Join(raiz, Carpeta, d.ID+".json")); string(despues) != string(antes) {
		t.Errorf("un rechazo no cambia el defecto:\n%s\n%s", antes, despues)
	}

	for _, c := range []struct{ caso, id string }{
		{"defecto con prueba pendiente", d.ID},
		{"defecto viejo sin prueba", "D000002"},
	} {
		g, err := Arreglar(raiz, c.id, "control implementado", "./p:TestP")
		if err != nil {
			t.Fatalf("%s: %v", c.caso, err)
		}
		if l, _ := leerUno(raiz, c.id); l.Arreglo != "control implementado" || l.Prueba != "./p:TestP" || g.Prueba != "./p:TestP" {
			t.Errorf("%s: guarda arreglo y prueba; quedó %+v", c.caso, l)
		}
	}
	// Choque con el gate de pruebas nombradas (E001/F017/S005): el defecto arreglado pasa y ya no cuenta como viejo.
	if ps, viejos := Revisar(raiz); len(ps) != 0 || viejos != 0 {
		t.Errorf("el gate acepta los defectos arreglados; dio %v, %d viejos", ps, viejos)
	}
}
