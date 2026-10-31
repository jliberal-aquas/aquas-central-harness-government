package backlog

// E001/F002/S007: alcance vigilado (opción B, decisión humana 2026-10-21). El slice en curso
// declara sus rutas (`aquas backlog scope <id> ruta…`; «dir/**» = todo lo de dentro). Tocar
// otra ruta exige ampliar el alcance con motivo de una línea, que queda en la ficha. Siempre
// permitidos: derivados del compilador y datos que escribe aquas (trabajo/, registro/, …).
// Sin slice en curso no se restringe. La guardia lo aplica a Write/Edit y a destinos de Bash.

import (
	"errors"
	"fmt"
	"path"
	"strings"
)

// Ampliacion: ruta agregada al alcance con su motivo.
type Ampliacion struct {
	Ruta   string `json:"ruta"`
	Motivo string `json:"motivo"`
}

// limpia normaliza una ruta: barras «/», sin «./» inicial ni segmentos redundantes.
func limpia(r string) string {
	return strings.TrimPrefix(path.Clean(strings.ReplaceAll(r, `\`, "/")), "./")
}

// Alcance declara (reemplaza) las rutas del slice.
func Alcance(raiz, id string, rutas []string) (Ficha, error) {
	if len(rutas) == 0 {
		return Ficha{}, errors.New("alcance vacío: aquas backlog scope <id> ruta… («dir/**» = todo lo de dentro)")
	}
	f, err := activa(raiz, id)
	if err != nil {
		return Ficha{}, err
	}
	if f.Tipo != "slice" {
		return Ficha{}, fmt.Errorf("%s: el alcance es de slices", id)
	}
	f.Alcance = nil
	for _, r := range rutas {
		f.Alcance = append(f.Alcance, limpia(r))
	}
	return f, escribir(raiz, Activo, f)
}

// Ampliar agrega una ruta al alcance; el motivo es obligatorio y de una línea.
func Ampliar(raiz, id, ruta, motivo string) (Ficha, error) {
	if motivo = strings.TrimSpace(motivo); motivo == "" || strings.ContainsAny(motivo, "\r\n") {
		return Ficha{}, errors.New("motivo obligatorio, de una línea")
	}
	f, err := activa(raiz, id)
	if err != nil {
		return Ficha{}, err
	}
	f.Ampliaciones = append(f.Ampliaciones, Ampliacion{Ruta: limpia(ruta), Motivo: motivo})
	return f, escribir(raiz, Activo, f)
}

// cubre: «dir/**» cubre todo lo de dentro; si no, patrón de path.Match o ruta exacta.
func cubre(patron, rel string) bool {
	if d, ok := strings.CutSuffix(patron, "/**"); ok {
		return strings.HasPrefix(rel, d+"/")
	}
	ok, _ := path.Match(patron, rel)
	return ok
}

// FueraDeAlcance: "" si rel (relativa a la raíz) se puede tocar; si no, el motivo y cómo seguir.
func FueraDeAlcance(raiz, rel string) string {
	rel = limpia(rel)
	if excluida(rel) || derivados(raiz)[rel] {
		return ""
	}
	id, hay, err := SliceEnCurso(raiz)
	if err != nil || !hay {
		return "" // backlog ilegible lo reporta el gate; sin slice en curso no hay alcance
	}
	f, err := activa(raiz, id)
	if err != nil {
		return ""
	}
	if len(f.Alcance) == 0 {
		return fmt.Sprintf("%s está en curso sin alcance declarado: declara el alcance con `aquas backlog scope %s ruta…` antes de escribir %s", id, id, rel)
	}
	for _, p := range f.Alcance {
		if cubre(p, rel) {
			return ""
		}
	}
	for _, a := range f.Ampliaciones {
		if cubre(a.Ruta, rel) {
			return ""
		}
	}
	return fmt.Sprintf("%s fuera del alcance de %s (%s): si hace falta, `aquas backlog scope %s extend %s motivo=…`; si es otro trabajo, va al backlog",
		rel, id, strings.Join(f.Alcance, ", "), id, rel)
}
