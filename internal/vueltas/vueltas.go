// Package vueltas cuenta corridas y fallas de pruebas y gate en el historial de Claude Code (E000/F004/S001).
// Solo lee el historial; nunca lo modifica.
package vueltas

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
)

// Conteo de un tipo de corrida: cuántas, cuántas fallaron y por qué.
type Conteo struct {
	Corridas, Fallas int
	Motivos          map[string]int
}

// tipos: qué comando Bash cuenta como corrida de cada tipo. «grep verificar» no cuenta.
var tipos = []struct {
	nombre string
	cmd    *regexp.Regexp
}{
	{"go test", regexp.MustCompile(`\bgo test\b`)},
	{"verificar", regexp.MustCompile(`(aquas|aq\.exe)\s+verificar\b`)},
}

var (
	pruebaFallida = regexp.MustCompile(`--- FAIL: (\S+)`)
	gateFallido   = regexp.MustCompile(`GATE FALLIDO · (\S+)`)
)

// motivos de una salida para un tipo; vacío = sin señal propia de fallo.
func motivos(tipo, salida string) []string {
	var m []string
	if tipo == "go test" {
		for _, s := range pruebaFallida.FindAllStringSubmatch(salida, -1) {
			if p := strings.SplitN(s[1], "/", 2)[0]; !slices.Contains(m, p) { // subprueba cuenta bajo su prueba
				m = append(m, p)
			}
		}
		if strings.Contains(salida, "build failed") {
			m = append(m, "compilación")
		}
		return m
	}
	// verificar: cada sección «GATE FALLIDO · x»; la wiki se desglosa por regla (E000/F004/S002).
	seccion, reglas := "", 0
	cerrarWiki := func() {
		if seccion == "wiki" && reglas == 0 {
			m = append(m, "wiki") // sin detalle legible: queda el motivo grueso
		}
	}
	for _, l := range strings.Split(salida, "\n") {
		if s := gateFallido.FindStringSubmatch(l); s != nil {
			cerrarWiki()
			seccion, reglas = s[1], 0
			if seccion != "wiki" {
				m = append(m, seccion)
			}
			continue
		}
		item, ok := strings.CutPrefix(l, "  - ")
		if seccion != "wiki" || !ok {
			continue
		}
		if _, txt, ok := strings.Cut(item, ": "); ok {
			reglas++
			if r := "wiki: " + regla(txt); !slices.Contains(m, r) { // misma regla en dos páginas = una vuelta
				m = append(m, r)
			}
		}
	}
	cerrarWiki()
	return m
}

var (
	citaR      = regexp.MustCompile(`«[^»]*»`)
	parentesis = regexp.MustCompile(`\s*\(.*$`) // pista final; puede traer paréntesis anidados
	numeroR    = regexp.MustCompile(`\d+`)
)

// regla: el mensaje del gate sin lo que cambia de página a página (citas, paréntesis, números).
func regla(txt string) string {
	txt, _, _ = strings.Cut(txt, ": ") // «prueba sin slice dueño: TestX» → la regla, sin el caso
	txt = citaR.ReplaceAllString(txt, "«…»")
	txt = parentesis.ReplaceAllString(txt, "")
	return strings.TrimSpace(numeroR.ReplaceAllString(txt, "N"))
}

type bloque struct {
	Type  string `json:"type"`
	ID    string `json:"id"`
	Name  string `json:"name"`
	Input struct {
		Command string `json:"command"`
	} `json:"input"`
	ToolUseID string          `json:"tool_use_id"`
	IsError   bool            `json:"is_error"`
	Content   json.RawMessage `json:"content"`
}

type corrida struct {
	tipos  []string
	salida string
	error  bool
	listo  bool
}

