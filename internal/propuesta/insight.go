package propuesta

import (
	"regexp"
	"strings"
)

// defecto: número de un defecto registrado con `aquas defect new`.
var defecto = regexp.MustCompile(`\bD\d{6}\b`)

// RevisarInsight devuelve "" si pasa; si no, el motivo del freno: insight PROPUESTO sin defecto registrado (E001/F002/S016).
// Solo mira el valor después de «Estado:» en la línea del campo Insights; la palabra citada en otra parte no cuenta (X01, D000009).
func RevisarInsight(msg string) string {
	for _, l := range strings.Split(strings.ReplaceAll(msg, "\r\n", "\n"), "\n") {
		i := strings.LastIndex(l, "Estado:")
		if !strings.Contains(l, "Insights:") || i < 0 || defecto.MatchString(l) {
			continue
		}
		if !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(l[i+len("Estado:"):])), "PROPUESTO") {
			continue
		}
		return "Insight PROPUESTO sin defecto: un control no implementado se pierde en el chat. Regístralo con " +
			"`aquas defect new firma=… que=… causa=… arreglo=… cercioro=… relacion=…` y nombra su número (D000NNN) en el campo Insights."
	}
	return ""
}
