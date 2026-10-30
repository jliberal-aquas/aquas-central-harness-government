package defecto

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestStopFrenaFalloSinDefecto (E001/F004/S002): un fallo contado en el cierre trae un defecto que existe.
// Rojo cuando: un ¿Salió algo mal? sin defecto no frena el cierre.
// Verde cuando: sin defecto o con uno inexistente frena; con No, con — o con un defecto que existe pasa.
func TestStopFrenaFalloSinDefecto(t *testing.T) {
	raiz := t.TempDir()
	if err := os.MkdirAll(filepath.Join(raiz, Carpeta), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(raiz, Carpeta, "D000001.json"), []byte(`{"id":"D000001"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(raiz, Carpeta, "D000003.json"), 0o755); err != nil {
		t.Fatal(err)
	}
	campo := "- 🤦‍♂️ **¿Salió algo mal?:** "
	casos := []struct {
		nombre, msg, frena string // frena: "" = pasa; si no, texto que debe traer el motivo
	}{
		{"sin campo", "Todo listo.", ""},
		{"No", campo + "No", ""},
		{"No con punto", campo + "No.", ""},
		{"raya", campo + "—", ""},
		{"negrita después de los dos puntos", "🤦‍♂️ **¿Salió algo mal?**: No", ""},
		{"mención sin dos puntos", "el campo «🤦‍♂️ ¿Salió algo mal?» puede contar un fallo", ""},
		{"defecto que existe", campo + "la prueba tenía un error (D000001)", ""},
		{"fallo sin defecto", campo + "la prueba tenía un error", "Fallo sin defecto"},
		{"valor vacío", campo, "Fallo sin defecto"},
		{"No con más texto", campo + "No hubo nada grave", "Fallo sin defecto"},
		{"número mal formado", campo + "falló (D12)", "Fallo sin defecto"},
		{"defecto inexistente", campo + "falló (D000002)", "D000002"},
		{"uno existe y otro no", campo + "falló (D000001, D000002)", "D000002"},
		// Hallazgo A1.1: una cita del campo en el cuerpo no es el campo.
		{"cita en el cuerpo", "2. Escribo «¿Salió algo mal?: sí, el gancho no frenó».\n" + campo + "No", ""},
		{"cita entre comillas", "Ejemplo: «🤦‍♂️ ¿Salió algo mal?: la prueba falló»", ""},
		// Hallazgo A1.3: un número con más dígitos no es D000NNN.
		{"número con siete dígitos", campo + "falló (D0000012)", "Fallo sin defecto"},
		// Hallazgo A1.2: una carpeta con nombre de defecto no es un defecto registrado.
		{"carpeta con nombre de defecto", campo + "falló (D000003)", "D000003"},
		{"dos campos, el segundo sin defecto", campo + "No\n" + campo + "se cayó el gate", "Fallo sin defecto"},
	}
	for _, c := range casos {
		m := RevisarFallo(raiz, c.msg)
		if c.frena == "" && m != "" {
			t.Errorf("%s: debe pasar; frenó con %q", c.nombre, m)
		}
		if c.frena != "" && !strings.Contains(m, c.frena) {
			t.Errorf("%s: debe frenar nombrando %q; dio %q", c.nombre, c.frena, m)
		}
	}
	// Registro ausente o ilegible: ningún número existe, así que frena.
	vacia := t.TempDir()
	if m := RevisarFallo(vacia, campo+"falló (D000001)"); !strings.Contains(m, "D000001") {
		t.Errorf("sin registro: frena nombrando D000001; dio %q", m)
	}
	rota := t.TempDir()
	if err := os.MkdirAll(filepath.Join(rota, "registro"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(rota, Carpeta), []byte("no soy carpeta"), 0o644); err != nil {
		t.Fatal(err)
	}
	if m := RevisarFallo(rota, campo+"falló (D000001)"); !strings.Contains(m, "D000001") {
		t.Errorf("registro ilegible: frena nombrando D000001; dio %q", m)
	}
}
