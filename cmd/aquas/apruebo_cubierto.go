package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/backlog"
	"github.com/jliberal-aquas/aquas-central-harness-government/internal/propuesta"
)

// pedidoCubierto (E001/F017/S002, D000069, D000084, D000099): freno del cierre si el agente pide un apruebo que la
// aprobación vigente de una feature ya cubre; "" si no. Cubierto: el slice está en la foto del apruebo de su feature,
// o la feature ya aprobada tiene todos sus slices activos en la foto. El «dale» queda como criterio humano.
func pedidoCubierto(raiz, msg string) string {
	pedidos := propuesta.PedidosDeApruebo(msg)
	if len(pedidos) == 0 {
		return ""
	}
	fs, err := backlog.Todas(raiz)
	if err != nil {
		return "" // sin backlog legible no se afirma que algo esté cubierto
	}
	por := map[string]backlog.Ficha{}
	for _, f := range fs {
		por[f.ID] = f
	}
	// aprobada: feature activa con aprobación cuya huella sigue igual; si el contrato cambió, el arnés pide el
	// apruebo otra vez y pedirlo es legítimo (A1.1).
	aprobada := func(id string) (backlog.Ficha, bool) {
		f, ok := por[id]
		if !ok || f.Tipo != "feature" || f.Aprobacion == nil || f.Estado == "CERRADO" || f.Estado == "CANCELADO" {
			return f, false
		}
		h, err := backlog.Huella(raiz, f)
		return f, err == nil && h == f.Aprobacion.Huella
	}
	const sigue = "sigue sin preguntar y vuelve al humano solo por duda real, cambio de alcance, acción irreversible no cubierta o feature terminada"
	for _, p := range pedidos {
		id := p.ID
		if i := strings.LastIndex(id, "/S"); i > 0 {
			if f, ok := aprobada(id[:i]); ok {
				if aprobadas, enFoto := f.Aprobacion.Foto[id]; enFoto && pruebasIntactas(por[id], aprobadas) {
					return fmt.Sprintf("Cierre: pides «apruebo %s» y el apruebo de %s ya cubre ese slice: %s (E001/F017/S002).", id, f.ID, sigue)
				}
			}
			continue
		}
		// Pedir «con push» sobre una aprobación sin push es legítimo: el push es una acción externa no cubierta (A2.2).
		if f, ok := aprobada(id); ok && todosEnFoto(fs, f) && (!p.Push || f.Aprobacion.Push) {
			return fmt.Sprintf("Cierre: pides «apruebo %s» y su aprobación vigente ya cubre todos sus slices: %s (E001/F017/S002).", id, sigue)
		}
	}
	return ""
}

// pruebasIntactas: cada prueba aprobada en la foto sigue en el slice; si cambió, el arnés pide el apruebo otra vez.
func pruebasIntactas(slice backlog.Ficha, aprobadas []backlog.Prueba) bool {
	for _, a := range aprobadas {
		if !slices.Contains(slice.Pruebas, a) {
			return false
		}
	}
	return true
}

// todosEnFoto: true si cada slice activo de la feature está en la foto de su aprobación.
func todosEnFoto(fs []backlog.Ficha, feature backlog.Ficha) bool {
	for _, f := range fs {
		if f.Padre != feature.ID || f.Tipo != "slice" || f.Estado == "CERRADO" || f.Estado == "CANCELADO" {
			continue
		}
		if _, ok := feature.Aprobacion.Foto[f.ID]; !ok {
			return false
		}
	}
	return true
}
