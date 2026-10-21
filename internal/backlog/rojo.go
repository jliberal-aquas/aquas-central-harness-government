package backlog

// E001/F002/S003: pruebas con intención en la ficha; rojo prueba por prueba con `go test -json`;
// verde después; un solo slice en curso.

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Prueba: nombre "./paquete:TestNombre" e intención (qué demuestra).
type Prueba struct {
	Nombre    string `json:"prueba"`
	Intencion string `json:"intencion"`
}

// Registro: una corrida de las pruebas de la ficha. Resultado: prueba → primera línea de su fallo (rojo).
type Registro struct {
	Fecha     string            `json:"fecha"`
	Head      string            `json:"head"`
	Resultado map[string]string `json:"resultado,omitempty"`
}

var reNombrePrueba = regexp.MustCompile(`^\./[\w.-]+(/[\w.-]+)*:Test([A-Z0-9_]\w*)?$`)

const formatoPrueba = "./paquete:TestNombre=intención"

// problemaPrueba: por qué una prueba de la ficha está mal escrita ("" si está bien).
func problemaPrueba(p Prueba) string {
	if !reNombrePrueba.MatchString(p.Nombre) {
		return fmt.Sprintf("prueba %q: usa %s", p.Nombre, formatoPrueba)
	}
	if strings.TrimSpace(p.Intencion) == "" {
		return fmt.Sprintf("prueba %s sin intención: di qué demuestra (%s)", p.Nombre, formatoPrueba)
	}
	return ""
}

// listaPruebas convierte "./paquete:TestNombre=intención" en pruebas válidas, sin repetidas.
func listaPruebas(xs []string) ([]Prueba, error) {
	if len(xs) == 0 {
		return nil, fmt.Errorf("falta al menos una prueba: prueba=%s", formatoPrueba)
	}
	var ps []Prueba
	for _, x := range xs {
		n, i, _ := strings.Cut(x, "=")
		p := Prueba{strings.TrimSpace(n), strings.TrimSpace(i)}
		if m := problemaPrueba(p); m != "" {
			return nil, errors.New(m)
		}
		for _, q := range ps {
			if q.Nombre == p.Nombre {
				return nil, fmt.Errorf("prueba %s repetida", p.Nombre)
			}
		}
		ps = append(ps, p)
	}
	return ps, nil
}

func corriendo(f Ficha) bool { return f.Estado == "EN_CURSO" || f.Estado == "VERIFICANDO" }

