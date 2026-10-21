package turno

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// E001/F002/S001: auditoría de turnos. El hook UserPromptSubmit abre el turno con el mensaje
// literal del humano; el hook Stop guarda la respuesta del agente. Nadie más escribe ahí.

var octubre = time.Date(2026, 10, 21, 18, 0, 0, 0, time.UTC)

func leer(t *testing.T, raiz, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(raiz, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func abrir(t *testing.T, raiz, sesion, literal string, ahora time.Time) string {
	t.Helper()
	rel, err := Abrir(raiz, sesion, literal, ahora)
	if err != nil {
		t.Fatal(err)
	}
	return rel
}

func TestAbrirGuardaLiteralIntacto(t *testing.T) {
	raiz := t.TempDir()
	literal := "Arregla esto:\n```go\nx := 1\n```\n## Respuesta\nfalsa, es parte del mensaje\n  con sangría  "
	rel := abrir(t, raiz, "s1", literal, octubre)
	if rel != "auditoria/turnos/2026-10/T000001.md" {
		t.Fatalf("primer turno de octubre en auditoria/turnos/2026-10/T000001.md; dio %s", rel)
	}
	c := leer(t, raiz, rel)
	if !strings.Contains(c, "## Literal\n\n````text\n"+literal+"\n````\n") {
		t.Errorf("el literal va intacto en una cerca más larga que sus backticks:\n%s", c)
	}
	if !strings.HasSuffix(c, "\n````\n\n## Respuesta\n\n—\n") {
		t.Errorf("recién abierto, la sección real de respuesta va tras la cerca y queda pendiente («—»):\n%s", c)
	}
	if !strings.HasPrefix(c, "# T000001\n") || !strings.Contains(c, "- Sesión: `s1`\n") || !strings.Contains(c, "- Abierto: 2026-10-21T18:00:00Z\n") {
		t.Errorf("cabecera con id, sesión y hora de apertura:\n%s", c)
	}
}

func TestNumeracionSinReutilizar(t *testing.T) {
	raiz := t.TempDir()
	quiere := []string{
		"auditoria/turnos/2026-10/T000001.md",
		"auditoria/turnos/2026-10/T000002.md",
		"auditoria/turnos/2026-11/T000003.md",
	}
	fechas := []time.Time{octubre, octubre.Add(time.Minute), octubre.AddDate(0, 0, 11)}
	for i, f := range fechas {
		if rel := abrir(t, raiz, "s1", "hola", f); rel != quiere[i] {
			t.Errorf("turno %d: quiero %s; dio %s", i+1, quiere[i], rel)
		}
	}
	// Un hueco no se rellena: el siguiente es el mayor + 1, aunque se haya borrado uno.
	os.Remove(filepath.Join(raiz, "auditoria", "turnos", "2026-10", "T000002.md"))
	if rel := abrir(t, raiz, "s2", "otra", fechas[2]); rel != "auditoria/turnos/2026-11/T000004.md" {
		t.Errorf("tras borrar T000002 sigue T000004; dio %s", rel)
	}
}

func TestCerrarGuardaRespuesta(t *testing.T) {
	raiz := t.TempDir()
	rel := abrir(t, raiz, "s1", "pedido", octubre)
	otro := abrir(t, raiz, "s2", "pedido de otra sesión", octubre)
	got, err := Cerrar(raiz, "s1", "primera respuesta", octubre.Add(time.Minute))
	if err != nil || got != rel {
		t.Fatalf("cerrar s1 escribe su turno %s; dio %q, %v", rel, got, err)
	}
	c := leer(t, raiz, rel)
	if !strings.Contains(c, "## Respuesta\n\n```text\nprimera respuesta\n```\n") || !strings.Contains(c, "- Cerrado: 2026-10-21T18:01:00Z\n") {
		t.Errorf("respuesta y hora de cierre en el turno:\n%s", c)
	}
	if !strings.Contains(c, "````text\npedido\n````") && !strings.Contains(c, "```text\npedido\n```") {
		t.Errorf("el cierre conserva el literal:\n%s", c)
	}
	// El hook Stop se repite si otro hook frena el cierre: gana la última respuesta, sin duplicar.
	if _, err := Cerrar(raiz, "s1", "respuesta final", octubre.Add(2*time.Minute)); err != nil {
		t.Fatal(err)
	}
	c = leer(t, raiz, rel)
	if strings.Contains(c, "primera respuesta") || !strings.Contains(c, "respuesta final") || strings.Count(c, "## Respuesta") != 1 {
		t.Errorf("segundo cierre reemplaza la respuesta:\n%s", c)
	}
	if o := leer(t, raiz, otro); !strings.Contains(o, "## Respuesta\n\n—\n") {
		t.Errorf("cerrar s1 no toca el turno de s2:\n%s", o)
	}
}

func TestCerrarSinTurnoAbierto(t *testing.T) {
	raiz := t.TempDir()
	rel, err := Cerrar(raiz, "s1", "respuesta", octubre)
	if err != nil || rel != "" {
		t.Fatalf("sin turno abierto: no escribe ni falla; dio %q, %v", rel, err)
	}
	if _, err := os.Stat(filepath.Join(raiz, "auditoria")); !os.IsNotExist(err) {
		t.Errorf("sin turno abierto no crea auditoria/")
	}
	abrir(t, raiz, "s1", "pedido", octubre)
	if rel, err := Cerrar(raiz, "otra", "respuesta", octubre); err != nil || rel != "" {
		t.Errorf("sesión sin turno propio: no escribe en el de otra; dio %q, %v", rel, err)
	}
}
