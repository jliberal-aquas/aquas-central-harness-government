package backlog

// E001/F009/S003: backlog editar cambia una ficha que aún no arranca y guarda lo anterior.

import (
	"errors"
	"time"
)

// Editar cambia campos de una ficha en BORRADOR o LISTO y guarda qué cambió, el valor anterior y el motivo.
func Editar(raiz, id string, campos []string, ahora time.Time) (Ficha, error) {
	return Ficha{}, errors.New("sin implementar (E001/F009/S003)")
}
