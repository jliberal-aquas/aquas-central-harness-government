package wiki

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func escribir(t *testing.T, raiz, rel, contenido string) {
	t.Helper()
	p := filepath.Join(raiz, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(contenido), 0o644); err != nil {
		t.Fatal(err)
	}
}

func leer(t *testing.T, raiz, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(raiz, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// reemplazar cambia viejo por nuevo en rel; falla si viejo no aparece.
func reemplazar(t *testing.T, raiz, rel, viejo, nuevo string) {
	t.Helper()
	c := leer(t, raiz, rel)
	if !strings.Contains(c, viejo) {
		t.Fatalf("%s no contiene %q", rel, viejo)
	}
	escribir(t, raiz, rel, strings.Replace(c, viejo, nuevo, 1))
}

func tSeccion(titulo, cuerpo string) string { return "## " + titulo + "\n\n" + cuerpo + "\n\n" }

func mermaid(tipo string) string { return "```mermaid\n" + tipo + "\n    A --> B\n```\n" }

// tMigas y tContrato derivan del id del ejemplo: solo las páginas E/F/S los llevan.
func tMigas(ej string) string {
	ids := strings.Split(strings.TrimPrefix(ej, "EJ-"), "-")
	if !strings.HasPrefix(ej, "EJ-E") {
		return ""
	}
	sube := []string{"../README.md", "../../README.md", "../../README.md"}[len(ids)-1]
	m := "[Inicio](" + sube + ")"
	switch len(ids) {
	case 2:
		m += " › [" + ids[0] + "](../README.md)"
	case 3:
		m += " › [" + ids[0] + "](../README.md) › [" + ids[1] + "](README.md)"
	}
	return m + " › " + ids[len(ids)-1] + "\n\n"
}

func tContrato(ej string) string {
	if !strings.HasPrefix(ej, "EJ-E") {
		return ""
	}
	return tSeccion("Contrato y evidencia", "Lo prometido y su prueba.")
}

func tPagina(ej, continua, diagrama, nav string) string {
	return "# Título\n\n" + tMigas(ej) +
		tSeccion("En una frase", "Algo simple.") +
		tSeccion("Para qué existe", "Problema → solución.") +
		tSeccion("Analogía", "Como un mapa.") +
		tSeccion("Diagrama", diagrama) +
		tSeccion("Ejemplo de punta a punta", `<a id="`+ej+`"></a>`+"\n\n"+continua+"Texto del ejemplo.") +
		tSeccion("Paso a paso", "1. Uno.") +
		tSeccion("Cómo se usa", "Se lee.") + tContrato(ej) +
		tSeccion("Navegación", nav)
}

const (
	tInicio = "docs/README.md"
	tE001   = "docs/E001/README.md"
	tF001   = "docs/E001/F001/README.md"
	tS001   = "docs/E001/F001/S001.md"
)

func wikiValida(t *testing.T) string {
	t.Helper()
	raiz := t.TempDir()
	escribir(t, raiz, tInicio, tPagina("EJ-INICIO", "", mermaid("flowchart LR"),
		"- Abajo: [E001](E001/README.md)\n- Glosario: [Glosario](glosario.md)"))
	escribir(t, raiz, "docs/glosario.md", "# Glosario\n\nPalabras técnicas.\n\n## zeta\n\nUna letra.\n\nEjemplo: la zeta va al final.\n")
	escribir(t, raiz, tE001, tPagina("EJ-E001", "", mermaid("flowchart TD"),
		"- ↑ Subir: [Inicio](../README.md)\n- Abajo: [F001](F001/README.md)"))
	escribir(t, raiz, tF001, tPagina("EJ-E001-F001", "continúa: [EJ-E001](../README.md#EJ-E001)\n\n",
		mermaid("flowchart LR")+"\n"+mermaid("stateDiagram-v2"),
		"- ↑ Subir: [E001](../README.md)\n- Abajo: [S001](S001.md)"))
	escribir(t, raiz, tS001, tPagina("EJ-E001-F001-S001", "continúa: [EJ-E001-F001](README.md#EJ-E001-F001)\n\n",
		mermaid("sequenceDiagram"), "- ↑ Subir: [F001](README.md)"))
	return raiz
}

func verificar(t *testing.T, raiz string) []string {
	t.Helper()
	p, err := Verificar(raiz)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func exigeProblema(t *testing.T, raiz, sub string) {
	t.Helper()
	p := verificar(t, raiz)
	for _, x := range p {
		if strings.Contains(x, sub) {
			return
		}
	}
	t.Fatalf("se esperaba problema con %q; hubo: %q", sub, p)
}

func TestWikiValidaPasa(t *testing.T) {
	if p := verificar(t, wikiValida(t)); len(p) != 0 {
		t.Fatalf("wiki válida no debe tener problemas: %q", p)
	}
}

func TestEnlaceEnBloqueDeCodigoSeIgnora(t *testing.T) {
	raiz := wikiValida(t)
	reemplazar(t, raiz, tS001, "1. Uno.", "1. Uno.\n\n```\n[x](noexiste.md)\n```\n\nY `[y](tampoco.md)`.")
	if p := verificar(t, raiz); len(p) != 0 {
		t.Fatalf("enlaces en código no se validan: %q", p)
	}
}

func TestAnclaPorTituloVale(t *testing.T) {
	raiz := wikiValida(t)
	reemplazar(t, raiz, tS001, "Se lee.", "Ver [x](README.md#para-qué-existe) y [y](#cómo-se-usa).")
	if p := verificar(t, raiz); len(p) != 0 {
		t.Fatalf("anclas de título válidas: %q", p)
	}
}

func TestFallas(t *testing.T) {
	casos := []struct {
		nombre string
		mutar  func(t *testing.T, raiz string)
		espera string
	}{
		{"sin tInicio", func(t *testing.T, r string) { os.Remove(filepath.Join(r, "docs", "README.md")) },
			"falta docs/README.md"},
		{"falta sección", func(t *testing.T, r string) { reemplazar(t, r, tS001, "## Analogía", "## Otra cosa") },
			"falta sección «Analogía»"},
		{"sección vacía", func(t *testing.T, r string) { reemplazar(t, r, tS001, "Se lee.", "") },
			"sección vacía «Cómo se usa»"},
		{"orden", func(t *testing.T, r string) {
			reemplazar(t, r, tS001, "## En una frase", "## TMP")
			reemplazar(t, r, tS001, "## Para qué existe", "## En una frase")
			reemplazar(t, r, tS001, "## TMP", "## Para qué existe")
		}, "fuera de orden"},
		{"F sin estados", func(t *testing.T, r string) { reemplazar(t, r, tF001, "stateDiagram-v2", "flowchart TD") },
			"falta diagrama stateDiagram"},
		{"E sin flujo", func(t *testing.T, r string) { reemplazar(t, r, tE001, "flowchart TD", "sequenceDiagram") },
			"falta diagrama flowchart"},
		{"S sin datos ni secuencia", func(t *testing.T, r string) { reemplazar(t, r, tS001, "sequenceDiagram", "flowchart LR") },
			"falta diagrama de datos o secuencia"},
		{"enlace roto", func(t *testing.T, r string) { reemplazar(t, r, tS001, "Se lee.", "Ver [x](no-existe.md).") },
			"enlace roto"},
		{"ancla inexistente", func(t *testing.T, r string) { reemplazar(t, r, tS001, "Se lee.", "Ver [x](README.md#nada).") },
			"ancla inexistente"},
		{"sin continúa", func(t *testing.T, r string) {
			reemplazar(t, r, tF001, "continúa: [EJ-E001](../README.md#EJ-E001)", "")
		}, "falta «continúa:»"},
		{"continúa a otro padre", func(t *testing.T, r string) {
			reemplazar(t, r, tS001, "[EJ-E001-F001](README.md#EJ-E001-F001)", "[EJ-E001](../README.md#EJ-E001)")
		}, "continúa debe apuntar a EJ-E001-F001"},
		{"sin ancla de ejemplo", func(t *testing.T, r string) { reemplazar(t, r, tF001, `<a id="EJ-E001-F001"></a>`, "") },
			"falta ancla EJ-E001-F001"},
		{"página huérfana", func(t *testing.T, r string) {
			escribir(t, r, "docs/E001/F001/S002.md", tPagina("EJ-E001-F001-S002",
				"continúa: [EJ-E001-F001](README.md#EJ-E001-F001)\n\n", mermaid("erDiagram"), "- ↑ Subir: [F001](README.md)"))
		}, "inalcanzable"},
		{"padre no enlaza hijo", func(t *testing.T, r string) {
			escribir(t, r, "docs/E001/F001/S002.md", tPagina("EJ-E001-F001-S002",
				"continúa: [EJ-E001-F001](README.md#EJ-E001-F001)\n\n", mermaid("erDiagram"), "- ↑ Subir: [F001](README.md)"))
		}, "Navegación no enlaza docs/E001/F001/S002.md"},
		{"sin subir", func(t *testing.T, r string) { reemplazar(t, r, tS001, "- ↑ Subir: [F001](README.md)", "- Nada.") },
			"Navegación sin ↑ Subir"},
		{"fuera del árbol", func(t *testing.T, r string) { escribir(t, r, "docs/notas.md", "# Notas\n") },
			"fuera del árbol"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			raiz := wikiValida(t)
			c.mutar(t, raiz)
			exigeProblema(t, raiz, c.espera)
		})
	}
}
