package canario

import "os"

// Permisos ignora el error a propósito: errcheck debe verlo.
func Permisos(p string) { os.Chmod(p, 0o755) }
