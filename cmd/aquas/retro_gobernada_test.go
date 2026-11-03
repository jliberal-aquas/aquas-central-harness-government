package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRetrospectivaGobernada (E001/F018): integral de la retrospectiva gobernada. Crece con cada slice cerrado.
// Rojo cuando: pasado el umbral next no pide retro o dos defectos de la misma familia no suman.
// Verde cuando: next pide retro en el umbral y la familia acumula ocurrencias.
func TestRetrospectivaGobernada(t *testing.T) {
	t.Run("S003 defecto exige familia de la lista", defectoExigeFamilia)
	t.Run("S001 /retro viaja al hijo", retroViajaAlHijo)
}

// retroViajaAlHijo (S001): install deja en el hijo /retro junto a /backlog, con el aquas instalado (principio 11).
func retroViajaAlHijo(t *testing.T) {
	dir := t.TempDir()
	if s := aquas(t, dir, "", "install"); s.codigo != 0 {
		t.Fatalf("install: exit %d\n%s%s", s.codigo, s.out, s.err)
	}
	for _, nombre := range []string{"backlog", "retro"} {
		b, err := os.ReadFile(filepath.Join(dir, ".claude", "skills", nombre, "SKILL.md"))
		if err != nil || !strings.HasPrefix(string(b), "---\n") || strings.Contains(string(b), "aquas.sh") {
			t.Errorf("/%s del hijo: debe existir, abrir con frontmatter y no usar aquas.sh; %v\n%s", nombre, err, b)
		}
	}
	if b, _ := os.ReadFile(filepath.Join(dir, ".claude", "skills", "retro", "SKILL.md")); !strings.Contains(string(b), "## Reparto") {
		t.Errorf("/retro del hijo sin su regla de reparto:\n%s", b)
	}
	// Choque: un /retro propio del hijo no se pisa; install falla nombrándolo y no copia la central.
	propio := t.TempDir()
	p := filepath.Join(propio, ".claude", "skills", "retro", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("mi retro\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Hallazgo A2.1: el rechazo es el chequeo previo de install, no el de build tras copiar la central.
	if s := aquas(t, propio, "", "install"); s.codigo == 0 || !strings.Contains(s.err, "install: ") || !strings.Contains(s.err, "ya tiene un /retro propio") {
		t.Errorf("install con /retro propio: debió fallar antes de copiar, nombrándolo; %+v", s)
	}
	if _, err := os.Stat(filepath.Join(propio, ".aquas")); err == nil {
		t.Errorf("install con /retro propio: no debió copiar la central")
	}
	if b, _ := os.ReadFile(p); string(b) != "mi retro\n" {
		t.Errorf("install pisó el /retro propio")
	}
}

// defectoExigeFamilia (S003) sobre el binario: `defect new` sin familia o con una inventada falla y nombra las
// válidas; con una de la lista pasa y la guarda.
func defectoExigeFamilia(t *testing.T) {
	raiz := t.TempDir()
	if err := os.MkdirAll(filepath.Join(raiz, "p"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(raiz, "p", "p_test.go"), []byte("package p\n\nimport \"testing\"\n\nfunc TestP(t *testing.T) {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	base := []string{"defect", "new", "que=q", "causa=c", "arreglo=a", "cercioro=x", "relacion=nuevo", "prueba=./p:TestP"}
	for _, c := range []struct {
		extra  []string
		codigo int
	}{
		{[]string{"firma=sin"}, 1},
		{[]string{"firma=inventada", "familia=inventada"}, 1},
		{[]string{"firma=buena", "familia=freno"}, 0},
	} {
		s := aquas(t, raiz, "", append(append([]string{}, base...), c.extra...)...)
		if s.codigo != c.codigo {
			t.Errorf("defect new %v: código %d, quiere %d; %+v", c.extra, s.codigo, c.codigo, s)
		}
		if c.codigo != 0 && !strings.Contains(s.err, "freno") {
			t.Errorf("defect new %v: el rechazo nombra las familias válidas; dio %q", c.extra, s.err)
		}
	}
	b, err := os.ReadFile(filepath.Join(raiz, "registro", "defectos", "D000001.json"))
	if err != nil || !strings.Contains(string(b), `"familia": "freno"`) {
		t.Errorf("el defecto guarda su familia; dio %s %v", b, err)
	}
	// S001 (hallazgo A1.1): defect show muestra la familia, que /retro usa para agrupar.
	if s := aquas(t, raiz, "", "defect", "show"); !strings.Contains(s.out, "familia freno") {
		t.Errorf("defect show debe mostrar la familia; dio %q", s.out)
	}
	// Gate: un defecto viejo sin familia solo avisa.
	viejo := `{"id":"D000002","firma":"v","que":"q","causa":"c","arreglo":"a","como_me_cercioro":"x","ocurrencias":[],"prueba":"./p:TestP"}`
	if err := os.WriteFile(filepath.Join(raiz, "registro", "defectos", "D000002.json"), []byte(viejo), 0o644); err != nil {
		t.Fatal(err)
	}
	if frena, aviso := gateDefectos(raiz); len(frena) != 0 || !strings.Contains(aviso, "1 defectos anteriores a E001/F018/S003 sin familia") {
		t.Errorf("un viejo sin familia no frena y avisa; dio %v, %q", frena, aviso)
	}
}
