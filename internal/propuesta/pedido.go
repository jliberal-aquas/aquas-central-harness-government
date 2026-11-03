package propuesta

import (
	"regexp"
	"strings"
)

// E001/F017/S002: qué pide el agente al humano en los campos de pedido del cierre. Solo esos dos campos cuentan
// (con sus viñetas): en el cuerpo, el agente puede citar lo que dijo el humano sin pedirlo.
var (
	reCampoPedido = regexp.MustCompile(`\*\*(Decisión|¿Qué necesito de ti\?):\*\*`)
	// El id termina en algo que no es letra, dígito ni «/»: «E001/F017/S0031» no se lee como E001/F017 (A1.8).
	reApruebo = regexp.MustCompile(`(?i)\bapruebo\s+(E\d{3}/F\d{3}(?:/S\d{3})?)(?:[^\w/]|$)`)
	// Solo las etiquetas del cierre cortan un campo; una viñeta «**Rama:** …» sigue dentro (A4.3).
	reEtiqueta      = regexp.MustCompile(`\*\*(Resultado del turno|Archivos relevantes|Plan|Objetivo|Información|Decisión|Insights|¿Salió algo mal\?|¿Qué necesito de ti\?|Próximo movimiento):\*\*`)
	reTopeDelPedido = regexp.MustCompile(`(?i)^[^,;.]*\btope\s+(\d+|N)\b`) // tope en la misma cláusula del apruebo (A2.8)
	reConPush       = regexp.MustCompile(`(?i)^[^,;.]*\bcon\s+push\b`)
	reNada          = regexp.MustCompile(`(?i)^[ \t*]*(nada|—)(\W|$)`)
	reVineta        = regexp.MustCompile(`^(- |• |\* |\d+[.)] )`)
)

// Pedido: un «apruebo <id>» pedido al humano; Push si pide «con push» en la misma cláusula.
type Pedido struct {
	ID   string
	Push bool
}

// camposDePedido: el texto de cada campo «Decisión» y «¿Qué necesito de ti?», con las viñetas que lo siguen.
// Un campo cuya primera línea empieza con «Nada» o «—» no pide nada (A1.3, A2.5, A2.6).
func camposDePedido(msg string) []string {
	var campos []string
	lineas := strings.Split(strings.ReplaceAll(msg, "\r\n", "\n"), "\n")
	var abierta byte
	largo := 0
	for i, l := range lineas {
		// Un bloque de código (``` o ~~~, misma regla que Revisar) no es el cierre (A4.6, A5.3).
		if c, n := cercaDe(strings.TrimSpace(l)); n > 0 && (abierta == 0 || (c == abierta && n >= largo)) {
			if abierta == 0 {
				abierta, largo = c, n
			} else {
				abierta = 0
			}
			continue
		}
		loc := reCampoPedido.FindStringIndex(l)
		if abierta != 0 || loc == nil || reNada.MatchString(l[loc[1]:]) {
			continue
		}
		texto := l[loc[1]:]
		for _, sig := range lineas[i+1:] {
			t := strings.TrimSpace(sig)
			if !reVineta.MatchString(t) || reEtiqueta.MatchString(t) { // otra etiqueta del cierre corta el campo (A3.1)
				break
			}
			texto += "\n" + t
		}
		campos = append(campos, texto)
	}
	return campos
}

// PedidosDeApruebo devuelve cada «apruebo <id>» pedido en los campos de pedido (id en mayúsculas), sin los de
// feature que piden un tope nuevo. Un «dale» no se revisa: sin id es ambiguo y queda como criterio humano (A3).
func PedidosDeApruebo(msg string) (pedidos []Pedido) {
	for _, c := range camposDePedido(msg) {
		ms := reApruebo.FindAllStringSubmatchIndex(c, -1)
		for k, m := range ms {
			// La cola llega hasta el siguiente «apruebo», otra cita que abre o el fin de línea (A4.5). La comilla que
			// cierra no corta: «apruebo X» con push sigue siendo un pedido con push (A5.2).
			fin := len(c)
			if k+1 < len(ms) {
				fin = ms[k+1][0]
			}
			cola := c[m[3]:fin]
			if j := strings.IndexAny(cola, "«\n"); j >= 0 {
				cola = cola[:j]
			}
			id := strings.ToUpper(c[m[2]:m[3]])
			if !strings.Contains(id, "/S") && reTopeDelPedido.MatchString(cola) {
				continue // pedir tope lo exige el freno de presupuesto (A1.7)
			}
			pedidos = append(pedidos, Pedido{ID: id, Push: reConPush.MatchString(cola)})
		}
	}
	return pedidos
}
