package defecto

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/backlog"
)

// RevisarPrueba (E001/F017/S005): prueba=./paquete:TestNombre debe existir como func en un _test.go del paquete;
// prueba=pendiente:<id> debe nombrar una ficha que exista (activa o archivada) y no esté CANCELADO.
func RevisarPrueba(raiz, v string) error {
	v = strings.TrimSpace(v)
	if v == "" {
		return fmt.Errorf("falta el campo prueba: ./paquete:TestNombre de la prueba que vigila el arreglo, o pendiente:<id de ficha> que la hará")
	}
	if id, ok := strings.CutPrefix(v, "pendiente:"); ok {
		if id == "" {
			return fmt.Errorf("prueba=pendiente: sin ficha: usa pendiente:<id de ficha>")
		}
		f, _, err := backlog.Ver(raiz, id)
		if err != nil {
			return fmt.Errorf("prueba=%s: la ficha no existe", v)
		}
		if f.Estado == "CANCELADO" {
			return fmt.Errorf("prueba=%s: la ficha está CANCELADO y nunca hará la prueba", v)
		}
		return nil
	}
	// Mismo formato que las fichas (backlog.EsNombrePrueba); el paquete no sale del repo (A1.2).
	paquete, nombre, _ := strings.Cut(strings.TrimPrefix(v, "./"), ":")
	if !backlog.EsNombrePrueba(v) || slices.Contains(strings.Split(paquete, "/"), "..") {
		return fmt.Errorf("prueba=%s: formato ./paquete:TestNombre dentro del repo, o pendiente:<id de ficha>", v)
	}
	if !existePrueba(filepath.Join(raiz, filepath.FromSlash(paquete)), nombre) {
		return fmt.Errorf("prueba=%s: la prueba no existe en el código", v)
	}
	return nil
}

// existePrueba: algún _test.go de la carpeta declara func nombre(.
func existePrueba(dir, nombre string) bool {
	xs, _ := filepath.Glob(filepath.Join(dir, "*_test.go"))
	re := regexp.MustCompile(`(?m)^func ` + regexp.QuoteMeta(nombre) + `\(`)
	for _, x := range xs {
		if b, err := os.ReadFile(x); err == nil && re.Match(b) {
			return true
		}
	}
	return false
}

// SinFamilia (E001/F018/S003): defectos anteriores a la familia obligatoria; el gate solo avisa.
func SinFamilia(raiz string) int {
	todos, _ := Leer(raiz)
	n := 0
	for _, d := range todos {
		if d.Familia == "" {
			n++
		}
	}
	return n
}

// Revisar (E001/F017/S005, gate): un problema por defecto cuya prueba nombrada ya no vale; los defectos viejos sin
// campo solo se cuentan (se informan, no frenan).
func Revisar(raiz string) (problemas []string, sinPrueba int) {
	todos, err := Leer(raiz)
	if err != nil {
		return []string{"registro de defectos ilegible: " + err.Error()}, 0
	}
	for _, d := range todos {
		if d.Prueba == "" {
			sinPrueba++
		} else if err := RevisarPrueba(raiz, d.Prueba); err != nil {
			problemas = append(problemas, d.ID+": "+err.Error())
		}
	}
	return problemas, sinPrueba
}
