// Package guardia: regla de oro «no escribir fuera del repositorio», como gate determinista.
//
// Se ejecuta como hook PreToolUse de Claude Code (`aquas guardia`). Exit 2 bloquea la herramienta.
//   - Write/Edit/MultiEdit/NotebookEdit: la ruta debe quedar dentro de la raíz del repo.
//   - Bash: se analizan redirecciones (>, >>, &>), cd, y los destinos de rm, mv, cp, mkdir, touch,
//     tee, ln, chmod, sed -i, perl -i, git clone/init/worktree add, curl -o, wget -O, dd of=, unzip -d.
//     Leer fuera está permitido. Ruta con variable desconocida = no se puede probar → se bloquea.
//   - Excepciones: solo las carpetas de gobierno/config/guardia.json (permitidos.go), con motivo.
//
// Límite (criterio humano): un programa que escribe por su cuenta (python, go run, node…) no se
// inspecciona; tampoco `git -C`, `tar -C` ni enlaces simbólicos que salgan del repo.
package guardia

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

const regla = "regla de oro: solo se escribe dentro del repositorio"

// Revisar recibe el JSON del hook PreToolUse; "" = permitido, si no, el motivo del bloqueo.
func Revisar(raiz string, entrada []byte) string {
	var e struct {
		ToolName  string `json:"tool_name"`
		Cwd       string `json:"cwd"`
		ToolInput struct {
			FilePath     string `json:"file_path"`
			NotebookPath string `json:"notebook_path"`
			Command      string `json:"command"`
		} `json:"tool_input"`
	}
	if err := json.Unmarshal(entrada, &e); err != nil {
		return "guardia: entrada del hook ilegible; se bloquea por seguridad (" + regla + ")"
	}
	switch e.ToolName {
	case "Write", "Edit", "MultiEdit":
		return RevisarArchivo(raiz, e.ToolInput.FilePath)
	case "NotebookEdit":
		return RevisarArchivo(raiz, e.ToolInput.NotebookPath)
	case "Bash":
		cwd := e.Cwd
		if cwd == "" {
			cwd = raiz
		}
		return RevisarBash(raiz, cwd, e.ToolInput.Command)
	case "PowerShell":
		return "PowerShell no se analiza: no se puede probar que no escribe fuera; usa Bash (" + regla + ")"
	}
	return ""
}

// RevisarArchivo valida una ruta de Write/Edit (absoluta o relativa a la raíz).
func RevisarArchivo(raiz, p string) string {
	if p == "" {
		return "ruta vacía: fuera del repositorio por defecto (" + regla + ")"
	}
	abs, ok := resolver(raiz, p)
	if !ok {
		return fmt.Sprintf("no se puede resolver la ruta «%s» (%s)", p, regla)
	}
	if !dentro(raiz, abs) && !permitido(raiz, abs) {
		return fmt.Sprintf("escritura fuera del repositorio bloqueada: %s (%s: %s)", p, regla, raiz)
	}
	return memoria(raiz, abs, p)
}

// memoria: .aquas/ guarda la memoria caliente; solo la escribe `aquas memoria fijar` (E001/F005).
func memoria(raiz, abs, p string) string {
	if dentro(filepath.Join(raiz, ".aquas"), abs) {
		return fmt.Sprintf("memoria caliente: %s solo se escribe con `aquas memoria fijar campo=valor`", p)
	}
	return ""
}

