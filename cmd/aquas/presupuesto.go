package main

// E001/F015/S002: hook PreToolUse («aquas budget tool»). Frena el gasto por feature (tope aprobado con la
// feature) y por turno (bucles desbocados); avisa del contexto.
// E000/F008/S005 (D000077, D000082): las llamadas, la feature, su tope y su hora salen del registro que escribe
// Claude Code fuera del repo; nada sale de contadores, foco ni fichas, que el agente puede tocar. Única excepción
// (E001/F017/S001): suelta el tope una feature CERRADO con todos sus slices CERRADO, y cada uno exigió auditoría
// APROBADO. Determinista solo con el aislamiento de Bash (S003); en Windows nativo es una segunda barrera,
// probabilística (D000079).

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/backlog"
)

const (
	topeTokensPorDefecto = 15_000_000 // E001/F019/S001: tokens ponderados sin «tope» en el apruebo de la feature
	frenoTurno           = 60         // llamadas del hilo principal por mensaje del humano
	avisoContexto        = 200_000    // tokens: desde aquí se pide /compact
)

// reTope: «tope» y las dos palabras que lo siguen; leerTope decide qué valen (E001/F019/S001).
var reTope = regexp.MustCompile(`(?i)\btope\s+(\S+)(?:\s+(\S+))?`)

var (
	reSoloDigitos = regexp.MustCompile(`^\d+$`)
	reMillones    = regexp.MustCompile(`(?i)^(\d+)m$`)
)

// leerTope: «tope 25M» o «tope 25 M» → tokens ponderados; «tope 900» solo, o seguido de «con» o «y» → llamadas (el viejo).
// Hallazgos A1.1, A2.1 y A3.1: cualquier otra forma («1,5M», «25k», «25MM», «15Mtok», «25 millones», «20000000 tokens»,
// «15 000 000», 0, desbordado) no vale y deja el de 15M; nunca cae a llamadas por adivinar. Una coma o punto final se ignora.
func leerTope(primera string, ap *apruebo) {
	t := reTope.FindStringSubmatch(primera)
	if t == nil {
		return
	}
	palabra := strings.TrimRight(t[1], ".,;:)")
	siguiente := strings.ToLower(strings.TrimRight(t[2], ".,;:)"))
	if m := reMillones.FindStringSubmatch(palabra); m != nil {
		palabra, siguiente = m[1], "m"
	} else if !reSoloDigitos.MatchString(palabra) {
		return
	}
	n, err := strconv.Atoi(palabra)
	switch {
	case err != nil || n <= 0:
	case siguiente == "m":
		if n <= maxTopeMillones {
			ap.tokens = int64(n) * 1_000_000
		}
	case siguiente == "" || siguiente == "con" || siguiente == "y": // hallazgo A3.1: llamadas solo con la gramática del apruebo
		ap.tope, ap.tokens = n, 0
	}
	// Cualquier otra palabra detrás («tokens», «000 000», «k», «millones», «más»): no se adivina; rige el de 15M.
}

// maxTopeMillones (hallazgo A1.6): más que esto desborda la cuenta en veinteavos; no vale y rige el de 15M.
const maxTopeMillones = 1_000_000

// pesoVeinteavos: el usage de una respuesta ponderado por precio, en veinteavos de token para no usar decimales.
// Pesos (tabla de precios oficial, fila Claude Opus 5.5, consultada 2026-11-03,
// https://platform.claude.com/docs/en/about-claude/pricing): entrada 1, escritura de caché 2 (1 h),
// lectura de caché 0,05, salida 5.
func pesoVeinteavos(u usoRespuesta) int64 {
	return 20*u.Entrada + 40*u.Escritura + u.Lectura + 100*u.Salida
}

// usoRespuesta: el usage de una respuesta del modelo en el registro.
type usoRespuesta struct {
	Entrada   int64 `json:"input_tokens"`
	Escritura int64 `json:"cache_creation_input_tokens"`
	Lectura   int64 `json:"cache_read_input_tokens"`
	Salida    int64 `json:"output_tokens"`
}

// gasto: el peso de una respuesta y su hora.
type gasto struct {
	hora time.Time
	peso int64
}

// millones: 15000000 → «15M»; 15000001 → «15,0M».
func millones(n int64) string {
	if n%1_000_000 == 0 {
		return fmt.Sprintf("%dM", n/1_000_000)
	}
	return strings.Replace(fmt.Sprintf("%.1fM", float64(n)/1e6), ".", ",", 1)
}

