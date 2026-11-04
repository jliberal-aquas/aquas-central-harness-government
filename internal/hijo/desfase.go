package hijo

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

// modulo: ruta de go install del binario aquas.
const modulo = "github.com/jliberal-aquas/aquas-central-harness-government/cmd/aquas"

// arregloUpdate: única salida del hijo desfasado; la guardia deja pasar `aquas update` por Bash.
const arregloUpdate = " → corre `aquas update` por Bash en la raíz del hijo (la guardia lo deja pasar)"

// Desfase (E003/F006/S002): "" si la central copiada en dir es la que lleva el binario; si no, el motivo del freno.
// Compara los SHA-256 del lock con los de central, no la cadena de versión. Si el binario es más viejo que el
// lock, pide go install de la versión del lock: aquas update se niega a bajar la central (BajaCentral).
func Desfase(dir string, central fs.FS, binario string) string {
	b, err := os.ReadFile(filepath.Join(dir, Carpeta, "lock.json"))
	if err != nil {
		return "no se pudo leer " + Carpeta + "/lock.json" + arregloUpdate
	}
	var l Lock
	if json.Unmarshal(b, &l) != nil {
		return Carpeta + "/lock.json ilegible" + arregloUpdate
	}
	quiere, err := hashes(central)
	if err != nil {
		return "no se pudo leer la central del binario: " + err.Error()
	}
	if maps.Equal(l.Archivos, quiere) {
		return ""
	}
	if masViejo(binario, l.Version) {
		return bajaria(binario, l.Version)
	}
	return fmt.Sprintf("la central de este hijo (%s) no es la de aquas (%s)%s", l.Version, binario, arregloUpdate)
}

// BajaCentral (E003/F006/S002, hallazgo A1.1): motivo para que aquas update no corra si el binario es más viejo
// que la central del hijo (la bajaría); "" si puede correr. Vale para quien sea que lo lance.
func BajaCentral(dir, binario string) string {
	b, err := os.ReadFile(filepath.Join(dir, Carpeta, "lock.json"))
	var l Lock
	if err != nil || json.Unmarshal(b, &l) != nil || !masViejo(binario, l.Version) {
		return ""
	}
	return bajaria(binario, l.Version)
}

// bajaria: motivo común cuando el binario es más viejo que el lock.
func bajaria(binario, lock string) string {
	return fmt.Sprintf("aquas (%s) es más viejo que la central de este hijo (%s) → el humano actualiza aquas: go install %s@%s",
		binario, lock, modulo, lock)
}

// hashes: SHA-256 de cada archivo de central, por ruta.
func hashes(central fs.FS) (map[string]string, error) {
	m := map[string]string{}
	err := fs.WalkDir(central, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := fs.ReadFile(central, p)
		if err != nil {
			return err
		}
		h := sha256.Sum256(b)
		m[p] = hex.EncodeToString(h[:])
		return nil
	})
	return m, err
}

// reVersion: vX.Y.Z con fecha opcional de pseudo-versión (vX.Y.Z-AAAA… o vX.Y.Z-0.AAAA…).
var reVersion = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)(?:-(?:0\.)?(\d{14})-[0-9a-f]{12})?(?:\+.*)?$`)

// masViejo: true solo si a es con certeza anterior a b. Una pseudo-versión vX.Y.Z-0.fecha va antes de la etiqueta vX.Y.Z.
// Orden desconocido (devel u otro formato) = false.
func masViejo(a, b string) bool {
	ma, mb := reVersion.FindStringSubmatch(a), reVersion.FindStringSubmatch(b)
	if ma == nil || mb == nil {
		return false
	}
	for i := 1; i <= 3; i++ {
		x, _ := strconv.Atoi(ma[i])
		y, _ := strconv.Atoi(mb[i])
		if x != y {
			return x < y
		}
	}
	switch fa, fb := ma[4], mb[4]; {
	case fa != "" && fb == "":
		return true
	case fa != "" && fb != "":
		return fa < fb
	}
	return false
}
