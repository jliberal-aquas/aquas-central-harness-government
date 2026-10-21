package compilador

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// E001/F006/S001: workers y auditor. El compilador acepta `tools` por agente y la config
// real declara los tres agentes pedidos por el humano (2026-10-21).

func TestAgenteConHerramientas(t *testing.T) {
	raiz := repoMinimo(t)
	escribir(raiz, RutaAgentes, []byte(`{"agentes":[
		{"name":"auditor","description":"Revisa","model":"opus","effort":"high","tools":["Read","Grep","Glob","Bash"],
		 "cuerpo":"gobierno/canonico/agentes/orq.md","max_caracteres":10000},
		{"name":"worker","description":"Ejecuta","model":"haiku","cuerpo":"gobierno/canonico/agentes/orq.md","max_caracteres":10000}]}`))
	if _, err := Compilar(raiz); err != nil {
		t.Fatal(err)
	}
	if a := leer(t, raiz, ".claude/agents/auditor.md"); !strings.Contains(a, "effort: high\ntools: Read, Grep, Glob, Bash\n---\n") {
		t.Errorf("auditor debe declarar sus herramientas en el frontmatter:\n%s", a)
	}
	if w := leer(t, raiz, ".claude/agents/worker.md"); strings.Contains(w, "tools:") {
		t.Errorf("sin `tools` en config, el agente hereda todas: no debe escribir la línea:\n%s", w)
	}
	verificarSinProblemas(t, raiz)
}

func TestAgentesRealesWorkersYAuditor(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", filepath.FromSlash(RutaAgentes)))
	if err != nil {
		t.Fatal(err)
	}
	var c struct{ Agentes []Agente }
	if err := json.Unmarshal(b, &c); err != nil {
		t.Fatal(err)
	}
	por := map[string]Agente{}
	for _, a := range c.Agentes {
		por[a.Name] = a
	}
	for nombre, modelo := range map[string]string{"worker-sonnet": "sonnet", "worker-haiku": "haiku", "auditor": "opus"} {
		a, ok := por[nombre]
		if !ok {
			t.Errorf("falta el agente %s", nombre)
			continue
		}
		if a.Model != modelo {
			t.Errorf("%s: modelo %q, se esperaba %q", nombre, a.Model, modelo)
		}
		if len(a.Comunicacion) > 0 {
			t.Errorf("%s no habla con el humano: no carga reglas de comunicación", nombre)
		}
	}
	au := por["auditor"]
	if au.Effort != "high" {
		t.Errorf("auditor: effort %q, se esperaba high", au.Effort)
	}
	if len(au.Tools) == 0 || slices.ContainsFunc(au.Tools, func(h string) bool {
		return h == "Edit" || h == "Write" || h == "MultiEdit" || h == "NotebookEdit"
	}) {
		t.Errorf("auditor solo lee: herramientas declaradas y sin Edit/Write; tiene %v", au.Tools)
	}
}
