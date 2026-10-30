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
	ors, _ := oracionesConLinea(lineas, nil)
	return ors
}

// palabrasDe: palabras de un texto de prosa, sin enlaces ni HTML.
func palabrasDe(texto string) []string {
	texto = reEnlaceTexto.ReplaceAllString(texto, "$1")
	texto = reHTML.ReplaceAllString(texto, " ")
	return strings.FieldsFunc(texto, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
}

// sinAbierto quita del final de s una etiqueta HTML o un enlace que cierra en la línea siguiente:
// sus palabras cuentan donde cierran. Un «<» o «[» suelto, que no cierra ahí, no se toca (hallazgo A2.1).
func sinAbierto(s, siguiente string) string {
	if i := strings.LastIndex(s, "<"); i > strings.LastIndex(s, ">") && strings.Contains(siguiente, ">") {
		s = s[:i]
	}
	i := strings.LastIndex(s, "[")
	switch {
	case i < 0:
	case i > strings.LastIndex(s, "]") && strings.Contains(siguiente, "]("):
		s = s[:i]
	case strings.LastIndex(s, "](") > i && strings.LastIndex(s, "](") > strings.LastIndex(s, ")") && strings.Contains(siguiente, ")"):
		s = s[:i]
	}
	return s
}

// oracionesConLinea (E000/F006/S004): como oraciones, y además la línea donde empieza cada una,
// tomada de nums (una por entrada de lineas). Sin nums, las líneas salen en 0.
func oracionesConLinea(lineas []string, nums []int) ([][]string, []int) {
	var res [][]string
	var donde []int
	var parrafo []string
	var parrafoLin []int
	cerrar := func() {
		texto := strings.Join(parrafo, " ")
		// palabras acumuladas hasta cada fragmento, limpiando el prefijo entero: un enlace
		// partido en dos líneas se reconoce igual que en el texto unido (hallazgo A1.2).
		var hasta []int
		for j := range parrafo {
			siguiente := ""
			if j+1 < len(parrafo) {
				siguiente = parrafo[j+1]
			}
			hasta = append(hasta, len(palabrasDe(sinAbierto(strings.Join(parrafo[:j+1], " "), siguiente))))
		}
		texto = reEnlaceTexto.ReplaceAllString(texto, "$1")
		texto = reHTML.ReplaceAllString(texto, " ")
		vistas := 0
		for _, o := range reFinOracion.Split(texto, -1) {
			ps := strings.FieldsFunc(o, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
			if len(ps) == 0 {
				continue
			}
			linea := parrafoLin[len(parrafoLin)-1] // red: nunca línea 0 si la cuenta no cuadra
			for j, h := range hasta {
				if vistas < h {
					linea = parrafoLin[j]
					break
				}
			}
			res, donde = append(res, ps), append(donde, linea)
			vistas += len(ps)
		}
		parrafo, parrafoLin = nil, nil
	}
	for i, l := range lineas {
		n := 0
		if i < len(nums) {
			n = nums[i]
		}
		t := strings.TrimSpace(l)
		switch {
		case t == "":
			cerrar()
		case reItem.MatchString(t):
			cerrar()
			parrafo, parrafoLin = append(parrafo, reItem.ReplaceAllString(t, "")), append(parrafoLin, n)
		default:
			parrafo, parrafoLin = append(parrafo, t), append(parrafoLin, n)
		}
	}
	cerrar()
	return res, donde
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
	ps := largas(pg)
	ors := oraciones(pg.prosa)
	if len(ors) == 0 {
		return ps
	}
	if p := Perspicuidad(ors); p < PisoLegibilidad {
		ps = append(ps, fmt.Sprintf("%s: legibilidad %s < %s (INFLESZ «bastante fácil»); usa palabras y oraciones más cortas",
			pg.rel, coma(p), coma(PisoLegibilidad)))
	}
	return ps
}

// largas: un aviso por oración de más de MaxPalabrasOracion, con la línea donde empieza.
func largas(pg *pagina) []string {
	var ps []string
	ors, lin := oracionesConLinea(pg.prosa, pg.prosaLinea)
	for i, o := range ors {
		if len(o) > MaxPalabrasOracion {
			ps = append(ps, fmt.Sprintf("%s:%d: oración de %d palabras (máx. %d): «%s…»",
				pg.rel, lin[i], len(o), MaxPalabrasOracion, strings.Join(o[:6], " ")))
		}
	}
	return ps
}

// OracionesLargas (E000/F006/S004): avisos de oración larga de una página de docs/, para el hook al guardar.
func OracionesLargas(raiz, rel string) ([]string, error) {
	pg := &pagina{rel: rel}
	if err := analizar(raiz, pg); err != nil {
		return nil, err
	}
	return largas(pg), nil
}

// coma formatea un número con un decimal y coma decimal.
func coma(x float64) string { return strings.Replace(fmt.Sprintf("%.1f", x), ".", ",", 1) }
