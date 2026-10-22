package hijo

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// Update reemplaza la central sellada por otra versión y reescribe el lock.
// No toca lo local (huecos, CLAUDE.md). Devuelve los archivos sellados que
// alguien había alterado: quedan restaurados, y conviene reportarlo.
func Update(dest string, central fs.FS, version string) ([]string, error) {
	if version == "" {
		return nil, errors.New("update: falta la versión de la central")
	}
	rest, err := alterados(dest)
	if err != nil {
		return nil, err
	}
	base := filepath.Join(dest, Carpeta, "central")
	// Lo sellado es solo lectura; en Windows hay que habilitar escritura para borrarlo.
	filepath.WalkDir(base, func(p string, d fs.DirEntry, err error) error {
		if err == nil {
			os.Chmod(p, 0o755)
		}
		return nil
	})
	if err := os.RemoveAll(base); err != nil {
		return nil, err
	}
	return rest, copiarCentral(dest, central, version)
}

// alterados lista los archivos sellados que ya no coinciden con el lock.
func alterados(dest string) ([]string, error) {
	b, err := os.ReadFile(filepath.Join(dest, Carpeta, "lock.json"))
	if err != nil {
		return nil, fmt.Errorf("update: %s no tiene gobierno instalado (usa install)", dest)
	}
	var l Lock
	if err := json.Unmarshal(b, &l); err != nil {
		return nil, err
	}
	var out []string
	for p, want := range l.Archivos {
		c, err := os.ReadFile(filepath.Join(dest, Carpeta, "central", filepath.FromSlash(p)))
		h := sha256.Sum256(c)
		if err != nil || hex.EncodeToString(h[:]) != want {
			out = append(out, p)
		}
	}
	sort.Strings(out)
	return out, nil
}
