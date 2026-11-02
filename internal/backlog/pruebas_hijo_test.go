package backlog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// conPlantilla escribe gobierno/config/pruebas.json del repo.
func conPlantilla(t *testing.T, raiz, txt string) {
	t.Helper()
	p := filepath.Join(raiz, filepath.FromSlash(RutaPruebas))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(txt), 0o644); err != nil {
		t.Fatal(err)
	}
}

func archivoEn(t *testing.T, raiz, rel, txt string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(raiz, rel), []byte(txt), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestPruebasDelHijoSinGo (E003/F001/S004): con gobierno/config/pruebas.json, rojo y verde corren la
// plantilla, una prueba a la vez, y su salida decide; sin archivo sigue go test.
func TestPruebasDelHijoSinGo(t *testing.T) {
	const prueba = "prueba=p.sh:arreglado=pide el arreglo|r|v"
	const script = "test -f \"$1\" || { echo \"falta $1\"; exit 1; }\n"
	ahora := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)

	// Sin archivo: la prueba sigue siendo de Go; un nombre libre se rechaza como hoy.
	raiz := base(t)
	if _, err := Nuevo(raiz, []string{"id=E001/F002/S009", "titulo=T", "prioridad=P1", "motivo=m", "relacion=nueva", "que_hace=hace algo", prueba}); err == nil {
		t.Error("sin pruebas.json, una prueba sin forma ./paquete:TestNombre debió rechazarse")
	}

	// Entrada mal escrita: cada caso da un error claro que nombra pruebas.json, y no crea la ficha.
	for _, c := range []struct{ txt, dice string }{
		{`{"version": 1, "comando": `, "no es JSON válido"},
		{`{"version": 1, "comando": "  "}`, "comando vacío"},
		{`{"version": 1, "comando": "sh {ruta}"}`, "{nombre}"},
		{`{"version": 2, "comando": "sh {ruta} {nombre}"}`, "version"},
	} {
		conPlantilla(t, raiz, c.txt)
		_, err := Nuevo(raiz, []string{"id=E001/F002/S009", "titulo=T", "prioridad=P1", "motivo=m", "relacion=nueva", "que_hace=hace algo", prueba})
		if err == nil || !strings.Contains(err.Error(), "pruebas.json") || !strings.Contains(err.Error(), c.dice) {
			t.Errorf("pruebas.json %q: debió fallar nombrando pruebas.json y %q; dio %v", c.txt, c.dice, err)
		}
		if p := Validar(raiz); !strings.Contains(strings.Join(p, "\n"), "pruebas.json") {
			t.Errorf("pruebas.json %q: el gate debió reportarlo; dio %v", c.txt, p)
		}
	}

	// Slice: rojo corre la plantilla y guarda su salida; verde pasa tras el arreglo.
	conPlantilla(t, raiz, `{"version": 1, "comando": "sh {ruta} {nombre}"}`)
	archivoEn(t, raiz, "p.sh", script)
	enCurso(t, raiz, "E001/F002/S009", "p.sh:arreglado=pide el arreglo|r|v")
	f, err := Rojo(raiz, "E001/F002/S009", ahora)
	if err != nil {
		t.Fatalf("rojo con plantilla: %v", err)
	}
	if m := f.Rojo.Resultado["p.sh:arreglado"]; !strings.Contains(m, "falta arreglado") {
		t.Errorf("el motivo del rojo debió ser la salida del comando; dio %q", m)
	}
	if _, err := Verde(raiz, "E001/F002/S009", ahora); err == nil {
		t.Error("verde sin el arreglo debió rechazarse")
	}
	archivoEn(t, raiz, "arreglado", "")
	if _, err := Verde(raiz, "E001/F002/S009", ahora); err != nil {
		t.Errorf("verde con el arreglo: %v", err)
	}
	if p := Validar(raiz); len(p) != 0 {
		t.Errorf("el gate no debió quejarse de pruebas con plantilla: %v", p)
	}

	// Nombre mal escrito (hallazgo A1.6): la ruta no sale del repo ni se vuelve opción; el error muestra el formato libre.
	for _, mala := range []string{"../../x.sh:y", "-rf:y", "p.sh:-x", "a b.sh:y", "p.sh:y;rm"} {
		_, err := listaPruebasDe(raiz, []string{mala + "=v|r|v"})
		if err == nil || !strings.Contains(err.Error(), formatoLibre) {
			t.Errorf("prueba %q: debió rechazarse mostrando %s; dio %v", mala, formatoLibre, err)
		}
	}
	if _, err := listaPruebasDe(raiz, []string{"p.sh:y"}); err == nil || !strings.Contains(err.Error(), formatoLibre) {
		t.Errorf("sin intención, el error debió mostrar el formato libre; dio %v", err)
	}

	// Feature: también corre con la plantilla.
	nuevo(t, raiz, "id=E001/F003", "titulo=F", "prioridad=P1", "motivo=m", "prueba=p.sh:otro=pide otro|r|v")
	if _, err := Rojo(raiz, "E001/F003", ahora); err != nil {
		t.Errorf("rojo de feature con plantilla: %v", err)
	}

	// Una a una: cada prueba corre sola, así una no tapa el resultado de otra.
	r, err := correr(raiz, []Prueba{{Nombre: "p.sh:arreglado"}, {Nombre: "p.sh:otro"}})
	if err != nil || r["p.sh:arreglado"].accion != "pass" || r["p.sh:otro"].accion != "fail" {
		t.Errorf("cada prueba debió tener su propio resultado; dio %+v %v", r, err)
	}

	// Choque con otro control: la plantilla no salta la aprobación del contrato antes del rojo.
	g := nuevo(t, raiz, "id=E001/F002/S010", "titulo=T", "prioridad=P1", "motivo=m", "prueba=p.sh:otro=pide otro|r|v")
	g.Estado = "EN_CURSO"
	pisar(t, raiz, Activo, g)
	if _, err := Rojo(raiz, "E001/F002/S010", ahora); err == nil {
		t.Error("rojo sin aprobación debió rechazarse también con plantilla")
	}

	// Fallo a mitad de camino: un comando que no existe no cuenta como rojo ni como verde.
	conPlantilla(t, raiz, `{"version": 1, "comando": "no-existe-aquas-xyz {ruta} {nombre}"}`)
	r, err = correr(raiz, []Prueba{{Nombre: "p.sh:arreglado"}})
	if err != nil {
		t.Fatal(err)
	}
	if x := r["p.sh:arreglado"]; x.accion == "fail" || x.accion == "pass" || !strings.Contains(x.motivo(""), "no cuenta") {
		t.Errorf("un comando que no existe debió quedar «no cuenta»; dio %+v (%s)", x, x.motivo(""))
	}
}
