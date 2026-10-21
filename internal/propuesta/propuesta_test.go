package propuesta

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// E001/F002/S006: una propuesta de proceso o diseño en el chat lleva diagrama en texto.
// El hook Stop lee la respuesta completa (last_assistant_message) y frena si una línea
// «PROPUESTA: …» no tiene después un bloque de código con flechas o cajas.

const cerca = "```"

func TestPropuestaSinDiagramaFrena(t *testing.T) {
	msg := "DECISIÓN: alcance.\n\nPROPUESTA: el sello arma la lista y frena el cierre.\n\nSin dibujo.\n"
	if m := Revisar(msg); !strings.Contains(m, "PROPUESTA sin diagrama") || !strings.Contains(m, "el sello arma la lista") {
		t.Errorf("propuesta sin diagrama: frena y nombra la propuesta; dio %q", m)
	}
	// Con negritas o viñeta delante también es propuesta.
	if m := Revisar("**PROPUESTA:** otra cosa\n"); m == "" {
		t.Errorf("«**PROPUESTA:**» también cuenta como propuesta")
	}
	// Cada propuesta necesita su diagrama: el de la primera no tapa a la segunda.
	dos := "PROPUESTA: uno\n" + cerca + "text\nA -> B\n" + cerca + "\nPROPUESTA: dos\nsin dibujo\n"
	if m := Revisar(dos); !strings.Contains(m, "dos") {
		t.Errorf("la segunda propuesta sin diagrama frena; dio %q", m)
	}
}

func TestPropuestaConDiagramaPasa(t *testing.T) {
	for _, d := range []string{"cerrar ──► CERRADO", "A -> B", "A → B", "┌───┐\n│ x │\n└───┘"} {
		msg := "PROPUESTA: el flujo.\n\n" + cerca + "text\n" + d + "\n" + cerca + "\n"
		if m := Revisar(msg); m != "" {
			t.Errorf("con diagrama %q pasa; dio %q", d, m)
		}
	}
}

func TestBloqueSinFlechasNoEsDiagrama(t *testing.T) {
	msg := "PROPUESTA: correr las pruebas.\n\n" + cerca + "bash\ngo test ./...\n" + cerca + "\n"
	if m := Revisar(msg); !strings.Contains(m, "PROPUESTA sin diagrama") {
		t.Errorf("un bloque sin flechas ni cajas no es diagrama; dio %q", m)
	}
	// Una flecha fuera de un bloque de código tampoco cuenta.
	if m := Revisar("PROPUESTA: A -> B en prosa\n"); m == "" {
		t.Errorf("flecha en prosa no es diagrama")
	}
}

func TestSinMarcaPasa(t *testing.T) {
	for _, msg := range []string{
		"TERMINADO: listo.\n",
		"Recomiendo una línea que empiece con `PROPUESTA:`.\n",
		"Ejemplo:\n" + cerca + "text\nPROPUESTA: dentro de un bloque no cuenta\n" + cerca + "\n",
		"",
	} {
		if m := Revisar(msg); m != "" {
			t.Errorf("sin línea PROPUESTA: pasa; %q dio %q", msg, m)
		}
	}
}

// La plantilla real de hooks trae `aquas propuesta` en Stop.
func TestPlantillaTraeHookDePropuesta(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "gobierno", "config", "claude_settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	var p struct {
		Hooks []struct{ Evento, Orden string }
	}
	if err := json.Unmarshal(b, &p); err != nil {
		t.Fatal(err)
	}
	for _, h := range p.Hooks {
		if h.Evento == "Stop" && h.Orden == "propuesta" {
			return
		}
	}
	t.Errorf("falta {evento: Stop, orden: propuesta} en la plantilla; hay %+v", p.Hooks)
}
