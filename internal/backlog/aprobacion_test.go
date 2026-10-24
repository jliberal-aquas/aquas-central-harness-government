package backlog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// E001/F010/S001: rojo y EN_CURSO exigen la aprobación humana del contrato, atada a su huella.

const paginaContrato = `# S002

## Para qué existe

El humano aprueba lo que se va a construir.

<!-- GENERADO por aquas desde trabajo/: no editar a mano. -->

| Prueba | Rojo | Verde |
|---|---|---|
| ` + "`TestFalla`" + ` | — | — |

<!-- /GENERADO -->

## Cómo se usa

Texto fuera del contrato.
`

// conExigencia: raíz con la página del slice S002 y el slice S002 en BORRADOR con alcance.
func conExigencia(t *testing.T) string {
	t.Helper()
	raiz := modulo(t)
	esc := func(rel, txt string) {
		ruta := filepath.Join(raiz, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(ruta, []byte(txt), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	esc("docs/E001/F002/S002.md", paginaContrato)
	nuevo(t, raiz, "id=E001/F002/S002", "titulo=T", "prioridad=P1", "motivo=m")
	if _, err := Alcance(raiz, "E001/F002/S002", []string{"p/**"}); err != nil {
		t.Fatal(err)
	}
	return raiz
}

func editarPagina(t *testing.T, raiz, antes, despues string) {
	t.Helper()
	ruta := filepath.Join(raiz, "docs", "E001", "F002", "S002.md")
	b, err := os.ReadFile(ruta)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), antes) {
		t.Fatalf("la página no contiene %q", antes)
	}
	if err := os.WriteFile(ruta, []byte(strings.Replace(string(b), antes, despues, 1)), 0o644); err != nil {
		t.Fatal(err)
	}
}

// Rojo cuando: un slice sin aprobación pasa a EN_CURSO o registra rojo.
// Verde cuando: se rechaza nombrando «apruebo» y el id; con la aprobación vigente se acepta.
func TestRojoExigeAprobacion(t *testing.T) {
	const id = "E001/F002/S002"
	raiz := conExigencia(t)
	_, err := Estado(raiz, id, "EN_CURSO")
	if err == nil || !strings.Contains(err.Error(), "apruebo "+id) {
		t.Fatalf("EN_CURSO sin aprobación: error que pide «apruebo %s»; dio %v", id, err)
	}
	if _, err := Rojo(raiz, id, hora); err == nil || !strings.Contains(err.Error(), "apruebo "+id) {
		t.Fatalf("rojo sin aprobación: error que pide «apruebo %s»; dio %v", id, err)
	}
	if _, err := Aprobar(raiz, id, "apruebo "+id, hora); err != nil {
		t.Fatal(err)
	}
	if _, err := Estado(raiz, id, "EN_CURSO"); err != nil {
		t.Fatalf("EN_CURSO con aprobación vigente debe aceptarse: %v", err)
	}
	if _, err := Rojo(raiz, id, hora); err != nil {
		t.Fatalf("rojo con aprobación vigente debe aceptarse: %v", err)
	}
}

// Rojo cuando: cambian las pruebas, el alcance o «Para qué existe» y el rojo se acepta con la aprobación vieja.
// Verde cuando: se rechaza pidiendo aprobar de nuevo; lo generado por aquas no la anula.
func TestAprobacionAtadaAHuella(t *testing.T) {
	const id = "E001/F002/S002"
	cambios := map[string]func(t *testing.T, raiz string){
		"para qué existe": func(t *testing.T, raiz string) {
			editarPagina(t, raiz, "lo que se va a construir", "otra cosa distinta")
		},
		"alcance": func(t *testing.T, raiz string) {
			if _, err := Alcance(raiz, id, []string{"p/**", "q/**"}); err != nil {
				t.Fatal(err)
			}
		},
		"pruebas": func(t *testing.T, raiz string) {
			if _, err := Pruebas(raiz, id, []string{"./p:TestFalla=otra intención|r|v"}); err != nil {
				t.Fatal(err)
			}
		},
	}
	for nombre, cambiar := range cambios {
		raiz := conExigencia(t)
		if _, err := Aprobar(raiz, id, "apruebo "+id, hora); err != nil {
			t.Fatal(err)
		}
		editarPagina(t, raiz, "| `TestFalla` | — | — |", "| `TestFalla` | 2026-10-23 | — |") // bloque generado: no cuenta
		if _, err := Rojo(raiz, id, hora); err != nil && strings.Contains(err.Error(), "apruebo") {
			t.Fatalf("%s: cambiar solo lo generado no anula la aprobación: %v", nombre, err)
		}
		cambiar(t, raiz)
		if _, err := Rojo(raiz, id, hora); err == nil || !strings.Contains(err.Error(), "apruebo "+id) {
			t.Errorf("cambio de %s tras aprobar: el rojo se rechaza y pide «apruebo %s»; dio %v", nombre, id, err)
		}
		if _, err := Estado(raiz, id, "EN_CURSO"); err == nil || !strings.Contains(err.Error(), "apruebo "+id) {
			t.Errorf("cambio de %s tras aprobar: EN_CURSO se rechaza y pide «apruebo %s»; dio %v", nombre, id, err)
		}
		if _, err := Aprobar(raiz, id, "apruebo "+id, hora); err != nil {
			t.Fatal(err)
		}
		if _, err := Estado(raiz, id, "EN_CURSO"); err != nil {
			t.Errorf("cambio de %s y aprobación nueva: EN_CURSO se acepta; dio %v", nombre, err)
		}
	}
}

// Rojo cuando: un slice BORRADOR, LISTO, BLOQUEADO o EN_CURSO sin rojo ni aprobación arranca o registra rojo sin ella; o uno con rojo previo y sin aprobación (anterior a F010) pide «apruebo».
// Verde cuando: todos esos estados piden «apruebo» (el desvío por BLOQUEADO también); el slice con rojo previo sigue exento.
func TestExigenciaPorEstado(t *testing.T) {
	const id = "E001/F002/S002"
	raiz := conExigencia(t) // BORRADOR sin aprobación
	if _, err := Rojo(raiz, id, hora); err == nil || !strings.Contains(err.Error(), "apruebo "+id) {
		t.Fatalf("rojo en BORRADOR sin aprobación: pide «apruebo %s»; dio %v", id, err)
	}
	f, _, err := Ver(raiz, id)
	if err != nil {
		t.Fatal(err)
	}
	for _, desde := range []string{"LISTO", "BLOQUEADO"} {
		f.Estado = desde
		pisar(t, raiz, Activo, f)
		if _, err := Estado(raiz, id, "EN_CURSO"); err == nil || !strings.Contains(err.Error(), "apruebo "+id) {
			t.Fatalf("%s→EN_CURSO sin aprobación: pide «apruebo %s»; dio %v", desde, id, err)
		}
	}
	f.Estado = "EN_CURSO"
	pisar(t, raiz, Activo, f)
	if _, err := Rojo(raiz, id, hora); err == nil || !strings.Contains(err.Error(), "apruebo "+id) {
		t.Fatalf("EN_CURSO sin rojo ni aprobación: el rojo pide «apruebo %s»; dio %v", id, err)
	}
	f.Rojo = &Registro{Fecha: "2026-10-01", Head: "abc"} // slice anterior a F010: rojo previo, sin aprobación
	pisar(t, raiz, Activo, f)
	if _, err := Verde(raiz, id, hora); err != nil && strings.Contains(err.Error(), "apruebo") {
		t.Fatalf("slice con rojo previo y sin aprobación: exento; dio %v", err)
	}
}

// Rojo cuando: tras aprobar y arrancar, cambiar «Para qué existe», las pruebas o ampliar el alcance deja pasar el rojo, el verde o el cierre.
// Verde cuando: en EN_CURSO los tres se rechazan pidiendo «apruebo» con el id; sin cambios, el cierre no pide apruebo.
func TestDerivaEnCursoAnulaAprobacion(t *testing.T) {
	const id = "E001/F002/S002"
	cambios := map[string]func(t *testing.T, raiz string){
		"para qué existe": func(t *testing.T, raiz string) {
			editarPagina(t, raiz, "lo que se va a construir", "otra cosa distinta")
		},
		"pruebas": func(t *testing.T, raiz string) {
			if _, err := Pruebas(raiz, id, []string{"./p:TestFalla=otra intención|r|v"}); err != nil {
				t.Fatal(err)
			}
		},
		"alcance ampliado": func(t *testing.T, raiz string) {
			if _, err := Ampliar(raiz, id, "q/**", "hace falta q"); err != nil {
				t.Fatal(err)
			}
		},
	}
	pide := func(err error) bool { return err != nil && strings.Contains(err.Error(), "apruebo "+id) }
	for nombre, cambiar := range cambios {
		raiz := conExigencia(t)
		if _, err := Pruebas(raiz, id, []string{"./p:TestFalla=intención|r|v"}); err != nil {
			t.Fatal(err)
		}
		if _, err := Aprobar(raiz, id, "apruebo "+id, hora); err != nil {
			t.Fatal(err)
		}
		if _, err := Estado(raiz, id, "EN_CURSO"); err != nil {
			t.Fatal(err)
		}
		f, _, _ := Ver(raiz, id)
		f.Rojo = &Registro{Fecha: "2026-10-23", Head: "abc"}
		f.Verde = &Registro{Fecha: "2026-10-23", Head: "abc"}
		pisar(t, raiz, Activo, f)
		cambiar(t, raiz)
		if _, err := Rojo(raiz, id, hora); !pide(err) {
			t.Errorf("deriva por %s en EN_CURSO: el rojo pide «apruebo %s»; dio %v", nombre, id, err)
		}
		if _, err := Verde(raiz, id, hora); !pide(err) {
			t.Errorf("deriva por %s en EN_CURSO: el verde pide «apruebo %s»; dio %v", nombre, id, err)
		}
		if _, err := Estado(raiz, id, "CERRADO"); !pide(err) {
			t.Errorf("deriva por %s en EN_CURSO: el cierre pide «apruebo %s»; dio %v", nombre, id, err)
		}
	}
}

// Rojo cuando: ampliar el alcance tras aprobar deja vigente la aprobación.
// Verde cuando: Ampliar anula la aprobación; EN_CURSO y el rojo se rechazan pidiendo «apruebo» con el id.
func TestAmpliarAnulaAprobacion(t *testing.T) {
	const id = "E001/F002/S002"
	raiz := conExigencia(t)
	if _, err := Aprobar(raiz, id, "apruebo "+id, hora); err != nil {
		t.Fatal(err)
	}
	if _, err := Ampliar(raiz, id, "q/**", "hace falta q"); err != nil {
		t.Fatal(err)
	}
	if _, err := Estado(raiz, id, "EN_CURSO"); err == nil || !strings.Contains(err.Error(), "apruebo "+id) {
		t.Fatalf("ampliar tras aprobar: EN_CURSO se rechaza y pide «apruebo %s»; dio %v", id, err)
	}
	if _, err := Rojo(raiz, id, hora); err == nil || !strings.Contains(err.Error(), "apruebo "+id) {
		t.Fatalf("ampliar tras aprobar: el rojo se rechaza y pide «apruebo %s»; dio %v", id, err)
	}
}

// arrancar: pasa un slice a EN_CURSO como lo haría el flujo real, con la página del contrato y la aprobación humana ya sembradas.
// Lo que no es slice, un slice ya EN_CURSO o uno sin ficha activa va directo a Estado.
func arrancar(t *testing.T, raiz, id string) (Ficha, error) {
	t.Helper()
	f, archivada, err := Ver(raiz, id)
	if err == nil && !archivada && f.Tipo == "slice" && f.Estado != "EN_CURSO" {
		sembrarAprobacion(t, raiz, id)
	}
	return Estado(raiz, id, "EN_CURSO")
}

// sembrarAprobacion: deja la página del contrato y la aprobación humana del slice, como el flujo real antes del rojo.
func sembrarAprobacion(t *testing.T, raiz, id string) {
	t.Helper()
	ruta := filepath.Join(raiz, filepath.FromSlash(pagina(id)))
	if _, err := os.Stat(ruta); err != nil {
		if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(ruta, []byte("# "+id+"\n\n## Para qué existe\n\nContrato de prueba.\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := Aprobar(raiz, id, "apruebo "+id, hora); err != nil {
		t.Fatal(err)
	}
}
