package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// sellarCLI marca sin_cambio cada ítem pendiente que lista `aquas backlog sello <id>`.
func sellarCLI(t *testing.T, raiz, id string) {
	t.Helper()
	s := aquas(t, raiz, "", "backlog", "sello", id)
	if s.codigo != 0 {
		t.Fatalf("sello %s: %+v", id, s)
	}
	for _, l := range strings.Split(s.out, "\n") {
		if c, ok := strings.CutPrefix(l, "  [ ] "); ok {
			clave, _, _ := strings.Cut(c, " · ")
			if r := aquas(t, raiz, "", "backlog", "sello", id, clave, "sin_cambio", "motivo=revisado"); r.codigo != 0 {
				t.Fatalf("marcar %s: %+v", clave, r)
			}
		}
	}
}

// E001/F002/S004: por el CLI real, el sello se lista, frena el cierre y se marca.
func TestSelloCLIDePuntaAPunta(t *testing.T) {
	raiz := repoSano(t)
	g := func(args ...string) string {
		out, err := exec.Command("git", append([]string{"-C", raiz, "-c", "user.name=t", "-c", "user.email=t@t"}, args...)...).CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	g("add", "-A")
	g("commit", "-q", "-m", "base")
	reg := map[string]any{"fecha": "2026-10-21T10:00:00Z", "head": g("rev-parse", "HEAD"), "resultado": map[string]string{"./p:TestA": "falló"}}
	for _, f := range []map[string]any{
		{"id": "E001", "tipo": "iniciativa", "padre": "", "estado": "EN_CURSO", "secuencia": 1},
		{"id": "E001/F001", "tipo": "feature", "padre": "E001", "estado": "EN_CURSO", "secuencia": 2},
		{"id": "E001/F001/S001", "tipo": "slice", "padre": "E001/F001", "estado": "EN_CURSO", "secuencia": 3,
			"pruebas": []map[string]string{{"prueba": "./p:TestA", "intencion": "a"}}, "rojo": reg, "verde": reg},
	} {
		f["titulo"], f["prioridad"], f["motivo"], f["bloqueado_por"] = "t", "P1", "m", []string{}
		b, _ := json.MarshalIndent(f, "", "  ")
		p := filepath.Join(raiz, "trabajo", "activo", strings.ReplaceAll(f["id"].(string), "/", "-")+".json")
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	s := aquas(t, raiz, "", "backlog", "sello", "E001/F001/S001")
	if s.codigo != 0 || !strings.Contains(s.out, "  [ ] ficha:E001/F001 · ") {
		t.Fatalf("sello lista los ítems pendientes con [ ]; dio %+v", s)
	}
	if s := aquas(t, raiz, "", "backlog", "estado", "E001/F001/S001", "CERRADO"); s.codigo != 1 || !strings.Contains(s.err, "sello incompleto") {
		t.Fatalf("sin marcas el cierre se niega; dio %+v", s)
	}
	sellarCLI(t, raiz, "E001/F001/S001")
	if s := aquas(t, raiz, "", "backlog", "sello", "E001/F001/S001"); strings.Contains(s.out, "[ ]") || !strings.Contains(s.out, "  [x] ficha:E001 · sin_cambio: revisado") {
		t.Fatalf("tras marcar, todo con [x], marca y motivo; dio %+v", s)
	}
	if s := aquas(t, raiz, "", "backlog", "estado", "E001/F001/S001", "CERRADO"); s.codigo != 0 {
		t.Fatalf("con el sello completo cierra; dio %+v", s)
	}
}
