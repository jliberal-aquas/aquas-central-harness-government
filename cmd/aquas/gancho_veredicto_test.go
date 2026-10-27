package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fichaSlice escribe la ficha activa de un slice en el estado dado, con alcance p/** y p/p.go.
func fichaSlice(t *testing.T, raiz, id, estado string) {
	t.Helper()
	f := map[string]any{"id": id, "tipo": "slice", "titulo": "t", "estado": estado, "prioridad": "P1", "motivo": "m",
		"padre": "E001/F014", "bloqueado_por": []string{}, "secuencia": 1, "alcance": []string{"p/**"},
		"pruebas": []map[string]string{{"prueba": "./p:TestA", "intencion": "a"}}}
	b, _ := json.MarshalIndent(f, "", "  ")
	escribir(t, raiz, "trabajo/activo/"+strings.ReplaceAll(id, "/", "-")+".json", string(b))
}

// auditoriaDe lee el campo auditoria de la ficha activa; nil si no hay.
func auditoriaDe(t *testing.T, raiz, id string) map[string]any {
	t.Helper()
	m := fichaEnDisco(t, raiz, "trabajo/activo/"+strings.ReplaceAll(id, "/", "-")+".json")
	raw, ok := m["auditoria"]
	if !ok {
		return nil
	}
	var a map[string]any
	if err := json.Unmarshal(raw, &a); err != nil {
		t.Fatal(err)
	}
	return a
}

// cierreAuditor arma la entrada real de SubagentStop: sin last_assistant_message; el cierre va como
// SubagentHandback en el transcript del subagente (agent_transcript_path).
func cierreAuditor(t *testing.T, tipo, cuerpo string, activo bool) string {
	t.Helper()
	linea, _ := json.Marshal(map[string]any{"type": "assistant", "message": map[string]any{"role": "assistant",
		"content": []any{map[string]any{"type": "tool_use", "name": "SubagentHandback", "input": map[string]any{"message": cuerpo}}}}})
	ruta := filepath.Join(t.TempDir(), "agent.jsonl")
	if err := os.WriteFile(ruta, append(linea, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(map[string]any{"agent_type": tipo, "agent_transcript_path": ruta, "stop_hook_active": activo})
	return string(b)
}

// Rojo cuando: el veredicto no llega a la ficha, se anota un cierre de worker o pasa un cierre de auditor sin id.
// Verde cuando: el auditor con id anota fecha, veredicto y huella (gana el último), y sin id o con id fuera de curso se rechaza.
func TestGanchoAnotaVeredicto(t *testing.T) {
	t.Run("H1: el cierre real (transcript, sin last_assistant_message) queda anotado en la ficha", func(t *testing.T) {
		raiz := t.TempDir()
		escribir(t, raiz, "p/p.go", "package p\n")
		fichaSlice(t, raiz, "E001/F014/S002", "EN_CURSO")
		entrada := cierreAuditor(t, "auditor", "id: E001-F014-S002-A1\nveredicto: APROBADO\nhallazgos: —\nno_revisado: —", false)
		if strings.Contains(entrada, "last_assistant_message") {
			t.Fatal("la entrada debe tener la forma real, sin last_assistant_message")
		}
		if s := aquas(t, raiz, entrada, "cierre-agente"); s.codigo != 0 {
			t.Fatalf("exit 0; dio %+v", s)
		}
		if a := auditoriaDe(t, raiz, "E001/F014/S002"); a == nil || a["veredicto"] != "APROBADO" {
			t.Errorf("ficha: se esperaba APROBADO; dio %v", a)
		}
	})
	t.Run("D000053: el cierre sale del transcript del subagente", probarUltimoMensaje)
	const id = "E001/F014/S002"
	cuerpo := func(linea, v string) string {
		return linea + "veredicto: " + v + "\nhallazgos: —\nno_revisado: —"
	}
	nuevo := func() string {
		raiz := t.TempDir()
		escribir(t, raiz, "p/p.go", "package p\n")
		fichaSlice(t, raiz, id, "EN_CURSO")
		return raiz
	}
	t.Run("auditor con id anota fecha, veredicto y huella; gana el último", func(t *testing.T) {
		raiz := nuevo()
		if s := aquas(t, raiz, cierreAuditor(t, "auditor", cuerpo("id: E001-F014-S002-A1\n", "CON_HALLAZGOS"), false), "cierre-agente"); s.codigo != 0 {
			t.Fatalf("cierre válido: exit 0; dio %+v", s)
		}
		a := auditoriaDe(t, raiz, id)
		huella, _ := a["huella"].(map[string]any)
		if a == nil || a["veredicto"] != "CON_HALLAZGOS" || a["fecha"] == "" || huella["p/p.go"] == nil {
			t.Fatalf("ficha: se esperaba CON_HALLAZGOS con fecha y huella de p/p.go; dio %v", a)
		}
		if s := aquas(t, raiz, cierreAuditor(t, "auditor", cuerpo("id: E001-F014-S002-A2\n", "APROBADO"), false), "cierre-agente"); s.codigo != 0 {
			t.Fatalf("segundo cierre: exit 0; dio %+v", s)
		}
		if a := auditoriaDe(t, raiz, id); a["veredicto"] != "APROBADO" {
			t.Errorf("gana el último veredicto: se esperaba APROBADO; dio %v", a)
		}
	})
	t.Run("cierre de worker no anota", func(t *testing.T) {
		raiz := nuevo()
		w := "estado: HECHO\narchivos: —\nevidencia: —\nbloqueo: —\nfuera_de_alcance: —"
		if s := aquas(t, raiz, cierreAuditor(t, "worker-sonnet", w, false), "cierre-agente"); s.codigo != 0 {
			t.Fatalf("cierre de worker: exit 0; dio %+v", s)
		}
		if a := auditoriaDe(t, raiz, id); a != nil {
			t.Errorf("un worker no anota veredicto; dio %v", a)
		}
	})
	t.Run("cierre de auditor sin id se rechaza y no anota", func(t *testing.T) {
		raiz := nuevo()
		s := aquas(t, raiz, cierreAuditor(t, "auditor", cuerpo("", "APROBADO"), false), "cierre-agente")
		if s.codigo != 2 || !strings.Contains(s.err, "cierre inválido") || !strings.Contains(s.err, "id:") {
			t.Errorf("sin id: exit 2, «cierre inválido» y pide la línea id:; dio %+v", s)
		}
		if a := auditoriaDe(t, raiz, id); a != nil {
			t.Errorf("sin id no se anota; dio %v", a)
		}
	})
	t.Run("id de un tramo que no está en curso se rechaza nombrando el id", func(t *testing.T) {
		raiz := nuevo()
		fichaSlice(t, raiz, "E001/F014/S003", "BORRADOR")
		for _, otro := range []string{"E001-F014-S003-A1", "E001-F014-S099-A1"} {
			s := aquas(t, raiz, cierreAuditor(t, "auditor", cuerpo("id: "+otro+"\n", "APROBADO"), false), "cierre-agente")
			want := strings.ReplaceAll(strings.TrimSuffix(otro, "-A1"), "-", "/")
			if s.codigo != 2 || !strings.Contains(s.err, want) {
				t.Errorf("id %s: exit 2 nombrando %s; dio %+v", otro, want, s)
			}
		}
		if a := auditoriaDe(t, raiz, id); a != nil {
			t.Errorf("el tramo en curso no recibe veredictos de otro; dio %v", a)
		}
	})
	t.Run("stop_hook_active con cierre inválido deja cerrar sin anotar", func(t *testing.T) {
		raiz := nuevo()
		if s := aquas(t, raiz, cierreAuditor(t, "auditor", cuerpo("", "APROBADO"), true), "cierre-agente"); s.codigo != 0 {
			t.Fatalf("ya pedido una vez: exit 0; dio %+v", s)
		}
		if a := auditoriaDe(t, raiz, id); a != nil {
			t.Errorf("cierre inválido no se anota; dio %v", a)
		}
	})
	t.Run("stop_hook_active y anotar falla: avisa en stderr y deja cerrar (H1)", func(t *testing.T) {
		raiz := nuevo()
		escribir(t, raiz, "trabajo/activo/E001-F014-S002.json", "{roto")
		s := aquas(t, raiz, cierreAuditor(t, "auditor", cuerpo("id: E001-F014-S002-A1\n", "APROBADO"), true), "cierre-agente")
		if s.codigo != 0 || !strings.Contains(s.err, "no quedó anotado") || !strings.Contains(s.err, id) {
			t.Errorf("activo + ficha rota: exit 0 con aviso nombrando %s; dio %+v", id, s)
		}
	})
	t.Run("id: — es válido y no anota (H2)", func(t *testing.T) {
		raiz := nuevo()
		s := aquas(t, raiz, cierreAuditor(t, "auditor", cuerpo("id: —\n", "APROBADO"), false), "cierre-agente")
		if s.codigo != 0 || s.err != "" {
			t.Errorf("id: —: exit 0 y sin aviso; dio %+v", s)
		}
		if a := auditoriaDe(t, raiz, id); a != nil {
			t.Errorf("id: — no anota; dio %v", a)
		}
	})
	t.Run("auditor con mensaje vacío o ausente no pasa en silencio (D000053)", func(t *testing.T) {
		raiz := nuevo()
		for _, entrada := range []string{cierreAuditor(t, "auditor", "", false), `{"agent_type":"auditor"}`} {
			if s := aquas(t, raiz, entrada, "cierre-agente"); s.codigo != 2 || !strings.Contains(s.err, "cierre vacío") {
				t.Errorf("entrada %s: exit 2 «cierre vacío»; dio %+v", entrada, s)
			}
		}
		if s := aquas(t, raiz, cierreAuditor(t, "auditor", "", true), "cierre-agente"); s.codigo != 0 {
			t.Errorf("vacío con stop_hook_active: exit 0; dio %+v", s)
		}
	})
	t.Run("ficha ilegible: el gancho lo dice y no da el veredicto por anotado", func(t *testing.T) {
		raiz := nuevo()
		escribir(t, raiz, "trabajo/activo/E001-F014-S002.json", "{roto")
		s := aquas(t, raiz, cierreAuditor(t, "auditor", cuerpo("id: E001-F014-S002-A1\n", "APROBADO"), false), "cierre-agente")
		if s.codigo != 2 || !strings.Contains(s.err, id) {
			t.Errorf("ficha rota: exit 2 nombrando %s; dio %+v", id, s)
		}
	})
}

// aprobarGancho: el auditor cierra con APROBADO por el gancho real, como en el flujo de verdad.
func aprobarGancho(t *testing.T, raiz, id string) {
	t.Helper()
	cuerpo := "id: " + strings.ReplaceAll(id, "/", "-") + "-A1\nveredicto: APROBADO\nhallazgos: —\nno_revisado: —"
	if s := aquas(t, raiz, cierreAuditor(t, "auditor", cuerpo, false), "cierre-agente"); s.codigo != 0 {
		t.Fatalf("auditor APROBADO de %s: %+v", id, s)
	}
}
