package backlog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// lineaUsuario: una línea del registro de Claude Code con un mensaje de usuario (forma observada 2026-10-31).
func lineaUsuario(texto, origen string, contenido any) string {
	if contenido == nil {
		contenido = texto
	}
	l := map[string]any{"type": "user", "message": map[string]any{"role": "user", "content": contenido}}
	if origen != "" {
		l["origin"] = map[string]string{"kind": origen}
	}
	b, _ := json.Marshal(l)
	return string(b)
}

// registro: deja un .jsonl con esas líneas dentro de CLAUDE_CONFIG_DIR/projects y devuelve su ruta.
func registro(t *testing.T, lineas ...string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "claude")
	t.Setenv("CLAUDE_CONFIG_DIR", dir)
	ruta := filepath.Join(dir, "projects", "repo", "sesion.jsonl")
	if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ruta, []byte(strings.Join(lineas, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return ruta
}

// TestApruebaSoloConMensajeHumano (E000/F008/S002, D000077): una aprobación vale solo si su mensaje humano
// exacto está en el registro de la sesión que guardó el hook.
// Rojo cuando: una aprobación sin registro deja pasar a EN_CURSO.
// Verde cuando: solo pasa la que tiene el mensaje humano exacto.
func TestApruebaSoloConMensajeHumano(t *testing.T) {
	const id = "E001/F002/S002"
	const msg = "apruebo " + id
	falla := map[string]func(t *testing.T, raiz string) string{
		"sin registro":    func(t *testing.T, raiz string) string { return "" },
		"archivo ausente": func(t *testing.T, raiz string) string { return registro(t) + ".no" },
		"texto distinto": func(t *testing.T, raiz string) string {
			return registro(t, lineaUsuario("apruebo otra cosa", "human", nil))
		},
		"lo dijo el agente": func(t *testing.T, raiz string) string { return registro(t, lineaUsuario(msg, "", nil)) },
		"otra sesión":       func(t *testing.T, raiz string) string { return registro(t, lineaUsuario(msg, "peer", nil)) },
		"resultado de herramienta": func(t *testing.T, raiz string) string {
			return registro(t, lineaUsuario(msg, "human", []map[string]string{{"type": "tool_result", "content": msg}}))
		},
		"línea rota": func(t *testing.T, raiz string) string {
			return registro(t, `{"type":"user","message":{"content":"`+msg)
		},
		"fuera de projects": func(t *testing.T, raiz string) string {
			r := registro(t, lineaUsuario(msg, "human", nil))
			otra := filepath.Join(t.TempDir(), "sesion.jsonl")
			b, _ := os.ReadFile(r)
			_ = os.WriteFile(otra, b, 0o644)
			return otra
		},
		"dentro del repo": func(t *testing.T, raiz string) string {
			t.Setenv("CLAUDE_CONFIG_DIR", filepath.Join(raiz, ".tmp", "claude"))
			ruta := filepath.Join(raiz, ".tmp", "claude", "projects", "x", "s.jsonl")
			_ = os.MkdirAll(filepath.Dir(ruta), 0o755)
			_ = os.WriteFile(ruta, []byte(lineaUsuario(msg, "human", nil)+"\n"), 0o644)
			return ruta
		},
	}
	for nombre, reg := range falla {
		raiz := conExigencia(t)
		if _, err := Aprobar(raiz, id, msg, reg(t, raiz), hora); err != nil {
			t.Fatalf("%s: %v", nombre, err)
		}
		if _, err := Estado(raiz, id, "EN_CURSO"); err == nil || !strings.Contains(err.Error(), "registro") {
			t.Errorf("%s: EN_CURSO se rechaza nombrando el registro; dio %v", nombre, err)
		}
		f, _, _ := Ver(raiz, id) // verde y cierre también usan la aprobación (A1, hallazgo 4)
		f.Estado, f.Rojo, f.Verde = "EN_CURSO", &Registro{Fecha: "2026-10-01T00:00:00Z"}, &Registro{Fecha: "2026-10-01T00:00:00Z"}
		if err := escribir(raiz, Activo, f); err != nil {
			t.Fatal(err)
		}
		if _, err := Verde(raiz, id, hora); err == nil || !strings.Contains(err.Error(), "registro") {
			t.Errorf("%s: verde se rechaza nombrando el registro; dio %v", nombre, err)
		}
		if _, err := Estado(raiz, id, "CERRADO"); err == nil || !strings.Contains(err.Error(), "registro") {
			t.Errorf("%s: cierre se rechaza nombrando el registro; dio %v", nombre, err)
		}
		if _, err := Rojo(raiz, id, hora); err == nil || !strings.Contains(err.Error(), "registro") {
			t.Errorf("%s: rojo se rechaza nombrando el registro; dio %v", nombre, err)
		}
	}

	// Pasa: mensaje humano exacto, como texto o en bloques, entre líneas rotas y de otros.
	for nombre, lineas := range map[string][]string{
		"texto":   {`{roto`, lineaUsuario("hola", "human", nil), lineaUsuario(msg, "human", nil)},
		"bloques": {lineaUsuario(msg, "human", []map[string]string{{"type": "text", "text": msg}})},
	} {
		raiz := conExigencia(t)
		if _, err := Aprobar(raiz, id, msg, registro(t, lineas...), hora); err != nil {
			t.Fatal(err)
		}
		if _, err := Estado(raiz, id, "EN_CURSO"); err != nil {
			t.Errorf("%s: con respaldo humano EN_CURSO pasa; dio %v", nombre, err)
		}
		// Choque con la huella: contrato cambiado anula aunque haya respaldo.
		editarPagina(t, raiz, "lo que se va a construir", "otra cosa")
		if _, err := Rojo(raiz, id, hora); err == nil || !strings.Contains(err.Error(), "apruebo "+id) {
			t.Errorf("%s: contrato cambiado pide «apruebo» aunque haya respaldo; dio %v", nombre, err)
		}
	}

	// Sin aprobación y con rojo previo ya no queda exento: ese rojo también se finge a mano.
	raiz := conExigencia(t)
	f, _, _ := Ver(raiz, id)
	f.Rojo = &Registro{Fecha: "2026-10-01T00:00:00Z"}
	if err := escribir(raiz, Activo, f); err != nil {
		t.Fatal(err)
	}
	if _, err := Estado(raiz, id, "EN_CURSO"); err == nil || !strings.Contains(err.Error(), "apruebo "+id) {
		t.Errorf("rojo previo sin aprobación: pide «apruebo»; dio %v", err)
	}

	// Un mensaje humano real que no aprueba este id no respalda (A1, hallazgo 1).
	for _, m := range []string{"dale", "apruebo E000/F006/S010", "apruebo " + id + "?", "no apruebo " + id} {
		raiz := conExigencia(t)
		if _, err := Aprobar(raiz, id, m, respaldo(t, m), hora); err != nil {
			t.Fatal(err)
		}
		if _, err := Estado(raiz, id, "EN_CURSO"); err == nil || !strings.Contains(err.Error(), "apruebo "+id) {
			t.Errorf("mensaje %q: no aprueba %s y no respalda; dio %v", m, id, err)
		}
	}
	featureSinRespaldoNoCubre(t)
}

// featureSinRespaldoNoCubre (E000/F008/S002, A1 hallazgos 1 y 4): la aprobación de la feature cubre a sus
// slices solo con respaldo, y un push escrito en la ficha sin «con push» en el mensaje no vale.
func featureSinRespaldoNoCubre(t *testing.T) {
	t.Helper()
	const fe, sl = "E007/F001", "E007/F001/S001"
	raiz := modulo(t)
	for _, rel := range []string{"docs/E007/README.md", "docs/E007/F001/README.md", "docs/E007/F001/S001.md"} {
		ruta := filepath.Join(raiz, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(ruta, []byte("# x\n\n## Para qué existe\n\nContrato.\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	nuevo(t, raiz, "id=E007", "titulo=I", "prioridad=P1", "motivo=m")
	nuevo(t, raiz, "id="+fe, "titulo=F", "prioridad=P1", "motivo=m")
	nuevo(t, raiz, "id="+sl, "titulo=A", "prioridad=P1", "motivo=m", "prueba=./p:TestA=a|r|v")
	revisar := func() error {
		f, _, err := Ver(raiz, sl)
		if err != nil {
			t.Fatal(err)
		}
		return aprobada(raiz, f)
	}
	if _, err := Aprobar(raiz, fe, "apruebo "+fe, "", hora); err != nil {
		t.Fatal(err)
	}
	if err := revisar(); err == nil || !strings.Contains(err.Error(), "apruebo "+fe) {
		t.Errorf("feature aprobada sin registro: el slice pide «apruebo %s»; dio %v", fe, err)
	}
	p, err := Aprobar(raiz, fe, "apruebo "+fe, respaldo(t, "apruebo "+fe), hora)
	if err != nil {
		t.Fatal(err)
	}
	if err := revisar(); err != nil {
		t.Errorf("feature con respaldo cubre al slice; dio %v", err)
	}
	p.Aprobacion.Push = true // escrito a mano: el mensaje no dice «con push»
	if err := escribir(raiz, Activo, p); err != nil {
		t.Fatal(err)
	}
	if err := revisar(); err == nil || !strings.Contains(err.Error(), "con push") {
		t.Errorf("push sin «con push» en el mensaje no vale; dio %v", err)
	}
}

// respaldo: registro con el mensaje humano exacto; lo usan las pruebas que solo necesitan una aprobación válida.
func respaldo(t *testing.T, msg string) string {
	t.Helper()
	return registro(t, lineaUsuario(msg, "human", nil))
}
