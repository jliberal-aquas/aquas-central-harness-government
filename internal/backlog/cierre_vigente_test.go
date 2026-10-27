package backlog

// E001/F014/S001: CERRADO exige que el código Go del alcance siga igual que en el verde.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// escribirEn escribe un archivo bajo raiz (crea carpetas).
func escribirEn(t *testing.T, raiz, rel, txt string) {
	t.Helper()
	ruta := filepath.Join(raiz, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ruta, []byte(txt), 0o644); err != nil {
		t.Fatal(err)
	}
}

// sliceConVerde: slice EN_CURSO con alcance «p/**» y «docs/**», con rojo y verde reales,
// sellos marcados y listo para pedir CERRADO. p/p.go (Go) y docs/nota.md son archivos del alcance.
func sliceConVerde(t *testing.T, alcance ...string) (raiz, id string) {
	t.Helper()
	if len(alcance) == 0 {
		alcance = []string{"p/**", "docs/**"}
	}
	raiz, id = modulo(t), "E001/F002/S001"
	for _, a := range [][]string{{"init", "-q"}, {"-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "--allow-empty", "-m", "base"}} {
		if out, err := exec.Command("git", append([]string{"-C", raiz}, a...)...).CombinedOutput(); err != nil { // D000050: git propio, no el de la central
			t.Fatalf("git %v: %v %s", a, err, out)
		}
	}
	escribirEn(t, raiz, "p/p.go", "package p\n\nconst V = 1\n")
	escribirEn(t, raiz, "docs/nota.md", "# nota\n")
	nuevo(t, raiz, "id="+id, "titulo=T", "prioridad=P1", "motivo=m", "prueba=./p:TestBandera=pide el arreglo|r|v")
	if _, err := Alcance(raiz, id, alcance); err != nil {
		t.Fatal(err)
	}
	if _, err := arrancar(t, raiz, id); err != nil {
		t.Fatalf("EN_CURSO: %v", err)
	}
	if _, err := Rojo(raiz, id, hora); err != nil {
		t.Fatalf("rojo: %v", err)
	}
	escribirEn(t, raiz, "p/arreglado", "")
	if _, err := Verde(raiz, id, hora.Add(time.Hour)); err != nil {
		t.Fatalf("verde: %v", err)
	}
	marcarTodo(t, raiz, id)
	return raiz, id
}

// Rojo cuando: cierra tras editar un .go del alcance.
// Verde cuando: rechaza nombrando el cambio y cierra si nada cambió o solo cambiaron docs.
func TestCierreExigeVerdeVigente(t *testing.T) {
	t.Run("editar un .go del alcance rechaza y nombra el archivo", func(t *testing.T) {
		raiz, id := sliceConVerde(t)
		escribirEn(t, raiz, "p/p.go", "package p\n\nconst V = 2\n")
		_, err := Estado(raiz, id, "CERRADO")
		if err == nil || !strings.Contains(err.Error(), "p/p.go") {
			t.Errorf("cerrar con p/p.go cambiado: se esperaba rechazo nombrando p/p.go; dio %v", err)
		}
	})
	t.Run("alcance sin Go: cierra si sigue sin Go, rechaza un .go nuevo", func(t *testing.T) {
		raiz, id := sliceConVerde(t, "docs/**")
		escribirEn(t, raiz, "docs/nuevo.go", "package docs\n")
		if _, err := Estado(raiz, id, "CERRADO"); err == nil || !strings.Contains(err.Error(), "docs/nuevo.go") {
			t.Errorf("huella vacía y luego un .go: se esperaba rechazo nombrando docs/nuevo.go; dio %v", err)
		}
		raiz, id = sliceConVerde(t, "docs/**")
		if _, err := Estado(raiz, id, "CERRADO"); err != nil {
			t.Errorf("huella vacía sin cambios debió cerrar: %v", err)
		}
	})
	t.Run("comodín en carpeta registra el verde y cubre el archivo", func(t *testing.T) {
		raiz, id := sliceConVerde(t, "*/p.go", "docs/**")
		escribirEn(t, raiz, "p/p.go", "package p\n\nconst V = 3\n")
		if _, err := Estado(raiz, id, "CERRADO"); err == nil || !strings.Contains(err.Error(), "p/p.go") {
			t.Errorf("alcance */p.go con p/p.go cambiado: se esperaba rechazo; dio %v", err)
		}
	})
	t.Run("sin cambios cierra", func(t *testing.T) {
		raiz, id := sliceConVerde(t)
		if _, err := Estado(raiz, id, "CERRADO"); err != nil {
			t.Errorf("verde sin cambios posteriores: debe cerrar; dio %v", err)
		}
	})
	t.Run("cambio solo en un .md del alcance cierra", func(t *testing.T) {
		raiz, id := sliceConVerde(t)
		escribirEn(t, raiz, "docs/nota.md", "# nota\n\ncambio solo en docs\n")
		if _, err := Estado(raiz, id, "CERRADO"); err != nil {
			t.Errorf("cambio en docs no entra en la huella: debe cerrar; dio %v", err)
		}
	})
	t.Run(".go nuevo en el alcance rechaza", func(t *testing.T) {
		raiz, id := sliceConVerde(t)
		escribirEn(t, raiz, "p/nuevo.go", "package p\n")
		_, err := Estado(raiz, id, "CERRADO")
		if err == nil || !strings.Contains(err.Error(), "p/nuevo.go") {
			t.Errorf("cerrar con p/nuevo.go añadido: se esperaba rechazo nombrando p/nuevo.go; dio %v", err)
		}
	})
	t.Run(".go borrado del alcance rechaza", func(t *testing.T) {
		raiz, id := sliceConVerde(t)
		if err := os.Remove(filepath.Join(raiz, "p", "p.go")); err != nil {
			t.Fatal(err)
		}
		_, err := Estado(raiz, id, "CERRADO")
		if err == nil || !strings.Contains(err.Error(), "p/p.go") {
			t.Errorf("cerrar con p/p.go borrado: se esperaba rechazo nombrando p/p.go; dio %v", err)
		}
	})
	t.Run("verde sin huella (registro viejo) cierra como hoy", func(t *testing.T) {
		raiz, id := sliceConVerde(t)
		f := leerFicha(t, activo(raiz, nombre(id)))
		f.Verde = &Registro{Fecha: f.Verde.Fecha, Head: f.Verde.Head} // como antes de este tramo: sin huella
		pisar(t, raiz, Activo, f)
		escribirEn(t, raiz, "p/p.go", "package p\n\nconst V = 3\n")
		if _, err := Estado(raiz, id, "CERRADO"); err != nil {
			t.Errorf("verde viejo sin huella: cierra como hoy; dio %v", err)
		}
	})
}

// Rojo cuando: cierra con código cambiado tras el verde o sin veredicto.
// Verde cuando: ambos rechazos ocurren y el caso sano cierra.
// Hoy cubre la parte de S001; el caso del veredicto fallará hasta S002 (cuando exista, el caso sano debe llevar veredicto APROBADO).
func TestCierreHonesto(t *testing.T) {
	t.Run("código cambiado tras el verde rechaza", func(t *testing.T) {
		raiz, id := sliceConVerde(t)
		escribirEn(t, raiz, "p/p.go", "package p\n\nconst V = 9\n")
		if _, err := Estado(raiz, id, "CERRADO"); err == nil {
			t.Errorf("cerró con código cambiado tras el verde")
		}
	})
	t.Run("sin veredicto APROBADO rechaza", func(t *testing.T) {
		raiz, id := sliceConVerde(t)
		_, err := Estado(raiz, id, "CERRADO")
		if err == nil || !strings.Contains(err.Error(), "veredicto") {
			t.Errorf("cerrar sin veredicto del auditor: se esperaba rechazo que mencione «veredicto»; dio %v", err)
		}
	})
	t.Run("caso sano cierra", func(t *testing.T) {
		raiz, id := sliceConVerde(t)
		if _, err := Estado(raiz, id, "CERRADO"); err != nil {
			t.Errorf("caso sano (hoy sin veredicto): debe cerrar; dio %v", err)
		}
	})
}
