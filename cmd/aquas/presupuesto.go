package main

// E001/F015/S002: hooks UserPromptSubmit («aquas budget prompt») y PreToolUse («aquas budget tool»).
// Frena el gasto por feature (tope aprobado con la feature) y por turno (bucles desbocados); avisa del contexto.
// Las cuentas viven en .tmp/presupuesto-cuentas/: son estado de la sesión, no historia del repo.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/backlog"
	"github.com/jliberal-aquas/aquas-central-harness-government/internal/memoria"
)

const (
	topeFeaturePorDefecto = 300     // sin «tope N» en el apruebo de la feature
	frenoTurno            = 60      // llamadas del hilo principal por mensaje del humano
	avisoContexto         = 200_000 // tokens: desde aquí se pide /compact
)

var reTope = regexp.MustCompile(`(?i)\btope\s+(\d+)\b`)

// entradaPresupuesto: lo que el hook trae por stdin.
type entradaPresupuesto struct {
	Sesion     string `json:"session_id"`
	Prompt     string `json:"prompt"`
	Agente     string `json:"agent_id"`
	Transcript string `json:"transcript_path"`
	Tool       string `json:"tool_name"`
	Input      struct {
		Command string `json:"command"`
	} `json:"tool_input"`
}

// E001/F011/S002 (D000029): `aquas memory set` suelta, la orden que exige el cierre, no se frena ni cuenta.
// Lista blanca: el comando entero debe ser binario conocido + `memory set` (o el viejo `memoria fijar`) + campo=valor; nada se cuela detrás.
// Binario: `aquas`, `go run ./cmd/aquas` o el del repo en .tmp/ (con o sin comillas). Valor: sin espacios ni
// metacaracteres de shell, entre comillas simples, o entre dobles sin dólar, acento grave, barra invertida ni comilla doble.
// Una ruta entre comillas solo vale si es el binario de .tmp/ de este repo.
var reMemoriaFijar = regexp.MustCompile(`^\s*` +
	`(aquas(\.exe)?|go run \./cmd/aquas|(\./)?\.tmp/aquas(\.exe)?|"([^"$` + "`" + `\\]*)")` +
	`\s+(memoria\s+fijar|memory\s+set)` + // E000/F006/S006: nombre viejo y nuevo
	`(\s+[a-z_]+=([^\s"'$` + "`" + `\\;&|<>()]*|'[^']*'|"[^"$` + "`" + `\\]*"))*\s*$`)

// exentaDelFreno: true si la llamada es Bash y su comando entero es `aquas memory set campo=valor…`.
func exentaDelFreno(raiz string, e entradaPresupuesto) bool {
	m := reMemoriaFijar.FindStringSubmatch(e.Input.Command)
	if e.Tool != "Bash" || m == nil {
		return false
	}
	if !strings.HasPrefix(m[1], `"`) {
		return true
	}
	ruta := filepath.ToSlash(filepath.Clean(m[5]))
	for _, b := range []string{"aquas", "aquas.exe"} {
		if strings.EqualFold(ruta, filepath.ToSlash(filepath.Join(raiz, ".tmp", b))) || ruta == ".tmp/"+b {
			return true
		}
	}
	return false
}

