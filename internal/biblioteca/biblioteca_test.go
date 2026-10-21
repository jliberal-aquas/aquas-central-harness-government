package biblioteca

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"unicode/utf8"
)

// E001/F003/S001: leer un lote de rutas en una llamada, con rangos, tope y cuenta.

func repo(t *testing.T) string {
	t.Helper()
	raiz := t.TempDir()
	var largo strings.Builder
	for i := 1; i <= 300; i++ {
		fmt.Fprintf(&largo, "línea %d del archivo largo\n", i)
	}
	for rel, c := range map[string]string{
		"a.go":      "package a\n\nfunc A() {}\n",
		"docs/b.md": "# B\r\nuno\r\ndos\r\n",
		"largo.txt": largo.String(),
		"vacio.txt": "",
	} {
		p := filepath.Join(raiz, filepath.FromSlash(rel))
		os.MkdirAll(filepath.Dir(p), 0o755)
		os.WriteFile(p, []byte(c), 0o644)
	}
	return raiz
}

func TestLeerLoteNumerado(t *testing.T) {
	raiz := repo(t)
	txt, _, _ := Leer(raiz, Pedido{Rutas: []string{"a.go", "docs/b.md"}})
	for _, w := range []string{
		"== a.go · 1-3 de 3\n1\tpackage a\n2\t\n3\tfunc A() {}\n",
		"== docs/b.md · 1-3 de 3\n1\t# B\n2\tuno\n3\tdos\n",
	} {
		if !strings.Contains(txt, w) {
			t.Errorf("falta el bloque:\n%s\nen:\n%s", w, txt)
		}
	}
	if strings.Contains(txt, "\r") {
		t.Error("CRLF debe salir como LF")
	}
}

func TestLeerRango(t *testing.T) {
	raiz := repo(t)
	txt, _, _ := Leer(raiz, Pedido{Rutas: []string{"largo.txt:10-12", "largo.txt:299-400"}})
	if !strings.Contains(txt, "== largo.txt · 10-12 de 300\n10\tlínea 10 del archivo largo\n11\t") || strings.Contains(txt, "13\tlínea 13") {
		t.Errorf("rango 10-12 exacto:\n%s", txt)
	}
	if !strings.Contains(txt, "== largo.txt · 299-300 de 300\n") {
		t.Errorf("un rango que pasa el final se corta en la última línea:\n%s", txt)
	}
}

func TestLeerErroresPorRutaNoFrenanElLote(t *testing.T) {
	raiz := repo(t)
	abs := filepath.Join(raiz, "a.go")
	txt, _, _ := Leer(raiz, Pedido{Rutas: []string{"no-existe.go", "../fuera.x", abs, "a.go:5-2", "a.go"}})
	// E000/F005/S003: cada ruta tiene su propia línea; la absoluta se comprueba con su nombre.
	if n := strings.Count(txt, "== ") - strings.Count(txt, "== devueltos "); n != 5 {
		t.Errorf("cinco rutas, cinco encabezados (sin contar el resumen); hubo %d:\n%s", n, txt)
	}
	for _, w := range []string{
		"== no-existe.go · error: no existe",
		"== ../fuera.x · error: fuera del repo",
		"== " + abs + " · error: fuera del repo", // ruta absoluta: solo relativas al repo
		"== a.go:5-2 · error: rango inválido",
		"== a.go · 1-3 de 3",
	} {
		if !strings.Contains(txt, w) {
			t.Errorf("falta %q en:\n%s", w, txt)
		}
	}
}

// archivoIlegible: ruta relativa que el SO no deja leer, con un error distinto de «no existe».
// Windows: «<» hace inválido el nombre. Otros SO: archivo sin permisos.
func archivoIlegible(t *testing.T, raiz string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		return "x<y.txt"
	}
	p := filepath.Join(raiz, "secreto.txt")
	if err := os.WriteFile(p, []byte("s"), 0o000); err != nil {
		t.Fatal(err)
	}
	if _, err := os.ReadFile(p); err == nil {
		t.Skip("este usuario lee archivos sin permiso (root): no hay archivo ilegible")
	}
	return "secreto.txt"
}

