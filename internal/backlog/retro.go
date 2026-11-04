package backlog

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

// E001/F018/S002: disparador de retrospectiva. retro close registra la retro con la cuenta del momento;
// next compara esa cuenta con la actual y, en el umbral, pide /retro. Solo avisa: no frena nada.

// CarpetaRetros: registro de retros, una por archivo R000N.json. Solo la escribe `aquas retro close`.
const CarpetaRetros = "registro/retros"

// carpetaDefectos: registro de defectos; aquí solo se cuentan sus archivos (sin importar el paquete defecto).
const carpetaDefectos = "registro/defectos"

// UmbralRetro: slices CERRADO o defectos nuevos desde la última retro que piden otra.
const UmbralRetro = 10

// Retro: un registro de retrospectiva.
type Retro struct {
	ID             string           `json:"id"`
	Fecha          string           `json:"fecha"`
	Head           string           `json:"head"`
	SlicesCerrados int              `json:"slices_cerrados"`
	Defectos       int              `json:"defectos"`
	Fichas         []string         `json:"fichas"`
	Motivo         string           `json:"motivo"`
	Aprobacion     *AprobacionRetro `json:"aprobacion"`
}

// ArchivoAprobacionRetro: «apruebo retro» del humano. Lo escribe el hook del mensaje; retro close lo gasta.
const ArchivoAprobacionRetro = CarpetaRetros + "/aprobacion.json"

// AprobacionRetro: el mensaje del humano que autoriza cerrar una retro (decisión 1A, hallazgo A1.1).
type AprobacionRetro struct {
	Fecha    string `json:"fecha"`
	Mensaje  string `json:"mensaje"`
	Registro string `json:"registro"`
}

// reAprueboRetro: la línea empieza por «apruebo retro»; «apruebo retrospectiva» no cuenta.
var reAprueboRetro = regexp.MustCompile(`(?i)^\s*apruebo\s+retro(?:\s|$)`)

// EsAprobacionRetro: la primera línea no vacía del humano es «apruebo retro …» y no es pregunta.
func EsAprobacionRetro(prompt string) bool {
	if deAgente(prompt) {
		return false
	}
	for _, l := range strings.Split(strings.ReplaceAll(prompt, "\r\n", "\n"), "\n") {
		if strings.TrimSpace(l) == "" {
			continue
		}
		return !strings.HasSuffix(strings.TrimSpace(l), "?") && reAprueboRetro.MatchString(l)
	}
	return false
}

// AprobarRetro guarda la aprobación; la llama solo el hook del mensaje del humano. Una nueva reemplaza la anterior.
func AprobarRetro(raiz, mensaje, registro string, ahora time.Time) error {
	b, err := json.MarshalIndent(AprobacionRetro{Fecha: ahora.UTC().Format(time.RFC3339), Mensaje: strings.TrimSpace(mensaje), Registro: registro}, "", "  ")
	if err != nil {
		return err
	}
	p := filepath.Join(raiz, filepath.FromSlash(ArchivoAprobacionRetro))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, append(b, '\n'), 0o644)
}

