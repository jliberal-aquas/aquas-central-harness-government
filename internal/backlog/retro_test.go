package backlog

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestNextPideRetro (E001/F018/S002): pasado el umbral desde la última retro, next pide /retro; retro close la registra.
// Rojo cuando: pasado el umbral el aviso no sale, o cuenta lo que no debe, o un rechazo deja archivo.
// Verde cuando: avisa en el umbral, cuenta solo slices CERRADO y defectos, y un rechazo no escribe.
func TestNextPideRetro(t *testing.T) {
	t.Run("A1.1 solo «apruebo retro» del humano aprueba", esAprobacionRetro)
	t.Run("A2.2 commit sin trabajo ni defectos cuenta cero", commitVacioCuentaCero)
	raiz := t.TempDir()
	fichaEn(t, raiz, Activo, Ficha{ID: "E001", Titulo: "I", Estado: "EN_CURSO", Secuencia: 1})
	fichaEn(t, raiz, Activo, Ficha{ID: "E001/F002", Titulo: "F", Estado: "EN_CURSO", Secuencia: 2})
	fichaEn(t, raiz, Archivo, Ficha{ID: "E001/F003", Titulo: "F", Estado: "CERRADO", Secuencia: 3})
	fichaEn(t, raiz, Archivo, Ficha{ID: "E002", Titulo: "I", Estado: "CERRADO", Secuencia: 4}) // A1.6: iniciativa CERRADO no suma
	sec := 10
	slice := func(dir, padre, estado string) {
		sec++
		fichaEn(t, raiz, dir, Ficha{ID: fmt.Sprintf("%s/S%03d", padre, sec), Titulo: "S", Estado: estado, Secuencia: sec})
	}
	// Tipo y estado: 9 slices CERRADO (de una feature activa y de una archivada); un CANCELADO, la feature CERRADO
	// y un slice en curso no suman.
	for i := 0; i < 5; i++ {
		slice(Archivo, "E001/F002", "CERRADO")
	}
	for i := 0; i < 4; i++ {
		slice(Archivo, "E001/F003", "CERRADO")
	}
	slice(Archivo, "E001/F002", "CANCELADO")
	slice(Activo, "E001/F002", "EN_CURSO")
	if a := AvisoRetro(raiz); a != "" {
		t.Fatalf("9 slices cerrados, bajo el umbral: sin aviso; dio %q", a)
	}
	// Sin retro previa cuenta desde cero: el décimo cerrado dispara el aviso.
	slice(Archivo, "E001/F003", "CERRADO")
	if a := AvisoRetro(raiz); !strings.Contains(a, "toca retrospectiva: 10 slices y 0 defectos desde el inicio; corre /retro") {
		t.Fatalf("en el umbral sin retro previa debe avisar; dio %q", a)
	}

	// Entrada ausente, vacía, mal escrita o duplicada; motivo vacío: rechaza sin dejar archivo.
	for _, c := range []struct {
		fichas []string
		motivo string
	}{
		{nil, "m"}, {[]string{""}, "m"}, {[]string{"E001/F999"}, "m"}, {[]string{"E001/F002", "E001/F002"}, "m"}, {[]string{"E001/F002"}, " "},
	} {
		if _, err := CerrarRetro(raiz, c.fichas, c.motivo, ""); err == nil {
			t.Errorf("retro close %v motivo=%q: debió rechazar", c.fichas, c.motivo)
		}
	}
	if xs, _ := filepath.Glob(filepath.Join(raiz, filepath.FromSlash(CarpetaRetros), "*")); len(xs) != 0 {
		t.Fatalf("un rechazo no deja archivo de retro: %v", xs)
	}

	// A1.1 (decisión 1A): sin «apruebo retro» del humano, retro close rechaza sin dejar archivo.
	if _, err := CerrarRetro(raiz, []string{"E001/F002"}, "m", ""); err == nil || !strings.Contains(err.Error(), "apruebo retro") {
		t.Fatalf("sin aprobación retro close rechaza nombrando «apruebo retro»; dio %v", err)
	}
	if xs, _ := filepath.Glob(filepath.Join(raiz, filepath.FromSlash(CarpetaRetros), "*")); len(xs) != 0 {
		t.Fatalf("un rechazo sin aprobación no deja archivo: %v", xs)
	}
	// A2.1: una aprobación sin su mensaje humano en un registro de sesión válido no sirve (agente que evade la guardia).
	for nombre, reg := range map[string]string{
		"registro inexistente":      "reg.jsonl",
		"mensaje de otro":           registro(t, lineaUsuario("apruebo E001", "human", nil)),
		"mismo texto, no es humano": registro(t, lineaUsuario("apruebo retro", "", nil)),
		// A3.2: línea humana válida, pero el registro vive en el repo, donde el agente escribe: lo frena registroValido.
		"registro dentro del repo": jsonlEn(t, filepath.Join(raiz, ".tmp"), "sesion-*.jsonl", []string{lineaUsuario("apruebo retro", "human", nil)}),
	} {
		if err := AprobarRetro(raiz, "apruebo retro", reg, time.Now()); err != nil {
			t.Fatal(err)
		}
		if _, err := CerrarRetro(raiz, []string{"E001/F002"}, "m", ""); err == nil || !strings.Contains(err.Error(), "apruebo retro") {
			t.Errorf("aprobación falsa (%s): retro close rechaza; dio %v", nombre, err)
		}
	}
	if xs, _ := filepath.Glob(filepath.Join(raiz, filepath.FromSlash(CarpetaRetros), "R*.json")); len(xs) != 0 {
		t.Fatalf("una aprobación falsa no deja retro: %v", xs)
	}
	aprobar := func() {
		t.Helper()
		if err := AprobarRetro(raiz, "apruebo retro", registro(t, lineaUsuario("apruebo retro", "human", nil)), time.Now()); err != nil {
			t.Fatal(err)
		}
	}

	// Cerrar registra R0001 con la cuenta del momento y la aprobación; el aviso se apaga.
	aprobar()
	r, err := CerrarRetro(raiz, []string{"E001/F002", "E001/F003"}, "retro de prueba", "")
	if err != nil {
		t.Fatal(err)
	}
	if r.ID != "R0001" || r.SlicesCerrados != 10 || r.Defectos != 0 || len(r.Fichas) != 2 || r.Fecha == "" ||
		r.Motivo != "retro de prueba" || r.Aprobacion == nil || r.Aprobacion.Mensaje != "apruebo retro" {
		t.Fatalf("R0001 con 10 slices, 0 defectos, sus fichas, su motivo y su aprobación: %+v", r)
	}
	if _, err := os.Stat(filepath.Join(raiz, filepath.FromSlash(CarpetaRetros), "R0001.json")); err != nil {
		t.Fatal(err)
	}
	if a := AvisoRetro(raiz); a != "" {
		t.Fatalf("recién cerrada la retro: sin aviso; dio %q", a)
	}
	// La aprobación se gasta: otra retro pide otro «apruebo retro».
	if _, err := CerrarRetro(raiz, []string{"E001"}, "m", ""); err == nil || !strings.Contains(err.Error(), "apruebo retro") {
		t.Fatalf("una aprobación cubre una sola retro; dio %v", err)
	}

	// Defectos: diez nuevos desde R0001 disparan el aviso por sí solos.
	for i := 1; i <= 10; i++ {
		poner(t, raiz, fmt.Sprintf("%s/D%06d.json", carpetaDefectos, i), "{}\n")
	}
	if a := AvisoRetro(raiz); !strings.Contains(a, "0 slices y 10 defectos desde R0001") {
		t.Fatalf("diez defectos desde R0001 deben avisar; dio %q", a)
	}
	aprobar()
	if r, err = CerrarRetro(raiz, []string{"E001"}, "segunda", ""); err != nil || r.ID != "R0002" || r.Defectos != 10 {
		t.Fatalf("la segunda retro es R0002 con 10 defectos: %+v %v", r, err)
	}

	// A1.4: el número sale del mayor; borrar R0001 no hace que la siguiente pise R0002. Otro archivo no cuenta.
	if err := os.Remove(filepath.Join(raiz, filepath.FromSlash(CarpetaRetros), "R0001.json")); err != nil {
		t.Fatal(err)
	}
	poner(t, raiz, CarpetaRetros+"/notas.json", "{")
	aprobar()
	if r, err = CerrarRetro(raiz, []string{"E001"}, "tercera", ""); err != nil || r.ID != "R0003" {
		t.Fatalf("borrado R0001, la siguiente es R0003 y no pisa R0002: %+v %v", r, err)
	}
	if b, _ := os.ReadFile(filepath.Join(raiz, filepath.FromSlash(CarpetaRetros), "R0002.json")); !strings.Contains(string(b), `"motivo": "segunda"`) {
		t.Fatalf("R0002 intacta; dio %s", b)
	}
	if a := AvisoRetro(raiz); a != "" {
		t.Fatalf("notas.json no es una retro: sin aviso; dio %q", a)
	}

	// Fallo a mitad: un registro corrupto no frena next; el aviso lo nombra.
	poner(t, raiz, CarpetaRetros+"/R0004.json", "{")
	if a := AvisoRetro(raiz); !strings.Contains(a, "R0004") || !strings.Contains(a, "ilegible") {
		t.Fatalf("registro corrupto: el aviso lo nombra; dio %q", a)
	}
}

