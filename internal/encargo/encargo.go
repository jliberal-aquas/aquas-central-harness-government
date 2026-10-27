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
	segunda *regexp.Regexp // si no es nil, segunda línea no vacía (auditor: veredicto tras el id)
	campo2  string         // nombre de la segunda línea, para el mensaje
	resto   []string       // campos obligatorios después
}

var (
	reIDAuditor = regexp.MustCompile(`^id: (E\d{3}-F\d{3}-S\d{3})-A\d+\s*$`)
	reVeredicto = regexp.MustCompile(`^veredicto: (APROBADO|CON_HALLAZGOS|RECHAZADO)\s*$`)
	worker      = contrato{regexp.MustCompile(`^estado: (HECHO|PARCIAL|BLOQUEADO)\s*$`), "estado: HECHO | PARCIAL | BLOQUEADO", nil, "", []string{"archivos", "evidencia", "bloqueo", "fuera_de_alcance"}}
	// «id: —»: auditoría sin tramo (feature, arreglo u otra auditoría); válido, el gancho no anota (S002-T2 H2).
	reIDAuditorOSin = regexp.MustCompile(`^id: (E\d{3}-F\d{3}-S\d{3}-A\d+|—)\s*$`)
	auditor         = contrato{reIDAuditorOSin, "id: E000-F000-S000-A1 (id del encargo del auditor; «id: —» si no audita un tramo)", reVeredicto, "veredicto: APROBADO | CON_HALLAZGOS | RECHAZADO", []string{"hallazgos", "no_revisado"}}
	porTipo         = map[string]contrato{"worker-sonnet": worker, "worker-haiku": worker, "auditor": auditor}
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
	if len(lineas) == 0 { // D000053: un cierre vacío nunca pasa en silencio
		return falla("cierre vacío: la primera línea debe ser «" + c.campo + "»; el cierre es el único texto, sin prosa")
	}
	if !c.primera.MatchString(strings.TrimSpace(lineas[0])) {
		return falla("la primera línea debe ser «" + c.campo + "»; el cierre es el único texto, sin prosa")
	}
	resto := lineas[1:]
	if c.segunda != nil {
		if len(resto) == 0 || !c.segunda.MatchString(strings.TrimSpace(resto[0])) {
			return falla("la segunda línea debe ser «" + c.campo2 + "»")
		}
		resto = resto[1:]
	}
	// Cada línea siguiente abre un campo («campo: valor») o continúa el anterior
	// («- …», «· …» o sangría). Otra cosa es texto fuera del cierre.
	valor := map[string]bool{} // campo visto → tiene valor
	actual := ""
	for _, l := range resto {
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

// Veredicto extrae del cierre de un auditor ya válido el id del tramo («E001/F014/S002») y su veredicto.
func Veredicto(mensaje string) (id, veredicto string, ok bool) {
	var ls []string
	for _, l := range strings.Split(strings.ReplaceAll(mensaje, "\r\n", "\n"), "\n") {
		if t := strings.TrimSpace(l); t != "" && !strings.HasPrefix(t, "```") {
			ls = append(ls, t)
		}
	}
	if len(ls) < 2 {
		return "", "", false
	}
	a, b := reIDAuditor.FindStringSubmatch(ls[0]), reVeredicto.FindStringSubmatch(ls[1])
	if a == nil || b == nil {
		return "", "", false
	}
	return strings.ReplaceAll(a[1], "-", "/"), b[1], true
}

// falla: mensaje de cierre inválido que pide reescribirlo según el canónico.
func falla(motivo string) string {
	return "cierre inválido: " + motivo + ". Reescribe solo el CIERRE según " + canonico + "."
}
