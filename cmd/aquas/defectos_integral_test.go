package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestDefectosDePuntaAPunta (E001/F004, integral): un fallo se registra, se repite y a la tercera aparición
// nace su slice; el cierre que lo cita pasa y el que calla el número frena.
// Rojo cuando: la tercera aparición no crea el slice o el cierre no exige el defecto.
// Verde cuando: nace un solo slice P1 en E000/F007 y el cierre solo pasa citando un defecto que existe.
func TestDefectosDePuntaAPunta(t *testing.T) {
	raiz := t.TempDir()
	ficha := func(id, titulo string) {
		t.Helper()
		if s := aquas(t, raiz, "", "backlog", "nuevo", "id="+id, "titulo="+titulo, "prioridad=P1", "motivo=prueba",
			"que_hace=agrupa trabajo", "relacion=nueva", "prueba=./internal/x:TestX=valida algo|rojo cuando falla|verde cuando pasa"); s.codigo != 0 {
			t.Fatalf("backlog nuevo %s: %+v", id, s)
		}
	}
	ficha("E000", "Mantenimiento")
	ficha("E000/F007", "Reincidencias de defectos")
	cierre := func(msg string) salida {
		t.Helper()
		b, _ := json.Marshal(map[string]string{"hook_event_name": "Stop", "last_assistant_message": msg})
		return aquas(t, raiz, string(b), "propuesta")
	}
	campo := "- 🤦‍♂️ **¿Salió algo mal?:** "

	// 1. Un fallo contado sin número frena el cierre.
	if s := cierre(campo + "el JSON tenía un salto crudo"); s.codigo != 2 || !strings.Contains(s.err, "Fallo sin defecto") {
		t.Fatalf("fallo sin defecto debe frenar: %+v", s)
	}
	// 2. Se registra; el cierre que lo cita pasa.
	s := aquas(t, raiz, "", "defecto", "nuevo", "firma=json-salto-crudo", "que=salto crudo en JSON", "causa=comillas dobles",
		"arreglo=backticks", "cercioro=prueba", "relacion=nuevo")
	if s.codigo != 0 || !strings.Contains(s.out, "D000001") {
		t.Fatalf("defecto nuevo: %+v", s)
	}
	if s := cierre(campo + "el JSON tenía un salto crudo (D000001)"); s.codigo != 0 {
		t.Fatalf("cierre con defecto que existe debe pasar: %+v", s)
	}
	// 3. Segunda aparición: no nace slice. Tercera: nace uno, P1, en E000/F007.
	if s := aquas(t, raiz, "", "defecto", "ocurre", "D000001", "T000001"); s.codigo != 0 || strings.Contains(s.out, "slice") {
		t.Fatalf("segunda aparición: %+v", s)
	}
	if s := aquas(t, raiz, "", "defecto", "ocurre", "D000001", "T000002"); s.codigo != 0 || !strings.Contains(s.out, "slice E000/F007/S001") {
		t.Fatalf("tercera aparición debe nombrar su slice: %+v", s)
	}
	if s := aquas(t, raiz, "", "backlog", "ver", "E000/F007/S001"); s.codigo != 0 ||
		!strings.Contains(s.out, "P1") || !strings.Contains(s.out, "Reincidencia D000001: json-salto-crudo") {
		t.Fatalf("slice de reincidencia: %+v", s)
	}
	// 4. Cuarta aparición: no nace otro.
	if s := aquas(t, raiz, "", "defecto", "ocurre", "D000001", "T000003"); s.codigo != 0 {
		t.Fatalf("cuarta aparición: %+v", s)
	}
	if s := aquas(t, raiz, "", "backlog", "ver", "E000/F007/S002"); s.codigo == 0 {
		t.Fatalf("la cuarta aparición creó otro slice: %+v", s)
	}
}
