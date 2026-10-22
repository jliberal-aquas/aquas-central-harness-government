package backlog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// E001/F002/S002: backlog en fichas JSON; solo aquas escribe; lo terminal se archiva.

func nuevo(t *testing.T, raiz string, campos ...string) Ficha {
	t.Helper()
	if !strings.Contains(strings.Join(campos, " "), "prueba=") {
		campos = append(campos, "prueba=./p:TestFalla=falla a propósito|r|v") // desde S003 toda ficha nace con pruebas
	}
	f, err := Nuevo(raiz, campos)
	if err != nil {
		t.Fatalf("Nuevo(%v): %v", campos, err)
	}
	return f
}

func leerFicha(t *testing.T, ruta string) Ficha {
	t.Helper()
	b, err := os.ReadFile(ruta)
	if err != nil {
		t.Fatalf("leer %s: %v", ruta, err)
	}
	var f Ficha
	if err := json.Unmarshal(b, &f); err != nil {
		t.Fatalf("JSON de %s: %v", ruta, err)
	}
	return f
}

func activo(raiz, nombre string) string  { return filepath.Join(raiz, Activo, nombre) }
func archivo(raiz, nombre string) string { return filepath.Join(raiz, Archivo, nombre) }

func base(t *testing.T) string {
	t.Helper()
	raiz := t.TempDir()
	nuevo(t, raiz, "id=E001", "titulo=Gobierno", "prioridad=P1", "motivo=raíz")
	nuevo(t, raiz, "id=E001/F002", "titulo=Backlog", "prioridad=P1", "motivo=orden")
	return raiz
}

func TestNuevoCreaFicha(t *testing.T) {
	raiz := base(t)
	f := nuevo(t, raiz, "id=E001/F002/S002", "titulo=Backlog en JSON", "prioridad=P1", "motivo=lo pidió Jesús", "bloqueado_por=E001/F002")
	g := leerFicha(t, activo(raiz, "E001-F002-S002.json"))
	want := Ficha{ID: "E001/F002/S002", Tipo: "slice", Titulo: "Backlog en JSON", Estado: "BORRADOR", Prioridad: "P1",
		Motivo: "lo pidió Jesús", Padre: "E001/F002", BloqueadoPor: []string{"E001/F002"}, Secuencia: 3}
	for _, x := range []Ficha{f, g} {
		if x.ID != want.ID || x.Tipo != want.Tipo || x.Titulo != want.Titulo || x.Estado != want.Estado ||
			x.Prioridad != want.Prioridad || x.Motivo != want.Motivo || x.Padre != want.Padre ||
			strings.Join(x.BloqueadoPor, ",") != "E001/F002" || x.Secuencia != want.Secuencia {
			t.Errorf("ficha = %+v; se esperaba %+v", x, want)
		}
	}
	if e := leerFicha(t, activo(raiz, "E001.json")); e.Tipo != "iniciativa" || e.Padre != "" || e.Secuencia != 1 {
		t.Errorf("iniciativa sin padre y secuencia 1: %+v", e)
	}
	if e := leerFicha(t, activo(raiz, "E001-F002.json")); e.Tipo != "feature" || e.Padre != "E001" {
		t.Errorf("feature con padre E001: %+v", e)
	}
	if g.BloqueadoPor == nil {
		t.Error("bloqueado_por nunca es null")
	}
	if s := nuevo(t, raiz, "id=E001/F002/S003", "titulo=Rojo", "prioridad=P2", "motivo=sigue"); s.BloqueadoPor == nil || len(s.BloqueadoPor) != 0 {
		t.Errorf("sin bloqueos = lista vacía, no null: %#v", s.BloqueadoPor)
	}
}

