package compilador

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// E001/F010/S001: la definición de terminado común es un canónico que llega al orquestador compilado.

// Rojo cuando: el orquestador compilado no trae la regla de terminado común.
// Verde cuando: el orquestador la enlaza en su config y el bloque de la regla en el compilado trae el título con la versión y sus requisitos.
func TestDoDComunCompilado(t *testing.T) {
	const regla = "gobierno/canonico/reglas/03_DOD_COMUN.md"
	b, err := os.ReadFile(filepath.Join("..", "..", filepath.FromSlash(RutaAgentes)))
	if err != nil {
		t.Fatal(err)
	}
	var c struct{ Agentes []Agente }
	if err := json.Unmarshal(b, &c); err != nil {
		t.Fatal(err)
	}
	enlazada := false
	for _, a := range c.Agentes {
		if a.Name == "orchestrator" {
			enlazada = slices.Contains(a.Reglas, regla)
		}
	}
	if !enlazada {
		t.Errorf("agentes.json: el orquestador debe enlazar %s", regla)
	}
	comp, err := os.ReadFile(filepath.Join("..", "..", ".claude", "agents", "orchestrator.md"))
	if err != nil {
		t.Fatal(err)
	}
	const titulo = "# Definición de terminado común · v0.1"
	texto := string(comp)
	ini := strings.Index(texto, titulo)
	if ini < 0 {
		t.Fatalf("orquestador compilado sin %q", titulo)
	}
	bloque := texto[ini:]
	if fin := strings.Index(bloque[len(titulo):], "\n# "); fin >= 0 { // el bloque llega hasta el siguiente título de primer nivel
		bloque = bloque[:len(titulo)+fin]
	}
	for _, want := range []string{"prueba integral", "Gate (`aquas verify`)", "APROBADO", "Commit con el porqué"} {
		if !strings.Contains(bloque, want) {
			t.Errorf("el bloque de la regla DoD en el orquestador compilado no trae %q", want)
		}
	}
}
