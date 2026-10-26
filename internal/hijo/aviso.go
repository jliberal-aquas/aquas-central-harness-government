package hijo

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
)

// reFecha toma la fecha que Go pone en una pseudo-versión: v0.0.0-AAAAMMDDhhmmss-hash,
// o vX.Y.Z-0.AAAAMMDDhhmmss-hash cuando ya hay una etiqueta previa.
var reFecha = regexp.MustCompile(`[-.](\d{14})-[0-9a-f]{12}(\+.*)?$`)

// AvisoVersion (E002/F002/S003) devuelve el aviso de versión del hijo, o "" si no hay nada que avisar.
// Avisa cuando el lock y el ejecutable tienen fechas distintas. Solo avisa: nunca falla.
func AvisoVersion(dir, binario string) string {
	b, err := os.ReadFile(filepath.Join(dir, Carpeta, "lock.json"))
	if err != nil {
		return ""
	}
	var l Lock
	if json.Unmarshal(b, &l) != nil {
		return ""
	}
	fl, fb := reFecha.FindStringSubmatch(l.Version), reFecha.FindStringSubmatch(binario)
	if fl == nil || fb == nil {
		return "" // devel o versión con etiqueta: no hay fecha que comparar
	}
	// Mismo largo y solo dígitos: el orden de texto es el orden de fecha.
	switch {
	case fl[1] < fb[1]:
		return "aviso: el gobierno de este hijo (" + l.Version + ") es más viejo que aquas (" + binario + "): corre aquas update"
	case fl[1] > fb[1]:
		return "aviso: aquas (" + binario + ") es más viejo que el gobierno de este hijo (" + l.Version + "): actualiza aquas con go install"
	}
	return ""
}
