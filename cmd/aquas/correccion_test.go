package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/defecto"
)

// TestCorreccionExigeDefecto (E001/F017/S007): el cierre de un turno humano de tipo corrección nombra su defecto.
// Rojo cuando: una corrección del humano cierra con «¿Salió algo mal?: No» o sin un D000NNN registrado.
// Verde cuando: la corrección cierra nombrando un defecto que existe; los demás tipos y los turnos de agente no lo exigen.
func TestCorreccionExigeDefecto(t *testing.T) {
	raiz := t.TempDir()
	if err := os.MkdirAll(filepath.Join(raiz, "registro", "defectos"), 0o755); err != nil {
		t.Fatal(err)
	}
	d := `{"id":"D000001","firma":"f","que":"q","causa":"c","arreglo":"a","como_me_cercioro":"x","ocurrencias":[]}`
	if err := os.WriteFile(filepath.Join(raiz, "registro", "defectos", "D000001.json"), []byte(d), 0o644); err != nil {
		t.Fatal(err)
	}
	cinco := "1. Qué falló: x.\n2. Por qué falló: y.\n3. Cómo lo arreglo: z.\n4. Cómo evito que vuelva: w.\n5. Qué regla/gancho/agente cambia: v.\n"
	// Hallazgo A1.2: el gancho ya frena por RevisarFallo; aquí RevisarCorreccion decide sola.
	for _, c := range []struct {
		caso, msg string
		frena     bool
	}{
		{"defecto registrado", "**¿Salió algo mal?:** D000001", false},
		{"defecto inexistente", "**¿Salió algo mal?:** D999999", true},
		{"uno registrado y otro no", "**¿Salió algo mal?:** D000001 y D999999", true},
		{"número mal escrito", "**¿Salió algo mal?:** D104", true},
	} {
		if m := defecto.RevisarCorreccion(raiz, "corrección", c.msg); (m != "") != c.frena {
			t.Errorf("RevisarCorreccion %s: frena=%v, quiere %v; %q", c.caso, m != "", c.frena, m)
		}
	}
	casos := []struct {
		caso, literal, respuesta string
		codigo                   int
	}{
		{"corrección con «No»", "eso está mal", "Tipo: corrección\nPedido refinado: arreglar X.\n\n🤦‍♂️ **¿Salió algo mal?:** No", 2},
		{"corrección sin el campo", "eso está mal", "Tipo: corrección\nPedido refinado: arreglar X.\n\nFALLO: arreglado.", 2},
		{"corrección con defecto registrado", "eso está mal", "Tipo: corrección\nPedido refinado: arreglar X.\n\n🤦‍♂️ **¿Salió algo mal?:** sí, D000001 (repetido).", 0},
		{"corrección con defecto inexistente", "eso está mal", "Tipo: corrección\nPedido refinado: arreglar X.\n\n🤦‍♂️ **¿Salió algo mal?:** sí, D999999.", 2},
		{"corrección con número mal escrito", "eso está mal", "Tipo: corrección\nPedido refinado: arreglar X.\n\n🤦‍♂️ **¿Salió algo mal?:** sí, D104.", 2},
		{"corrección con las 5 preguntas y su defecto", "eso está mal", "Tipo: corrección\nPedido refinado: arreglar X.\n\n" + cinco + "\n🤦‍♂️ **¿Salió algo mal?:** sí, D000001.", 0},
		{"petición con «No»", "haz X", "Tipo: petición\nPedido refinado: hacer X.\n\n🤦‍♂️ **¿Salió algo mal?:** No", 0},
		{"corrección en turno de agente", "<agent-message from=\"a1\">informe</agent-message>", "Tipo: corrección\nPedido refinado: x.\n\n🤦‍♂️ **¿Salió algo mal?:** No", 0},
	}
	for i, c := range casos {
		sesion := fmt.Sprintf("s%d", i)
		b, _ := json.Marshal(map[string]string{"session_id": sesion, "prompt": c.literal})
		if s := aquas(t, raiz, string(b), "turn", "open"); s.codigo != 0 {
			t.Fatalf("%s: turn open: %+v", c.caso, s)
		}
		b, _ = json.Marshal(map[string]string{"session_id": sesion, "last_assistant_message": c.respuesta})
		if s := aquas(t, raiz, string(b), "turn", "close"); s.codigo != 0 {
			t.Fatalf("%s: turn close: %+v", c.caso, s)
		}
		if s := aquas(t, raiz, string(b), "proposal"); s.codigo != c.codigo {
			t.Errorf("%s: código %d, quiere %d; %+v", c.caso, s.codigo, c.codigo, s)
		}
	}
	// Hallazgo A1.3: segunda vuelta del gancho tras un freno; la respuesta corta no trae «Tipo:» y vale el guardado.
	sesion := "vueltas"
	b, _ := json.Marshal(map[string]string{"session_id": sesion, "prompt": "eso está mal"})
	if s := aquas(t, raiz, string(b), "turn", "open"); s.codigo != 0 {
		t.Fatalf("turn open: %+v", s)
	}
	for _, v := range []struct {
		respuesta string
		codigo    int
	}{
		{"Tipo: corrección\nPedido refinado: arreglar X.\n\n🤦‍♂️ **¿Salió algo mal?:** D000001", 0},
		{"Corrijo el campo.\n\n🤦‍♂️ **¿Salió algo mal?:** No", 2},
	} {
		b, _ = json.Marshal(map[string]string{"session_id": sesion, "last_assistant_message": v.respuesta})
		if s := aquas(t, raiz, string(b), "turn", "close"); s.codigo != 0 {
			t.Fatalf("turn close: %+v", s)
		}
		if s := aquas(t, raiz, string(b), "proposal"); s.codigo != v.codigo {
			t.Errorf("vuelta %q: código %d, quiere %d; %+v", v.respuesta, s.codigo, v.codigo, s)
		}
	}
}
