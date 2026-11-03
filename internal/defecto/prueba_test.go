package defecto

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// semilla deja en raiz una prueba real (./p:TestP) para que los defectos de las pruebas la nombren.
func semilla(t *testing.T, raiz string) {
	t.Helper()
	escribirArchivo(t, raiz, "p/p_test.go", "package p\n\nimport \"testing\"\n\nfunc TestP(t *testing.T) {}\n")
}

func escribirArchivo(t *testing.T, raiz, rel, texto string) {
	t.Helper()
	p := filepath.Join(raiz, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(texto), 0o644); err != nil {
		t.Fatal(err)
	}
}

func ficha(t *testing.T, raiz, carpeta, id, estado string) {
	t.Helper()
	b, _ := json.Marshal(map[string]any{"id": id, "tipo": "slice", "titulo": "ficha " + id, "estado": estado,
		"prioridad": "P1", "motivo": "m", "padre": id[:9], "bloqueado_por": []string{}, "secuencia": 1})
	escribirArchivo(t, raiz, "trabajo/"+carpeta+"/"+strings.ReplaceAll(id, "/", "-")+".json", string(b))
}

// TestDefectoNombraPrueba (E001/F017/S005): cada defecto nuevo nombra la prueba que lo vigila o la ficha que la hará.
// Rojo cuando: un defecto se registra sin prueba o con una prueba que no existe.
// Verde cuando: el defecto apunta a una prueba real o a una ficha viva, y el gate frena la que desaparece.
func TestDefectoNombraPrueba(t *testing.T) {
	raiz := t.TempDir()
	semilla(t, raiz)
	escribirArchivo(t, raiz, "internal/x/y_test.go", "package x\n\nimport \"testing\"\n\nfunc TestY_2(t *testing.T) {}\n")
	ficha(t, raiz, "activo", "E001/F002/S001", "BORRADOR")
	ficha(t, raiz, "archivo", "E001/F002/S002", "CERRADO")
	ficha(t, raiz, "archivo", "E001/F002/S003", "CANCELADO")

	con := func(firma, prueba string, hay bool) map[string]string {
		c := campos(firma)
		delete(c, "prueba")
		if hay {
			c["prueba"] = prueba
		}
		return c
	}
	buenos := []struct{ caso, prueba string }{
		{"prueba real", "./p:TestP"},
		{"prueba real repetida: una prueba vigila varios defectos", "./p:TestP"},
		{"paquete anidado, mismo formato que las fichas", "./internal/x:TestY_2"},
		{"pendiente a ficha en BORRADOR", "pendiente:E001/F002/S001"},
		{"pendiente a ficha CERRADO y archivada", "pendiente:E001/F002/S002"},
	}
	for i, c := range buenos {
		d, err := Nuevo(raiz, con("bueno-"+string(rune('a'+i)), c.prueba, true))
		if err != nil {
			t.Errorf("%s: debió registrar: %v", c.caso, err)
			continue
		}
		if g, _ := leerUno(raiz, d.ID); g.Prueba != c.prueba {
			t.Errorf("%s: el defecto guarda prueba=%q; dio %q", c.caso, c.prueba, g.Prueba)
		}
	}

	antes, _ := Leer(raiz)
	malos := []struct {
		caso, prueba string
		hay          bool
		dice         string
	}{
		{"campo ausente", "", false, "falta el campo prueba"},
		{"campo vacío", "   ", true, "falta el campo prueba"},
		{"sin paquete", "TestP", true, "./paquete:TestNombre"},
		{"nombre que no es prueba", "./p:Probar", true, "./paquete:TestNombre"},
		{"paquete fuera del repo", "./../fuera:TestP", true, "dentro del repo"},
		{"dos pruebas en un campo", "./p:TestP,./p:TestP", true, "./paquete:TestNombre"},
		{"prueba que no existe", "./p:TestNoHay", true, "no existe"},
		{"paquete que no existe", "./q:TestP", true, "no existe"},
		{"pendiente a ficha CANCELADO", "pendiente:E001/F002/S003", true, "CANCELADO"},
		{"pendiente a ficha que no existe", "pendiente:E001/F002/S099", true, "no existe"},
		{"pendiente sin ficha", "pendiente:", true, "pendiente:<id de ficha>"},
	}
	for i, c := range malos {
		if _, err := Nuevo(raiz, con("malo-"+string(rune('a'+i)), c.prueba, c.hay)); err == nil || !strings.Contains(err.Error(), c.dice) {
			t.Errorf("%s: debía fallar mencionando %q; dio %v", c.caso, c.dice, err)
		}
	}
	// Fallo a mitad de camino: un rechazo no deja defecto escrito.
	if despues, _ := Leer(raiz); len(despues) != len(antes) {
		t.Errorf("los rechazos no escriben defectos: había %d, hay %d", len(antes), len(despues))
	}

	// Gate: un defecto viejo sin campo se informa y no frena; una prueba nombrada que desaparece frena.
	escribirArchivo(t, raiz, Carpeta+"/D000099.json", `{"id":"D000099","firma":"viejo","que":"q","causa":"c","arreglo":"a","como_me_cercioro":"x","ocurrencias":[]}`)
	if errs, viejos := Revisar(raiz); len(errs) != 0 || viejos != 1 {
		t.Errorf("con todo en orden: sin errores y 1 defecto viejo sin prueba; dio %v, %d", errs, viejos)
	}
	if err := os.Remove(filepath.Join(raiz, "internal", "x", "y_test.go")); err != nil {
		t.Fatal(err)
	}
	if errs, _ := Revisar(raiz); len(errs) != 1 || !strings.Contains(errs[0], "D000003") || !strings.Contains(errs[0], "./internal/x:TestY_2") {
		t.Errorf("la prueba borrada frena el gate y nombra su defecto; dio %v", errs)
	}
}
