package compilador

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// TestSkillRetro (E001/F018/S001): /retro compila desde su canónico con fuentes, categorías, reparto y salida.
// Rojo cuando: falta el skill, alguna sección obligatoria, alguna categoría o la regla de reparto.
// Verde cuando: compila con todas, declarado con tope en documentos.json.
func TestSkillRetro(t *testing.T) {
	const ruta = ".claude/skills/retro/SKILL.md"
	salidas, err := Planificar("../..")
	if err != nil {
		t.Fatal(err)
	}
	var skill string
	for _, s := range salidas {
		if s.Ruta == ruta {
			skill = string(s.Contenido)
		}
	}
	if skill == "" {
		t.Fatalf("no hay salida %s", ruta)
	}
	_, resto, ok := strings.Cut(strings.TrimPrefix(skill, "---\n"), "\n---\n")
	if !strings.HasPrefix(skill, "---\n") || !ok || !strings.HasPrefix(resto, "<!-- GENERADO") {
		t.Fatalf("el skill debe abrir con frontmatter y seguir con la marca GENERADO:\n%s", skill)
	}
	// Entrada ausente: cada sección obligatoria y lo que debe decir dentro.
	secciones := map[string][]string{
		"## Fuentes":    {"familia", "measure runs", "correcciones del humano", "registro de sesión"},
		"## Categorías": {"control que falta", "sin conectar", "criterio", "navegación", "herramientas", "no cambia nada", "acceso a información"},
		"## Reparto":    {"mecánico", "gate", "auditor", "puntero"},
		"## Salida":     {"gravedad", "recomendación", "backlog new", "retro close"},
	}
	for titulo, claves := range secciones {
		_, cuerpo, ok := strings.Cut(resto, "\n"+titulo+"\n")
		if !ok {
			t.Errorf("falta la sección %q", titulo)
			continue
		}
		cuerpo, _, _ = strings.Cut(cuerpo, "\n## ")
		for _, c := range claves {
			if !strings.Contains(strings.ToLower(cuerpo), c) {
				t.Errorf("sección %q sin %q", titulo, c)
			}
		}
	}
	// Quién habla: lo invoca el humano o el orquestador cuando next lo pide; el skill no crea fichas solo.
	// Hallazgo A1.5: la frase exacta, no «humano» suelto, que aparece en otras secciones.
	for _, c := range []string{"La invoca el humano con `/retro`", "el orquestador, cuando `aquas backlog next` avisa «toca retrospectiva»",
		"no crea fichas", "se reporta con `aquas report`"} {
		if !strings.Contains(resto, c) {
			t.Errorf("el skill debe decir quién lo invoca y su límite: falta %q", c)
		}
	}
	// Choque con el sello y el tope: declarado en documentos.json con su fuente y max_caracteres.
	b, err := os.ReadFile("../../" + RutaDocumentos)
	if err != nil {
		t.Fatal(err)
	}
	var cfg struct {
		Documentos []struct {
			Salida  string   `json:"salida"`
			Fuentes []string `json:"fuentes"`
			Max     int      `json:"max_caracteres"`
		} `json:"documentos"`
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		t.Fatal(err)
	}
	declarado := false
	for _, d := range cfg.Documentos {
		if d.Salida == ruta {
			declarado = len(d.Fuentes) == 1 && d.Fuentes[0] == "gobierno/canonico/skills/retro.md" && d.Max > 0
		}
	}
	if !declarado {
		t.Errorf("%s debe declararse con fuente gobierno/canonico/skills/retro.md y max_caracteres", ruta)
	}
	// Hallazgo A2.2: un canónico de skill sin declarar viajaría al hijo sin compilarse aquí; el gate lo nombra.
	raiz := repoMinimo(t)
	if err := escribir(raiz, "gobierno/canonico/skills/x.md", []byte("---\nname: x\n---\n# X\n")); err != nil {
		t.Fatal(err)
	}
	if _, err := Compilar(raiz); err != nil {
		t.Fatal(err)
	}
	if p, err := Verificar(raiz); err != nil || !strings.Contains(strings.Join(p, "\n"), "skills/x.md: skill sin declarar") {
		t.Errorf("skill sin declarar: Verificar debió nombrarlo; dio %v %v", p, err)
	}
}
