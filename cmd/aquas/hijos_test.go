package main

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// copiarGobierno copia gobierno/ del repo real a raiz: compilar necesita canónicos y config.
func copiarGobierno(t *testing.T, raiz string) {
	t.Helper()
	org := filepath.Join("..", "..", "gobierno")
	err := filepath.WalkDir(org, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(org, p)
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		dst := filepath.Join(raiz, "gobierno", rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		return os.WriteFile(dst, b, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
}

// E001/F002/S010: `aquas compilar` rehace el bloque de hijos cuando cambia una ficha.
func TestCompilarRegeneraHijos(t *testing.T) {
	raiz := t.TempDir()
	copiarGobierno(t, raiz)
	if out, err := exec.Command("git", "-C", raiz, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out) // iniciar exige ser la raíz de su propio repo
	}
	os.MkdirAll(filepath.Join(raiz, ".claude"), 0o755) // en un clon real ya existe: trae los agentes compilados
	if s := aquas(t, raiz, "", "iniciar"); s.codigo != 0 {
		t.Fatalf("iniciar: %+v", s)
	}
	inicio := filepath.Join(raiz, "docs", "README.md")
	os.MkdirAll(filepath.Dir(inicio), 0o755)
	os.WriteFile(inicio, []byte("# Inicio\n\n## Navegación\n\n- Glosario\n"), 0o644)
	if s := aquas(t, raiz, "", "backlog", "nuevo", "id=E001", "titulo=Gobierno", "prioridad=P1", "motivo=m", "prueba=./p:TestA=a"); s.codigo != 0 {
		t.Fatalf("backlog nuevo: %+v", s)
	}
	pagina := func() string {
		b, _ := os.ReadFile(inicio)
		return string(b)
	}
	s := aquas(t, raiz, "", "compilar")
	if s.codigo != 0 || !strings.Contains(s.out, "docs/README.md") {
		t.Fatalf("compilar escribe el inicio y lo informa; dio %+v", s)
	}
	if !strings.Contains(pagina(), "## Navegación\n\n<!-- GENERADO") || !strings.Contains(pagina(), "| E001 | Gobierno | Planeado |\n") {
		t.Fatalf("el inicio lista E001 como Planeado, sin enlace (no tiene página):\n%s", pagina())
	}
	if s := aquas(t, raiz, "", "backlog", "estado", "E001", "EN_CURSO"); s.codigo != 0 {
		t.Fatalf("backlog estado: %+v", s)
	}
	if s := aquas(t, raiz, "", "compilar"); s.codigo != 0 || !strings.Contains(s.out, "docs/README.md") {
		t.Fatalf("tras cambiar la ficha, compilar reescribe el inicio; dio %+v", s)
	}
	if c := pagina(); !strings.Contains(c, "| E001 | Gobierno | En curso |\n") || strings.Contains(c, "Planeado") || !strings.Contains(c, "- Glosario\n") {
		t.Fatalf("el bloque cambia a En curso y lo escrito a mano se conserva:\n%s", c)
	}
	if s := aquas(t, raiz, "", "compilar"); strings.Contains(s.out, "docs/README.md") {
		t.Errorf("sin cambios en fichas, compilar no toca el inicio; dio %+v", s)
	}
}
