package hijo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestUpdateRespetaLock(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "hijo")
	v1 := fstest.MapFS{
		"gobierno/canonico/constitucion.md": {Data: []byte(constitucionConHuecos)},
		"gobierno/canonico/vieja.md":        {Data: []byte("se va en v2\n")},
	}
	if err := Install(dir, v1, "v1.0.0"); err != nil {
		t.Fatal(err)
	}
	local := filepath.Join(dir, CarpetaHuecos, "mision.md")
	if err := os.MkdirAll(filepath.Dir(local), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(local, []byte("- Landing.\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Editar lo sellado hace caer el gate (build).
	sellado := filepath.Join(dir, Carpeta, "central", "gobierno", "canonico", "constitucion.md")
	if err := os.Chmod(sellado, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sellado, []byte("# alterado\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Build(dir); err == nil {
		t.Fatal("build sobre lo sellado alterado debió fallar")
	}

	// Update cambia de versión, restaura lo alterado y conserva lo local.
	v2 := fstest.MapFS{"gobierno/canonico/constitucion.md": {Data: []byte(constitucionConHuecos + "2. Regla nueva.\n")}}
	rest, err := Update(dir, v2, "v2.0.0")
	if err != nil {
		t.Fatalf("update falló: %v", err)
	}
	if len(rest) != 1 || rest[0] != "gobierno/canonico/constitucion.md" {
		t.Errorf("restaurados = %v; quería la constitución alterada", rest)
	}
	if b, _ := os.ReadFile(filepath.Join(dir, Carpeta, "lock.json")); !strings.Contains(string(b), "v2.0.0") {
		t.Errorf("lock sin v2.0.0:\n%s", b)
	}
	if _, err := os.Stat(filepath.Join(dir, Carpeta, "central", "gobierno", "canonico", "vieja.md")); err == nil {
		t.Error("quedó un archivo que v2 ya no trae")
	}
	if b, _ := os.ReadFile(local); string(b) != "- Landing.\n" {
		t.Errorf("update pisó lo local: %q", b)
	}
	if err := Build(dir); err != nil {
		t.Fatalf("build tras update falló: %v", err)
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "CLAUDE.md")); !strings.Contains(string(b), "Regla nueva") {
		t.Error("CLAUDE.md no trae la regla de v2")
	}

	// Falla: sin instalación previa, o sin versión.
	if _, err := Update(filepath.Join(t.TempDir(), "nada"), v2, "v2.0.0"); err == nil || !strings.Contains(err.Error(), "install") {
		t.Errorf("update sin instalación debió pedir install; err=%v", err)
	}
	if _, err := Update(dir, v2, ""); err == nil {
		t.Error("update sin versión debió fallar")
	}
}
