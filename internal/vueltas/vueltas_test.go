package vueltas

import (
	"reflect"
	"strings"
	"testing"
)

// linea arma una línea del historial de Claude Code: una llamada Bash y su resultado.
func uso(id, cmd string) string {
	return `{"message":{"content":[{"type":"tool_use","id":"` + id + `","name":"Bash","input":{"command":"` + cmd + `"}}]}}`
}

func res(id, salida string, err bool) string {
	e := "false"
	if err {
		e = "true"
	}
	return `{"message":{"content":[{"type":"tool_result","tool_use_id":"` + id + `","is_error":` + e + `,"content":"` + salida + `"}]}}`
}

func TestCuentaVueltasPorMotivo(t *testing.T) {
	historial := strings.Join([]string{
		uso("a", "go test ./..."), res("a", "ok  \\tpaquete", false),
		uso("b", "go test ./internal/x"), res("b", "--- FAIL: TestA (0.00s)\\nFAIL", true),
		uso("c", "go test ./internal/y"), res("c", "FAIL\\tpaquete [build failed]", true),
		uso("h", "go test ./internal/w"), res("h", "--- FAIL: TestA/sub (0.00s)\\n--- FAIL: TestA (0.00s)", true),
		uso("d", "go test ./internal/z"), res("d", "panic sin señal", true),
		uso("e", ".tmp/aq.exe verificar"), res("e", "GATE FALLIDO · wiki docs/ incompleta:\\n- a\\nGATE FALLIDO · backlog trabajo/ (fichas):\\n- b", true),
		uso("f", "go run ./cmd/aquas verificar"), res("f", "GATE OK", false),
		uso("g", "grep -n verificar main.go"), res("g", "GATE FALLIDO", false),
		`línea que no es JSON`,
	}, "\n")

	got, err := Contar(strings.NewReader(historial))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]Conteo{
		"go test":   {Corridas: 5, Fallas: 4, Motivos: map[string]int{"TestA": 2, "compilación": 1, "otro": 1}},
		"verificar": {Corridas: 2, Fallas: 1, Motivos: map[string]int{"wiki": 1, "backlog": 1}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("conteo\n got %+v\nwant %+v", got, want)
	}
	seis := map[string]Conteo{"verificar": {Corridas: 9, Fallas: 9, Motivos: map[string]int{"a": 1, "b": 1, "c": 1, "d": 2, "e": 1, "wiki": 3}}}
	if l, w := Linea("X", seis), "X · verificar 9/9 (wiki 3, d 2, a 1, b 1, c 1)"; l != w {
		t.Fatalf("línea\n got %s\nwant %s", l, w)
	}
}

func TestMotivoDeWikiPorRegla(t *testing.T) {
	salida := "GATE FALLIDO · wiki docs/ incompleta:\n" +
		"  - docs/a.md: falta sección «Contrato y evidencia»\n" +
		"  - docs/b.md: falta sección «Paso a paso»\n" +
		"  - docs/b.md: primera aparición de «agente» sin enlace al glosario (escribe [agente](../glosario.md#agente))\n" +
		"GATE FALLIDO · derivados desfasados:\n  - CLAUDE.md"
	historial := uso("a", ".tmp/aq.exe verificar") + "\n" + res("a", strings.ReplaceAll(salida, "\n", `\n`), true) // escapado como en el JSON real
	got, err := Contar(strings.NewReader(historial))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]int{
		"wiki: falta sección «…»":                               1, // misma regla en dos páginas = una vuelta
		"wiki: primera aparición de «…» sin enlace al glosario": 1,
		"derivados": 1,
	}
	if !reflect.DeepEqual(got["verificar"].Motivos, want) {
		t.Fatalf("motivos\n got %v\nwant %v", got["verificar"].Motivos, want)
	}
}
