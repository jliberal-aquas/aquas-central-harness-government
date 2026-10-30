package wiki

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	reCita    = regexp.MustCompile("`aquas ([^`]*)`")
	reCase    = regexp.MustCompile(`case\s+"[^:]*:`)
	reLiteral = regexp.MustCompile(`"([a-z][a-z-]*)"`)
	rePalabra = regexp.MustCompile(`^[a-z][a-z-]*$`)
	// E000/F006/S006: la tabla de nombres en inglés de cmd/aquas; cada llave es un comando.
	reIngles = regexp.MustCompile(`(?s)var comandos = map\[string\]string\{(.*?)\n\}`)
	reLlave  = regexp.MustCompile(`"([a-z][a-z-]*)":`)
)

// RevisarComandos: todo `aquas X [Y]` citado entre comillas invertidas en docs/ y en
// gobierno/canonico/ existe en el CLI. X debe ser un `case` de cmd/aquas; Y, si es una
// palabra, debe aparecer como literal en el código de cmd/ o internal/.
// Sin cmd/aquas no hay contra qué comparar y no revisa nada.
// Lo que va sin comillas invertidas (p. ej. un comando en construcción) no se revisa.
func RevisarComandos(raiz string) ([]string, error) {
	comandos, literales := map[string]bool{}, map[string]bool{}
	for _, dir := range []string{"cmd", "internal"} {
		err := filepath.WalkDir(filepath.Join(raiz, dir), func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				if os.IsNotExist(err) {
					return filepath.SkipDir
				}
				return err
			}
			if d.IsDir() || !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
				return nil
			}
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			for _, m := range reLiteral.FindAllStringSubmatch(string(b), -1) {
				literales[m[1]] = true
			}
			if filepath.ToSlash(filepath.Dir(p)) == filepath.ToSlash(filepath.Join(raiz, "cmd", "aquas")) {
				for _, c := range reCase.FindAllString(string(b), -1) {
					for _, m := range reLiteral.FindAllStringSubmatch(c, -1) {
						comandos[m[1]] = true
					}
				}
				for _, t := range reIngles.FindAllStringSubmatch(string(b), -1) {
					for _, m := range reLlave.FindAllStringSubmatch(t[1], -1) {
						comandos[m[1]] = true
					}
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	if len(comandos) == 0 {
		return nil, nil
	}
	var problemas []string
	for _, dir := range []string{"docs", "gobierno/canonico"} {
		err := filepath.WalkDir(filepath.Join(raiz, dir), func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				if os.IsNotExist(err) {
					return filepath.SkipDir
				}
				return err
			}
			if d.IsDir() || !strings.HasSuffix(p, ".md") {
				return nil
			}
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			rel := relativa(raiz, p)
			for _, m := range reCita.FindAllStringSubmatch(string(b), -1) {
				ps := strings.Fields(m[1])
				if len(ps) == 0 || !rePalabra.MatchString(ps[0]) {
					continue
				}
				cita := "aquas " + ps[0]
				ok := comandos[ps[0]]
				if ok && len(ps) > 1 && rePalabra.MatchString(ps[1]) {
					cita += " " + ps[1]
					ok = literales[ps[1]]
				}
				if !ok {
					problemas = append(problemas, fmt.Sprintf("%s: comando citado que el CLI no tiene: `%s`", rel, cita))
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return problemas, nil
}
