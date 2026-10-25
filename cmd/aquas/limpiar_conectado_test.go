package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestLimpiarConectado (E000/F006/S011): el binario despacha `limpiar` y la regla de encargos lo nombra.
// Rojo cuando: el despacho falta o la regla no nombra `aquas limpiar`.
// Verde cuando: el binario borra por subcomando, rechaza un id inválido y la regla lo nombra.
func TestLimpiarConectado(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "aquas.exe")
	if out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	raiz := t.TempDir()
	dir := filepath.Join(raiz, ".tmp", "E000-F006-S010-A1")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	correr := func(id string) error {
		c := exec.Command(bin, "limpiar", id)
		c.Dir = raiz
		return c.Run()
	}
	// Criterio 2: id inválido → salida distinta de 0, nada borrado.
	if err := correr("s008r"); err == nil {
		t.Error("id inválido debió salir distinto de 0")
	}
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("id inválido no debió borrar: %v", err)
	}
	// Criterio 1: id válido → salida 0 y carpeta borrada.
	if err := correr("E000/F006/S010-A1"); err != nil {
		t.Errorf("id válido debió salir 0: %v", err)
	}
	if _, err := os.Stat(dir); err == nil {
		t.Error("la carpeta del encargo debió desaparecer")
	}
	// Caso borde: carpeta inexistente → salida 0 con aviso por stdout.
	c := exec.Command(bin, "limpiar", "E000/F006/S010-A1")
	c.Dir = raiz
	if out, err := c.Output(); err != nil || !strings.HasPrefix(string(out), "nada que borrar") {
		t.Errorf("sin carpeta debió salir 0 con «nada que borrar»; dio %q, %v", out, err)
	}
	// Criterio 3: la regla canónica de encargos nombra la herramienta.
	b, err := os.ReadFile(filepath.Join("..", "..", "gobierno", "canonico", "reglas", "02_COMUNICACION_ENTRE_AGENTES.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "aquas limpiar") {
		t.Error("la regla de encargos debió nombrar `aquas limpiar`")
	}
}