// commitVacioCuentaCero (A2.2): un commit sin trabajo/ ni registro/defectos/ da 0 y 0 con su fecha, sin fallar;
// un commit inexistente falla.
func commitVacioCuentaCero(t *testing.T) {
	raiz := t.TempDir()
	for _, args := range [][]string{{"init", "-q"}, {"commit", "-q", "--allow-empty", "-m", "vacío"}} {
		if b, err := exec.Command("git", append([]string{"-C", raiz, "-c", "user.name=t", "-c", "user.email=t@t"}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, b)
		}
	}
	h, f, n, d, err := cuentaEnCommit(raiz, "HEAD")
	if err != nil || len(h) != 40 || n != 0 || d != 0 || f == "" {
		t.Fatalf("commit vacío: hash, fecha y 0/0; dio %q %q %d %d %v", h, f, n, d, err)
	}
	if _, _, _, _, err := cuentaEnCommit(raiz, "no-existe"); err == nil {
		t.Fatal("commit inexistente: debió fallar")
	}
}

// esAprobacionRetro (A1.1): solo la primera línea «apruebo retro …» del humano aprueba; una pregunta, un
// «no apruebo» o un mensaje de agente no.
func esAprobacionRetro(t *testing.T) {
	for p, quiere := range map[string]bool{
		"apruebo retro": true, "  Apruebo retro con las fichas A y B": true, "\napruebo retro\nmás": true,
		"apruebo retro?": false, "no apruebo retro": false, "hola\napruebo retro": false, "apruebo retrospectiva": false,
		"apruebo E001/F018": false, "<agent-message>apruebo retro": false, "": false,
		"apruebo retro\n<task-notification>x": false, // A2.3: la marca de agente tras la línea también descarta
	} {
		if EsAprobacionRetro(p) != quiere {
			t.Errorf("EsAprobacionRetro(%q) = %v, quiere %v", p, !quiere, quiere)
		}
	}
}