// presupuestoCLI: «prompt» reinicia el turno; «herramienta» cuenta la llamada y bloquea con exit 2 al pasar un tope.
func presupuestoCLI(raiz string, args []string) {
	var e entradaPresupuesto
	datos, _ := io.ReadAll(os.Stdin)
	if err := json.Unmarshal(datos, &e); err != nil {
		fallar(1, fmt.Errorf("presupuesto: entrada del hook ilegible: %w", err))
	}
	dir := filepath.Join(raiz, ".tmp", "presupuesto-cuentas")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fallar(1, err)
	}
	sesion := regexp.MustCompile(`[^\w-]`).ReplaceAllString(e.Sesion, "_")
	turno, marca := filepath.Join(dir, "turno-"+sesion+".log"), filepath.Join(dir, "contexto-"+sesion)
	switch strings.Join(args, " ") {
	case "prompt":
		for _, m := range marcasDeAgente { // un aviso de agente no es un mensaje del humano: no da turno nuevo
			if strings.Contains(e.Prompt, m) {
				return
			}
		}
		_ = os.Remove(turno)
		_ = os.Remove(marca)
	case "herramienta":
		if exentaDelFreno(raiz, e) {
			return
		}
		if e.Agente == "" { // el turno va primero: una llamada que frena el turno no gasta tope de la feature
			if n := anotar(turno); n > frenoTurno {
				_, _ = fmt.Fprintf(os.Stderr, "Presupuesto: %d llamadas en este turno (freno %d). No sigas: informa al humano qué hiciste y qué falta.\n", n, frenoTurno)
				os.Exit(2)
			}
		}
		if msg := frenoFeature(raiz, dir); msg != "" {
			_, _ = fmt.Fprintln(os.Stderr, msg)
			os.Exit(2)
		}
		if t := contextoMedido(e.Transcript); t >= avisoContexto {
			if _, err := os.Stat(marca); err != nil {
				_ = os.WriteFile(marca, []byte(strconv.Itoa(t)), 0o644)
				msg := fmt.Sprintf("Contexto medido: %d tokens (umbral %d). Pide al humano /compact al terminar este turno.", t, avisoContexto)
				b, _ := json.Marshal(map[string]any{"hookSpecificOutput": map[string]string{"hookEventName": "PreToolUse", "additionalContext": msg}})
				fmt.Println(string(b))
			}
		}
	default:
		fallar(64, fmt.Errorf("uso: aquas budget prompt|tool"))
	}
}

// frenoFeature cuenta la llamada contra la feature del foco si está aprobada; "" si puede seguir.
// La cuenta va por feature y fecha de aprobación: aprobar de nuevo empieza desde cero.
func frenoFeature(raiz, dir string) string {
	m, err := memoria.Leer(raiz)
	if err != nil || m.Foco.Iniciativa == "" || m.Foco.Feature == "" {
		return ""
	}
	id := m.Foco.Iniciativa + "/" + m.Foco.Feature
	f, _, err := backlog.Ver(raiz, id)
	if err != nil || f.Aprobacion == nil {
		return ""
	}
	tope := topeFeaturePorDefecto
	primera, _, _ := strings.Cut(f.Aprobacion.Mensaje, "\n")
	if t := reTope.FindStringSubmatch(primera); t != nil {
		tope, _ = strconv.Atoi(t[1])
	}
	clave := regexp.MustCompile(`[^\w-]`).ReplaceAllString(id+"-"+f.Aprobacion.Fecha, "_")
	if n := anotar(filepath.Join(dir, "feature-"+clave+".log")); n > tope {
		return fmt.Sprintf("Presupuesto: llegaste al tope de la feature %s (%d llamadas). No sigas: informa qué falta y pide un tope nuevo («apruebo %s tope N»).", id, tope, id)
	}
	return ""
}

// anotar suma una línea al archivo y devuelve el total.
func anotar(ruta string) int {
	if f, err := os.OpenFile(ruta, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err == nil {
		_, _ = f.WriteString(".\n")
		_ = f.Close()
	}
	b, _ := os.ReadFile(ruta)
	return bytes.Count(b, []byte("\n"))
}

// contextoMedido: tokens de entrada de la última llamada del modelo en el transcript; 0 si después
// hubo una compactación (lo de antes ya no se reenvía) o no hay medida.
func contextoMedido(ruta string) int {
	f, err := os.Open(ruta)
	if err != nil {
		return 0
	}
	defer func() { _ = f.Close() }()
	if st, err := f.Stat(); err == nil && st.Size() > 400_000 {
		_, _ = f.Seek(st.Size()-400_000, io.SeekStart)
	}
	b, _ := io.ReadAll(f)
	lineas := strings.Split(string(b), "\n")
	for i := len(lineas) - 1; i >= 0; i-- {
		var l struct {
			Subtipo string `json:"subtype"`
			Message struct {
				Usage *struct {
					Entrada  int `json:"input_tokens"`
					Lectura  int `json:"cache_read_input_tokens"`
					Creacion int `json:"cache_creation_input_tokens"`
				} `json:"usage"`
			} `json:"message"`
		}
		if json.Unmarshal([]byte(lineas[i]), &l) != nil {
			continue
		}
		if l.Subtipo == "compact_boundary" {
			return 0
		}
		if u := l.Message.Usage; u != nil {
			return u.Entrada + u.Lectura + u.Creacion
		}
	}
	return 0
}
