package guardia

import (
	"fmt"
	"regexp"
	"strings"
)

// E000/F008/S001 (D000077): órdenes que solo lanza Claude Code como hook. Si el agente las
// lanza por Bash con un JSON inventado, falsifica un apruebo, un veredicto o el contador del tope.
//
// Se busca en el texto del comando, sin heredocs, y no palabra por palabra: así se cubren
// redirecciones, prefijos (env, exec, nohup, time), $(…), `sh -lc`, if/then, llaves y comillas.
// Freno probabilístico: xargs por stdin, variables y nombres armados lo evaden (ver contrato).
// Precio aceptado: un texto que cite la orden, con o sin comillas, también se bloquea.

// OrdenesHook: cada orden reservada, con nombre nuevo y viejo; las de dos palabras admiten cruces.
var OrdenesHook = []string{
	"approval", "aprobacion", "agent-close", "cierre-agente", "budget", "presupuesto",
	"turn", "turno", "proposal", "propuesta", "guard", "guardia", "hook",
	"memory start", "memory close", "memory inicio", "memory cierre",
	"memoria start", "memoria close", "memoria inicio", "memoria cierre",
}

// reHook: quién llama (binario de aquas, lanzador, go run o una variable) + redirecciones + orden.
var reHook = func() *regexp.Regexp {
	var ords []string
	for _, o := range OrdenesHook {
		ords = append(ords, strings.ReplaceAll(regexp.QuoteMeta(o), " ", `\s+`))
	}
	llamada := `(?:\S*?(?:aquas|aquas\.exe|aq\.exe|aquas\.sh)["']?` + // binario o lanzador, con ruta
		`|\bgo\s+run(?:\s+-\S+)*\s+\S+` + // go run [opciones] <paquete o archivo>: «.», main.go
		`|\$\{?\w+\}?["']?)` // variable usada como comando
	redir := `(?:\s+\d*[<>]+&?\s*\S+)*`
	return regexp.MustCompile(`(?i)(?:^|[\s;&|(){}!` + "`" + `"'=])` + llamada + redir + `\s+(` + strings.Join(ords, "|") + `)(?:$|[\s;&|)<>` + "`" + `])`)
}()

// reSinComillas: comillas, escapes y continuaciones de línea; bash los quita antes de ejecutar (A2).
var reSinComillas = regexp.MustCompile(`\\\r?\n|["'\\]`)

// RevisarHook: motivo de bloqueo si el comando lanza una orden de hook; «» si no.
func RevisarHook(comando string) string {
	m := reHook.FindStringSubmatch(reSinComillas.ReplaceAllString(sinHeredocs(comando), ""))
	if m == nil {
		return ""
	}
	orden := strings.Join(strings.Fields(m[1]), " ")
	return fmt.Sprintf("orden de hook «aquas %s»: solo la lanza Claude Code, no el agente por Bash (D000077)", orden)
}
