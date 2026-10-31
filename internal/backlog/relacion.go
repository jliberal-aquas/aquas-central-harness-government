package backlog

// E001/F009/S001: toda ficha nueva declara su relación con lo existente; sin ella, backlog new
// se niega y muestra las fichas y defectos parecidos. Ampliar o corregir algo cerrado se niega.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

// maxParecidas: cuántas parecidas muestra backlog new.
const maxParecidas = 5

// usoRelacion: forma válida del campo relacion.
const usoRelacion = "relacion=nueva | amplia:<id> | corrige:<id>"

// palabras: conjunto de palabras de tres o más letras, en minúscula.
func palabras(s string) map[string]bool {
	ws := map[string]bool{}
	for _, w := range strings.FieldsFunc(strings.ToLower(s), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) }) {
		if len([]rune(w)) >= 3 {
			ws[w] = true
		}
	}
	return ws
}

// parecida: una ficha o defecto con palabras en común con el título nuevo.
type parecida struct {
	id, texto string
	comunes   int
}

// parecidas: hasta maxParecidas fichas (activas y archivadas) o defectos con más palabras en común con titulo.
func parecidas(raiz, titulo string, fichas ...map[string]Ficha) []parecida {
	nuevo := palabras(titulo)
	var ps []parecida
	sumar := func(id, texto string) {
		n := 0
		for w := range palabras(texto) {
			if nuevo[w] {
				n++
			}
		}
		if n > 0 {
			ps = append(ps, parecida{id, texto, n})
		}
	}
	for _, m := range fichas {
		for id, f := range m {
			sumar(id, f.Titulo)
		}
	}
	rutas, _ := filepath.Glob(filepath.Join(raiz, "registro", "defectos", "D*.json"))
	for _, r := range rutas {
		var d struct{ ID, Que string }
		if b, err := os.ReadFile(r); err == nil && json.Unmarshal(b, &d) == nil {
			sumar(d.ID, d.Que)
		}
	}
	sort.Slice(ps, func(i, j int) bool {
		if ps[i].comunes != ps[j].comunes {
			return ps[i].comunes > ps[j].comunes
		}
		return ps[i].id < ps[j].id
	})
	if len(ps) > maxParecidas {
		ps = ps[:maxParecidas]
	}
	return ps
}

// revisarRelacion valida el campo relacion contra las fichas existentes; sin él, lista las parecidas.
func revisarRelacion(raiz, rel, titulo string, act, arc map[string]Ficha) error {
	if rel == "" {
		var b strings.Builder
		fmt.Fprintf(&b, "falta %s; revisa si ya existe:", usoRelacion)
		ps := parecidas(raiz, titulo, act, arc)
		if len(ps) == 0 {
			b.WriteString(" sin parecidas")
		}
		for _, p := range ps {
			fmt.Fprintf(&b, "\n  - %s · %s", p.id, p.texto)
		}
		return fmt.Errorf("%s", b.String())
	}
	if rel == "nueva" {
		return nil
	}
	tipo, id, ok := strings.Cut(rel, ":")
	if !ok || (tipo != "amplia" && tipo != "corrige") {
		return fmt.Errorf("relacion %q inválida: %s", rel, usoRelacion)
	}
	if _, enArc := arc[id]; enArc {
		if tipo == "amplia" {
			return fmt.Errorf("%s está cerrada: lo nuevo va en una feature nueva que la nombre (relacion=nueva)", id)
		}
		return fmt.Errorf("%s está cerrada: corrígela con aquas backlog fix %s titulo=… motivo=…", id, id)
	}
	if _, enAct := act[id]; !enAct {
		return fmt.Errorf("relacion: %s no existe", id)
	}
	return nil
}

// Parecidas: líneas «id · texto» de las fichas y defectos con más palabras en común con texto (E001/F009/S002).
func Parecidas(raiz, texto string) ([]string, error) {
	act, arc, err := todas(raiz)
	if err != nil {
		return nil, err
	}
	var ls []string
	for _, p := range parecidas(raiz, texto, act, arc) {
		ls = append(ls, p.id+" · "+p.texto)
	}
	return ls, nil
}
