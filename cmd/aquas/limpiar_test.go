package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestLimpiarSoloTmpDelEncargo (E000/F006/S009): aquas limpiar borra .tmp/<id>/ del encargo y nada más.
// Rojo cuando: el comando no existe o borra fuera de .tmp/<id>/.
// Verde cuando: borra la carpeta del encargo y rechaza los demás casos sin borrar nada.
func TestLimpiarSoloTmpDelEncargo(t *testing.T) {
	raiz := t.TempDir()
	escribir := func(rel string) {
		p := filepath.Join(raiz, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	existe := func(rel string) bool {
		_, err := os.Lstat(filepath.Join(raiz, filepath.FromSlash(rel)))
		return err == nil
	}
	vecinos := []string{".tmp/aquas-hook.exe", ".tmp/s008r/a.txt", ".tmp/E000-F006-S010-T1/b.txt", "fuera.txt"}
	for _, v := range vecinos {
		escribir(v)
	}
	escribir(".tmp/E000-F006-S010-A1/sub/c.txt")

	// Criterio 1: id válido con «/» borra solo su carpeta.
	if _, err := limpiar(raiz, []string{"E000/F006/S010-A1"}); err != nil {
		t.Fatalf("id válido debió borrar: %v", err)
	}
	if existe(".tmp/E000-F006-S010-A1") {
		t.Error("la carpeta del encargo debió desaparecer")
	}
	for _, v := range vecinos {
		if !existe(v) {
			t.Errorf("%s debió quedar intacto", v)
		}
	}

	// Criterio 2: ids inválidos se rechazan sin borrar.
	for _, malo := range []string{"", ".tmp", "..", "../fuera", "s008r", "E000-F006-S010", "E000-F006-S010-A1/..", "E000-F006-S010-T1/b.txt", filepath.Join(raiz, ".tmp"), "E000-F006-S010-X1",
		"E000/F006/S010/T1", "E000/F006-S010-T1", `E000\F006\S010-T1`, "e000-f006-s010-t1", "E000-F006-S010-T1\n", "E000-F006-S010-T1 "} {
		if _, err := limpiar(raiz, []string{malo}); err == nil {
			t.Errorf("id %q debió rechazarse", malo)
		}
	}
	if _, err := limpiar(raiz, nil); err == nil {
		t.Error("sin id debió rechazarse")
	}
	for _, v := range vecinos {
		if !existe(v) {
			t.Errorf("%s debió quedar intacto tras los rechazos", v)
		}
	}

	// Criterio 3: un enlace dentro frena todo y no se sigue.
	escribir("afuera/tesoro.txt")
	escribir(".tmp/E000-F006-S011-A1/d.txt")
	enlace := filepath.Join(raiz, ".tmp", "E000-F006-S011-A1", "enlace")
	destino := filepath.Join(raiz, "afuera")
	if err := os.Symlink(destino, enlace); err != nil {
		if runtime.GOOS != "windows" {
			t.Fatal(err)
		}
		if out, err := exec.Command("cmd", "/c", "mklink", "/J", enlace, destino).CombinedOutput(); err != nil {
			t.Fatalf("no se pudo crear el enlace: %v %s", err, out)
		}
	}
	if _, err := limpiar(raiz, []string{"E000-F006-S011-A1"}); err == nil {
		t.Error("carpeta con enlace debió rechazarse")
	}
	if !existe(".tmp/E000-F006-S011-A1/d.txt") || !existe("afuera/tesoro.txt") {
		t.Error("con enlace no debió borrarse nada")
	}

	// Criterio 4: id válido sin carpeta avisa y termina bien.
	msg, err := limpiar(raiz, []string{"E001-F002-S019-T2"})
	if err != nil || !strings.HasPrefix(msg, "nada que borrar") {
		t.Errorf("sin carpeta debió avisar «nada que borrar» sin error; dio %q, %v", msg, err)
	}

	// La carpeta del encargo misma como enlace: se rechaza y el destino queda.
	escribir("afuera2/x.txt")
	enlazar(t, filepath.Join(raiz, ".tmp", "E000-F006-S012-A1"), filepath.Join(raiz, "afuera2"))
	if _, err := limpiar(raiz, []string{"E000-F006-S012-A1"}); err == nil || !existe("afuera2/x.txt") {
		t.Errorf("carpeta del encargo como enlace debió rechazarse sin borrar; err=%v", err)
	}

	// .tmp como enlace: se rechaza y nada se borra.
	raiz2 := t.TempDir()
	real := filepath.Join(raiz2, "real", "E000-F006-S010-A1")
	if err := os.MkdirAll(real, 0o755); err != nil {
		t.Fatal(err)
	}
	enlazar(t, filepath.Join(raiz2, ".tmp"), filepath.Join(raiz2, "real"))
	if _, err := limpiar(raiz2, []string{"E000-F006-S010-A1"}); err == nil {
		t.Error(".tmp como enlace debió rechazarse")
	}
	if _, err := os.Stat(real); err != nil {
		t.Errorf(".tmp como enlace no debió borrar: %v", err)
	}

	// Un archivo con nombre de id no es una carpeta de encargo: se rechaza y queda.
	escribir(".tmp/E000-F006-S013-T1")
	if _, err := limpiar(raiz, []string{"E000-F006-S013-T1"}); err == nil || !existe(".tmp/E000-F006-S013-T1") {
		t.Errorf("archivo con nombre de id debió rechazarse sin borrar; err=%v", err)
	}
}

// enlazar crea un enlace a una carpeta: symlink, o junction en Windows sin privilegio.
func enlazar(t *testing.T, enlace, destino string) {
	t.Helper()
	if err := os.Symlink(destino, enlace); err == nil {
		return
	} else if runtime.GOOS != "windows" {
		t.Fatal(err)
	}
	if out, err := exec.Command("cmd", "/c", "mklink", "/J", enlace, destino).CombinedOutput(); err != nil {
		t.Fatalf("no se pudo crear el enlace: %v %s", err, out)
	}
}