// Hallazgos del auditor (2026-10-21): archivo vacío, carpeta y misma ruta con otra grafía.
func TestLeerCasosBorde(t *testing.T) {
	raiz := repo(t)
	txt, _, completos := Leer(raiz, Pedido{Rutas: []string{"vacio.txt", "docs", "a.go:1-1", "./a.go:2-2", "a.go:3-3"}})
	for _, w := range []string{"== vacio.txt · vacío (0 líneas)\n", "== docs · error: es una carpeta\n"} {
		if !strings.Contains(txt, w) {
			t.Errorf("falta %q en:\n%s", w, txt)
		}
	}
	if strings.Contains(txt, raiz) {
		t.Errorf("los errores no muestran la ruta absoluta de la máquina:\n%s", txt)
	}
	// E000/F005/S003: un archivo que el SO no deja leer llega a «no se pudo leer», sin la ruta absoluta.
	ilegible := archivoIlegible(t, raiz)
	txt, _, _ = Leer(raiz, Pedido{Rutas: []string{ilegible}})
	if !strings.HasPrefix(txt, "== "+ilegible+" · error: no se pudo leer\n== devueltos ") || strings.Contains(txt, raiz) {
		t.Errorf("archivo ilegible: una línea con «no se pudo leer» y sin ruta absoluta; dio:\n%s", txt)
	}
	if completos != utf8.RuneCountInString("package a\n\nfunc A() {}\n") {
		t.Errorf("a.go pedido tres veces, con otra grafía, cuenta una sola vez en completos: %d", completos)
	}
}

func TestTopeCortaYLoDice(t *testing.T) {
	raiz := repo(t)
	txt, devueltos, _ := Leer(raiz, Pedido{Rutas: []string{"largo.txt", "a.go"}, Tope: 500})
	if devueltos > 500 || utf8.RuneCountInString(txt) > 700 {
		t.Fatalf("el tope limita lo devuelto: %d caracteres", devueltos)
	}
	if !strings.Contains(txt, "[cortado por tope: 500 caracteres") {
		t.Errorf("el corte se avisa, con el tope:\n%s", txt)
	}
}

func TestCuentaDeAhorro(t *testing.T) {
	raiz := repo(t)
	txt, devueltos, completos := Leer(raiz, Pedido{Rutas: []string{"largo.txt:1-5"}})
	b, _ := os.ReadFile(filepath.Join(raiz, "largo.txt"))
	if completos != utf8.RuneCount(b) {
		t.Errorf("completos = caracteres del archivo entero: %d, se esperaba %d", completos, utf8.RuneCount(b))
	}
	if devueltos == 0 || devueltos >= completos {
		t.Errorf("devueltos debe ser menor que completos: %d de %d", devueltos, completos)
	}
	if !strings.HasSuffix(txt, fmt.Sprintf("== devueltos %d de %d caracteres\n", devueltos, completos)) {
		t.Errorf("la última línea dice la cuenta:\n%s", txt)
	}
}

// Prueba de ahorro acordada con el humano (2026-10-21): la misma tarea por el bibliotecario
// y por lectura completa numerada (como la herramienta Read), comparando caracteres.
// Tarea: ver la cabecera (10 líneas) de tres archivos.
func TestAhorroFrenteALecturaCompleta(t *testing.T) {
	raiz := repo(t)
	txt, _, _ := Leer(raiz, Pedido{Rutas: []string{"largo.txt:1-10", "a.go:1-10", "docs/b.md:1-10"}})
	completa := 0
	for _, r := range []string{"largo.txt", "a.go", "docs/b.md"} {
		b, _ := os.ReadFile(filepath.Join(raiz, filepath.FromSlash(r)))
		for i, l := range strings.Split(strings.TrimRight(string(b), "\n"), "\n") {
			completa += utf8.RuneCountInString(fmt.Sprintf("%6d\t%s\n", i+1, l))
		}
	}
	got := utf8.RuneCountInString(txt)
	if got == 0 || got*4 > completa {
		t.Fatalf("el bibliotecario debe devolver menos de un cuarto: %d frente a %d caracteres", got, completa)
	}
	t.Logf("ahorro: %d frente a %d caracteres (%.0f%%)", got, completa, 100-float64(got)*100/float64(completa))
}
