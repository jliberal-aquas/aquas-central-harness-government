package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestComandosEnIngles (E000/F006/S006, humano 2026-10-30: 1a 2a 3a): comandos y subcomandos en
// inglés; los nombres viejos siguen andando y avisan el nuevo; estados, campos y mensajes no cambian.
// Rojo cuando: un comando en inglés no responde o un nombre viejo no avisa.
// Verde cuando: el inglés responde igual que el viejo y el viejo avisa el nombre nuevo.
func TestComandosEnIngles(t *testing.T) {
	// De punta a punta: mismo resultado, y solo el nombre viejo avisa.
	raiz := t.TempDir()
	escribir(t, raiz, "a.txt", "uno\ndos\n")
	pedido := `{"rutas":["a.txt"]}`
	nuevo := aquas(t, raiz, pedido, "library", "read")
	viejo := aquas(t, raiz, pedido, "biblioteca", "leer")
	if nuevo.codigo != 0 || nuevo.out == "" || nuevo.out != viejo.out || viejo.codigo != 0 {
		t.Fatalf("library read debe responder igual que biblioteca leer:\nnuevo %d %q %q\nviejo %d %q", nuevo.codigo, nuevo.out, nuevo.err, viejo.codigo, viejo.out)
	}
	if strings.Contains(nuevo.err, "ahora se escribe") {
		t.Errorf("el nombre nuevo no avisa: %q", nuevo.err)
	}
	for _, w := range []string{"«biblioteca» ahora se escribe «library»", "«leer» ahora se escribe «read»"} {
		if !strings.Contains(viejo.err, w) {
			t.Errorf("falta el aviso %q en %q", w, viejo.err)
		}
	}

	// Cada posición: comando, subcomando y la palabra «ampliar» de backlog alcance.
	casos := []struct{ entra, sale, aviso string }{
		{"verify", "verificar", ""},
		{"compile", "compilar", ""},
		{"backlog next", "backlog siguiente", ""},
		{"backlog state E000/F006/S006 EN_CURSO", "backlog estado E000/F006/S006 EN_CURSO", ""},
		{"backlog scope E000/F006/S006 extend a.go motivo=x", "backlog alcance E000/F006/S006 ampliar a.go motivo=x", ""},
		{"memory set tarea=t", "memoria fijar tarea=t", ""},
		{"memory show", "memoria ver", ""}, // hallazgo A1.1
		{"memoria ver", "memoria ver", "«memoria» ahora se escribe «memory»"},
		{"turn close", "turno cerrar", ""},
		{"budget tool", "presupuesto herramienta", ""},
		{"agent-close", "cierre-agente", ""},
		{"guard", "guardia", ""},
		{"hook", "hook", ""},
		{"backlog estado X EN_CURSO", "backlog estado X EN_CURSO", "«estado» ahora se escribe «state»"},
		{"memoria cierre", "memoria cierre", "«memoria» ahora se escribe «memory»"},
		{"backlog alcance X ampliar a.go motivo=m", "backlog alcance X ampliar a.go motivo=m", "«ampliar» ahora se escribe «extend»"},
		{"backlog nuevo titulo=ver", "backlog nuevo titulo=ver", "«nuevo» ahora se escribe «new»"}, // solo la posición del subcomando se traduce
		{"build ver", "build ver", ""},         // build no tiene subcomandos: su argumento es una ruta
		{"desconocido x", "desconocido x", ""}, // el error de comando desconocido sigue saliendo en main
	}
	for _, c := range casos {
		args, avisos := traducir(strings.Fields(c.entra))
		if got := strings.Join(args, " "); got != c.sale {
			t.Errorf("%q → %q; esperado %q", c.entra, got, c.sale)
		}
		if got := strings.Join(avisos, "\n"); (c.aviso == "") != (got == "") || !strings.Contains(got, c.aviso) {
			t.Errorf("%q: avisos %q; esperado %q", c.entra, got, c.aviso)
		}
	}

	// En cada tabla (hallazgos A1.4 y A1.5): ningún nombre nuevo es a la vez nombre viejo de otro,
	// y cada nombre interno tiene un solo nombre nuevo, para que el aviso sea siempre el mismo.
	tablas := map[string]map[string]string{"comandos": comandos, "alcance": alcanceAmpliar}
	for c, t := range subcomandos {
		tablas[c] = t
	}
	for nombre, tabla := range tablas {
		internos := map[string]string{}
		for nuevoN, viejoN := range tabla {
			if otro, ok := tabla[viejoN]; ok && viejoN != nuevoN && otro != viejoN {
				t.Errorf("%s: %q es nombre viejo de %q y nombre nuevo de %q", nombre, viejoN, nuevoN, otro)
			}
			if previo, ok := internos[viejoN]; ok {
				t.Errorf("%s: %q tiene dos nombres nuevos: %q y %q", nombre, viejoN, previo, nuevoN)
			}
			internos[viejoN] = nuevoN
		}
	}

	// Los hooks y el pre-commit ya usan los nombres nuevos (decisión 2a).
	datos, err := os.ReadFile(filepath.Join("..", "..", "gobierno", "config", "claude_settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cfg struct {
		Hooks []struct {
			Orden string `json:"orden"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(datos, &cfg); err != nil || len(cfg.Hooks) == 0 {
		t.Fatalf("claude_settings.json sin hooks legibles: %v", err)
	}
	for _, h := range cfg.Hooks {
		if _, avisos := traducir(strings.Fields(h.Orden)); len(avisos) > 0 {
			t.Errorf("hook con nombre viejo %q: %v", h.Orden, avisos)
		}
	}
	pc, err := os.ReadFile(filepath.Join("..", "..", ".githooks", "pre-commit"))
	if err != nil || !strings.Contains(string(pc), ".exe verify") {
		t.Errorf("pre-commit debe correr «verify»: %v\n%s", err, pc)
	}
}
