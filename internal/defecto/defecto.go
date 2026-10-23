// Package defecto: registro de defectos D000001 por repo, en Git (E001/F004/S001).
// Un archivo por defecto en registro/defectos/; solo se escribe con `aquas defecto`.
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
	ID          string   `json:"id"`
	Firma       string   `json:"firma"`
	Que         string   `json:"que"`
	Causa       string   `json:"causa"`
	Arreglo     string   `json:"arreglo"`
	Cercioro    string   `json:"como_me_cercioro"`
	Ocurrencias []string `json:"ocurrencias"`
	Relacion    string   `json:"relacion,omitempty"` // E001/F009/S002: nuevo o el id de una ficha
}

// Campos que pide `nuevo`, en orden.
var Campos = []string{"firma", "que", "causa", "arreglo", "cercioro"}

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
		if strings.TrimSpace(c[k]) == "" {
			return Defecto{}, fmt.Errorf("falta el campo %s", k)
		}
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
			return Defecto{}, fmt.Errorf("la firma %s ya existe en %s: usa `aquas defecto ocurre %s T000NNN`", d.Firma, d.ID, d.ID)
		}
	}
	if err := revisarRelacion(raiz, c["relacion"], c["que"], todos); err != nil {
		return Defecto{}, err
	}
	d := Defecto{
		ID:       fmt.Sprintf("D%06d", len(todos)+1),
		Relacion: c["relacion"],
		Firma:    c["firma"], Que: c["que"], Causa: c["causa"], Arreglo: c["arreglo"], Cercioro: c["cercioro"],
		Ocurrencias: []string{},
	}
	return d, escribir(raiz, d)
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
	if slices.Contains(d.Ocurrencias, turno) {
		return Defecto{}, fmt.Errorf("%s ya tiene el turno %s", id, turno)
	}
	d.Ocurrencias = append(d.Ocurrencias, turno)
	return d, escribir(raiz, d)
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

// Texto: una línea por defecto para `aquas defecto ver`.
func Texto(d Defecto) string {
	return fmt.Sprintf("%s · %s · %d ocurrencias · %s\n", d.ID, d.Firma, len(d.Ocurrencias), d.Que)
}

// revisarRelacion (E001/F009/S002): relacion=nuevo o el id de una ficha; sin ella, lista defectos y fichas parecidos.
// repite:<D> se niega: el mismo defecto se cuenta con `aquas defecto ocurre`.
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
				return fmt.Errorf("es el mismo defecto: suma la ocurrencia con `aquas defecto ocurre %s T000NNN`", id)
			}
		}
		return fmt.Errorf("relacion: el defecto %s no existe", id)
	}
	if _, _, err := backlog.Ver(raiz, rel); err != nil {
		return fmt.Errorf("relacion %q: la ficha no existe o el valor es inválido (usa nuevo, repite:<D> o el id de una ficha)", rel)
	}
	return nil
}
