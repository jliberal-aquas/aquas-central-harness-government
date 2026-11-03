package main

import "testing"

// TestRetrospectivaSinReincidencia (E001/F017): cada familia de la retrospectiva del 2026-11-02 tiene su control,
// con un caso que debe fallar y uno que debe pasar. Crece con cada slice cerrado.
// Rojo cuando: una familia con slice cerrado vuelve a pasar sin freno.
// Verde cuando: el control de cada slice cerrado frena su caso malo y deja pasar el bueno.
func TestRetrospectivaSinReincidencia(t *testing.T) {
	t.Run("S001 freno atado al trabajo activo", TestFrenoSigueAlTrabajoActivo)
}
