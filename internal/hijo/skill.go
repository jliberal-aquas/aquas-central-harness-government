package hijo

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// E001/F007/S003: skills del hijo, sacados de la copia sellada de los canónicos centrales.
// E001/F018/S001: viajan todos los de gobierno/canonico/skills, no solo /backlog.
const (
	// carpetaSkills: canónicos de skills en la central; cada <nombre>.md va a .claude/skills/<nombre>/SKILL.md.
	carpetaSkills = "gobierno/canonico/skills"
	// comandoCentral: cómo corre aquas el central (desde su código); el hijo usa el binario instalado.
	comandoCentral = "sh .claude/hooks/aquas.sh "
	comandoHijo    = "aquas "
)

// marcaSkill: comienzo de la marca que distingue el SKILL.md generado de uno propio del hijo.
const marcaSkill = "<!-- GENERADO por `aquas build` desde .aquas/central/"

// fuentesSkill: canónicos de skills que trae la central, ordenados.
func fuentesSkill(central fs.FS) []string {
	xs, _ := fs.Glob(central, carpetaSkills+"/*.md")
	sort.Strings(xs)
	return xs
}

// rutaSkill: dónde queda en el hijo el skill de esa fuente.
func rutaSkill(fuente string) string {
	return ".claude/skills/" + nombreSkill(fuente) + "/SKILL.md"
}

// nombreSkill: nombre del comando del chat, sacado del archivo canónico (retro.md → retro).
func nombreSkill(fuente string) string {
	return strings.TrimSuffix(path.Base(fuente), ".md")
}

// skillPropio (hallazgo A1.1): error si el hijo ya tiene en ruta un SKILL.md que no generó aquas; build no lo pisa.
func skillPropio(dir, ruta string) error {
	b, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(ruta)))
	if err != nil || strings.Contains(string(b), marcaSkill) {
		return nil
	}
	nombre := path.Base(path.Dir(ruta))
	return fmt.Errorf("%s ya existe y no lo generó aquas: renómbralo o muévelo y corre aquas build (no se pisa un /%s propio)", ruta, nombre)
}

// armarSkills devuelve ruta → SKILL.md del hijo por cada skill de la central sellada; vacío si no trae ninguno
// (central anterior a S003): no es error. Falla si una fuente corre un comando sin el del central, o no abre
// con frontmatter.
func armarSkills(dir string) (map[string]string, error) {
	central := os.DirFS(filepath.Join(dir, Carpeta, "central"))
	out := map[string]string{}
	for _, f := range fuentesSkill(central) {
		b, err := fs.ReadFile(central, f)
		if err != nil {
			return nil, err
		}
		nombre := nombreSkill(f)
		src := strings.ReplaceAll(string(b), "\r\n", "\n")
		// Hallazgo A1.7: cada línea que corre un comando lo hace con el aquas del central, que aquí se reemplaza.
		for _, l := range strings.Split(src, "\n") {
			if strings.Contains(l, "!`") && !strings.Contains(l, comandoCentral) {
				return nil, fmt.Errorf("skill /%s de la central corre un comando sin %q: la central cambió el comando; ajusta internal/hijo/skill.go", nombre, comandoCentral)
			}
		}
		fm, resto, ok := strings.Cut(strings.TrimPrefix(src, "---\n"), "\n---\n")
		if !strings.HasPrefix(src, "---\n") || !ok {
			return nil, fmt.Errorf("skill /%s de la central sin frontmatter", nombre)
		}
		cab := marcaSkill + f + ". NO EDITAR. -->\n\n"
		txt := "---\n" + fm + "\n---\n" + cab + strings.TrimLeft(resto, "\n")
		out[rutaSkill(f)] = strings.ReplaceAll(txt, comandoCentral, comandoHijo)
	}
	return out, nil
}

// rutasOrdenadas: claves de skills en orden, para escribir y avisar siempre igual.
func rutasOrdenadas(skills map[string]string) []string {
	rs := make([]string, 0, len(skills))
	for r := range skills {
		rs = append(rs, r)
	}
	sort.Strings(rs)
	return rs
}
