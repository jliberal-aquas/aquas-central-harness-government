package guardia

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Lista de permitidos (gobierno/config/guardia.json): rutas fuera del repo donde sí se puede
// escribir, cada una con su motivo. Lo que no está en la lista se bloquea.

func conPermitidos(t *testing.T, contenido string) (raiz, fuera string) {
	t.Helper()
	raiz = raizPrueba(t)
	fuera = filepath.Join(filepath.Dir(raiz), "cache")
	p := filepath.Join(raiz, "gobierno", "config", "guardia.json")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(contenido), 0o644); err != nil {
		t.Fatal(err)
	}
	return raiz, fuera
}

func TestPermitidoFueraPasa(t *testing.T) {
	t.Setenv("AQUAS_PRUEBA", "")
	raiz, fuera := conPermitidos(t, "")
	t.Setenv("AQUAS_PRUEBA", filepath.Dir(raiz))
	conPermitidosEn(t, raiz, `{"permitidos_fuera":[{"ruta":"${AQUAS_PRUEBA}/cache","motivo":"caché de prueba"}]}`)
	if m := RevisarArchivo(raiz, filepath.Join(fuera, "x.bin")); m != "" {
		t.Fatalf("ruta permitida debe pasar: %s", m)
	}
	if m := RevisarBash(raiz, raiz, `rm -rf "`+filepath.ToSlash(fuera)+`/viejo"`); m != "" {
		t.Fatalf("bash a ruta permitida debe pasar: %s", m)
	}
	if p := ValidarConfig(raiz); len(p) != 0 {
		t.Fatalf("config válida: %q", p)
	}
}

func TestNoPermitidoSigueBloqueado(t *testing.T) {
	raiz, fuera := conPermitidos(t, "")
	conPermitidosEn(t, raiz, `{"permitidos_fuera":[{"ruta":"`+filepath.ToSlash(fuera)+`","motivo":"caché"}]}`)
	for _, p := range []string{
		filepath.Join(fuera+"-gemelo", "x"),          // prefijo igual, carpeta distinta
		filepath.Join(filepath.Dir(raiz), "otra", "x"), // hermana no listada
	} {
		if m := RevisarArchivo(raiz, p); !strings.Contains(m, "fuera del repositorio") {
			t.Errorf("%s no está en la lista; debe fallar: %q", p, m)
		}
	}
}

func TestSinConfigTodoFueraBloqueado(t *testing.T) {
	raiz := raizPrueba(t)
	// E000/F005/S002: bloquea por ruta fuera, y nombra la ruta.
	p := filepath.Join(filepath.Dir(raiz), "cache", "x")
	if m := RevisarArchivo(raiz, p); !strings.Contains(m, "escritura fuera del repositorio bloqueada: "+p) {
		t.Fatalf("sin lista de permitidos, fuera debe fallar nombrando %s; obtuve %q", p, m)
	}
}

func TestConfigInvalida(t *testing.T) {
	casos := []struct{ nombre, json, espera string }{
		{"json roto", `{`, "gobierno/config/guardia.json: ilegible"},
		{"sin motivo", `{"permitidos_fuera":[{"ruta":"C:/x"}]}`, "sin motivo"},
		{"ruta vacía", `{"permitidos_fuera":[{"ruta":"","motivo":"m"}]}`, "ruta vacía"},
		{"ruta relativa", `{"permitidos_fuera":[{"ruta":"cache","motivo":"m"}]}`, "no es absoluta"},
		{"permite todo", `{"permitidos_fuera":[{"ruta":"C:/","motivo":"m"}]}`, "raíz de unidad"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			raiz, fuera := conPermitidos(t, c.json)
			p := ValidarConfig(raiz)
			if !strings.Contains(strings.Join(p, "\n"), c.espera) {
				t.Fatalf("se esperaba %q; hubo %q", c.espera, p)
			}
			if c.nombre == "json roto" && RevisarArchivo(raiz, filepath.Join(fuera, "x")) == "" {
				t.Fatal("config ilegible no debe permitir nada fuera")
			}
		})
	}
}

func conPermitidosEn(t *testing.T, raiz, contenido string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(raiz, "gobierno", "config", "guardia.json"), []byte(contenido), 0o644); err != nil {
		t.Fatal(err)
	}
}
