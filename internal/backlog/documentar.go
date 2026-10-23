package backlog

// E001/F008/S007: completar los textos de pruebas viejas sin borrar su rojo ni su verde.

import (
	"fmt"
	"strings"
)

// Documentar reescribe valida, rojo cuando y verde cuando de pruebas ya registradas.
// Sirve en fichas archivadas y en arreglos (id padre/Xnn): no toca estado, rojo ni verde, solo los textos.
func Documentar(raiz, id string, lista []string) (Ficha, error) {
	ps, err := listaPruebas(lista)
	if err != nil {
		return Ficha{}, err
	}
	padre, arreglo := id, -1
	if k := strings.LastIndex(id, "/X"); k > 0 {
		padre = id[:k]
	}
	f, archivada, err := Ver(raiz, padre)
	if err != nil {
		return Ficha{}, err
	}
	pruebas := f.Pruebas
	if padre != id {
		for j, a := range f.Arreglos {
			if a.ID == id {
				arreglo, pruebas = j, a.Pruebas
			}
		}
		if arreglo < 0 {
			return Ficha{}, fmt.Errorf("%s no existe", id)
		}
	}
	for _, p := range ps {
		i := -1
		for j, q := range pruebas {
			if q.Nombre == p.Nombre {
				i = j
			}
		}
		if i < 0 {
			return Ficha{}, fmt.Errorf("%s no tiene la prueba %s: documentar solo completa pruebas registradas", id, p.Nombre)
		}
		pruebas[i] = p
	}
	return f, escribir(raiz, dirDe(archivada), f)
}
