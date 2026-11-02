package hijo

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RutaPreCommit: el pre-commit del hijo, relativo a su raíz.
const RutaPreCommit = ".githooks/pre-commit"

// contenidoPreCommit: el hijo no tiene fuentes Go; corre el binario aquas instalado.
const contenidoPreCommit = "#!/bin/sh\nexec aquas verify\n"

// correVerify: el último renglón con comando del pre-commit (sin espacios, \r de CRLF ni comentarios) es
// `aquas verify` o `exec aquas verify`. Así un pre-commit propio unido a mano vale (hallazgo A1.1), y su
// salida es la del gate: en sh el hook sale con el código del último comando (hallazgo A2.1).
func correVerify(contenido string) bool {
	ultimo := ""
	for _, l := range strings.Split(contenido, "\n") {
		if l = strings.TrimSpace(l); l != "" && !strings.HasPrefix(l, "#") {
			ultimo = l
		}
	}
	return ultimo == "aquas verify" || ultimo == "exec aquas verify"
}

// RevisarPreCommit (E003/F001/S005, hallazgos A1.2 y A1.3): problema si el hijo no tiene repo git propio,
// su pre-commit no corre aquas verify o core.hooksPath no es .githooks; "" si está activo. El control
// git:pre-commit es «todos»: sin él, el gate del hijo no puede dar OK.
func RevisarPreCommit(dir string) string {
	const arreglo = ": el control git:pre-commit vale en cada hijo; corre git init (si falta) y aquas update"
	if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
		return "sin repo git propio" + arreglo
	}
	if b, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(RutaPreCommit))); err != nil || !correVerify(string(b)) {
		return RutaPreCommit + " ausente o sin «aquas verify»" + arreglo
	}
	if h, _ := exec.Command("git", "-C", dir, "config", "--get", "core.hooksPath").Output(); strings.TrimSpace(string(h)) != ".githooks" {
		return "core.hooksPath no es .githooks" + arreglo
	}
	return ""
}

// AsegurarPreCommit (E003/F001/S005) escribe .githooks/pre-commit con `aquas verify` y fija core.hooksPath.
// Sin repo git no hay dónde fijarlo: no hace nada. Un pre-commit propio distinto o un core.hooksPath ajeno no se pisan.
func AsegurarPreCommit(dir string) error {
	// El repo es el de dir mismo: un .git de una carpeta madre no cuenta (no se toca su configuración).
	if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
		return nil
	}
	if err := exec.Command("git", "-C", dir, "rev-parse", "--git-dir").Run(); err != nil {
		return nil // .git no usable (p. ej. falso)
	}
	p := filepath.Join(dir, filepath.FromSlash(RutaPreCommit))
	b, err := os.ReadFile(p)
	switch {
	case err == nil && !correVerify(string(b)):
		return fmt.Errorf("%s ya existe y es propio; únele a mano la línea «aquas verify» como último comando (o reemplázalo por esta: %q) y corre aquas update", RutaPreCommit, strings.TrimSpace(contenidoPreCommit))
	case err != nil && !errors.Is(err, fs.ErrNotExist):
		return err
	}
	actual, _ := exec.Command("git", "-C", dir, "config", "--get", "core.hooksPath").Output()
	if h := strings.TrimSpace(string(actual)); h != "" && h != ".githooks" {
		return fmt.Errorf("core.hooksPath ya apunta a «%s»; unifícalo: pasa tus ganchos a .githooks (con aquas verify en pre-commit), fija core.hooksPath=.githooks y corre aquas update", h)
	}
	if err != nil { // no existía
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(p, []byte(contenidoPreCommit), 0o755); err != nil {
			return err
		}
	}
	if out, err := exec.Command("git", "-C", dir, "config", "core.hooksPath", ".githooks").CombinedOutput(); err != nil {
		return fmt.Errorf("git config core.hooksPath: %w: %s", err, out)
	}
	return nil
}
