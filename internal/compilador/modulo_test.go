package compilador

import (
	"os"
	"strings"
	"testing"
)

// Ruta pública del repo central en GitHub: `go install` exige que el módulo sea igual.
const moduloRemoto = "github.com/jliberal-aquas/aquas-central-harness-government"

// TestModuloIgualRemoto: go.mod y la constitución canónica nombran el módulo igual al remoto.
func TestModuloIgualRemoto(t *testing.T) {
	gomod, err := os.ReadFile("../../go.mod")
	if err != nil {
		t.Fatal(err)
	}
	primera, _, _ := strings.Cut(string(gomod), "\n")
	if got := strings.TrimSpace(strings.TrimPrefix(primera, "module")); got != moduloRemoto {
		t.Fatalf("go.mod: módulo %q, quiero %q", got, moduloRemoto)
	}
	con, err := os.ReadFile("../../gobierno/canonico/constitucion.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(con), "Módulo: `"+moduloRemoto+"`") {
		t.Fatalf("constitución canónica no nombra el módulo %q", moduloRemoto)
	}
}
