package memoria

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// E001/F005/S001: memoria caliente con campos fijos, escrita solo por aquas y sellada.

var ahora = time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)

func fijar(t *testing.T, raiz string, pares ...string) Memoria {
	t.Helper()
	m, err := Fijar(raiz, pares, ahora)
	if err != nil {
		t.Fatalf("Fijar(%q): %v", pares, err)
	}
	return m
}

func TestFijarYLeer(t *testing.T) {
	raiz := t.TempDir()
	fijar(t, raiz, "iniciativa=E001", "feature=F005", "slice=S001", "tarea=tests en rojo",
		"siguiente=programar el paquete", "pendiente=guardia", "pendiente=sello", "decision=¿archivo?")
	m, err := Leer(raiz)
	if err != nil {
		t.Fatal(err)
	}
	if m.Foco != (Foco{"E001", "F005", "S001", "tests en rojo"}) || m.Siguiente != "programar el paquete" {
		t.Fatalf("foco o siguiente distintos: %+v", m)
	}
	if strings.Join(m.Pendientes, "|") != "guardia|sello" || strings.Join(m.Decisiones, "|") != "¿archivo?" {
		t.Fatalf("listas distintas: %+v", m)
	}
	if m.Actualizado != "2026-10-20T12:00:00Z" || m.Sello == "" {
		t.Fatalf("falta actualizado o sello: %+v", m)
	}
}

func TestFijarConservaLoNoTocado(t *testing.T) {
	raiz := t.TempDir()
	fijar(t, raiz, "iniciativa=E001", "feature=F005", "pendiente=a", "pendiente=b")
	m := fijar(t, raiz, "tarea=otra", "pendiente=c")
	if m.Foco.Feature != "F005" || m.Foco.Tarea != "otra" {
		t.Fatalf("debe conservar feature y cambiar tarea: %+v", m.Foco)
	}
	if strings.Join(m.Pendientes, "|") != "c" {
		t.Fatalf("una lista nombrada se reemplaza entera: %q", m.Pendientes)
	}
	if m = fijar(t, raiz, "pendiente="); len(m.Pendientes) != 0 {
		t.Fatalf("«pendiente=» vacía la lista: %q", m.Pendientes)
	}
}

func TestFijarRechaza(t *testing.T) {
	for _, c := range []struct{ par, espera string }{
		{"color=rojo", "campo desconocido: color"},
		{"tarea", "se espera campo=valor"},
		{"iniciativa=X1", "iniciativa debe ser E y 3 dígitos"},
		{"feature=F5", "feature debe ser F y 3 dígitos"},
		{"slice=s001", "slice debe ser S y 3 dígitos"},
	} {
		raiz := t.TempDir()
		_, err := Fijar(raiz, []string{c.par}, ahora)
		if err == nil || !strings.Contains(err.Error(), c.espera) {
			t.Errorf("%q: se esperaba error con %q; dio %v", c.par, c.espera, err)
		}
		if _, err := os.Stat(filepath.Join(raiz, Ruta)); err == nil {
			t.Errorf("%q: un error no debe escribir la memoria", c.par)
		}
	}
}

func TestSelloDetectaEdicionAMano(t *testing.T) {
	raiz := t.TempDir()
	fijar(t, raiz, "iniciativa=E001", "tarea=original")
	p := filepath.Join(raiz, Ruta)
	b, _ := os.ReadFile(p)
	if err := os.WriteFile(p, []byte(strings.Replace(string(b), "original", "trucada", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Leer(raiz); !errors.Is(err, ErrSelloRoto) {
		t.Fatalf("edición fuera de aquas debe dar ErrSelloRoto; dio %v", err)
	}
	if _, err := Fijar(raiz, []string{"tarea=x"}, ahora); !errors.Is(err, ErrSelloRoto) {
		t.Fatalf("fijar sobre memoria trucada no debe tapar el fallo; dio %v", err)
	}
}

func TestSinMemoria(t *testing.T) {
	if _, err := Leer(t.TempDir()); !errors.Is(err, ErrSinMemoria) {
		t.Fatalf("sin archivo debe dar ErrSinMemoria; dio %v", err)
	}
}

func TestRutasDelFoco(t *testing.T) {
	r := Rutas(Foco{Iniciativa: "E001", Feature: "F005", Slice: "S001"})
	quiere := Mapa{
		Iniciativa: "docs/E001/README.md",
		Feature:    "docs/E001/F005/README.md",
		Slice:      "docs/E001/F005/S001.md",
		Backlog:    "trabajo/activo",
	}
	if r != quiere {
		t.Fatalf("rutas: %+v", r)
	}
	if r := Rutas(Foco{Iniciativa: "E001"}); r.Feature != "" || r.Slice != "" {
		t.Fatalf("sin feature ni slice no se inventan rutas: %+v", r)
	}
}

func TestVerMuestraFocoYRutas(t *testing.T) {
	raiz := t.TempDir()
	m := fijar(t, raiz, "iniciativa=E001", "feature=F005", "slice=S001", "tarea=x", "siguiente=y", "pendiente=p1")
	txt := Texto(m)
	for _, s := range []string{"E001 › F005 › S001", "tarea: x", "siguiente: y", "- p1",
		"docs/E001/F005/S001.md", "backlog: trabajo/activo", "aquas memory set"} { // E000/F005/S003: literal, no la constante
		if !strings.Contains(txt, s) {
			t.Errorf("ver debe mostrar %q; dio:\n%s", s, txt)
		}
	}
}

// E001/F002/S003: la ruta del backlog de la memoria apunta a las fichas reales, no al canónico viejo.
func TestRutaBacklogEsTrabajo(t *testing.T) {
	r := Rutas(Foco{})
	if r.Backlog != "trabajo/activo" {
		t.Fatalf("backlog = %q; se esperaba trabajo/activo", r.Backlog)
	}
	if st, err := os.Stat(filepath.Join("..", "..", filepath.FromSlash(r.Backlog))); err != nil || !st.IsDir() {
		t.Fatalf("la ruta del backlog debe existir en el repo: %v", err)
	}
}
