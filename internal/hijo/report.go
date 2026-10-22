package hijo

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RepoCentral: dónde viven los issues de gobierno.
const RepoCentral = "jliberal-aquas/aquas-central-harness-government"

// Report arma el borrador de issue para la central: firma del hijo, versión
// del lock, qué pasó, cómo reproducirlo y sellados alterados. No envía nada:
// enviarlo es una acción externa que decide el humano.
func Report(dir, quePaso, repro string) (string, error) {
	if strings.TrimSpace(quePaso) == "" || strings.TrimSpace(repro) == "" {
		return "", errors.New("report: faltan qué pasó y cómo reproducirlo")
	}
	b, err := os.ReadFile(filepath.Join(dir, Carpeta, "lock.json"))
	if err != nil {
		return "", fmt.Errorf("report: %s no tiene gobierno instalado (usa install)", dir)
	}
	var l Lock
	if err := json.Unmarshal(b, &l); err != nil {
		return "", err
	}
	alt, err := alterados(dir)
	if err != nil {
		return "", err
	}
	sellados := "ninguno"
	if len(alt) > 0 {
		sellados = strings.Join(alt, ", ")
	}
	abs, _ := filepath.Abs(dir)
	return fmt.Sprintf("Reportado desde el hijo «%s» · central %s\n\n## Qué pasó\n%s\n\n## Cómo reproducirlo\n%s\n\nSellados alterados: %s\n",
		filepath.Base(abs), l.Version, quePaso, repro, sellados), nil
}
