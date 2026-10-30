package wiki

import (
	"fmt"
	"math"
	"strings"
	"testing"
)

func TestSilabas(t *testing.T) {
	casos := map[string]int{
		"casa": 2, "día": 2, "cuidado": 3, "leer": 2, "muy": 1, "hay": 1, "aéreo": 4,
		"guion": 1, "país": 2, "huevo": 2, "ley": 1, "oír": 2, "poeta": 3, "aquí": 2,
		"guía": 2, "queso": 2, "pingüino": 3, "ayer": 2, "y": 1, "Uruguay": 3,
		"S001": 1, "murciélago": 4, "teatro": 3, "ciudad": 2, "baúl": 2, "Documentación": 5,
	}
	for palabra, esperado := range casos {
		if got := Silabas(palabra); got != esperado {
			t.Errorf("Silabas(%q) = %d; esperado %d", palabra, got, esperado)
		}
	}
}

func TestPerspicuidad(t *testing.T) {
	// 4 palabras, 6 sílabas, 1 oración → 206,835 − 62,3·(6/4) − 4/1 = 109,385
	got := Perspicuidad(oraciones([]string{"El gato come pan."}))
	if math.Abs(got-109.385) > 1e-9 {
		t.Fatalf("Perspicuidad = %v; esperado 109,385", got)
	}
}

func TestOraciones(t *testing.T) {
	lineas := []string{
		"Primera frase. Segunda frase con [un enlace](x.md) y **negrita**.",
		"sigue el mismo párrafo",
		"",
		"- Un ítem sin punto",
		"- Otro ítem",
		"",
		"El número 3.5 no corta.",
		"",
		"**El problema.** Sin revisor.",
	}
	got := oraciones(lineas)
	esperado := []string{
		"Primera frase",
		"Segunda frase con un enlace y negrita",
		"sigue el mismo párrafo",
		"Un ítem sin punto",
		"Otro ítem",
		"El número 3 5 no corta",
		"El problema",
		"Sin revisor",
	}
	if len(got) != len(esperado) {
		t.Fatalf("oraciones = %q; esperado %q", got, esperado)
	}
	for i := range esperado {
		if strings.Join(got[i], " ") != esperado[i] {
			t.Errorf("oración %d = %q; esperado %q", i, strings.Join(got[i], " "), esperado[i])
		}
	}
}

func TestLegibilidadIgnoraCodigoTablasYNavegacion(t *testing.T) {
	larga := strings.Repeat("palabra ", 25)
	raiz := wikiValida(t)
	reemplazar(t, raiz, tS001, "1. Uno.", "1. Uno.\n\n```\n"+larga+"\n```\n\n| "+larga+" |\n|---|\n| x |\n\nY `"+larga+"`.")
	reemplazar(t, raiz, tS001, "- ↑ Subir: [F001](README.md)", "- ↑ Subir: [F001](README.md) "+larga)
	if p := verificar(t, raiz); len(p) != 0 {
		t.Fatalf("código, tablas y Navegación no se miden: %q", p)
	}
}

func TestFallasLegibilidad(t *testing.T) {
	casos := []struct {
		nombre, nuevo, espera string
	}{
		{"oración larga", strings.TrimSpace(strings.Repeat("sol ", 21)) + ".", "oración de 21 palabras (máx. 20)"},
		{"legibilidad baja", "Documentación administrativa interdepartamental extraordinariamente complicada.", "legibilidad"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			raiz := wikiValida(t)
			reemplazar(t, raiz, tS001, "Se lee.", c.nuevo)
			exigeProblema(t, raiz, c.espera)
		})
	}
}

// TestOracionesConLinea (E000/F006/S004, hallazgos A1.1 y A1.2): cada oración sabe en qué línea empieza,
// también si no es la primera del párrafo o si antes hay un enlace partido en dos líneas.
func TestOracionesConLinea(t *testing.T) {
	lineas := []string{"Corta.", "Segunda empieza aquí", "y sigue.", "", "Uno dos [tres", "cuatro](../../a/b/c/d/e.md) fin.", "Otra.", "<b", ">Final.</b> Y [un](../../x/y", "z.md) cierre."}
	nums := []int{3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	_, got := oracionesConLinea(lineas, nums)
	want := []int{3, 4, 7, 9, 11, 11}
	if len(got) != len(want) {
		t.Fatalf("líneas = %v; esperado %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("oración %d empieza en %d; esperado %d", i, got[i], want[i])
		}
	}
	// Hallazgo A2.1: «<» o «[» sueltos en prosa no son etiqueta ni enlace; no desplazan la línea.
	sueltos := []string{"Si n < 3 sale.", "Sigue aquí.", "Un nombre con «<» y [sin cerrar.", "Final."}
	_, got = oracionesConLinea(sueltos, []int{20, 21, 22, 23})
	if fmt.Sprint(got) != "[20 21 22 23]" {
		t.Errorf("sueltos: líneas = %v; esperado [20 21 22 23]", got)
	}
	_, got = oracionesConLinea([]string{"Si n < 3 sale. Otra más."}, []int{30})
	if fmt.Sprint(got) != "[30 30]" {
		t.Errorf("suelto en una sola línea: %v", got)
	}
}
