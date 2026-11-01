package herencia

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func escribir(t *testing.T, raiz, rel, contenido string) {
	t.Helper()
	p := filepath.Join(raiz, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(contenido), 0o644); err != nil {
		t.Fatal(err)
	}
}

// filas arma un herencia.json con una fila «todos» por control, más las filas extra dadas.
func filas(controles []string, extra ...string) string {
	var fs []string
	for _, c := range controles {
		fs = append(fs, `{"control": "`+c+`", "destino": "todos", "motivo": "m"}`)
	}
	fs = append(fs, extra...)
	return `{"version": 1, "controles": [` + strings.Join(fs, ", ") + `]}`
}

// TestCadaControlTieneFila (E003/F001/S002): cada control de la central tiene una fila válida en herencia.json.
// Rojo cuando: un control no tiene fila o una fila es inválida.
// Verde cuando: controles y filas coinciden, en la muestra y en el repo real.
func TestCadaControlTieneFila(t *testing.T) {
	raiz := t.TempDir()
	escribir(t, raiz, "gobierno/config/claude_settings.json", `{"hooks": [
		{"evento": "PreToolUse", "orden": "guard"},
		{"evento": "PreToolUse", "herramientas": ["Bash"], "orden": "guard"},
		{"evento": "Stop", "orden": "turn close"}]}`)
	escribir(t, raiz, ".githooks/pre-commit", "#!/bin/sh\n")
	controles, err := Controles(raiz)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"gancho:guard", "gancho:turn close", "git:pre-commit"}
	for _, c := range Chequeos {
		want = append(want, "gate:"+c)
	}
	if strings.Join(controles, ",") != strings.Join(want, ",") {
		t.Fatalf("controles: el gancho repetido cuenta una vez, más .githooks y Chequeos; esperado %v, dio %v", want, controles)
	}

	casos := []struct {
		nombre, json string
		quiere       []string // trozos que deben salir; vacío = sin problemas
	}{
		{"todo bien", filas(want), nil},
		{"control sin fila", filas(want[1:]), []string{"gancho:guard sin fila"}},
		{"fila sin control", filas(want, `{"control": "gancho:viejo", "destino": "todos", "motivo": "m"}`), []string{"gancho:viejo", "no existe"}},
		{"destino desconocido", filas(want[1:], `{"control": "gancho:guard", "destino": "tipos", "motivo": "m"}`), []string{"gancho:guard", "destino «tipos»"}},
		{"motivo vacío", filas(want[1:], `{"control": "gancho:guard", "destino": "central", "motivo": " "}`), []string{"gancho:guard", "motivo"}},
		{"fila repetida", filas(want, `{"control": "gancho:guard", "destino": "central", "motivo": "m"}`), []string{"gancho:guard", "repetida"}},
		{"json roto", `{"version": 1, "controles": [`, []string{"herencia.json ilegible"}},
		{"versión desconocida", `{"version": 2, "controles": []}`, []string{"versión"}},
	}
	for _, c := range casos {
		escribir(t, raiz, Ruta, c.json)
		got := strings.Join(Validar(raiz), "\n")
		if len(c.quiere) == 0 && got != "" {
			t.Errorf("%s: sin problemas; dio %q", c.nombre, got)
		}
		for _, q := range c.quiere {
			if !strings.Contains(got, q) {
				t.Errorf("%s: el problema debió decir %q; dio %q", c.nombre, q, got)
			}
		}
	}
	if err := os.Remove(filepath.Join(raiz, filepath.FromSlash(Ruta))); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(Validar(raiz), "\n"); !strings.Contains(got, "falta "+Ruta) {
		t.Errorf("archivo ausente: debió decir «falta %s»; dio %q", Ruta, got)
	}

	// Repo real: la central declara cada uno de sus controles.
	real, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	if ps := Validar(real); len(ps) > 0 {
		t.Errorf("herencia.json de la central: %v", ps)
	}
}
