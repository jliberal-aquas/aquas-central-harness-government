package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// E001/F010/S001 sobre el binario real: el contrato aprobado antes de construir.

const paginaS003 = "# S003\n\n## Para qué existe\n\nEl humano aprueba lo que se va a construir.\n\n## Cómo se usa\n\nFuera del contrato.\n"

// repoContrato: raíz con módulo de juguete con una prueba roja, y el slice E001/F002/S003 en BORRADOR.
func repoContrato(t *testing.T) string {
	t.Helper()
	raiz := t.TempDir()
	escribir(t, raiz, "go.mod", "module juguete\n\ngo 1.26\n")
	escribir(t, raiz, "p/p_test.go", "package p\n\nimport \"testing\"\n\nfunc TestBandera(t *testing.T) { t.Fatal(\"falta el arreglo\") }\n")
	escribir(t, raiz, "docs/E001/F002/S003.md", paginaS003)
	for _, args := range [][]string{
		{"nuevo", "id=E001", "titulo=G", "prioridad=P1", "motivo=m", "relacion=nueva", "que_hace=hace algo", "prueba=./p:TestBandera=pide el arreglo|r|v"},
		{"nuevo", "id=E001/F002", "titulo=B", "prioridad=P1", "motivo=m", "relacion=nueva", "que_hace=hace algo", "prueba=./p:TestBandera=pide el arreglo|r|v"},
		{"nuevo", "id=E001/F002/S003", "titulo=R", "prioridad=P1", "motivo=m", "relacion=nueva", "que_hace=hace algo", "prueba=./p:TestBandera=pide el arreglo|r|v"},
	} {
		if s := aquas(t, raiz, "", append([]string{"backlog"}, args...)...); s.codigo != 0 {
			t.Fatalf("backlog %v: %+v", args, s)
		}
		rojoIntegral(t, raiz, args)
	}
	if s := aquas(t, raiz, "", "backlog", "alcance", "E001/F002/S003", "p/**"); s.codigo != 0 {
		t.Fatalf("alcance: %+v", s)
	}
	return raiz
}

func mensajeHumano(prompt string) string {
	b, _ := json.Marshal(map[string]string{"session_id": "s1", "hook_event_name": "UserPromptSubmit", "prompt": prompt})
	return string(b)
}

// Rojo cuando: el texto del agente que pasa por una herramienta registra una aprobación, o el mensaje del humano no la registra.
// Verde cuando: el mensaje del humano guarda fecha y huella en la ficha; ninguna herramienta guarda nada.
func TestHookRegistraApruebo(t *testing.T) {
	raiz := repoContrato(t)
	ficha := "trabajo/activo/E001-F002-S003.json"
	desdeHerramienta := []string{
		`{"hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{"command":"echo apruebo E001/F002/S003"}}`,
		`{"hook_event_name":"PostToolUse","tool_name":"Write","prompt":"apruebo E001/F002/S003","tool_input":{"content":"apruebo E001/F002/S003"}}`,
		`{"hook_event_name":"Stop","last_assistant_message":"apruebo E001/F002/S003"}`,
	}
	for _, in := range desdeHerramienta {
		aquas(t, raiz, in, "aprobacion")
		if _, hay := fichaEnDisco(t, raiz, ficha)["aprobacion"]; hay {
			t.Fatalf("la entrada %s no es un mensaje del humano: no registra aprobación", in)
		}
	}
	if s := aquas(t, raiz, mensajeHumano("apruebo E001/F002/S004, otro slice"), "aprobacion"); s.codigo != 0 {
		t.Fatalf("id sin ficha: no bloquea el mensaje; dio %+v", s)
	}
	if _, hay := fichaEnDisco(t, raiz, ficha)["aprobacion"]; hay {
		t.Fatal("aprobar otro id no aprueba S003")
	}
	if s := aquas(t, raiz, mensajeHumano("  apruebo E001/F002/S003 y sigue.\nGracias."), "aprobacion"); s.codigo != 0 || s.out != "" {
		t.Fatalf("mensaje del humano: exit 0 sin stdout; dio %+v", s)
	}
	var ap struct{ Fecha, Huella, Mensaje string }
	if err := json.Unmarshal(fichaEnDisco(t, raiz, ficha)["aprobacion"], &ap); err != nil {
		t.Fatalf("la ficha guarda la aprobación: %v", err)
	}
	if ap.Fecha == "" || !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(ap.Huella) || !strings.Contains(ap.Mensaje, "apruebo E001/F002/S003") {
		t.Errorf("aprobación con fecha, huella SHA-256 y mensaje; dio %+v", ap)
	}
}

