// Package biblioteca: el bibliotecario (E001/F003). Lee un lote de rutas del repo en una
// sola llamada y entrega solo lo pedido, numerado, con tope y cuenta de ahorro.
package biblioteca

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// TopeDefecto: freno de desborde extremo, no meta (constitución: topes con holgura amplia).
const TopeDefecto = 200000

// Pedido: lo que llega por stdin a `aquas biblioteca leer`.
//
//	{"rutas": ["internal/x.go", "docs/README.md:10-40"], "tope": 0}
type Pedido struct {
	Rutas []string `json:"rutas"`
	Tope  int      `json:"tope"` // 0 = TopeDefecto
}

// reRango: «ruta:desde-hasta» al final de la ruta.
var reRango = regexp.MustCompile(`^(.+):(\d+)-(\d+)$`)

// Leer devuelve el texto para el agente y cuántos caracteres devolvió de cuántos tiene
// la lectura completa de los mismos archivos. Un error en una ruta no frena el lote.
func Leer(raiz string, p Pedido) (texto string, devueltos, completos int) {
	tope := p.Tope
	if tope <= 0 {
		tope = TopeDefecto
	}
	var b strings.Builder
	cortado := false
	escribir := func(s string) {
		if cortado {
			return
		}
		if n := utf8.RuneCountInString(s); devueltos+n > tope {
			r := []rune(s)
			b.WriteString(string(r[:tope-devueltos]))
			devueltos = tope
			cortado = true
			fmt.Fprintf(&b, "\n[cortado por tope: %d caracteres; pide rangos más chicos]\n", tope)
			return
		}
		b.WriteString(s)
		devueltos += utf8.RuneCountInString(s)
	}
	// avisar: línea de error; la lectura entera también la traería, así que cuenta en completos,
	// una vez por línea distinta, igual que cada archivo (hallazgo A1.4).
	avisos := map[string]bool{}
	avisar := func(s string) {
		if !avisos[s] {
			avisos[s] = true
			completos += utf8.RuneCountInString(s)
		}
		escribir(s)
	}
	leidos := map[string][]string{} // cada archivo se lee y se cuenta una sola vez
	for _, ruta := range p.Rutas {
		rel, desde, hasta := ruta, 0, 0
		if m := reRango.FindStringSubmatch(ruta); m != nil {
			rel = m[1]
			desde, _ = strconv.Atoi(m[2])
			hasta, _ = strconv.Atoi(m[3])
			if desde < 1 || hasta < desde {
				avisar("== " + ruta + " · error: rango inválido (desde-hasta, desde ≥ 1)\n")
				continue
			}
		}
		if !filepath.IsLocal(filepath.FromSlash(rel)) {
			avisar("== " + ruta + " · error: fuera del repo (solo rutas relativas a la raíz)\n")
			continue
		}
		rel = filepath.ToSlash(filepath.Clean(filepath.FromSlash(rel))) // «./a.go» y «a.go» = mismo archivo
		lineas, ok := leidos[rel]
		if !ok {
			abs := filepath.Join(raiz, filepath.FromSlash(rel))
			if info, err := os.Stat(abs); err == nil && info.IsDir() {
				avisar("== " + ruta + " · error: es una carpeta\n")
				continue
			}
			datos, err := os.ReadFile(abs)
			if err != nil {
				motivo := "no se pudo leer" // sin la ruta absoluta ni el mensaje del SO
				if errors.Is(err, os.ErrNotExist) {
					motivo = "no existe"
				}
				avisar("== " + ruta + " · error: " + motivo + "\n")
				continue
			}
			s := strings.TrimSuffix(strings.ReplaceAll(string(datos), "\r\n", "\n"), "\n")
			lineas = []string{}
			if s != "" {
				lineas = strings.Split(s, "\n")
			}
			completos += utf8.RuneCountInString(bloque(rel, lineas, 1, len(lineas))) // mismo formato que lo devuelto (S007)
			leidos[rel] = lineas
		}
		if len(lineas) == 0 {
			escribir(bloque(rel, lineas, 0, 0))
			continue
		}
		if desde == 0 {
			desde, hasta = 1, len(lineas)
		}
		if hasta > len(lineas) {
			hasta = len(lineas)
		}
		if desde > len(lineas) {
			escribir(fmt.Sprintf("== %s · error: el archivo tiene %d líneas\n", ruta, len(lineas)))
			continue
		}
		escribir(bloque(rel, lineas, desde, hasta))
	}
	fmt.Fprintf(&b, "== devueltos %d de %d caracteres", devueltos, completos)
	if devueltos > completos { // hallazgos A1.1 y A2.1: muchos bloques chicos o repetidos cuestan más que leer entero
		b.WriteString(" (más que leer entero de una vez: cada rango o aviso suma su cabecera y lo repetido cuenta doble)")
	}
	b.WriteString("\n")
	return b.String(), devueltos, completos
}

// bloque: lo que Leer devuelve de las líneas desde-hasta de un archivo, cabecera incluida.
// Sirve también para contar completos en el mismo formato (E000/F006/S007).
func bloque(rel string, lineas []string, desde, hasta int) string {
	if len(lineas) == 0 {
		return "== " + rel + " · vacío (0 líneas)\n"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "== %s · %d-%d de %d\n", rel, desde, hasta, len(lineas))
	for i := desde; i <= hasta; i++ {
		fmt.Fprintf(&b, "%d\t%s\n", i, lineas[i-1])
	}
	return b.String()
}
