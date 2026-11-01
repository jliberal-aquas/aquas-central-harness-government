package backlog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSliceNuevoPideFeature (E000/F008/S006, D000084): un slice fuera de la foto de una feature aprobada pide
// «apruebo» de la feature, no el suyo; backlog new lo avisa al crearlo.
// Rojo cuando: el rechazo pide «apruebo» del slice.
// Verde cuando: pide «apruebo» de la feature y el aviso al crear lo dice.
func TestSliceNuevoPideFeature(t *testing.T) {
	raiz := modulo(t)
	for _, rel := range []string{"docs/E007/README.md", "docs/E007/F001/README.md", "docs/E007/F001/S001.md",
		"docs/E007/F001/S002.md", "docs/E007/F001/S003.md", "docs/E007/F002/README.md"} {
		ruta := filepath.Join(raiz, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(ruta, []byte("# x\n\n## Para qué existe\n\nContrato.\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	crear := func(campos ...string) string {
		t.Helper()
		f := nuevo(t, raiz, campos...)
		return AvisoFeatureAprobada(raiz, f)
	}
	revisar := func(id string) error {
		t.Helper()
		f, _, err := Ver(raiz, id)
		if err != nil {
			t.Fatal(err)
		}
		return aprobada(raiz, f)
	}
	aprobar := func(id, msg string) {
		t.Helper()
		if _, err := Aprobar(raiz, id, msg, respaldo(t, msg), hora); err != nil {
			t.Fatal(err)
		}
	}

	// Cada tipo de ficha: sin feature aprobada nada avisa.
	for _, c := range [][]string{{"id=E007", "titulo=I"}, {"id=E007/F001", "titulo=F"}, {"id=E007/F001/S001", "titulo=A", "prueba=./p:TestA=a|r|v"}} {
		if a := crear(append(c, "prioridad=P1", "motivo=m")...); a != "" {
			t.Errorf("%s: sin feature aprobada no hay aviso; dio %q", c[0], a)
		}
	}
	aprobar("E007/F001", "apruebo E007/F001 con push")
	if err := revisar("E007/F001/S001"); err != nil {
		t.Errorf("S001 está en la foto: avanza; dio %v", err)
	}
	if _, err := Estado(raiz, "E007/F001/S001", "CANCELADO"); err != nil {
		t.Fatal(err)
	}

	// Slice nuevo en feature aprobada: avisa al crear y el rechazo pide la feature, aunque sea el único abierto.
	if a := crear("id=E007/F001/S002", "titulo=B", "prioridad=P1", "motivo=m", "prueba=./p:TestB=b|r|v"); !strings.Contains(a, "«apruebo E007/F001 con push»") {
		t.Errorf("crear un slice en una feature aprobada avisa que pide «apruebo E007/F001 con push»; dio %q", a)
	}
	if a := crear("id=E007/F002", "titulo=G", "prioridad=P1", "motivo=m"); a != "" {
		t.Errorf("una feature nueva no avisa; dio %q", a)
	}
	err := revisar("E007/F001/S002")
	if err == nil || !strings.Contains(err.Error(), "«apruebo E007/F001 con push»") || strings.Contains(err.Error(), "apruebo E007/F001/S002") {
		t.Errorf("slice fuera de la foto: pide «apruebo E007/F001 con push», no el del slice; dio %v", err)
	}

	// Fallo a mitad de camino: el apruebo nuevo de la feature lo cubre.
	aprobar("E007/F001", "apruebo E007/F001 con push")
	if err := revisar("E007/F001/S002"); err != nil {
		t.Errorf("tras el apruebo nuevo de la feature, S002 avanza; dio %v", err)
	}

	// Choque con otro control: la aprobación propia del slice sigue valiendo primero.
	crear("id=E007/F001/S003", "titulo=C", "prioridad=P1", "motivo=m", "prueba=./p:TestC=c|r|v")
	aprobar("E007/F001/S003", "apruebo E007/F001/S003")
	if err := revisar("E007/F001/S003"); err != nil {
		t.Errorf("la aprobación propia del slice vale; dio %v", err)
	}
}
