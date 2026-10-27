package encargo

import (
	"strings"
	"testing"
)

// E001/F006/S002: el hook SubagentStop rechaza un cierre sin sus campos fijos.

const cierreWorker = "estado: HECHO\narchivos: internal/x.go\nevidencia: go test ./... → ok\nbloqueo: —\nfuera_de_alcance: —\n"
const cierreAuditor = "id: E001-F014-S002-A1\nveredicto: CON_HALLAZGOS\nhallazgos: falta prueba · x_test.go:3 · lectura · cobertura · confirmado\nno_revisado: —\n"

func TestCierreValidoPasa(t *testing.T) {
	for _, c := range []struct{ tipo, msg string }{
		{"worker-haiku", cierreWorker},
		{"worker-sonnet", strings.Replace(cierreWorker, "HECHO", "BLOQUEADO", 1)},
		{"auditor", cierreAuditor},
		{"auditor", "\n" + strings.Replace(cierreAuditor, "CON_HALLAZGOS", "APROBADO", 1)},
		{"auditor", "```\n" + cierreAuditor + "```\n"},
		{"auditor", strings.Replace(cierreAuditor, "id: E001-F014-S002-A1", "id: —", 1)}, // H2: auditoría sin tramo
		{"worker-haiku", "estado: PARCIAL\n  archivos: a.go\nevidencia:\n- go test → ok\n- verificar → GATE OK\nbloqueo: —\nfuera_de_alcance: —\r\n"},
	} {
		if m := Revisar(c.tipo, c.msg, false); m != "" {
			t.Errorf("%s: cierre válido debe pasar; dio %q", c.tipo, m)
		}
	}
	// E001/F014/S002: del cierre de auditor válido se extrae el id del tramo y el veredicto.
	id, v, ok := Veredicto("```\n" + strings.Replace(cierreAuditor, "S002-A1", "S002-A12", 1) + "```\n")
	if !ok || id != "E001/F014/S002" || v != "CON_HALLAZGOS" {
		t.Errorf("se esperaba E001/F014/S002 y CON_HALLAZGOS; dio %q %q %v", id, v, ok)
	}
	if _, _, ok := Veredicto(strings.Replace(cierreAuditor, "id: E001-F014-S002-A1", "id: —", 1)); ok {
		t.Errorf("id: — no trae tramo que anotar")
	}
	if _, _, ok := Veredicto(cierreWorker); ok {
		t.Errorf("un cierre de worker no trae veredicto")
	}
}

func TestCierreInvalidoBloquea(t *testing.T) {
	casos := []struct{ nombre, tipo, msg, espera string }{
		{"prosa antes del cierre", "worker-haiku", "Listo, hice el cambio.\n" + cierreWorker, "primera línea"},
		{"estado inventado", "worker-sonnet", strings.Replace(cierreWorker, "HECHO", "TERMINADO", 1), "estado"},
		{"falta evidencia", "worker-haiku", strings.Replace(cierreWorker, "evidencia: go test ./... → ok\n", "", 1), "evidencia"},
		{"falta fuera_de_alcance", "worker-haiku", strings.Replace(cierreWorker, "fuera_de_alcance: —\n", "", 1), "fuera_de_alcance"},
		{"auditor con cierre de worker", "auditor", cierreWorker, "id:"},
		{"auditor sin id", "auditor", strings.Replace(cierreAuditor, "id: E001-F014-S002-A1\n", "", 1), "id:"},
		{"auditor sin id: el ejemplo del mensaje es genérico", "auditor", cierreWorker, "E000-F000-S000-A1"},
		{"auditor con mensaje vacío", "auditor", "", "cierre vacío"},
		{"auditor con id mal formado", "auditor", strings.Replace(cierreAuditor, "S002-A1", "S002", 1), "id:"},
		{"auditor con id de feature", "auditor", strings.Replace(cierreAuditor, "-S002-A1", "-A1", 1), "id:"},
		{"auditor con veredicto inventado", "auditor", strings.Replace(cierreAuditor, "CON_HALLAZGOS", "TAL_VEZ", 1), "segunda línea"},
		{"falta no_revisado", "auditor", strings.Replace(cierreAuditor, "no_revisado: —\n", "", 1), "no_revisado"},
		{"vacío", "worker-sonnet", "", "primera línea"},
		{"texto después del cierre", "worker-haiku", cierreWorker + "\nGracias, cualquier duda me dices.\n", "texto fuera del cierre"},
		{"campo vacío", "worker-haiku", strings.Replace(cierreWorker, "evidencia: go test ./... → ok", "evidencia:", 1), "vacíos: evidencia"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			m := Revisar(c.tipo, c.msg, false)
			if !strings.Contains(m, c.espera) || !strings.Contains(m, "02_COMUNICACION_ENTRE_AGENTES") {
				t.Fatalf("debe bloquear nombrando %q y el canónico; dio %q", c.espera, m)
			}
		})
	}
}

func TestCierreNoBloqueaEnBucleNiAOtros(t *testing.T) {
	if m := Revisar("worker-haiku", "prosa libre", true); m != "" {
		t.Errorf("si ya se pidió una vez (stop_hook_active), deja cerrar; dio %q", m)
	}
	if m := Revisar("Explore", "prosa libre", false); m != "" {
		t.Errorf("agentes ajenos al contrato no se revisan; dio %q", m)
	}
}
