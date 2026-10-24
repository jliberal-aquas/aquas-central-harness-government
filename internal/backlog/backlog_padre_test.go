package backlog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// E001/F002/S017: un padre no cierra con hijos abiertos; el gate detecta al huérfano.

// fichaEn escribe una ficha en activo o archivo con los campos mínimos válidos.
func fichaEn(t *testing.T, raiz, dir string, f Ficha) {
	t.Helper()
	f.Tipo, f.Padre = tipoYPadre(f.ID)
	if f.BloqueadoPor == nil {
		f.BloqueadoPor = []string{}
	}
	f.Prioridad, f.Motivo = "P1", "m"
	b, _ := json.MarshalIndent(f, "", "  ")
	poner(t, raiz, dir+"/"+nombre(f.ID), string(b)+"\n")
}

func regPadre() *Registro {
	return &Registro{Fecha: "2026-10-23T10:00:00Z", Resultado: map[string]string{"./p:TestP": "falló"}}
}

// listoParaCerrar: padre con pruebas, rojo y verde (cierre propio ya satisfecho).
func listoParaCerrar(id string, estado string, sec int) Ficha {
	return Ficha{ID: id, Titulo: "T", Estado: estado, Secuencia: sec,
		Pruebas: []Prueba{{Nombre: "./p:TestP", Intencion: "x", RojoCuando: "r", VerdeCuando: "v"}}, Rojo: regPadre(), Verde: regPadre()}
}