// CerrarRetro registra la retro en R000N.json: fichas que salieron de ella (deben existir, sin repetir), motivo y la
// aprobación del humano, que se gasta. commit vacío = la cuenta de hoy; si no, la de ese commit (decisión 2A).
// Todo se valida antes de escribir: un rechazo no deja archivo ni gasta la aprobación.
func CerrarRetro(raiz string, fichas []string, motivo, commit string) (Retro, error) {
	if strings.TrimSpace(motivo) == "" {
		return Retro{}, fmt.Errorf("retro close: falta motivo=")
	}
	if len(fichas) == 0 {
		return Retro{}, fmt.Errorf("retro close: falta fichas=A,B (las fichas que salieron de la retro)")
	}
	for i, id := range fichas {
		if id == "" {
			return Retro{}, fmt.Errorf("retro close: fichas= trae un id vacío")
		}
		if slices.Contains(fichas[:i], id) {
			return Retro{}, fmt.Errorf("retro close: %s repetida en fichas=", id)
		}
		if _, _, err := Ver(raiz, id); err != nil {
			return Retro{}, fmt.Errorf("retro close: la ficha %s no existe", id)
		}
	}
	pa := filepath.Join(raiz, filepath.FromSlash(ArchivoAprobacionRetro))
	var ap AprobacionRetro
	b, err := os.ReadFile(pa)
	if err == nil {
		err = json.Unmarshal(b, &ap)
	}
	if err == nil {
		err = aprobacionRetroValida(raiz, ap)
	}
	if err != nil {
		return Retro{}, fmt.Errorf("retro close: falta «apruebo retro» válido del humano en el chat (el hook la registra); %v", err)
	}
	rs, err := retros(raiz)
	if err != nil {
		return Retro{}, err
	}
	r := Retro{Fecha: time.Now().UTC().Format(time.RFC3339), Head: head(raiz), Fichas: fichas, Motivo: motivo, Aprobacion: &ap}
	if commit == "" {
		r.SlicesCerrados, r.Defectos, err = cuentaRetro(raiz)
	} else {
		r.Head, r.Fecha, r.SlicesCerrados, r.Defectos, err = cuentaEnCommit(raiz, commit)
	}
	if err != nil {
		return Retro{}, err
	}
	sig := 1 // hallazgo A1.4: el siguiente número sale del mayor, no de contar archivos
	if len(rs) > 0 {
		k, _ := strconv.Atoi(reRetro.FindStringSubmatch(filepath.Base(rs[len(rs)-1]))[1])
		sig = k + 1
	}
	r.ID = fmt.Sprintf("R%04d", sig)
	b, err = json.MarshalIndent(r, "", "  ")
	if err != nil {
		return Retro{}, err
	}
	dir := filepath.Join(raiz, filepath.FromSlash(CarpetaRetros))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Retro{}, err
	}
	fh, err := os.OpenFile(filepath.Join(dir, r.ID+".json"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644) // nunca pisa una retro
	if err != nil {
		return Retro{}, err
	}
	_, err = fh.Write(append(b, '\n'))
	if cerr := fh.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Remove(pa) // una aprobación, una retro
	}
	return r, err
}

// aprobacionRetroValida (hallazgo A2.1, D000077): el mensaje es «apruebo retro» y está como mensaje humano en un
// registro de sesión válido, la misma segunda barrera que respaldada() pone a los apruebos de fichas.
func aprobacionRetroValida(raiz string, ap AprobacionRetro) error {
	if !EsAprobacionRetro(ap.Mensaje) {
		return fmt.Errorf("el mensaje guardado %q no es «apruebo retro»", ap.Mensaje)
	}
	reg, err := registroValido(raiz, ap.Registro)
	if err != nil {
		return fmt.Errorf("su registro de sesión %v", err)
	}
	if ok, err := mensajeHumano(reg, ap.Mensaje); err != nil || !ok {
		return fmt.Errorf("el registro de sesión no tiene el mensaje humano %q", ap.Mensaje)
	}
	return nil
}

