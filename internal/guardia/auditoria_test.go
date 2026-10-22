package guardia

import (
	"strings"
	"testing"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/turno"
)

// E001/F002/S001: la auditoría solo la escribe `aquas turno` desde los hooks. La carpeta sale
// del paquete turno (lección de E000/F005/S005): si se muda, la guardia protege la nueva.
func TestGuardiaProtegeAuditoria(t *testing.T) {
	raiz := raizPrueba(t)
	p := turno.Carpeta + "/2026-10/T000001.md"
	if m := RevisarArchivo(raiz, p); !strings.Contains(m, "auditoría: "+p) {
		t.Errorf("Write a %s: se esperaba bloqueo de auditoría; motivo %q", p, m)
	}
	bash := `{"tool_name":"Bash","tool_input":{"command":"echo x >> ` + p + `"}}`
	if m := Revisar(raiz, []byte(bash)); !strings.Contains(m, "auditoría: ") {
		t.Errorf("Bash hacia %s: se esperaba bloqueo de auditoría; motivo %q", p, m)
	}
	if m := RevisarArchivo(raiz, "docs/auditoria.md"); m != "" {
		t.Errorf("un nombre parecido fuera de la carpeta no se bloquea; motivo %q", m)
	}
}
