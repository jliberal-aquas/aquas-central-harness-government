package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCommitExigeSliceEnCurso (E000/F014/S003): no se commitea código fuera de un slice en curso.
// Rojo cuando: un commit con código pasa sin slice EN_CURSO que lo cubra.
// Verde cuando: se rechaza nombrando la ruta y pasa con el alcance declarado.
func TestCommitExigeSliceEnCurso(t *testing.T) {
	const falta = ": ningún slice EN_CURSO lo cubre"
	verificar := func(t *testing.T, raiz string) string {
		t.Helper()
		s := aquas(t, raiz, "", "verify")
		return s.out + s.err
	}
	preparar := func(t *testing.T, raiz string, rels ...string) {
		t.Helper()
		for _, r := range rels {
			escribir(t, raiz, r, "x\n")
		}
		git(t, raiz, append([]string{"add", "--"}, rels...)...)
	}

	t.Run("central: código sin slice en curso frena; gobierno y alcance pasan", func(t *testing.T) {
		raiz := repoContrato(t) // E001/F002/S003 en BORRADOR con alcance p/**
		git(t, raiz, "init", "-q")
		if o := verificar(t, raiz); strings.Contains(o, falta) {
			t.Fatalf("sin archivos preparados no cambia nada; dio:\n%s", o)
		}
		preparar(t, raiz, "p/nuevo.go", "docs/E001/F002/x.md", "trabajo/nota.txt", "gobierno/x.md", "CLAUDE.md", ".claude/x.json", ".githooks/x", "registro/x.json", ".aquas/x.json")
		o := verificar(t, raiz)
		if !strings.Contains(o, "p/nuevo.go"+falta) {
			t.Errorf("slice en BORRADOR no cubre: debe nombrar p/nuevo.go; dio:\n%s", o)
		}
		for _, g := range []string{"docs/E001/F002/x.md", "trabajo/nota.txt", "gobierno/x.md", "CLAUDE.md", ".claude/x.json", ".githooks/x", "registro/x.json", ".aquas/x.json"} {
			if strings.Contains(o, g+falta) {
				t.Errorf("gobierno %s no exige slice; dio:\n%s", g, o)
			}
		}
		aprobarCLI(t, raiz, "E001/F002/S003")
		if s := aquas(t, raiz, "", "backlog", "state", "E001/F002/S003", "EN_CURSO"); s.codigo != 0 {
			t.Fatalf("EN_CURSO: %+v", s)
		}
		preparar(t, raiz, "src/otro.ts")
		o = verificar(t, raiz)
		if strings.Contains(o, "p/nuevo.go"+falta) || !strings.Contains(o, "src/otro.ts"+falta) {
			t.Errorf("EN_CURSO cubre p/** y no src/otro.ts; dio:\n%s", o)
		}

		// El slice que se cierra en este commit (ficha nueva en trabajo/archivo/, CERRADO) cubre su alcance.
		b, err := os.ReadFile(filepath.Join(raiz, "trabajo", "activo", "E001-F002-S003.json"))
		if err != nil {
			t.Fatal(err)
		}
		cerrada := strings.NewReplacer(`"E001/F002/S003"`, `"E001/F002/S009"`, `"EN_CURSO"`, `"CERRADO"`, `"p/**"`, `"src/**"`).Replace(string(b))
		escribir(t, raiz, "trabajo/archivo/E001-F002-S009.json", cerrada)
		git(t, raiz, "add", "trabajo/archivo/E001-F002-S009.json")
		if o := verificar(t, raiz); !strings.Contains(o, "src/otro.ts"+falta) {
			t.Errorf("ficha CERRADO sin auditoría APROBADO (fabricada) no cubre (A1.4); dio:\n%s", o)
		}
		auditada := strings.Replace(cerrada, `"estado": "CERRADO"`, `"estado": "CERRADO", "auditoria": {"fecha": "2026-11-05T00:00:00Z", "veredicto": "APROBADO"}`, 1)
		escribir(t, raiz, "trabajo/archivo/E001-F002-S009.json", auditada)
		git(t, raiz, "add", "trabajo/archivo/E001-F002-S009.json")
		if o := verificar(t, raiz); strings.Contains(o, "src/otro.ts"+falta) {
			t.Errorf("el slice cerrado y auditado en este commit cubre su alcance; dio:\n%s", o)
		}

		// Arreglo sobre un slice ya archivado (`backlog fix`): su ficha cambia (M) en este commit y cubre (A1.1).
		git(t, raiz, "-c", "core.hooksPath=/dev/null", "commit", "-q", "-m", "base")
		// Un slice archivado que no se cierra ni recibe arreglo en este commit no cubre, aunque una orden de aquas
		// (edit, document) toque su ficha (A2.1, A2.2).
		preparar(t, raiz, "src/sin_slice.ts")
		escribir(t, raiz, "trabajo/archivo/E001-F002-S009.json", strings.Replace(auditada, `"titulo"`, `"ediciones": [], "titulo"`, 1))
		git(t, raiz, "add", "trabajo/archivo/E001-F002-S009.json")
		if o := verificar(t, raiz); !strings.Contains(o, "src/sin_slice.ts"+falta) {
			t.Errorf("slice archivado sin cierre ni arreglo en este commit no cubre; dio:\n%s", o)
		}
		git(t, raiz, "rm", "-q", "--cached", "src/sin_slice.ts")
		escribir(t, raiz, "src/arreglo.ts", "y\n")
		const arreglo = `"arreglos": [{"id": "E001/F002/S009/X01", "titulo": "x", "motivo": "m", "fecha": "2026-11-05"%s}], "titulo"`
		escribir(t, raiz, "trabajo/archivo/E001-F002-S009.json", strings.Replace(auditada, `"titulo"`, fmt.Sprintf(arreglo, ""), 1))
		git(t, raiz, "add", "src/arreglo.ts", "trabajo/archivo/E001-F002-S009.json")
		if o := verificar(t, raiz); !strings.Contains(o, "src/arreglo.ts"+falta) {
			t.Errorf("un arreglo sin rojo ni verde no abre el paso (A3.1); dio:\n%s", o)
		}
		probado := `, "rojo": {"fecha": "2026-11-05T00:00:00Z", "resultado": {"TestX": "fail"}}, "verde": {"fecha": "2026-11-05T00:01:00Z", "resultado": {"TestX": "pass"}}`
		escribir(t, raiz, "trabajo/archivo/E001-F002-S009.json", strings.Replace(auditada, `"titulo"`, fmt.Sprintf(arreglo, probado), 1))
		git(t, raiz, "add", "trabajo/archivo/E001-F002-S009.json")
		if o := verificar(t, raiz); strings.Contains(o, "src/arreglo.ts"+falta) {
			t.Errorf("el arreglo con rojo y verde de un slice archivado cubre su alcance; dio:\n%s", o)
		}

		// Ficha en HEAD ilegible: no se sabe si el arreglo es nuevo; falla cerrado y no cubre (A3.2).
		escribir(t, raiz, "trabajo/archivo/E001-F002-S010.json", "{")
		git(t, raiz, "add", "trabajo/archivo/E001-F002-S010.json")
		git(t, raiz, "-c", "core.hooksPath=/dev/null", "commit", "-q", "-m", "ficha rota", "--", "trabajo/archivo/E001-F002-S010.json")
		git(t, raiz, "reset", "-q", "--", "trabajo/archivo/E001-F002-S009.json")
		reparada := strings.NewReplacer(`"E001/F002/S009"`, `"E001/F002/S010"`, `E001/F002/S009/X01`, `E001/F002/S010/X01`).Replace(strings.Replace(auditada, `"titulo"`, fmt.Sprintf(arreglo, probado), 1))
		escribir(t, raiz, "trabajo/archivo/E001-F002-S010.json", reparada)
		git(t, raiz, "add", "trabajo/archivo/E001-F002-S010.json")
		if o := verificar(t, raiz); !strings.Contains(o, "src/arreglo.ts"+falta) {
			t.Errorf("ficha ilegible en HEAD: no cubre; dio:\n%s", o)
		}
		git(t, raiz, "rm", "-q", "--cached", "src/arreglo.ts")

		// README.md raíz es la portada que exige el gate de portadas: no pide slice (A1.2).
		preparar(t, raiz, "README.md")
		if o := verificar(t, raiz); strings.Contains(o, "README.md"+falta) {
			t.Errorf("README.md raíz no exige slice; dio:\n%s", o)
		}
	})

	t.Run("hijo: el mismo control en su pre-commit", func(t *testing.T) {
		dir := t.TempDir()
		if s := aquas(t, dir, "", "install"); s.codigo != 0 {
			t.Fatalf("install: %+v", s)
		}
		antes := aquas(t, dir, "", "verify")
		preparar(t, dir, "src/validez.ts")
		s := aquas(t, dir, "", "verify")
		if !strings.Contains(s.out+s.err, "src/validez.ts"+falta) || !strings.Contains(s.out+s.err, "commit: 1 archivos sin slice EN_CURSO") {
			t.Errorf("en el hijo el código sin slice EN_CURSO también frena; dio %+v", s)
		}
		if antes.codigo == 0 && s.codigo == 0 {
			t.Errorf("el hijo pasaba el gate; con código sin slice debe frenar (exit≠0); dio %+v", s)
		}
	})

	t.Run("git ilegible: lo dice y frena", func(t *testing.T) {
		raiz := repoContrato(t)
		escribir(t, raiz, ".git", "gitdir: no/existe\n")
		s := aquas(t, raiz, "", "verify")
		if s.codigo == 0 || !strings.Contains(s.out+s.err, "no pude leer los archivos preparados") {
			t.Errorf("git ilegible: frena diciéndolo; dio %+v", s)
		}
	})
}
