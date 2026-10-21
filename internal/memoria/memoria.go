// Package memoria: memoria caliente del foco de trabajo (E001/F005).
//
// Vive en .aquas/memoria.json (oculto, no versionado). Solo la escribe `aquas memoria fijar`;
// la guardia bloquea Write/Edit/Bash hacia .aquas/. El sello SHA-256 delata una edición por
// otra vía (un script): Leer devuelve ErrSelloRoto y Fijar se niega a taparla.
package memoria

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	Ruta    = ".aquas/memoria.json"
	Backlog = "gobierno/canonico/metodologias/E001_GOBIERNO_DEL_TRABAJO.md"
)

var (
	ErrSinMemoria = errors.New("sin memoria caliente: aún no se fijó un foco (aquas memoria fijar)")
	ErrSelloRoto  = errors.New("memoria caliente editada fuera de aquas: el sello no coincide")
)

type Foco struct {
	Iniciativa string `json:"iniciativa"`
	Feature    string `json:"feature"`
	Slice      string `json:"slice"`
	Tarea      string `json:"tarea"`
}

// Memoria: campos fijos; no admite campos extra.
type Memoria struct {
	Foco        Foco     `json:"foco"`
	Siguiente   string   `json:"siguiente"`
	Pendientes  []string `json:"pendientes"`
	Decisiones  []string `json:"decisiones"`
	Actualizado string   `json:"actualizado"`
	Sello       string   `json:"sello"`
}

// Mapa: rutas derivadas del foco; no se guardan.
type Mapa struct{ Iniciativa, Feature, Slice, Backlog string }

var formato = map[string]*regexp.Regexp{
	"iniciativa": regexp.MustCompile(`^E\d{3}$`),
	"feature":    regexp.MustCompile(`^F\d{3}$`),
	"slice":      regexp.MustCompile(`^S\d{3}$`),
}

// Fijar aplica pares campo=valor sobre la memoria actual y la sella.
// pendiente/decision: los valores nombrados reemplazan la lista entera; «pendiente=» la vacía.
func Fijar(raiz string, pares []string, ahora time.Time) (Memoria, error) {
	m, err := Leer(raiz)
	if err != nil && !errors.Is(err, ErrSinMemoria) {
		return Memoria{}, err
	}
	listas := map[string][]string{}
	for _, par := range pares {
		campo, valor, ok := strings.Cut(par, "=")
		if !ok {
			return Memoria{}, fmt.Errorf("se espera campo=valor; llegó %q", par)
		}
		if re := formato[campo]; re != nil && valor != "" && !re.MatchString(valor) {
			return Memoria{}, fmt.Errorf("%s debe ser %s y 3 dígitos; llegó %q", campo, re.String()[1:2], valor)
		}
		switch campo {
		case "iniciativa":
			m.Foco.Iniciativa = valor
		case "feature":
			m.Foco.Feature = valor
		case "slice":
			m.Foco.Slice = valor
		case "tarea":
			m.Foco.Tarea = valor
		case "siguiente":
			m.Siguiente = valor
		case "pendiente", "decision":
			if _, ya := listas[campo]; !ya {
				listas[campo] = []string{}
			}
			if valor != "" {
				listas[campo] = append(listas[campo], valor)
			}
		default:
			return Memoria{}, fmt.Errorf("campo desconocido: %s (válidos: iniciativa, feature, slice, tarea, siguiente, pendiente, decision)", campo)
		}
	}
	if l, ok := listas["pendiente"]; ok {
		m.Pendientes = l
	}
	if l, ok := listas["decision"]; ok {
		m.Decisiones = l
	}
	m.Actualizado = ahora.UTC().Format(time.RFC3339)
	m.Sello = ""
	m.Sello = huella(m)
	return m, escribir(raiz, m)
}

// Leer carga la memoria y comprueba su sello.
func Leer(raiz string) (Memoria, error) {
	b, err := os.ReadFile(filepath.Join(raiz, Ruta))
	if errors.Is(err, os.ErrNotExist) {
		return Memoria{}, ErrSinMemoria
	}
	if err != nil {
		return Memoria{}, err
	}
	var m Memoria
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err := d.Decode(&m); err != nil {
		return Memoria{}, fmt.Errorf("%w (%v)", ErrSelloRoto, err)
	}
	sello := m.Sello
	m.Sello = ""
	if sello == "" || huella(m) != sello {
		return Memoria{}, ErrSelloRoto
	}
	m.Sello = sello
	return m, nil
}

func Rutas(f Foco) Mapa {
	r := Mapa{Backlog: Backlog}
	if f.Iniciativa == "" {
		return r
	}
	r.Iniciativa = "docs/" + f.Iniciativa + "/README.md"
	if f.Feature == "" {
		return r
	}
	r.Feature = "docs/" + f.Iniciativa + "/" + f.Feature + "/README.md"
	if f.Slice != "" {
		r.Slice = "docs/" + f.Iniciativa + "/" + f.Feature + "/" + f.Slice + ".md"
	}
	return r
}

// Texto: lo que imprime `aquas memoria ver` (y, en S002, el hook de inicio de sesión).
func Texto(m Memoria) string {
	var b strings.Builder
	var ids []string
	for _, s := range []string{m.Foco.Iniciativa, m.Foco.Feature, m.Foco.Slice} {
		if s != "" {
			ids = append(ids, s)
		}
	}
	foco := strings.Join(ids, " › ")
	if foco == "" {
		foco = "—"
	}
	fmt.Fprintf(&b, "MEMORIA CALIENTE · actualizada %s\nfoco: %s\ntarea: %s\nsiguiente: %s\n", m.Actualizado, foco, guion(m.Foco.Tarea), guion(m.Siguiente))
	lista(&b, "pendientes", m.Pendientes)
	lista(&b, "decisiones abiertas", m.Decisiones)
	r := Rutas(m.Foco)
	b.WriteString("rutas:\n")
	for _, p := range [][2]string{{"iniciativa", r.Iniciativa}, {"feature", r.Feature}, {"slice", r.Slice}, {"backlog", r.Backlog}} {
		if p[1] != "" {
			fmt.Fprintf(&b, "  %s: %s\n", p[0], p[1])
		}
	}
	b.WriteString("Actualizar: aquas memoria fijar campo=valor … (no editar el archivo a mano)\n")
	return b.String()
}

func guion(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

func lista(b *strings.Builder, titulo string, xs []string) {
	if len(xs) == 0 {
		fmt.Fprintf(b, "%s: —\n", titulo)
		return
	}
	fmt.Fprintf(b, "%s:\n", titulo)
	for _, x := range xs {
		fmt.Fprintf(b, "- %s\n", x)
	}
}

func codificar(m Memoria) []byte {
	b, _ := json.MarshalIndent(m, "", "  ") // struct de strings: no falla
	return append(b, '\n')
}

func huella(m Memoria) string {
	s := sha256.Sum256(codificar(m))
	return hex.EncodeToString(s[:])
}

// escribir: archivo temporal + rename, para no dejar una memoria a medias.
func escribir(raiz string, m Memoria) error {
	p := filepath.Join(raiz, Ruta)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, codificar(m), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}
