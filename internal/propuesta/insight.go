package propuesta

import (
	"regexp"
	"strings"
)

// defecto: número de un defecto registrado con `aquas defecto nuevo`.
var defecto = regexp.MustCompile(`\bD\d{6}\b`)

// RevisarInsight devuelve "" si pasa; si no, el motivo del freno: insight PROPUESTO sin defecto registrado (E001/F002/S016).
// Solo mira la línea del campo Insights del cierre; un PROPUESTO en otra parte no cuenta.
func RevisarInsight(msg string) string {
	for _, l := range strings.Split(strings.ReplaceAll(msg, "\r\n", "\n"), "\n") {
		if !strings.Contains(l, "Insights:") || !strings.Contains(l, "PROPUESTO") || defecto.MatchString(l) {
			continue
		}
		return "Insight PROPUESTO sin defecto: un control no implementado se pierde en el chat. Regístralo con " +
			"`aquas defecto nuevo firma=… que=… causa=… arreglo=… cercioro=…` y nombra su número (D000NNN) en el campo Insights."
	}
	return ""
}
