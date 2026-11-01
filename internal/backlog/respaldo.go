package backlog

// E000/F008/S002 (D000077): una aprobación vale solo si su mensaje humano exacto está en el registro de la sesión
// (transcript_path de Claude Code). El registro vive fuera del repo; con el aislamiento de Bash (S003) el agente
// no lo puede escribir. En Windows nativo es una segunda barrera, probabilística (D000079).

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

// carpetaRegistros: <CLAUDE_CONFIG_DIR o ~/.claude>/projects, donde Claude Code guarda los registros de sesión.
func carpetaRegistros() (string, error) {
	base := os.Getenv("CLAUDE_CONFIG_DIR")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".claude")
	}
	return filepath.Join(base, "projects"), nil
}

// bajo: la ruta real hija queda dentro de la ruta real padre.
func bajo(padre, hija string) bool {
	r, err := filepath.Rel(padre, hija)
	return err == nil && r != "." && r != ".." && !strings.HasPrefix(r, ".."+string(filepath.Separator)) && !filepath.IsAbs(r)
}

// real: ruta absoluta sin enlaces; si no existe, error.
func real(p string) (string, error) {
	a, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(a)
}

// respaldada: nil si a.Mensaje aprueba justo ese id (con «con push» solo si a.Push) y está como mensaje humano
// en un registro de sesión válido; si no, la causa. Así un mensaje humano real de otra cosa no sirve (A1, hallazgo 1).
func respaldada(raiz, id string, a *Aprobacion) error {
	pide := fmt.Sprintf("el humano debe escribir «apruebo %s» otra vez (el hook la registra con su registro)", id)
	primera, _, _ := strings.Cut(strings.TrimSpace(a.Mensaje), "\n")
	if got, ok := IDAprobado(a.Mensaje); !ok || got != id || a.Push && !reConPush.MatchString(primera) {
		return fmt.Errorf("el mensaje guardado en la aprobación de %s no es «apruebo %s» con su «con push»: %s", id, id, pide)
	}
	if a.Registro == "" {
		return fmt.Errorf("la aprobación de %s no tiene registro de sesión: %s", id, pide)
	}
	reg, err := registroValido(raiz, a.Registro)
	if err != nil {
		return fmt.Errorf("el registro de la aprobación de %s %v: %s", id, err, pide)
	}
	ok, err := mensajeHumano(reg, a.Mensaje)
	if err != nil || !ok {
		return fmt.Errorf("el registro de sesión no tiene el mensaje humano «%s» que aprobó %s: %s", a.Mensaje, id, pide)
	}
	return nil
}

// registroValido: la ruta real de un registro de Claude Code, si es un .jsonl de la carpeta de registros y queda
// fuera del repo, donde el agente escribe; si no, la causa (S002 y S004).
func registroValido(raiz, ruta string) (string, error) {
	reg, err := real(ruta)
	if err != nil {
		return "", errors.New("no se puede leer")
	}
	proy, err := carpetaRegistros()
	if err == nil {
		proy, err = real(proy)
	}
	r, errR := real(raiz)
	switch {
	case err != nil || errR != nil:
		return "", errors.New("no se ubica: falta la carpeta de registros de Claude Code")
	case !strings.EqualFold(filepath.Ext(reg), ".jsonl") || !bajo(proy, reg):
		return "", fmt.Errorf("no es un .jsonl de %s", proy)
	case bajo(r, reg) || bajo(r, proy):
		return "", errors.New("está dentro del repo, donde el agente escribe")
	}
	return reg, nil
}

// lineaRegistro: lo que se lee de cada línea del registro (forma observada el 2026-10-31).
type lineaRegistro struct {
	Type    string `json:"type"`
	Message struct {
		Content json.RawMessage `json:"content"`
	} `json:"message"`
	Origin struct {
		Kind string `json:"kind"`
	} `json:"origin"`
}

// mensajeHumano: alguna línea de usuario con origen humano trae ese texto exacto; un mensaje guardado con 300
// runas (recortado por Aprobar) se compara como prefijo. Líneas rotas y de otros se ignoran.
func mensajeHumano(ruta, mensaje string) (bool, error) {
	fh, err := os.Open(ruta)
	if err != nil {
		return false, err
	}
	defer func() { _ = fh.Close() }()
	recortado := utf8.RuneCountInString(mensaje) >= 300
	rd := bufio.NewReader(fh)
	for {
		l, err := rd.ReadBytes('\n')
		if len(l) > 0 && bytes.Contains(l, []byte(`"human"`)) {
			var lr lineaRegistro
			if json.Unmarshal(l, &lr) == nil && lr.Type == "user" && lr.Origin.Kind == "human" {
				t := strings.TrimSpace(textoUsuario(lr.Message.Content))
				if t != "" && (t == mensaje || recortado && strings.HasPrefix(t, mensaje)) {
					return true, nil
				}
			}
		}
		if errors.Is(err, io.EOF) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
	}
}

// textoUsuario: contenido como texto, o la unión de sus bloques «text»; otros bloques (tool_result) no cuentan.
func textoUsuario(c json.RawMessage) string {
	var s string
	if json.Unmarshal(c, &s) == nil {
		return s
	}
	var bloques []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(c, &bloques) != nil {
		return ""
	}
	var partes []string
	for _, b := range bloques {
		if b.Type == "text" {
			partes = append(partes, b.Text)
		}
	}
	return strings.Join(partes, "\n")
}

// reApruebo: la línea empieza (tras espacios) por «apruebo <id>». «no apruebo …» no cuenta.
var reApruebo = regexp.MustCompile(`(?i)^\s*apruebo\s+(E\d{3}(?:/F\d{3}(?:/S\d{3})?)?)\b`)

// MarcasDeAgente: un prompt con alguna de estas marcas es un mensaje de agente, no del humano.
var MarcasDeAgente = []string{"<agent-message", "[Subagent hand-back]", "<task-notification>"}

// IDAprobado: id de la PRIMERA línea no vacía si es «apruebo <id>»; una línea que termina en «?» es pregunta.
// Todo prompt con marcas de mensaje de agente se descarta.
func IDAprobado(prompt string) (string, bool) {
	for _, m := range MarcasDeAgente {
		if strings.Contains(prompt, m) {
			return "", false
		}
	}
	for _, l := range strings.Split(strings.ReplaceAll(prompt, "\r\n", "\n"), "\n") {
		if strings.TrimSpace(l) == "" {
			continue
		}
		if strings.HasSuffix(strings.TrimSpace(l), "?") {
			return "", false
		}
		if m := reApruebo.FindStringSubmatch(l); m != nil {
			return m[1], true
		}
		return "", false
	}
	return "", false
}
