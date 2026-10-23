// Package encargo: contrato de cierre de los agentes worker y auditor (E001/F006/S002).
// Canónico: gobierno/canonico/reglas/02_COMUNICACION_ENTRE_AGENTES.md.
package encargo

import (
	"regexp"
	"slices"
	"strings"
)

const canonico = "gobierno/canonico/reglas/02_COMUNICACION_ENTRE_AGENTES.md"

type contrato struct {
	primera *regexp.Regexp // primera línea no vacía
	campo   string         // nombre de la primera línea, para el mensaje
	resto   []string       // campos obligatorios después
}

var (
	worker  = contrato{regexp.MustCompile(`^estado: (HECHO|PARCIAL|BLOQUEADO)\s*$`), "estado: HECHO | PARCIAL | BLOQUEADO", []string{"archivos", "evidencia", "bloqueo", "fuera_de_alcance"}}
	auditor = contrato{regexp.MustCompile(`^veredicto: (APROBADO|CON_HALLAZGOS|RECHAZADO)\s*$`), "veredicto: APROBADO | CON_HALLAZGOS | RECHAZADO", []string{"hallazgos", "no_revisado"}}
	porTipo = map[string]contrato{"worker-sonnet": worker, "worker-haiku": worker, "auditor": auditor}
)

// Revisar decide el hook SubagentStop: "" deja cerrar; si no, qué falta en el cierre.
// stop_hook_active: ya se pidió una vez → deja cerrar, para no entrar en bucle.
func Revisar(tipo, mensaje string, stopHookActive bool) string {
	c, ok := porTipo[tipo]
	if !ok || stopHookActive {
		return ""
	}
	// Líneas no vacías, sin la cerca ``` que un modelo puede copiar del canónico.
	var lineas []string
	for _, l := range strings.Split(strings.ReplaceAll(mensaje, "\r\n", "\n"), "\n") {
		if t := strings.TrimSpace(l); t != "" && !strings.HasPrefix(t, "```") {
			lineas = append(lineas, l)
		}
	}
	if len(lineas) == 0 || !c.primera.MatchString(strings.TrimSpace(lineas[0])) {
		return falla("la primera línea debe ser «" + c.campo + "»; el cierre es el único texto, sin prosa")
	}
	// Cada línea siguiente abre un campo («campo: valor») o continúa el anterior
	// («- …», «· …» o sangría). Otra cosa es texto fuera del cierre.
	valor := map[string]bool{} // campo visto → tiene valor
	actual := ""
	for _, l := range lineas[1:] {
		t := strings.TrimSpace(l)
		if campo, v, ok := strings.Cut(t, ":"); ok && slices.Contains(c.resto, campo) {
			actual = campo
			valor[campo] = strings.TrimSpace(v) != ""
			continue
		}
		if actual != "" && (strings.HasPrefix(t, "- ") || strings.HasPrefix(t, "· ") || t != l) {
			valor[actual] = true
			continue
		}
		return falla("texto fuera del cierre: «" + t + "»; el cierre es el único texto")
	}
	var faltan, vacios []string
	for _, campo := range c.resto {
		conValor, hay := valor[campo]
		if !hay {
			faltan = append(faltan, campo)
		} else if !conValor {
			vacios = append(vacios, campo)
		}
	}
	if len(faltan) > 0 {
		return falla("faltan campos: " + strings.Join(faltan, ", ") + " (usa «—» si no aplica)")
	}
	if len(vacios) > 0 {
		return falla("campos vacíos: " + strings.Join(vacios, ", ") + " (usa «—» si no aplica)")
	}
	return ""
}

// falla: mensaje de cierre inválido que pide reescribirlo según el canónico.
func falla(motivo string) string {
	return "cierre inválido: " + motivo + ". Reescribe solo el CIERRE según " + canonico + "."
}
