// Package compilador genera los archivos derivados (CLAUDE.md, agentes) a partir
// de los canónicos de gobierno/ y los sella con SHA-256.
package compilador

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"
)

const (
	RutaAgentes    = "gobierno/config/agentes.json"
	RutaDocumentos = "gobierno/config/documentos.json"
	RutaSellos     = "gobierno/config/sellos.json"
	PrefijoCanon   = "gobierno/canonico/"
	DirAgentes     = ".claude/agents"
)

type Agente struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Model        string   `json:"model"`
	Effort       string   `json:"effort"`
	Cuerpo       string   `json:"cuerpo"`
	Reglas       []string `json:"reglas"`
	Comunicacion []string `json:"comunicacion"`
	Max          int      `json:"max_caracteres"` // presupuesto de contexto, obligatorio
}

type Documento struct {
	Salida  string   `json:"salida"`
	Fuentes []string `json:"fuentes"`
	Max     int      `json:"max_caracteres"` // presupuesto de contexto, obligatorio
}

// Salida es un archivo derivado calculado en memoria.
type Salida struct {
	Ruta      string
	Contenido []byte
	Fuentes   []string // canónicos + config que la determinan
	Max       int      // presupuesto en caracteres (runas)
}

type Sello struct {
	Salidas map[string]SelloSalida `json:"salidas"`
}

type SelloSalida struct {
	SHA256  string            `json:"sha256"`
	Fuentes map[string]string `json:"fuentes"`
}

// Informe resume lo que hizo Compilar.
type Informe struct {
	Escritos         []string // salidas reescritas en disco
	FuentesCambiadas []string // fuentes cuyo hash difiere del sello anterior
}

func marca(config string) string {
	return fmt.Sprintf("<!-- GENERADO por `aquas compilar` desde %s. NO EDITAR: edita los canónicos y recompila. -->", config)
}

// Planificar lee la config y calcula todas las salidas sin tocar el disco.
func Planificar(raiz string) ([]Salida, error) {
	var agentes struct {
		Agentes []Agente `json:"agentes"`
	}
	var docs struct {
		Documentos []Documento `json:"documentos"`
	}
	if err := leerJSON(raiz, RutaAgentes, &agentes); err != nil {
		return nil, err
	}
	if err := leerJSON(raiz, RutaDocumentos, &docs); err != nil {
		return nil, err
	}

	var salidas []Salida
	for _, d := range docs.Documentos {
		if d.Salida == "" || len(d.Fuentes) == 0 || d.Max <= 0 {
			return nil, fmt.Errorf("%s: documento requiere salida, fuentes y max_caracteres > 0", RutaDocumentos)
		}
		partes, err := leerCanonicos(raiz, d.Fuentes)
		if err != nil {
			return nil, err
		}
		contenido := marca(RutaDocumentos) + "\n\n" + strings.Join(partes, "\n\n") + "\n"
		salidas = append(salidas, Salida{
			Ruta:      d.Salida,
			Contenido: []byte(contenido),
			Fuentes:   append([]string{RutaDocumentos}, d.Fuentes...),
			Max:       d.Max,
		})
	}

	for _, a := range agentes.Agentes {
		s, err := compilarAgente(raiz, a)
		if err != nil {
			return nil, err
		}
		salidas = append(salidas, s)
	}
	sort.Slice(salidas, func(i, j int) bool { return salidas[i].Ruta < salidas[j].Ruta })
	return salidas, nil
}

func compilarAgente(raiz string, a Agente) (Salida, error) {
	if a.Name == "" || a.Description == "" || a.Model == "" || a.Cuerpo == "" || a.Max <= 0 {
		return Salida{}, fmt.Errorf("%s: agente %q requiere name, description, model, cuerpo y max_caracteres > 0", RutaAgentes, a.Name)
	}
	fuentes := append([]string{a.Cuerpo}, a.Reglas...)
	fuentes = append(fuentes, a.Comunicacion...)
	partes, err := leerCanonicos(raiz, fuentes)
	if err != nil {
		return Salida{}, err
	}

	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("name: " + a.Name + "\n")
	b.WriteString("description: " + citarYAML(a.Description) + "\n")
	b.WriteString("model: " + a.Model + "\n")
	if a.Effort != "" {
		b.WriteString("effort: " + a.Effort + "\n")
	}
	b.WriteString("---\n")
	b.WriteString(marca(RutaAgentes) + "\n\n")
	b.WriteString(strings.Join(partes, "\n\n") + "\n")

	return Salida{
		Ruta:      path.Join(DirAgentes, a.Name+".md"),
		Contenido: []byte(b.String()),
		Fuentes:   append([]string{RutaAgentes}, fuentes...),
		Max:       a.Max,
	}, nil
}

