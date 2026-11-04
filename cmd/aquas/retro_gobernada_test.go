package main

import (
	"fmt"
	"os"
	"os/exec"
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
	t.Run("S002 retro close y aviso en next", retroCloseYAviso)
}

// retroCloseYAviso (S002) sobre el binario: next avisa pasado el umbral, retro close lo apaga, la guardia
// no deja escribir registro/retros/ a mano y un rechazo sale con código 1.
func retroCloseYAviso(t *testing.T) {
	raiz := t.TempDir()
	if s := aquas(t, raiz, "", "backlog", "new", "id=E001", "titulo=I", "prioridad=P1", "motivo=m", "que_hace=Prueba la retro.", "prueba=./p:TestP=valida|rojo|verde", "relacion=nueva"); s.codigo != 0 {
		t.Fatalf("backlog new: %+v", s)
	}
	if err := os.MkdirAll(filepath.Join(raiz, "registro", "defectos"), 0o755); err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 10; i++ {
		if err := os.WriteFile(filepath.Join(raiz, "registro", "defectos", fmt.Sprintf("D%06d.json", i)), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// A1.3 no frena: next sale con 0 y muestra la ficha después del aviso.
	s := aquas(t, raiz, "", "backlog", "next")
	i, j := strings.Index(s.out, "toca retrospectiva: 0 slices y 10 defectos desde el inicio"), strings.LastIndex(s.out, "E001")
	if s.codigo != 0 || i < 0 || j < i {
		t.Errorf("next avisa en el umbral, sale con 0 y muestra la ficha tras el aviso; dio %+v", s)
	}
	// 2A: la retro del 2-nov se registra con la cuenta de su commit, no la de hoy.
	const fechaVieja = "2026-11-02T22:07:54Z"
	git := func(args ...string) string {
		t.Helper()
		c := exec.Command("git", append([]string{"-C", raiz, "-c", "user.name=t", "-c", "user.email=t@t"}, args...)...)
		c.Env = append(os.Environ(), "GIT_COMMITTER_DATE="+fechaVieja, "GIT_AUTHOR_DATE="+fechaVieja) // A3.1: fecha del commit ≠ hoy
		b, err := c.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, b)
		}
		return strings.TrimSpace(string(b))
	}
	git("init", "-q")
	git("add", "-A")
	git("commit", "-q", "-m", "retro vieja")
	viejo := git("rev-parse", "HEAD")
	for i := 11; i <= 13; i++ {
		if err := os.WriteFile(filepath.Join(raiz, "registro", "defectos", fmt.Sprintf("D%06d.json", i)), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// 1A: sin «apruebo retro» del humano no se cierra.
	if s := aquas(t, raiz, "", "retro", "close", "fichas=E001", "motivo=m"); s.codigo != 1 || !strings.Contains(s.err, "apruebo retro") {
		t.Errorf("retro close sin «apruebo retro»: código 1 y lo nombra; dio %+v", s)
	}
	// A2.1: el hook con un registro inventado guarda la aprobación, pero retro close no la acepta.
	if s := aquas(t, raiz, `{"hook_event_name":"UserPromptSubmit","prompt":"apruebo retro","transcript_path":"r.jsonl"}`, "approval"); s.codigo != 0 {
		t.Fatalf("hook approval con registro inventado: %+v", s)
	}
	if s := aquas(t, raiz, "", "retro", "close", "fichas=E001", "motivo=m"); s.codigo != 1 || !strings.Contains(s.err, "apruebo retro") {
		t.Errorf("aprobación sin mensaje humano en el registro: código 1; dio %+v", s)
	}
	if s := aquas(t, raiz, mensajeHumano(t, "apruebo retro"), "approval"); s.codigo != 0 {
		t.Fatalf("hook approval con «apruebo retro»: %+v", s)
	}
	// A1.6: ficha repetida, clave desconocida o repetida y commit inexistente salen con 1 sin gastar la aprobación.
	for _, args := range [][]string{
		{"fichas=E001,E001", "motivo=m"}, {"fichas=E001", "motivo=m", "x=1"}, {"fichas=E001", "motivo=m", "motivo=n"},
		{"fichas=E001", "motivo=m", "commit=no-existe"},
	} {
		if s := aquas(t, raiz, "", append([]string{"retro", "close"}, args...)...); s.codigo != 1 {
			t.Errorf("retro close %v: código 1; dio %+v", args, s)
		}
	}
	if s := aquas(t, raiz, "", "retro", "close", "fichas=E001", "motivo=retro vieja", "commit="+viejo); s.codigo != 0 || !strings.Contains(s.out, "R0001 registrada · 0 slices cerrados y 10 defectos") {
		t.Fatalf("retro close commit=: cuenta la del commit; %+v", s)
	}
	b, err := os.ReadFile(filepath.Join(raiz, "registro", "retros", "R0001.json"))
	if err != nil || !strings.Contains(string(b), `"head": "`+viejo+`"`) || !strings.Contains(string(b), `"fecha": "`+fechaVieja+`"`) || !strings.Contains(string(b), `"motivo": "retro vieja"`) || !strings.Contains(string(b), `"mensaje": "apruebo retro"`) {
		t.Errorf("R0001 guarda el commit, el motivo y la aprobación; dio %s %v", b, err)
	}
	if s := aquas(t, raiz, "", "backlog", "next"); strings.Contains(s.out, "toca retrospectiva") {
		t.Errorf("3 defectos desde R0001: sin aviso; dio %+v", s)
	}
	if s := aquas(t, repoSano(t), `{"tool_name":"Write","tool_input":{"file_path":"registro/retros/R0002.json"}}`, "guard"); s.codigo != 2 || !strings.Contains(s.err, "aquas retro close") {
		t.Errorf("la guardia debe frenar registro/retros/ a mano; dio %+v", s)
	}
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
