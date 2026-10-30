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
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
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
	// S007: completos = lo que da leer entero, una vez, cada ruta distinta del pedido.
	if _, entero, _ := Leer(raiz, Pedido{Rutas: []string{"vacio.txt", "docs", "a.go"}}); completos != entero {
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
	if _, entero, _ := Leer(raiz, Pedido{Rutas: []string{"largo.txt"}}); completos != entero {
		t.Errorf("completos = lo que devuelve leer el archivo entero: %d, se esperaba %d", completos, entero)
	}
	if devueltos == 0 || devueltos >= completos {
		t.Errorf("devueltos debe ser menor que completos: %d de %d", devueltos, completos)
	}
	if !strings.HasSuffix(txt, fmt.Sprintf("== devueltos %d de %d caracteres\n", devueltos, completos)) {
		t.Errorf("la última línea dice la cuenta:\n%s", txt)
	}
}

// TestCuentaDevueltos (E000/F006/S007, Fable #10 «devueltos 60081 de 56670»): devueltos y completos
// se miden en el mismo formato, cabecera y números de línea incluidos.
// Rojo cuando: el total devuelto supera lo pedido. Verde cuando: el total coincide con lo devuelto.
func TestCuentaDevueltos(t *testing.T) {
	raiz := repo(t)
	cuerpo := func(txt string) int { // lo devuelto sin la línea final de la cuenta
		return utf8.RuneCountInString(txt[:strings.LastIndex(txt, "== devueltos")])
	}
	casos := []struct {
		nombre string
		rutas  []string
	}{
		{"un archivo entero", []string{"largo.txt"}},
		{"varios enteros, con CRLF", []string{"a.go", "docs/b.md", "largo.txt"}},
		{"vacío", []string{"vacio.txt"}},
		{"solo errores", []string{"no-existe.go", "docs", "../fuera.go", "a.go:5-2"}},
		{"entero y error", []string{"a.go", "no-existe.go"}},
	}
	for _, c := range casos {
		txt, devueltos, completos := Leer(raiz, Pedido{Rutas: c.rutas})
		if devueltos != completos {
			t.Errorf("%s: lectura entera debe dar devueltos = completos: %d de %d", c.nombre, devueltos, completos)
		}
		if n := cuerpo(txt); devueltos != n {
			t.Errorf("%s: devueltos = %d pero el texto trae %d caracteres", c.nombre, devueltos, n)
		}
	}
	// Rango y archivo repetido: el archivo cuenta una vez en completos; devueltos ≤ completos.
	txt, devueltos, completos := Leer(raiz, Pedido{Rutas: []string{"largo.txt:1-5", "./largo.txt:10-12", "a.go"}})
	if devueltos >= completos || devueltos != cuerpo(txt) {
		t.Errorf("rangos: devueltos %d de %d; el texto trae %d", devueltos, completos, cuerpo(txt))
	}
	// Hallazgo A1.4: un error repetido cuenta una vez en completos, como un archivo repetido.
	if _, d, c := Leer(raiz, Pedido{Rutas: []string{"no.go", "no.go"}}); c*2 != d {
		t.Errorf("error repetido: devueltos %d, completos %d; esperaba completos = mitad", d, c)
	}
	// Hallazgos A1.1 y A2.1: rangos chicos disjuntos, repetidos o solapados, o un error repetido,
	// devuelven más que la lectura entera; la nota lo explica sin culpar a una sola causa.
	for _, rutas := range [][]string{{"a.go:1-1", "a.go:3-3"}, {"a.go", "a.go"}, {"largo.txt:1-200", "largo.txt:100-300"}, {"no.go", "no.go"}} {
		txt, d, c := Leer(raiz, Pedido{Rutas: rutas})
		if d <= c || d != cuerpo(txt) || !strings.HasSuffix(txt, "(más que leer entero de una vez: cada rango o aviso suma su cabecera y lo repetido cuenta doble)\n") {
			t.Errorf("%v: devueltos %d de %d sin explicar el exceso:\n%s", rutas, d, c, txt[strings.LastIndex(txt, "== devueltos"):])
		}
	}
	// Hallazgo A1.2: rango que empieza después del final; su aviso cuenta solo en devueltos.
	txt, d, c := Leer(raiz, Pedido{Rutas: []string{"a.go:9-9"}})
	if _, entero, _ := Leer(raiz, Pedido{Rutas: []string{"a.go"}}); c != entero || d != cuerpo(txt) || !strings.Contains(txt, "error: el archivo tiene 3 líneas") {
		t.Errorf("rango fuera del archivo: devueltos %d, completos %d (entero %d):\n%s", d, c, entero, txt)
	}
	// Hallazgo A1.3: con corte por tope, devueltos = tope y completos sigue siendo el archivo entero.
	txt, d, c = Leer(raiz, Pedido{Rutas: []string{"largo.txt"}, Tope: 500})
	if _, entero, _ := Leer(raiz, Pedido{Rutas: []string{"largo.txt"}}); d != 500 || c != entero || !strings.HasSuffix(txt, fmt.Sprintf("== devueltos 500 de %d caracteres\n", entero)) {
		t.Errorf("tope: devueltos %d de %d (entero %d)", d, c, entero)
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