func bytesDe(t *testing.T, ruta string) string {
	t.Helper()
	b, err := os.ReadFile(ruta)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestPadreNoCierraConHijosAbiertos(t *testing.T) {
	for _, destino := range []string{"CERRADO", "CANCELADO"} {
		for _, origen := range []string{"BORRADOR", "EN_CURSO", "BLOQUEADO"} {
			for _, hijoEstado := range []string{"EN_CURSO", "BORRADOR", "BLOQUEADO"} {
				raiz := t.TempDir()
				fichaEn(t, raiz, Activo, Ficha{ID: "E001", Titulo: "I", Estado: "EN_CURSO", Secuencia: 1})
				fichaEn(t, raiz, Activo, listoParaCerrar("E001/F002", origen, 2))
				fichaEn(t, raiz, Activo, Ficha{ID: "E001/F002/S001", Titulo: "S", Estado: hijoEstado, Secuencia: 3})
				antes := bytesDe(t, filepath.Join(raiz, Activo, "E001-F002.json"))
				_, err := Estado(raiz, "E001/F002", destino)
				if err == nil || !strings.Contains(err.Error(), "E001/F002/S001") || !strings.Contains(err.Error(), hijoEstado) {
					t.Errorf("%s→%s con slice %s: debe rechazar nombrando hijo y estado; dio %v", origen, destino, hijoEstado, err)
				}
				if bytesDe(t, filepath.Join(raiz, Activo, "E001-F002.json")) != antes {
					t.Errorf("%s→%s: el rechazo no debe escribir la ficha", origen, destino)
				}
				if _, e := os.Stat(filepath.Join(raiz, Archivo, "E001-F002.json")); !os.IsNotExist(e) {
					t.Errorf("%s→%s: el rechazo no debe archivar", origen, destino)
				}
			}
		}
	}

	t.Run("iniciativa con feature activa", func(t *testing.T) {
		for _, destino := range []string{"CERRADO", "CANCELADO"} {
			raiz := t.TempDir()
			fichaEn(t, raiz, Activo, listoParaCerrar("E001", "EN_CURSO", 1))
			fichaEn(t, raiz, Activo, Ficha{ID: "E001/F002", Titulo: "F", Estado: "BORRADOR", Secuencia: 2})
			if _, err := Estado(raiz, "E001", destino); err == nil || !strings.Contains(err.Error(), "E001/F002") {
				t.Errorf("%s: debe rechazar nombrando la feature; dio %v", destino, err)
			}
		}
	})

	t.Run("mensaje nombra todos los hijos activos", func(t *testing.T) {
		raiz := t.TempDir()
		fichaEn(t, raiz, Activo, listoParaCerrar("E001/F002", "EN_CURSO", 1))
		fichaEn(t, raiz, Activo, Ficha{ID: "E001/F002/S001", Titulo: "S", Estado: "EN_CURSO", Secuencia: 2})
		fichaEn(t, raiz, Activo, Ficha{ID: "E001/F002/S002", Titulo: "S", Estado: "BLOQUEADO", Secuencia: 3})
		_, err := Estado(raiz, "E001/F002", "CERRADO")
		if err == nil || !strings.Contains(err.Error(), "E001/F002/S001") || !strings.Contains(err.Error(), "E001/F002/S002") {
			t.Errorf("debe nombrar ambos hijos; dio %v", err)
		}
	})

	t.Run("hijos todos terminales pasa", func(t *testing.T) {
		for _, destino := range []string{"CERRADO", "CANCELADO"} {
			raiz := t.TempDir()
			fichaEn(t, raiz, Activo, listoParaCerrar("E001/F002", "EN_CURSO", 1))
			fichaEn(t, raiz, Archivo, Ficha{ID: "E001/F002/S001", Titulo: "S", Estado: "CERRADO", Secuencia: 2})
			fichaEn(t, raiz, Archivo, Ficha{ID: "E001/F002/S002", Titulo: "S", Estado: "CANCELADO", Secuencia: 3})
			if _, err := Estado(raiz, "E001/F002", destino); err != nil {
				t.Errorf("%s con hijos terminales debe pasar; dio %v", destino, err)
			}
			if _, e := os.Stat(filepath.Join(raiz, Archivo, "E001-F002.json")); e != nil {
				t.Errorf("%s: debe quedar archivada: %v", destino, e)
			}
		}
	})

	t.Run("padre sin hijos cierra", func(t *testing.T) {
		raiz := t.TempDir()
		fichaEn(t, raiz, Activo, listoParaCerrar("E001/F002", "EN_CURSO", 1))
		if _, err := Estado(raiz, "E001/F002", "CERRADO"); err != nil {
			t.Errorf("sin hijos cierra como hoy; dio %v", err)
		}
	})

	t.Run("hijo de otro padre no cuenta", func(t *testing.T) {
		raiz := t.TempDir()
		fichaEn(t, raiz, Activo, listoParaCerrar("E001/F002", "EN_CURSO", 1))
		fichaEn(t, raiz, Activo, Ficha{ID: "E001/F003/S001", Titulo: "S", Estado: "EN_CURSO", Secuencia: 2})
		if _, err := Estado(raiz, "E001/F002", "CANCELADO"); err != nil {
			t.Errorf("el hijo de otra feature no bloquea; dio %v", err)
		}
	})

	t.Run("ficha hija ilegible es error y no cierra", func(t *testing.T) {
		raiz := t.TempDir()
		fichaEn(t, raiz, Activo, listoParaCerrar("E001/F002", "EN_CURSO", 1))
		poner(t, raiz, Activo+"/E001-F002-S001.json", "{no es json")
		if _, err := Estado(raiz, "E001/F002", "CANCELADO"); err == nil || !strings.Contains(err.Error(), "ilegible") {
			t.Errorf("ficha ilegible debe ser error; dio %v", err)
		}
		if _, e := os.Stat(filepath.Join(raiz, Archivo, "E001-F002.json")); !os.IsNotExist(e) {
			t.Error("no debe archivar con una ficha ilegible")
		}
	})

	t.Run("slice sin hijos cancela como hoy", func(t *testing.T) {
		raiz := t.TempDir()
		fichaEn(t, raiz, Activo, Ficha{ID: "E001/F002/S001", Titulo: "S", Estado: "BORRADOR", Secuencia: 1})
		if _, err := Estado(raiz, "E001/F002/S001", "CANCELADO"); err != nil {
			t.Errorf("slice cancela como hoy; dio %v", err)
		}
	})
}

func TestGateDetectaHijoHuerfano(t *testing.T) {
	huerfanos := func(p []string) []string {
		var r []string
		for _, m := range p {
			if strings.Contains(m, "archivado") {
				r = append(r, m)
			}
		}
		return r
	}
	t.Run("hijo activo con padre archivado falla", func(t *testing.T) {
		raiz := t.TempDir()
		fichaEn(t, raiz, Activo, Ficha{ID: "E001", Titulo: "I", Estado: "EN_CURSO", Secuencia: 1})
		fichaEn(t, raiz, Archivo, Ficha{ID: "E001/F003", Titulo: "F", Estado: "CANCELADO", Secuencia: 2})
		fichaEn(t, raiz, Activo, Ficha{ID: "E001/F003/S003", Titulo: "S", Estado: "BORRADOR", Secuencia: 3})
		h := huerfanos(Validar(raiz))
		if len(h) != 1 || !strings.Contains(h[0], "E001-F003-S003.json") || !strings.Contains(h[0], "padre E001/F003") {
			t.Errorf("el gate debe nombrar hijo y padre; dio %v", h)
		}
	})
	t.Run("feature archivada con iniciativa activa no es huérfana", func(t *testing.T) {
		raiz := t.TempDir()
		fichaEn(t, raiz, Activo, Ficha{ID: "E001", Titulo: "I", Estado: "EN_CURSO", Secuencia: 1})
		fichaEn(t, raiz, Archivo, Ficha{ID: "E001/F003", Titulo: "F", Estado: "CERRADO", Secuencia: 2})
		fichaEn(t, raiz, Archivo, Ficha{ID: "E001/F003/S003", Titulo: "S", Estado: "CERRADO", Secuencia: 3})
		if h := huerfanos(Validar(raiz)); len(h) != 0 {
			t.Errorf("hijo y padre archivados: sin problema; dio %v", h)
		}
	})
	t.Run("hijo activo con padre activo no es huérfano", func(t *testing.T) {
		raiz := t.TempDir()
		fichaEn(t, raiz, Activo, Ficha{ID: "E001", Titulo: "I", Estado: "EN_CURSO", Secuencia: 1})
		fichaEn(t, raiz, Activo, Ficha{ID: "E001/F003", Titulo: "F", Estado: "EN_CURSO", Secuencia: 2})
		if h := huerfanos(Validar(raiz)); len(h) != 0 {
			t.Errorf("sin huérfanos; dio %v", h)
		}
	})
	t.Run("padre inexistente lo reporta la revisión de padre, no la de huérfano", func(t *testing.T) {
		raiz := t.TempDir()
		fichaEn(t, raiz, Activo, Ficha{ID: "E001/F003/S003", Titulo: "S", Estado: "BORRADOR", Secuencia: 1})
		p := Validar(raiz)
		if h := huerfanos(p); len(h) != 0 {
			t.Errorf("no es huérfano archivado; dio %v", h)
		}
		if !strings.Contains(strings.Join(p, "\n"), "no existe") {
			t.Errorf("debe seguir reportando padre inexistente; dio %v", p)
		}
	})
}
