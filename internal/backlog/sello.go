package backlog

// E001/F002/S004: sello de cierre. Antes de cerrar un slice, aquas arma la lista de lo que pudo
// quedar desfasado y el cierre se niega hasta que cada ítem tenga marca (sin_cambio | actualizado)
// y motivo de una línea. Dos niveles:
//   - Árbol: feature padre, iniciativa, dependientes (fichas abiertas con bloqueado_por este slice)
//     y las páginas wiki del slice, la feature y la iniciativa. Un hermano sin dependencia no entra.
//   - Archivos: tocados = `git diff` desde el HEAD del rojo + archivos nuevos. Por cada archivo
//     del repo que nombra la ruta exacta de un tocado, o que importa (Go) su paquete, un ítem con
//     la línea relevante. Se excluyen los datos: trabajo/, registro/, auditoria/, .tmp/, .aquas/.
// La lista se arma de nuevo en cada consulta: tocar otro archivo agrega ítems sin marca.
// Sin HEAD en el rojo (repo sin commits o sin Git) se exige el árbol y un aviso; solo se omite el nivel archivos (S013).
// El nivel trabajo (otras iniciativas que comparten archivos o datos) es F002/S012.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/compilador"
)

// Marca: la revisión de un ítem del sello.
type Marca struct {
	Clave  string `json:"clave"`
	Marca  string `json:"marca"`
	Motivo string `json:"motivo"`
}

// Item: algo que revisar antes de cerrar un slice; Marca nil = pendiente.
type Item struct {
	Clave, Linea string
	Marca        *Marca
}

var excluidas = []string{"trabajo/", "registro/", "auditoria/", ".tmp/", ".aquas/"}

func excluida(p string) bool {
	for _, e := range excluidas {
		if strings.HasPrefix(p, e) {
			return true
		}
	}
	return false
}

