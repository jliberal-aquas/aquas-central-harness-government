package memoria

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// E001/F005/S003: al cerrar el slice, la memoria se archiva sellada en registro/memoria/.

func focoCompleto(t *testing.T, raiz string) Memoria {
	t.Helper()
	return fijar(t, raiz, "iniciativa=E001", "feature=F005", "slice=S003", "tarea=archivar",
		"siguiente=cerrar", "pendiente=limpiar .tmp", "decision=¿gate?")
}

func TestRutaArchivo(t *testing.T) {
	if got := RutaArchivo(Foco{"E001", "F005", "S003", "x"}); got != "registro/memoria/E001-F005-S003.json" {
		t.Fatalf("ruta: %q", got)
	}
	if got := RutaArchivo(Foco{Iniciativa: "E001", Feature: "F005"}); got != "" {
		t.Fatalf("sin slice no hay ruta; dio %q", got)
	}
}

func TestArchivarCopiaSellada(t *testing.T) {
	raiz := repoGit(t)
	focoCompleto(t, raiz)
	caliente, _ := os.ReadFile(filepath.Join(raiz, Ruta))
	rel, err := Archivar(raiz, ahora)
	if err != nil {
		t.Fatal(err)
	}
	if rel != "registro/memoria/E001-F005-S003.json" {
		t.Fatalf("ruta devuelta: %q", rel)
	}
	b, err := os.ReadFile(filepath.Join(raiz, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b, caliente) {
		t.Fatal("el archivo debe guardar los mismos bytes sellados de la memoria")
	}
	m, err := LeerArchivo(filepath.Join(raiz, filepath.FromSlash(rel)))
	if err != nil || m.Foco.Tarea != "archivar" {
		t.Fatalf("el archivo debe leerse con su sello: %v %+v", err, m)
	}
}

func TestArchivarDejaFocoParaElSiguiente(t *testing.T) {
	raiz := repoGit(t)
	focoCompleto(t, raiz)
	if _, err := Archivar(raiz, ahora); err != nil {
		t.Fatal(err)
	}
	m, err := Leer(raiz)
	if err != nil {
		t.Fatalf("la memoria debe seguir sellada: %v", err)
	}
	if m.Foco != (Foco{Iniciativa: "E001", Feature: "F005"}) || m.Siguiente != "cerrar" {
		t.Fatalf("conserva E, F y siguiente; vacía slice y tarea: %+v", m)
	}
	if strings.Join(m.Pendientes, "|") != "limpiar .tmp" || strings.Join(m.Decisiones, "|") != "¿gate?" {
		t.Fatalf("pendientes y decisiones siguen: %+v", m)
	}
}

func TestArchivarRechaza(t *testing.T) {
	casos := []struct {
		nombre string
		armar  func(t *testing.T, raiz string)
		espera string
	}{
		{"sin memoria", func(t *testing.T, r string) {}, "sin memoria caliente"},
		{"sin slice", func(t *testing.T, r string) { fijar(t, r, "iniciativa=E001", "feature=F005") }, "slice"},
		{"sello roto", func(t *testing.T, r string) {
			focoCompleto(t, r)
			p := filepath.Join(r, Ruta)
			b, _ := os.ReadFile(p)
			if err := os.WriteFile(p, bytes.Replace(b, []byte("archivar"), []byte("trucada"), 1), 0o644); err != nil {
				t.Fatal(err)
			}
		}, "el sello no coincide"},
		{"memoria vieja", func(t *testing.T, r string) {
			focoCompleto(t, r)
			escribirArchivo(t, r, "a.txt", "cambio sin fijar\n")
		}, "memoria caliente vieja"},
		{"ya archivado", func(t *testing.T, r string) {
			focoCompleto(t, r)
			escribirArchivo(t, r, "registro/memoria/E001-F005-S003.json", "previo\n")
			fijar(t, r, "tarea=archivar") // al día con el archivo previo en disco
		}, "ya existe"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			raiz := repoGit(t)
			c.armar(t, raiz)
			antes, _ := os.ReadFile(filepath.Join(raiz, Ruta))
			_, err := Archivar(raiz, ahora)
			if err == nil || !strings.Contains(err.Error(), c.espera) {
				t.Fatalf("debe rechazar con %q; dio %v", c.espera, err)
			}
			despues, _ := os.ReadFile(filepath.Join(raiz, Ruta))
			if !bytes.Equal(antes, despues) {
				t.Fatal("un rechazo no toca la memoria")
			}
			if c.nombre != "ya archivado" {
				if _, err := os.Stat(filepath.Join(raiz, "registro")); !errors.Is(err, os.ErrNotExist) {
					t.Fatal("un rechazo no escribe el archivo")
				}
			}
		})
	}
}

func TestArchivoDetectaEdicion(t *testing.T) {
	raiz := repoGit(t)
	focoCompleto(t, raiz)
	rel, err := Archivar(raiz, ahora)
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(raiz, filepath.FromSlash(rel))
	b, _ := os.ReadFile(p)
	if err := os.WriteFile(p, bytes.Replace(b, []byte("archivar"), []byte("trucada"), 1), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LeerArchivo(p); !errors.Is(err, ErrSelloRoto) {
		t.Fatalf("archivo editado debe dar ErrSelloRoto; dio %v", err)
	}
}
