package guardia

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestEscribirExigeLeerContrato (E001/F017/S008): con un slice EN_CURSO, escribir dentro de su alcance exige que el
// registro de Claude Code muestre una lectura de su página desde el último /compact.
// Rojo cuando: se escribe sin leer la página desde el último compact.
// Verde cuando: frena sin lectura y pasa con Read, cat o biblioteca.
func TestEscribirExigeLeerContrato(t *testing.T) {
	raiz := raizPrueba(t)
	pagina := "docs/E001/F002/S007.md"
	abs := filepath.Join(raiz, filepath.FromSlash(pagina))
	slice := func(estado string) {
		t.Helper()
		dir := filepath.Join(raiz, "trabajo", "activo")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		f := `{"id": "E001/F002/S007", "tipo": "slice", "titulo": "A", "estado": "` + estado + `", "prioridad": "P1", "motivo": "m",
  "padre": "E001/F002", "bloqueado_por": [], "secuencia": 1, "alcance": ["internal/a/**", "docs/**"]}`
		if err := os.WriteFile(filepath.Join(dir, "E001-F002-S007.json"), []byte(f), 0o644); err != nil {
			t.Fatal(err)
		}
		p := `{"id": "E001/F002", "tipo": "feature", "titulo": "F", "estado": "EN_CURSO", "prioridad": "P1", "motivo": "m",
  "padre": "E001", "bloqueado_por": [], "secuencia": 1}`
		if err := os.WriteFile(filepath.Join(dir, "E001-F002.json"), []byte(p), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	uso := func(nombre string, input map[string]string) string {
		b, _ := json.Marshal(map[string]any{"type": "assistant", "message": map[string]any{"content": []any{
			map[string]any{"type": "tool_use", "name": nombre, "input": input}}}})
		return string(b)
	}
	leer := uso("Read", map[string]string{"file_path": abs})
	compact := `{"type":"system","subtype":"compact_boundary","content":"Conversation compacted"}`
	otro := uso("Read", map[string]string{"file_path": filepath.Join(raiz, "docs", "E001", "F002", "S001.md")})
	registros := filepath.Join(raiz, ".tmp", "registros")
	n := 0
	registro := func(lineas ...string) string {
		n++
		r := filepath.Join(registros, "s"+string(rune('a'+n))+".jsonl")
		if err := os.MkdirAll(filepath.Dir(r), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(r, []byte(strings.Join(lineas, "\n")+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		return r
	}
	entrada := func(tool, x, reg, agente string) []byte {
		in := map[string]string{"file_path": x}
		if tool == "Bash" {
			in = map[string]string{"command": x}
		}
		b, _ := json.Marshal(map[string]any{"tool_name": tool, "tool_input": in, "transcript_path": reg, "agent_id": agente, "cwd": raiz})
		return b
	}
	// Subagente: su registro vive en <sesión>/subagents/agent-<id>.jsonl, junto al de la sesión principal.
	principal := registro(leer)
	subagente := func(id, linea string) {
		p := filepath.Join(strings.TrimSuffix(principal, ".jsonl"), "subagents", "agent-"+id+".jsonl")
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(linea+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	subagente("x1", otro)
	subagente("x2", leer)

	slice("EN_CURSO")
	casos := []struct {
		caso            string
		tool, x, reg, a string
		dice            string // "" = pasa
	}{
		{"sin lectura frena y nombra la página", "Write", "internal/a/x.go", registro(otro), "", pagina},
		{"Read de la página pasa", "Edit", "internal/a/x.go", registro(leer), "", ""},
		{"lectura antes del compact no vale", "Write", "internal/a/x.go", registro(leer, compact, otro), "", pagina},
		{"cat después del compact pasa", "Write", "internal/a/x.go", registro(leer, compact, uso("Bash", map[string]string{"command": "cat " + pagina})), "", ""},
		{"sed de la página pasa", "Write", "internal/a/x.go", registro(uso("Bash", map[string]string{"command": "sed -n 1,40p \"" + pagina + "\""})), "", ""},
		{"head de la página pasa", "Write", "internal/a/x.go", registro(uso("Bash", map[string]string{"command": "head -50 " + pagina})), "", ""},
		{"biblioteca con la página pasa", "Write", "internal/a/x.go", registro(uso("Bash", map[string]string{"command": "./.tmp/aquas.exe library read " + pagina + " internal/a/x.go"})), "", ""},
		{"Bash que solo nombra la página no vale", "Write", "internal/a/x.go", registro(uso("Bash", map[string]string{"command": "git add " + pagina})), "", pagina},
		{"cat de otro archivo y la página en otro tramo no vale", "Write", "internal/a/x.go", registro(uso("Bash", map[string]string{"command": "cat a.txt && git add " + pagina})), "", pagina},
		{"sed -i sobre la página no es lectura", "Write", "internal/a/x.go", registro(uso("Bash", map[string]string{"command": "sed -i 's/a/b/' " + pagina})), "", pagina},
		{"lectura en un tramo tras cd pasa", "Write", "internal/a/x.go", registro(uso("Bash", map[string]string{"command": "cd x && cat " + pagina})), "", ""},
		{"Bash que escribe en el alcance sin lectura frena", "Bash", "echo x > internal/a/x.go", registro(otro), "", pagina},
		{"escribir la propia página no exige", "Write", pagina, registro(otro), "", ""},
		{"escribir en .tmp no exige", "Write", ".tmp/x.txt", registro(otro), "", ""},
		{"aquas backlog no exige", "Bash", "./.tmp/aquas.exe backlog state E001/F002/S007 CERRADO", registro(otro), "", ""},
		{"fuera del alcance frena por alcance", "Write", "cmd/z.go", registro(leer), "", "fuera del alcance"},
		{"registro ausente frena", "Write", "internal/a/x.go", filepath.Join(registros, "no-hay.jsonl"), "", "registro"},
		{"registro vacío en el hook frena", "Write", "internal/a/x.go", "", "", "registro"},
		{"registro ilegible frena", "Write", "internal/a/x.go", registros, "", "registro"},
		{"subagente sin lectura propia frena aunque la sesión leyó", "Write", "internal/a/x.go", principal, "x1", pagina},
		{"subagente que leyó pasa", "Write", "internal/a/x.go", principal, "x2", ""},
	}
	for _, c := range casos {
		m := Revisar(raiz, entrada(c.tool, c.x, c.reg, c.a))
		if c.dice == "" && m != "" || c.dice != "" && !strings.Contains(m, c.dice) {
			t.Errorf("%s: motivo %q, quiere %q", c.caso, m, c.dice)
		}
	}
	// Sin slice EN_CURSO no se exige: BORRADOR (feature en curso sin slice en curso), CERRADO aún en activo y
	// slice archivado en trabajo/archivo (hallazgo A1.1).
	sinLectura := registro(otro)
	for _, estado := range []string{"BORRADOR", "CERRADO", "archivado"} {
		if estado == "archivado" {
			arch := filepath.Join(raiz, "trabajo", "archivo")
			if err := os.MkdirAll(arch, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.Rename(filepath.Join(raiz, "trabajo", "activo", "E001-F002-S007.json"), filepath.Join(arch, "E001-F002-S007.json")); err != nil {
				t.Fatal(err)
			}
		} else {
			slice(estado)
		}
		if m := Revisar(raiz, entrada("Write", "internal/a/x.go", sinLectura, "")); m != "" {
			t.Errorf("slice %s: no exige lectura; motivo %q", estado, m)
		}
	}
}
