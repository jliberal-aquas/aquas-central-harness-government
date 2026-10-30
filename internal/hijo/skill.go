package hijo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// E001/F007/S003: /backlog en el hijo, sacado de la copia sellada del canónico central.
const (
	FuenteSkill = "gobierno/canonico/skills/backlog.md"
	RutaSkill   = ".claude/skills/backlog/SKILL.md"
	// comandoCentral: cómo corre aquas el central (desde su código); el hijo usa el binario instalado.
	comandoCentral = "sh .claude/hooks/aquas.sh "
	comandoHijo    = "aquas "
)

// marcaSkill: comienzo de la marca que distingue el SKILL.md generado de uno propio del hijo.
const marcaSkill = "<!-- GENERADO por `aquas build` desde .aquas/central/"

// skillPropio (hallazgo A1.1): error si el hijo ya tiene un SKILL.md de /backlog que no generó aquas; build no lo pisa.
func skillPropio(dir string) error {
	b, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(RutaSkill)))
	if err != nil || strings.Contains(string(b), marcaSkill) {
		return nil
	}
	return fmt.Errorf("%s ya existe y no lo generó aquas: renómbralo o muévelo y corre aquas build (no se pisa un /backlog propio)", RutaSkill)
}

// armarSkill devuelve el SKILL.md del hijo. hay=false si la central sellada no trae el skill (central anterior
// a S003): no es error. Falla si la fuente no trae el comando del central, o no abre con frontmatter.
func armarSkill(dir string) (txt string, hay bool, err error) {
	b, err := os.ReadFile(filepath.Join(dir, Carpeta, "central", filepath.FromSlash(FuenteSkill)))
	if os.IsNotExist(err) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	src := strings.ReplaceAll(string(b), "\r\n", "\n")
	if !strings.Contains(src, comandoCentral) {
		return "", false, fmt.Errorf("skill /backlog de la central sin %q: la central cambió el comando; ajusta internal/hijo/skill.go", comandoCentral)
	}
	fm, resto, ok := strings.Cut(strings.TrimPrefix(src, "---\n"), "\n---\n")
	if !strings.HasPrefix(src, "---\n") || !ok {
		return "", false, fmt.Errorf("skill /backlog de la central sin frontmatter")
	}
	cab := marcaSkill + FuenteSkill + ". NO EDITAR. -->\n\n"
	out := "---\n" + fm + "\n---\n" + cab + strings.TrimLeft(resto, "\n")
	return strings.ReplaceAll(out, comandoCentral, comandoHijo), true, nil
}
