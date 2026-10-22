package hijo

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

// TestInstallCopiaCentralConLock: install copia la central a .aquas/central con un lock
// de versión y hashes; no instala sobre una carpeta con contenido ni sin versión.
func TestInstallCopiaCentralConLock(t *testing.T) {
	const con = "# Constitución\n"
	central := fstest.MapFS{
		"gobierno/canonico/constitucion.md": {Data: []byte(con)},
		"gobierno/config/agentes.json":      {Data: []byte("{}\n")},
	}
	dest := filepath.Join(t.TempDir(), "landing")
	if err := Install(dest, central, "v0.1.0"); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dest, Carpeta, "central", "gobierno", "canonico", "constitucion.md"))
	if err != nil || string(b) != con {
		t.Fatalf("constitución copiada: %q, %v", b, err)
	}
	var l Lock
	b, err = os.ReadFile(filepath.Join(dest, Carpeta, "lock.json"))
	if err != nil {
		t.Fatal("sin lock:", err)
	}
	if err := json.Unmarshal(b, &l); err != nil {
		t.Fatal(err)
	}
	h := sha256.Sum256([]byte(con))
	if l.Version != "v0.1.0" || len(l.Archivos) != 2 || l.Archivos["gobierno/canonico/constitucion.md"] != hex.EncodeToString(h[:]) {
		t.Fatalf("lock: %+v", l)
	}

	if err := Install(dest, central, "v0.1.0"); err == nil {
		t.Fatal("instalar dos veces en la misma carpeta debe fallar (eso es update)")
	}
	lleno := t.TempDir()
	if err := os.WriteFile(filepath.Join(lleno, "x.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Install(lleno, central, "v0.1.0"); err == nil {
		t.Fatal("carpeta con contenido debe fallar")
	}
	if err := Install(filepath.Join(t.TempDir(), "otro"), central, ""); err == nil {
		t.Fatal("sin versión debe fallar")
	}
}
