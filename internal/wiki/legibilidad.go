package wiki

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// Legibilidad de la prosa de cada página (S002).
//
// Índice de perspicuidad Szigriszt-Pazos (Flesch-Szigriszt):
//
//	P = 206,835 − 62,3 · (sílabas / palabras) − (palabras / oraciones)
//
// Escala INFLESZ: <40 muy difícil · 40–54 algo difícil · 55–64 normal · 65–79 bastante fácil · ≥80 muy fácil.
// Fuente de la escala: Ballesteros-Peña y Fernández-Aedo, An. Sist. Sanit. Navar. 2013;36(3):397-406, tabla 2.
// Constantes: Ríos Hernández, Chasqui 135 (2017), citando la tesis de Szigriszt (UCM, 1993), que redondea 206,835 a 207.
const (
	PisoLegibilidad    = 65.0 // INFLESZ «bastante fácil»
	MaxPalabrasOracion = 20
)

var (
	reEnlaceTexto = regexp.MustCompile(`!?\[([^\]]*)\]\([^)]*\)`)
	reHTML        = regexp.MustCompile(`<[^>]+>`)
	reFinOracion  = regexp.MustCompile(`[.!?…]+[*_»"”')]*(?:\s|$)`) // admite cierre de negrita, comillas o paréntesis
	reItem        = regexp.MustCompile(`^(?:[-*+]|\d+[.)]|>)\s+`)
)

// Silabas cuenta las sílabas de una palabra en español por núcleos vocálicos,
// separando hiatos (dos vocales fuertes, o vocal débil acentuada). Mínimo 1.
func Silabas(palabra string) int {
	rs := []rune(strings.ToLower(palabra))
	vocal := make([]bool, len(rs))
	for i, r := range rs {
		switch r {
		case 'a', 'e', 'i', 'o', 'á', 'é', 'í', 'ó', 'ú', 'ü':
			vocal[i] = true
		case 'u':
			// u muda en «que, qui, gue, gui»
			muda := i > 0 && (rs[i-1] == 'q' || rs[i-1] == 'g' && i+1 < len(rs) && strings.ContainsRune("eiéí", rs[i+1]))
			vocal[i] = !muda
		case 'y':
			// y es vocal al final o antes de consonante, si no inicia sílaba ante vocal
			vocal[i] = i+1 == len(rs) || !esVocalLetra(rs[i+1])
		}
	}
	n := 0
	for i := range rs {
		if !vocal[i] {
			continue
		}
		if i == 0 || !vocal[i-1] || hiato(rs[i-1], rs[i]) {
			n++
		}
	}
	if n == 0 {
		return 1
	}
	return n
}

// esVocalLetra: la runa es vocal castellana, con o sin tilde.
func esVocalLetra(r rune) bool { return strings.ContainsRune("aeiouáéíóúü", r) }

// hiato: las dos vocales seguidas se cuentan como sílabas separadas.
func hiato(a, b rune) bool {
	fuerte := func(r rune) bool { return strings.ContainsRune("aeoáéó", r) }
	debilTilde := func(r rune) bool { return r == 'í' || r == 'ú' }
	return fuerte(a) && fuerte(b) || debilTilde(a) || debilTilde(b)
}

// oraciones parte la prosa en oraciones de palabras. Línea vacía o ítem de lista
// cierra el párrafo; el párrafo también cierra oración.
func oraciones(lineas []string) [][]string {
	var res [][]string
	var parrafo []string
	cerrar := func() {
		texto := strings.Join(parrafo, " ")
		parrafo = nil
		texto = reEnlaceTexto.ReplaceAllString(texto, "$1")
		texto = reHTML.ReplaceAllString(texto, " ")
		for _, o := range reFinOracion.Split(texto, -1) {
			ps := strings.FieldsFunc(o, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
			if len(ps) > 0 {
				res = append(res, ps)
			}
		}
	}
	for _, l := range lineas {
		t := strings.TrimSpace(l)
		switch {
		case t == "":
			cerrar()
		case reItem.MatchString(t):
			cerrar()
			parrafo = append(parrafo, reItem.ReplaceAllString(t, ""))
		default:
			parrafo = append(parrafo, t)
		}
	}
	cerrar()
	return res
}

// Perspicuidad calcula el índice Szigriszt-Pazos; 0 si no hay palabras.
func Perspicuidad(ors [][]string) float64 {
	palabras, silabas := 0, 0
	for _, o := range ors {
		for _, p := range o {
			palabras++
			silabas += Silabas(p)
		}
	}
	if palabras == 0 {
		return 0
	}
	w := float64(palabras)
	return 206.835 - 62.3*float64(silabas)/w - w/float64(len(ors))
}

// revisarLegibilidad: avisos por oraciones largas o prosa difícil de leer.
func revisarLegibilidad(pg *pagina) []string {
	var ps []string
	ors := oraciones(pg.prosa)
	for _, o := range ors {
		if len(o) > MaxPalabrasOracion {
			ps = append(ps, fmt.Sprintf("%s: oración de %d palabras (máx. %d): «%s…»",
				pg.rel, len(o), MaxPalabrasOracion, strings.Join(o[:6], " ")))
		}
	}
	if len(ors) == 0 {
		return ps
	}
	if p := Perspicuidad(ors); p < PisoLegibilidad {
		ps = append(ps, fmt.Sprintf("%s: legibilidad %s < %s (INFLESZ «bastante fácil»); usa palabras y oraciones más cortas",
			pg.rel, coma(p), coma(PisoLegibilidad)))
	}
	return ps
}

// coma formatea un número con un decimal y coma decimal.
func coma(x float64) string { return strings.Replace(fmt.Sprintf("%.1f", x), ".", ",", 1) }
