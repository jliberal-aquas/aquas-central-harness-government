package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestVerifyExigeAislamiento (E000/F008/S003, A1 hallazgos 2 y 4): en Linux sin ajustes gestionados estrictos,
// `aquas verify` falla en la central y en un hijo, y el arranque lo avisa; con ellos, nada se dice.
func TestVerifyExigeAislamiento(t *testing.T) {
	gestion := t.TempDir()
	t.Setenv("AQUAS_PRUEBA_GOOS", "linux")
	t.Setenv("AQUAS_PRUEBA_GESTION", gestion)
	central := t.TempDir()
	hijo := t.TempDir()
	escribir(t, hijo, ".aquas/lock.json", "{}")
	// Ajustes estrictos de usuario y de repo no cuentan: el agente los edita (A2, hallazgo 1).
	const estricto = `{"sandbox":{"enabled":true,"allowUnsandboxedCommands":false,"failIfUnavailable":true}}`
	usuario := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", usuario)
	escribir(t, usuario, "settings.json", estricto)
	for _, raiz := range []string{central, hijo} {
		escribir(t, raiz, ".claude/settings.json", estricto)
		escribir(t, raiz, ".claude/settings.local.json", estricto)
	}
	const marca = "aislamiento de Bash no estricto"
	for _, raiz := range []string{central, hijo} {
		if s := aquas(t, raiz, "", "verificar"); s.codigo == 0 || !strings.Contains(s.out+s.err, marca) {
			t.Errorf("%s: verify falla nombrando el aislamiento; dio %+v", raiz, s)
		}
	}
	if s := aquas(t, central, `{"source":"startup"}`, "memoria", "inicio"); !strings.Contains(s.out, "AVISO "+marca) {
		t.Errorf("el arranque avisa en Linux sin aislamiento estricto; dio %+v", s)
	}

	if err := os.WriteFile(filepath.Join(gestion, "managed-settings.json"), []byte(`{"sandbox":{"enabled":true,"allowUnsandboxedCommands":false,"failIfUnavailable":true}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, raiz := range []string{central, hijo} {
		if s := aquas(t, raiz, "", "verificar"); strings.Contains(s.out+s.err, "aislamiento") {
			t.Errorf("%s: con ajustes gestionados estrictos verify no nombra el aislamiento; dio %+v", raiz, s)
		}
	}
	if s := aquas(t, central, `{"source":"startup"}`, "memoria", "inicio"); strings.Contains(s.out, "aislamiento") {
		t.Errorf("con ajustes gestionados estrictos el arranque no avisa; dio %+v", s)
	}
}
