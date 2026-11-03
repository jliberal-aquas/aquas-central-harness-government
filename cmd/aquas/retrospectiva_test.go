package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/wiki"
)

// TestRetrospectivaSinReincidencia (E001/F017): cada familia de la retrospectiva del 2026-11-02 tiene su control,
// con un caso que debe fallar y uno que debe pasar. Crece con cada slice cerrado.
// Rojo cuando: una familia con slice cerrado vuelve a pasar sin freno.
// Verde cuando: el control de cada slice cerrado frena su caso malo y deja pasar el bueno.
func TestRetrospectivaSinReincidencia(t *testing.T) {
	t.Run("S001 freno atado al trabajo activo", TestFrenoSigueAlTrabajoActivo)
	t.Run("S002 no pedir apruebo ya cubierto", TestNoPideAprueboCubierto)
	t.Run("S003 turnos separan humano de agente", turnosSeparanOrigen)
	t.Run("S004 turno guarda tipo y pedido refinado", turnoExigeRefinado)
	t.Run("S005 defecto nombra la prueba", defectoExigePrueba)
	t.Run("S007 corrección exige defecto", TestCorreccionExigeDefecto)
	t.Run("S008 escribir exige leer el contrato", escribirExigeLeerContrato)
	t.Run("S009 freno por turno cede con feature aprobada", TestFrenoTurnoCedeConFeatureAprobada)
	t.Run("S006 página al día con su código", paginaAlDia)
}

