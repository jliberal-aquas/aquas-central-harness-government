// Package defecto: registro de defectos D000001 por repo, en Git (E001/F004/S001).
// Un archivo por defecto en registro/defectos/; solo se escribe con `aquas defect`.
package defecto

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/backlog"
)

// Carpeta del registro, relativa a la raíz del repo.
const Carpeta = "registro/defectos"

// Defecto: un fallo con firma estable y los turnos en que ocurrió.
type Defecto struct {
	ID           string   `json:"id"`
	Firma        string   `json:"firma"`
	Que          string   `json:"que"`
	Causa        string   `json:"causa"`
	Arreglo      string   `json:"arreglo"`
	Cercioro     string   `json:"como_me_cercioro"`
	Ocurrencias  []string `json:"ocurrencias"`
	Relacion     string   `json:"relacion,omitempty"`      // E001/F009/S002: nuevo o el id de una ficha
	Slice        string   `json:"slice,omitempty"`         // E001/F004/S003: slice de reincidencia ya creado
	Prueba       string   `json:"prueba,omitempty"`        // E001/F017/S005: ./paquete:TestNombre o pendiente:<id de ficha>
	Familia      string   `json:"familia,omitempty"`       // E001/F018/S003: una de Familias
	TramoFamilia string   `json:"tramo_familia,omitempty"` // E001/F018/S003: tramo de reincidencia de su familia
}

// Familias (E001/F018/S003): lista cerrada, tomada de la retrospectiva del 2026-11-02. Agregar una = sí del humano.
var Familias = []string{"comunicacion", "freno", "aprobaciones", "hijos", "documentacion", "asesor", "mecanica",
	"memoria", "herramientas", "falso-verde", "contrato", "alcance"}

// TituloFamilia: título del tramo de reincidencia de una familia en Feature.
func TituloFamilia(familia string) string { return "Reincidencia familia " + familia }

// Campos que pide `nuevo`, en orden.
var Campos = []string{"firma", "que", "causa", "arreglo", "cercioro", "prueba", "familia"} // prueba: E001/F017/S005; familia: E001/F018/S003

var (
	kebab  = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	turnoR = regexp.MustCompile(`^T\d{6}$`)
	idR    = regexp.MustCompile(`^D\d{6}$`)
)

// Nuevo crea el siguiente D000NNN. Todos los campos son obligatorios; la firma es kebab-case y única.
func Nuevo(raiz string, c map[string]string) (Defecto, error) {
	for k := range c {
		if !slices.Contains(Campos, k) && k != "relacion" {
			return Defecto{}, fmt.Errorf("campo desconocido %q: usa %s, relacion", k, strings.Join(Campos, ", "))
		}
	}
	for _, k := range Campos {
		if k == "prueba" || k == "familia" {
			continue // RevisarPrueba y la lista de familias explican qué falta
		}
		if strings.TrimSpace(c[k]) == "" {
			return Defecto{}, fmt.Errorf("falta el campo %s", k)
		}
	}
	if !slices.Contains(Familias, c["familia"]) {
		return Defecto{}, fmt.Errorf("familia %q: usa una de %s (agregar una pide el sí del humano)", c["familia"], strings.Join(Familias, ", "))
	}
	if !kebab.MatchString(c["firma"]) {
		return Defecto{}, fmt.Errorf("firma %q: debe ser kebab-case, ej. hook-exit-go-run", c["firma"])
	}
	todos, err := Leer(raiz)
	if err != nil {
		return Defecto{}, err
	}
	for _, d := range todos {
		if d.Firma == c["firma"] {
			return Defecto{}, fmt.Errorf("la firma %s ya existe en %s: usa `aquas defect occur %s T000NNN`", d.Firma, d.ID, d.ID)
		}
	}
	if err := revisarRelacion(raiz, c["relacion"], c["que"], todos); err != nil {
		return Defecto{}, err
	}
	if err := RevisarPrueba(raiz, c["prueba"]); err != nil {
		return Defecto{}, err
	}
	d := Defecto{
		ID:       fmt.Sprintf("D%06d", len(todos)+1),
		Relacion: c["relacion"],
		Firma:    c["firma"], Que: c["que"], Causa: c["causa"], Arreglo: c["arreglo"], Cercioro: c["cercioro"], Prueba: strings.TrimSpace(c["prueba"]),
		Ocurrencias: []string{}, Familia: c["familia"],
	}
	if err := escribir(raiz, d); err != nil {
		return Defecto{}, err
	}
	return revisarFamilia(raiz, d)
}

