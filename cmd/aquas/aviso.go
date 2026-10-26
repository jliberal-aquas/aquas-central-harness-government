package main

import (
	"fmt"
	"io"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/hijo"
)

// avisarVersion (E002/F002/S003) escribe el aviso de versión del hijo en w, salvo en install y update.
// No devuelve nada: el aviso nunca cambia el resultado del comando.
func avisarVersion(w io.Writer, comando, raiz, version string) {
	if comando == "install" || comando == "update" {
		return // ellos mismos ponen la versión del lock
	}
	if a := hijo.AvisoVersion(raiz, version); a != "" {
		fmt.Fprintln(w, a)
	}
}
