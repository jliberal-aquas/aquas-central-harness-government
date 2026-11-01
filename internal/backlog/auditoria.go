package backlog

// E001/F014/S002: el veredicto del auditor queda anotado en la ficha; CERRADO lo exige.

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"strings"
	"time"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/encargo"
)

// Auditoria: último veredicto del auditor de un slice. Huella: la del código Go del alcance al anotarlo.
type Auditoria struct {
	Fecha     string             `json:"fecha"`
	Veredicto string             `json:"veredicto"`
	Huella    *map[string]string `json:"huella,omitempty"`
	Registro  string             `json:"registro,omitempty"` // E000/F008/S004: agent_transcript_path del auditor
}

// AnotarAuditoria anota el veredicto del auditor y su registro en la ficha de un slice EN_CURSO; gana el último.
func AnotarAuditoria(raiz, id, veredicto, registro string, ahora time.Time) error {
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
	f.Auditoria = &Auditoria{Fecha: ahora.UTC().Format(time.RFC3339), Veredicto: veredicto, Huella: h, Registro: registro}
	if err := escribir(raiz, Activo, f); err != nil {
		return fmt.Errorf("no se anota el veredicto de %s: %w", id, err)
	}
	return nil
}

// veredictoVigente: CERRADO de un slice exige el último veredicto APROBADO, posterior al verde y sobre el mismo código.
func veredictoVigente(raiz string, f Ficha) error {
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
		return fmt.Errorf("%s no cierra: fecha del verde ilegible (%q), no se puede comparar el veredicto; corre aquas backlog green %s", f.ID, f.Verde.Fecha, f.ID)
	}
	if errA != nil {
		return fmt.Errorf("%s no cierra: fecha del veredicto ilegible (%q); lanza otra auditoría", f.ID, a.Fecha)
	}
	if !fa.After(fv) {
		return fmt.Errorf("%s no cierra: el veredicto APROBADO (%s) es anterior al verde (%s); lanza otra auditoría", f.ID, a.Fecha, f.Verde.Fecha)
	}
	if err := veredictoRespaldado(raiz, f.ID, a.Registro, fv); err != nil {
		return fmt.Errorf("%s no cierra: %w; lanza otra auditoría", f.ID, err)
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

// veredictoRespaldado (E000/F008/S004, D000077): el último cierre de ese slice que el auditor escribió en su
// registro dice APROBADO y es posterior al verde. El texto del orquestador, de otro agente o del encargo no cuenta.
func veredictoRespaldado(raiz, id, registro string, verde time.Time) error {
	if registro == "" {
		return errors.New("el veredicto no trae registro del auditor")
	}
	reg, err := registroValido(raiz, registro)
	if err != nil {
		return fmt.Errorf("el registro del auditor %v", err)
	}
	v, cuando, err := ultimoVeredicto(reg, id)
	switch {
	case err != nil:
		return fmt.Errorf("el registro del auditor no se puede leer: %v", err)
	case v == "":
		return fmt.Errorf("el registro del auditor no trae un veredicto de %s escrito por el auditor", id)
	case v != "APROBADO":
		return fmt.Errorf("el último veredicto de %s en el registro del auditor es %s", id, v)
	case !cuando.After(verde):
		return fmt.Errorf("el APROBADO del registro del auditor (%s) es anterior al verde", cuando.Format(time.RFC3339))
	}
	return nil
}

// ultimoVeredicto: veredicto y hora del último cierre de ese id en las líneas de asistente del auditor
// (texto o mensaje de SubagentHandback; forma observada el 2026-10-31). "" si no hay ninguno.
func ultimoVeredicto(ruta, id string) (string, time.Time, error) {
	fh, err := os.Open(ruta)
	if err != nil {
		return "", time.Time{}, err
	}
	defer func() { _ = fh.Close() }()
	var v string
	var cuando time.Time
	rd := bufio.NewReader(fh)
	for {
		l, errL := rd.ReadBytes('\n')
		var lr struct {
			Type      string `json:"type"`
			Agente    string `json:"attributionAgent"`
			Timestamp string `json:"timestamp"`
			Message   struct {
				Content []json.RawMessage `json:"content"`
			} `json:"message"`
		}
		if len(l) > 0 && json.Unmarshal(l, &lr) == nil && lr.Type == "assistant" && lr.Agente == "auditor" {
			ts, errT := time.Parse(time.RFC3339Nano, lr.Timestamp)
			for _, c := range lr.Message.Content {
				var p struct {
					Type  string `json:"type"`
					Text  string `json:"text"`
					Name  string `json:"name"`
					Input struct {
						Message string `json:"message"`
					} `json:"input"`
				}
				if json.Unmarshal(c, &p) != nil {
					continue
				}
				texto := p.Text
				if p.Type == "tool_use" && p.Name == "SubagentHandback" {
					texto = p.Input.Message
				} else if p.Type != "text" {
					continue
				}
				if got, ver, ok := encargo.Veredicto(texto); ok && got == id && errT == nil {
					v, cuando = ver, ts
				}
			}
		}
		if errors.Is(errL, io.EOF) {
			return v, cuando, nil
		}
		if errL != nil {
			return "", time.Time{}, errL
		}
	}
}
