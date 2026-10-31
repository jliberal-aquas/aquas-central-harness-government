package guardia

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Lista de permitidos: gobierno/config/guardia.json declara las únicas rutas fuera del repo donde
// se puede escribir, cada una con su motivo. Sin archivo o ilegible = nada permitido fuera.
// Una ruta con ${VAR} sin definir en este equipo se ignora (p. ej. LOCALAPPDATA fuera de Windows).

const RutaConfig = "gobierno/config/guardia.json"

type permitidoFuera struct {
	Ruta   string `json:"ruta"`
	Motivo string `json:"motivo"`
}

var reVarConfig = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

// cargar devuelve las carpetas permitidas ya resueltas y los problemas de la config.
func cargar(raiz string) ([]string, []string) {
	b, err := os.ReadFile(filepath.Join(raiz, filepath.FromSlash(RutaConfig)))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	var c struct {
		Permitidos []permitidoFuera `json:"permitidos_fuera"`
	}
	if err == nil {
		err = json.Unmarshal(b, &c)
	}
	if err != nil {
		return nil, []string{RutaConfig + ": ilegible (" + err.Error() + "); nada se permite fuera"}
	}
	var dirs, ps []string
	for i, p := range c.Permitidos {
		pre := RutaConfig + ": permitidos_fuera[" + itoa(i) + "] "
		if strings.TrimSpace(p.Motivo) == "" {
			ps = append(ps, pre+"sin motivo")
		}
		if strings.TrimSpace(p.Ruta) == "" {
			ps = append(ps, pre+"ruta vacía")
			continue
		}
		r, ok := expandirConfig(p.Ruta)
		if !ok {
			continue // variable no definida en este equipo
		}
		r = filepath.Clean(r)
		switch {
		case !filepath.IsAbs(r):
			ps = append(ps, pre+"ruta «"+p.Ruta+"» no es absoluta")
		case filepath.Dir(r) == r:
			ps = append(ps, pre+"ruta «"+p.Ruta+"» es la raíz de unidad: permitiría todo")
		case strings.TrimSpace(p.Motivo) != "":
			dirs = append(dirs, r)
		}
	}
	return dirs, ps
}

// expandirConfig sustituye variables de entorno ${X} de la config; ok falso si falta alguna.
func expandirConfig(s string) (string, bool) {
	ok := true
	s = reVarConfig.ReplaceAllStringFunc(s, func(m string) string {
		v := os.Getenv(m[2 : len(m)-1])
		if v == "" {
			ok = false
		}
		return v
	})
	if s == "~" || strings.HasPrefix(s, "~/") {
		h, err := os.UserHomeDir()
		if err != nil {
			return "", false
		}
		s = h + s[1:]
	}
	return s, ok
}

// permitido: abs cae dentro de alguna carpeta de la lista.
func permitido(raiz, abs string) bool {
	dirs, _ := cargar(raiz)
	for _, d := range dirs {
		if dentro(d, abs) {
			return true
		}
	}
	return false
}

// ValidarConfig es el gate de la lista de permitidos (lo llama `aquas verify`).
func ValidarConfig(raiz string) []string {
	_, ps := cargar(raiz)
	return ps
}

// itoa convierte un entero a texto.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b []byte
	for ; i > 0; i /= 10 {
		b = append([]byte{byte('0' + i%10)}, b...)
	}
	return string(b)
}
