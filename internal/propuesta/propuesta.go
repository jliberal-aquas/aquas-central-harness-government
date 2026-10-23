// Package propuesta revisa que una propuesta de proceso o diseño en el chat traiga diagrama
// (E001/F002/S006). El hook Stop pasa la respuesta completa del agente (last_assistant_message);
// cada línea «PROPUESTA: …» fuera de un bloque de código necesita después, antes de la siguiente
// propuesta, un bloque de código con flechas o cajas. El chat no dibuja Mermaid: va en texto.
// Límite (criterio, no gate): si el agente no escribe «PROPUESTA:», el hook no se entera.
package propuesta

import "strings"

// trazos: lo que convierte un bloque de código en diagrama. Un bloque con un comando no lo es.
var trazos = []string{"→", "->", "─►", "►", "▼", "▲", "◄", "<-", "┌", "┐", "└", "┘", "│"}

// marca: texto de la propuesta si la línea empieza con PROPUESTA: (admite viñeta, título, cita o negrita).
func marca(l string) (string, bool) {
	r, ok := strings.CutPrefix(strings.TrimLeft(l, " \t*#>-"), "PROPUESTA:")
	if !ok {
		return "", false
	}
	r = strings.TrimSpace(strings.TrimLeft(r, "* "))
	if c := []rune(r); len(c) > 60 {
		r = string(c[:60]) + "…"
	}
	return r, true
}

// diagrama: el texto trae algún trazo de diagrama (flechas o cajas).
func diagrama(c string) bool {
	for _, t := range trazos {
		if strings.Contains(c, t) {
			return true
		}
	}
	return false
}

// cercaDe: el carácter y el largo de una cerca de bloque (``` o ~~~), o largo 0 si no es cerca.
func cercaDe(t string) (byte, int) {
	if len(t) < 3 || (t[0] != '`' && t[0] != '~') {
		return 0, 0
	}
	n := 0
	for n < len(t) && t[n] == t[0] {
		n++
	}
	if n < 3 {
		return 0, 0
	}
	return t[0], n
}

// Revisar devuelve "" si pasa; si no, el motivo del freno con las propuestas sin diagrama.
func Revisar(msg string) string {
	var faltan []string
	pendiente, hay := "", false
	var abierta byte
	largo := 0
	var bloque strings.Builder
	for _, l := range strings.Split(strings.ReplaceAll(msg, "\r\n", "\n"), "\n") {
		t := strings.TrimSpace(l)
		if abierta == 0 {
			if c, n := cercaDe(t); n > 0 {
				abierta, largo = c, n
				bloque.Reset()
				continue
			}
			if p, ok := marca(l); ok {
				if hay {
					faltan = append(faltan, pendiente)
				}
				pendiente, hay = p, true
			}
			continue
		}
		if c, n := cercaDe(t); c == abierta && n >= largo && strings.Trim(t, string(c)) == "" {
			if hay && diagrama(bloque.String()) {
				hay = false
			}
			abierta = 0
			continue
		}
		bloque.WriteString(l + "\n")
	}
	if hay {
		faltan = append(faltan, pendiente)
	}
	if len(faltan) == 0 {
		return ""
	}
	return "PROPUESTA sin diagrama: «" + strings.Join(faltan, "», «") + "». Agrega después de cada una un bloque de código" +
		" con flechas (→, ->, ─►) o cajas (┌ │ └): el chat no dibuja Mermaid, el diagrama va en texto."
}