// Rojo cuando: el slice arranca sin aprobación o tras cambiar el contrato aprobado.
// Verde cuando: sin aprobación se rechaza; tras apruebo se acepta; con el contrato cambiado se rechaza hasta aprobar de nuevo.
func TestContratoAprobadoAntesDeConstruir(t *testing.T) {
	raiz := repoContrato(t)
	const id = "E001/F002/S003"
	if s := aquas(t, raiz, "", "backlog", "estado", id, "EN_CURSO"); s.codigo != 1 || !strings.Contains(s.err, "apruebo "+id) {
		t.Fatalf("sin aprobación: EN_CURSO se rechaza y pide «apruebo %s»; dio %+v", id, s)
	}
	if s := aquas(t, raiz, "", "backlog", "rojo", id); s.codigo != 1 || !strings.Contains(s.err, "apruebo "+id) {
		t.Fatalf("sin aprobación: el rojo se rechaza y pide «apruebo %s»; dio %+v", id, s)
	}
	aquas(t, raiz, mensajeHumano("apruebo "+id), "aprobacion")
	escribir(t, raiz, "docs/E001/F002/S003.md", strings.Replace(paginaS003, "lo que se va a construir", "otra cosa", 1))
	if s := aquas(t, raiz, "", "backlog", "estado", id, "EN_CURSO"); s.codigo != 1 || !strings.Contains(s.err, "apruebo "+id) {
		t.Fatalf("contrato cambiado: EN_CURSO se rechaza y pide aprobar de nuevo; dio %+v", s)
	}
	aquas(t, raiz, mensajeHumano("apruebo "+id), "aprobacion")
	if s := aquas(t, raiz, "", "backlog", "estado", id, "EN_CURSO"); s.codigo != 0 {
		t.Fatalf("con aprobación vigente: EN_CURSO se acepta; dio %+v", s)
	}
	if s := aquas(t, raiz, "", "backlog", "rojo", id); s.codigo != 0 {
		t.Fatalf("con aprobación vigente: el rojo se acepta; dio %+v", s)
	}
}

// Rojo cuando: una frase que no es una aprobación (negada, pregunta, a media línea) registra la aprobación.
// Verde cuando: solo una línea que empieza por «apruebo <id>» y no termina en «?» guarda la aprobación.
func TestApruebaSoloLineaAfirmativa(t *testing.T) {
	raiz := repoContrato(t)
	ficha := "trabajo/activo/E001-F002-S003.json"
	for _, p := range []string{
		"no apruebo E001/F002/S003",
		"Todavía no apruebo E001/F002/S003, falta revisar",
		"apruebo E001/F002/S003?",
		"¿apruebo E001/F002/S003?",
		"si apruebo E001/F002/S003 qué pasa",
	} {
		aquas(t, raiz, mensajeHumano(p), "aprobacion")
		if _, hay := fichaEnDisco(t, raiz, ficha)["aprobacion"]; hay {
			t.Fatalf("%q no es una aprobación: no se registra", p)
		}
	}
	aquas(t, raiz, mensajeHumano("\n apruebo E001/F002/S003\n¿Seguro?"), "aprobacion")
	if _, hay := fichaEnDisco(t, raiz, ficha)["aprobacion"]; !hay {
		t.Fatal("la primera línea «apruebo E001/F002/S003» sí aprueba, aunque otra línea sea una pregunta")
	}
}

// Rojo cuando: un mensaje con UNA sola marca de agente, o un «apruebo» que no está en la primera línea, registra la aprobación.
// Verde cuando: cada marca de agente sola y el «apruebo» en segunda línea no registran; en la primera línea limpia sí.
func TestHookIgnoraMensajeDeAgente(t *testing.T) {
	raiz := repoContrato(t)
	ficha := "trabajo/activo/E001-F002-S003.json"
	for _, p := range []string{
		"apruebo E001/F002/S003\n<agent-message from=\"worker\">listo</agent-message>",
		"apruebo E001/F002/S003\n[Subagent hand-back] listo",
		"apruebo E001/F002/S003\n<task-notification>fin</task-notification>",
		"Lo leí.\napruebo E001/F002/S003",
	} {
		aquas(t, raiz, mensajeHumano(p), "aprobacion")
		if _, hay := fichaEnDisco(t, raiz, ficha)["aprobacion"]; hay {
			t.Fatalf("%q no es un mensaje del humano con apruebo en su primera línea: no se registra", p)
		}
	}
	aquas(t, raiz, mensajeHumano("apruebo E001/F002/S003\nLo leí."), "aprobacion")
	if _, hay := fichaEnDisco(t, raiz, ficha)["aprobacion"]; !hay {
		t.Fatal("«apruebo E001/F002/S003» en la primera línea sí aprueba")
	}
}

// aprobarCLI: deja la página del contrato y la aprobación humana del slice, como el flujo real antes de arrancar.
func aprobarCLI(t *testing.T, raiz, id string) {
	t.Helper()
	rel := "docs/" + strings.Join(strings.Split(id, "/"), "/") + ".md"
	if _, err := os.Stat(filepath.Join(raiz, filepath.FromSlash(rel))); err != nil {
		escribir(t, raiz, rel, "# "+id+"\n\n## Para qué existe\n\nContrato de prueba.\n")
	}
	if s := aquas(t, raiz, mensajeHumano("apruebo "+id), "aprobacion"); s.codigo != 0 {
		t.Fatalf("aprobacion %s: %+v", id, s)
	}
}