// Ocurre suma un turno T000NNN al defecto; el mismo turno no se cuenta dos veces.
func Ocurre(raiz, id, turno string) (Defecto, error) {
	if !turnoR.MatchString(turno) {
		return Defecto{}, fmt.Errorf("turno %q: formato T000001", turno)
	}
	d, err := leerUno(raiz, id)
	if err != nil {
		return Defecto{}, err
	}
	// Hallazgo A1.6: con el umbral alcanzado y sin slice, repetir el mismo turno reintenta sin sumar ocurrencia.
	pendiente := 1+len(d.Ocurrencias) >= Umbral && d.Slice == "" && d.TramoFamilia == "" // A2.2: el de familia también cuenta
	if slices.Contains(d.Ocurrencias, turno) {
		if !pendiente {
			return Defecto{}, fmt.Errorf("%s ya tiene el turno %s", id, turno)
		}
	} else {
		d.Ocurrencias = append(d.Ocurrencias, turno)
		if err := escribir(raiz, d); err != nil {
			return Defecto{}, err
		}
	}
	if d.Slice != "" || d.TramoFamilia != "" { // ya lo ataca su tramo propio o el de su familia (hallazgo A1.3)
		return d, nil
	}
	if 1+len(d.Ocurrencias) < Umbral {
		return revisarFamilia(raiz, d)
	}
	s, err := sliceReincidencia(raiz, d)
	if err != nil {
		return Defecto{}, fmt.Errorf("%s llegó a %d apariciones y su slice no nació; la ocurrencia quedó guardada; arregla la causa y repite `aquas defect occur %s %s`: %w", id, 1+len(d.Ocurrencias), id, turno, err)
	}
	d.Slice = s
	return d, escribir(raiz, d)
}

// Umbral (E001/F004/S003): apariciones (registro + ocurrencias) que convierten un defecto en slice de Feature.
const Umbral = 3

// Feature: donde nacen los slices de reincidencia; vive en backlog, que la salta mientras esté vacía.
const Feature = backlog.FeatureReincidencias

// sliceReincidencia devuelve el slice del defecto: el que ya lleve su título o uno nuevo P1 en Feature.
// Reusa el existente (activo o archivado) tras un fallo a mitad; si no, toma el SNNN siguiente al mayor.
// Feature archivada se niega: un hijo activo de padre archivado rompe el gate (hallazgo A1.1).
func sliceReincidencia(raiz string, d Defecto) (string, error) {
	titulo := "Reincidencia " + d.ID + ": " + d.Firma
	fs, err := hijosVivos(raiz)
	if err != nil {
		return "", err
	}
	for _, f := range fs {
		if f.Titulo == titulo {
			return f.ID, nil
		}
	}
	id := siguienteSlice(fs)
	_, err = backlog.Nuevo(raiz, []string{"id=" + id, "titulo=" + titulo, "prioridad=P1", "relacion=nueva",
		fmt.Sprintf("motivo=%s apareció %d veces (E001/F004/S003)", d.ID, 1+len(d.Ocurrencias)),
		"que_hace=ataca la causa de " + d.ID + " para que no vuelva a ocurrir",
		"prueba=./internal/defecto:TestReincidencia" + d.ID + "=valida que " + d.ID + " no vuelve|rojo cuando " + d.Firma + " se reproduce|verde cuando el control lo impide"})
	return id, err
}

