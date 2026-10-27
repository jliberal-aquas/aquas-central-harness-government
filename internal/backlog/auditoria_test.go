package backlog

// E001/F014/S002: CERRADO exige el último veredicto APROBADO, posterior al verde y sobre el mismo código.

import (
	"strings"
	"testing"
	"time"
)

// aprobar anota un veredicto del auditor en la hora dada (mismo camino que el gancho cierre-agente).
func aprobar(t *testing.T, raiz, id, veredicto string, ahora time.Time) {
	t.Helper()
	if err := AnotarAuditoria(raiz, id, veredicto, ahora); err != nil {
		t.Fatalf("anotar %s: %v", veredicto, err)
	}
}

// cerrarDa exige que CERRADO falle con un mensaje que contenga «veredicto» y el texto de su causa propia (D000052).
func cerrarDa(t *testing.T, raiz, id, causa string) {
	t.Helper()
	_, err := Estado(raiz, id, "CERRADO")
	if err == nil || !strings.Contains(err.Error(), "veredicto") || !strings.Contains(err.Error(), causa) {
		t.Errorf("cerrar: se esperaba rechazo por veredicto con la causa %q; dio %v", causa, err)
	}
}

// Rojo cuando: cierra sin veredicto, con CON_HALLAZGOS o RECHAZADO como último, con APROBADO anterior al verde o con otra huella.
// Verde cuando: solo cierra con APROBADO posterior al verde y huella igual; el último veredicto manda.
func TestCierreExigeAuditorAprobado(t *testing.T) {
	t.Run("sin veredicto rechaza y pide auditar", func(t *testing.T) {
		raiz, id := sliceConVerde(t)
		cerrarDa(t, raiz, id, "no hay veredicto")
	})
	t.Run("CON_HALLAZGOS y RECHAZADO rechazan nombrando el veredicto", func(t *testing.T) {
		for _, v := range []string{"CON_HALLAZGOS", "RECHAZADO"} {
			raiz, id := sliceConVerde(t)
			aprobar(t, raiz, id, v, hora.Add(2*time.Hour))
			cerrarDa(t, raiz, id, "último veredicto es "+v)
		}
	})
	t.Run("APROBADO anterior al verde rechaza", func(t *testing.T) {
		raiz, id := sliceConVerde(t)
		aprobar(t, raiz, id, "APROBADO", hora.Add(30*time.Minute))
		cerrarDa(t, raiz, id, "anterior al verde")
	})
	t.Run("APROBADO a la misma hora del verde rechaza", func(t *testing.T) {
		raiz, id := sliceConVerde(t)
		aprobar(t, raiz, id, "APROBADO", hora.Add(time.Hour))
		cerrarDa(t, raiz, id, "anterior al verde")
	})
	t.Run("APROBADO con otra huella rechaza aunque el código volvió", func(t *testing.T) {
		raiz, id := sliceConVerde(t)
		escribirEn(t, raiz, "p/p.go", "package p\n\nconst V = 2\n")
		aprobar(t, raiz, id, "APROBADO", hora.Add(2*time.Hour))     // huella de otro código
		escribirEn(t, raiz, "p/p.go", "package p\n\nconst V = 1\n") // el código vuelve al del verde
		cerrarDa(t, raiz, id, "otro código")
	})
	t.Run("fecha del verde ilegible rechaza con causa propia", func(t *testing.T) {
		raiz, id := sliceConVerde(t)
		aprobar(t, raiz, id, "APROBADO", hora.Add(2*time.Hour))
		f := leerFicha(t, activo(raiz, nombre(id)))
		f.Verde = &Registro{Fecha: "no-es-fecha", Head: f.Verde.Head, Huella: f.Verde.Huella}
		pisar(t, raiz, Activo, f)
		cerrarDa(t, raiz, id, "fecha del verde ilegible")
	})
	t.Run("verde repetido tras el APROBADO deja el APROBADO viejo", func(t *testing.T) {
		raiz, id := sliceConVerde(t)
		aprobar(t, raiz, id, "APROBADO", hora.Add(2*time.Hour))
		if _, err := Verde(raiz, id, hora.Add(3*time.Hour)); err != nil {
			t.Fatalf("verde: %v", err)
		}
		cerrarDa(t, raiz, id, "anterior al verde")
	})
	t.Run("APROBADO posterior y misma huella cierra", func(t *testing.T) {
		raiz, id := sliceConVerde(t)
		aprobar(t, raiz, id, "APROBADO", hora.Add(2*time.Hour))
		if _, err := Estado(raiz, id, "CERRADO"); err != nil {
			t.Errorf("APROBADO vigente: debe cerrar; dio %v", err)
		}
	})
	t.Run("gana el último veredicto", func(t *testing.T) {
		raiz, id := sliceConVerde(t)
		aprobar(t, raiz, id, "CON_HALLAZGOS", hora.Add(2*time.Hour))
		aprobar(t, raiz, id, "APROBADO", hora.Add(3*time.Hour))
		if _, err := Estado(raiz, id, "CERRADO"); err != nil {
			t.Errorf("CON_HALLAZGOS y luego APROBADO: debe cerrar; dio %v", err)
		}
		raiz, id = sliceConVerde(t)
		aprobar(t, raiz, id, "APROBADO", hora.Add(2*time.Hour))
		aprobar(t, raiz, id, "RECHAZADO", hora.Add(3*time.Hour))
		cerrarDa(t, raiz, id, "último veredicto es RECHAZADO")
	})
	t.Run("verde viejo sin huella: exige APROBADO posterior y no compara huella", func(t *testing.T) {
		raiz, id := sliceConVerde(t)
		f := leerFicha(t, activo(raiz, nombre(id)))
		f.Verde = &Registro{Fecha: f.Verde.Fecha, Head: f.Verde.Head}
		pisar(t, raiz, Activo, f)
		cerrarDa(t, raiz, id, "no hay veredicto")
		aprobar(t, raiz, id, "APROBADO", hora.Add(2*time.Hour))
		escribirEn(t, raiz, "p/p.go", "package p\n\nconst V = 7\n") // huella distinta de la anotada: no se compara
		if _, err := Estado(raiz, id, "CERRADO"); err != nil {
			t.Errorf("verde sin huella con APROBADO posterior: debe cerrar; dio %v", err)
		}
	})
	t.Run("anotar solo en un slice en curso, nombrando el id", func(t *testing.T) {
		raiz := modulo(t)
		nuevo(t, raiz, "id=E001/F002/S005", "titulo=T", "prioridad=P1", "motivo=m", "prueba=./p:TestBandera=a|r|v")
		for _, id := range []string{"E001/F002/S005", "E001/F002/S099"} { // borrador; inexistente
			if err := AnotarAuditoria(raiz, id, "APROBADO", hora); err == nil || !strings.Contains(err.Error(), id) {
				t.Errorf("anotar en %s: se esperaba error que nombre el id; dio %v", id, err)
			}
		}
	})
	t.Run("anotar guarda fecha, veredicto y huella en la ficha", func(t *testing.T) {
		raiz, id := sliceConVerde(t)
		aprobar(t, raiz, id, "CON_HALLAZGOS", hora.Add(2*time.Hour))
		a := leerFicha(t, activo(raiz, nombre(id))).Auditoria
		if a == nil || a.Fecha != "2026-10-21T14:00:00Z" || a.Veredicto != "CON_HALLAZGOS" || a.Huella == nil || (*a.Huella)["p/p.go"] == "" {
			t.Errorf("auditoría anotada: se esperaba fecha 2026-10-21T14:00:00Z, CON_HALLAZGOS y huella de p/p.go; dio %+v", a)
		}
	})
}
