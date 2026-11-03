package turno

import (
	"strings"
	"testing"
)

// TestTurnoGuardaRefinado (E001/F017/S004): la respuesta a un turno humano declara tipo y pedido refinado; el turno
// los guarda y su falta frena el cierre.
// Rojo cuando: la respuesta no trae refinado y el cierre pasa.
// Verde cuando: el refinado queda en el turno y su falta frena.
func TestTurnoGuardaRefinado(t *testing.T) {
	desafio := "DESAFÍO\n- Supuestos: x\n- Contraargumentos: sin objeción real\n- Riesgos: y\n- Recomendación: hacerlo\n"
	casos := []struct {
		caso, literal, respuesta string
		frena                    bool
		tipo, refinado           string // lo que queda en el turno
	}{
		{"humano con las dos líneas", "dale, sigue", "Tipo: confirmación\nPedido refinado: seguir con S004 dentro del apruebo.\n\nINFORMACIÓN: hecho.", false, "confirmación", "seguir con S004 dentro del apruebo."},
		{"humano sin las líneas", "arregla el gate", "INFORMACIÓN: hecho.", true, "—", "—"},
		{"agente sin las líneas", "<agent-message from=\"a1\">informe</agent-message>", "INFORMACIÓN: recibido.", false, "—", "—"},
		{"tipo fuera de la lista", "haz X", "Tipo: urgente\nPedido refinado: hacer X.", true, "—", "hacer X."},
		{"refinado vacío", "haz X", "Tipo: petición\nPedido refinado:   \n", true, "petición", "—"},
		{"tipo en mayúsculas y negrita", "haz X", "**Tipo:** Petición.\n**Pedido refinado:** hacer X para Y.", false, "petición", "hacer X para Y."},
		{"en viñeta tras el DESAFÍO", "nuevo pedido largo", desafio + "- Tipo: decisión\n- Pedido refinado: elegir A sobre B.\n", false, "decisión", "elegir A sobre B."},
		{"duplicada: cuenta la primera", "haz X", "Tipo: petición\nPedido refinado: hacer X.\nTipo: pregunta\nPedido refinado: otra cosa.", false, "petición", "hacer X."},
		{"en una cita", "haz X", "> Tipo: corrección\n> Pedido refinado: arreglar X.", false, "corrección", "arreglar X."},
		{"solo en un bloque de código", "haz X", "```\nTipo: petición\nPedido refinado: hacer X.\n```\n", true, "—", "—"},
	}
	for i, c := range casos {
		raiz := t.TempDir()
		sesion := "s" + string(rune('a'+i))
		rel := abrir(t, raiz, sesion, c.literal, octubre)
		if s := leer(t, raiz, rel); !strings.Contains(s, "- Tipo: —\n- Pedido refinado: —\n") {
			t.Errorf("%s: recién abierto, tipo y refinado pendientes («—»):\n%s", c.caso, s)
		}
		if _, err := Cerrar(raiz, sesion, c.respuesta, octubre); err != nil {
			t.Fatal(err)
		}
		if m := RevisarRefinado(raiz, sesion, c.respuesta); (m != "") != c.frena {
			t.Errorf("%s: frena=%v; dio %q", c.caso, c.frena, m)
		}
		s := leer(t, raiz, rel)
		if !strings.Contains(s, "- Tipo: "+c.tipo+"\n- Pedido refinado: "+c.refinado+"\n") {
			t.Errorf("%s: el turno guarda «Tipo: %s» y «Pedido refinado: %s»:\n%s", c.caso, c.tipo, c.refinado, s)
		}
	}

	// Fallo a mitad de camino: otro freno repite el cierre con una corrección corta; el refinado ya dicho se conserva.
	raiz := t.TempDir()
	rel := abrir(t, raiz, "r", "haz X", octubre)
	if _, err := Cerrar(raiz, "r", "Tipo: petición\nPedido refinado: hacer X.\n\nTERMINADO: X.", octubre); err != nil {
		t.Fatal(err)
	}
	if _, err := Cerrar(raiz, "r", "Corregido: agregué el defecto D000001.", octubre); err != nil {
		t.Fatal(err)
	}
	if m := RevisarRefinado(raiz, "r", "Corregido: agregué el defecto D000001."); m != "" {
		t.Errorf("la corrección tras otro freno no pide de nuevo el refinado: %q", m)
	}
	if s := leer(t, raiz, rel); !strings.Contains(s, "- Tipo: petición\n- Pedido refinado: hacer X.\n") || !strings.Contains(s, "Corregido: agregué") {
		t.Errorf("tras repetir el cierre, el turno guarda la última respuesta y conserva el refinado:\n%s", s)
	}

	// Sin turno abierto (sesión desconocida) no hay a qué exigirle refinado.
	if m := RevisarRefinado(t.TempDir(), "nadie", "hola"); m != "" {
		t.Errorf("sin turno abierto no frena: %q", m)
	}
}