// paginaAlDia (S006): con el slice EN_CURSO, código sin commit y su página sin tocar desde el commit, la revisión
// de la wiki nombra la página; al editar su texto, pasa.
func paginaAlDia(t *testing.T) {
	raiz := t.TempDir()
	escribir(t, raiz, "trabajo/activo/E001-F002-S001.json", `{"id": "E001/F002/S001", "tipo": "slice", "titulo": "A", "estado": "EN_CURSO", "prioridad": "P1", "motivo": "m", "padre": "E001/F002", "bloqueado_por": [], "secuencia": 1}`)
	escribir(t, raiz, "docs/E001/F002/S001.md", "# S001\n\ncontrato v1\n")
	escribir(t, raiz, "internal/x/x.go", "package x\n")
	for _, a := range [][]string{{"init", "-q"}, {"add", "-A"}, {"commit", "-q", "-m", "base"}} {
		if out, err := exec.Command("git", append([]string{"-C", raiz, "-c", "user.name=t", "-c", "user.email=t@t"}, a...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", a, err, out)
		}
	}
	escribir(t, raiz, "internal/x/x.go", "package x\n\nvar A = 1\n")
	if ps := wiki.RevisarAlDia(raiz); len(ps) != 1 || !strings.Contains(ps[0], "docs/E001/F002/S001.md") {
		t.Errorf("código cambiado y página sin tocar: la revisión nombra la página; dio %v", ps)
	}
	escribir(t, raiz, "docs/E001/F002/S001.md", "# S001\n\ncontrato v2\n")
	if ps := wiki.RevisarAlDia(raiz); len(ps) != 0 {
		t.Errorf("página editada después del código: pasa; dio %v", ps)
	}
}

// escribirExigeLeerContrato (S008) sobre el binario: con un slice EN_CURSO, `aquas guard` frena escribir en su
// alcance si el registro no muestra lectura de su página; con la lectura pasa.
func escribirExigeLeerContrato(t *testing.T) {
	raiz := repoSano(t)
	dir := filepath.Join(raiz, "trabajo", "activo")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	f := `{"id": "E001/F002/S007", "tipo": "slice", "titulo": "A", "estado": "EN_CURSO", "prioridad": "P1", "motivo": "m",
  "padre": "E001/F002", "bloqueado_por": [], "secuencia": 1, "alcance": ["internal/**"]}`
	if err := os.WriteFile(filepath.Join(dir, "E001-F002-S007.json"), []byte(f), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		linea  string
		codigo int
	}{
		{`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Read","input":{"file_path":"x.md"}}]}}`, 2},
		{`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"cat docs/E001/F002/S007.md"}}]}}`, 0},
	} {
		reg := filepath.Join(raiz, "registro.jsonl")
		if err := os.WriteFile(reg, []byte(c.linea+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		b, _ := json.Marshal(map[string]any{"tool_name": "Write", "tool_input": map[string]string{"file_path": "internal/x.go"}, "transcript_path": reg, "cwd": raiz})
		if s := aquas(t, raiz, string(b), "guard"); s.codigo != c.codigo || c.codigo == 2 && !strings.Contains(s.err, "docs/E001/F002/S007.md") {
			t.Errorf("registro %s: código %d, quiere %d; %+v", c.linea, s.codigo, c.codigo, s)
		}
	}
}

// defectoExigePrueba (S005) sobre el binario: `defect new` sin prueba o con una que no existe falla; con una real pasa.
func defectoExigePrueba(t *testing.T) {
	raiz := t.TempDir()
	if err := os.MkdirAll(filepath.Join(raiz, "p"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(raiz, "p", "p_test.go"), []byte("package p\n\nimport \"testing\"\n\nfunc TestP(t *testing.T) {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	base := []string{"defect", "new", "firma=f", "que=q", "causa=c", "arreglo=a", "cercioro=x", "relacion=nuevo", "familia=freno"}
	for _, c := range []struct {
		extra  []string
		codigo int
	}{
		{nil, 1},
		{[]string{"prueba=./p:TestNoHay"}, 1},
		{[]string{"prueba=./p:TestP"}, 0},
	} {
		if s := aquas(t, raiz, "", append(append([]string{}, base...), c.extra...)...); s.codigo != c.codigo {
			t.Errorf("defect new %v: código %d, quiere %d; %+v", c.extra, s.codigo, c.codigo, s)
		}
	}
	// Gate: un defecto viejo sin campo solo avisa; la prueba nombrada que desaparece frena.
	viejo := `{"id":"D000002","firma":"v","que":"q","causa":"c","arreglo":"a","como_me_cercioro":"x","ocurrencias":[]}`
	if err := os.WriteFile(filepath.Join(raiz, "registro", "defectos", "D000002.json"), []byte(viejo), 0o644); err != nil {
		t.Fatal(err)
	}
	if frena, aviso := gateDefectos(raiz); len(frena) != 0 || !strings.Contains(aviso, "1 defectos") {
		t.Errorf("con la prueba presente: no frena y avisa 1 viejo; dio %v, %q", frena, aviso)
	}
	if err := os.Remove(filepath.Join(raiz, "p", "p_test.go")); err != nil {
		t.Fatal(err)
	}
	if frena, _ := gateDefectos(raiz); len(frena) != 1 || !strings.Contains(frena[0], "D000001") {
		t.Errorf("la prueba borrada frena el gate y nombra D000001; dio %v", frena)
	}
}

// turnoExigeRefinado (S004) sobre el binario: el cierre de un turno humano sin «Tipo» y «Pedido refinado» frena;
// con las dos líneas pasa y el turno las guarda.
func turnoExigeRefinado(t *testing.T) {
	raiz := t.TempDir()
	for _, c := range []struct {
		respuesta string
		codigo    int
	}{
		{"INFORMACIÓN: hecho.", 2},
		{"Tipo: petición\nPedido refinado: hacer X para Y.\n\nINFORMACIÓN: hecho.", 0},
	} {
		b, _ := json.Marshal(map[string]string{"session_id": "h", "prompt": "haz X"})
		if s := aquas(t, raiz, string(b), "turn", "open"); s.codigo != 0 {
			t.Fatalf("turn open: %+v", s)
		}
		b, _ = json.Marshal(map[string]string{"session_id": "h", "last_assistant_message": c.respuesta})
		if s := aquas(t, raiz, string(b), "turn", "close"); s.codigo != 0 {
			t.Fatalf("turn close: %+v", s)
		}
		if s := aquas(t, raiz, string(b), "proposal"); s.codigo != c.codigo {
			t.Errorf("cierre con %q: código %d, quiere %d; %+v", c.respuesta, s.codigo, c.codigo, s)
		}
	}
	turnos, _ := filepath.Glob(filepath.Join(raiz, "auditoria", "turnos", "*", "T000002.md"))
	if len(turnos) != 1 {
		t.Fatalf("dos turnos abiertos; dio %v", turnos)
	}
	if b, _ := os.ReadFile(turnos[0]); !strings.Contains(string(b), "- Tipo: petición\n- Pedido refinado: hacer X para Y.\n") {
		t.Errorf("el turno guarda tipo y refinado:\n%s", b)
	}
}

// turnosSeparanOrigen (S003) sobre el binario: el hook de turnos guarda el origen de cada mensaje.
func turnosSeparanOrigen(t *testing.T) {
	raiz := t.TempDir()
	for _, c := range []struct{ sesion, prompt, origen string }{
		{"h", "dale, sigue", "humano"},
		{"a", "<agent-message from=\"x\">informe</agent-message>", "agente"},
	} {
		b, _ := json.Marshal(map[string]string{"session_id": c.sesion, "prompt": c.prompt})
		if s := aquas(t, raiz, string(b), "turn", "open"); s.codigo != 0 {
			t.Fatalf("turn open: %+v", s)
		}
	}
	turnos, _ := filepath.Glob(filepath.Join(raiz, "auditoria", "turnos", "*", "T*.md"))
	var origenes []string
	for _, f := range turnos {
		b, _ := os.ReadFile(f)
		for _, o := range []string{"humano", "agente"} {
			if strings.Contains(string(b), "- Origen: "+o+"\n") {
				origenes = append(origenes, o)
			}
		}
	}
	if strings.Join(origenes, ",") != "humano,agente" {
		t.Errorf("los dos turnos deben quedar con origen humano y agente, en ese orden; dio %v", origenes)
	}
}
