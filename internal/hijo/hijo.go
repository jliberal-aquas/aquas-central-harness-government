// Package hijo instala y mantiene el gobierno central en un proyecto hijo.
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
	"strings"
)

// Carpeta del gobierno heredado dentro del hijo.
const Carpeta = ".aquas"

// Lock: versión instalada y SHA-256 de cada archivo de la central copiado.
type Lock struct {
	Version  string            `json:"version"`
	Archivos map[string]string `json:"archivos"`
}

// Install copia la central sellada en dest/.aquas/central y escribe dest/.aquas/lock.json.
// E002/F002/S002: acepta una carpeta con código sin tocarlo; ante un conflicto falla sin escribir nada.
func Install(dest string, central fs.FS, version string) error {
	if version == "" {
		return errors.New("install: falta la versión de la central")
	}
	var huecos []string
	if b, err := fs.ReadFile(central, "gobierno/canonico/constitucion.md"); err == nil {
		vistos := map[string]bool{}
		for _, m := range reHueco.FindAllStringSubmatch(string(b), -1) {
			if !vistos[m[1]] {
				vistos[m[1]] = true
				huecos = append(huecos, m[1]+".md")
			}
		}
		sort.Strings(huecos)
	}
	conflictos := []struct{ ruta, aviso string }{
		{Carpeta, "ya tiene gobierno instalado; usa aquas update"},
		{"CLAUDE.md", "ya tiene un CLAUDE.md propio; sácalo de la carpeta (p. ej. renómbralo CLAUDE.previo.md), corre aquas install y pasa su contenido a los huecos de " +
			CarpetaHuecos + "/: " + strings.Join(huecos, ", ")},
	}
	// E001/F007/S003 (hallazgo A2.1): si la central trae /backlog, un SKILL.md previo del hijo es conflicto.
	if _, err := fs.Stat(central, FuenteSkill); err == nil {
		conflictos = append(conflictos, struct{ ruta, aviso string }{RutaSkill,
			"ya tiene un /backlog propio en " + RutaSkill + "; renómbralo o muévelo y corre aquas install"})
	}
	for _, c := range conflictos {
		_, err := os.Lstat(filepath.Join(dest, c.ruta))
		switch {
		case err == nil:
			return fmt.Errorf("install: %s %s", dest, c.aviso)
		case !errors.Is(err, fs.ErrNotExist):
			return err
		}
	}
	return copiarCentral(dest, central, version)
}

// copiarCentral escribe la central como solo lectura y su lock con versión y hashes.
func copiarCentral(dest string, central fs.FS, version string) error {
	l := Lock{Version: version, Archivos: map[string]string{}}
	base := filepath.Join(dest, Carpeta, "central")
	err := fs.WalkDir(central, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := fs.ReadFile(central, p)
		if err != nil {
			return err
		}
		destino := filepath.Join(base, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(destino), 0o755); err != nil {
			return err
		}
		// Solo lectura: lo sellado no se edita en el hijo.
		if err := os.WriteFile(destino, b, 0o444); err != nil {
			return err
		}
		h := sha256.Sum256(b)
		l.Archivos[p] = hex.EncodeToString(h[:])
		return nil
	})
	if err != nil {
		return err
	}
	if len(l.Archivos) == 0 {
		return errors.New("install: la central está vacía")
	}
	b, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dest, Carpeta, "lock.json"), append(b, '\n'), 0o644)
}

// AMedias (E002/F002/S002): error de install cuando la central ya quedó copiada y falló lo siguiente.
// Reintentar es update, no install: install rechaza una carpeta que ya tiene .aquas/.
func AMedias(err error) error {
	return fmt.Errorf("install quedó a medias (la central ya está copiada): %w; tras corregirlo corre aquas update", err)
}
