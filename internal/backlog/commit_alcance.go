package backlog

// E000/F014/S003: un commit no lleva código fuera de un slice en curso. La guardia solo vigila el alcance con
// un slice EN_CURSO; sin él, el agente del hijo escribió y commiteó código de slices en LISTO (D000125).
// El pre-commit (`aquas verify`) revisa los archivos preparados: cada uno de código debe caer en el alcance del
// slice EN_CURSO, o del slice cerrado y auditado que se archiva en este commit o recibe en él un arreglo nuevo.
// Un arreglo nuevo cubre solo con su rojo y su verde registrados (A3.1).
// Límite: el gate no valida la huella de la auditoría archivada; lo que frena una ficha fabricada es la guardia,
// que no deja escribir trabajo/ a mano. Un arreglo no pide apruebo del humano: hueco aparte, fuera de este slice.

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// gobiernoSinSlice: rutas de gobierno que un commit no exige cubrir con un slice.
var gobiernoSinSlice = []string{"docs/", "trabajo/", "registro/", ".aquas/", ".claude/", "gobierno/", ".githooks/"}

// esGobierno: rel es de gobierno (o CLAUDE.md, o un derivado que escribe aquas).
func esGobierno(raiz, rel string) bool {
	if rel == "CLAUDE.md" || rel == "README.md" || derivados(raiz)[rel] { // README.md raíz: portada de la wiki (A1.2)
		return true
	}
	for _, g := range gobiernoSinSlice {
		if strings.HasPrefix(rel, g) {
			return true
		}
	}
	return false
}

// enHEAD: la ficha tal como está en el último commit; false si no estaba (entra nueva en este commit).
// Falla cerrado (A3.2): si estaba en HEAD y no se puede leer, devuelve error y el llamador no cubre.
func enHEAD(raiz, rel string) (Ficha, bool, error) {
	if exec.Command("git", "-C", raiz, "rev-parse", "--verify", "-q", "HEAD").Run() != nil {
		return Ficha{}, false, nil // repo sin commits: toda ficha entra nueva
	}
	ls, err := exec.Command("git", "-C", raiz, "ls-tree", "--name-only", "HEAD", "--", rel).Output()
	if err != nil {
		return Ficha{}, false, err
	}
	if strings.TrimSpace(string(ls)) == "" {
		return Ficha{}, false, nil
	}
	out, err := exec.Command("git", "-C", raiz, "show", "HEAD:"+rel).Output()
	if err != nil {
		return Ficha{}, false, err
	}
	var f Ficha
	if err := json.Unmarshal(out, &f); err != nil {
		return Ficha{}, false, err
	}
	return f, true, nil
}

// arregloProbado: el último arreglo nuevo tiene rojo y verde propios (A3.1): un `backlog fix` sin prueba no abre el paso.
func arregloProbado(f Ficha, previos int) bool {
	if len(f.Arreglos) <= previos {
		return false
	}
	a := f.Arreglos[len(f.Arreglos)-1]
	return a.Rojo != nil && a.Verde != nil
}

// CommitSinSlice: problemas de los archivos preparados para commit; nada sin repo git o sin archivos preparados.
func CommitSinSlice(raiz string) []string {
	if _, err := os.Stat(filepath.Join(raiz, ".git")); err != nil {
		return nil // sin repo git no hay commit que revisar
	}
	preparados, err := gitLineas(raiz, "diff", "--cached", "--name-only", "--diff-filter=ACMRD")
	if err != nil {
		return []string{"no pude leer los archivos preparados con git: " + err.Error()}
	}
	// Ficha archivada que entra o cambia en este commit: el cierre del slice o un arreglo (`backlog fix`) sobre él.
	archivadas, err := gitLineas(raiz, "diff", "--cached", "--name-only", "--diff-filter=AM", "--", "trabajo/archivo/")
	if err != nil {
		return []string{"no pude leer los archivos preparados con git: " + err.Error()}
	}
	var cubren []Ficha
	id, hay, err := SliceEnCurso(raiz)
	if err != nil {
		return []string{"no pude leer el backlog para saber qué slice está en curso: " + err.Error()}
	}
	if hay {
		if f, err := activa(raiz, id); err == nil {
			cubren = append(cubren, f)
		}
	}
	for _, n := range archivadas {
		// Solo un slice cerrado con auditoría APROBADO (A1.4) que se cierra ahora o recibe un arreglo nuevo:
		// otra escritura de aquas sobre una ficha archivada (edit, document, code) no abre el paso (A2.1).
		f, err := leer(filepath.Join(raiz, filepath.FromSlash(n)))
		if err != nil || f.Tipo != "slice" || f.Estado != "CERRADO" || f.Auditoria == nil || f.Auditoria.Veredicto != "APROBADO" {
			continue
		}
		previo, ok, err := enHEAD(raiz, n)
		if err == nil && (!ok || arregloProbado(f, len(previo.Arreglos))) {
			cubren = append(cubren, f)
		}
	}
	var ps []string
	for _, rel := range preparados {
		if rel = limpia(rel); rel == "" || esGobierno(raiz, rel) {
			continue
		}
		cubierto := false
		for _, f := range cubren {
			cubierto = cubierto || cubreFicha(f, rel)
		}
		if !cubierto {
			ps = append(ps, rel+": ningún slice EN_CURSO lo cubre: aprueba y arranca su slice, o amplía el alcance del slice en curso")
		}
	}
	return ps
}
