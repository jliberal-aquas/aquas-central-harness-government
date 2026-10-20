package wiki

import "testing"

// Páginas de referencia (S003): docs/ref/<nombre>.md documentan lo que ya existe.
// Llevan las 8 secciones, ancla EJ-REF-<nombre>, al menos un diagrama Mermaid,
// ↑ Subir al inicio y el inicio las enlaza. No llevan «continúa:».

const tRef = "docs/ref/constitucion.md"

func conRef(t *testing.T) string {
	t.Helper()
	raiz := wikiValida(t)
	escribir(t, raiz, tRef, tPagina("EJ-REF-constitucion", "", mermaid("flowchart LR"),
		"- ↑ Subir: [Inicio](../README.md)"))
	reemplazar(t, raiz, tInicio, "- Glosario: [Glosario](glosario.md)",
		"- Glosario: [Glosario](glosario.md)\n- Referencia: [Constitución](ref/constitucion.md)")
	return raiz
}

func TestRefValidaPasa(t *testing.T) {
	if p := verificar(t, conRef(t)); len(p) != 0 {
		t.Fatalf("página de referencia válida debe pasar: %q", p)
	}
}

func TestRefAceptaCualquierDiagrama(t *testing.T) {
	raiz := conRef(t)
	reemplazar(t, raiz, tRef, "flowchart LR", "sequenceDiagram")
	if p := verificar(t, raiz); len(p) != 0 {
		t.Fatalf("referencia con sequenceDiagram debe pasar: %q", p)
	}
}

func TestFallasRef(t *testing.T) {
	casos := []struct {
		nombre string
		mutar  func(t *testing.T, raiz string)
		espera string
	}{
		{"sin ancla", func(t *testing.T, r string) { reemplazar(t, r, tRef, `<a id="EJ-REF-constitucion"></a>`, "") },
			"docs/ref/constitucion.md: falta ancla EJ-REF-constitucion"},
		{"sin diagrama", func(t *testing.T, r string) {
			reemplazar(t, r, tRef, "```mermaid\nflowchart LR\n    A --> B\n```\n", "Texto sin dibujo.")
		}, "docs/ref/constitucion.md: falta diagrama Mermaid en «Diagrama»"},
		{"inicio no la enlaza", func(t *testing.T, r string) {
			reemplazar(t, r, tInicio, "\n- Referencia: [Constitución](ref/constitucion.md)", "")
		}, "docs/README.md: Navegación no enlaza docs/ref/constitucion.md"},
		{"sin subir", func(t *testing.T, r string) { reemplazar(t, r, tRef, "- ↑ Subir: [Inicio](../README.md)", "- Nada.") },
			"docs/ref/constitucion.md: Navegación sin ↑ Subir"},
		{"falta sección", func(t *testing.T, r string) { reemplazar(t, r, tRef, "## Analogía", "## Otra") },
			"docs/ref/constitucion.md: falta sección «Analogía»"},
		{"nombre inválido", func(t *testing.T, r string) { escribir(t, r, "docs/ref/Mal_Nombre.md", "# X\n") },
			"docs/ref/Mal_Nombre.md: fuera del árbol"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			raiz := conRef(t)
			c.mutar(t, raiz)
			exigeProblema(t, raiz, c.espera)
		})
	}
}