// sliceEnCurso: el primer slice EN_CURSO o VERIFICANDO entre las activas, por id.
func sliceEnCurso(act map[string]Ficha) (string, bool) {
	var ids []string
	for id, f := range act {
		if f.Tipo == "slice" && corriendo(f) {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	if len(ids) == 0 {
		return "", false
	}
	return ids[0], true
}

// SliceEnCurso: el único slice EN_CURSO o VERIFICANDO, si hay.
func SliceEnCurso(raiz string) (string, bool, error) {
	act, _, err := todas(raiz)
	if err != nil {
		return "", false, err
	}
	id, hay := sliceEnCurso(act)
	return id, hay, nil
}

// Pruebas fija las pruebas de la ficha; cambiarlas borra rojo y verde.
func Pruebas(raiz, id string, lista []string) (Ficha, error) {
	ps, err := listaPruebas(lista)
	if err != nil {
		return Ficha{}, err
	}
	f, err := activa(raiz, id)
	if err != nil {
		return Ficha{}, err
	}
	f.Pruebas, f.Rojo, f.Verde = ps, nil, nil
	return f, escribir(raiz, Activo, f)
}

// corrible: ficha activa con pruebas; un slice además debe estar en curso.
func corrible(raiz, id string) (Ficha, error) {
	f, err := activa(raiz, id)
	if err != nil {
		return Ficha{}, err
	}
	if f.Tipo == "slice" && !corriendo(f) {
		return Ficha{}, fmt.Errorf("%s no está en curso (%s): rojo y verde son del slice en curso", id, f.Estado)
	}
	if len(f.Pruebas) == 0 {
		return Ficha{}, fmt.Errorf("%s sin pruebas: aquas backlog pruebas %s %s", id, id, formatoPrueba)
	}
	return f, nil
}

// Rojo corre cada prueba sola con `go test -json`; exige que cada una falle por sí misma.
func Rojo(raiz, id string, ahora time.Time) (Ficha, error) {
	f, err := corrible(raiz, id)
	if err != nil {
		return Ficha{}, err
	}
	res := map[string]string{}
	var malas []string
	for _, p := range f.Pruebas {
		r, err := correr(raiz, []Prueba{p})
		if err != nil {
			return Ficha{}, err
		}
		if x := r[p.Nombre]; x.accion == "fail" {
			res[p.Nombre] = x.salida
		} else {
			malas = append(malas, fmt.Sprintf("%s: %s", p.Nombre, x.motivo("corrida sola, pasa: no falla por sí misma")))
		}
	}
	if len(malas) > 0 {
		return Ficha{}, fmt.Errorf("rojo rechazado; cada prueba debe fallar sola por su motivo:\n  - %s", strings.Join(malas, "\n  - "))
	}
	f.Rojo, f.Verde = &Registro{Fecha: ahora.UTC().Format(time.RFC3339), Head: head(raiz), Resultado: res}, nil
	return f, escribir(raiz, Activo, f)
}

// Verde corre las pruebas juntas tras el rojo; exige que todas pasen.
func Verde(raiz, id string, ahora time.Time) (Ficha, error) {
	f, err := corrible(raiz, id)
	if err != nil {
		return Ficha{}, err
	}
	if f.Rojo == nil {
		return Ficha{}, fmt.Errorf("%s sin rojo: primero aquas backlog rojo %s", id, id)
	}
	r, err := correr(raiz, f.Pruebas)
	if err != nil {
		return Ficha{}, err
	}
	var malas []string
	for _, p := range f.Pruebas {
		if x := r[p.Nombre]; x.accion != "pass" {
			malas = append(malas, fmt.Sprintf("%s: %s", p.Nombre, x.motivo("")))
		}
	}
	if len(malas) > 0 {
		return Ficha{}, fmt.Errorf("verde rechazado; todas deben pasar:\n  - %s", strings.Join(malas, "\n  - "))
	}
	f.Verde = &Registro{Fecha: ahora.UTC().Format(time.RFC3339), Head: head(raiz)}
	return f, escribir(raiz, Activo, f)
}

// resultado de una prueba en `go test -json`. accion "" = no hubo evento de la prueba.
type resultado struct {
	accion, salida        string
	lineas                []string
	noCompila, sinPaquete bool
}

func (r resultado) motivo(siPasa string) string {
	switch {
	case r.accion == "pass" && siPasa != "":
		return siPasa
	case r.accion == "pass":
		return "pasó"
	case r.accion == "fail":
		return "falló: " + r.salida
	case r.accion == "skip":
		return "se saltó: no cuenta"
	case r.sinPaquete:
		return "paquete no existe: no cuenta"
	case r.noCompila:
		return "no compila: no cuenta"
	}
	return "no existe en el paquete"
}

// correr: `go test -json -count=1 -run ^(A|B)$ ./paquete` por paquete; resultado por prueba.
func correr(raiz string, ps []Prueba) (map[string]resultado, error) {
	porPaquete := map[string][]string{}
	var paquetes []string
	for _, p := range ps {
		pk, n, _ := strings.Cut(p.Nombre, ":")
		if _, ya := porPaquete[pk]; !ya {
			paquetes = append(paquetes, pk)
		}
		porPaquete[pk] = append(porPaquete[pk], n)
	}
	out := map[string]resultado{}
	for _, pk := range paquetes {
		c := exec.Command("go", "test", "-json", "-count=1", "-run", "^("+strings.Join(porPaquete[pk], "|")+")$", pk)
		c.Dir = raiz
		var o bytes.Buffer
		c.Stdout, c.Stderr = &o, &o
		if err := c.Run(); err != nil {
			if _, salio := err.(*exec.ExitError); !salio {
				return nil, fmt.Errorf("no se pudo correr go test: %v", err)
			}
		}
		texto := o.String()
		noCompila := strings.Contains(texto, "[build failed]") || strings.Contains(texto, "[setup failed]") || strings.Contains(texto, `"Action":"build-fail"`)
		sinPaquete := strings.Contains(texto, "directory not found") || strings.Contains(texto, "cannot find") ||
			strings.Contains(texto, "no Go files") || strings.Contains(texto, "is not in std")
		porPrueba := map[string]*resultado{}
		for _, n := range porPaquete[pk] {
			porPrueba[n] = &resultado{noCompila: noCompila, sinPaquete: sinPaquete}
		}
		sc := bufio.NewScanner(strings.NewReader(texto))
		sc.Buffer(make([]byte, 1<<20), 1<<20)
		for sc.Scan() {
			var e struct{ Action, Test, Output string }
			if json.Unmarshal(sc.Bytes(), &e) != nil {
				continue
			}
			r := porPrueba[e.Test]
			if r == nil {
				continue
			}
			switch e.Action {
			case "pass", "fail", "skip":
				r.accion = e.Action
			case "output":
				// Motivo = la línea del fallo (Fatal/Error dejan «archivo_test.go:N:»), no un t.Log previo.
				l := strings.TrimSpace(e.Output)
				if l != "" && !strings.HasPrefix(l, "=== ") && !strings.HasPrefix(l, "--- ") {
					r.lineas = append(r.lineas, l)
				}
			}
		}
		for n, r := range porPrueba {
			r.salida = motivoDelFallo(r.lineas)
			out[pk+":"+n] = *r
		}
	}
	return out, nil
}

// motivoDelFallo: la última línea «archivo.go:N: …» (la que hizo fallar); si no hay, la primera.
func motivoDelFallo(ls []string) string {
	for i := len(ls) - 1; i >= 0; i-- {
		if reLineaGo.MatchString(ls[i]) {
			return ls[i]
		}
	}
	if len(ls) > 0 {
		return ls[0]
	}
	return ""
}

var reLineaGo = regexp.MustCompile(`^\S+\.go:\d+: `)

// head: commit actual del repo, o "" fuera de Git.
func head(raiz string) string {
	b, err := exec.Command("git", "-C", raiz, "rev-parse", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

// cierre: CERRADO exige pruebas, rojo y verde.
func cierre(f Ficha) error {
	switch {
	case len(f.Pruebas) == 0:
		return fmt.Errorf("%s no cierra sin pruebas: aquas backlog pruebas %s %s", f.ID, f.ID, formatoPrueba)
	case f.Rojo == nil:
		return fmt.Errorf("%s no cierra sin rojo: aquas backlog rojo %s", f.ID, f.ID)
	case f.Verde == nil:
		return fmt.Errorf("%s no cierra sin verde: aquas backlog verde %s", f.ID, f.ID)
	}
	return nil
}

// problemasRojoVerde: chequeos del gate sobre pruebas, rojo y verde de una ficha.
func problemasRojoVerde(rel string, f Ficha) []string {
	var p []string
	vistas := map[string]bool{}
	for _, x := range f.Pruebas {
		if m := problemaPrueba(x); m != "" {
			p = append(p, rel+": "+m)
		}
		if vistas[x.Nombre] {
			p = append(p, fmt.Sprintf("%s: prueba %s repetida", rel, x.Nombre))
		}
		vistas[x.Nombre] = true
	}
	if f.Verde != nil && f.Rojo == nil {
		p = append(p, rel+": verde sin rojo")
	}
	if f.Rojo != nil {
		var ns []string
		for n := range f.Rojo.Resultado {
			if !vistas[n] {
				ns = append(ns, n)
			}
		}
		sort.Strings(ns)
		for _, n := range ns {
			p = append(p, fmt.Sprintf("%s: el rojo registra %s, que no está en sus pruebas", rel, n))
		}
		if f.Verde != nil && f.Verde.Fecha < f.Rojo.Fecha {
			p = append(p, rel+": verde antes del rojo")
		}
	}
	if f.Estado == "CERRADO" && len(f.Pruebas) > 0 {
		if f.Rojo == nil {
			p = append(p, rel+": cerrada sin rojo")
		}
		if f.Verde == nil {
			p = append(p, rel+": cerrada sin verde")
		}
	}
	return p
}

// Detalle: pruebas con su intención, y rojo y verde, para `ver <id>`, `pruebas`, `rojo` y `verde`.
func Detalle(f Ficha) string {
	var b strings.Builder
	for _, p := range f.Pruebas {
		fmt.Fprintf(&b, "  prueba %s · %s\n", p.Nombre, p.Intencion)
		if f.Rojo != nil && f.Rojo.Resultado[p.Nombre] != "" {
			fmt.Fprintf(&b, "    rojo: %s\n", f.Rojo.Resultado[p.Nombre])
		}
	}
	marca := func(r *Registro) string {
		if r == nil {
			return "—"
		}
		return r.Fecha
	}
	fmt.Fprintf(&b, "  rojo %s · verde %s\n", marca(f.Rojo), marca(f.Verde))
	return b.String()
}
