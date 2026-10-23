package backlog

// E001/F008/S007: completar los textos de pruebas viejas sin borrar su rojo ni su verde.

import "fmt"

// Documentar reescribe valida, rojo cuando y verde cuando de pruebas ya registradas.
// Sirve también en fichas archivadas: no toca estado, rojo ni verde, solo los textos.
func Documentar(raiz, id string, lista []string) (Ficha, error) {
	ps, err := listaPruebas(lista)
	if err != nil {
		return Ficha{}, err
	}
	f, archivada, err := Ver(raiz, id)
	if err != nil {
		return Ficha{}, err
	}
	for _, p := range ps {
		i := -1
		for j, q := range f.Pruebas {
			if q.Nombre == p.Nombre {
				i = j
			}
		}
		if i < 0 {
			return Ficha{}, fmt.Errorf("%s no tiene la prueba %s: documentar solo completa pruebas registradas", id, p.Nombre)
		}
		f.Pruebas[i] = p
	}
	return f, escribir(raiz, dirDe(archivada), f)
}

// Deuda: pruebas registradas sin «rojo cuando» o «verde cuando», en fichas abiertas y archivadas.
func Deuda(raiz string) (int, error) {
	fs, err := Todas(raiz)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, f := range fs {
		for _, p := range f.Pruebas {
			if p.RojoCuando == "" || p.VerdeCuando == "" {
				n++
			}
		}
	}
	return n, nil
}
