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

// Prueba: nombre "./paquete:TestNombre", qué valida (Intencion) y cuándo queda en rojo y en verde.
type Prueba struct {
	Nombre      string `json:"prueba"`
	Intencion   string `json:"intencion"`
	RojoCuando  string `json:"rojo_cuando,omitempty"`
	VerdeCuando string `json:"verde_cuando,omitempty"`
}

// Registro: una corrida de las pruebas de la ficha. Resultado: prueba → primera línea de su fallo (rojo).
type Registro struct {
	Fecha     string            `json:"fecha"`
	Head      string            `json:"head"`
	Resultado map[string]string `json:"resultado,omitempty"`
	// X02: slice cerrado del que una feature heredó este rojo; "" = rojo propio.
	Origen string `json:"origen,omitempty"`
}

var reNombrePrueba = regexp.MustCompile(`^\./[\w.-]+(/[\w.-]+)*:Test([A-Z0-9_]\w*)?$`)

const formatoPrueba = "./paquete:TestNombre=valida|rojo cuando|verde cuando"

// problemaPrueba: por qué una prueba de la ficha está mal escrita ("" si está bien).
func problemaPrueba(p Prueba) string {
	if !reNombrePrueba.MatchString(p.Nombre) {
		return fmt.Sprintf("prueba %q: usa %s", p.Nombre, formatoPrueba)
	}
	if strings.TrimSpace(p.Intencion) == "" {
		return fmt.Sprintf("prueba %s sin intención: di qué demuestra (%s)", p.Nombre, formatoPrueba)
	}
	if p.RojoCuando == "" || p.VerdeCuando == "" { // E001/F008/S007: sin excepción para pruebas viejas
		return fmt.Sprintf("prueba %s sin «rojo cuando» o «verde cuando»: aquas backlog documentar (%s)", p.Nombre, formatoPrueba)
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
		if k := strings.Count(i, "|") + 1; k > 3 { // D000024: el sobrante no se pega a «verde cuando»
			return nil, fmt.Errorf("prueba %s con %d partes separadas por |: deben ser 3 (valida|rojo cuando|verde cuando); escribe «o» en vez de | dentro del texto", strings.TrimSpace(n), k)
		}
		c := append(strings.SplitN(i, "|", 3), "", "")
		p := Prueba{Nombre: strings.TrimSpace(n), Intencion: strings.TrimSpace(c[0]), RojoCuando: strings.TrimSpace(c[1]), VerdeCuando: strings.TrimSpace(c[2])}
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

// corriendo: la ficha está EN_CURSO o VERIFICANDO.
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
	if _, ok := padreDeArreglo(id); ok { // D000020
		return pruebasArreglo(raiz, id, ps)
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
	if _, ok := padreDeArreglo(id); ok { // D000020
		return rojoArreglo(raiz, id, ahora)
	}
	if f0, err := activa(raiz, id); err == nil { // E001/F010/S001: el contrato se aprueba antes del rojo
		if err := aprobada(raiz, f0); err != nil {
			return Ficha{}, err
		}
	}
	f, err := corrible(raiz, id)
	if err != nil {
		return Ficha{}, err
	}
	r, err := rojoDe(raiz, f.Pruebas, ahora)
	if err != nil && f.Tipo == "feature" {
		if h := rojoHeredado(raiz, f); h != nil {
			r, err = h, nil
		}
	}
	if err != nil {
		return Ficha{}, err
	}
	f.Rojo, f.Verde = r, nil
	return f, escribir(raiz, Activo, f)
}

// rojoHeredado (X02): rojo que un slice cerrado de la feature registró para todas sus pruebas
// integrales; nil si ninguno las respalda. La integral que ya pasó no vuelve a caer: este rojo es su evidencia.
func rojoHeredado(raiz string, f Ficha) *Registro {
	_, arc, err := todas(raiz)
	if err != nil {
		return nil
	}
	for _, s := range arc {
		if s.Tipo != "slice" || s.Padre != f.ID || s.Rojo == nil || s.Estado != "CERRADO" { // D000032: CANCELADO no respalda
			continue
		}
		res := map[string]string{}
		for _, p := range f.Pruebas {
			if v, ok := s.Rojo.Resultado[p.Nombre]; ok {
				res[p.Nombre] = v
			}
		}
		if len(res) == len(f.Pruebas) {
			return &Registro{Fecha: s.Rojo.Fecha, Head: s.Rojo.Head, Resultado: res, Origen: s.ID}
		}
	}
	return nil
}

// Verde corre las pruebas juntas tras el rojo; exige que todas pasen.
func Verde(raiz, id string, ahora time.Time) (Ficha, error) {
	if _, ok := padreDeArreglo(id); ok { // D000020
		return verdeArreglo(raiz, id, ahora)
	}
	if f0, err := activa(raiz, id); err == nil { // E001/F010/S001: el verde compara la huella aprobada
		if err := aprobada(raiz, f0); err != nil {
			return Ficha{}, err
		}
	}
	f, err := corrible(raiz, id)
	if err != nil {
		return Ficha{}, err
	}
	r, err := verdeDe(raiz, id, f.Pruebas, f.Rojo, ahora)
	if err != nil {
		return Ficha{}, err
	}
	f.Verde = r
	return f, escribir(raiz, Activo, f)
}

// resultado de una prueba en `go test -json`. accion "" = no hubo evento de la prueba.
type resultado struct {
	accion, salida        string
	lineas                []string
	noCompila, sinPaquete bool
}

// motivo: texto que explica el resultado de una prueba; siPasa se usa cuando pasó.
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
