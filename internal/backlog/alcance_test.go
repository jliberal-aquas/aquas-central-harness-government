package backlog

import (
	"strings"
	"testing"
)

// E001/F002/S007: alcance vigilado (opción B, decisión humana 2026-10-21). El slice en curso
// declara sus rutas; tocar otra exige ampliar el alcance con motivo. Derivados y datos de aquas
// siempre permitidos. Sin slice en curso no se restringe.

func enCursoConAlcance(t *testing.T) (string, string) {
	t.Helper()
	raiz, id := t.TempDir(), "E001/F002/S007"
	fichaJSON(t, raiz, Ficha{ID: id, Titulo: "A", Estado: "EN_CURSO", Secuencia: 1})
	return raiz, id
}

func TestAlcanceSeDeclaraYAmplia(t *testing.T) {
	raiz, id := enCursoConAlcance(t)
	if _, err := Alcance(raiz, id, []string{"internal/a/**", "docs/x.md"}); err != nil {
		t.Fatal(err)
	}
	if _, err := Ampliar(raiz, id, "cmd/x/main.go", ""); err == nil {
		t.Errorf("ampliar sin motivo: error")
	}
	if _, err := Ampliar(raiz, id, "cmd/x/main.go", "hace falta el comando"); err != nil {
		t.Fatal(err)
	}
	poner(t, raiz, "gobierno/config/sellos.json", "{\"salidas\": {\"CLAUDE.md\": {}}}\n")
	for _, p := range []string{"internal/a/b/c.go", "internal/a/x.go", "docs/x.md", "cmd/x/main.go", "CLAUDE.md", "gobierno/config/sellos.json", "trabajo/activo/x.json", ".tmp/y"} {
		if m := FueraDeAlcance(raiz, p); m != "" {
			t.Errorf("%s está dentro (declarado, ampliado, derivado o dato de aquas); dio %q", p, m)
		}
	}
	for _, p := range []string{"docs/y.md", "internal/ab/x.go", "docs/x.md.bak"} {
		m := FueraDeAlcance(raiz, p)
		if !strings.Contains(m, "fuera del alcance de "+id) || !strings.Contains(m, "aquas backlog scope "+id+" extend "+p+" motivo=") {
			t.Errorf("%s está fuera: lo dice y cómo ampliar; dio %q", p, m)
		}
	}
	f, _, _ := Ver(raiz, id)
	if len(f.Ampliaciones) != 1 || f.Ampliaciones[0].Motivo != "hace falta el comando" {
		t.Errorf("la ampliación queda registrada con su motivo; dio %+v", f.Ampliaciones)
	}
}

func TestSinAlcanceDeclaradoBloquea(t *testing.T) {
	raiz, id := enCursoConAlcance(t)
	if m := FueraDeAlcance(raiz, "docs/y.md"); !strings.Contains(m, "declara el alcance") || !strings.Contains(m, "aquas backlog scope "+id) {
		t.Errorf("slice en curso sin alcance: frena y dice cómo declararlo; dio %q", m)
	}
}

func TestSinSliceEnCursoNoRestringe(t *testing.T) {
	raiz := t.TempDir()
	fichaJSON(t, raiz, Ficha{ID: "E001/F002/S007", Titulo: "A", Estado: "BORRADOR", Secuencia: 1})
	if m := FueraDeAlcance(raiz, "docs/y.md"); m != "" {
		t.Errorf("sin slice en curso no hay alcance que vigilar; dio %q", m)
	}
}
