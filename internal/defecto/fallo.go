package defecto

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// reCampoFallo: el campo «¿Salió algo mal?» seguido de «:», con la negrita antes o después de los dos puntos.
	// Va al inicio de la línea, tras viñeta, emoji o negrita. Una cita en el cuerpo (tras letra, número o comillas)
	// o una mención sin «:» no es el campo (hallazgo A1.1).
	reCampoFallo = regexp.MustCompile("^[^\\p{L}\\p{N}«\"`]*¿Salió algo mal\\?(?:\\*\\*)?:(?:\\*\\*)?(.*)$")
	reNumero     = regexp.MustCompile(`\bD\d{6}\b`)
)

// RevisarFallo (E001/F004/S002) devuelve "" si pasa; si no, el motivo del freno. El campo «¿Salió algo mal?»
// pasa con «No» o «—»; si cuenta un fallo, debe nombrar al menos un D000NNN y todos deben existir en el registro.
func RevisarFallo(raiz, msg string) string {
	for _, l := range strings.Split(strings.ReplaceAll(msg, "\r\n", "\n"), "\n") {
		m := reCampoFallo.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		v := strings.TrimSuffix(strings.TrimSpace(strings.Trim(strings.TrimSpace(m[1]), "*")), ".")
		if v == "No" || v == "—" {
			continue
		}
		ids := reNumero.FindAllString(v, -1)
		if len(ids) == 0 {
			return "Fallo sin defecto: el campo «¿Salió algo mal?» cuenta un fallo y no nombra su defecto. Regístralo con " +
				"`aquas defecto nuevo firma=… que=… causa=… arreglo=… cercioro=… relacion=…` (o `aquas defecto ocurre D000NNN T000NNN` si se repite) " +
				"y nombra su número en el campo; si no hubo fallo, escribe «No»."
		}
		hay := registrados(raiz)
		for _, id := range ids {
			if !hay[id] {
				return "Fallo con defecto inexistente: " + id + " no está en " + Carpeta + ". Regístralo con `aquas defecto nuevo …` y usa el número que devuelve."
			}
		}
	}
	return ""
}

// registrados: números con archivo en el registro. Registro ausente o ilegible = ninguno (hallazgo A1.2).
func registrados(raiz string) map[string]bool {
	es, _ := os.ReadDir(filepath.Join(raiz, Carpeta))
	hay := map[string]bool{}
	for _, e := range es {
		if id, ok := strings.CutSuffix(e.Name(), ".json"); ok && !e.IsDir() {
			hay[id] = true
		}
	}
	return hay
}
