package propuesta

import (
	"strings"
	"unicode"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/backlog"
)

// operativas: palabras que vuelven operativo un mensaje corto (contrato de E000/F010/S001).
var operativas = map[string]bool{"sí": true, "si": true, "dale": true, "ok": true, "sigue": true, "continúa": true, "listo": true, "hazlo": true, "aplica": true,
	"recomendado": true, "aprobado": true, "sube": true} // E000/F010/X01: medidas en el registro real, decisión humana 2026-11-02

// partesDesafio: lo que el bloque nombra después de su línea DESAFÍO.
var partesDesafio = []string{"Supuestos", "Contraargumentos", "Riesgos", "Recomendación"}

// Operativo: «apruebo …», o 6 palabras o menos con alguna palabra operativa. Heurística: control probabilístico.
func Operativo(msg string) bool {
	m := strings.ToLower(strings.TrimSpace(msg))
	palabras := strings.FieldsFunc(m, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
	if len(palabras) > 0 && palabras[0] == "apruebo" { // A2.2: «Apruebo», «apruebo.» y «apruebo X» empiezan con apruebo
		return true
	}
	if len(palabras) > 6 {
		return false
	}
	for _, p := range palabras {
		if operativas[p] {
			return true
		}
	}
	return false
}

// RevisarDesafio (E000/F010/S001): si el último mensaje humano del registro es un pedido nuevo, la respuesta debe
// traer una línea que empieza con DESAFÍO y, después, las cuatro partes; si no, el motivo del freno. No revisa el
// turno que corrige otro freno (activo), ni sin registro legible o sin mensaje humano.
func RevisarDesafio(registro, respuesta string, activo bool) string {
	if activo || registro == "" {
		return ""
	}
	humano, err := backlog.UltimoMensajeHumano(registro)
	if err != nil || humano == "" || Operativo(humano) {
		return ""
	}
	var resto strings.Builder
	hay := false
	for _, l := range strings.Split(respuesta, "\n") {
		if hay {
			resto.WriteString(l + "\n")
		} else if strings.HasPrefix(strings.TrimLeft(l, " \t*#>_-"), "DESAFÍO") {
			hay = true
		}
	}
	var falta []string
	if !hay {
		falta = append(falta, "la línea DESAFÍO")
	}
	for _, p := range partesDesafio {
		if !strings.Contains(resto.String(), p) {
			falta = append(falta, p)
		}
	}
	if len(falta) == 0 {
		return ""
	}
	return "pedido nuevo sin DESAFÍO completo: falta " + strings.Join(falta, ", ") +
		". Agrega una línea DESAFÍO con Supuestos, Contraargumentos (o «sin objeción real»), Riesgos y Recomendación."
}
