package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/backlog"
	"github.com/jliberal-aquas/aquas-central-harness-government/internal/wiki"
)

// TestEditaQueHace (E000/F007/S002): `aquas backlog edit <id> que_hace=… motivo=…` corrige el «qué hace» de una ficha
// en cualquier estado, activa o archivada, con la regla de `new`.
// Rojo cuando: edit rechaza que_hace.
// Verde cuando: edit lo cambia y la tabla de la feature lo muestra.
func TestEditaQueHace(t *testing.T) {
	raiz := repoFeature(t)
	escribir(t, raiz, "docs/E001/F002/README.md", "# F002\n\n## Para qué existe\n\nContrato de la feature.\n\n## Navegación\n")
	escribir(t, raiz, "docs/E001/F002/S003.md", "# S003\n\n## Para qué existe\n\nContrato del slice.\n")
	queHace := func(rel string) string {
		t.Helper()
		var q string
		_ = json.Unmarshal(fichaEnDisco(t, raiz, rel)["que_hace"], &q)
		return q
	}
	editar := func(args ...string) salida {
		t.Helper()
		return aquas(t, raiz, "", append([]string{"backlog", "edit"}, args...)...)
	}
	huella := func() string {
		t.Helper()
		f, _, err := backlog.Ver(raiz, "E001/F002/S003")
		if err != nil {
			t.Fatal(err)
		}
		h, err := backlog.Huella(raiz, f)
		if err != nil {
			t.Fatal(err)
		}
		return h
	}
	antesHuella := huella()

	// Slice y feature activos.
	if s := editar("E001/F002/S003", "que_hace=marca el defecto como arreglado", "motivo=el texto quedó viejo"); s.codigo != 0 {
		t.Fatalf("slice en BORRADOR: %+v", s)
	}
	if q := queHace("trabajo/activo/E001-F002-S003.json"); q != "marca el defecto como arreglado" {
		t.Errorf("el slice guarda el texto nuevo; quedó %q", q)
	}
	if s := editar("E001/F002", "que_hace=agrupa los arreglos de defectos", "motivo=m"); s.codigo != 0 {
		t.Errorf("feature: %+v", s)
	}
	// Choque con la aprobación: que_hace no está en la huella.
	if huella() != antesHuella {
		t.Errorf("editar que_hace no cambia la huella de aprobación")
	}
	// La tabla del padre muestra el texto nuevo.
	if _, err := wiki.Generar(raiz); err != nil { // lo que corre `aquas compile` sobre docs/
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(filepath.Join(raiz, "docs", "E001", "F002", "README.md")); !strings.Contains(string(b), "marca el defecto como arreglado") {
		t.Errorf("la tabla de la feature muestra el texto nuevo:\n%s", b)
	}

	// Slice CERRADO y archivado: que_hace se corrige; el título no.
	arch := `{"id": "E001/F002/S004", "tipo": "slice", "titulo": "Viejo", "estado": "CERRADO", "prioridad": "P1", "motivo": "m", "padre": "E001/F002", "bloqueado_por": [], "secuencia": 2, "que_hace": "texto viejo"}`
	escribir(t, raiz, "trabajo/archivo/E001-F002-S004.json", arch)
	if s := editar("E001/F002/S004", "que_hace=texto corregido", "motivo=quedó viejo"); s.codigo != 0 {
		t.Errorf("slice archivado: %+v", s)
	}
	if q := queHace("trabajo/archivo/E001-F002-S004.json"); q != "texto corregido" {
		t.Errorf("el archivado guarda el texto nuevo en su carpeta; quedó %q", q)
	}
	if _, err := os.Stat(filepath.Join(raiz, "trabajo", "activo", "E001-F002-S004.json")); err == nil {
		t.Errorf("editar un archivado no lo devuelve a activo")
	}
	if s := editar("E001/F002/S004", "titulo=Nuevo", "motivo=m"); s.codigo == 0 {
		t.Errorf("el título de un archivado sigue sin editarse")
	}

	// Entradas malas: no cambian la ficha.
	largo := strings.TrimSpace(strings.Repeat("palabra ", 21))
	for _, c := range []struct {
		caso string
		args []string
	}{
		{"que_hace vacío", []string{"E001/F002/S003", "que_hace=  ", "motivo=m"}},
		{"que_hace de 21 palabras", []string{"E001/F002/S003", "que_hace=" + largo, "motivo=m"}},
		{"sin motivo", []string{"E001/F002/S003", "que_hace=otro texto"}},
		{"que_hace repetido", []string{"E001/F002/S003", "que_hace=a", "que_hace=b", "motivo=m"}},
		{"ficha inexistente", []string{"E001/F002/S099", "que_hace=otro", "motivo=m"}},
	} {
		if s := editar(c.args...); s.codigo == 0 {
			t.Errorf("%s: debía fallar; %+v", c.caso, s)
		}
	}
	if q := queHace("trabajo/activo/E001-F002-S003.json"); q != "marca el defecto como arreglado" {
		t.Errorf("los rechazos no cambian la ficha; quedó %q", q)
	}
}