// entradaPresupuesto: lo que el hook trae por stdin.
type entradaPresupuesto struct {
	Sesion     string `json:"session_id"`
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
// E001/F017/S001 (D000101): también `aquas defect new|occur`, argumentos sin «=» y un `cd <raíz> &&` delante;
// el `cd` solo vale si va a la raíz de este repo.
// Entre palabras solo espacio o tabulador: un salto de línea nunca separa (D000105, hallazgo A1.3).
var reMemoriaFijar = regexp.MustCompile(`^[ \t]*` +
	`(cd[ \t]+("([^"$` + "`" + `\\]*)"|'([^']*)'|([^\s"'$` + "`" + `\\;&|<>()]+))[ \t]*&&[ \t]*)?` +
	`(aquas(\.exe)?|go run \./cmd/aquas|(\./)?\.tmp/aquas(\.exe)?|"([^"$` + "`" + `\\]*)")` +
	`[ \t]+(memoria[ \t]+fijar|memory[ \t]+set|defecto[ \t]+(nuevo|ocurre)|defect[ \t]+(new|occur))` + // E000/F006/S006: nombre viejo y nuevo
	`([ \t]+(?:[^\s"'$` + "`" + `\\;&|<>()]+|'[^']*'|"[^"$` + "`" + `\\]*")+)*[ \t]*\n?$`) // cada argumento: trozos sueltos o entre comillas, pegados

// rutaComparableEn: la ruta limpia con «/»; en Windows la forma Git Bash «/d/…» pasa a «D:/…» (A2.4; sistema dado: E000/F007/S001).
// Conserva la raíz de unidad «D:/» (A1.1), no confunde «D:» a secas con ella (A1.2) y mantiene la doble barra UNC (A1.3).
func rutaComparableEn(goos, r string) string {
	r = strings.ReplaceAll(r, `\`, "/")
	unc := strings.HasPrefix(r, "//")
	raizUnidad := len(r) >= 3 && r[1] == ':' && r[2] == '/'
	r = path.Clean(r)
	if unc {
		r = "/" + r
	}
	if raizUnidad && len(r) == 2 {
		r += "/"
	}
	if goos == "windows" && len(r) >= 2 && r[0] == '/' && unicode.IsLetter(rune(r[1])) && (len(r) == 2 || r[2] == '/') {
		r = strings.ToUpper(r[1:2]) + ":" + r[2:]
		if len(r) == 2 {
			r += "/"
		}
	}
	return r
}

// exentaDelFreno: true si el comando Bash entero es `aquas memory set …` o `aquas defect new|occur …`.
// Puede llevar delante un `cd` a la raíz de este repo, y a ninguna otra carpeta.
func exentaDelFreno(raiz, herramienta, comando string) bool {
	return exentaDelFrenoEn(runtime.GOOS, raiz, herramienta, comando)
}

// rutasIguales (E000/F007/S001, hallazgos L1 y L2; no confundir con entorno.mismaRuta, que no entiende «/d/…»): misma carpeta según el sistema; mayúsculas iguales solo en Windows.
func rutasIguales(goos, a, b string) bool {
	a, b = rutaComparableEn(goos, a), rutaComparableEn(goos, b)
	if goos == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}

// exentaDelFrenoEn: exentaDelFreno para un sistema dado (E000/F007/S001).
func exentaDelFrenoEn(goos, raiz, herramienta, comando string) bool {
	m := reMemoriaFijar.FindStringSubmatch(comando)
	if herramienta != "Bash" || m == nil {
		return false
	}
	if m[1] != "" && !rutasIguales(goos, m[3]+m[4]+m[5], raiz) {
		return false
	}
	if !strings.HasPrefix(m[6], `"`) {
		return true
	}
	ruta := path.Clean(strings.ReplaceAll(m[10], `\`, "/"))
	for _, b := range []string{"aquas", "aquas.exe"} {
		if rutasIguales(goos, ruta, raiz+"/.tmp/"+b) || ruta == ".tmp/"+b {
			return true
		}
	}
	return false
}

// presupuestoCLI: «herramienta» cuenta el gasto en el registro (tokens ponderados o llamadas, E001/F019/S001) y bloquea con exit 2 al pasar un tope.
func presupuestoCLI(raiz string, args []string) {
	if strings.Join(args, " ") != "herramienta" {
		fallar(64, fmt.Errorf("uso: aquas budget tool"))
	}
	var e entradaPresupuesto
	datos, _ := io.ReadAll(os.Stdin)
	if err := json.Unmarshal(datos, &e); err != nil {
		fallar(1, fmt.Errorf("presupuesto: entrada del hook ilegible: %w", err))
	}
	if exentaDelFreno(raiz, e.Tool, e.Input.Command) {
		return
	}
	sesiones, err := escanearProyecto(raiz, e.Transcript)
	if err != nil {
		frenar("Presupuesto: %v. Sin registro no hay cuenta: informa al humano.", err)
	}
	// E001/F017/S009 (D000109): con una feature aprobada y abierta solo cuenta su tope; el freno del turno cede.
	msgFeature, vigente := frenoFeature(raiz, e.Transcript, sesiones)
	turno := ""
	if e.Agente == "" { // el freno del turno es del hilo principal
		propia, ok := sesiones[filepath.Clean(e.Transcript)]
		if !ok { // registro fuera de la carpeta leída: se lee aparte
			if propia, err = escanear(raiz, e.Transcript); err != nil {
				frenar("Presupuesto: no se puede leer el registro de la sesión %s: %v.", e.Transcript, err)
			}
		}
		if n := len(propia.turno); n > frenoTurno && !vigente {
			frenar("Presupuesto: %d llamadas en este turno (freno %d). No sigas: informa al humano qué hiciste y qué falta.", n, frenoTurno)
		}
		turno = strconv.FormatInt(propia.inicio, 10)
	}
	if msgFeature != "" {
		frenar("%s", msgFeature)
	}
	avisarContexto(raiz, e, turno)
}

// frenar escribe el motivo en stderr y bloquea la llamada.
func frenar(formato string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, formato+"\n", a...)
	os.Exit(2)
}

// avisarContexto pide /compact una vez por turno cuando el contexto medido pasa el umbral; la marca guarda el
// turno avisado (es solo un aviso: borrarla repite el aviso, no quita un freno).
func avisarContexto(raiz string, e entradaPresupuesto, turno string) {
	t := contextoMedido(e.Transcript)
	if t < avisoContexto {
		return
	}
	dir := filepath.Join(raiz, ".tmp", "presupuesto-cuentas")
	marca := filepath.Join(dir, "contexto-"+regexp.MustCompile(`[^\w-]`).ReplaceAllString(e.Sesion, "_"))
	if b, err := os.ReadFile(marca); err == nil && string(b) == turno {
		return
	}
	_ = os.MkdirAll(dir, 0o755)
	_ = os.WriteFile(marca, []byte(turno), 0o644)
	msg := fmt.Sprintf("Contexto medido: %d tokens (umbral %d). Pide al humano /compact al terminar este turno.", t, avisoContexto)
	b, _ := json.Marshal(map[string]any{"hookSpecificOutput": map[string]string{"hookEventName": "PreToolUse", "additionalContext": msg}})
	fmt.Println(string(b))
}

// lineaLlamadas: lo que se lee de cada línea del registro (forma observada el 2026-10-31).
type lineaLlamadas struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Message   struct {
		ID      string          `json:"id"`
		Content json.RawMessage `json:"content"`
		Usage   *usoRespuesta   `json:"usage"`
	} `json:"message"`
	Origin struct {
		Kind string `json:"kind"`
	} `json:"origin"`
}

// bloqueRegistro: un bloque del contenido de un mensaje.
type bloqueRegistro struct {
	Type  string `json:"type"`
	ID    string `json:"id"`
	Name  string `json:"name"`
	Text  string `json:"text"`
	Input struct {
		Command string `json:"command"`
	} `json:"input"`
}

// recorrerRegistro abre un registro y entrega cada línea legible con su posición; las rotas se ignoran.
// Solo se decodifican las líneas que traen una llamada, un mensaje humano o un usage (E001/F019/S001).
func recorrerRegistro(ruta string, f func(pos int64, l lineaLlamadas)) error {
	fh, err := os.Open(ruta)
	if err != nil {
		return err
	}
	defer func() { _ = fh.Close() }()
	rd := bufio.NewReader(fh)
	var pos int64
	for {
		b, err := rd.ReadBytes('\n')
		if len(b) > 0 && (bytes.Contains(b, []byte(`"tool_use"`)) || bytes.Contains(b, []byte(`"human"`)) || bytes.Contains(b, []byte(`"usage"`))) {
			var l lineaLlamadas
			if json.Unmarshal(b, &l) == nil {
				f(pos, l)
			}
		}
		pos += int64(len(b))
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

// llamadas: ids de los tool_use de una línea de asistente, sin las llamadas exentas.
func llamadas(raiz string, l lineaLlamadas) []string {
	if l.Type != "assistant" {
		return nil
	}
	var bloques []bloqueRegistro
	if json.Unmarshal(l.Message.Content, &bloques) != nil {
		return nil
	}
	var ids []string
	for _, b := range bloques {
		if b.Type == "tool_use" && b.ID != "" && !exentaDelFreno(raiz, b.Name, b.Input.Command) {
			ids = append(ids, b.ID)
		}
	}
	return ids
}

// apruebo: un «apruebo» de feature escrito por el humano en el registro.
// tokens > 0: el tope es en tokens ponderados; si no, tope cuenta llamadas («tope N» viejo).
type apruebo struct {
	feature string
	tope    int
	tokens  int64
	hora    time.Time
}

// escaneo: lo que deja un registro: cada llamada con su hora, las del turno abierto, la hora (UnixNano) en que empieza ese turno,
// el último apruebo de feature y el gasto de cada respuesta por message.id (E001/F019/S001).
type escaneo struct {
	llamadas map[string]time.Time
	turno    map[string]bool
	inicio   int64
	apruebo  *apruebo
	gastos   map[string]gasto
}

var reFeature = regexp.MustCompile(`^E\d{3}/F\d{3}$`)

// escanear lee un registro una vez. D000096: el turno y el apruebo vigente se eligen por hora, no por posición:
// Claude Code escribe a veces el mensaje humano nuevo antes de líneas más viejas. El turno son las llamadas con hora
// igual o posterior al mensaje humano más reciente; un mensaje sin hora no abre turno.
func escanear(raiz, ruta string) (escaneo, error) {
	r := escaneo{llamadas: map[string]time.Time{}, turno: map[string]bool{}, gastos: map[string]gasto{}}
	var desde time.Time
	err := recorrerRegistro(ruta, func(_ int64, l lineaLlamadas) {
		hora, _ := time.Parse(time.RFC3339, l.Timestamp)
		if texto := backlog.TextoHumano(l.Type, l.Origin.Kind, l.Message.Content); texto != "" {
			if hora.After(desde) {
				desde = hora
			}
			if id, ok := backlog.IDAprobado(texto); ok && reFeature.MatchString(id) && !hora.IsZero() && (r.apruebo == nil || !hora.Before(r.apruebo.hora)) {
				ap := apruebo{feature: id, tokens: topeTokensPorDefecto, hora: hora}
				primera, _, _ := strings.Cut(texto, "\n")
				leerTope(primera, &ap)
				r.apruebo = &ap
			}
			return
		}
		if u := l.Message.Usage; l.Type == "assistant" && u != nil && l.Message.ID != "" {
			// Hallazgo A1.3: tras /compact Claude Code repite respuestas con usage en cero; gana el mayor peso del id.
			// Hallazgo A2.2: a igual peso se queda la hora más temprana; una copia tardía no mete la respuesta en la ventana.
			p := pesoVeinteavos(*u)
			viejo, hay := r.gastos[l.Message.ID]
			if !hay || p > viejo.peso || (p == viejo.peso && hora.Before(viejo.hora)) {
				r.gastos[l.Message.ID] = gasto{hora: hora, peso: p}
			}
		}
		for _, id := range llamadas(raiz, l) {
			r.llamadas[id] = hora
		}
	})
	for id, h := range r.llamadas {
		if !h.Before(desde) {
			r.turno[id] = true
		}
	}
	r.inicio = desde.UnixNano()
	return r, err
}

// carpetaProyecto: la carpeta de las sesiones del proyecto; desde un registro de subagente sube a ella.
func carpetaProyecto(ruta string) string {
	dir := filepath.Dir(ruta)
	if filepath.Base(dir) == "subagents" {
		return filepath.Dir(filepath.Dir(dir))
	}
	return dir
}

// escanearProyecto lee cada registro de sesión de la carpeta del proyecto; sin registro legible, error.
func escanearProyecto(raiz, ruta string) (map[string]escaneo, error) {
	if ruta == "" {
		return nil, errors.New("el hook no trajo el registro de la sesión (transcript_path)")
	}
	if _, err := os.Stat(ruta); err != nil {
		return nil, fmt.Errorf("no se puede leer el registro de la sesión %s", ruta)
	}
	rutas, _ := filepath.Glob(filepath.Join(carpetaProyecto(ruta), "*.jsonl"))
	r := map[string]escaneo{}
	for _, s := range rutas {
		e, err := escanear(raiz, s)
		if err != nil {
			return nil, fmt.Errorf("no se puede leer el registro %s: %w", s, err)
		}
		r[filepath.Clean(s)] = e
	}
	return r, nil
}

// featureCerrada (E001/F017/S001, D000091): true solo si la feature y todos sus slices (al menos uno) están en CERRADO.
// Cerrar un slice exige auditoría APROBADO; cancelar no la pide, así que un slice CANCELADO o una feature sin
// slices no sueltan el tope (hallazgo A1.2). Activa, sin ficha o backlog ilegible: false (no se inventa salida).
func featureCerrada(raiz, id string) bool {
	fs, err := backlog.Todas(raiz)
	if err != nil {
		return false
	}
	feature, slices := false, 0
	for _, f := range fs {
		switch {
		case f.ID == id:
			feature = f.Estado == "CERRADO"
		case f.Padre == id && f.Tipo == "slice":
			if f.Estado != "CERRADO" {
				return false
			}
			slices++
		}
	}
	return feature && slices > 0
}

// frenoFeature: el último apruebo de feature del humano fija feature, tope y hora; se cuentan los tokens ponderados
// (o las llamadas distintas, con «tope N» viejo) desde esa hora en las sesiones y en sus subagentes. "" si puede seguir o no hay apruebo de feature.
// vigente: hay apruebo de feature y la feature sigue abierta (E001/F017/S009).
func frenoFeature(raiz, ruta string, sesiones map[string]escaneo) (msg string, vigente bool) {
	var ap *apruebo
	for _, s := range sesiones {
		if s.apruebo != nil && (ap == nil || s.apruebo.hora.After(ap.hora)) {
			ap = s.apruebo
		}
	}
	if ap == nil || featureCerrada(raiz, ap.feature) {
		return "", false
	}
	ids := map[string]bool{}
	pesos := map[string]int64{} // E001/F019/S001: por message.id, una vez aunque se repita en varias líneas
	sumar := func(e escaneo) {
		for id, hora := range e.llamadas {
			if !hora.Before(ap.hora) {
				ids[id] = true
			}
		}
		for id, g := range e.gastos {
			if !g.hora.Before(ap.hora) && g.peso > pesos[id] {
				pesos[id] = g.peso
			}
		}
	}
	for _, s := range sesiones {
		sumar(s)
	}
	subagentes, _ := filepath.Glob(filepath.Join(carpetaProyecto(ruta), "*", "subagents", "*.jsonl"))
	for _, sub := range subagentes {
		if st, err := os.Stat(sub); err != nil || st.ModTime().Before(ap.hora) {
			continue
		}
		e, err := escanear(raiz, sub)
		if err != nil {
			return fmt.Sprintf("Presupuesto: no se puede leer el registro %s: %v.", sub, err), true
		}
		sumar(e)
	}
	if ap.tokens > 0 {
		var veinteavos int64
		for _, p := range pesos {
			veinteavos += p
		}
		if veinteavos > 20*ap.tokens {
			return fmt.Sprintf("Presupuesto: llegaste al tope de la feature %s (%s tokens ponderados; gastados %s). No sigas: informa qué falta y pide un tope nuevo («apruebo %s tope NM»).",
				ap.feature, millones(ap.tokens), millones(veinteavos/20), ap.feature), true
		}
		return "", true
	}
	if len(ids) > ap.tope {
		return fmt.Sprintf("Presupuesto: llegaste al tope de la feature %s (%d llamadas). No sigas: informa qué falta y pide un tope nuevo («apruebo %s tope N»).", ap.feature, ap.tope, ap.feature), true
	}
	return "", true
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