func gitLineas(raiz string, args ...string) ([]string, error) {
	out, err := exec.Command("git", append([]string{"-C", raiz, "-c", "core.quotepath=off"}, args...)...).Output()
	if err != nil {
		return nil, fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	var xs []string
	for _, l := range strings.Split(strings.ReplaceAll(string(out), "\r\n", "\n"), "\n") {
		if l = strings.TrimSpace(l); l != "" {
			xs = append(xs, l)
		}
	}
	return xs, nil
}

func pagina(id string) string {
	p := strings.Split(id, "/")
	if len(p) == 3 {
		return "docs/" + p[0] + "/" + p[1] + "/" + p[2] + ".md"
	}
	return "docs/" + strings.Join(p, "/") + "/README.md"
}

// arbol: padre, abuelo, dependientes abiertos y páginas.
func arbol(raiz string, f Ficha) ([]Item, error) {
	var xs []Item
	if f.Padre != "" {
		xs = append(xs, Item{Clave: "ficha:" + f.Padre, Linea: "feature padre"})
		if e, _ := tipoYPadre(f.Padre); e == "feature" {
			_, abuelo := tipoYPadre(f.Padre)
			xs = append(xs, Item{Clave: "ficha:" + abuelo, Linea: "iniciativa"})
		}
	}
	act, _, err := todas(raiz)
	if err != nil {
		return nil, err
	}
	var deps []string
	for _, a := range act {
		if esta(a.BloqueadoPor, f.ID) {
			deps = append(deps, a.ID)
		}
	}
	sort.Strings(deps)
	for _, d := range deps {
		xs = append(xs, Item{Clave: "ficha:" + d, Linea: "dependiente: " + act[d].Titulo})
	}
	for id := f.ID; id != ""; _, id = tipoYPadre(id) {
		xs = append(xs, Item{Clave: "pagina:" + pagina(id), Linea: "página de " + id})
	}
	return xs, nil
}

func moduloGo(raiz string) string {
	b, err := os.ReadFile(filepath.Join(raiz, "go.mod"))
	if err != nil {
		return ""
	}
	for _, l := range strings.Split(string(b), "\n") {
		if m, ok := strings.CutPrefix(strings.TrimSpace(l), "module "); ok {
			return strings.TrimSpace(m)
		}
	}
	return ""
}

func lineaCon(c, s string) string {
	for _, l := range strings.Split(c, "\n") {
		if strings.Contains(l, s) {
			l = strings.TrimSpace(l)
			if r := []rune(l); len(r) > 120 {
				l = string(r[:120]) + "…"
			}
			return l
		}
	}
	return ""
}

// derivados: salidas del compilador según su archivo de sellos, más ese archivo (S014). Ya los vigila
// el gate de derivados: revisarlos en el sello solo mete ruido. Sin archivo de sellos: ninguno.
func derivados(raiz string) map[string]bool {
	d := map[string]bool{compilador.RutaSellos: true}
	b, err := os.ReadFile(filepath.Join(raiz, filepath.FromSlash(compilador.RutaSellos)))
	if err != nil {
		return d
	}
	var s struct {
		Salidas map[string]json.RawMessage `json:"salidas"`
	}
	if json.Unmarshal(b, &s) == nil {
		for p := range s.Salidas {
			d[p] = true
		}
	}
	return d
}

// archivos: quién nombra o importa lo tocado desde el HEAD del rojo.
func archivos(raiz, head string) ([]Item, error) {
	cambiados, err := gitLineas(raiz, "diff", "--name-only", head)
	if err != nil {
		return nil, err
	}
	nuevos, err := gitLineas(raiz, "ls-files", "-o", "--exclude-standard")
	if err != nil {
		return nil, err
	}
	todos, err := gitLineas(raiz, "ls-files", "-co", "--exclude-standard")
	if err != nil {
		return nil, err
	}
	tocado := map[string]bool{}
	var tocados []string
	paquetes := map[string]bool{}
	der := derivados(raiz)
	for _, p := range append(cambiados, nuevos...) {
		if excluida(p) || tocado[p] || der[p] {
			continue
		}
		tocado[p] = true
		tocados = append(tocados, p)
		if strings.HasSuffix(p, ".go") {
			paquetes[path.Dir(p)] = true
		}
	}
	sort.Strings(tocados)
	var pkgs []string
	for p := range paquetes {
		pkgs = append(pkgs, p)
	}
	sort.Strings(pkgs)
	mod := moduloGo(raiz)
	var xs []Item
	for _, a := range todos {
		if excluida(a) || tocado[a] {
			continue
		}
		b, err := os.ReadFile(filepath.Join(raiz, filepath.FromSlash(a)))
		if err != nil || len(b) > 2<<20 || strings.ContainsRune(string(b), 0) {
			continue // borrado, enorme o binario
		}
		c := string(b)
		for _, t := range tocados {
			if strings.Contains(c, t) {
				xs = append(xs, Item{Clave: "nombra:" + a + ">" + t, Linea: lineaCon(c, t)})
			}
		}
		if mod != "" && strings.HasSuffix(a, ".go") {
			for _, p := range pkgs {
				imp := `"` + mod + "/" + p + `"`
				if strings.Contains(c, imp) {
					xs = append(xs, Item{Clave: "importa:" + a + ">" + p, Linea: lineaCon(c, imp)})
				}
			}
		}
	}
	return xs, nil
}

// Sello arma la lista del slice y le pega las marcas guardadas en su ficha.
func Sello(raiz, id string) ([]Item, error) {
	f, _, err := Ver(raiz, id)
	if err != nil {
		return nil, err
	}
	return sello(raiz, f)
}

func sello(raiz string, f Ficha) ([]Item, error) {
	if f.Tipo != "slice" {
		return nil, fmt.Errorf("%s: el sello es de slices", f.ID)
	}
	xs, err := arbol(raiz, f)
	if err != nil {
		return nil, err
	}
	if f.Rojo != nil && f.Rojo.Head != "" {
		ys, err := archivos(raiz, f.Rojo.Head)
		if err != nil {
			return nil, err
		}
		xs = append(xs, ys...)
	} else {
		// S013: sin commit base el árbol se exige igual; el aviso también se marca.
		xs = append(xs, Item{Clave: "aviso:sin-commit-base", Linea: "nivel archivos omitido: el rojo no guardó commit (repo sin commits o sin Git)"})
	}
	marcas := map[string]Marca{}
	for _, m := range f.Sello {
		marcas[m.Clave] = m
	}
	for i := range xs {
		if m, ok := marcas[xs[i].Clave]; ok {
			xs[i].Marca = &m
		}
	}
	return xs, nil
}

// selloCompleto: CERRADO exige marca en cada ítem, haya o no commit base (S013).
func selloCompleto(raiz string, f Ficha) error {
	if f.Tipo != "slice" {
		return nil
	}
	xs, err := sello(raiz, f)
	if err != nil {
		return err
	}
	var faltan []string
	for _, x := range xs {
		if x.Marca == nil {
			faltan = append(faltan, x.Clave)
		}
	}
	if len(faltan) == 0 {
		return nil
	}
	muestra := faltan
	if len(muestra) > 5 {
		muestra = append(muestra[:5:5], "…")
	}
	return fmt.Errorf("%s no cierra: sello incompleto, %d sin marca: %s → aquas backlog sello %s <clave> sin_cambio|actualizado motivo=…",
		f.ID, len(faltan), strings.Join(muestra, ", "), f.ID)
}

// Marcar guarda la revisión de un ítem del sello de un slice activo.
func Marcar(raiz, id, clave, marca, motivo string) (Ficha, error) {
	if marca != "sin_cambio" && marca != "actualizado" {
		return Ficha{}, fmt.Errorf("marca %q inválida: sin_cambio | actualizado", marca)
	}
	if motivo = strings.TrimSpace(motivo); motivo == "" || strings.ContainsAny(motivo, "\r\n") {
		return Ficha{}, errors.New("motivo obligatorio, de una línea")
	}
	f, err := activa(raiz, id)
	if err != nil {
		return Ficha{}, err
	}
	xs, err := sello(raiz, f)
	if err != nil {
		return Ficha{}, err
	}
	hay := false
	for _, x := range xs {
		hay = hay || x.Clave == clave
	}
	if !hay {
		return Ficha{}, fmt.Errorf("%s no está en el sello de %s: aquas backlog sello %s", clave, id, id)
	}
	m := Marca{Clave: clave, Marca: marca, Motivo: motivo}
	for i := range f.Sello {
		if f.Sello[i].Clave == clave {
			f.Sello[i] = m
			return f, escribir(raiz, Activo, f)
		}
	}
	f.Sello = append(f.Sello, m)
	return f, escribir(raiz, Activo, f)
}

// TextoSello: una línea por ítem; [ ] pendiente, [x] marcado con su marca y motivo.
func TextoSello(xs []Item) string {
	var b strings.Builder
	for _, x := range xs {
		if x.Marca == nil {
			fmt.Fprintf(&b, "  [ ] %s · %s\n", x.Clave, x.Linea)
		} else {
			fmt.Fprintf(&b, "  [x] %s · %s: %s\n", x.Clave, x.Marca.Marca, x.Marca.Motivo)
		}
	}
	return b.String()
}
