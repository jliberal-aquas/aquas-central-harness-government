package encargo

import (
	"strings"
	"testing"
)

// E001/F006/S002: el hook SubagentStop rechaza un cierre sin sus campos fijos.

const cierreWorker = "estado: HECHO\narchivos: internal/x.go\nevidencia: go test ./... → ok\nbloqueo: —\nfuera_de_alcance: —\n"
const cierreAuditor = "veredicto: CON_HALLAZGOS\nhallazgos: falta prueba · x_test.go:3 · lectura · cobertura · confirmado\nno_revisado: —\n"

func TestCierreValidoPasa(t *testing.T) {
	for _, c := range []struct{ tipo, msg string }{
		{"worker-haiku", cierreWorker},
		{"worker-sonnet", strings.Replace(cierreWorker, "HECHO", "BLOQUEADO", 1)},
		{"auditor", cierreAuditor},
		{"auditor", "\n" + strings.Replace(cierreAuditor, "CON_HALLAZGOS", "APROBADO", 1)},
		{"auditor", "```\n" + cierreAuditor + "```\n"},
		{"worker-haiku", "estado: PARCIAL\n  archivos: a.go\nevidencia:\n- go test → ok\n- verificar → GATE OK\nbloqueo: —\nfuera_de_alcance: —\r\n"},
	} {
		if m := Revisar(c.tipo, c.msg, false); m != "" {
			t.Errorf("%s: cierre válido debe pasar; dio %q", c.tipo, m)
		}
	}
}

func TestCierreInvalidoBloquea(t *testing.T) {
	casos := []struct{ nombre, tipo, msg, espera string }{
		{"prosa antes del cierre", "worker-haiku", "Listo, hice el cambio.\n" + cierreWorker, "primera línea"},
		{"estado inventado", "worker-sonnet", strings.Replace(cierreWorker, "HECHO", "TERMINADO", 1), "estado"},
		{"falta evidencia", "worker-haiku", strings.Replace(cierreWorker, "evidencia: go test ./... → ok\n", "", 1), "evidencia"},
		{"falta fuera_de_alcance", "worker-haiku", strings.Replace(cierreWorker, "fuera_de_alcance: —\n", "", 1), "fuera_de_alcance"},
		{"auditor con cierre de worker", "auditor", cierreWorker, "veredicto"},
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