// Contar lee un historial .jsonl y agrupa por tipo ("go test", "verificar").
// Falla = señal propia del tipo (FAIL, build failed, GATE FALLIDO); si la llamada
// dio error sin ninguna señal, falla cada tipo presente con motivo «otro».
// Llamada sin resultado (bloqueada o cortada) no cuenta: no corrió.
func Contar(r io.Reader) (map[string]Conteo, error) {
	var orden []string
	corridas := map[string]*corrida{}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 1<<20), 64<<20)
	for sc.Scan() {
		var l struct {
			Message struct {
				Content json.RawMessage `json:"content"`
			} `json:"message"`
		}
		if json.Unmarshal(sc.Bytes(), &l) != nil {
			continue
		}
		var bs []bloque
		if json.Unmarshal(l.Message.Content, &bs) != nil {
			continue
		}
		for _, b := range bs {
			switch {
			case b.Type == "tool_use" && b.Name == "Bash":
				var ts []string
				for _, t := range tipos {
					if t.cmd.MatchString(b.Input.Command) {
						ts = append(ts, t.nombre)
					}
				}
				if len(ts) > 0 {
					corridas[b.ID] = &corrida{tipos: ts}
					orden = append(orden, b.ID)
				}
			case b.Type == "tool_result":
				if c, ok := corridas[b.ToolUseID]; ok {
					c.salida, c.error, c.listo = texto(b.Content), b.IsError, true
				}
			}
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	out := map[string]Conteo{}
	for _, id := range orden {
		c := corridas[id]
		if !c.listo {
			continue
		}
		porTipo := map[string][]string{}
		señal := false
		for _, t := range c.tipos {
			porTipo[t] = motivos(t, c.salida)
			señal = señal || len(porTipo[t]) > 0
		}
		for _, t := range c.tipos {
			k := out[t]
			if k.Motivos == nil {
				k.Motivos = map[string]int{}
			}
			k.Corridas++
			m := porTipo[t]
			if len(m) == 0 && c.error && !señal {
				m = []string{"otro"}
			}
			if len(m) > 0 {
				k.Fallas++
			}
			for _, x := range m {
				k.Motivos[x]++
			}
			out[t] = k
		}
	}
	return out, nil
}

// texto: el contenido de un resultado viene como cadena o como lista de bloques {type,text}.
func texto(raw json.RawMessage) string {
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	var bs []struct {
		Text string `json:"text"`
	}
	if json.Unmarshal(raw, &bs) == nil {
		var b strings.Builder
		for _, x := range bs {
			b.WriteString(x.Text + "\n")
		}
		return b.String()
	}
	return string(raw)
}

// Carpeta del historial de Claude Code para este repo: ~/.claude/projects/<raíz con :\/ y espacios → ->.
func Carpeta(home, raiz string) string {
	slug := strings.NewReplacer(":", "-", `\`, "-", "/", "-", " ", "-").Replace(raiz)
	return filepath.Join(home, ".claude", "projects", slug)
}

// Medir imprime una línea por historial y el total. Sin archivos: todos los .jsonl de Carpeta.
func Medir(w io.Writer, home, raiz string, archivos []string) error {
	if len(archivos) == 0 {
		var err error
		if archivos, err = filepath.Glob(filepath.Join(Carpeta(home, raiz), "*.jsonl")); err != nil {
			return err
		}
	}
	total := map[string]Conteo{}
	for _, a := range archivos {
		f, err := os.Open(a)
		if err != nil {
			return err
		}
		c, err := Contar(f)
		f.Close()
		if err != nil {
			return fmt.Errorf("%s: %w", a, err)
		}
		fmt.Fprintln(w, Linea(strings.TrimSuffix(filepath.Base(a), ".jsonl"), c))
		for t, k := range c {
			s := total[t]
			if s.Motivos == nil {
				s.Motivos = map[string]int{}
			}
			s.Corridas, s.Fallas = s.Corridas+k.Corridas, s.Fallas+k.Fallas
			for m, n := range k.Motivos {
				s.Motivos[m] += n
			}
			total[t] = s
		}
	}
	fmt.Fprintln(w, Linea("TOTAL", total))
	return nil
}

// Linea: «nombre · go test 92/47 (TestA 3, compilación 2) · verificar 102/37 (wiki 20)»; motivos de más a menos.
func Linea(nombre string, c map[string]Conteo) string {
	var b strings.Builder
	b.WriteString(nombre)
	for _, t := range tipos {
		k, ok := c[t.nombre]
		if !ok {
			continue
		}
		fmt.Fprintf(&b, " · %s %d/%d", t.nombre, k.Corridas, k.Fallas)
		ms := make([]string, 0, len(k.Motivos))
		for m := range k.Motivos {
			ms = append(ms, m)
		}
		sort.Slice(ms, func(i, j int) bool {
			if k.Motivos[ms[i]] != k.Motivos[ms[j]] {
				return k.Motivos[ms[i]] > k.Motivos[ms[j]]
			}
			return ms[i] < ms[j]
		})
		for i, m := range ms {
			ms[i] = fmt.Sprintf("%s %d", m, k.Motivos[m])
		}
		if len(ms) > 5 {
			ms = ms[:5] // solo los 5 principales: se lee de un vistazo
		}
		if len(ms) > 0 {
			fmt.Fprintf(&b, " (%s)", strings.Join(ms, ", "))
		}
	}
	return b.String()
}
