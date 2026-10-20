package wiki

import (
	"os"
	"path/filepath"
	"testing"
)

const tGlosario = "docs/glosario.md"

func TestGlosarioEnlazadoPasa(t *testing.T) {
	raiz := wikiValida(t)
	reemplazar(t, raiz, tGlosario, "## zeta", "## commit\n\nGuardado en git.\n\nEjemplo: un commit por cambio.\n\n## zeta")
	reemplazar(t, raiz, tS001, "Se lee.",
		"La [zeta](../../glosario.md#zeta) manda. Otra zeta sin enlace ya vale. Y `zeta` en código no cuenta. El pre-commit corre antes del [commit](../../glosario.md#commit).")
	if p := verificar(t, raiz); len(p) != 0 {
		t.Fatalf("primera aparición enlazada debe pasar: %q", p)
	}
}

func TestFallasGlosario(t *testing.T) {
	casos := []struct {
		nombre string
		mutar  func(t *testing.T, raiz string)
		espera string
	}{
		{"sin glosario", func(t *testing.T, r string) { os.Remove(filepath.Join(r, "docs", "glosario.md")) },
			"falta docs/glosario.md"},
		{"término sin enlace", func(t *testing.T, r string) { reemplazar(t, r, tS001, "Se lee.", "La zeta manda.") },
			"primera aparición de «zeta» sin enlace al glosario (escribe [zeta](../../glosario.md#zeta))"},
		{"mayúsculas cuentan", func(t *testing.T, r string) { reemplazar(t, r, tE001, "Se lee.", "Zeta manda.") },
			"docs/E001/README.md: primera aparición de «zeta»"},
		{"variante sin enlace", func(t *testing.T, r string) {
			reemplazar(t, r, tGlosario, "Una letra.", "Una letra.\n\nVariantes: zetas")
			reemplazar(t, r, tS001, "Se lee.", "Las zetas mandan.")
		}, "primera aparición de «zeta»"},
		{"enlace a otra entrada", func(t *testing.T, r string) {
			reemplazar(t, r, tS001, "Se lee.", "La [zeta](../../glosario.md#otra) manda.")
		}, "primera aparición de «zeta» sin enlace"},
		{"entrada sin ejemplo", func(t *testing.T, r string) { reemplazar(t, r, tGlosario, "Ejemplo: la zeta va al final.", "") },
			"docs/glosario.md: entrada «zeta» sin línea «Ejemplo:»"},
		{"entrada sin definición", func(t *testing.T, r string) { reemplazar(t, r, tGlosario, "Una letra.", "") },
			"docs/glosario.md: entrada «zeta» sin definición"},
		{"glosario inalcanzable", func(t *testing.T, r string) {
			reemplazar(t, r, tInicio, "\n- Glosario: [Glosario](glosario.md)", "")
		}, "docs/glosario.md: inalcanzable"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			raiz := wikiValida(t)
			c.mutar(t, raiz)
			exigeProblema(t, raiz, c.espera)
		})
	}
}

func TestExtensionDeArchivoNoEsTermino(t *testing.T) {
	raiz := wikiValida(t)
	reemplazar(t, raiz, tS001, "Se lee.", "Abre nota.zeta hoy.")
	if p := verificar(t, raiz); len(p) != 0 {
		t.Fatalf("«.zeta» como extensión no es término: %q", p)
	}
}
