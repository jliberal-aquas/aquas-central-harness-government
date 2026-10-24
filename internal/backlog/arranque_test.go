package backlog

import (
	"strings"
	"testing"
)

// TestArranqueExigeRojoDeLaIntegral: un slice no pasa a EN_CURSO si su feature declara una integral sin rojo (D000023).
// Rojo cuando: el slice arranca con la integral de su feature sin rojo, o no arranca aunque la integral ya tenga rojo o la feature no declare pruebas.
// Verde cuando: sin rojo se niega nombrando el comando; con rojo, o sin integral, arranca.
func TestArranqueExigeRojoDeLaIntegral(t *testing.T) {
	raiz := base(t)
	f, _, err := Ver(raiz, "E001/F002")
	if err != nil {
		t.Fatal(err)
	}
	f.Rojo = nil
	pisar(t, raiz, Activo, f)
	nuevo(t, raiz, "id=E001/F002/S001", "titulo=T", "prioridad=P1", "motivo=m")
	if _, err := arrancar(t, raiz, "E001/F002/S001"); err == nil || !strings.Contains(err.Error(), "aquas backlog rojo E001/F002") {
		t.Errorf("integral sin rojo: el slice debió negarse a arrancar; dio %v", err)
	}
	f.Rojo = &Registro{Fecha: "2026-10-21T12:00:00Z", Resultado: map[string]string{"./p:TestFalla": "x"}}
	pisar(t, raiz, Activo, f)
	if _, err := arrancar(t, raiz, "E001/F002/S001"); err != nil {
		t.Errorf("integral con rojo: debió arrancar; dio %v", err)
	}
	if _, err := Estado(raiz, "E001/F002/S001", "CANCELADO"); err != nil { // un solo slice en curso
		t.Fatal(err)
	}
	nuevo(t, raiz, "id=E001/F003", "titulo=Sin integral", "prioridad=P1", "motivo=m")
	g, _, _ := Ver(raiz, "E001/F003")
	g.Pruebas = nil
	pisar(t, raiz, Activo, g)
	nuevo(t, raiz, "id=E001/F003/S001", "titulo=T", "prioridad=P1", "motivo=m")
	if _, err := arrancar(t, raiz, "E001/F003/S001"); err != nil {
		t.Errorf("feature sin integral: debió arrancar; dio %v", err)
	}
}