// resolver lleva p a ruta absoluta del sistema. Entiende ~, /c/… (Git Bash) y relativas a base.
func resolver(base, p string) (string, bool) {
	if p == "~" || strings.HasPrefix(p, "~/") || strings.HasPrefix(p, `~\`) {
		h, err := os.UserHomeDir()
		if err != nil {
			return "", false
		}
		p = h + p[1:]
	}
	if runtime.GOOS == "windows" {
		if m := reUnidadBash.FindStringSubmatch(p); m != nil {
			p = strings.ToUpper(m[1]) + ":/" + m[2]
		} else if strings.HasPrefix(p, "/") || strings.HasPrefix(p, `\`) && !strings.HasPrefix(p, `\\`) {
			return "/", true // raíz de Git Bash (C:\Program Files\Git…): fuera
		}
	}
	if !filepath.IsAbs(p) {
		if base == "" {
			return "", false
		}
		p = filepath.Join(base, p)
	}
	return filepath.Clean(p), true
}

var reUnidadBash = regexp.MustCompile(`^/([a-zA-Z])(?:/(.*))?$`)

func dentro(raiz, abs string) bool {
	rel, err := filepath.Rel(raiz, abs) // en Windows compara sin distinguir mayúsculas
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}

// ---- Bash ----

type palabra struct {
	texto    string // sin comillas; variables como \x00NOMBRE\x00
	redirige bool   // operador de redirección a archivo (>, >>, &>)
	op       bool   // separador de comandos (&&, ||, ;, |, &, salto)
}

var reHeredoc = regexp.MustCompile(`<<-?\s*(?:'([^']+)'|"([^"]+)"|\\?([A-Za-z_][A-Za-z0-9_]*))`)

// sinHeredocs quita el cuerpo de cada heredoc: su contenido es texto, no comandos.
func sinHeredocs(c string) string {
	var out []string
	fin := ""
	for _, l := range strings.Split(strings.ReplaceAll(c, "\r\n", "\n"), "\n") {
		if fin != "" {
			if strings.TrimSpace(l) == fin {
				fin = ""
			}
			continue
		}
		if m := reHeredoc.FindStringSubmatch(l); m != nil {
			fin = m[1] + m[2] + m[3]
			l = reHeredoc.ReplaceAllString(l, " ")
		}
		out = append(out, l)
	}
	return strings.Join(out, "\n")
}

func lexear(c string) []palabra {
	var ps []palabra
	var b strings.Builder
	hay := false
	cerrar := func() {
		if hay {
			ps = append(ps, palabra{texto: b.String()})
		}
		b.Reset()
		hay = false
	}
	r := []rune(c)
	simple, doble := false, false
	for i := 0; i < len(r); i++ {
		ch := r[i]
		switch {
		case simple:
			if ch == '\'' {
				simple = false
			} else {
				b.WriteRune(ch)
			}
			continue
		case ch == '\\' && i+1 < len(r):
			i++
			b.WriteRune(r[i])
			hay = true
			continue
		case ch == '\'' && !doble:
			simple, hay = true, true
			continue
		case ch == '"':
			doble, hay = !doble, true
			continue
		case ch == '$' || ch == '`' || ch == '%':
			n, salto := variable(r, i)
			if salto > 0 {
				b.WriteString("\x00" + n + "\x00")
				hay = true
				i += salto - 1
				continue
			}
		}
		if doble {
			b.WriteRune(ch)
			continue
		}
		switch ch {
		case ' ', '\t':
			cerrar()
		case '\n', ';', '|', '&':
			if ch == '&' && i+1 < len(r) && r[i+1] == '>' { // &> archivo
				cerrar()
				i++
				if i+1 < len(r) && r[i+1] == '>' {
					i++
				}
				ps = append(ps, palabra{redirige: true})
				continue
			}
			cerrar()
			if i+1 < len(r) && (r[i+1] == ch || ch == '|' && r[i+1] == '&') {
				i++
			}
			ps = append(ps, palabra{op: true})
		case '>':
			fd := b.String()
			if hay && strings.Trim(fd, "0123456789") != "" {
				cerrar()
			} else {
				b.Reset()
				hay = false
			}
			if i+1 < len(r) && r[i+1] == '>' {
				i++
			}
			if i+1 < len(r) && r[i+1] == '&' { // >&1, 2>&1: duplica descriptor, no es archivo
				i++
				for i+1 < len(r) && (r[i+1] >= '0' && r[i+1] <= '9' || r[i+1] == '-') {
					i++
				}
				continue
			}
			if i+1 < len(r) && r[i+1] == '|' {
				i++
			}
			ps = append(ps, palabra{redirige: true})
		case '<', '(', ')':
			cerrar()
		default:
			b.WriteRune(ch)
			hay = true
		}
	}
	cerrar()
	return ps
}

// variable reconoce $NOMBRE, ${NOMBRE}, $(…), `…` y %NOMBRE%; devuelve nombre ("?" = desconocida) y largo.
func variable(r []rune, i int) (string, int) {
	switch r[i] {
	case '`':
		for j := i + 1; j < len(r); j++ {
			if r[j] == '`' {
				return "?", j - i + 1
			}
		}
		return "?", len(r) - i
	case '%':
		for j := i + 1; j < len(r); j++ {
			if r[j] == '%' && j > i+1 {
				return "?", j - i + 1
			}
			if !esNombre(r[j]) {
				break
			}
		}
		return "", 0
	}
	if i+1 >= len(r) {
		return "", 0
	}
	switch n := r[i+1]; {
	case n == '(':
		prof := 0
		for j := i + 1; j < len(r); j++ {
			if r[j] == '(' {
				prof++
			} else if r[j] == ')' {
				if prof--; prof == 0 {
					return "?", j - i + 1
				}
			}
		}
		return "?", len(r) - i
	case n == '{':
		for j := i + 2; j < len(r); j++ {
			if r[j] == '}' {
				return string(r[i+2 : j]), j - i + 1
			}
		}
		return "?", len(r) - i
	case esNombre(n):
		j := i + 1
		for j < len(r) && esNombre(r[j]) {
			j++
		}
		return string(r[i+1 : j]), j - i
	}
	return "", 0
}

func esNombre(c rune) bool {
	return c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9'
}

var reAsignacion = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=`)

// RevisarBash analiza un comando de Bash; cwd es el directorio donde arranca.
func RevisarBash(raiz, cwd, comando string) string {
	vars := map[string]string{}
	actual, cwdOK := cwd, true
	ps := lexear(sinHeredocs(comando))
	for ini := 0; ini < len(ps); {
		fin := ini
		for fin < len(ps) && !ps[fin].op {
			fin++
		}
		if m := segmento(raiz, &actual, &cwdOK, vars, ps[ini:fin]); m != "" {
			return m
		}
		ini = fin + 1
	}
	return ""
}

func segmento(raiz string, cwd *string, cwdOK *bool, vars map[string]string, ps []palabra) string {
	var args []string
	var destinos []string
	soloAsignaciones := true
	for i := 0; i < len(ps); i++ {
		if ps[i].redirige {
			if i+1 < len(ps) {
				i++
				destinos = append(destinos, ps[i].texto)
			}
			continue
		}
		if len(args) == 0 && reAsignacion.MatchString(ps[i].texto) {
			continue
		}
		soloAsignaciones = false
		args = append(args, ps[i].texto)
	}
	if soloAsignaciones {
		for _, p := range ps {
			if k, v, ok := strings.Cut(p.texto, "="); ok && !p.redirige {
				if e, ok := expandir(v, vars, *cwd); ok {
					vars[k] = e
				} else {
					vars[k] = "\x00?\x00"
				}
			}
		}
	}
	if len(args) > 0 {
		cmd := filepath.Base(strings.ReplaceAll(args[0], `\`, "/"))
		resto := args[1:]
		pos := posicionales(resto)
		switch cmd {
		case "cd", "pushd":
			dst := "~"
			if len(pos) > 0 {
				dst = pos[0]
			}
			if e, ok := expandir(dst, vars, *cwd); ok {
				if abs, ok := resolver(*cwd, e); ok {
					*cwd = abs
					return ""
				}
			}
			*cwdOK = false
		case "rm", "rmdir", "mkdir", "touch", "chmod", "chown", "ln", "truncate", "shred", "unlink", "install", "tee", "mv":
			destinos = append(destinos, pos...)
		case "cp":
			if len(pos) > 0 {
				destinos = append(destinos, pos[len(pos)-1])
			}
		case "sed", "perl":
			for _, a := range resto {
				if strings.HasPrefix(a, "-i") || strings.HasPrefix(a, "--in-place") || strings.HasPrefix(a, "-pi") {
					if len(pos) > 1 {
						destinos = append(destinos, pos[1:]...)
					}
					break
				}
			}
		case "git":
			if len(pos) >= 1 {
				switch pos[0] {
				case "clone":
					if len(pos) >= 3 {
						destinos = append(destinos, pos[len(pos)-1])
					}
				case "init":
					destinos = append(destinos, pos[1:]...)
				case "worktree":
					if len(pos) >= 3 && pos[1] == "add" {
						destinos = append(destinos, pos[2])
					}
				}
			}
		case "curl", "wget", "unzip":
			for i, a := range resto {
				if (a == "-o" || a == "--output" || a == "-O" || a == "--output-document" || a == "-P" || a == "-d") && i+1 < len(resto) {
					if cmd == "curl" && a == "-O" {
						continue // curl -O guarda con el nombre remoto en cwd
					}
					destinos = append(destinos, resto[i+1])
				}
			}
		case "dd":
			for _, a := range resto {
				if v, ok := strings.CutPrefix(a, "of="); ok {
					destinos = append(destinos, v)
				}
			}
		}
	}
	for _, d := range destinos {
		if m := revisarDestino(raiz, *cwd, *cwdOK, vars, d); m != "" {
			return m
		}
	}
	return ""
}

func posicionales(args []string) []string {
	var p []string
	for _, a := range args {
		if a != "-" && strings.HasPrefix(a, "-") {
			continue
		}
		p = append(p, a)
	}
	return p
}

func revisarDestino(raiz, cwd string, cwdOK bool, vars map[string]string, d string) string {
	e, ok := expandir(d, vars, cwd)
	if !ok {
		return fmt.Sprintf("no se puede resolver la ruta «%s»: escribe la ruta literal (%s)", legible(d), regla)
	}
	switch e {
	case "/dev/null", "/dev/stdout", "/dev/stderr", "NUL", "nul":
		return ""
	}
	if !cwdOK && !filepath.IsAbs(e) && !strings.HasPrefix(e, "/") && !strings.HasPrefix(e, "~") {
		return fmt.Sprintf("no se puede resolver la ruta «%s»: un cd anterior no es literal (%s)", e, regla)
	}
	abs, ok := resolver(cwd, e)
	if !ok {
		return fmt.Sprintf("no se puede resolver la ruta «%s» (%s)", e, regla)
	}
	if !dentro(raiz, abs) && !permitido(raiz, abs) {
		return fmt.Sprintf("escritura fuera del repositorio bloqueada: %s (%s: %s)", e, regla, raiz)
	}
	return memoria(raiz, abs, e)
}

var reVar = regexp.MustCompile("\x00([^\x00]*)\x00")

// expandir sustituye variables conocidas; HOME/USERPROFILE/TEMP/TMP salen del entorno real.
func expandir(s string, vars map[string]string, cwd string) (string, bool) {
	ok := true
	out := reVar.ReplaceAllStringFunc(s, func(m string) string {
		n := m[1 : len(m)-1]
		if v, hay := vars[n]; hay && !strings.Contains(v, "\x00") {
			return v
		}
		switch n {
		case "PWD":
			return cwd
		case "HOME", "USERPROFILE", "TEMP", "TMP", "TMPDIR", "APPDATA", "LOCALAPPDATA":
			if v := os.Getenv(n); v != "" {
				return v
			}
			if h, err := os.UserHomeDir(); err == nil && (n == "HOME" || n == "USERPROFILE") {
				return h
			}
		}
		ok = false
		return ""
	})
	return out, ok
}

func legible(s string) string { return reVar.ReplaceAllString(s, "$$$1") }