// Compilar escribe las salidas que cambiaron y actualiza el sello.
func Compilar(raiz string) (Informe, error) {
	var inf Informe
	salidas, err := Planificar(raiz)
	if err != nil {
		return inf, err
	}
	nuevo, err := sellar(raiz, salidas)
	if err != nil {
		return inf, err
	}
	anterior, _ := leerSello(raiz)
	inf.FuentesCambiadas = diferencias(anterior, nuevo)

	for _, s := range salidas {
		actual, _ := os.ReadFile(filepath.Join(raiz, s.Ruta))
		if bytes.Equal(actual, s.Contenido) {
			continue
		}
		if err := escribir(raiz, s.Ruta, s.Contenido); err != nil {
			return inf, err
		}
		inf.Escritos = append(inf.Escritos, s.Ruta)
	}

	datosSello, err := codificarSello(nuevo)
	if err != nil {
		return inf, err
	}
	actual, _ := os.ReadFile(filepath.Join(raiz, RutaSellos))
	if !bytes.Equal(actual, datosSello) {
		if err := escribir(raiz, RutaSellos, datosSello); err != nil {
			return inf, err
		}
		inf.Escritos = append(inf.Escritos, RutaSellos)
	}
	return inf, nil
}

// Verificar es el gate: devuelve un problema por cada salida o sello desfasado.
func Verificar(raiz string) ([]string, error) {
	salidas, err := Planificar(raiz)
	if err != nil {
		return nil, err
	}
	var problemas []string
	for _, s := range salidas {
		actual, err := os.ReadFile(filepath.Join(raiz, s.Ruta))
		switch {
		case err != nil:
			problemas = append(problemas, s.Ruta+": no existe")
		case !bytes.Equal(normalizar(actual), s.Contenido):
			problemas = append(problemas, s.Ruta+": desfasado o editado a mano")
		}
		if n := utf8.RuneCount(s.Contenido); n > s.Max {
			problemas = append(problemas, fmt.Sprintf("%s: excede presupuesto (%d > %d caracteres)", s.Ruta, n, s.Max))
		}
	}
	nuevo, err := sellar(raiz, salidas)
	if err != nil {
		return nil, err
	}
	datosSello, err := codificarSello(nuevo)
	if err != nil {
		return nil, err
	}
	actual, err := os.ReadFile(filepath.Join(raiz, RutaSellos))
	if err != nil || !bytes.Equal(normalizar(actual), datosSello) {
		problemas = append(problemas, RutaSellos+": sello desfasado")
	}
	return problemas, nil
}

// EsGobernado indica si una ruta relativa (con /) afecta la compilación.
func EsGobernado(rel string) bool {
	return strings.HasPrefix(rel, PrefijoCanon) ||
		rel == RutaAgentes || rel == RutaDocumentos
}

// EsSalida indica si una ruta relativa es un archivo compilado.
func EsSalida(raiz, rel string) bool {
	salidas, err := Planificar(raiz)
	if err != nil {
		return false
	}
	for _, s := range salidas {
		if s.Ruta == rel {
			return true
		}
	}
	return false
}

func sellar(raiz string, salidas []Salida) (Sello, error) {
	sello := Sello{Salidas: map[string]SelloSalida{}}
	for _, s := range salidas {
		ss := SelloSalida{SHA256: huella(s.Contenido), Fuentes: map[string]string{}}
		for _, f := range s.Fuentes {
			datos, err := os.ReadFile(filepath.Join(raiz, f))
			if err != nil {
				return sello, err
			}
			ss.Fuentes[f] = huella(normalizar(datos))
		}
		sello.Salidas[s.Ruta] = ss
	}
	return sello, nil
}

func diferencias(anterior, nuevo Sello) []string {
	vistos := map[string]bool{}
	var cambiadas []string
	for ruta, ns := range nuevo.Salidas {
		as := anterior.Salidas[ruta]
		for f, h := range ns.Fuentes {
			if as.Fuentes[f] != h && !vistos[f] {
				vistos[f] = true
				cambiadas = append(cambiadas, f)
			}
		}
	}
	sort.Strings(cambiadas)
	return cambiadas
}

func leerCanonicos(raiz string, rutas []string) ([]string, error) {
	partes := make([]string, 0, len(rutas))
	for _, r := range rutas {
		if !strings.HasPrefix(r, PrefijoCanon) {
			return nil, fmt.Errorf("%s: fuera de %s; solo se compilan canónicos", r, PrefijoCanon)
		}
		datos, err := os.ReadFile(filepath.Join(raiz, r))
		if err != nil {
			return nil, fmt.Errorf("leer canónico: %w", err)
		}
		partes = append(partes, strings.TrimRight(string(normalizar(datos)), "\n"))
	}
	return partes, nil
}

func leerJSON(raiz, rel string, v any) error {
	datos, err := os.ReadFile(filepath.Join(raiz, rel))
	if err != nil {
		return err
	}
	dec := json.NewDecoder(bytes.NewReader(datos))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("%s: %w", rel, err)
	}
	return nil
}

func leerSello(raiz string) (Sello, error) {
	var s Sello
	datos, err := os.ReadFile(filepath.Join(raiz, RutaSellos))
	if err != nil {
		return s, err
	}
	err = json.Unmarshal(datos, &s)
	return s, err
}

func codificarSello(s Sello) ([]byte, error) {
	datos, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(datos, '\n'), nil
}

// citarYAML produce un escalar YAML entre comillas dobles (JSON es YAML válido).
func citarYAML(s string) string {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(s)
	return strings.TrimRight(b.String(), "\n")
}

func normalizar(b []byte) []byte {
	return bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n"))
}

func huella(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func escribir(raiz, rel string, datos []byte) error {
	destino := filepath.Join(raiz, rel)
	if err := os.MkdirAll(filepath.Dir(destino), 0o755); err != nil {
		return err
	}
	return os.WriteFile(destino, datos, 0o644)
}
