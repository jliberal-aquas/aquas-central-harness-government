package backlog

// E001/F009/S003: backlog edit cambia una ficha que aún no arranca y guarda lo anterior; que_hace en cualquier estado (E000/F007/S002).

import (
	"fmt"
	"strings"
	"time"
)

// Edicion: un campo cambiado antes de arrancar, con su valor anterior y el motivo.
type Edicion struct {
	Fecha   string `json:"fecha"`
	Campo   string `json:"campo"`
	Antes   string `json:"antes"`
	Despues string `json:"despues"`
	Motivo  string `json:"motivo"`
}

// Editar cambia título o bloqueado_por (solo en BORRADOR o LISTO) o que_hace (en cualquier estado) y guarda el cambio con su motivo.
// que_hace (E000/F007/S002) se corrige en cualquier estado, activa o archivada: describe la ficha y no entra en la
// huella de aprobación; sigue la regla de `new` (una oración de MaxPalabrasQueHace palabras como máximo).
func Editar(raiz, id string, campos []string, ahora time.Time) (Ficha, error) {
	v := map[string]string{}
	var orden []string
	for _, c := range campos {
		k, val, ok := strings.Cut(c, "=")
		if !ok || !esta([]string{"titulo", "bloqueado_por", "que_hace", "motivo"}, k) {
			return Ficha{}, fmt.Errorf("campo %q no editable: usa titulo=…, bloqueado_por=A,B o que_hace=…, más motivo=…", c)
		}
		if _, rep := v[k]; rep {
			return Ficha{}, fmt.Errorf("campo %s repetido", k)
		}
		v[k] = strings.TrimSpace(val)
		if k != "motivo" {
			orden = append(orden, k)
		}
	}
	if v["motivo"] == "" {
		return Ficha{}, fmt.Errorf("falta motivo=…: di por qué cambia")
	}
	if len(orden) == 0 {
		return Ficha{}, fmt.Errorf("nada que editar: usa titulo=…, bloqueado_por=A,B o que_hace=…")
	}
	if q, hay := v["que_hace"]; hay {
		if q == "" {
			return Ficha{}, fmt.Errorf("que_hace vacío: una oración que diga qué hace")
		}
		if n := len(strings.Fields(q)); n > MaxPalabrasQueHace {
			return Ficha{}, fmt.Errorf("que_hace tiene %d palabras: el máximo es %d", n, MaxPalabrasQueHace)
		}
	}
	act, arc, err := todas(raiz)
	if err != nil {
		return Ficha{}, err
	}
	f, hay := act[id]
	archivada := false
	if !hay && len(orden) == 1 && orden[0] == "que_hace" { // solo que_hace: también archivadas
		f, archivada = arc[id]
		hay = archivada
	}
	soloQueHace := len(orden) == 1 && orden[0] == "que_hace"
	if !hay || (!soloQueHace && f.Estado != "BORRADOR" && f.Estado != "LISTO") {
		if _, enArchivo := arc[id]; !hay && !enArchivo {
			return Ficha{}, fmt.Errorf("%s no existe en el backlog", id)
		}
		return Ficha{}, fmt.Errorf("%s solo se edita en BORRADOR o LISTO: si ya arrancó, usa backlog scope; si cerró, backlog fix o una ficha nueva", id)
	}
	existe := func(x string) bool { _, a := act[x]; _, b := arc[x]; return a || b }
	fecha := ahora.UTC().Format(time.RFC3339)
	for _, k := range orden {
		antes, despues := f.Titulo, v[k]
		switch k {
		case "que_hace":
			antes, f.QueHace = f.QueHace, despues
		case "titulo":
			if despues == "" {
				return Ficha{}, fmt.Errorf("titulo vacío")
			}
			f.Titulo = despues
		default:
			antes = strings.Join(f.BloqueadoPor, ",")
			bloq := []string{}
			for _, b := range strings.Split(despues, ",") {
				if b = strings.TrimSpace(b); b == "" {
					continue
				}
				if !existe(b) || b == id {
					return Ficha{}, fmt.Errorf("bloqueado_por: %s no existe o es la misma ficha", b)
				}
				if !esta(bloq, b) {
					bloq = append(bloq, b)
				}
			}
			f.BloqueadoPor, despues = bloq, strings.Join(bloq, ",")
		}
		f.Ediciones = append(f.Ediciones, Edicion{Fecha: fecha, Campo: k, Antes: antes, Despues: despues, Motivo: v["motivo"]})
	}
	return f, escribir(raiz, dirDe(archivada), f)
}
