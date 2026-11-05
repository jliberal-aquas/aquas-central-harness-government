package backlog

// E000/F013/S003: una respuesta corta del humano aprueba la única línea «apruebo …» que el agente propuso en su
// turno; un mensaje escrito en cola (adjunto queued_command) cuenta como mensaje del humano.

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"regexp"
	"strings"
)

// reCorta: respuestas cortas que aprueban, solas o con signo final.
var reCorta = regexp.MustCompile(`(?i)^(sí|si|dale|ejecútalo|ejecuta|apruebo|ok)[.!]*$`)

// campoPedido: el campo del cierre donde el agente pide input (01_COMUNICACION_HUMANO); su texto puede ser la propuesta.
const campoPedido = "¿Qué necesito de ti?"

// Adjunto: el adjunto de una línea del registro; Claude Code guarda así el mensaje que el humano escribe mientras
// el agente trabaja (type queued_command, origin human; forma observada el 2026-11-04). Prompt es texto o bloques.
type Adjunto struct {
	Type   string          `json:"type"`
	Prompt json.RawMessage `json:"prompt"`
	Origin struct {
		Kind string `json:"kind"`
	} `json:"origin"`
}

// rePropuesta: la propuesta exacta, sin otro texto: «apruebo <id>» con «con push», «tope N» o «tope NM» (A2, hallazgo 2:
// una opción en prosa, como «o dime si lo prefieres con push», no se aprueba).
var rePropuesta = regexp.MustCompile(`(?i)^apruebo\s+E\d{3}(?:/F\d{3}(?:/S\d{3})?)?(?:\s+(?:con\s+push|tope\s+\d+\s*m?|y))*$`)

// lineaPropuesta: la línea como propuesta, sola en su línea o sola en el campo de pedido del cierre, con comillas,
// código o negrita alrededor; "" si no lo es.
func lineaPropuesta(l string) string {
	if i := strings.Index(l, campoPedido); i >= 0 {
		l = l[i+len(campoPedido):]
	}
	l = strings.TrimRight(strings.Trim(strings.TrimSpace(l), "`*>-:«»\"' "), " .,;:`«»\"'*")
	if !rePropuesta.MatchString(l) {
		return ""
	}
	return l
}

// PropuestaAprobada: si respuesta es corta, la única propuesta distinta del texto del agente; si no, "".
// Dos propuestas distintas no aprueban nada; una línea con dos no es propuesta (rePropuesta).
func PropuestaAprobada(respuesta, agente string) string {
	r := strings.TrimSpace(respuesta)
	if deAgente(r) || !reCorta.MatchString(r) {
		return ""
	}
	var hallada string
	for _, l := range strings.Split(strings.ReplaceAll(agente, "\r\n", "\n"), "\n") {
		switch p := lineaPropuesta(l); {
		case p != "" && hallada != "" && p != hallada:
			return ""
		case p != "":
			hallada = p
		}
	}
	return hallada
}

// Turno recorre un registro en orden: guarda los textos del agente con «apruebo» desde el último mensaje humano
// (y los del turno anterior a ese mensaje) y cuenta las líneas del agente desde ese mensaje.
type Turno struct {
	agente  []string
	previo  []string // textos del agente antes del último mensaje humano
	despues int      // líneas del agente desde el último mensaje humano
	crudo   string   // texto tal cual del último mensaje humano
}

// Linea: el texto que vale como mensaje del humano en esta línea, o "" si no es del humano. Una línea de usuario
// con origen humano o un mensaje en cola de origen humano es del humano; si es corta, vale la propuesta del agente.
func (tu *Turno) Linea(tipo, origen string, contenido json.RawMessage, adj Adjunto) string {
	var t string
	switch {
	case tipo == "assistant":
		tu.despues++
		if bytes.Contains(contenido, []byte("pruebo")) { // sin «apruebo» no hay propuesta: no se arma el texto
			tu.agente = append(tu.agente, textoUsuario(contenido))
		}
		return ""
	case tipo == "user":
		t = TextoHumano(tipo, origen, contenido)
	case tipo == "attachment" && adj.Type == "queued_command" && adj.Origin.Kind == "human":
		if t = strings.TrimSpace(textoUsuario(adj.Prompt)); deAgente(t) {
			t = ""
		}
	}
	if t == "" {
		return ""
	}
	tu.crudo, tu.previo = t, tu.agente
	if p := PropuestaAprobada(t, strings.Join(tu.agente, "\n")); p != "" {
		t = p
	}
	tu.agente, tu.despues = nil, 0
	return t
}

// recorrer pasa por Turno cada línea humana o del agente con «apruebo»; del resto del agente solo se cuenta la línea.
func recorrer(ruta string, f func(vale string)) (*Turno, error) {
	fh, err := os.Open(ruta)
	if err != nil {
		return nil, err
	}
	defer func() { _ = fh.Close() }()
	var tu Turno
	rd := bufio.NewReader(fh)
	for {
		l, err := rd.ReadBytes('\n')
		agente := bytes.Contains(l, []byte(`"type":"assistant"`)) // A2, hallazgo 7: las api-request-blob no son del agente
		switch {
		case bytes.Contains(l, []byte(`"human"`)) || agente && bytes.Contains(l, []byte("pruebo")):
			var lr lineaRegistro
			if json.Unmarshal(l, &lr) == nil {
				f(tu.Linea(lr.Type, lr.Origin.Kind, lr.Message.Content, lr.Attachment))
			}
		case agente:
			tu.despues++
		}
		if errors.Is(err, io.EOF) {
			return &tu, nil
		}
		if err != nil {
			return nil, err
		}
	}
}

// Resolver: lo que vale el prompt con que el humano abre el turno (hook UserPromptSubmit): la propuesta del agente
// si es una respuesta corta, o el prompt tal cual. Si el prompt ya está escrito al final del registro, vale la
// propuesta del turno anterior a él; si no, la del turno en curso (A1, hallazgo 1). Sin registro legible, el prompt.
func Resolver(ruta, prompt string) string {
	prompt = strings.TrimSpace(prompt)
	if !reCorta.MatchString(prompt) { // solo una respuesta corta lee el registro
		return prompt
	}
	tu, err := recorrer(ruta, func(string) {})
	if err != nil {
		return prompt
	}
	agente := tu.agente
	if tu.despues == 0 && tu.crudo == prompt {
		agente = tu.previo
	}
	if p := PropuestaAprobada(prompt, strings.Join(agente, "\n")); p != "" {
		return p
	}
	return prompt
}
