package guardia

// E001/F017/S008: escribir en el alcance del slice EN_CURSO exige haber leído su página desde el último /compact.

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/backlog"
)

// lectura: la revisión de S008 para la llamada en curso; nil = no se exige (llamadas directas, pruebas viejas).
var lectura func(rel string) string

// reLector: un tramo de Bash que lee: empieza con cat, sed (sin -i) o head, o es la biblioteca de aquas (A1.4).
var (
	reLector   = regexp.MustCompile(`^\s*(cat|sed|head)\s|\blibrary\s+read\b`)
	reSedEdita = regexp.MustCompile(`^\s*sed\s(.*\s)?-[a-zA-Z]*i`)
	reTramo    = regexp.MustCompile(`&&|\|\||[;|\n]`)
)

// bashLee: algún tramo del comando lee la página: el lector y la ruta van en el mismo tramo.
func bashLee(comando, rel string) bool {
	for _, t := range reTramo.Split(filepath.ToSlash(comando), -1) {
		if reLector.MatchString(t) && !reSedEdita.MatchString(t) && strings.Contains(t, rel) {
			return true
		}
	}
	return false
}

// registroDe: el registro de quien llama; un subagente usa el suyo, <sesión>/subagents/agent-<id>.jsonl.
func registroDe(transcript, agente string) string {
	if agente == "" || transcript == "" || filepath.Base(filepath.Dir(transcript)) == "subagents" {
		return transcript
	}
	return filepath.Join(strings.TrimSuffix(transcript, ".jsonl"), "subagents", "agent-"+agente+".jsonl")
}

// exigirLectura arma la revisión: para rel dentro del alcance del slice EN_CURSO (que no sea su propia página),
// el registro debe mostrar una lectura de docs/<id>.md después del último compact.
func exigirLectura(raiz, registro string) func(string) string {
	return func(rel string) string {
		id, ok := backlog.EnAlcance(raiz, rel)
		pagina := "docs/" + id + ".md"
		if !ok || rel == pagina {
			return ""
		}
		leida, err := leyoPagina(registro, filepath.Join(raiz, filepath.FromSlash(pagina)), pagina)
		if err != nil {
			return "lee el contrato " + pagina + " antes de escribir " + rel + ": el registro de la sesión no se pudo leer (" + err.Error() + "); no se inventa la lectura (E001/F017/S008)"
		}
		if !leida {
			return "lee el contrato " + pagina + " antes de escribir " + rel + " (Read, cat, sed, head o `aquas library read`): desde el último /compact no hay lectura (E001/F017/S008)"
		}
		return ""
	}
}

// leyoPagina recorre el registro: un compact borra lo leído; un Read de abs o un Bash lector que nombra rel cuenta.
func leyoPagina(registro, abs, rel string) (bool, error) {
	if registro == "" {
		return false, errors.New("el hook no trajo transcript_path")
	}
	fh, err := os.Open(registro)
	if err != nil {
		return false, err
	}
	defer func() { _ = fh.Close() }()
	rd := bufio.NewReader(fh)
	leida := false
	for {
		b, err := rd.ReadBytes('\n')
		switch {
		case bytes.Contains(b, []byte(`"compact_boundary"`)):
			leida = false
		case bytes.Contains(b, []byte(`"tool_use"`)) && lineaLee(b, abs, rel):
			leida = true
		}
		if errors.Is(err, io.EOF) {
			return leida, nil
		}
		if err != nil {
			return false, err
		}
	}
}

// lineaLee: la línea del registro trae una llamada Read a la página o un Bash lector con su ruta.
func lineaLee(b []byte, abs, rel string) bool {
	var l struct {
		Message struct {
			Content json.RawMessage `json:"content"`
		} `json:"message"`
	}
	var bloques []struct {
		Type  string `json:"type"`
		Name  string `json:"name"`
		Input struct {
			FilePath string `json:"file_path"`
			Command  string `json:"command"`
		} `json:"input"`
	}
	if json.Unmarshal(b, &l) != nil || json.Unmarshal(l.Message.Content, &bloques) != nil {
		return false
	}
	for _, u := range bloques {
		if u.Type != "tool_use" {
			continue
		}
		if u.Name == "Read" && u.Input.FilePath != "" && strings.EqualFold(filepath.Clean(u.Input.FilePath), filepath.Clean(abs)) {
			return true
		}
		if u.Name == "Bash" && bashLee(u.Input.Command, rel) {
			return true
		}
	}
	return false
}
