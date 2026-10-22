package hijo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

const constitucionConHuecos = `# CONSTITUCIÓN

## Misión
<!-- hueco: mision -->
- Misión de la central.
<!-- /hueco -->

## Principios
1. Regla final de la central.
`

// instalarConHuecos deja un hijo instalado y sus huecos locales escritos.
func instalarConHuecos(t *testing.T, huecos map[string]string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "hijo")
	central := fstest.MapFS{"gobierno/canonico/constitucion.md": {Data: []byte(constitucionConHuecos)}}
	if err := Install(dir, central, "v1.0.0"); err != nil {
		t.Fatal(err)
	}
	for nombre, txt := range huecos {
		p := filepath.Join(dir, CarpetaHuecos, nombre+".md")
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(txt), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestBuildRellenaHuecosSinTocarFinal(t *testing.T) {
	// Pasa: el hueco se llena con lo local y lo final queda intacto.
	dir := instalarConHuecos(t, map[string]string{"mision": "- Landing de Aquas.\n"})
	if err := Build(dir); err != nil {
		t.Fatalf("build válido falló: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}
	out := string(b)
	for _, debe := range []string{"GENERADO", "- Landing de Aquas.", "1. Regla final de la central."} {
		if !strings.Contains(out, debe) {
			t.Errorf("CLAUDE.md sin %q:\n%s", debe, out)
		}
	}
	if strings.Contains(out, "Misión de la central") || strings.Contains(out, "<!-- hueco") {
		t.Errorf("quedó el relleno o el marcador de la central:\n%s", out)
	}

	// Falla: lo local apunta a una sección final (no es hueco).
	dir = instalarConHuecos(t, map[string]string{"mision": "x\n", "principios": "otra regla\n"})
	if err := Build(dir); err == nil || !strings.Contains(err.Error(), "principios") {
		t.Errorf("redefinir una sección final debió fallar nombrándola; err=%v", err)
	}

	// Falla: hueco sin llenar.
	dir = instalarConHuecos(t, nil)
	if err := Build(dir); err == nil || !strings.Contains(err.Error(), "mision") {
		t.Errorf("hueco sin llenar debió fallar nombrándolo; err=%v", err)
	}

	// Falla: alguien editó lo sellado.
	dir = instalarConHuecos(t, map[string]string{"mision": "x\n"})
	sellado := filepath.Join(dir, Carpeta, "central", "gobierno", "canonico", "constitucion.md")
	if err := os.Chmod(sellado, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sellado, []byte("# alterado\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Build(dir); err == nil || !strings.Contains(err.Error(), "sellado") {
		t.Errorf("central alterada debió fallar por sello; err=%v", err)
	}
}

const constitucionConTitulo = `<!-- hueco: titulo -->
# CONSTITUCIÓN · central
<!-- /hueco -->

## Compilación
<!-- solo-central -->
- Compilar: go run ./cmd/aquas compilar.
<!-- /solo-central -->
<!-- solo-hijo
- Compilar: aquas build.
-->
`

func TestBuildTituloYCompilacionDelHijo(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "hijo")
	central := fstest.MapFS{"gobierno/canonico/constitucion.md": {Data: []byte(constitucionConTitulo)}}
	if err := Install(dir, central, "v1.0.0"); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, CarpetaHuecos, "titulo.md")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("# CONSTITUCIÓN · landing\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Build(dir); err != nil {
		t.Fatalf("build falló: %v", err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	out := string(b)
	for _, debe := range []string{"# CONSTITUCIÓN · landing", "- Compilar: aquas build."} {
		if !strings.Contains(out, debe) {
			t.Errorf("CLAUDE.md sin %q:\n%s", debe, out)
		}
	}
	for _, no := range []string{"· central", "cmd/aquas", "solo-central", "solo-hijo"} {
		if strings.Contains(out, no) {
			t.Errorf("CLAUDE.md del hijo trae %q de la central:\n%s", no, out)
		}
	}
}
