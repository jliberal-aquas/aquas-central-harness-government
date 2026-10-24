package backlog

// E001/F010/S001: aprobación humana del contrato de un slice, atada a su huella. El rojo y el paso
// a EN_CURSO exigen una aprobación vigente (desde cualquier estado, mientras no haya rojo); rojo, verde y cierre
// comparan la huella con la aprobada. Solo el hook del mensaje del humano (`aquas aprobacion`) registra aprobaciones.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"
)

// Aprobacion: el humano aprobó el contrato de la ficha; Huella sella pruebas, alcance y «Para qué existe».
type Aprobacion struct {
	Fecha   string `json:"fecha"`
	Huella  string `json:"huella"`
	Mensaje string `json:"mensaje"`
}

// seccionContrato: texto de «## Para qué existe» de la página, sin los bloques GENERADO de aquas.
func seccionContrato(texto string) (string, bool) {
	var b strings.Builder
	dentro, generado, hallada := false, false, false
	for _, l := range strings.Split(strings.ReplaceAll(texto, "\r\n", "\n"), "\n") {
		t := strings.TrimSpace(l)
		switch {
		case strings.HasPrefix(t, "## "):
			if dentro {
				return strings.TrimSpace(b.String()), true
			}
			dentro = t == "## Para qué existe"
			hallada = hallada || dentro
		case !dentro:
		case strings.HasPrefix(t, "<!-- GENERADO"):
			generado = true
		case strings.HasPrefix(t, "<!-- /GENERADO"):
			generado = false
		case !generado:
			b.WriteString(strings.TrimRight(l, " \t") + "\n")
		}
	}
	return strings.TrimSpace(b.String()), hallada
}

// Huella: SHA-256 de las pruebas, el alcance (con sus ampliaciones) y la sección «Para qué existe» de la página.
func Huella(raiz string, f Ficha) (string, error) {
	rel := pagina(f.ID)
	b, err := os.ReadFile(filepath.Join(raiz, filepath.FromSlash(rel)))
	if err != nil {
		return "", fmt.Errorf("%s: falta la página del contrato %s", f.ID, rel)
	}
	sec, ok := seccionContrato(string(b))
	if !ok {
		return "", fmt.Errorf("%s: %s no tiene la sección «## Para qué existe»", f.ID, rel)
	}
	ps, _ := json.Marshal(f.Pruebas)
	al, _ := json.Marshal(f.Alcance)
	am, _ := json.Marshal(f.Ampliaciones)
	h := sha256.Sum256([]byte("pruebas\n" + string(ps) + "\nalcance\n" + string(al) + "\nampliaciones\n" + string(am) + "\npara qué existe\n" + sec))
	return hex.EncodeToString(h[:]), nil
}

// Aprobar registra la aprobación del contrato del slice con la huella actual. Lo llama solo el hook del mensaje del humano.
func Aprobar(raiz, id, mensaje string, ahora time.Time) (Ficha, error) {
	f, err := activa(raiz, id)
	if err != nil {
		return Ficha{}, err
	}
	if f.Tipo != "slice" {
		return Ficha{}, fmt.Errorf("%s: la aprobación es de slices", id)
	}
	h, err := Huella(raiz, f)
	if err != nil {
		return Ficha{}, err
	}
	if m := strings.TrimSpace(mensaje); utf8.RuneCountInString(m) > 300 {
		mensaje = string([]rune(m)[:300])
	} else {
		mensaje = m
	}
	f.Aprobacion = &Aprobacion{Fecha: ahora.UTC().Format(time.RFC3339), Huella: h, Mensaje: mensaje}
	return f, escribir(raiz, Activo, f)
}

// aprobada: nil si el slice puede avanzar (EN_CURSO, rojo, verde, cierre); si no, el paso que falta.
// Sin aprobación: se rechaza mientras la ficha no tenga rojo (desde cualquier estado). Con rojo y sin aprobación
// es un slice anterior a F010: exento. Con aprobación: la huella actual debe ser la aprobada, también en EN_CURSO.
// Protege contra errores y deriva del contrato, no contra un agente que quiera engañar (para eso, auditor y diff).
func aprobada(raiz string, f Ficha) error {
	if f.Tipo != "slice" {
		return nil
	}
	if f.Aprobacion == nil {
		if f.Rojo != nil {
			return nil
		}
		return fmt.Errorf("%s sin aprobación del contrato: el humano debe escribir «apruebo %s» en el chat (el hook la registra)", f.ID, f.ID)
	}
	h, err := Huella(raiz, f)
	if err != nil {
		return err
	}
	if h != f.Aprobacion.Huella {
		return fmt.Errorf("el contrato de %s cambió desde su aprobación (pruebas, alcance o «Para qué existe»): el humano debe escribir «apruebo %s» otra vez", f.ID, f.ID)
	}
	return nil
}
