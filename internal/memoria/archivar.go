package memoria

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// --- S003: archivo al cerrar el slice ---

// Archivo: carpeta versionada donde queda la memoria de cada slice cerrado.
const Archivo = "registro/memoria"

// RutaArchivo: registro/memoria/<E>-<F>-<S>.json; "" si el foco no llega a slice.
func RutaArchivo(f Foco) string {
	if f.Iniciativa == "" || f.Feature == "" || f.Slice == "" {
		return ""
	}
	return Archivo + "/" + f.Iniciativa + "-" + f.Feature + "-" + f.Slice + ".json"
}

// Archivar copia los bytes sellados de la memoria a registro/memoria/<E>-<F>-<S>.json y deja
// la memoria lista para el slice siguiente: conserva iniciativa, feature, pendientes y
// decisiones y siguiente (qué viene después del slice); vacía slice y tarea. Devuelve la ruta relativa del archivo.
// Exige foco con slice, sello bien y memoria al día; nunca pisa un archivo existente.
func Archivar(raiz string, ahora time.Time) (string, error) {
	m, err := Leer(raiz)
	if err != nil {
		return "", err
	}
	rel := RutaArchivo(m.Foco)
	if rel == "" {
		return "", errors.New("archivar exige foco con iniciativa, feature y slice: aquas memoria fijar slice=S…")
	}
	if HuellaGit(raiz) != m.HuellaGit {
		return "", fmt.Errorf("memoria caliente vieja: el repo cambió desde %s; fija el estado final antes de archivar (aquas memoria fijar tarea=… siguiente=…)", m.Actualizado)
	}
	dest := filepath.Join(raiz, filepath.FromSlash(rel))
	if _, err := os.Stat(dest); err == nil {
		return "", fmt.Errorf("%s ya existe: ese slice ya se archivó y no se pisa", rel)
	}
	b, err := os.ReadFile(filepath.Join(raiz, Ruta))
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", err
	}
	f, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644) // O_EXCL: nunca pisa
	if err != nil {
		return "", err
	}
	if _, err := f.Write(b); err != nil {
		_ = f.Close() // limpieza tras fallo de escritura; el error útil es el de Write
		_ = os.Remove(dest)
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	// Foco del slice siguiente: misma feature, sin slice; «siguiente» se conserva porque al cerrar
	// un slice dice qué viene (fallo visto 2026-10-21). La huella ya incluye el archivo nuevo.
	if _, err := Fijar(raiz, []string{"slice=", "tarea="}, ahora); err != nil {
		return rel, err
	}
	return rel, nil
}

// LeerArchivo carga una memoria archivada y comprueba su sello.
func LeerArchivo(ruta string) (Memoria, error) {
	b, err := os.ReadFile(ruta)
	if err != nil {
		return Memoria{}, err
	}
	return decodificar(b)
}
