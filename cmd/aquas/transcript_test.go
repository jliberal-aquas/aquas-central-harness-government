package main

import (
	"os"
	"path/filepath"
	"testing"
)

// probarUltimoMensaje (D000053): corre dentro de TestGanchoAnotaVeredicto. Rojo cuando el cierre no se
// lee del transcript; verde cuando gana el último SubagentHandback o texto del asistente.
func probarUltimoMensaje(t *testing.T) {
	dir := t.TempDir()
	escribir := func(nombre, s string) string {
		p := filepath.Join(dir, nombre)
		if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	usuario := `{"type":"user","message":{"role":"user","content":"encargo"}}` + "\n"
	texto := `{"type":"assistant","message":{"role":"assistant","content":[{"type":"thinking","thinking":""},{"type":"text","text":"borrador"}]}}` + "\n"
	entrega := `{"type":"assistant","message":{"role":"assistant","content":[{"type":"tool_use","name":"SubagentHandback","input":{"message":"id: —\nveredicto: APROBADO"}}]}}` + "\n"
	resultado := `{"type":"user","message":{"role":"user","content":[{"type":"tool_result","content":"ok"}]}}`
	if got := ultimoMensaje(escribir("handback.jsonl", usuario+texto+entrega+resultado)); got != "id: —\nveredicto: APROBADO" {
		t.Errorf("con SubagentHandback al final se esperaba su mensaje; dio %q", got)
	}
	if got := ultimoMensaje(escribir("texto.jsonl", usuario+entrega+texto)); got != "borrador" {
		t.Errorf("con texto al final se esperaba «borrador»; dio %q", got)
	}
	if got := ultimoMensaje(filepath.Join(dir, "no-existe.jsonl")); got != "" {
		t.Errorf("sin transcript se esperaba vacío; dio %q", got)
	}
	// H2: una línea de usuario con parte text después del cierre no lo pisa.
	usuarioTexto := `{"type":"user","message":{"role":"user","content":[{"type":"text","text":"eco del usuario"}]}}` + "\n"
	if got := ultimoMensaje(escribir("usuario-texto.jsonl", usuario+texto+usuarioTexto)); got != "borrador" {
		t.Errorf("texto de usuario no cuenta: se esperaba «borrador»; dio %q", got)
	}
	// H3: texto del asistente solo con espacios no pisa el cierre.
	blanco := `{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"  \n "}]}}` + "\n"
	if got := ultimoMensaje(escribir("blanco.jsonl", usuario+entrega+blanco)); got != "id: —\nveredicto: APROBADO" {
		t.Errorf("texto en blanco no pisa el cierre; dio %q", got)
	}
	// H4: una parte con forma rara no descarta las demás de la línea.
	rara := `{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":5},"suelto",{"type":"text","text":"bueno"}]}}` + "\n"
	if got := ultimoMensaje(escribir("rara.jsonl", usuario+rara)); got != "bueno" {
		t.Errorf("parte rara no descarta las demás: se esperaba «bueno»; dio %q", got)
	}
}
