package main

import "github.com/jliberal-aquas/aquas-central-harness-government/internal/calidad"

// revisarGo (E001/F013/S001) corre la calidad Go de la central para el gate.
func revisarGo(raiz string) []string { return calidad.RevisarCentral(raiz) }
