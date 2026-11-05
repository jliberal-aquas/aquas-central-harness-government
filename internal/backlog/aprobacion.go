package backlog

// E001/F010/S001: aprobación humana del contrato de un slice, atada a su huella. El rojo y el paso
// a EN_CURSO exigen una aprobación vigente (desde cualquier estado); rojo, verde y cierre
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
	// E000/F008/S002: transcript_path de la sesión donde el humano escribió el mensaje; aprobada lo comprueba.
	Registro string `json:"registro,omitempty"`
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

// MarcaPendiente: relleno de la página nacida (E000/F014/S001); el mismo texto que internal/hijo usa en sus huecos.
const MarcaPendiente = "PENDIENTE:"

// reCeldaPendiente: celda de tabla Markdown cuyo texto empieza con la marca: una fila sin decidir.
var reCeldaPendiente = regexp.MustCompile(`\|\s*` + MarcaPendiente)

// contratoCompleto (E000/F014/S001, hallazgo A1): la página del slice existe y su «Para qué existe» no trae
// filas PENDIENTE; si no, el humano aprobaría un contrato sin decidir.
func contratoCompleto(raiz, id string) error {
	rel := pagina(id)
	b, err := os.ReadFile(filepath.Join(raiz, filepath.FromSlash(rel)))
	if err != nil {
		return fmt.Errorf("%s: falta la página del contrato %s", id, rel)
	}
	sec, ok := seccionContrato(string(b))
	if !ok {
		return fmt.Errorf("%s: %s no tiene la sección «## Para qué existe»", id, rel)
	}
	if reCeldaPendiente.MatchString(sec) { // solo una celda que empieza con la marca; mencionarla en prosa no frena (A2)
		return fmt.Errorf("%s: el contrato de %s tiene filas %s: complétalo antes del apruebo", id, rel, MarcaPendiente)
	}
	return nil
}

// Aprobar registra la aprobación del contrato de un slice o de una feature (con la foto de sus slices) con la huella
// actual. Lo llama solo el hook del mensaje del humano.
func Aprobar(raiz, id, mensaje, registro string, ahora time.Time) (Ficha, error) {
	f, err := activa(raiz, id)
	if err != nil {
		return Ficha{}, err
	}
	var push bool
	var foto map[string][]Prueba
	switch f.Tipo {
	case "slice":
		if err := contratoCompleto(raiz, f.ID); err != nil {
			return Ficha{}, err
		}
	case "feature": // E001/F015/S001: la aprobación de la feature guarda la foto de sus slices activos
		act, _, err := todas(raiz)
		if err != nil {
			return Ficha{}, err
		}
		foto = map[string][]Prueba{}
		for _, s := range act {
			if s.Tipo == "slice" && s.Padre == f.ID {
				if err := contratoCompleto(raiz, s.ID); err != nil { // la feature cubre solo slices con contrato
					return Ficha{}, err
				}
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
	f.Aprobacion = &Aprobacion{Fecha: ahora.UTC().Format(time.RFC3339), Huella: h, Mensaje: mensaje, Registro: registro, Push: push, Foto: foto}
	return f, escribir(raiz, Activo, f)
}

// aprobada: nil si el slice puede avanzar (EN_CURSO, rojo, verde, cierre); si no, el paso que falta.
// Sin aprobación se rechaza desde cualquier estado, también con rojo previo (E000/F008/S002: ese rojo se finge a mano).
// Con aprobación: la huella actual debe ser la aprobada y su mensaje humano debe estar en el registro (respaldada).
// E001/F015: vale primero la aprobación propia vigente; si no, la foto de la feature aprobada (cubiertaPorFeature).
func aprobada(raiz string, f Ficha) error {
	if f.Tipo != "slice" {
		return nil
	}
	var sinRespaldo error
	if f.Aprobacion != nil { // E001/F015/S001: la aprobación propia vigente vale primero
		if h, err := Huella(raiz, f); err == nil && h == f.Aprobacion.Huella {
			if sinRespaldo = respaldada(raiz, f.ID, f.Aprobacion); sinRespaldo == nil {
				return nil
			}
		}
	}
	if p, err := activa(raiz, f.Padre); err == nil && p.Aprobacion != nil { // E001/F015/S001: si no, la foto de la feature
		if aprobadas, ok := p.Aprobacion.Foto[f.ID]; ok {
			if err := respaldada(raiz, p.ID, p.Aprobacion); err != nil {
				return err
			}
			return cubiertaPorFeature(raiz, p, f, aprobadas)
		}
		// E000/F008/S006 (D000084): fuera de la foto de una feature aprobada se pide la feature, no el slice.
		return fmt.Errorf("%s no está en la aprobación de %s (se creó después): el humano debe escribir %s otra vez (cubre todos sus slices; el hook la registra)", f.ID, p.ID, pideFeature(p))
	}
	if sinRespaldo != nil {
		return sinRespaldo
	}
	if f.Aprobacion == nil {
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

// AvisoFeatureAprobada: aviso al crear un slice cuya feature ya está aprobada; "" si no aplica.
// E000/F008/S006 (D000084): la creación no se frena; el slice queda fuera de la foto hasta el apruebo nuevo.
func AvisoFeatureAprobada(raiz string, f Ficha) string {
	if f.Tipo != "slice" {
		return ""
	}
	p, err := activa(raiz, f.Padre)
	if err != nil || p.Aprobacion == nil {
		return ""
	}
	return fmt.Sprintf("Aviso: %s ya está aprobada; %s necesitará %s otra vez.\n", p.ID, f.ID, pideFeature(p))
}

// pideFeature: el apruebo de la feature que hay que pedir, con «con push» si la aprobación anterior lo tenía.
func pideFeature(p Ficha) string {
	if p.Aprobacion != nil && p.Aprobacion.Push {
		return fmt.Sprintf("«apruebo %s con push»", p.ID)
	}
	return fmt.Sprintf("«apruebo %s»", p.ID)
}
