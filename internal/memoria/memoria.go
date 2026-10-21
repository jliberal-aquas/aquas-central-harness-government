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
	"os/exec"
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
	HuellaGit   string   `json:"huella_git"`
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
	m.HuellaGit = HuellaGit(raiz)
	m.Actualizado = ahora.UTC().Format(time.RFC3339)
	m.Sello = ""
	m.Sello = huella(m)
	return m, escribir(raiz, m)
}

// reSello: línea del sello tal como la escribe codificar. El sello cubre los bytes escritos
// (con el sello vacío), no la estructura actual: un campo nuevo no rompe una memoria vieja.
var reSello = regexp.MustCompile(`(?m)^(\s*)"sello": "([0-9a-f]*)"`)

// Leer carga la memoria y comprueba su sello.
func Leer(raiz string) (Memoria, error) {
	b, err := os.ReadFile(filepath.Join(raiz, Ruta))
	if errors.Is(err, os.ErrNotExist) {
		return Memoria{}, ErrSinMemoria
	}
	if err != nil {
		return Memoria{}, err
	}
	return decodificar(b)
}

// decodificar: campos fijos y sello sobre los bytes tal como se escribieron.
func decodificar(b []byte) (Memoria, error) {
	var m Memoria
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err := d.Decode(&m); err != nil {
		return Memoria{}, fmt.Errorf("%w (%v)", ErrSelloRoto, err)
	}
	sello := reSello.FindSubmatch(b)
	if sello == nil || len(sello[2]) == 0 {
		return Memoria{}, ErrSelloRoto
	}
	sinSello := reSello.ReplaceAll(b, []byte(`$1"sello": ""`))
	if h := sha256.Sum256(sinSello); hex.EncodeToString(h[:]) != string(sello[2]) {
		return Memoria{}, ErrSelloRoto
	}
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

// --- S002: huella de Git y hooks ---

// HuellaGit resume el estado del repo: HEAD + cada ruta de `git status` con el contenido de su
// archivo. Excluye .aquas/. Si raiz no es la raíz de un repo propio, devuelve "".
func HuellaGit(raiz string) string {
	top, err := exec.Command("git", "-C", raiz, "rev-parse", "--show-toplevel").Output()
	if err != nil || !mismaCarpeta(strings.TrimSpace(string(top)), raiz) {
		return ""
	}
	head, _ := exec.Command("git", "-C", raiz, "rev-parse", "HEAD").Output() // sin commits: vacío
	st, err := exec.Command("git", "-C", raiz, "status", "--porcelain=v1", "-z", "--untracked-files=all",
		"--", ".", ":(exclude).aquas").Output()
	if err != nil {
		return ""
	}
	h := sha256.New()
	h.Write(head)
	for _, e := range strings.Split(string(st), "\x00") {
		if len(e) < 4 {
			continue
		}
		h.Write([]byte(e + "\x00"))
		if c, err := os.ReadFile(filepath.Join(raiz, filepath.FromSlash(e[3:]))); err == nil {
			s := sha256.Sum256(c)
			h.Write(s[:])
		}
	}
	return hex.EncodeToString(h.Sum(nil))
}

func mismaCarpeta(a, b string) bool {
	ia, err1 := os.Stat(filepath.FromSlash(a))
	ib, err2 := os.Stat(b)
	return err1 == nil && err2 == nil && os.SameFile(ia, ib)
}

// Cierre decide el hook Stop: "" deja cerrar el turno; si no, el motivo para seguir.
func Cierre(raiz string, stopHookActive bool) string {
	if stopHookActive {
		return "" // ya se pidió una vez: no bloquear en bucle
	}
	m, err := Leer(raiz)
	switch {
	case errors.Is(err, ErrSinMemoria):
		return ""
	case errors.Is(err, ErrSelloRoto):
		return "aquas memoria cierre: " + err.Error() + ". No la rehagas: avisa al humano para que revise " + Ruta + "."
	case err != nil:
		return "aquas memoria cierre: " + err.Error()
	}
	if m.Foco == (Foco{}) {
		return ""
	}
	if HuellaGit(raiz) != m.HuellaGit {
		return "memoria caliente vieja: el repo cambió desde la última actualización (" + m.Actualizado +
			"). Antes de cerrar, actualiza foco, tarea y siguiente con: aquas memoria fijar tarea=… siguiente=…"
	}
	return ""
}

// Inicio: texto que el hook SessionStart agrega al contexto. Nunca bloquea.
func Inicio(raiz string) string {
	m, err := Leer(raiz)
	switch {
	case errors.Is(err, ErrSelloRoto):
		return "AVISO: " + err.Error() + ". No confíes en su contenido; avisa al humano.\n"
	case err != nil:
		return "aquas: " + err.Error() + "\n"
	}
	return Texto(m)
}
