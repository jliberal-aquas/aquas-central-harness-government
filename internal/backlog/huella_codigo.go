package backlog

// E001/F014/S001: el verde guarda el SHA-256 de cada .go del alcance; CERRADO lo recalcula.

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// huellaVerde: huella para el registro del verde. Nunca nil: sin archivos Go queda vacía,
// distinta de un verde anterior a este tramo (nil), y un .go que aparezca después impide cerrar.
func huellaVerde(raiz string, f Ficha) (*map[string]string, error) {
	h, err := huellaCodigo(raiz, f)
	if err != nil {
		return nil, err
	}
	return &h, nil
}

// baseFija: carpeta del patrón antes del primer tramo con comodín («internal/*/x.go» → «internal»).
func baseFija(p string) string {
	partes := strings.Split(p, "/")
	var fijas []string
	for _, s := range partes[:len(partes)-1] {
		if strings.ContainsAny(s, "*?[") {
			break
		}
		fijas = append(fijas, s)
	}
	return strings.Join(fijas, "/")
}

// fueraDelRepo: patrón absoluto, con unidad («C:») o con un tramo «..»; la huella no lee fuera de la raíz.
func fueraDelRepo(p string) bool {
	if strings.HasPrefix(p, "/") || filepath.IsAbs(filepath.FromSlash(p)) || filepath.VolumeName(filepath.FromSlash(p)) != "" || strings.ContainsAny(p, `:\`) {
		return true
	}
	for _, s := range strings.Split(p, "/") {
		if s == ".." {
			return true
		}
	}
	return false
}

// huellaCodigo: ruta relativa con «/» → SHA-256 de cada .go cubierto por el alcance (con ampliaciones).
// Omite carpetas .git, .tmp y testdata. Sin archivos Go devuelve un mapa vacío. Un archivo ilegible es error.
func huellaCodigo(raiz string, f Ficha) (map[string]string, error) {
	patrones := append([]string{}, f.Alcance...)
	for _, a := range f.Ampliaciones {
		patrones = append(patrones, a.Ruta)
	}
	res := map[string]string{}
	vistos := map[string]bool{}
	for _, p := range patrones {
		if fueraDelRepo(p) {
			return nil, fmt.Errorf("alcance fuera del repo: %s; usa rutas relativas a la raíz, sin «..»", p)
		}
		inicio := filepath.Join(raiz, filepath.FromSlash(baseFija(p)))
		err := filepath.WalkDir(inicio, func(ruta string, d fs.DirEntry, err error) error {
			if err != nil {
				if os.IsNotExist(err) && ruta == inicio {
					return nil // carpeta aún inexistente: sin archivos
				}
				return err
			}
			if d.IsDir() {
				if n := d.Name(); n == ".git" || n == ".tmp" || n == "testdata" {
					return filepath.SkipDir
				}
				return nil
			}
			rel, err := filepath.Rel(raiz, ruta)
			if err != nil {
				return err
			}
			rel = filepath.ToSlash(rel)
			if !strings.HasSuffix(rel, ".go") || vistos[rel] || !cubre(p, rel) {
				return nil
			}
			vistos[rel] = true
			b, err := os.ReadFile(ruta)
			if err != nil {
				return fmt.Errorf("no se pudo leer %s: %w", rel, err)
			}
			h := sha256.Sum256(b)
			res[rel] = hex.EncodeToString(h[:])
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

// codigoCambio: rutas (ordenadas) que difieren entre la huella del verde y el código de ahora.
// Solo se llama con f.Verde.Huella distinta de nil.
func codigoCambio(raiz string, f Ficha) ([]string, error) {
	ahora, err := huellaCodigo(raiz, f)
	if err != nil {
		return nil, err
	}
	antes := *f.Verde.Huella
	var difieren []string
	for r, h := range antes {
		if ahora[r] != h {
			difieren = append(difieren, r)
		}
	}
	for r := range ahora {
		if _, ok := antes[r]; !ok {
			difieren = append(difieren, r)
		}
	}
	sort.Strings(difieren)
	return difieren, nil
}