func TestNuevoRechaza(t *testing.T) {
	raiz := base(t)
	ok := []string{"titulo=T", "prioridad=P1", "motivo=m", "prueba=./p:TestA=a|r|v"}
	for _, c := range []struct {
		nombre string
		campos []string
		dice   string
	}{
		{"id mal formado", append([]string{"id=E1/F2"}, ok...), "id"},
		{"padre inexistente", append([]string{"id=E001/F009/S001"}, ok...), "padre"},
		{"id repetido", append([]string{"id=E001/F002"}, ok...), "ya existe"},
		{"prioridad inválida", []string{"id=E001/F002/S001", "titulo=T", "prioridad=P9", "motivo=m", "prueba=./p:TestA=a|r|v"}, "prioridad"},
		{"sin motivo", []string{"id=E001/F002/S001", "titulo=T", "prioridad=P1", "prueba=./p:TestA=a|r|v"}, "motivo"},
		{"sin título", []string{"id=E001/F002/S001", "prioridad=P1", "motivo=m", "prueba=./p:TestA=a|r|v"}, "titulo"},
		{"campo desconocido", append([]string{"id=E001/F002/S001", "color=rojo"}, ok...), "color"},
		{"bloqueo inexistente", append([]string{"id=E001/F002/S001", "bloqueado_por=E009"}, ok...), "E009"},
		{"campo repetido", append([]string{"id=E001/F002/S001", "id=E001/F002/S009"}, ok...), "repetido"},
		{"bloqueo repetido", append([]string{"id=E001/F002/S001", "bloqueado_por=E001,E001"}, ok...), "repetido"},
	} {
		if _, err := Nuevo(raiz, c.campos); err == nil || !strings.Contains(err.Error(), c.dice) {
			t.Errorf("%s: se esperaba error con %q; dio %v", c.nombre, c.dice, err)
		}
	}
	// Un id ya archivado tampoco se reutiliza.
	nuevo(t, raiz, "id=E001/F002/S001", "titulo=T", "prioridad=P1", "motivo=m")
	if _, err := Estado(raiz, "E001/F002/S001", "CANCELADO"); err != nil {
		t.Fatal(err)
	}
	if _, err := Nuevo(raiz, append([]string{"id=E001/F002/S001"}, ok...)); err == nil || !strings.Contains(err.Error(), "ya existe") {
		t.Errorf("id archivado no se reutiliza; dio %v", err)
	}
}

