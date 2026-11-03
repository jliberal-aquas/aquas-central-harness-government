package turno

import (
	"strings"
	"testing"
)

// TestTurnoMarcaOrigen (E001/F017/S003): cada turno dice si vino del humano o de un agente.
// Rojo cuando: un informe de subagente queda como mensaje del humano.
// Verde cuando: el informe queda con origen agente y el humano con origen humano.
func TestTurnoMarcaOrigen(t *testing.T) {
	raiz := t.TempDir()
	casos := []struct{ caso, literal, origen string }{
		{"humano", "dale, sigue", "humano"},
		{"informe de subagente", "<agent-message from=\"a1\">\nid: x\n</agent-message>", "agente"},
		{"notificación de tarea", "<task-notification>\n<task-id>b1</task-id>\n</task-notification>", "agente"},
		{"espacios antes", "  \n<agent-message from=\"a2\">x</agent-message>", "agente"},
		{"humano que cita una etiqueta", "¿qué es <agent-message> en el registro?", "humano"},
		{"literal vacío", "", "humano"},
		{"cerco de cuatro comillas", "````\n<agent-message>\n````", "humano"},
	}
	for i, c := range casos {
		sesion := "s" + string(rune('a'+i))
		rel := abrir(t, raiz, sesion, c.literal, octubre)
		// Abierto, sin cierre: el origen ya está escrito (fallo a mitad de camino).
		if s := leer(t, raiz, rel); !strings.Contains(s, "- Origen: "+c.origen+"\n") {
			t.Errorf("%s: el turno abierto debe decir «- Origen: %s»; dio:\n%s", c.caso, c.origen, s)
		}
		if _, err := Cerrar(raiz, sesion, "respuesta", octubre); err != nil {
			t.Fatal(err)
		}
		if s := leer(t, raiz, rel); !strings.Contains(s, "- Origen: "+c.origen+"\n") {
			t.Errorf("%s: el turno cerrado conserva «- Origen: %s»; dio:\n%s", c.caso, c.origen, s)
		}
	}
	if got := Origen("<task-notification>x"); got != "agente" {
		t.Errorf("Origen exportado para otros controles (S004, S007): dio %q", got)
	}
}
