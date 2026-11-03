package defecto

import (
	"strings"
	"testing"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/backlog"
)

// conFamilia: campos válidos de un defecto con su familia.
func conFamilia(firma, familia string) map[string]string {
	c := campos(firma)
	c["familia"] = familia
	return c
}

// tramosFamilia: slices de E000/F007 con el título del tramo de esa familia, activos o archivados.
func tramosFamilia(t *testing.T, raiz, familia string) []backlog.Ficha {
	t.Helper()
	var out []backlog.Ficha
	for _, f := range hijosF007(t, raiz) {
		if f.Titulo == TituloFamilia(familia) {
			out = append(out, f)
		}
	}
	return out
}

// TestReincidenciaPorFamilia (E001/F018/S003): cada defecto nuevo lleva familia de la lista cerrada y tres apariciones
// de una familia sin tramo crean el tramo «Reincidencia familia <x>» en E000/F007, marcando esos defectos.
// Rojo cuando: familia fuera de lista se acepta o dos firmas de una familia no suman.
// Verde cuando: rechaza familia desconocida y suma por familia.
func TestReincidenciaPorFamilia(t *testing.T) {
	raiz := t.TempDir()
	semilla(t, raiz)
	nuevaFicha(t, raiz, "id=E000", "titulo=Mantenimiento")

	// Entrada ausente, vacía, mal escrita o desconocida: rechaza, nombra las válidas y no escribe nada.
	sin := campos("sin-familia")
	delete(sin, "familia")
	for nombre, c := range map[string]map[string]string{
		"ausente": sin, "vacía": conFamilia("x-vacia", ""), "mayúsculas": conFamilia("x-mayus", "Freno"), "desconocida": conFamilia("x-desc", "inventada"),
	} {
		if _, err := Nuevo(raiz, c); err == nil || !strings.Contains(err.Error(), "familia") || !strings.Contains(err.Error(), "freno") {
			t.Errorf("familia %s: debe rechazar nombrando las válidas; dio %v", nombre, err)
		}
	}
	if ds, _ := Leer(raiz); len(ds) != 0 {
		t.Fatalf("un rechazo no escribe defectos; hay %d", len(ds))
	}

	// Estado de origen: un defecto viejo sin familia, con tres apariciones, no cuenta para ninguna familia.
	escribirArchivo(t, raiz, Carpeta+"/D000001.json",
		`{"id":"D000001","firma":"viejo","que":"q","causa":"c","arreglo":"a","como_me_cercioro":"x","ocurrencias":["T000001","T000002"],"slice":"E000/F007/S900"}`)

	// Fallo a mitad: sin E000/F007, la tercera aparición de la familia falla, el defecto queda y nadie queda marcado.
	for _, f := range []string{"freno-a", "freno-b"} {
		if _, err := Nuevo(raiz, conFamilia(f, "freno")); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
	}
	if _, err := Nuevo(raiz, conFamilia("freno-c", "freno")); err == nil || !strings.Contains(err.Error(), Feature) {
		t.Fatalf("sin %s la tercera aparición de la familia debe fallar nombrándola; dio %v", Feature, err)
	}
	ds, _ := Leer(raiz)
	if len(ds) != 4 {
		t.Fatalf("el tercer defecto debe quedar guardado: hay %d", len(ds))
	}
	for _, d := range ds {
		if d.TramoFamilia != "" {
			t.Fatalf("un fallo a mitad no marca defectos: %+v", d)
		}
	}

	// Con la feature: el siguiente defecto de la familia crea el tramo y marca los cuatro de freno.
	nuevaFicha(t, raiz, "id=E000/F007", "titulo=Reincidencias de defectos")
	d, err := Nuevo(raiz, conFamilia("freno-d", "freno"))
	if err != nil {
		t.Fatal(err)
	}
	tramos := tramosFamilia(t, raiz, "freno")
	if len(tramos) != 1 || tramos[0].Prioridad != "P1" || d.TramoFamilia != tramos[0].ID {
		t.Fatalf("debe nacer un tramo P1 de la familia y marcar el defecto; tramos %+v, defecto %+v", tramos, d)
	}
	ds, _ = Leer(raiz)
	for _, x := range ds {
		if (x.Familia == "freno") != (x.TramoFamilia == tramos[0].ID) {
			t.Errorf("solo los defectos de freno quedan marcados: %+v", x)
		}
	}

	// Ya marcados no cuentan de nuevo; con el tramo abierto, tres apariciones nuevas lo reutilizan (hallazgo A1.1).
	for _, f := range []string{"freno-e", "freno-i"} {
		if d, err = Nuevo(raiz, conFamilia(f, "freno")); err != nil || d.TramoFamilia != "" {
			t.Fatalf("%s: un defecto nuevo tras el tramo no lo hereda bajo el umbral: %+v %v", f, d, err)
		}
	}
	if d, err = Nuevo(raiz, conFamilia("freno-j", "freno")); err != nil || d.TramoFamilia != tramos[0].ID {
		t.Fatalf("con el tramo abierto se reutiliza: %+v %v", d, err)
	}
	if n := len(tramosFamilia(t, raiz, "freno")); n != 1 {
		t.Fatalf("no se duplica el tramo de familia: hay %d", n)
	}

	// Choque con el conteo por defecto (hallazgo A1.3): freno-e, ya en el tramo de familia, llega a su umbral y no
	// recibe tramo propio.
	e, _ := leerUno(raiz, "D000006") // freno-e
	if _, err = Ocurre(raiz, e.ID, "T000011"); err != nil {
		t.Fatal(err)
	}
	if e, err = Ocurre(raiz, e.ID, "T000012"); err != nil || e.Slice != "" || e.TramoFamilia != tramos[0].ID {
		t.Fatalf("en el tramo de familia no nace tramo propio: %+v %v", e, err)
	}

	// Otra familia suma aparte (hallazgo A1.5); un defecto con tramo propio no cuenta para ella.
	a, err := Nuevo(raiz, conFamilia("hijo-a", "hijos"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Ocurre(raiz, a.ID, "T000013"); err != nil {
		t.Fatal(err)
	}
	if a, err = Ocurre(raiz, a.ID, "T000014"); err != nil || a.Slice == "" || a.TramoFamilia != "" {
		t.Fatalf("hijo-a llega a su umbral: tramo propio y sin familia: %+v %v", a, err)
	}
	// Un defecto de otra familia, sin marcar, presente mientras hijos suma (hallazgo A2.1): si las familias se
	// mezclaran, hijo-c ya llegaría al umbral.
	m, err := Nuevo(raiz, conFamilia("memo-a", "memoria"))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"hijo-b", "hijo-c"} {
		if d, err = Nuevo(raiz, conFamilia(f, "hijos")); err != nil || d.TramoFamilia != "" {
			t.Fatalf("%s: hijo-a con tramo propio no suma; bajo el umbral: %+v %v", f, d, err)
		}
	}
	if d, err = Nuevo(raiz, conFamilia("hijo-d", "hijos")); err != nil || d.TramoFamilia == "" || d.TramoFamilia == tramos[0].ID {
		t.Fatalf("hijos junta tres y nace su propio tramo: %+v %v", d, err)
	}
	if a, _ = leerUno(raiz, a.ID); a.TramoFamilia != "" {
		t.Fatalf("hijo-a, con tramo propio, no queda marcado: %+v", a)
	}
	if m, _ = leerUno(raiz, m.ID); m.TramoFamilia != "" {
		t.Fatalf("memo-a, de otra familia, no queda marcado: %+v", m)
	}
	// Turno repetido en un defecto ya en tramo de familia sigue fallando (hallazgo A2.2).
	if _, err = Ocurre(raiz, e.ID, "T000012"); err == nil || !strings.Contains(err.Error(), "ya tiene el turno") {
		t.Fatalf("turno repetido con tramo de familia debe fallar; dio %v", err)
	}

	// Padre e hijo: con el tramo de familia archivado, tres apariciones nuevas crean otro; abierto, se reutiliza.
	if _, err := backlog.Estado(raiz, tramos[0].ID, "CANCELADO"); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"freno-f", "freno-g"} {
		if _, err := Nuevo(raiz, conFamilia(f, "freno")); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
	}
	d, err = Nuevo(raiz, conFamilia("freno-h", "freno"))
	if err != nil {
		t.Fatal(err)
	}
	nuevos := tramosFamilia(t, raiz, "freno")
	if len(nuevos) != 2 || d.TramoFamilia == "" || d.TramoFamilia == tramos[0].ID {
		t.Fatalf("con el tramo archivado nace otro; tramos %+v, defecto %+v", nuevos, d)
	}
}
