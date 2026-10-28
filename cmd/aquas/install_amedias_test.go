package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestInstallAMediasCLI (E002/F002/S002): si install o update fallan tras copiar la central, la CLI manda a update.
// Rojo cuando: el aviso de la CLI no dice cómo seguir o manda a repetir install.
// Verde cuando: install a medias pide «corre aquas update» sin «repite» y update a medias pide «repite aquas update».
func TestInstallAMediasCLI(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".claude", "settings.json"), []byte("{ roto"), 0o644); err != nil {
		t.Fatal(err)
	}
	s := aquas(t, dir, "", "install")
	if s.codigo == 0 || !strings.Contains(s.err, "corre aquas update") || strings.Contains(s.err, "repite") {
		t.Errorf("install a medias: se esperaba fallo que mande a aquas update sin «repite»; exit %d\n%s", s.codigo, s.err)
	}
	s = aquas(t, dir, "", "update")
	if s.codigo == 0 || !strings.Contains(s.err, "repite aquas update") {
		t.Errorf("update a medias: se esperaba fallo que pida repetir aquas update; exit %d\n%s", s.codigo, s.err)
	}
}
