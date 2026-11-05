package backlog

// E000/F014/S001 (D000125): un slice nace con su página de contrato. Sin página, la aprobación del humano
// fallaba recién al escribir «apruebo»; ahora ese estado inválido no existe. Página ya escrita: no se toca.

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// paginaNace escribe docs/<E>/<F>/<S>.md de un slice nuevo si no existe; feature e iniciativa no cambian.
func paginaNace(raiz string, f Ficha) error {
	if f.Tipo != "slice" {
		return nil
	}
	rel := pagina(f.ID)
	ruta := filepath.Join(raiz, filepath.FromSlash(rel))
	if _, err := os.Stat(ruta); err == nil {
		return nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("no pude revisar la página %s: %w", rel, err)
	}
	if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
		return fmt.Errorf("no pude crear la página %s: %w", rel, err)
	}
	if err := os.WriteFile(ruta, []byte(paginaInicial(f)), 0o644); err != nil {
		return fmt.Errorf("no pude crear la página %s: %w", rel, err)
	}
	return nil
}

// paginaInicial: esqueleto que pasa el gate wiki; el agente completa el contrato y los casos borde antes del rojo.
func paginaInicial(f Ficha) string {
	p := strings.Split(f.ID, "/")
	e, fe, s := p[0], p[1], p[2]
	pendiente := MarcaPendiente + " el agente lo decide antes del apruebo"
	var casos strings.Builder
	for _, c := range []string{
		"cada tipo de ficha", "cada estado de origen", "padre e hijo, activo y archivado",
		"entrada ausente, vacía, mal escrita o duplicada", "choque con otro control",
		"quién habla: humano, agente o herramienta", "fallo a mitad de camino",
	} {
		casos.WriteString("| " + c + " | no | " + pendiente + " |\n")
	}
	return "# " + s + " · " + f.Titulo + "\n\n" +
		"[Inicio](../../README.md) › [" + e + "](../README.md) › [" + fe + "](README.md) › " + s + "\n\n" +
		f.QueHace + "\n\n" +
		"## En una frase\n\n" + f.QueHace + "\n\n" +
		"## Para qué existe\n\n" +
		"**El problema.** " + f.Motivo + "\n\n" +
		"**La solución.** " + f.QueHace + "\n\n" +
		"**Contrato.**\n\n| Regla | Valor |\n|---|---|\n| " + pendiente + " | — |\n\n" +
		"**Casos borde.**\n\n| Categoría | Aplica | Prueba o motivo |\n|---|---|---|\n" + casos.String() + "\n" +
		"## Analogía\n\n" + pendiente + ".\n\n" +
		"## Diagrama\n\n```mermaid\nsequenceDiagram\n    participant A as agente\n    participant Q as aquas\n    A->>Q: " + s + "\n```\n\n" +
		"## Ejemplo de punta a punta\n\n<a id=\"EJ-" + e + "-" + fe + "-" + s + "\"></a>\n\n" +
		"continúa: [EJ-" + e + "-" + fe + "](README.md#EJ-" + e + "-" + fe + ")\n\n1. " + pendiente + ".\n\n" +
		"## Paso a paso\n\n1. " + pendiente + ".\n\n" +
		"## Cómo se usa\n\n- " + pendiente + ".\n\n" +
		"## Contrato y evidencia\n\n" +
		"## Navegación\n\n- ↑ Subir: [" + fe + "](README.md)\n"
}
