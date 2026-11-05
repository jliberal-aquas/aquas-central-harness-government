package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestHijoSinGoCicloCompleto (integral E000/F015): un hijo sin Go recorre contrato, casos borde, aprobación, rojo,
// verde y gate sin chocar con supuestos de Go.
// Rojo cuando: un control exige algo propio de Go a un hijo con pruebas.json.
// Verde cuando: el ciclo completo pasa con nombres ruta:nombre y un nombre inventado se sigue rechazando.
func TestHijoSinGoCicloCompleto(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init", "-q")
	if s := aquas(t, dir, "", "install"); s.codigo != 0 {
		t.Fatalf("install: %+v", s)
	}
	escribir(t, dir, "gobierno/config/pruebas.json", `{"version": 1, "comando": "sh {ruta} {nombre}"}`)
	escribir(t, dir, "p.sh", "test -f \"$1\" || { echo \"falta $1\"; exit 1; }\n")

	// S002: pedir ayuda no ejecuta la orden.
	if s := aquas(t, dir, "", "backlog", "new", "--help"); s.codigo != 0 || !strings.Contains(s.out+s.err, "uso") {
		t.Errorf("S002: backlog new --help debió mostrar el uso y salir 0; dio %+v", s)
	}
	if _, err := os.Stat(filepath.Join(dir, "trabajo", "activo")); err == nil {
		if xs, _ := os.ReadDir(filepath.Join(dir, "trabajo", "activo")); len(xs) > 0 {
			t.Errorf("S002: --help escribió fichas: %v", xs)
		}
	}

	prueba := "prueba=p.sh:arreglado=pide el arreglo|r|v"
	for _, args := range [][]string{
		{"nuevo", "id=E001", "titulo=G", "prioridad=P1", "motivo=m", "relacion=nueva", "que_hace=hace algo", prueba},
		{"nuevo", "id=E001/F001", "titulo=B", "prioridad=P1", "motivo=m", "relacion=nueva", "que_hace=hace algo", prueba},
		{"nuevo", "id=E001/F001/S001", "titulo=R", "prioridad=P1", "motivo=m", "relacion=nueva", "que_hace=hace algo", prueba},
	} {
		if s := aquas(t, dir, "", append([]string{"backlog"}, args...)...); s.codigo != 0 {
			t.Fatalf("backlog %v: %+v", args, s)
		}
	}
	if s := aquas(t, dir, "", "backlog", "rojo", "E001/F001"); s.codigo != 0 {
		t.Fatalf("rojo de la integral: %+v", s)
	}
	escribir(t, dir, "docs/README.md", "# Inicio\n\n- [E001](E001/README.md)\n")
	escribir(t, dir, "docs/E001/README.md", "# E001\n\n- [F001](F001/README.md)\n")
	escribir(t, dir, "docs/E001/F001/README.md", "# F001\n\n- [S001](S001.md)\n")

	// S001: la tabla de casos borde cita la prueba con el nombre ruta:nombre de la ficha.
	base, err := os.ReadFile(filepath.Join(dir, "docs", "E001", "F001", "S001.md"))
	if err != nil {
		t.Fatalf("backlog new debió crear la página del slice: %v", err)
	}
	pagina := func(nombre string) string {
		var b strings.Builder
		for _, l := range strings.Split(string(base), "\n") {
			if c, _, ok := strings.Cut(strings.TrimPrefix(l, "| "), " | no | PENDIENTE"); ok {
				l = "| " + c + " | sí | `" + nombre + "` |"
			}
			b.WriteString(strings.ReplaceAll(l, "PENDIENTE: el agente lo decide antes del apruebo", "decidido") + "\n")
		}
		return b.String()
	}
	escribir(t, dir, "docs/E001/F001/S001.md", pagina("p.sh:arreglado"))
	aprobarCLI(t, dir, "E001/F001/S001")
	if s := aquas(t, dir, "", "backlog", "estado", "E001/F001/S001", "EN_CURSO"); s.codigo != 0 {
		t.Fatalf("EN_CURSO: %+v", s)
	}
	if s := aquas(t, dir, "", "backlog", "rojo", "E001/F001/S001"); s.codigo != 0 {
		t.Fatalf("rojo: %+v", s)
	}
	escribir(t, dir, "arreglado", "")
	if s := aquas(t, dir, "", "backlog", "verde", "E001/F001/S001"); s.codigo != 0 {
		t.Fatalf("verde: %+v", s)
	}
	s := aquas(t, dir, "", "verify")
	for _, mal := range []string{"con sí nombra una prueba", "prueba nombrada no existe"} {
		if strings.Contains(s.out+s.err, mal) {
			t.Errorf("S001: el gate del hijo rechaza la prueba ruta:nombre de la ficha («%s»): %s", mal, s.out+s.err)
		}
	}

	// S001: un nombre que la ficha no declara se sigue rechazando.
	escribir(t, dir, "docs/E001/F001/S001.md", pagina("p.sh:inventada"))
	if s := aquas(t, dir, "", "verify"); !strings.Contains(s.out+s.err, "prueba nombrada no existe en el código: p.sh:inventada") {
		t.Errorf("S001: el gate debió rechazar p.sh:inventada; dio %+v", s)
	}
}
