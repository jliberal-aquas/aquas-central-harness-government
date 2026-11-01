package main

import (
	"io/fs"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	gobcentral "github.com/jliberal-aquas/aquas-central-harness-government"
)

// TestBinarioSinEntornoLocal (E003/F001/S001): lo que el binario lleva de gobierno/config es la config versionada,
// sin archivos locales de la máquina que compiló.
// Rojo cuando: entra un *.local.json o falta un archivo versionado.
// Verde cuando: lo embebido coincide con `git ls-files gobierno/config` y no trae ningún *.local.json.
func TestBinarioSinEntornoLocal(t *testing.T) {
	var embebidos []string
	err := fs.WalkDir(gobcentral.Central, "gobierno/config", func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			embebidos = append(embebidos, strings.TrimPrefix(p, "gobierno/config/"))
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	// A1 hallazgo 2: «versionado» = lo que Git rastrea, no lo que haya en disco.
	out, err := exec.Command("git", "-C", filepath.Join("..", ".."), "ls-files", "gobierno/config").Output()
	if err != nil {
		t.Fatal(err)
	}
	var versionados []string
	for _, l := range strings.Fields(string(out)) {
		versionados = append(versionados, strings.TrimPrefix(l, "gobierno/config/"))
	}
	for _, e := range embebidos {
		if strings.HasSuffix(e, ".local.json") {
			t.Errorf("el binario lleva un archivo local: %s", e)
		}
	}
	sort.Strings(embebidos)
	sort.Strings(versionados)
	if strings.Join(embebidos, ",") != strings.Join(versionados, ",") {
		t.Errorf("el binario debe llevar justo la config versionada, sin *.local.json:\n embebidos   %v\n versionados %v", embebidos, versionados)
	}
}
