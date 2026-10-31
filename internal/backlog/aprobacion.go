package backlog

// E001/F010/S001: aprobación humana del contrato de un slice, atada a su huella. El rojo y el paso
// a EN_CURSO exigen una aprobación vigente (desde cualquier estado, mientras no haya rojo); rojo, verde y cierre
// comparan la huella con la aprobada. Solo el hook del mensaje del humano (`aquas approval`) registra aprobaciones.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

// Aprobacion: el humano aprobó el contrato de la ficha; Huella sella pruebas, alcance y «Para qué existe».
type Aprobacion struct {
	Fecha   string `json:"fecha"`
	Huella  string `json:"huella"`
	Mensaje string `json:"mensaje"`
	// E001/F015/S001: solo en features. Push: el humano escribió «con push». Foto: pruebas aprobadas de cada slice.
	Push bool                `json:"push,omitempty"`
	Foto map[string][]Prueba `json:"foto,omitempty"`
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

// Aprobar registra la aprobación del contrato de un slice o de una feature (con la foto de sus slices) con la huella
// actual. Lo llama solo el hook del mensaje del humano.
func Aprobar(raiz, id, mensaje string, ahora time.Time) (Ficha, error) {
	f, err := activa(raiz, id)
	if err != nil {
		return Ficha{}, err
	}
	var push bool
	var foto map[string][]Prueba
	switch f.Tipo {
	case "slice":
	case "feature": // E001/F015/S001: la aprobación de la feature guarda la foto de sus slices activos
		act, _, err := todas(raiz)
		if err != nil {
			return Ficha{}, err
		}
		foto = map[string][]Prueba{}
		for _, s := range act {
			if s.Tipo == "slice" && s.Padre == f.ID {
				foto[s.ID] = s.Pruebas
			}
		}
		if len(foto) == 0 {
			return Ficha{}, fmt.Errorf("%s: la feature no tiene slices activos; no hay nada que aprobar", id)
		}
		primera, _, _ := strings.Cut(strings.TrimSpace(mensaje), "\n")
		push = reConPush.MatchString(primera)
	default:
		return Ficha{}, fmt.Errorf("%s: se aprueban slices o features", id)
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
	f.Aprobacion = &Aprobacion{Fecha: ahora.UTC().Format(time.RFC3339), Huella: h, Mensaje: mensaje, Push: push, Foto: foto}
	return f, escribir(raiz, Activo, f)
}

// aprobada: nil si el slice puede avanzar (EN_CURSO, rojo, verde, cierre); si no, el paso que falta.
// Sin aprobación: se rechaza mientras la ficha no tenga rojo (desde cualquier estado). Con rojo y sin aprobación
// es un slice anterior a F010: exento. Con aprobación: la huella actual debe ser la aprobada, también en EN_CURSO.
// E001/F015: vale primero la aprobación propia vigente; si no, la foto de la feature aprobada (cubiertaPorFeature).
// Protege contra errores y deriva del contrato, no contra un agente que quiera engañar (para eso, auditor y diff).
func aprobada(raiz string, f Ficha) error {
	if f.Tipo != "slice" {
		return nil
	}
	if f.Aprobacion != nil { // E001/F015/S001: la aprobación propia vigente vale primero
		if h, err := Huella(raiz, f); err == nil && h == f.Aprobacion.Huella {
			return nil
		}
	}
	if p, err := activa(raiz, f.Padre); err == nil && p.Aprobacion != nil { // E001/F015/S001: si no, la foto de la feature
		if aprobadas, ok := p.Aprobacion.Foto[f.ID]; ok {
			return cubiertaPorFeature(raiz, p, f, aprobadas)
		}
	}
	if f.Aprobacion == nil {
		if f.Rojo != nil {
			return nil
		}
		if n := abiertos(raiz, f.Padre); n >= 2 { // E001/F015/X01 (D000069): una aprobación de feature cubre a todos
			return fmt.Errorf("%s sin aprobación del contrato: su feature tiene %d slices abiertos; pide una sola vez «apruebo %s con push» (cubre todos; el hook la registra)", f.ID, n, f.Padre)
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

// abiertos (E001/F015/X01): slices activos (ni CERRADO ni CANCELADO) de la feature; error de lectura = 0.
func abiertos(raiz, feature string) int {
	act, _, err := todas(raiz)
	if err != nil {
		return 0
	}
	n := 0
	for _, f := range act {
		if f.Tipo == "slice" && f.Padre == feature {
			n++
		}
	}
	return n
}

// reConPush: la primera línea del mensaje es «apruebo <id> …» y dice «con push» en cualquier lugar tras el id,
// también después de «tope N» (E001/F015/S001, S003).
var reConPush = regexp.MustCompile(`(?i)^\s*apruebo\s+\S+\s.*\bcon\s+push\b`)

// cubiertaPorFeature (E001/F015/S001): el slice avanza si el contrato de su feature no cambió y cada prueba
// aprobada sigue idéntica. Agregar pruebas o cambiar el alcance del slice no invalida la aprobación.
func cubiertaPorFeature(raiz string, p, f Ficha, aprobadas []Prueba) error {
	h, err := Huella(raiz, p)
	if err != nil {
		return err
	}
	if h != p.Aprobacion.Huella {
		return fmt.Errorf("el contrato de la feature %s cambió desde su aprobación: el humano debe escribir «apruebo %s» otra vez", p.ID, p.ID)
	}
	for _, a := range aprobadas {
		if !slices.Contains(f.Pruebas, a) {
			return fmt.Errorf("%s: la prueba aprobada %s cambió o ya no está: el humano debe escribir «apruebo %s» otra vez", f.ID, a.Nombre, p.ID)
		}
	}
	return nil
}