// revisarFamilia (E001/F018/S003): si la familia de d junta Umbral apariciones entre defectos sin tramo propio ni de
// familia, crea o reutiliza su tramo y los marca. Si el tramo no nace, nadie queda marcado.
func revisarFamilia(raiz string, d Defecto) (Defecto, error) {
	if d.Familia == "" || d.TramoFamilia != "" || d.Slice != "" {
		return d, nil
	}
	todos, err := Leer(raiz)
	if err != nil {
		return Defecto{}, err
	}
	var cuentan []Defecto
	n := 0
	for _, x := range todos {
		if x.Familia == d.Familia && x.TramoFamilia == "" && x.Slice == "" {
			cuentan = append(cuentan, x)
			n += 1 + len(x.Ocurrencias)
		}
	}
	if n < Umbral {
		return d, nil
	}
	s, err := tramoFamilia(raiz, d.Familia, n)
	if err != nil {
		return Defecto{}, fmt.Errorf("%s quedó guardado, pero la familia %s llegó a %d apariciones y su tramo no nació; arregla la causa y el próximo defecto de la familia lo reintenta: %w", d.ID, d.Familia, n, err)
	}
	for _, x := range cuentan {
		x.TramoFamilia = s
		if err := escribir(raiz, x); err != nil {
			return Defecto{}, err
		}
		if x.ID == d.ID {
			d = x
		}
	}
	return d, nil
}

// tramoFamilia devuelve el tramo abierto de la familia o crea uno nuevo P1 en Feature; uno archivado no se reutiliza.
func tramoFamilia(raiz, familia string, n int) (string, error) {
	titulo := TituloFamilia(familia)
	fs, err := hijosVivos(raiz)
	if err != nil {
		return "", err
	}
	for _, f := range fs {
		if f.Titulo == titulo && f.Estado != "CERRADO" && f.Estado != "CANCELADO" {
			return f.ID, nil
		}
	}
	id := siguienteSlice(fs)
	nombre := ""
	for _, p := range strings.Split(familia, "-") {
		nombre += strings.ToUpper(p[:1]) + p[1:]
	}
	_, err = backlog.Nuevo(raiz, []string{"id=" + id, "titulo=" + titulo, "prioridad=P1", "relacion=nueva",
		fmt.Sprintf("motivo=la familia %s juntó %d apariciones (E001/F018/S003)", familia, n),
		"que_hace=ataca la causa común de los defectos de la familia " + familia,
		"prueba=./internal/defecto:TestReincidenciaFamilia" + nombre + "=valida que la familia " + familia + " no vuelve|rojo cuando un defecto de la familia se reproduce|verde cuando el control lo impide"})
	return id, err
}

// hijosVivos: los slices de Feature, activos o archivados; Feature archivada se niega (hallazgo A1.1 de E001/F004/S003).
func hijosVivos(raiz string) ([]backlog.Ficha, error) {
	if f, archivada, err := backlog.Ver(raiz, Feature); err != nil {
		return nil, err
	} else if archivada {
		return nil, fmt.Errorf("%s está archivada (%s): un slice activo bajo ella rompe el gate", Feature, f.Estado)
	}
	fs, err := backlog.Todas(raiz)
	if err != nil {
		return nil, err
	}
	var out []backlog.Ficha
	for _, f := range fs {
		if f.Padre == Feature {
			out = append(out, f)
		}
	}
	return out, nil
}

// siguienteSlice: el SNNN siguiente al mayor de Feature, no el primer hueco (hallazgo A1.4 de E001/F004/S003).
func siguienteSlice(fs []backlog.Ficha) string {
	n := 0
	for _, f := range fs {
		var k int
		if _, err := fmt.Sscanf(strings.TrimPrefix(f.ID, Feature+"/S"), "%d", &k); err == nil && k > n {
			n = k
		}
	}
	return fmt.Sprintf("%s/S%03d", Feature, n+1)
}

