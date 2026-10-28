package hijo

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// CarpetaHuecos: donde el hijo escribe lo suyo, un archivo por hueco (<nombre>.md).
const CarpetaHuecos = "gobierno/canonico/huecos"

// reHueco: bloque que la central deja abierto; lo de adentro es su relleno propio.
var reHueco = regexp.MustCompile(`(?s)<!-- hueco: ([a-z0-9-]+) -->\r?\n.*?<!-- /hueco -->\r?\n`)

// reSoloCentral: texto que vale solo en la central; build lo quita en el hijo.
var reSoloCentral = regexp.MustCompile(`(?s)<!-- solo-central -->\r?\n.*?<!-- /solo-central -->\r?\n`)

// reSoloHijo: texto comentado en la central; build lo descomenta en el hijo.
var reSoloHijo = regexp.MustCompile(`(?s)<!-- solo-hijo\r?\n(.*?)-->\r?\n`)

// Build arma el CLAUDE.md del hijo: central sellada + huecos locales.
// Falla si lo sellado no coincide con el lock, si un hueco queda vacío
// o si lo local nombra algo que no es hueco (una sección final).
func Build(dir string) error {
	txt, err := armar(dir)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte(txt), 0o644)
}

// armar devuelve el CLAUDE.md que Build escribiría, sin escribirlo.
func armar(dir string) (string, error) {
	central := filepath.Join(dir, Carpeta, "central")
	if err := verificarSello(dir, central); err != nil {
		return "", err
	}
	b, err := os.ReadFile(filepath.Join(central, "gobierno", "canonico", "constitucion.md"))
	if err != nil {
		return "", err
	}
	locales, err := leerHuecos(filepath.Join(dir, filepath.FromSlash(CarpetaHuecos)))
	if err != nil {
		return "", err
	}
	usados := map[string]bool{}
	var faltan []string
	out := reHueco.ReplaceAllStringFunc(string(b), func(bloque string) string {
		nombre := reHueco.FindStringSubmatch(bloque)[1]
		txt, ok := locales[nombre]
		if !ok {
			faltan = append(faltan, nombre)
			return bloque
		}
		usados[nombre] = true
		return txt
	})
	if len(faltan) > 0 {
		return "", fmt.Errorf("hueco sin llenar: %s (escribe %s/<nombre>.md)", strings.Join(faltan, ", "), CarpetaHuecos)
	}
	var ajenos []string
	for nombre := range locales {
		if !usados[nombre] {
			ajenos = append(ajenos, nombre)
		}
	}
	if len(ajenos) > 0 {
		sort.Strings(ajenos)
		return "", fmt.Errorf("no son huecos de la central (sección final, no se redefine): %s", strings.Join(ajenos, ", "))
	}
	out = reSoloCentral.ReplaceAllString(out, "")
	out = reSoloHijo.ReplaceAllString(out, "$1")
	cab := "<!-- GENERADO por `aquas build` desde .aquas/central + " + CarpetaHuecos + ". NO EDITAR. -->\n\n"
	return cab + out, nil
}

// verificarSello compara cada archivo de la central con su hash del lock.
func verificarSello(dir, central string) error {
	b, err := os.ReadFile(filepath.Join(dir, Carpeta, "lock.json"))
	if err != nil {
		return fmt.Errorf("sin lock: ¿corriste aquas install? (%w)", err)
	}
	var l Lock
	if err := json.Unmarshal(b, &l); err != nil {
		return err
	}
	for p, want := range l.Archivos {
		c, err := os.ReadFile(filepath.Join(central, filepath.FromSlash(p)))
		if err != nil {
			return fmt.Errorf("sellado ausente: %s", p)
		}
		h := sha256.Sum256(c)
		if hex.EncodeToString(h[:]) != want {
			return fmt.Errorf("sellado alterado: %s no coincide con el lock %s", p, l.Version)
		}
	}
	return nil
}

// leerHuecos devuelve nombre → contenido de cada <nombre>.md local.
func leerHuecos(dir string) (map[string]string, error) {
	m := map[string]string{}
	es, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return m, nil
	}
	if err != nil {
		return nil, err
	}
	for _, e := range es {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		txt := string(b)
		if !strings.HasSuffix(txt, "\n") {
			txt += "\n"
		}
		m[strings.TrimSuffix(e.Name(), ".md")] = txt
	}
	return m, nil
}

// Verificar (X02, E002/F002/S004): gate del hijo; nil si su central sellada, sus huecos sin PENDIENTE y su CLAUDE.md están al día.
func Verificar(dir string) []string {
	quiero, err := armar(dir)
	if err != nil {
		return []string{err.Error()}
	}
	prob, err := pendientes(dir)
	if err != nil {
		return []string{err.Error()}
	}
	if hay, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md")); err != nil || string(hay) != quiero {
		prob = append(prob, "CLAUDE.md desfasado o ausente: corre aquas build "+dir)
	}
	return prob
}

// EsHijo: dir tiene un gobierno central instalado (lock.json de aquas install).
func EsHijo(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, Carpeta, "lock.json"))
	return err == nil
}

// MarcaPendiente: texto de relleno que PrepararHuecos escribe en cada hueco nuevo.
const MarcaPendiente = "PENDIENTE:"

// pendientes (E002/F002/S004): un problema por hueco que conserva la marca de relleno, en orden.
func pendientes(dir string) ([]string, error) {
	hs, err := leerHuecos(filepath.Join(dir, filepath.FromSlash(CarpetaHuecos)))
	if err != nil {
		return nil, err
	}
	nombres := make([]string, 0, len(hs))
	for n, txt := range hs {
		if strings.Contains(txt, MarcaPendiente) {
			nombres = append(nombres, n)
		}
	}
	sort.Strings(nombres)
	var prob []string
	for _, n := range nombres {
		prob = append(prob, "hueco pendiente: llena "+CarpetaHuecos+"/"+n+".md y corre aquas build")
	}
	return prob, nil
}
