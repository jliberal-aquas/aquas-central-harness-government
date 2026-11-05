package main

// E000/F014/S002 (D000125): ningún gancho falla mudo. Claude Code no muestra el exit 1 ni el stderr de un
// exit 0 (code.claude.com/docs/en/hooks, consultada 2026-11-04). Por eso, lanzado como gancho, aquas se corre
// a sí mismo como proceso hijo y traduce su salida: 0 y 2 pasan tal cual; cualquier otro código pasa a
// exit 0 con systemMessage (lo ve el humano) y, en los eventos que lo aceptan, additionalContext (lo ve el
// agente). La guardia falla cerrada: su fallo sale con exit 2 y la causa.

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ganchos: orden interna → evento de Claude Code que la lanza. Toda orden de gobierno/config/claude_settings.json
// debe estar aquí; TestGanchosNuncaMudos lo exige.
var ganchos = map[string]string{
	"guardia":                 "PreToolUse",
	"presupuesto herramienta": "PreToolUse",
	"hook":                    "PostToolUse",
	"memoria inicio":          "SessionStart",
	"turno abrir":             "UserPromptSubmit",
	"aprobacion":              "UserPromptSubmit",
	"turno cerrar":            "Stop",
	"memoria cierre":          "Stop",
	"propuesta":               "Stop",
	"cierre-agente":           "SubagentStop",
}

// frenaSiFalla: ganchos cuyo fallo frena la acción (falla cerrado) en vez de avisar.
var frenaSiFalla = map[string]bool{"guardia": true}

// conContexto: eventos que aceptan hookSpecificOutput.additionalContext (referencia de hooks, 2026-11-04).
// PreToolUse: presupuesto.go ya lo emite (E001/F019/S002). Stop y SubagentStop no lo aceptan: ahí el aviso
// llega solo al humano, por systemMessage.
var conContexto = map[string]bool{"UserPromptSubmit": true, "SessionStart": true, "PostToolUse": true, "PreToolUse": true}

const envGancho = "AQUAS_GANCHO_HIJO"

// avisoNombreViejo: texto del aviso de main por un nombre de comando viejo; no es un fallo del gancho.
const avisoNombreViejo = "el nombre viejo dejará de existir"

// claveGancho: la orden de gancho que nombran los argumentos internos, o "".
func claveGancho(args []string) string {
	if len(args) > 1 {
		if _, ok := ganchos[args[0]+" "+args[1]]; ok {
			return args[0] + " " + args[1]
		}
	}
	if len(args) > 0 {
		if _, ok := ganchos[args[0]]; ok {
			return args[0]
		}
	}
	return ""
}

// eventoGancho: evento de Claude Code de la orden; false si no es un gancho.
func eventoGancho(args []string) (string, bool) {
	k := claveGancho(args)
	return ganchos[k], k != ""
}

// nuncaMudo: si args es una orden de gancho, la corre como proceso hijo, traduce su salida y termina el proceso.
// Fuera de un gancho, o ya dentro del hijo, no hace nada.
// originales: los argumentos tal como llegaron; el hijo los traduce y avisa los nombres viejos una sola vez.
func nuncaMudo(args, originales []string) {
	clave := claveGancho(args)
	args = originales
	if clave == "" || os.Getenv(envGancho) == "1" {
		return
	}
	exe, err := os.Executable()
	if err != nil {
		avisarGancho(clave, fmt.Errorf("no encuentro el ejecutable de aquas: %w", err))
	}
	c := exec.Command(exe, args...)
	c.Env = append(os.Environ(), envGancho+"=1")
	c.Stdin = os.Stdin
	var out, errb bytes.Buffer
	c.Stdout, c.Stderr = &out, &errb
	err = c.Run()
	codigo := 0
	var x *exec.ExitError
	if errors.As(err, &x) {
		codigo = x.ExitCode()
	} else if err != nil {
		avisarGancho(clave, err)
	}
	switch codigo {
	case 2:
		_, _ = os.Stderr.Write(errb.Bytes())
		_, _ = os.Stdout.Write(out.Bytes())
		os.Exit(2)
	case 0:
		_, _ = os.Stderr.Write(errb.Bytes())
		_, _ = os.Stdout.Write(conAviso(clave, out.Bytes(), errb.String()))
		os.Exit(0)
	}
	avisarGancho(clave, errors.New(causa(errb.String(), codigo)))
}

// conAviso: con exit 0 el stderr no lo ve nadie; si trae algo, viaja como systemMessage junto a la salida del hijo.
func conAviso(clave string, out []byte, stderr string) []byte {
	var ls []string
	for _, l := range strings.Split(stderr, "\n") {
		if l = strings.TrimSpace(l); l != "" && !strings.Contains(l, avisoNombreViejo) {
			ls = append(ls, strings.TrimSpace(strings.TrimPrefix(l, "aquas:")))
		}
	}
	if len(ls) == 0 {
		return out
	}
	msg := "aquas " + clave + ": " + strings.Join(ls, " · ")
	salida := map[string]any{}
	texto := strings.TrimSpace(string(out))
	if texto != "" && json.Unmarshal(out, &salida) != nil { // texto plano: en estos eventos es contexto del agente
		salida = map[string]any{}
		if ev := ganchos[clave]; conContexto[ev] {
			salida["hookSpecificOutput"] = map[string]string{"hookEventName": ev, "additionalContext": texto}
		}
	}
	if salida == nil { // el hijo imprimió null
		salida = map[string]any{}
	}
	if previo, hay := salida["systemMessage"].(string); hay && previo != "" {
		msg = previo + " · " + msg
	}
	salida["systemMessage"] = msg
	b, _ := json.Marshal(salida)
	return append(b, '\n')
}

// causa: líneas de stderr del hijo sin el prefijo «aquas:» ni el aviso de nombre viejo.
func causa(stderr string, codigo int) string {
	var ls []string
	for _, l := range strings.Split(stderr, "\n") {
		l = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(l), "aquas:"))
		if l != "" && !strings.Contains(l, avisoNombreViejo) {
			ls = append(ls, l)
		}
	}
	if len(ls) == 0 {
		return fmt.Sprintf("salió con código %d sin decir la causa", codigo)
	}
	return strings.Join(ls, " · ")
}

// avisarGancho: el fallo llega a quien corresponde y el proceso termina con 0 (aviso) o 2 (freno).
func avisarGancho(clave string, err error) {
	codigo, out, errTxt := fallaGancho(clave, err)
	_, _ = fmt.Fprint(os.Stdout, out)
	_, _ = fmt.Fprint(os.Stderr, errTxt)
	os.Exit(codigo)
}

// fallaGancho: código, stdout y stderr del fallo de un gancho; 2 con la causa si frena, 0 con aviso si no.
func fallaGancho(clave string, err error) (int, string, string) {
	msg := "aquas " + clave + " falló: " + err.Error()
	if clave == "aprobacion" {
		msg = "aprobación NO registrada: " + err.Error() + ". Arregla la causa y pide la línea apruebo otra vez."
	}
	if frenaSiFalla[clave] {
		return 2, "", msg + "\n"
	}
	salida := map[string]any{"systemMessage": msg}
	if ev := ganchos[clave]; conContexto[ev] {
		salida["hookSpecificOutput"] = map[string]string{"hookEventName": ev, "additionalContext": msg}
	}
	b, _ := json.Marshal(salida)
	return 0, string(b) + "\n", ""
}