// Leer devuelve todos los defectos, ordenados por ID. Sin carpeta = registro vacío.
func Leer(raiz string) ([]Defecto, error) {
	es, err := os.ReadDir(filepath.Join(raiz, Carpeta))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []Defecto
	for _, e := range es {
		id, ok := strings.CutSuffix(e.Name(), ".json")
		if !ok || !idR.MatchString(id) {
			continue
		}
		d, err := leerUno(raiz, id)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

// leerUno carga el defecto con ese ID; valida el formato D000001.
func leerUno(raiz, id string) (Defecto, error) {
	if !idR.MatchString(id) {
		return Defecto{}, fmt.Errorf("id %q: formato D000001", id)
	}
	b, err := os.ReadFile(filepath.Join(raiz, Carpeta, id+".json"))
	if err != nil {
		return Defecto{}, fmt.Errorf("defecto %s: %w", id, err)
	}
	var d Defecto
	if err := json.Unmarshal(b, &d); err != nil {
		return Defecto{}, fmt.Errorf("defecto %s: %w", id, err)
	}
	return d, nil
}

// escribir guarda el defecto como JSON en su carpeta.
func escribir(raiz string, d Defecto) error {
	dir := filepath.Join(raiz, Carpeta)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, d.ID+".json"), append(b, '\n'), 0o644)
}

// Texto: una línea por defecto para `aquas defect show`.
func Texto(d Defecto) string {
	if d.Slice != "" {
		return fmt.Sprintf("%s · %s · %d ocurrencias · slice %s · %s\n", d.ID, d.Firma, len(d.Ocurrencias), d.Slice, d.Que)
	}
	return fmt.Sprintf("%s · %s · %d ocurrencias · %s\n", d.ID, d.Firma, len(d.Ocurrencias), d.Que)
}

// revisarRelacion (E001/F009/S002): relacion=nuevo o el id de una ficha; sin ella, lista defectos y fichas parecidos.
// repite:<D> se niega: el mismo defecto se cuenta con `aquas defect occur`.
func revisarRelacion(raiz, rel, que string, todos []Defecto) error {
	switch {
	case rel == "":
		ps, err := backlog.Parecidas(raiz, que)
		if err != nil {
			return err
		}
		m := "falta relacion=nuevo | repite:<D> | <id de ficha>; revisa si ya existe:"
		if len(ps) == 0 {
			m += " sin parecidas"
		}
		for _, p := range ps {
			m += "\n  - " + p
		}
		return fmt.Errorf("%s", m)
	case rel == "nuevo":
		return nil
	case strings.HasPrefix(rel, "repite:"):
		id := strings.TrimPrefix(rel, "repite:")
		for _, d := range todos {
			if d.ID == id {
				return fmt.Errorf("es el mismo defecto: suma la ocurrencia con `aquas defect occur %s T000NNN`", id)
			}
		}
		return fmt.Errorf("relacion: el defecto %s no existe", id)
	}
	if _, _, err := backlog.Ver(raiz, rel); err != nil {
		return fmt.Errorf("relacion %q: la ficha no existe o el valor es inválido (usa nuevo, repite:<D> o el id de una ficha)", rel)
	}
	return nil
}

// Arreglar (E000/F007/S003, D000109): marca el defecto como arreglado. Reemplaza arreglo y prueba; la prueba debe ser
// real (./paquete:TestNombre que existe), no pendiente:. Un rechazo no toca el archivo.
func Arreglar(raiz, id, arreglo, prueba string) (Defecto, error) {
	d, err := leerUno(raiz, id)
	if err != nil {
		return Defecto{}, err
	}
	if strings.TrimSpace(arreglo) == "" {
		return Defecto{}, fmt.Errorf("falta el campo arreglo: di qué control quedó implementado")
	}
	prueba = strings.TrimSpace(prueba)
	if strings.HasPrefix(prueba, "pendiente:") {
		return Defecto{}, fmt.Errorf("prueba=%s: un defecto arreglado nombra su prueba real (./paquete:TestNombre) y no una ficha pendiente", prueba)
	}
	if err := RevisarPrueba(raiz, prueba); err != nil {
		return Defecto{}, err
	}
	d.Arreglo, d.Prueba = strings.TrimSpace(arreglo), prueba
	return d, escribir(raiz, d)
}
