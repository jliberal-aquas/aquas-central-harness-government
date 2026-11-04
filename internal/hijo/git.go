package hijo

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// AsegurarGit (E003/F006/S001) deja dir como raíz de un repo git: lo inicia si no está en ninguno; falla si está dentro de otro.
func AsegurarGit(dir string) error {
	if _, err := exec.LookPath("git"); err != nil {
		return errors.New("git no está en el PATH: instálalo y repite")
	}
	c := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel")
	c.Env = append(os.Environ(), "LC_ALL=C", "LANG=C")
	var out, errb bytes.Buffer
	c.Stdout, c.Stderr = &out, &errb
	if err := c.Run(); err != nil {
		if !strings.Contains(errb.String(), "not a git repository") {
			return fmt.Errorf("git no pudo leer la carpeta: %s", strings.TrimSpace(errb.String()))
		}
		if o, err := exec.Command("git", "-C", dir, "init", "-q").CombinedOutput(); err != nil {
			return fmt.Errorf("git init falló: %s", strings.TrimSpace(string(o)))
		}
		return nil
	}
	top := strings.TrimSpace(out.String())
	if !igualRuta(top, dir) {
		return fmt.Errorf("la carpeta está dentro del repo %s: instala en su raíz o fuera de él", top)
	}
	return nil
}

// igualRuta compara dos rutas limpias y absolutas; en Windows sin distinguir mayúsculas.
func igualRuta(a, b string) bool {
	a, _ = filepath.Abs(a)
	b, _ = filepath.Abs(b)
	a, b = filepath.ToSlash(filepath.Clean(a)), filepath.ToSlash(filepath.Clean(b))
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}
