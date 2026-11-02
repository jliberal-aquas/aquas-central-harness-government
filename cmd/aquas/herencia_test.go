package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// filaHerencia: una fila de gobierno/config/herencia.json (E003/F001/S002).
type filaHerencia struct {
	Control string `json:"control"`
	Destino string `json:"destino"`
	Motivo  string `json:"motivo"`
}

// TestHerenciaDePuntaAPunta (integral E003/F001): un hijo instalado tiene activo cada control que la central marca «todos».
// Rojo cuando: un control marcado todos no queda activo en el hijo.
// Verde cuando: ganchos, agentes, pruebas sin Go, verify y pre-commit funcionan en el hijo.
func TestHerenciaDePuntaAPunta(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init", "-q")
	leer := func(rel string) string {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatalf("el hijo debió tener %s: %v", rel, err)
		}
		return string(b)
	}
	if s := aquas(t, dir, "", "install"); s.codigo != 0 {
		t.Fatalf("install: %+v", s)
	}

	// S001: la central que viaja no lleva archivos locales de la máquina que compiló.
	if strings.Contains(leer(".aquas/lock.json"), ".local.json") {
		t.Error("S001: el lock del hijo lista un archivo local de la central")
	}

	// S002 + S003: cada gancho marcado «todos» queda en el settings del hijo; los agentes llegan.
	b, err := os.ReadFile(filepath.Join("..", "..", "gobierno", "config", "herencia.json"))
	if err != nil {
		t.Fatalf("S002: la central debió tener gobierno/config/herencia.json: %v", err)
	}
	var h struct {
		Controles []filaHerencia `json:"controles"`
	}
	if err := json.Unmarshal(b, &h); err != nil {
		t.Fatalf("S002: herencia.json ilegible: %v", err)
	}
	settings := leer(".claude/settings.json")
	for _, f := range h.Controles {
		if orden, ok := strings.CutPrefix(f.Control, "gancho:"); ok && f.Destino == "todos" && !strings.Contains(settings, "aquas "+orden) {
			t.Errorf("S003: el gancho %q viaja a todos y no está en el settings del hijo", orden)
		}
	}
	for _, a := range []string{"worker-sonnet", "worker-haiku", "auditor"} {
		if _, err := os.Stat(filepath.Join(dir, ".claude", "agents", a+".md")); err != nil {
			t.Errorf("S003: el agente %s debió llegar al hijo: %v", a, err)
		}
	}
	// S003 (D000090): sin git, install no falla; avisa que el entorno quedó sin declarar.
	sinGit := t.TempDir()
	if s := aquas(t, sinGit, "", "install"); s.codigo != 0 || !strings.Contains(s.out, "AVISO: entorno sin declarar") {
		t.Errorf("S003: install sin git debe salir 0 y avisar del entorno; dio %+v", s)
	}
	// S003: la guardia funciona en el hijo (exige el entorno que install declara).
	if s := aquas(t, dir, `{"tool_name":"Write","tool_input":{"file_path":"../fuera.txt"}}`, "guardia"); s.codigo != 2 || !strings.Contains(s.err, "bloqueada: ../fuera.txt") {
		t.Errorf("S003: en el hijo, escribir fuera debe bloquearse (exit 2); dio %+v", s)
	}
	if s := aquas(t, dir, `{"tool_name":"Write","tool_input":{"file_path":"docs/a.md"}}`, "guardia"); s.codigo != 0 {
		t.Errorf("S003: en el hijo, escribir dentro debe pasar; dio %+v", s)
	}
	for _, f := range h.Controles {
		if orden, ok := strings.CutPrefix(f.Control, "gancho:"); ok && f.Destino == "todos" {
			// Cada gancho corre en el hijo: 0 deja pasar, 2 frena con motivo; 1 sería un fallo interno (p. ej. falta un archivo).
			if s := aquas(t, dir, `{"session_id":"s1","prompt":"hola","tool_name":"Read","tool_input":{}}`, strings.Fields(orden)...); s.codigo != 0 && s.codigo != 2 {
				t.Errorf("S003: el gancho %q falla en el hijo: %+v", orden, s)
			}
		}
	}

	// S004: rojo y verde corren la plantilla de pruebas.json, sin Go.
	escribir(t, dir, "gobierno/config/pruebas.json", `{"version": 1, "comando": "sh {ruta} {nombre}"}`)
	escribir(t, dir, "p.sh", "test -f \"$1\" || { echo \"falta $1\"; exit 1; }\n")
	prueba := "prueba=p.sh:arreglado=pide el arreglo|r|v"
	for _, args := range [][]string{
		{"nuevo", "id=E001", "titulo=G", "prioridad=P1", "motivo=m", "relacion=nueva", "que_hace=hace algo", prueba},
		{"nuevo", "id=E001/F001", "titulo=B", "prioridad=P1", "motivo=m", "relacion=nueva", "que_hace=hace algo", prueba},
		{"nuevo", "id=E001/F001/S001", "titulo=R", "prioridad=P1", "motivo=m", "relacion=nueva", "que_hace=hace algo", prueba},
	} {
		if s := aquas(t, dir, "", append([]string{"backlog"}, args...)...); s.codigo != 0 {
			t.Fatalf("S004: backlog %v: %+v", args, s)
		}
		rojoIntegral(t, dir, args)
	}
	aprobarCLI(t, dir, "E001/F001/S001")
	if s := aquas(t, dir, "", "backlog", "estado", "E001/F001/S001", "EN_CURSO"); s.codigo != 0 {
		t.Fatalf("S004: EN_CURSO: %+v", s)
	}
	if s := aquas(t, dir, "", "backlog", "rojo", "E001/F001/S001"); s.codigo != 0 || !strings.Contains(s.out, "falta arreglado") {
		t.Errorf("S004: el rojo debió correr la plantilla y guardar su salida; dio %+v", s)
	}
	escribir(t, dir, "arreglado", "")
	if s := aquas(t, dir, "", "backlog", "verde", "E001/F001/S001"); s.codigo != 0 {
		t.Errorf("S004: el verde debió pasar con la plantilla; dio %+v", s)
	}

	// S005: el pre-commit del hijo corre verify, y verify revisa las fichas del hijo.
	if !strings.Contains(leer(".githooks/pre-commit"), "aquas verify") {
		t.Error("S005: el pre-commit del hijo debió correr aquas verify")
	}
	if out, err := exec.Command("git", "-C", dir, "config", "core.hooksPath").Output(); err != nil || strings.TrimSpace(string(out)) != ".githooks" {
		t.Errorf("S005: core.hooksPath del hijo debió ser .githooks; dio %q %v", out, err)
	}
	escribir(t, dir, "trabajo/activo/E001-F001-S001.json", "{roto")
	if s := aquas(t, dir, "", "verify"); s.codigo == 0 || !strings.Contains(s.out+s.err, "backlog") {
		t.Errorf("S005: verify del hijo debió fallar por la ficha rota; dio %+v", s)
	}
}
