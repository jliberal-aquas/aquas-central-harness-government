package canario

import (
	"fmt"
	"strings"
)

// Cada función rompe a propósito un linter del gate.

func noUsada() {} // unused

// Ineficaz asigna un valor que nadie lee: ineffassign.
func Ineficaz() int {
	x := 1
	x = 2
	return x
}

// Formato pasa un texto a %d: govet.
func Formato() { fmt.Printf("%d\n", "texto") }


// Descarta descarta el resultado de una función pura: staticcheck (SA4017).
func Descarta() { strings.ToUpper("x") }
