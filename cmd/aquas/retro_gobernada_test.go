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
	// Gate: un defecto viejo sin familia solo avisa.
	viejo := `{"id":"D000002","firma":"v","que":"q","causa":"c","arreglo":"a","como_me_cercioro":"x","ocurrencias":[],"prueba":"./p:TestP"}`
	if err := os.WriteFile(filepath.Join(raiz, "registro", "defectos", "D000002.json"), []byte(viejo), 0o644); err != nil {
		t.Fatal(err)
	}
	if frena, aviso := gateDefectos(raiz); len(frena) != 0 || !strings.Contains(aviso, "1 defectos anteriores a E001/F018/S003 sin familia") {
		t.Errorf("un viejo sin familia no frena y avisa; dio %v, %q", frena, aviso)
	}
}
