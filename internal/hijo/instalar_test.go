package hijo

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
)

// foto devuelve ruta relativa → contenido de cada archivo bajo dir.
func foto(t *testing.T, dir string) map[string]string {
	t.Helper()
	m := map[string]string{}
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := os.ReadFile(p)
		rel, _ := filepath.Rel(dir, p)
		m[filepath.ToSlash(rel)] = string(b)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return m
}

// carpeta crea dir con los archivos dados.
func carpeta(t *testing.T, archivos map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for p, txt := range archivos {
		q := filepath.Join(dir, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(q), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(q, []byte(txt), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// TestInstallEnCarpetaConCodigo (E002/F002/S002): install en un repo con .git o código.
// Rojo cuando: install rechaza la carpeta con código, altera un archivo previo o escribe algo ante un conflicto.
// Verde cuando: instala con lo previo idéntico y ante cada conflicto falla sin escribir.
func TestInstallEnCarpetaConCodigo(t *testing.T) {
	central := fstest.MapFS{"gobierno/canonico/constitucion.md": {Data: []byte(constitucionConHuecos)}}
	previo := map[string]string{".git/HEAD": "ref: refs/heads/main\n", "index.html": "<h1>hola</h1>\n"}
	// Criterio 1: carpeta con .git y código → instala; lo previo idéntico.
	dir := carpeta(t, previo)
	if err := Install(dir, central, "v1.0.0"); err != nil {
		t.Fatalf("install en carpeta con código debió funcionar: %v", err)
	}
	if err := Listo(dir); err != nil {
		t.Fatal(err)
	}
	f := foto(t, dir)
	for p, txt := range previo {
		if f[p] != txt {
			t.Errorf("%s alterado: %q", p, f[p])
		}
	}
	if _, ok := f["CLAUDE.md"]; !ok {
		t.Error("install debió dejar CLAUDE.md")
	}
	// Criterios 2 y 3: conflicto → falla con el aviso y sin escribir nada.
	for _, c := range []struct {
		nombre string
		extra  map[string]string
		aviso  string
	}{
		{"ya es hijo", map[string]string{Carpeta + "/lock.json": "{}"}, "aquas update"},
		{"CLAUDE.md propio", map[string]string{"CLAUDE.md": "# Mis reglas\n"}, CarpetaHuecos},
	} {
		arch := map[string]string{}
		for p, txt := range previo {
			arch[p] = txt
		}
		for p, txt := range c.extra {
			arch[p] = txt
		}
		dir := carpeta(t, arch)
		antes := foto(t, dir)
		err := Install(dir, central, "v1.0.0")
		if err == nil || !strings.Contains(err.Error(), c.aviso) {
			t.Errorf("%s: se esperaba error que nombre %q; dio %v", c.nombre, c.aviso, err)
		}
		if despues := foto(t, dir); !reflect.DeepEqual(antes, despues) {
			t.Errorf("%s: install escribió pese al conflicto", c.nombre)
		}
	}
}
