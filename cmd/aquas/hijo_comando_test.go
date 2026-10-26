package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestHijoEnUnComando (integral E002/F002): instalar y actualizar un hijo es un comando cada uno.
// Rojo cuando: install o update no dejan CLAUDE.md armado.
// Verde cuando: cada uno, en la carpeta del hijo, termina con CLAUDE.md.
func TestHijoEnUnComando(t *testing.T) {
	dir := t.TempDir()
	claude := filepath.Join(dir, "CLAUDE.md")
	if s := aquas(t, dir, "", "install"); s.codigo != 0 {
		t.Fatalf("install: exit %d\n%s%s", s.codigo, s.out, s.err)
	}
	if _, err := os.Stat(claude); err != nil {
		t.Fatalf("install debió dejar CLAUDE.md: %v", err)
	}
	if err := os.Remove(claude); err != nil {
		t.Fatal(err)
	}
	if s := aquas(t, dir, "", "update"); s.codigo != 0 {
		t.Fatalf("update: exit %d\n%s%s", s.codigo, s.out, s.err)
	}
	if _, err := os.Stat(claude); err != nil {
		t.Errorf("update debió rearmar CLAUDE.md: %v", err)
	}
}

// TestInstallYUpdateArmanClaude (E002/F002/S001): install y update, sin ruta, crean huecos y arman CLAUDE.md.
// Rojo cuando: piden ruta, terminan sin CLAUDE.md o pisan un hueco escrito.
// Verde cuando: sin ruta arman CLAUDE.md, crean los huecos faltantes con PENDIENTE y respetan los escritos.
func TestInstallYUpdateArmanClaude(t *testing.T) {
	huecos := []string{"titulo", "mision", "mapa"}
	leer := func(p string) string {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	correr := func(dir, sub string) {
		t.Helper()
		if s := aquas(t, dir, "", sub); s.codigo != 0 {
			t.Fatalf("%s sin ruta debió salir 0: exit %d\n%s%s", sub, s.codigo, s.out, s.err)
		}
	}

	// Criterio 1: install en carpeta vacía crea huecos PENDIENTE y CLAUDE.md.
	dir := t.TempDir()
	hd := filepath.Join(dir, "gobierno", "canonico", "huecos")
	correr(dir, "install")
	for _, h := range huecos {
		if !strings.Contains(leer(filepath.Join(hd, h+".md")), "PENDIENTE") {
			t.Errorf("hueco %s debió nacer con PENDIENTE", h)
		}
	}
	if !strings.Contains(leer(filepath.Join(dir, "CLAUDE.md")), "PENDIENTE") {
		t.Error("CLAUDE.md debió armarse con los huecos PENDIENTE")
	}

	// Criterio 3: un hueco escrito queda igual; uno borrado se recrea.
	propia := "- Misión propia de la landing.\n"
	if err := os.WriteFile(filepath.Join(hd, "mision.md"), []byte(propia), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(hd, "titulo.md")); err != nil {
		t.Fatal(err)
	}
	correr(dir, "update")
	if leer(filepath.Join(hd, "mision.md")) != propia {
		t.Error("update no debió tocar un hueco escrito")
	}
	if !strings.Contains(leer(filepath.Join(hd, "titulo.md")), "PENDIENTE") {
		t.Error("update debió recrear el hueco faltante con PENDIENTE")
	}
	if !strings.Contains(leer(filepath.Join(dir, "CLAUDE.md")), "Misión propia de la landing.") {
		t.Error("CLAUDE.md debió llevar el hueco escrito")
	}

	// Criterio 2: hijo instalado sin huecos ni CLAUDE.md; update los crea y arma.
	if err := os.RemoveAll(filepath.Join(dir, "gobierno")); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(dir, "CLAUDE.md")); err != nil {
		t.Fatal(err)
	}
	correr(dir, "update")
	for _, h := range huecos {
		if _, err := os.Stat(filepath.Join(hd, h+".md")); err != nil {
			t.Errorf("update debió crear el hueco %s: %v", h, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "CLAUDE.md")); err != nil {
		t.Errorf("update debió armar CLAUDE.md: %v", err)
	}

	// Criterio 4: una ruta como argumento se rechaza con el uso correcto.
	otra := t.TempDir()
	for _, sub := range []string{"install", "update"} {
		s := aquas(t, dir, "", sub, otra)
		if s.codigo == 0 || !strings.Contains(s.err, "sin ruta") {
			t.Errorf("%s con ruta debió salir distinto de 0 con el uso «sin ruta»; exit %d, %q", sub, s.codigo, s.err)
		}
	}
	if es, _ := os.ReadDir(otra); len(es) > 0 {
		t.Error("con ruta no debió escribir nada en ella")
	}

	// Caso borde (fallo a medias): un archivo que no es hueco hace fallar build; sale distinto de 0,
	// pero la central nueva ya quedó copiada: el lock pasa de la versión marcada a la del binario.
	lock := filepath.Join(dir, ".aquas", "lock.json")
	var l map[string]any
	if err := json.Unmarshal([]byte(leer(lock)), &l); err != nil {
		t.Fatal(err)
	}
	l["version"] = "marcada-por-la-prueba"
	b, _ := json.Marshal(l)
	if err := os.WriteFile(lock, b, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hd, "otro.md"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if s := aquas(t, dir, "", "update"); s.codigo == 0 || !strings.Contains(s.err, "otro") {
		t.Errorf("update con un no-hueco debió fallar nombrándolo; exit %d, %q", s.codigo, s.err)
	}
	if strings.Contains(leer(lock), "marcada-por-la-prueba") {
		t.Error("tras el fallo de build la central nueva debió quedar copiada con su lock")
	}
}
