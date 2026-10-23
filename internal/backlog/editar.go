package backlog

// E001/F009/S003: backlog editar cambia una ficha que aún no arranca y guarda lo anterior.

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

// Editar cambia título o bloqueado_por de una ficha en BORRADOR o LISTO y guarda campo, antes, después y motivo.
func Editar(raiz, id string, campos []string, ahora time.Time) (Ficha, error) {
	v := map[string]string{}
	var orden []string
	for _, c := range campos {
		k, val, ok := strings.Cut(c, "=")
		if !ok || !esta([]string{"titulo", "bloqueado_por", "motivo"}, k) {
			return Ficha{}, fmt.Errorf("campo %q no editable: usa titulo=… o bloqueado_por=A,B, más motivo=…", c)
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
		return Ficha{}, fmt.Errorf("nada que editar: usa titulo=… o bloqueado_por=A,B")
	}
	act, arc, err := todas(raiz)
	if err != nil {
		return Ficha{}, err
	}
	f, hay := act[id]
	if !hay || (f.Estado != "BORRADOR" && f.Estado != "LISTO") {
		return Ficha{}, fmt.Errorf("%s solo se edita en BORRADOR o LISTO: si ya arrancó, usa backlog alcance; si cerró, backlog arreglo o una ficha nueva", id)
	}
	existe := func(x string) bool { _, a := act[x]; _, b := arc[x]; return a || b }
	fecha := ahora.UTC().Format(time.RFC3339)
	for _, k := range orden {
		antes, despues := f.Titulo, v[k]
		if k == "titulo" {
			if despues == "" {
				return Ficha{}, fmt.Errorf("titulo vacío")
			}
			f.Titulo = despues
		} else {
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
	return f, escribir(raiz, Activo, f)
}
