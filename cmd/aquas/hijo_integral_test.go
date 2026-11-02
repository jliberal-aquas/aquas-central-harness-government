package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestHijoInstaladoPasaSuGate: integral de E002/F001. Un hijo recién instalado en carpeta vacía
// arma su CLAUDE.md y pasa su propio gate sin errores.
// Rojo cuando: install, build (con los huecos llenos) o verificar con wiki mínima sale distinto de 0, o sin wiki pasa.
// Verde cuando: sin docs/ verificar falla, y con wiki mínima los pasos salen con 0 en la carpeta del hijo (D000098).
func TestHijoInstaladoPasaSuGate(t *testing.T) {
	base := t.TempDir()
	hijo := filepath.Join(base, "hijo")
	correr := func(paso ...string) {
		t.Helper()
		if s := aquas(t, base, "", paso...); s.codigo != 0 {
			t.Fatalf("aquas %v: exit %d; quiero 0\n%s%s", paso, s.codigo, s.out, s.err)
		}
	}
	// E002/F002/S001: install corre en la carpeta del hijo, sin ruta, y crea los huecos PENDIENTE.
	if err := os.MkdirAll(hijo, 0o755); err != nil {
		t.Fatal(err)
	}
	git(t, hijo, "init", "-q") // E003/F001/S005: el pre-commit (control «todos») necesita repo git propio

	if s := aquas(t, hijo, "", "install"); s.codigo != 0 {
		t.Fatalf("aquas install en el hijo: exit %d; quiero 0\n%s%s", s.codigo, s.out, s.err)
	}
	// los huecos son lo único que el hijo escribe a mano (E002/F001/S002)
	huecos := filepath.Join(hijo, "gobierno", "canonico", "huecos")
	for _, h := range []string{"titulo", "mision", "mapa"} {
		if err := os.WriteFile(filepath.Join(huecos, h+".md"), []byte("Hijo de prueba: "+h+".\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	correr("build", hijo)
	// E003/F001/S005 (D000098): la wiki es control «todos»; sin docs/ el gate del hijo falla, y con una wiki mínima pasa.
	if s := aquas(t, hijo, "", "verificar"); s.codigo == 0 || !strings.Contains(s.out+s.err, "falta docs/README.md") {
		t.Fatalf("aquas verificar en un hijo sin wiki debió fallar por docs/README.md; dio exit %d\n%s%s", s.codigo, s.out, s.err)
	}
	escribir(t, hijo, "docs/glosario.md", "# Glosario\n\nPalabras técnicas.\n\n## zeta\n\nUna letra.\n\nEjemplo: la zeta va al final.\n")
	escribir(t, hijo, "docs/README.md", "# Hijo\n\n## En una frase\n\nAlgo simple.\n\n## Para qué existe\n\nProblema → solución.\n\n"+
		"## Analogía\n\nComo un mapa.\n\n## Diagrama\n\n```mermaid\nflowchart LR\n    A --> B\n```\n\n"+
		"## Ejemplo de punta a punta\n\n<a id=\"EJ-INICIO\"></a>\n\nTexto del ejemplo.\n\n## Paso a paso\n\n1. Uno.\n\n"+
		"## Cómo se usa\n\nSe lee.\n\n## Navegación\n\n- Glosario: [Glosario](glosario.md)\n")
	if s := aquas(t, hijo, "", "verificar"); s.codigo != 0 {
		t.Fatalf("aquas verificar en el hijo: exit %d; quiero 0\n%s%s", s.codigo, s.out, s.err)
	}
}
