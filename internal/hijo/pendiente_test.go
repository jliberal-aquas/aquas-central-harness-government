package hijo

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"testing/fstest"
)

// Tres huecos, como la central real: los casos de «varios huecos» se ejercitan (D000056).
const constitucionTresHuecos = `<!-- hueco: titulo -->
# CONSTITUCIÓN · central
<!-- /hueco -->

## Misión
<!-- hueco: mision -->
- Misión de la central.
<!-- /hueco -->

## Mapa
<!-- hueco: mapa -->
- Mapa de la central.
<!-- /hueco -->
`

func avisoPendiente(n string) string {
	return "hueco pendiente: llena " + CarpetaHuecos + "/" + n + ".md y corre aquas build"
}

// TestVerificarHuecoPendiente (E002/F002/S004): el gate del hijo falla mientras un hueco conserve «PENDIENTE:».
// Rojo cuando: verificar pasa con un hueco PENDIENTE o no nombra cuál.
// Verde cuando: falla nombrando cada hueco pendiente y pasa con los huecos escritos.
func TestVerificarHuecoPendiente(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "hijo")
	central := fstest.MapFS{"gobierno/canonico/constitucion.md": {Data: []byte(constitucionTresHuecos)}}
	if err := Install(dir, central, "v1.0.0"); err != nil {
		t.Fatal(err)
	}
	if err := Listo(dir); err != nil {
		t.Fatal(err)
	}
	// Recién instalado: los tres, título incluido (su relleno es distinto al de los demás).
	quiero := []string{avisoPendiente("mapa"), avisoPendiente("mision"), avisoPendiente("titulo")}
	if got := Verificar(dir); !reflect.DeepEqual(got, quiero) {
		t.Errorf("recién instalado: se esperaba %q; dio %q", quiero, got)
	}
	hd := filepath.Join(dir, filepath.FromSlash(CarpetaHuecos))
	escribir := func(nombre, txt string) {
		if err := os.WriteFile(filepath.Join(hd, nombre+".md"), []byte(txt), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// Solo mision pendiente; «pendiente» en minúscula o «PENDIENTE» sin dos puntos no cuentan.
	escribir("titulo", "# CONSTITUCIÓN · landing · v0.1\n")
	escribir("mapa", "- `web/` → sitio; estado: PENDIENTE de revisión; queda una tarea pendiente.\n")
	if err := Build(dir); err != nil {
		t.Fatal(err)
	}
	quiero = []string{avisoPendiente("mision")}
	if got := Verificar(dir); !reflect.DeepEqual(got, quiero) {
		t.Errorf("solo mision pendiente: se esperaba %q; dio %q", quiero, got)
	}
	// Pendiente y CLAUDE.md desfasado: los dos avisos, cada uno con su texto.
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("a mano\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	quiero = []string{avisoPendiente("mision"), "CLAUDE.md desfasado o ausente: corre aquas build " + dir}
	if got := Verificar(dir); !reflect.DeepEqual(got, quiero) {
		t.Errorf("pendiente y desfasado: se esperaba %q; dio %q", quiero, got)
	}
	escribir("mision", "- Landing de Aquas.\n")
	if err := Build(dir); err != nil {
		t.Fatal(err)
	}
	if got := Verificar(dir); len(got) != 0 {
		t.Errorf("huecos escritos y CLAUDE.md al día: debió pasar; dio %q", got)
	}
}
