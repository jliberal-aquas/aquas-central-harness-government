// Package turno guarda la auditoría de turnos (E001/F002/S001). El hook UserPromptSubmit abre
// el turno con el mensaje literal del humano; el hook Stop guarda la respuesta del agente.
// Ninguno de los dos pasa por el agente. Cada turno es un Markdown en Carpeta/AAAA-MM/T000001.md,
// fuera de Git (decisión del canónico). Lo pendiente por sesión vive en Estado, dentro de .aquas/.
// Límite: el pedido refinado y la diferencia A→B no son de este slice.
package turno

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// Carpeta: donde vive la auditoría de turnos; la guardia la protege desde aquí.
	Carpeta = "auditoria/turnos"
	// Estado: el último turno de cada sesión, para que el cierre sepa dónde escribir.
	Estado = ".aquas/turno.json"
)

type abierto struct {
	ID      string `json:"id"`
	Sesion  string `json:"sesion"`
	Ruta    string `json:"ruta"`
	Literal string `json:"literal"`
	Abierto string `json:"abierto"`
}

var reID = regexp.MustCompile(`^T(\d{6})\.md$`)

// siguiente: el mayor número existente + 1; un hueco no se rellena (los ID no se reutilizan).
func siguiente(raiz string) (int, error) {
	meses, err := os.ReadDir(filepath.Join(raiz, filepath.FromSlash(Carpeta)))
	if os.IsNotExist(err) {
		return 1, nil
	} else if err != nil {
		return 0, err
	}
	mayor := 0
	for _, m := range meses {
		if !m.IsDir() {
			continue
		}
		xs, err := os.ReadDir(filepath.Join(raiz, filepath.FromSlash(Carpeta), m.Name()))
		if err != nil {
			return 0, err
		}
		for _, x := range xs {
			if s := reID.FindStringSubmatch(x.Name()); s != nil {
				if n, _ := strconv.Atoi(s[1]); n > mayor {
					mayor = n
				}
			}
		}
	}
	return mayor + 1, nil
}

func leerEstado(raiz string) (map[string]abierto, error) {
	e := map[string]abierto{}
	b, err := os.ReadFile(filepath.Join(raiz, filepath.FromSlash(Estado)))
	if os.IsNotExist(err) {
		return e, nil
	} else if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &e); err != nil {
		return nil, fmt.Errorf("%s ilegible: %w", Estado, err)
	}
	return e, nil
}

func escribir(raiz, rel string, b []byte) error {
	p := filepath.Join(raiz, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, b, 0o644)
}

// cerca: bloque de código cuya cerca es más larga que cualquier racha de backticks del texto,
// así el texto queda intacto aunque traiga sus propios bloques.
func cerca(s string) string {
	largo, racha := 3, 0
	for _, r := range s {
		if r == '`' {
			racha++
			if racha >= largo {
				largo = racha + 1
			}
		} else {
			racha = 0
		}
	}
	f := strings.Repeat("`", largo)
	return f + "text\n" + s + "\n" + f + "\n"
}

func render(a abierto, respuesta, cerrado string) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n- Sesión: `%s`\n- Abierto: %s\n- Cerrado: %s\n\n", a.ID, a.Sesion, a.Abierto, cerrado)
	b.WriteString("## Literal\n\n" + cerca(a.Literal) + "\n## Respuesta\n\n")
	if cerrado == "—" {
		b.WriteString("—\n")
	} else {
		b.WriteString(cerca(respuesta))
	}
	return []byte(b.String())
}

// Abrir crea el turno siguiente con el literal y lo deja como el abierto de la sesión.
func Abrir(raiz, sesion, literal string, ahora time.Time) (string, error) {
	e, err := leerEstado(raiz)
	if err != nil {
		return "", err
	}
	n, err := siguiente(raiz)
	if err != nil {
		return "", err
	}
	ahora = ahora.UTC()
	id := fmt.Sprintf("T%06d", n)
	a := abierto{ID: id, Sesion: sesion, Ruta: Carpeta + "/" + ahora.Format("2006-01") + "/" + id + ".md", Literal: literal, Abierto: ahora.Format(time.RFC3339)}
	if err := escribir(raiz, a.Ruta, render(a, "", "—")); err != nil {
		return "", err
	}
	e[sesion] = a
	b, _ := json.MarshalIndent(e, "", "  ")
	if err := escribir(raiz, Estado, append(b, '\n')); err != nil {
		return "", err
	}
	return a.Ruta, nil
}

// Cerrar guarda la respuesta en el último turno de la sesión; sin turno abierto, no escribe.
// Si Stop se repite (otro hook frenó el cierre), reescribe: gana la última respuesta.
func Cerrar(raiz, sesion, respuesta string, ahora time.Time) (string, error) {
	e, err := leerEstado(raiz)
	if err != nil {
		return "", err
	}
	a, ok := e[sesion]
	if !ok {
		return "", nil
	}
	if err := escribir(raiz, a.Ruta, render(a, respuesta, ahora.UTC().Format(time.RFC3339))); err != nil {
		return "", err
	}
	return a.Ruta, nil
}
