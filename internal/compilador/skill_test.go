package compilador

import (
	"strings"
	"testing"
)

// TestSkillBacklog (E001/F007/S002): el compilador genera /backlog con su frontmatter primero; solo lo invoca
// el humano y corre `aquas backlog tabla`.
// Rojo cuando: no hay salida .claude/skills/backlog/SKILL.md, la marca GENERADO va antes del frontmatter
// o falta disable-model-invocation.
// Verde cuando: el skill empieza con frontmatter, sigue la marca y su cuerpo corre la tabla.
func TestSkillBacklog(t *testing.T) {
	salidas, err := Planificar("../..")
	if err != nil {
		t.Fatal(err)
	}
	var skill string
	for _, s := range salidas {
		if s.Ruta == ".claude/skills/backlog/SKILL.md" {
			skill = string(s.Contenido)
		}
	}
	if skill == "" {
		t.Fatal("no hay salida .claude/skills/backlog/SKILL.md")
	}
	fm, resto, ok := strings.Cut(strings.TrimPrefix(skill, "---\n"), "\n---\n")
	if !strings.HasPrefix(skill, "---\n") || !ok {
		t.Fatalf("el skill debe empezar con frontmatter:\n%s", skill)
	}
	if !strings.Contains(fm, "disable-model-invocation: true") {
		t.Errorf("frontmatter sin disable-model-invocation: true:\n%s", fm)
	}
	if !strings.HasPrefix(resto, "<!-- GENERADO") {
		t.Errorf("la marca GENERADO debe seguir al frontmatter:\n%s", resto)
	}
	if !strings.Contains(resto, "!`") || !strings.Contains(resto, "backlog tabla") {
		t.Errorf("el cuerpo debe correr `backlog tabla`:\n%s", resto)
	}
	// Choque con el sello: un SKILL.md editado a mano lo detecta Verificar, igual que CLAUDE.md.
	raiz := repoMinimo(t)
	docs := `{"documentos":[{"salida":"CLAUDE.md","fuentes":["gobierno/canonico/constitucion.md"],"max_caracteres":10000},` +
		`{"salida":".claude/skills/backlog/SKILL.md","fuentes":["gobierno/canonico/skills/backlog.md"],"max_caracteres":1500}]}`
	if err := escribir(raiz, RutaDocumentos, []byte(docs)); err != nil {
		t.Fatal(err)
	}
	if err := escribir(raiz, "gobierno/canonico/skills/backlog.md", []byte("---\nname: backlog\n---\n# B\n")); err != nil {
		t.Fatal(err)
	}
	if _, err := Compilar(raiz); err != nil {
		t.Fatal(err)
	}
	if p, err := Verificar(raiz); err != nil || len(p) != 0 {
		t.Fatalf("recién compilado debe pasar: %v %v", p, err)
	}
	if err := escribir(raiz, ".claude/skills/backlog/SKILL.md", []byte("---\nname: backlog\n---\n# a mano\n")); err != nil {
		t.Fatal(err)
	}
	if p, _ := Verificar(raiz); len(p) == 0 {
		t.Error("SKILL.md editado a mano: Verificar debe reportarlo")
	}
	// Un documento sin frontmatter conserva la marca al inicio.
	for _, s := range salidas {
		if s.Ruta == "CLAUDE.md" && !strings.HasPrefix(string(s.Contenido), "<!-- GENERADO") {
			t.Errorf("CLAUDE.md debe empezar con la marca")
		}
	}
}