// cuentaEnCommit: commit completo, su fecha y la cuenta de slices CERRADO y defectos que tenía el repo en él.
// Extrae trabajo/ y registro/defectos/ de ese commit con git archive a una carpeta de .tmp/ que luego borra.
func cuentaEnCommit(raiz, commit string) (hash, fecha string, slicesCerrados, defectos int, err error) {
	salida := func(args ...string) ([]byte, error) {
		b, err := exec.Command("git", append([]string{"-C", raiz}, args...)...).Output()
		if err != nil {
			return nil, fmt.Errorf("retro close: commit=%s: git %s: %v", commit, args[0], err)
		}
		return b, nil
	}
	b, err := salida("rev-parse", "--verify", "--quiet", commit+"^{commit}")
	if err != nil {
		return "", "", 0, 0, err
	}
	hash = strings.TrimSpace(string(b))
	if b, err = salida("show", "-s", "--format=%cI", hash); err != nil {
		return "", "", 0, 0, err
	}
	t, err := time.Parse(time.RFC3339, strings.TrimSpace(string(b)))
	if err != nil {
		return "", "", 0, 0, err
	}
	// Hallazgo A2.2: un commit sin trabajo/ o sin registro/defectos/ cuenta 0 de lo que falta, no falla.
	if b, err = salida("ls-tree", "--name-only", hash, "--", "trabajo", carpetaDefectos); err != nil {
		return "", "", 0, 0, err
	}
	if rutas := strings.Fields(string(b)); len(rutas) == 0 {
		b = nil
	} else if b, err = salida(append([]string{"archive", "--format=tar", hash, "--"}, rutas...)...); err != nil {
		return "", "", 0, 0, err
	}
	if err = os.MkdirAll(filepath.Join(raiz, ".tmp"), 0o755); err != nil {
		return "", "", 0, 0, err
	}
	dir, err := os.MkdirTemp(filepath.Join(raiz, ".tmp"), "retro-")
	if err != nil {
		return "", "", 0, 0, err
	}
	defer func() { _ = os.RemoveAll(dir) }() // carpeta temporal: si no se borra, .tmp/ la limpia aquas clean
	tr := tar.NewReader(bytes.NewReader(b))
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", "", 0, 0, err
		}
		if h.Typeflag != tar.TypeReg || strings.Contains(h.Name, "..") {
			continue
		}
		p := filepath.Join(dir, filepath.FromSlash(h.Name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return "", "", 0, 0, err
		}
		c, err := io.ReadAll(tr)
		if err == nil {
			err = os.WriteFile(p, c, 0o644)
		}
		if err != nil {
			return "", "", 0, 0, err
		}
	}
	slicesCerrados, defectos, err = cuentaRetro(dir)
	return hash, t.UTC().Format(time.RFC3339), slicesCerrados, defectos, err
}

// reRetro: nombre válido de un registro de retro; otro archivo de la carpeta no cuenta.
var reRetro = regexp.MustCompile(`^R(\d{4})\.json$`)

// retros: registros R0000.json de la carpeta, ordenados por número.
func retros(raiz string) ([]string, error) {
	xs, err := filepath.Glob(filepath.Join(raiz, filepath.FromSlash(CarpetaRetros), "R*.json"))
	var out []string
	for _, x := range xs {
		if reRetro.MatchString(filepath.Base(x)) {
			out = append(out, x)
		}
	}
	sort.Strings(out)
	return out, err
}

// cuentaRetro: slices CERRADO (activos y archivados) y archivos de defecto, hoy.
func cuentaRetro(raiz string) (slicesCerrados, defectos int, err error) {
	fs, err := Todas(raiz)
	if err != nil {
		return 0, 0, err
	}
	for _, f := range fs {
		if f.Tipo == "slice" && f.Estado == "CERRADO" {
			slicesCerrados++
		}
	}
	ds, err := filepath.Glob(filepath.Join(raiz, filepath.FromSlash(carpetaDefectos), "D*.json"))
	return slicesCerrados, len(ds), err
}

// AvisoRetro: «toca retrospectiva: …» si desde la última retro hay UmbralRetro slices CERRADO o defectos nuevos;
// vacío si no toca. Nunca falla: un registro ilegible se nombra en el aviso (next no se frena).
func AvisoRetro(raiz string) string {
	n, d, err := cuentaRetro(raiz)
	if err != nil {
		return "retrospectiva: no se pudo contar (" + err.Error() + ")"
	}
	rs, err := retros(raiz)
	if err != nil {
		return "retrospectiva: registro de retros ilegible (" + err.Error() + ")"
	}
	desde := "el inicio"
	if len(rs) > 0 {
		ult := rs[len(rs)-1]
		var r Retro
		b, err := os.ReadFile(ult)
		if err == nil {
			err = json.Unmarshal(b, &r)
		}
		if err != nil {
			return fmt.Sprintf("retrospectiva: %s ilegible (%v); arréglalo o corre /retro", filepath.Base(ult), err)
		}
		n, d, desde = n-r.SlicesCerrados, d-r.Defectos, r.ID
	}
	if n < UmbralRetro && d < UmbralRetro {
		return ""
	}
	return fmt.Sprintf("toca retrospectiva: %d slices y %d defectos desde %s; corre /retro", n, d, desde)
}
