package backlog

import (
	"os"
	"strings"
	"testing"
)

// E001/F002/S011: la ficha guarda que_hace, una oración que dice qué hace.

func oracionDe(n int) string {
	p := make([]string, n)
	for i := range p {
		p[i] = "palabra"
	}
	return strings.Join(p, " ")
}

// campos de una ficha válida salvo que_hace.
func sinQueHace(id string) []string {
	return []string{"id=" + id, "titulo=T", "prioridad=P1", "motivo=m", "relacion=nueva", "prueba=./p:TestFalla=v|r|v"}
}

// Rojo cuando: nuevo acepta una ficha sin que_hace, o con solo espacios, 21 palabras o repetido; o escribe la ficha rechazada;
// o guarda 20 palabras sin que_hace, o una ficha vieja sin el campo deja de leerse.
// Verde cuando: iniciativa, feature y slice exigen que_hace, lo guardan, y la ficha sin el campo se lee bien.
func TestFichaGuardaQueHace(t *testing.T) {
	raiz := t.TempDir()
	ok := oracionDe(20)
	ids := []string{"E001", "E001/F001", "E001/F001/S001"}
	for _, id := range ids {
		archivo := activo(raiz, strings.ReplaceAll(id, "/", "-")+".json")
		malos := map[string][]string{
			"ausente":       sinQueHace(id),
			"solo espacios": append(sinQueHace(id), "que_hace=   "),
			"21 palabras":   append(sinQueHace(id), "que_hace="+oracionDe(21)),
			"repetido":      append(sinQueHace(id), "que_hace=uno", "que_hace=dos"),
		}
		for caso, campos := range malos {
			if _, err := Nuevo(raiz, campos); err == nil {
				t.Errorf("%s %s: nuevo debe rechazar", id, caso)
			}
			if _, err := os.Stat(archivo); err == nil {
				t.Errorf("%s %s: rechazada pero escribió la ficha", id, caso)
				_ = os.Remove(archivo)
			}
		}
		_, err := Nuevo(raiz, sinQueHace(id))
		if err == nil || !strings.Contains(err.Error(), "que_hace") {
			t.Errorf("%s ausente: el error debe nombrar que_hace; fue %v", id, err)
		}
		f, err := Nuevo(raiz, append(sinQueHace(id), "que_hace="+ok))
		if err != nil {
			t.Fatalf("%s con 20 palabras debe pasar: %v", id, err)
		}
		if g := leerFicha(t, archivo); g.QueHace != ok || f.QueHace != ok {
			t.Errorf("%s: que_hace guardado %q / devuelto %q; quería %q", id, g.QueHace, f.QueHace, ok)
		}
		b, _ := os.ReadFile(archivo)
		if !strings.Contains(string(b), `"que_hace": "`) {
			t.Errorf("%s: el JSON en disco no trae que_hace:\n%s", id, b)
		}
		if id == "E001/F001" { // el padre de un slice exige rojo (D000023)
			f.Rojo = &Registro{Fecha: "2026-10-21T00:00:00Z", Resultado: map[string]string{"./p:TestFalla": "falló"}}
			pisar(t, raiz, Activo, f)
		}
	}
	// Ficha vieja: sin el campo, sigue legible.
	vieja := Ficha{ID: "E002", Tipo: "iniciativa", Titulo: "Vieja", Estado: "BORRADOR", Prioridad: "P1", Motivo: "m", BloqueadoPor: []string{}, Secuencia: 9}
	pisar(t, raiz, Activo, vieja)
	g, _, err := Ver(raiz, "E002")
	if err != nil || g.QueHace != "" || g.Titulo != "Vieja" {
		t.Errorf("ficha sin que_hace debe leerse: %+v, %v", g, err)
	}
	if b, _ := os.ReadFile(activo(raiz, "E002.json")); strings.Contains(string(b), "que_hace") {
		t.Errorf("la ficha vieja no debe ganar que_hace:\n%s", b)
	}
	if MaxPalabrasQueHace != 20 {
		t.Errorf("tope de que_hace %d; quería 20", MaxPalabrasQueHace)
	}
}
