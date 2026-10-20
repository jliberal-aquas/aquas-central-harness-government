package compilador

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoMinimo crea un gobierno de prueba con un documento y un agente.
func repoMinimo(t *testing.T) string {
	t.Helper()
	raiz := t.TempDir()
	archivos := map[string]string{
		RutaDocumentos: `{"documentos":[{"salida":"CLAUDE.md","fuentes":["gobierno/canonico/constitucion.md"]}]}`,
		RutaAgentes: `{"agentes":[{"name":"orchestrator","description":"Orquesta: <todo> & más","model":"claude-opus-5-5","effort":"high",
			"cuerpo":"gobierno/canonico/agentes/orq.md","reglas":["gobierno/canonico/reglas/r1.md"],
			"comunicacion":["gobierno/canonico/comunicacion/c1.md"]}]}`,
		"gobierno/canonico/constitucion.md":    "# Constitución\r\n- regla A\r\n",
		"gobierno/canonico/agentes/orq.md":     "# Orquestador\n",
		"gobierno/canonico/reglas/r1.md":       "# Regla 1\n",
		"gobierno/canonico/comunicacion/c1.md": "# Comunicación 1\n- regla original\n",
	}
	for rel, contenido := range archivos {
		if err := escribir(raiz, rel, []byte(contenido)); err != nil {
			t.Fatal(err)
		}
	}
	return raiz
}

func leer(t *testing.T, raiz, rel string) string {
	t.Helper()
	datos, err := os.ReadFile(filepath.Join(raiz, rel))
	if err != nil {
		t.Fatal(err)
	}
	return string(datos)
}

func verificarSinProblemas(t *testing.T, raiz string) {
	t.Helper()
	problemas, err := Verificar(raiz)
	if err != nil {
		t.Fatal(err)
	}
	if len(problemas) > 0 {
		t.Fatalf("se esperaba gate OK, problemas: %v", problemas)
	}
}

func TestCompilarGeneraAgenteYDocumento(t *testing.T) {
	raiz := repoMinimo(t)
	if _, err := Compilar(raiz); err != nil {
		t.Fatal(err)
	}
	agente := leer(t, raiz, ".claude/agents/orchestrator.md")
	for _, esperado := range []string{
		"---\nname: orchestrator\n",
		`description: "Orquesta: <todo> & más"`,
		"model: claude-opus-5-5\n",
		"effort: high\n",
		"GENERADO",
		"# Orquestador\n\n# Regla 1\n\n# Comunicación 1",
	} {
		if !strings.Contains(agente, esperado) {
			t.Errorf("agente no contiene %q:\n%s", esperado, agente)
		}
	}
	if strings.Contains(leer(t, raiz, "CLAUDE.md"), "\r") {
		t.Error("CLAUDE.md conserva CRLF; debe salir en LF")
	}
	verificarSinProblemas(t, raiz)
}

// Caso que debe fallar y luego pasar: cambio en canónico de comunicación.
func TestCambioEnCanonicoRecompilaAgente(t *testing.T) {
	raiz := repoMinimo(t)
	if _, err := Compilar(raiz); err != nil {
		t.Fatal(err)
	}
	c1 := "gobierno/canonico/comunicacion/c1.md"
	if err := escribir(raiz, c1, []byte("# Comunicación 1\n- regla original\n- regla NUEVA\n")); err != nil {
		t.Fatal(err)
	}

	problemas, _ := Verificar(raiz)
	if len(problemas) == 0 {
		t.Fatal("gate debía fallar tras cambiar un canónico sin recompilar")
	}

	inf, err := Compilar(raiz)
	if err != nil {
		t.Fatal(err)
	}
	if len(inf.FuentesCambiadas) != 1 || inf.FuentesCambiadas[0] != c1 {
		t.Errorf("fuentes cambiadas = %v, se esperaba solo %s", inf.FuentesCambiadas, c1)
	}
	if !strings.Contains(leer(t, raiz, ".claude/agents/orchestrator.md"), "regla NUEVA") {
		t.Error("el agente no incluye la regla nueva")
	}
	verificarSinProblemas(t, raiz)
}

func TestEdicionManualDeDerivadoFallaGate(t *testing.T) {
	raiz := repoMinimo(t)
	if _, err := Compilar(raiz); err != nil {
		t.Fatal(err)
	}
	if err := escribir(raiz, "CLAUDE.md", []byte("editado a mano\n")); err != nil {
		t.Fatal(err)
	}
	problemas, _ := Verificar(raiz)
	if len(problemas) == 0 {
		t.Fatal("gate debía detectar la edición manual de CLAUDE.md")
	}
}

func TestFuenteFueraDeCanonicoSeRechaza(t *testing.T) {
	raiz := repoMinimo(t)
	escribir(raiz, RutaDocumentos, []byte(`{"documentos":[{"salida":"CLAUDE.md","fuentes":["README.md"]}]}`))
	escribir(raiz, "README.md", []byte("x"))
	if _, err := Compilar(raiz); err == nil {
		t.Fatal("se esperaba error: fuente fuera de gobierno/canonico/")
	}
}

func TestCompilarEsIdempotente(t *testing.T) {
	raiz := repoMinimo(t)
	Compilar(raiz)
	inf, err := Compilar(raiz)
	if err != nil {
		t.Fatal(err)
	}
	if len(inf.Escritos) != 0 {
		t.Errorf("segunda compilación reescribió %v; debía ser no-op", inf.Escritos)
	}
}
