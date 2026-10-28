package hijo

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// PrepararHuecos (E002/F002/S001): crea con texto PENDIENTE los huecos de la central
// que el hijo aún no escribió. Un hueco ya escrito no se toca.
func PrepararHuecos(dir string) error {
	// Misma fuente que armar: los huecos viven solo en la constitución central.
	b, err := os.ReadFile(filepath.Join(dir, Carpeta, "central", "gobierno", "canonico", "constitucion.md"))
	if err != nil {
		return err
	}
	nombres := map[string]bool{}
	for _, m := range reHueco.FindAllStringSubmatch(string(b), -1) {
		nombres[m[1]] = true
	}
	hd := filepath.Join(dir, filepath.FromSlash(CarpetaHuecos))
	if err := os.MkdirAll(hd, 0o755); err != nil {
		return err
	}
	for n := range nombres {
		p := filepath.Join(hd, n+".md")
		if _, err := os.Stat(p); err == nil {
			continue
		} else if !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		txt := "- " + MarcaPendiente + " escribe aquí el hueco «" + n + "» de este proyecto.\n"
		if n == "titulo" {
			txt = "# CONSTITUCIÓN · " + MarcaPendiente + " nombre del proyecto\n"
		}
		if err := os.WriteFile(p, []byte(txt), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// Listo deja al hijo con sus huecos y su CLAUDE.md armado.
func Listo(dir string) error {
	if err := PrepararHuecos(dir); err != nil {
		return err
	}
	return Build(dir)
}
