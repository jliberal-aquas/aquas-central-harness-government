package compilador

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestOrquestadorGateHumanoPorFeature (E001/F015/S003): el orquestador compilado trae las reglas del gate por feature.
// Rojo cuando: falta push por slice, arreglar y reportar o los cuatro casos de volver al humano.
// Verde cuando: el orquestador compilado nombra las tres reglas.
func TestOrquestadorGateHumanoPorFeature(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", ".claude", "agents", "orchestrator.md"))
	if err != nil {
		t.Fatal(err)
	}
	texto := string(b)
	for _, want := range []string{
		"sin apruebo propio ni «dale» por paso", "«tope N» del apruebo",
		"push de cada slice al cerrarlo",
		"arreglar y reportar",
		"duda real", "cambio de alcance", "acción irreversible o externa", "feature terminada",
	} {
		if !strings.Contains(texto, want) {
			t.Errorf("orquestador compilado sin %q", want)
		}
	}
}
