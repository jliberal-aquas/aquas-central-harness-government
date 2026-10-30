package defecto

import (
	"strings"
	"testing"
	"time"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/backlog"
)

// hijosF007: los slices de E000/F007, activos o archivados.
func hijosF007(t *testing.T, raiz string) []backlog.Ficha {
	t.Helper()
	fs, err := backlog.Todas(raiz)
	if err != nil {
		t.Fatal(err)
	}
	var out []backlog.Ficha
	for _, f := range fs {
		if f.Padre == Feature {
			out = append(out, f)
		}
	}
	return out
}

func nuevaFicha(t *testing.T, raiz string, campos ...string) {
	t.Helper()
	base := []string{"prioridad=P1", "motivo=prueba", "que_hace=hace algo", "relacion=nueva",
		"prueba=./internal/x:TestX=valida algo|rojo cuando falla|verde cuando pasa"}
	if _, err := backlog.Nuevo(raiz, append(campos, base...)); err != nil {
		t.Fatal(err)
	}
}

// TestReincidenciaCreaSlice (E001/F004/S003): la tercera aparición de un defecto crea un slice P1 en E000/F007, una sola vez.
// Rojo cuando: al llegar al umbral no crea el slice, o lo crea repetido.
// Verde cuando: crea un slice P1 en el umbral; debajo no crea nada.
func TestReincidenciaCreaSlice(t *testing.T) {
	raiz := t.TempDir()
	nuevaFicha(t, raiz, "id=E000", "titulo=Mantenimiento")

	d, err := Nuevo(raiz, campos("json-salto-crudo"))
	if err != nil {
		t.Fatal(err)
	}
	// Segunda aparición: debajo del umbral, no crea nada (y aún no exige E000/F007).
	if d, err = Ocurre(raiz, d.ID, "T000001"); err != nil || d.Slice != "" {
		t.Fatalf("segunda aparición: %+v %v", d, err)
	}
	// Tercera aparición sin E000/F007: falla por su ausencia (hallazgo A1.5), pero la ocurrencia queda guardada.
	if _, err = Ocurre(raiz, d.ID, "T000002"); err == nil || !strings.Contains(err.Error(), "E000/F007 no existe") {
		t.Fatalf("sin E000/F007 debe fallar por su ausencia; dio %v", err)
	}
	if d, _ = leerUno(raiz, d.ID); len(d.Ocurrencias) != 2 || d.Slice != "" {
		t.Fatalf("la ocurrencia debe quedar guardada sin slice: %+v", d)
	}
	// Con la feature y un hueco (S001 y S003, sin S002), repetir el mismo turno reintenta (hallazgo A1.6)
	// y toma el siguiente al mayor, no el primer hueco (hallazgo A1.4).
	nuevaFicha(t, raiz, "id=E000/F007", "titulo=Reincidencias de defectos")
	nuevaFicha(t, raiz, "id=E000/F007/S001", "titulo=Trabajo previo uno")
	nuevaFicha(t, raiz, "id=E000/F007/S003", "titulo=Trabajo previo tres")
	if d, err = Ocurre(raiz, d.ID, "T000002"); err != nil {
		t.Fatalf("reintento con el mismo turno: %v", err)
	}
	if d.Slice != "E000/F007/S004" || len(d.Ocurrencias) != 2 {
		t.Fatalf("debe crear E000/F007/S004 sin sumar ocurrencia; dio %q, %v", d.Slice, d.Ocurrencias)
	}
	f, _, err := backlog.Ver(raiz, d.Slice)
	if err != nil {
		t.Fatal(err)
	}
	if f.Prioridad != "P1" || f.Estado != "BORRADOR" || f.Titulo != "Reincidencia D000001: json-salto-crudo" ||
		len(f.Pruebas) != 1 || f.Pruebas[0].Nombre != "./internal/defecto:TestReincidenciaD000001" {
		t.Fatalf("slice creado: %+v", f)
	}
	if g, _ := leerUno(raiz, d.ID); g.Slice != d.Slice {
		t.Fatalf("el defecto no guardó su slice: %+v", g)
	}
	// Con slice guardado, el mismo turno vuelve a ser repetido.
	if _, err = Ocurre(raiz, d.ID, "T000002"); err == nil || !strings.Contains(err.Error(), "ya tiene") {
		t.Fatalf("turno repetido con slice: debe fallar; dio %v", err)
	}
	// Encima del umbral, aunque el slice cambie de título: no crea otro (guarda del campo slice, hallazgo A1.3).
	if _, err = backlog.Editar(raiz, d.Slice, []string{"titulo=Causa de los saltos crudos", "motivo=prueba"}, time.Now()); err != nil {
		t.Fatal(err)
	}
	if d, err = Ocurre(raiz, d.ID, "T000004"); err != nil || len(hijosF007(t, raiz)) != 3 {
		t.Fatalf("cuarta aparición creó otro slice: %d %v", len(hijosF007(t, raiz)), err)
	}
	// Turno repetido o mal escrito: falla sin crear nada.
	for _, turno := range []string{"T000004", "4"} {
		if _, err = Ocurre(raiz, d.ID, turno); err == nil || len(hijosF007(t, raiz)) != 3 {
			t.Fatalf("turno %q: debe fallar sin crear slice", turno)
		}
	}

	// Fallo a mitad: el slice nació pero el defecto no lo guardó; el reintento reutiliza el existente, aun archivado.
	d2, err := Nuevo(raiz, campos("otro-fallo"))
	if err != nil {
		t.Fatal(err)
	}
	nuevaFicha(t, raiz, "id=E000/F007/S005", "titulo=Reincidencia D000002: otro-fallo")
	if _, err = backlog.Estado(raiz, "E000/F007/S005", "CANCELADO"); err != nil {
		t.Fatal(err)
	}
	for _, turno := range []string{"T000010", "T000011"} {
		if d2, err = Ocurre(raiz, d2.ID, turno); err != nil {
			t.Fatal(err)
		}
	}
	if d2.Slice != "E000/F007/S005" || len(hijosF007(t, raiz)) != 4 {
		t.Fatalf("debe reutilizar E000/F007/S005 archivado sin crear otro: %+v %d", d2, len(hijosF007(t, raiz)))
	}

	// Feature archivada: se niega sin crear el hijo activo que rompería el gate (hallazgo A1.1).
	otra := t.TempDir()
	nuevaFicha(t, otra, "id=E000", "titulo=Mantenimiento")
	nuevaFicha(t, otra, "id=E000/F007", "titulo=Reincidencias de defectos")
	if _, err = backlog.Estado(otra, "E000/F007", "CANCELADO"); err != nil {
		t.Fatal(err)
	}
	d3, err := Nuevo(otra, campos("tercer-fallo"))
	if err != nil {
		t.Fatal(err)
	}
	_, _ = Ocurre(otra, d3.ID, "T000020")
	if _, err = Ocurre(otra, d3.ID, "T000021"); err == nil || !strings.Contains(err.Error(), "archivada") || len(hijosF007(t, otra)) != 0 {
		t.Fatalf("feature archivada: debe negarse sin crear slice; dio %v, %d hijos", err, len(hijosF007(t, otra)))
	}
	if errs := backlog.Validar(otra); len(errs) != 0 {
		t.Fatalf("el backlog quedó inválido: %v", errs)
	}
}
