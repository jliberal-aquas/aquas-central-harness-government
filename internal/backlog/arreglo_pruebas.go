package backlog

// D000020: un arreglo registra sus propias pruebas y pasa rojo y verde con su ID (padre/Xnn).

import (
	"fmt"
	"strings"
	"time"
)

// padreDeArreglo: el ID del slice o feature dueño, si id es de un arreglo (padre/Xnn).
func padreDeArreglo(id string) (string, bool) {
	k := strings.LastIndex(id, "/X")
	if k <= 0 {
		return "", false
	}
	return id[:k], true
}

// cambiarArreglo carga la ficha dueña, aplica cambio al arreglo id y la guarda donde vive (activa o archivada).
func cambiarArreglo(raiz, id string, cambio func(a *Arreglo) error) (Ficha, error) {
	padre, _ := padreDeArreglo(id)
	f, archivada, err := Ver(raiz, padre)
	if err != nil {
		return Ficha{}, err
	}
	for j := range f.Arreglos {
		if f.Arreglos[j].ID == id {
			if err := cambio(&f.Arreglos[j]); err != nil {
				return Ficha{}, err
			}
			return f, escribir(raiz, dirDe(archivada), f)
		}
	}
	return Ficha{}, fmt.Errorf("%s no existe", id)
}

// rojoDe corre cada prueba sola; exige que cada una falle por sí misma.
func rojoDe(raiz string, ps []Prueba, ahora time.Time) (*Registro, error) {
	res := map[string]string{}
	var malas []string
	for _, p := range ps {
		r, err := correr(raiz, []Prueba{p})
		if err != nil {
			return nil, err
		}
		if x := r[p.Nombre]; x.accion == "fail" {
			res[p.Nombre] = x.salida
		} else {
			malas = append(malas, fmt.Sprintf("%s: %s", p.Nombre, x.motivo("corrida sola, pasa: no falla por sí misma")))
		}
	}
	if len(malas) > 0 {
		return nil, fmt.Errorf("rojo rechazado; cada prueba debe fallar sola por su motivo:\n  - %s", strings.Join(malas, "\n  - "))
	}
	return &Registro{Fecha: ahora.UTC().Format(time.RFC3339), Head: head(raiz), Resultado: res}, nil
}

// verdeDe corre las pruebas juntas tras el rojo; exige que todas pasen.
func verdeDe(raiz, id string, ps []Prueba, rojo *Registro, ahora time.Time) (*Registro, error) {
	if rojo == nil {
		return nil, fmt.Errorf("%s sin rojo: primero aquas backlog rojo %s", id, id)
	}
	r, err := correr(raiz, ps)
	if err != nil {
		return nil, err
	}
	var malas []string
	for _, p := range ps {
		if x := r[p.Nombre]; x.accion != "pass" {
			malas = append(malas, fmt.Sprintf("%s: %s", p.Nombre, x.motivo("")))
		}
	}
	if len(malas) > 0 {
		return nil, fmt.Errorf("verde rechazado; todas deben pasar:\n  - %s", strings.Join(malas, "\n  - "))
	}
	return &Registro{Fecha: ahora.UTC().Format(time.RFC3339), Head: head(raiz)}, nil
}

// pruebasArreglo fija las pruebas del arreglo; cambiarlas borra su rojo y su verde.
func pruebasArreglo(raiz, id string, ps []Prueba) (Ficha, error) {
	return cambiarArreglo(raiz, id, func(a *Arreglo) error {
		a.Pruebas, a.Rojo, a.Verde = ps, nil, nil
		return nil
	})
}

// rojoArreglo registra el rojo del arreglo.
func rojoArreglo(raiz, id string, ahora time.Time) (Ficha, error) {
	return cambiarArreglo(raiz, id, func(a *Arreglo) error {
		if len(a.Pruebas) == 0 {
			return fmt.Errorf("%s sin pruebas: aquas backlog pruebas %s …", id, id)
		}
		r, err := rojoDe(raiz, a.Pruebas, ahora)
		if err == nil {
			a.Rojo, a.Verde = r, nil
		}
		return err
	})
}

// verdeArreglo registra el verde del arreglo.
func verdeArreglo(raiz, id string, ahora time.Time) (Ficha, error) {
	return cambiarArreglo(raiz, id, func(a *Arreglo) error {
		r, err := verdeDe(raiz, id, a.Pruebas, a.Rojo, ahora)
		if err == nil {
			a.Verde = r
		}
		return err
	})
}
