package defecto

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func campos(firma string) map[string]string {
	return map[string]string{"firma": firma, "que": "qué pasó", "causa": "por qué", "arreglo": "cómo se arregló", "cercioro": "cómo me cercioro"}
}

func TestRegistroDeDefectos(t *testing.T) {
	raiz := t.TempDir()

	d, err := Nuevo(raiz, campos("hook-exit-go-run"))
	if err != nil {
		t.Fatal(err)
	}
	if d.ID != "D000001" || d.Firma != "hook-exit-go-run" || len(d.Ocurrencias) != 0 {
		t.Fatalf("primer defecto: %+v", d)
	}
	if _, err := os.Stat(filepath.Join(raiz, Carpeta, "D000001.json")); err != nil {
		t.Fatalf("no quedó en %s: %v", Carpeta, err)
	}
	if d2, err := Nuevo(raiz, campos("otro-fallo")); err != nil || d2.ID != "D000002" {
		t.Fatalf("segundo defecto: %+v %v", d2, err)
	}

	// Rechazos, cada uno por su motivo.
	malos := map[string]struct {
		c    map[string]string
		pide string
	}{
		"firma repetida":    {campos("hook-exit-go-run"), "ya existe"},
		"campo vacío":       {map[string]string{"firma": "x-y", "que": "a", "causa": "", "arreglo": "c", "cercioro": "d"}, "causa"},
		"firma no kebab":    {campos("Hook Exit"), "firma"},
		"campo desconocido": {func() map[string]string { c := campos("z"); c["extra"] = "1"; return c }(), "extra"},
	}
	for nombre, m := range malos {
		if _, err := Nuevo(raiz, m.c); err == nil || !strings.Contains(err.Error(), m.pide) {
			t.Errorf("%s: error %v; debía mencionar %q", nombre, err, m.pide)
		}
	}

	// ocurre suma un turno; turno repetido o mal formado falla.
	if d, err = Ocurre(raiz, "D000001", "T000042"); err != nil || !reflect.DeepEqual(d.Ocurrencias, []string{"T000042"}) {
		t.Fatalf("ocurre: %+v %v", d, err)
	}
	if _, err := Ocurre(raiz, "D000001", "T000042"); err == nil || !strings.Contains(err.Error(), "ya") {
		t.Errorf("turno repetido debía fallar: %v", err)
	}
	if _, err := Ocurre(raiz, "D000001", "42"); err == nil || !strings.Contains(err.Error(), "turno") {
		t.Errorf("turno mal formado debía fallar: %v", err)
	}
	if _, err := Ocurre(raiz, "D000009", "T000043"); err == nil {
		t.Error("defecto inexistente debía fallar")
	}

	// Semilla real del repo.
	todos, err := Leer(filepath.Join("..", ".."))
	if err != nil || len(todos) == 0 || todos[0].ID != "D000001" || todos[0].Firma != "hook-exit-go-run" {
		t.Fatalf("semilla D000001 hook-exit-go-run ausente: %+v %v", todos, err)
	}
}
