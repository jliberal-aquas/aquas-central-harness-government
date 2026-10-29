package backlog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// E001/F002/S003: pruebas nombradas con su intención en la ficha; rojo prueba por prueba;
// verde después; cierre exige ambos.

var hora = time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC)

// Paquete ./p de un módulo de juguete: cada prueba ilustra un caso del rojo.
const pruebasP = `package p

import (
	"os"
	"testing"
)

var sucio bool

func TestFalla(t *testing.T)   { t.Log("ruido previo"); t.Fatal("motivo del contrato") }
func TestPasa(t *testing.T)    {}
func TestSalta(t *testing.T)   { t.Skip("saltada") }
func TestEnsucia(t *testing.T) { sucio = true }
func TestDepende(t *testing.T) {
	if sucio {
		t.Fatal("falla solo tras TestEnsucia")
	}
}
func TestBandera(t *testing.T) {
	if _, err := os.Stat("arreglado"); err != nil {
		t.Fatal("falta el arreglo")
	}
}
`

func modulo(t *testing.T) string {
	t.Helper()
	raiz := base(t)
	for rel, txt := range map[string]string{
		"go.mod":            "module juguete\n\ngo 1.26\n",
		"p/p_test.go":       pruebasP,
		"roto/roto_test.go": "package roto\n\nimport \"testing\"\n\nfunc TestX(t *testing.T) { no compila }\n",
	} {
		ruta := filepath.Join(raiz, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(ruta, []byte(txt), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return raiz
}

// enCurso: slice nuevo con esas pruebas ("./p:TestX=intención|r|v"), ya EN_CURSO.
func enCurso(t *testing.T, raiz, id string, pruebas ...string) {
	t.Helper()
	campos := []string{"id=" + id, "titulo=T", "prioridad=P1", "motivo=m"}
	for _, x := range pruebas {
		campos = append(campos, "prueba="+x)
	}
	nuevo(t, raiz, campos...)
	if _, err := arrancar(t, raiz, id); err != nil {
		t.Fatalf("EN_CURSO %s: %v", id, err)
	}
}

// pisar escribe la ficha a mano, como lo haría alguien saltándose aquas.
func pisar(t *testing.T, raiz, dir string, f Ficha) {
	t.Helper()
	b, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	d := filepath.Join(raiz, filepath.FromSlash(dir))
	if err := os.MkdirAll(d, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(d, nombre(f.ID)), b, 0o644); err != nil {
		t.Fatal(err)
	}
}

// conRojoYVerde pone a mano un rojo y un verde, para pruebas que solo necesitan cerrar.
func conRojoYVerde(t *testing.T, raiz, id string) {
	t.Helper()
	f := leerFicha(t, activo(raiz, nombre(id)))
	f.Rojo = &Registro{Fecha: "2026-10-21T12:00:00Z", Resultado: map[string]string{f.Pruebas[0].Nombre: "falló"}}
	f.Verde = &Registro{Fecha: "2026-10-21T13:00:00Z"}
	f.Auditoria = &Auditoria{Fecha: "2026-10-21T14:00:00Z", Veredicto: "APROBADO"} // E001/F014/S002: cerrar exige el APROBADO del auditor
	pisar(t, raiz, Activo, f)
}

func TestNuevoExigePruebas(t *testing.T) {
	raiz := base(t)
	ok := []string{"id=E001/F002/S001", "titulo=T", "prioridad=P1", "motivo=m", "que_hace=hace algo"}
	for _, c := range []struct {
		nombre string
		campos []string
		dice   string
	}{
		{"sin pruebas", nil, "prueba"},
		{"sin intención", []string{"prueba=./p:TestA"}, "intención"},
		{"intención vacía", []string{"prueba=./p:TestA= |r|v"}, "intención"},
		{"paquete sin ./", []string{"prueba=internal/p:TestA=x"}, "./paquete:TestNombre"},
		{"nombre sin Test", []string{"prueba=./p:testa=x"}, "./paquete:TestNombre"},
		{"sin paquete", []string{"prueba=TestA=x"}, "./paquete:TestNombre"},
		{"repetida", []string{"prueba=./p:TestA=x|r|v", "prueba=./p:TestA=y|r|v"}, "repetida"},
	} {
		if _, err := Nuevo(raiz, append(append([]string{}, ok...), c.campos...)); err == nil || !strings.Contains(err.Error(), c.dice) {
			t.Errorf("%s: se esperaba error con %q; dio %v", c.nombre, c.dice, err)
		}
	}
	f, err := Nuevo(raiz, append(ok, "relacion=nueva", "prueba=./p:TestFalla=falla por el contrato|r|v", "prueba=./cmd/x:TestB_2=otra = con igual|r|v"))
	if err != nil {
		t.Fatal(err)
	}
	g := leerFicha(t, activo(raiz, "E001-F002-S001.json"))
	quiere := []Prueba{{Nombre: "./p:TestFalla", Intencion: "falla por el contrato", RojoCuando: "r", VerdeCuando: "v"}, {Nombre: "./cmd/x:TestB_2", Intencion: "otra = con igual", RojoCuando: "r", VerdeCuando: "v"}}
	for _, x := range []Ficha{f, g} {
		if len(x.Pruebas) != 2 || x.Pruebas[0] != quiere[0] || x.Pruebas[1] != quiere[1] || x.Rojo != nil || x.Verde != nil {
			t.Errorf("pruebas con intención, en el orden dado, sin rojo ni verde: %+v", x)
		}
	}
}

func TestPruebasReiniciaRojoYVerde(t *testing.T) {
	raiz := base(t)
	f := nuevo(t, raiz, "id=E001/F002/S001", "titulo=T", "prioridad=P1", "motivo=m", "prueba=./p:TestFalla=falla|r|v")
	f.Rojo = &Registro{Fecha: "2026-10-20T10:00:00Z", Resultado: map[string]string{"./p:TestFalla": "x"}}
	f.Verde = &Registro{Fecha: "2026-10-20T11:00:00Z"}
	pisar(t, raiz, Activo, f)
	g, err := Pruebas(raiz, "E001/F002/S001", []string{"./p:TestPasa=pasa|r|v", "./p:TestBandera=pide el arreglo|r|v"})
	if err != nil {
		t.Fatal(err)
	}
	for _, x := range []Ficha{g, leerFicha(t, activo(raiz, "E001-F002-S001.json"))} {
		if len(x.Pruebas) != 2 || x.Pruebas[1] != (Prueba{Nombre: "./p:TestBandera", Intencion: "pide el arreglo", RojoCuando: "r", VerdeCuando: "v"}) || x.Rojo != nil || x.Verde != nil {
			t.Errorf("pruebas nuevas y rojo y verde borrados: %+v", x)
		}
	}
	for _, c := range []struct {
		nombre string
		lista  []string
		dice   string
	}{
		{"lista vacía", nil, "prueba"},
		{"formato malo", []string{"p:TestA=x"}, "./paquete:TestNombre"},
		{"sin intención", []string{"./p:TestA"}, "intención"},
	} {
		if _, err := Pruebas(raiz, "E001/F002/S001", c.lista); err == nil || !strings.Contains(err.Error(), c.dice) {
			t.Errorf("%s: se esperaba error con %q; dio %v", c.nombre, c.dice, err)
		}
	}
	if _, err := Estado(raiz, "E001/F002/S001", "CANCELADO"); err != nil {
		t.Fatal(err)
	}
	if _, err := Pruebas(raiz, "E001/F002/S001", []string{"./p:TestPasa=pasa|r|v"}); err == nil || !strings.Contains(err.Error(), "archivada") {
		t.Errorf("una archivada no cambia sus pruebas; dio %v", err)
	}
}

func TestRojoExigeQueCadaPruebaFalle(t *testing.T) {
	raiz := modulo(t)
	nuevo(t, raiz, "id=E001/F002/S001", "titulo=T", "prioridad=P1", "motivo=m", "prueba=./p:TestFalla=falla|r|v")
	sembrarAprobacion(t, raiz, "E001/F002/S001")
	if _, err := Rojo(raiz, "E001/F002/S001", hora); err == nil || !strings.Contains(err.Error(), "en curso") {
		t.Errorf("rojo de un slice que no está en curso; dio %v", err)
	}
	if _, err := arrancar(t, raiz, "E001/F002/S001"); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct{ prueba, dice string }{
		{"./p:TestPasa=pasa|r|v", "TestPasa"},
		{"./p:TestSalta=se salta|r|v", "TestSalta"},
	} {
		if _, err := Pruebas(raiz, "E001/F002/S001", []string{"./p:TestFalla=falla|r|v", c.prueba}); err != nil {
			t.Fatal(err)
		}
		sembrarAprobacion(t, raiz, "E001/F002/S001") // cambiar pruebas en curso pide aprobar de nuevo
		if _, err := Rojo(raiz, "E001/F002/S001", hora); err == nil || !strings.Contains(err.Error(), c.dice) || strings.Contains(err.Error(), "TestFalla") {
			t.Errorf("rojo rechaza solo %s; dio %v", c.dice, err)
		}
		if g := leerFicha(t, activo(raiz, "E001-F002-S001.json")); g.Rojo != nil {
			t.Errorf("rojo rechazado no se registra: %+v", g.Rojo)
		}
	}
	if _, err := Pruebas(raiz, "E001/F002/S001", []string{"./p:TestFalla=falla|r|v"}); err != nil {
		t.Fatal(err)
	}
	sembrarAprobacion(t, raiz, "E001/F002/S001") // cambiar pruebas en curso pide aprobar de nuevo
	f, err := Rojo(raiz, "E001/F002/S001", hora)
	if err != nil {
		t.Fatal(err)
	}
	for _, x := range []Ficha{f, leerFicha(t, activo(raiz, "E001-F002-S001.json"))} {
		if x.Rojo == nil || x.Rojo.Fecha != "2026-10-21T12:00:00Z" || !strings.Contains(x.Rojo.Resultado["./p:TestFalla"], "motivo del contrato") || strings.Contains(x.Rojo.Resultado["./p:TestFalla"], "ruido") {
			t.Errorf("rojo con fecha y el motivo del fallo de cada prueba: %+v", x.Rojo)
		}
	}
}

func TestRojoNoCuentaCompilacionNiAusencia(t *testing.T) {
	raiz := modulo(t)
	enCurso(t, raiz, "E001/F002/S001", "./roto:TestX=no compila|r|v")
	if _, err := Rojo(raiz, "E001/F002/S001", hora); err == nil || !strings.Contains(err.Error(), "no compila") {
		t.Errorf("un fallo de compilación no es rojo; dio %v", err)
	}
	if _, err := Pruebas(raiz, "E001/F002/S001", []string{"./p:TestNoExiste=no existe|r|v"}); err != nil {
		t.Fatal(err)
	}
	sembrarAprobacion(t, raiz, "E001/F002/S001") // cambiar pruebas en curso pide aprobar de nuevo
	if _, err := Rojo(raiz, "E001/F002/S001", hora); err == nil || !strings.Contains(err.Error(), "no existe") {
		t.Errorf("una prueba que no existe no es rojo; dio %v", err)
	}
	if _, err := Pruebas(raiz, "E001/F002/S001", []string{"./nada:TestX=paquete ausente|r|v"}); err != nil {
		t.Fatal(err)
	}
	sembrarAprobacion(t, raiz, "E001/F002/S001") // cambiar pruebas en curso pide aprobar de nuevo
	if _, err := Rojo(raiz, "E001/F002/S001", hora); err == nil || !strings.Contains(err.Error(), "paquete no existe") {
		t.Errorf("un paquete que no existe se nombra así, no como «no compila»; dio %v", err)
	}
}

func TestRojoPruebaAislada(t *testing.T) {
	raiz := modulo(t)
	// Juntas, TestDepende falla por culpa de TestEnsucia; sola, pasa: no es su rojo.
	enCurso(t, raiz, "E001/F002/S001", "./p:TestEnsucia=ensucia|r|v", "./p:TestDepende=depende|r|v")
	if _, err := Rojo(raiz, "E001/F002/S001", hora); err == nil || !strings.Contains(err.Error(), "TestDepende") {
		t.Errorf("rojo corre cada prueba sola; dio %v", err)
	}
}

func TestVerdeDespuesDelRojo(t *testing.T) {
	raiz := modulo(t)
	enCurso(t, raiz, "E001/F002/S001", "./p:TestBandera=pide el arreglo|r|v")
	if _, err := Verde(raiz, "E001/F002/S001", hora); err == nil || !strings.Contains(err.Error(), "rojo") {
		t.Errorf("verde sin rojo; dio %v", err)
	}
	if _, err := Rojo(raiz, "E001/F002/S001", hora); err != nil {
		t.Fatal(err)
	}
	if _, err := Verde(raiz, "E001/F002/S001", hora.Add(time.Hour)); err == nil || !strings.Contains(err.Error(), "TestBandera") {
		t.Errorf("verde con la prueba fallando; dio %v", err)
	}
	if err := os.WriteFile(filepath.Join(raiz, "p", "arreglado"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	f, err := Verde(raiz, "E001/F002/S001", hora.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	for _, x := range []Ficha{f, leerFicha(t, activo(raiz, "E001-F002-S001.json"))} {
		if x.Rojo == nil || x.Verde == nil || x.Verde.Fecha != "2026-10-21T13:00:00Z" {
			t.Errorf("verde registrado después del rojo: rojo %+v verde %+v", x.Rojo, x.Verde)
		}
	}
}

func TestCerrarExigeRojoYVerde(t *testing.T) {
	raiz := base(t)
	enCurso(t, raiz, "E001/F002/S001", "./p:TestFalla=falla|r|v")
	if _, err := Estado(raiz, "E001/F002/S001", "CERRADO"); err == nil || !strings.Contains(err.Error(), "rojo") {
		t.Errorf("cerrar sin rojo; dio %v", err)
	}
	f := leerFicha(t, activo(raiz, "E001-F002-S001.json"))
	f.Rojo = &Registro{Fecha: "2026-10-21T12:00:00Z", Resultado: map[string]string{"./p:TestFalla": "x"}}
	pisar(t, raiz, Activo, f)
	if _, err := Estado(raiz, "E001/F002/S001", "CERRADO"); err == nil || !strings.Contains(err.Error(), "verde") {
		t.Errorf("cerrar sin verde; dio %v", err)
	}
	f.Verde = &Registro{Fecha: "2026-10-21T13:00:00Z"}
	f.Auditoria = &Auditoria{Fecha: "2026-10-21T14:00:00Z", Veredicto: "APROBADO"} // E001/F014/S002
	pisar(t, raiz, Activo, f)
	marcarTodo(t, raiz, "E001/F002/S001") // desde S004 cerrar exige el sello marcado
	if g, err := Estado(raiz, "E001/F002/S001", "CERRADO"); err != nil || g.Verde == nil {
		t.Errorf("con rojo y verde cierra y los conserva: %+v %v", g, err)
	}
	// Ficha anterior a S003 (sin pruebas): para cerrarla primero se le ponen pruebas.
	pisar(t, raiz, Activo, Ficha{ID: "E001/F002/S002", Tipo: "slice", Titulo: "T", Estado: "EN_CURSO", Prioridad: "P1",
		Motivo: "m", Padre: "E001/F002", BloqueadoPor: []string{}, Secuencia: 9})
	sembrarAprobacion(t, raiz, "E001/F002/S002")
	if _, err := Estado(raiz, "E001/F002/S002", "CERRADO"); err == nil || !strings.Contains(err.Error(), "pruebas") {
		t.Errorf("cerrar sin pruebas; dio %v", err)
	}
	if _, err := Estado(raiz, "E001/F002/S002", "CANCELADO"); err != nil {
		t.Errorf("cancelar no exige pruebas: %v", err)
	}
}

func TestUnSoloSliceEnCurso(t *testing.T) {
	raiz := base(t)
	enCurso(t, raiz, "E001/F002/S001", "./p:TestFalla=falla|r|v")
	nuevo(t, raiz, "id=E001/F002/S002", "titulo=T", "prioridad=P2", "motivo=m", "prueba=./p:TestFalla=falla|r|v")
	for _, e := range []string{"EN_CURSO", "LISTO"} {
		if _, err := Estado(raiz, "E001/F002/S002", e); err == nil || !strings.Contains(err.Error(), "E001/F002/S001") {
			t.Errorf("con S001 en curso, S002 no pasa a %s; dio %v", e, err)
		}
	}
	if _, err := Prioridad(raiz, "E001/F002/S002", "P0", "urge"); err != nil {
		t.Errorf("la prioridad sí cambia en cualquier ficha: %v", err)
	}
	if _, err := arrancar(t, raiz, "E001/F002"); err != nil {
		t.Errorf("una feature sí cambia de estado: %v", err)
	}
	if _, err := Estado(raiz, "E001/F002/S001", "VERIFICANDO"); err != nil {
		t.Fatal(err)
	}
	if id, hay, err := SliceEnCurso(raiz); err != nil || !hay || id != "E001/F002/S001" {
		t.Errorf("VERIFICANDO sigue en curso: %q %v %v", id, hay, err)
	}
	if _, err := Estado(raiz, "E001/F002/S001", "CANCELADO"); err != nil {
		t.Fatal(err)
	}
	if _, hay, err := SliceEnCurso(raiz); err != nil || hay {
		t.Errorf("sin slice en curso: %v %v", hay, err)
	}
	if _, err := arrancar(t, raiz, "E001/F002/S002"); err != nil {
		t.Errorf("libre el turno, S002 entra: %v", err)
	}
}

func TestValidarRojoYVerde(t *testing.T) {
	raiz := base(t)
	s := func(n, estado string, pruebas []Prueba, rojo, verde *Registro) Ficha {
		return Ficha{ID: "E001/F002/S00" + n, Tipo: "slice", Titulo: "T", Estado: estado, Prioridad: "P1", Motivo: "m",
			Padre: "E001/F002", BloqueadoPor: []string{}, Secuencia: 10 + int(n[0]-'0'), Pruebas: pruebas, Rojo: rojo, Verde: verde}
	}
	r := func(fecha string, res ...string) *Registro {
		m := map[string]string{}
		for _, x := range res {
			m[x] = "falló"
		}
		return &Registro{Fecha: fecha, Resultado: m}
	}
	ok := []Prueba{{Nombre: "./p:TestA", Intencion: "a", RojoCuando: "r", VerdeCuando: "v"}}
	pisar(t, raiz, Activo, s("1", "BORRADOR", ok, nil, r("2026-10-21T13:00:00Z")))                                                          // verde sin rojo
	pisar(t, raiz, Activo, s("2", "BORRADOR", ok, r("2026-10-21T12:00:00Z", "./p:TestOtra"), nil))                                          // rojo de otra prueba
	pisar(t, raiz, Activo, s("3", "EN_CURSO", ok, nil, nil))                                                                                // dos en curso
	pisar(t, raiz, Activo, s("4", "VERIFICANDO", ok, nil, nil))                                                                             // dos en curso
	pisar(t, raiz, Archivo, s("5", "CERRADO", ok, r("2026-10-21T12:00:00Z", "./p:TestA"), nil))                                             // cerrado sin verde
	pisar(t, raiz, Activo, s("6", "BORRADOR", ok, r("2026-10-21T12:00:00Z", "./p:TestA"), r("2026-10-21T11:00:00Z")))                       // verde antes del rojo
	pisar(t, raiz, Activo, s("7", "BORRADOR", []Prueba{{Nombre: "p:TestA", Intencion: "a"}}, nil, nil))                                     // formato malo
	pisar(t, raiz, Archivo, s("9", "CERRADO", ok, nil, r("2026-10-21T13:00:00Z")))                                                          // cerrado sin rojo
	pisar(t, raiz, Activo, s("8", "BORRADOR", []Prueba{{Nombre: "./p:TestA", Intencion: "", RojoCuando: "r", VerdeCuando: "v"}}, nil, nil)) // sin intención
	p := strings.Join(Validar(raiz), "\n")
	for _, c := range []struct{ archivo, dice string }{
		{"E001-F002-S001.json", "verde sin rojo"},
		{"E001-F002-S002.json", "./p:TestOtra"},
		{"E001-F002-S004.json", "en curso"},
		{"E001-F002-S005.json", "sin verde"},
		{"E001-F002-S006.json", "antes del rojo"},
		{"E001-F002-S007.json", "./paquete:TestNombre"},
		{"E001-F002-S008.json", "intención"},
		{"E001-F002-S009.json", "cerrada sin rojo"},
	} {
		hallado := false
		for _, linea := range strings.Split(p, "\n") {
			if strings.Contains(linea, c.archivo) && strings.Contains(linea, c.dice) {
				hallado = true
			}
		}
		if !hallado {
			t.Errorf("Validar: falta %q en la línea de %s; dio:\n%s", c.dice, c.archivo, p)
		}
	}
}