func TestEstadoTerminalArchiva(t *testing.T) {
	raiz := base(t)
	nuevo(t, raiz, "id=E001/F002/S002", "titulo=T", "prioridad=P1", "motivo=m")
	if f, err := Estado(raiz, "E001/F002/S002", "EN_CURSO"); err != nil || f.Estado != "EN_CURSO" {
		t.Fatalf("EN_CURSO: %+v %v", f, err)
	}
	if _, err := Estado(raiz, "E001/F002/S002", "LISTISIMO"); err == nil || !strings.Contains(err.Error(), "estado") {
		t.Errorf("estado inválido debe fallar; dio %v", err)
	}
	conRojoYVerde(t, raiz, "E001/F002/S002")
	marcarTodo(t, raiz, "E001/F002/S002") // desde S004 cerrar exige el sello marcado
	if _, err := Estado(raiz, "E001/F002/S002", "CERRADO"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(activo(raiz, "E001-F002-S002.json")); !os.IsNotExist(err) {
		t.Error("la ficha cerrada no queda en activo")
	}
	if g := leerFicha(t, archivo(raiz, "E001-F002-S002.json")); g.Estado != "CERRADO" {
		t.Errorf("en archivo con estado CERRADO: %+v", g)
	}
	if f, archivada, err := Ver(raiz, "E001/F002/S002"); err != nil || !archivada || f.Estado != "CERRADO" {
		t.Errorf("Ver encuentra la archivada: %+v %v %v", f, archivada, err)
	}
	if _, err := Estado(raiz, "E001/F002/S002", "EN_CURSO"); err == nil || !strings.Contains(err.Error(), "archivada") {
		t.Errorf("una archivada no cambia; dio %v", err)
	}
	if _, err := Prioridad(raiz, "E001/F002/S002", "P0", "m"); err == nil || !strings.Contains(err.Error(), "archivada") {
		t.Errorf("una archivada no cambia de prioridad; dio %v", err)
	}
	if _, err := Estado(raiz, "E001/F009", "LISTO"); err == nil || !strings.Contains(err.Error(), "no existe") {
		t.Errorf("id inexistente; dio %v", err)
	}
}

func TestPrioridadConMotivo(t *testing.T) {
	raiz := base(t)
	nuevo(t, raiz, "id=E001/F002/S002", "titulo=T", "prioridad=P2", "motivo=viejo")
	if _, err := Prioridad(raiz, "E001/F002/S002", "P0", ""); err == nil || !strings.Contains(err.Error(), "motivo") {
		t.Errorf("sin motivo debe fallar; dio %v", err)
	}
	if _, err := Prioridad(raiz, "E001/F002/S002", "PX", "m"); err == nil || !strings.Contains(err.Error(), "prioridad") {
		t.Errorf("prioridad inválida; dio %v", err)
	}
	if _, err := Prioridad(raiz, "E001/F002/S002", "P0", "rompe el repo"); err != nil {
		t.Fatal(err)
	}
	if g := leerFicha(t, activo(raiz, "E001-F002-S002.json")); g.Prioridad != "P0" || g.Motivo != "rompe el repo" {
		t.Errorf("prioridad y motivo guardados: %+v", g)
	}
}

func TestSiguienteOrdenFijo(t *testing.T) {
	raiz := base(t)
	s := func(id, p string, extra ...string) {
		nuevo(t, raiz, append([]string{"id=" + id, "titulo=T", "prioridad=" + p, "motivo=m"}, extra...)...)
	}
	// Desde E001/F002/S008: feature sin slices exige definirse antes que cualquier slice.
	if f, hay, err := Siguiente(raiz); err != nil || !hay || f.ID != "E001/F002" {
		t.Fatalf("sin slices, siguiente pide definir la feature: %q %v %v", f.ID, hay, err)
	}
	s("E001/F002/S001", "P2")
	s("E001/F002/S002", "P1", "bloqueado_por=E001/F002/S001") // bloqueado: S001 no está cerrado
	s("E001/F002/S003", "P1")
	s("E001/F002/S004", "P1")
	s("E001/F002/S005", "SIN_TRIAGE")
	pide := func(want string) {
		t.Helper()
		f, hay, err := Siguiente(raiz)
		if err != nil || !hay || f.ID != want {
			t.Fatalf("siguiente = %q (%v, %v); se esperaba %s", f.ID, hay, err, want)
		}
	}
	cambia := func(id, e string) {
		t.Helper()
		if e == "CERRADO" {
			conRojoYVerde(t, raiz, id) // desde S003 cerrar exige rojo y verde
			marcarTodo(t, raiz, id)    // desde S004 cerrar exige el sello marcado
		}
		if _, err := Estado(raiz, id, e); err != nil {
			t.Fatalf("Estado(%s, %s): %v", id, e, err)
		}
	}
	pide("E001/F002/S003") // P1 desbloqueado, menor secuencia
	cambia("E001/F002/S004", "EN_CURSO")
	pide("E001/F002/S004") // en curso primero: un solo slice activo
	cambia("E001/F002/S004", "CERRADO")
	cambia("E001/F002/S003", "BLOQUEADO")
	pide("E001/F002/S001") // S003 bloqueado por estado; S002 por S001 → queda P2
	cambia("E001/F002/S001", "CERRADO")
	pide("E001/F002/S002") // S001 cerrado desbloquea S002
	cambia("E001/F002/S002", "CANCELADO")
	cambia("E001/F002/S003", "LISTO")
	pide("E001/F002/S003")
	act, err := Activas(raiz)
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, f := range act {
		ids = append(ids, f.ID)
	}
	if got := strings.Join(ids, " "); got != "E001/F002/S003 E001/F002/S005 E001 E001/F002" {
		t.Errorf("Activas en orden de siguiente (slices primero, luego el resto por secuencia): %s", got)
	}
	// Hallazgos del auditor (2026-10-21): VERIFICANDO cuenta como en curso; Activas pone lo
	// bloqueado después de lo libre; un bloqueador CANCELADO libera.
	s("E001/F002/S006", "P0")
	s("E001/F002/S007", "P0", "bloqueado_por=E001/F002/S006")
	s("E001/F002/S008", "P3")
	cambia("E001/F002/S008", "VERIFICANDO")
	pide("E001/F002/S008")
	act, _ = Activas(raiz)
	ids = nil
	for _, f := range act {
		ids = append(ids, f.ID)
	}
	if got := strings.Join(ids, " "); got != "E001/F002/S008 E001/F002/S006 E001/F002/S003 E001/F002/S005 E001/F002/S007 E001 E001/F002" {
		t.Errorf("en curso, luego libres por prioridad, luego bloqueadas: %s", got)
	}
	cambia("E001/F002/S008", "CERRADO")
	cambia("E001/F002/S006", "CANCELADO")
	pide("E001/F002/S007")
}

func TestValidarDetectaFichasMalas(t *testing.T) {
	raiz := base(t)
	nuevo(t, raiz, "id=E001/F002/S002", "titulo=T", "prioridad=P1", "motivo=m")
	if p := Validar(raiz); len(p) != 0 {
		t.Fatalf("backlog sano no da problemas: %v", p)
	}
	if p := Validar(t.TempDir()); len(p) != 0 {
		t.Fatalf("sin carpeta trabajo/ no hay problemas: %v", p)
	}
	escribir := func(rel, c string) {
		p := filepath.Join(raiz, filepath.FromSlash(rel))
		os.MkdirAll(filepath.Dir(p), 0o755)
		os.WriteFile(p, []byte(c), 0o644)
	}
	escribir(Activo+"/roto.json", "{no es json")
	escribir(Activo+"/E001-F003.json", `{"id":"E001/F003","tipo":"feature","titulo":"T","estado":"LISTO","prioridad":"P1","motivo":"m","padre":"E001","bloqueado_por":[],"secuencia":9,"color":"rojo"}`)
	escribir(Activo+"/E001-F004.json", `{"id":"E001/F007","tipo":"feature","titulo":"T","estado":"LISTO","prioridad":"P1","motivo":"m","padre":"E001","bloqueado_por":[],"secuencia":10}`)
	escribir(Activo+"/E001-F005.json", `{"id":"E001/F005","tipo":"feature","titulo":"T","estado":"CERRADO","prioridad":"P1","motivo":"m","padre":"E001","bloqueado_por":[],"secuencia":11}`)
	escribir(Activo+"/E002-F001.json", `{"id":"E002/F001","tipo":"feature","titulo":"T","estado":"LISTO","prioridad":"P1","motivo":"m","padre":"E002","bloqueado_por":[],"secuencia":12}`)
	escribir(Archivo+"/E001-F006.json", `{"id":"E001/F006","tipo":"feature","titulo":"T","estado":"LISTO","prioridad":"P1","motivo":"m","padre":"E001","bloqueado_por":[],"secuencia":13}`)
	escribir(Activo+"/E001-F008.json", `{"id":"E001/F008","tipo":"slice","titulo":"T","estado":"LISTO","prioridad":"P1","motivo":"m","padre":"E001","bloqueado_por":[],"secuencia":1}`)
	escribir(Archivo+"/E001.json", `{"id":"E001","tipo":"iniciativa","titulo":"T","estado":"CERRADO","prioridad":"P1","motivo":"m","padre":"","bloqueado_por":[],"secuencia":20}`)
	escribir(Activo+"/x.json", `{"id":"E1","tipo":"iniciativa","titulo":"T","estado":"LISTO","prioridad":"P1","motivo":"m","padre":"","bloqueado_por":[],"secuencia":21}`)
	escribir(Activo+"/E001-F009.json", `{"id":"E001/F009","tipo":"feature","titulo":"T","estado":"LISTO","prioridad":"P1","motivo":"m","padre":"E001","bloqueado_por":["E005"],"secuencia":22}`)
	escribir(Activo+"/E001-F010.json", `{"id":"E001/F010","tipo":"feature","titulo":"T","estado":"LISTO","prioridad":"P1","motivo":"m","padre":"E001","bloqueado_por":[]}`)
	escribir(Activo+"/E001-F011.json", `{"id":"E001/F011","tipo":"feature","titulo":"T","estado":"LISTO","prioridad":"P1","motivo":"m","padre":"E001","bloqueado_por":null,"secuencia":24}`)
	escribir(Activo+"/E001-F012.json", `{"id":"E001/F012","tipo":"feature","titulo":"T","estado":"LISTO","prioridad":"P1","motivo":"m","padre":"E001","bloqueado_por":[],"secuencia":25} {}`)
	// E000/F005/S003: padre incoherente con tipo correcto (existe, pero no es el del id).
	escribir(Activo+"/E001-F013.json", `{"id":"E001/F013","tipo":"feature","titulo":"T","estado":"LISTO","prioridad":"P1","motivo":"m","padre":"E001/F002","bloqueado_por":[],"secuencia":26}`)
	// E000/F005/S003: cada ficha mala da su propio problema, en su propia línea («archivo: motivo»).
	problemas := Validar(raiz)
	A, R := Activo+"/", Archivo+"/"
	for _, c := range []struct{ archivo, motivo string }{
		{A + "roto.json", ""},
		{A + "E001-F003.json", "color"},
		{A + "E001-F004.json", "el nombre no cuadra con el id E001/F007"},
		{A + "E001-F005.json", "CERRADO"},
		{A + "E002-F001.json", "padre E002 no existe"},
		{R + "E001-F006.json", "LISTO"},
		{A + "E001-F008.json", `tipo o padre no cuadran con el id (feature, "E001")`},
		{A + "E001-F013.json", `tipo o padre no cuadran con el id (feature, "E001")`},
		{A + "x.json", `id "E1" inválido`},
		{A + "E001-F009.json", "bloqueado_por E005 no existe"},
		{A + "E001-F010.json", "falta el campo secuencia"},
		{A + "E001-F011.json", "null"},
		{A + "E001-F012.json", "texto después"},
	} {
		hay := false
		for _, l := range problemas {
			if strings.HasPrefix(l, c.archivo+": ") && strings.Contains(l, c.motivo) {
				hay = true
			}
		}
		if !hay {
			t.Errorf("Validar debe señalar %s con %q; dio:\n%s", c.archivo, c.motivo, strings.Join(problemas, "\n"))
		}
	}
	p := strings.Join(problemas, "\n")
	for _, w := range []string{"secuencia 1 repite", "id E001 repite"} {
		if !strings.Contains(p, w) {
			t.Errorf("Validar debe señalar %q; dio:\n%s", w, p)
		}
	}
}
