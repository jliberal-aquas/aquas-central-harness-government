package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Id de encargo estricto: slice + sufijo de worker (T) o auditor (A). Nada más se borra.
// Acepta la forma con guiones o la del backlog (E000/F006/S010-A1); nunca separadores de más.
var reEncargo = regexp.MustCompile(`^E\d{3}-F\d{3}-S\d{3}-[AT]\d+$|^E\d{3}/F\d{3}/S\d{3}-[AT]\d+$`)

// limpiar (E000/F006/S009): borra .tmp/<id>/ de un encargo y nada más.
// Rechaza ids que no sean de encargo y carpetas con enlaces o junctions: ante la duda, no borra.
func limpiar(raiz string, args []string) (string, error) {
	if len(args) != 1 {
		return "", errors.New("uso: aquas limpiar <id-del-encargo>, p. ej. E000/F006/S010-A1")
	}
	if !reEncargo.MatchString(args[0]) {
		return "", fmt.Errorf("aquas limpiar: %q no es un id de encargo (E000-F006-S010-A1 o -T1); no se borra nada", args[0])
	}
	id := strings.ReplaceAll(args[0], "/", "-")
	tmp := filepath.Join(raiz, ".tmp")
	if fi, err := os.Lstat(tmp); err != nil || !fi.IsDir() || fi.Mode().Type() != fs.ModeDir {
		return "", errors.New("aquas limpiar: .tmp no es una carpeta normal; no se borra nada")
	}
	dir := filepath.Join(tmp, id)
	rel := ".tmp/" + id + "/"
	fi, err := os.Lstat(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return "nada que borrar: " + rel + " no existe", nil
	}
	if err != nil || fi.Mode().Type() != fs.ModeDir {
		return "", fmt.Errorf("aquas limpiar: %s no es una carpeta normal; no se borra nada", rel)
	}
	// Recorre antes de borrar: solo carpetas y archivos normales; un enlace o junction frena todo.
	err = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if t := d.Type(); t != 0 && t != fs.ModeDir {
			return fmt.Errorf("aquas limpiar: %s es un enlace o archivo especial; no se borra nada", p)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if err := os.RemoveAll(dir); err != nil {
		return "", fmt.Errorf("aquas limpiar: %w", err)
	}
	return "borrada " + rel, nil
}
