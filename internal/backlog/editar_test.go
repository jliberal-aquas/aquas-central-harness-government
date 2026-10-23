package backlog

import (
	"strings"
	"testing"
)

// TestEditarSoloAntesDeArrancar: backlog editar cambia una ficha que aún no arranca y guarda lo anterior.
// Rojo cuando: edita una ficha EN_CURSO o cerrada, acepta un campo no editable o sin motivo, o no deja historial.
// Verde cuando: en BORRADOR cambia título y bloqueo con historial; lo demás se niega.
func TestEditarSoloAntesDeArrancar(t *testing.T) {
	raiz := base(t)
	nuevo(t, raiz, "id=E001/F002/S001", "titulo=Viejo", "prioridad=P1", "motivo=m")
	nuevo(t, raiz, "id=E001/F002/S002", "titulo=Otro", "prioridad=P1", "motivo=m")
	f, err := Editar(raiz, "E001/F002/S001", []string{"titulo=Nuevo", "bloqueado_por=E001/F002/S002", "motivo=pidió Jesús"}, hora)
	if err != nil {
		t.Fatalf("editar en BORRADOR: %v", err)
	}
	g := leerFicha(t, activo(raiz, "E001-F002-S001.json"))
	for _, x := range []Ficha{f, g} {
		if x.Titulo != "Nuevo" || strings.Join(x.BloqueadoPor, ",") != "E001/F002/S002" || len(x.Ediciones) != 2 {
			t.Fatalf("edición no aplicada o sin historial: %+v", x)
		}
		e := x.Ediciones[0]
		if e.Campo != "titulo" || e.Antes != "Viejo" || e.Despues != "Nuevo" || e.Motivo != "pidió Jesús" || e.Fecha != "2026-10-21T12:00:00Z" {
			t.Errorf("historial incompleto: %+v", e)
		}
	}
	for _, c := range []struct {
		nombre string
		campos []string
		dice   string
	}{
		{"sin motivo", []string{"titulo=X"}, "motivo"},
		{"sin cambios", []string{"motivo=m"}, "nada que editar"},
		{"campo no editable", []string{"estado=CERRADO", "motivo=m"}, "no editable"},
		{"bloqueo inexistente", []string{"bloqueado_por=E001/F002/S099", "motivo=m"}, "no existe"},
	} {
		if _, err := Editar(raiz, "E001/F002/S001", c.campos, hora); err == nil || !strings.Contains(err.Error(), c.dice) {
			t.Errorf("%s: se esperaba error con %q; dio %v", c.nombre, c.dice, err)
		}
	}
	enCurso(t, raiz, "E001/F002/S003", "./p:TestFalla=falla|r|v")
	if _, err := Editar(raiz, "E001/F002/S003", []string{"titulo=X", "motivo=m"}, hora); err == nil || !strings.Contains(err.Error(), "BORRADOR o LISTO") {
		t.Errorf("editar EN_CURSO: debió negarse; dio %v", err)
	}
}
