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
	leidos := map[string][]string{} // cada archivo se lee y se cuenta una sola vez
	for _, ruta := range p.Rutas {
		rel, desde, hasta := ruta, 0, 0
		if m := reRango.FindStringSubmatch(ruta); m != nil {
			rel = m[1]
			desde, _ = strconv.Atoi(m[2])
			hasta, _ = strconv.Atoi(m[3])
			if desde < 1 || hasta < desde {
				escribir("== " + ruta + " · error: rango inválido (desde-hasta, desde ≥ 1)\n")
				continue
			}
		}
		if !filepath.IsLocal(filepath.FromSlash(rel)) {
			escribir("== " + ruta + " · error: fuera del repo (solo rutas relativas a la raíz)\n")
			continue
		}
		rel = filepath.ToSlash(filepath.Clean(filepath.FromSlash(rel))) // «./a.go» y «a.go» = mismo archivo
		lineas, ok := leidos[rel]
		if !ok {
			abs := filepath.Join(raiz, filepath.FromSlash(rel))
			if info, err := os.Stat(abs); err == nil && info.IsDir() {
				escribir("== " + ruta + " · error: es una carpeta\n")
				continue
			}
			datos, err := os.ReadFile(abs)
			if err != nil {
				motivo := "no se pudo leer" // sin la ruta absoluta ni el mensaje del SO
				if errors.Is(err, os.ErrNotExist) {
					motivo = "no existe"
				}
				escribir("== " + ruta + " · error: " + motivo + "\n")
				continue
			}
			completos += utf8.RuneCount(datos)
			s := strings.TrimSuffix(strings.ReplaceAll(string(datos), "\r\n", "\n"), "\n")
			lineas = []string{}
			if s != "" {
				lineas = strings.Split(s, "\n")
			}
			leidos[rel] = lineas
		}
		if len(lineas) == 0 {
			escribir("== " + rel + " · vacío (0 líneas)\n")
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
		var blq strings.Builder
		fmt.Fprintf(&blq, "== %s · %d-%d de %d\n", rel, desde, hasta, len(lineas))
		for i := desde; i <= hasta; i++ {
			fmt.Fprintf(&blq, "%d\t%s\n", i, lineas[i-1])
		}
		escribir(blq.String())
	}
	fmt.Fprintf(&b, "== devueltos %d de %d caracteres\n", devueltos, completos)
	return b.String(), devueltos, completos
}
