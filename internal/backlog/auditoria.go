package backlog

// E001/F014/S002: el veredicto del auditor queda anotado en la ficha; CERRADO lo exige.

import (
	"fmt"
	"maps"
	"strings"
	"time"
)

// Auditoria: último veredicto del auditor de un slice. Huella: la del código Go del alcance al anotarlo.
type Auditoria struct {
	Fecha     string             `json:"fecha"`
	Veredicto string             `json:"veredicto"`
	Huella    *map[string]string `json:"huella,omitempty"`
}

// AnotarAuditoria anota el veredicto del auditor en la ficha de un slice EN_CURSO; gana el último.
func AnotarAuditoria(raiz, id, veredicto string, ahora time.Time) error {
	f, err := activa(raiz, id)
	if err != nil {
		return fmt.Errorf("no se anota el veredicto de %s: %w", id, err)
	}
	if f.Tipo != "slice" || f.Estado != "EN_CURSO" {
		return fmt.Errorf("no se anota el veredicto de %s: debe ser un slice EN_CURSO y es %s %s", id, f.Tipo, f.Estado)
	}
	h, err := huellaVerde(raiz, f)
	if err != nil {
		return fmt.Errorf("no se anota el veredicto de %s: %w", id, err)
	}
	f.Auditoria = &Auditoria{Fecha: ahora.UTC().Format(time.RFC3339), Veredicto: veredicto, Huella: h}
	if err := escribir(raiz, Activo, f); err != nil {
		return fmt.Errorf("no se anota el veredicto de %s: %w", id, err)
	}
	return nil
}

// veredictoVigente: CERRADO de un slice exige el último veredicto APROBADO, posterior al verde y sobre el mismo código.
func veredictoVigente(f Ficha) error {
	a := f.Auditoria
	if a == nil {
		return fmt.Errorf("%s no cierra: no hay veredicto del auditor; lanza una auditoría (cierre con id: %s-A1)", f.ID, strings.ReplaceAll(f.ID, "/", "-"))
	}
	if a.Veredicto != "APROBADO" {
		return fmt.Errorf("%s no cierra: el último veredicto es %s, no APROBADO; corrige y lanza otra auditoría", f.ID, a.Veredicto)
	}
	fa, errA := time.Parse(time.RFC3339, a.Fecha)
	fv, errV := time.Parse(time.RFC3339, f.Verde.Fecha)
	if errV != nil {
		return fmt.Errorf("%s no cierra: fecha del verde ilegible (%q), no se puede comparar el veredicto; corre aquas backlog verde %s", f.ID, f.Verde.Fecha, f.ID)
	}
	if errA != nil {
		return fmt.Errorf("%s no cierra: fecha del veredicto ilegible (%q); lanza otra auditoría", f.ID, a.Fecha)
	}
	if !fa.After(fv) {
		return fmt.Errorf("%s no cierra: el veredicto APROBADO (%s) es anterior al verde (%s); lanza otra auditoría", f.ID, a.Fecha, f.Verde.Fecha)
	}
	if f.Verde.Huella != nil && !maps.Equal(*f.Verde.Huella, derefHuella(a.Huella)) {
		return fmt.Errorf("%s no cierra: el veredicto APROBADO se dio sobre otro código que el del verde; lanza otra auditoría", f.ID)
	}
	return nil
}

// derefHuella: la huella apuntada, o nil si no hay.
func derefHuella(h *map[string]string) map[string]string {
	if h == nil {
		return nil
	}
	return *h
}
