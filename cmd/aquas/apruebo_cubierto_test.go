package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// cierre: corre el hook Stop «propuesta» con el mensaje del agente (stop_hook_active: no revisa DESAFÍO).
func cierre(t *testing.T, raiz, msg string) salida {
	t.Helper()
	b, _ := json.Marshal(map[string]any{"last_assistant_message": msg, "stop_hook_active": true})
	return aquas(t, raiz, string(b), "proposal")
}

// TestNoPideAprueboCubierto (E001/F017/S002, D000069, D000084, D000099): el cierre frena si el agente pide, en sus
// campos de pedido, un apruebo que la aprobación de la feature ya cubre. El «dale» no se revisa (criterio humano).
// Rojo cuando: el agente pide apruebo de un slice o una feature cubiertos y el cierre pasa.
// Verde cuando: frena esos pedidos y deja pasar los no cubiertos, el tope y las citas fuera de los campos.
func TestNoPideAprueboCubierto(t *testing.T) {
	raiz := repoFeature(t)
	necesito := func(pedido string) string {
		return "INFORMACIÓN: algo.\n\n🧙‍♂️ **¿Qué necesito de ti?:** " + pedido + "\n🎯 **Próximo movimiento:** seguir."
	}
	decision := func(pedido string) string {
		return "INFORMACIÓN: algo.\n\n⁉️ **Decisión:** " + pedido + "\n🧙‍♂️ **¿Qué necesito de ti?:** Nada."
	}
	pasa := func(caso, msg string) {
		t.Helper()
		if s := cierre(t, raiz, msg); s.codigo != 0 {
			t.Errorf("%s: el cierre debe pasar; dio %+v", caso, s)
		}
	}
	frena := func(caso, msg string) {
		t.Helper()
		if s := cierre(t, raiz, msg); s.codigo != 2 || !strings.Contains(s.err, "ya cubre") {
			t.Errorf("%s: el cierre debe frenar diciendo que la aprobación ya lo cubre; dio %+v", caso, s)
		}
	}

	// Sin apruebo de la feature (BORRADOR): pedir apruebo pasa.
	pasa("feature sin apruebo, slice", necesito("«apruebo E001/F002/S003»"))
	pasa("feature sin apruebo, feature", necesito("«apruebo E001/F002 tope 300»"))
	pasa("feature sin apruebo, dale", necesito("un «dale» para seguir"))

	if s := aquas(t, raiz, mensajeHumano(t, "apruebo E001/F002 con push"), "aprobacion"); s.codigo != 0 {
		t.Fatalf("apruebo de la feature: %+v", s)
	}
	// Feature aprobada: el slice de la foto y la feature misma ya están cubiertos; el «dale» no se revisa.
	frena("slice cubierto", necesito("«apruebo E001/F002/S003»"))
	frena("slice cubierto, en Decisión", decision("¿apruebas E001/F002/S003? Escribe «apruebo E001/F002/S003»."))
	frena("feature ya aprobada", necesito("«apruebo E001/F002»"))
	pasa("dale no se revisa", necesito("«dale» para arrancar el siguiente slice"))
	pasa("Dale con mayúscula no se revisa", necesito("Dale, sigue."))
	// Pedir tope sí pasa (el freno de presupuesto lo exige).
	pasa("tope nuevo", necesito("«apruebo E001/F002 tope 500» para seguir: se agotó el tope"))
	// Id mal escrito o sin id: no se reconoce como pedido cubierto.
	pasa("id mal escrito", necesito("«apruebo E001/F02/S003»"))
	pasa("sin id", necesito("tu apruebo"))
	pasa("id con un dígito de más", necesito("«apruebo E001/F002/S0031»"))
	frena("id en minúsculas", necesito("«apruebo e001/f002/s003»"))
	frena("duplicada: uno cubierto y otro no", necesito("«apruebo E009/F009/S001» y «apruebo E001/F002/S003»"))
	frena("pedido en una viñeta del campo", "INFORMACIÓN: algo.\n\n🧙‍♂️ **¿Qué necesito de ti?:** dos cosas\n- «apruebo E001/F002/S003»\n🎯 **Próximo movimiento:** x")
	frena("mencionar el tope sin número no exime", necesito("«apruebo E001/F002/S003» (vamos en 40% del tope)"))
	pasa("dale para una acción irreversible no cubierta", necesito("«dale» para borrar la rama vieja"))
	pasa("campo Nada que cita la aprobación", necesito("Nada: tu «apruebo E001/F002» ya cubre S003."))
	pasa("campo nada en minúscula", necesito("nada; tu «apruebo E001/F002» ya cubre S003."))
	pasa("dale para otra cosa y luego seguir", necesito("«dale» para borrar la rama vieja; luego seguimos con S003"))
	frena("viñeta con la línea del campo vacía", "INFORMACIÓN: algo.\n\n🧙‍♂️ **¿Qué necesito de ti?:**\n- «apruebo E001/F002/S003»\n🎯 **Próximo movimiento:** x")
	frena("lista numerada", "INFORMACIÓN: algo.\n\n🧙‍♂️ **¿Qué necesito de ti?:** dos cosas\n1. «apruebo E001/F002/S003»\n🎯 **Próximo movimiento:** x")
	frena("tope de otra cosa en la misma línea", necesito("apruebo E001/F002/S003, y aparte el tope 500 de otra feature"))
	frena("tope de otra cláusula, feature (A4.1)", necesito("apruebo E001/F002, y aparte el tope 500 de otra feature"))
	frena("dos pedidos sin comillas (A4.5)", necesito("apruebo E009/F009/S001 y apruebo E001/F002/S003"))
	frena("viñeta con etiqueta propia (A4.3)", "INFORMACIÓN: algo.\n\n🧙‍♂️ **¿Qué necesito de ti?:** dos respuestas\n1. **Rama:** sí o no\n2. **Siguiente:** «apruebo E001/F002/S003»")
	pasa("campo con raya (A4.7)", necesito("— tu «apruebo E001/F002/S003» ya lo cubre"))
	pasa("pedido dentro de un bloque de código (A4.6)", "Ejemplo:\n```\n🧙‍♂️ **¿Qué necesito de ti?:** «apruebo E001/F002/S003»\n```\n🧙‍♂️ **¿Qué necesito de ti?:** Nada.")
	pasa("bloque con cerca ~~~ (A5.3)", "Ejemplo:\n~~~\n🧙‍♂️ **¿Qué necesito de ti?:** «apruebo E001/F002/S003»\n~~~\n🧙‍♂️ **¿Qué necesito de ti?:** Nada.")
	// La cola de cada pedido termina en el siguiente «apruebo» o en otra cita que abre (A5.1).
	frena("el tope del segundo pedido no exime al primero", necesito("apruebo E001/F002 y apruebo E009/F009 tope 500"))
	frena("el tope de otra cita no exime", necesito("«apruebo E001/F002» y aparte «revisa el tope 500»"))
	// Citas fuera de los campos de pedido no frenan (el agente cuenta lo que dijo el humano).
	pasa("cita en el cuerpo", "INFORMACIÓN: tu «apruebo E001/F002/S003» y tu «dale» quedaron registrados.\n\n🧙‍♂️ **¿Qué necesito de ti?:** Nada.")
	pasa("sin campos", "Sigo con E001/F002/S003.")
	// Cierre en viñetas: la etiqueta siguiente corta el campo; su cita no cuenta como pedido (A3.1).
	pasa("cierre en viñetas", "INFORMACIÓN: algo.\n\n- ⁉️ **Decisión:** ¿borro la rama vieja?\n- 🧙‍♂️ **¿Qué necesito de ti?:** sí o no sobre la rama\n- 🎯 **Próximo movimiento:** arrancar S003; tu «apruebo E001/F002/S003» ya lo cubre")

	// Slice nuevo fuera de la foto: pedir el apruebo de la feature otra vez pasa (D000084).
	if s := aquas(t, raiz, "", "backlog", "new", "id=E001/F002/S004", "titulo=N", "prioridad=P1", "motivo=m", "relacion=nueva",
		"que_hace=hace algo", "prueba=./p:TestBandera=pide el arreglo|r|v"); s.codigo != 0 {
		t.Fatalf("backlog new: %+v", s)
	}
	pasa("slice fuera de la foto", necesito("«apruebo E001/F002/S004»"))
	pasa("feature con slice fuera de la foto", necesito("«apruebo E001/F002 con push»"))
	frena("el slice de la foto sigue cubierto", necesito("«apruebo E001/F002/S003»"))

	pasa("feature inexistente", necesito("«apruebo E009/F009/S001»"))

	// Aprobación sin push: pedir «con push» pasa (acción externa no cubierta, A2.2); sin «con push», frena.
	if s := aquas(t, raiz, mensajeHumano(t, "apruebo E001/F002"), "aprobacion"); s.codigo != 0 {
		t.Fatalf("apruebo sin push: %+v", s)
	}
	pasa("pedir con push sobre apruebo sin push", necesito("«apruebo E001/F002 con push» para poder subir"))
	frena("feature ya aprobada sin push", necesito("«apruebo E001/F002»"))
	frena("con push en otra cláusula no cuenta (A4.1)", necesito("apruebo E001/F002, y aparte con push la rama vieja"))
	pasa("con push tras la comilla que cierra (A5.2)", necesito("«apruebo E001/F002» con push para poder subir"))

	// Cada caso cambia una cosa del repo, comprueba que pasa y la repone: el control vuelve a frenar (A2.3, A2.4).
	cambiar := func(caso, rel string, cambio func(string) string) {
		t.Helper()
		ruta := filepath.Join(raiz, filepath.FromSlash(rel))
		orig, err := os.ReadFile(ruta)
		if err != nil {
			t.Fatal(err)
		}
		escribir(t, raiz, rel, cambio(string(orig)))
		pasa(caso, necesito("«apruebo E001/F002/S003»"))
		escribir(t, raiz, rel, string(orig))
		frena(caso+" (repuesto)", necesito("«apruebo E001/F002/S003»"))
	}
	reEstado := regexp.MustCompile(`"estado":\s*"[A-Z_]+"`)
	for _, estado := range []string{"CERRADO", "CANCELADO"} {
		cambiar("feature "+estado, "trabajo/activo/E001-F002.json", func(s string) string {
			return reEstado.ReplaceAllString(s, `"estado": "`+estado+`"`)
		})
	}
	cambiar("prueba aprobada del slice cambiada", "trabajo/activo/E001-F002-S003.json", func(s string) string {
		return regexp.MustCompile(`"verde_cuando":\s*"[^"]*"`).ReplaceAllString(s, `"verde_cuando": "otro"`)
	})
	cambiar("contrato de la feature cambiado (A1.1)", "docs/E001/F002/README.md", func(string) string {
		return "# F002\n\n## Para qué existe\n\nContrato nuevo de la feature.\n"
	})
	cambiar("backlog ilegible", "trabajo/activo/E001-F002-S004.json", func(string) string { return "{roto" })
}
