package entorno

// D000026: un gate corrido con un binario de aquas compilado antes del último cambio da resultados
// falsos. El gate compara la fecha del binario con la del código de producción.

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// BinarioViejo (D000026): aviso si algún .go de producción es más nuevo que el binario exe; "" si está al día.
func BinarioViejo(raiz, exe string) string {
	info, err := os.Stat(exe)
	if err != nil {
		return ""
	}
	limite := info.ModTime()
	nuevo, cuando := "", limite
	for _, base := range []string{"cmd", "internal"} {
		_ = filepath.WalkDir(filepath.Join(raiz, base), func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
				return nil
			}
			if fi, err := d.Info(); err == nil && fi.ModTime().After(cuando) {
				r, _ := filepath.Rel(raiz, p)
				nuevo, cuando = filepath.ToSlash(r), fi.ModTime()
			}
			return nil
		})
	}
	if nuevo == "" {
		return ""
	}
	return fmt.Sprintf("binario de aquas más viejo que %s: el gate puede mentir; reconstruye con go build -o <binario> ./cmd/aquas", nuevo)
}
