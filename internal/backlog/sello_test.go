package backlog

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// E001/F002/S004: sello de cierre. Antes de cerrar un slice, aquas arma la lista de lo que
// puede haber quedado desfasado (árbol, dependientes, archivos que nombran lo tocado) y el
// cierre se niega hasta que cada ítem tenga marca y motivo.

func git(t *testing.T, raiz string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", raiz, "-c", "user.name=t", "-c", "user.email=t@t"}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func poner(t *testing.T, raiz, rel, c string) {
	t.Helper()
	p := filepath.Join(raiz, filepath.FromSlash(rel))
	os.MkdirAll(filepath.Dir(p), 0o755)
	if err := os.WriteFile(p, []byte(c), 0o644); err != nil {
		t.Fatal(err)
	}
}

func fichaJSON(t *testing.T, raiz string, f Ficha) {
	t.Helper()
	f.Tipo, f.Padre = tipoYPadre(f.ID)
	if f.BloqueadoPor == nil {
		f.BloqueadoPor = []string{}
	}
	f.Prioridad, f.Motivo = "P1", "m"
	b, _ := json.MarshalIndent(f, "", "  ")
	poner(t, raiz, Activo+"/"+nombre(f.ID), string(b)+"\n")
}

// repoSello: repo Git con S004 en curso (rojo y verde registrados en el commit base), un
// dependiente S005, un hermano independiente S006, y archivos que nombran a internal/a/a.go
// y a cmd/x/main.go. Después del rojo se toca internal/a/a.go y se crea internal/b/b.go.
func repoSello(t *testing.T) string {
	t.Helper()
	raiz := t.TempDir()
	git(t, raiz, "init", "-q")
	poner(t, raiz, "go.mod", "module ejemplo.com/m\n\ngo 1.22\n")
	poner(t, raiz, "internal/a/a.go", "package a\n\nfunc A() int { return 1 }\n")
	poner(t, raiz, "cmd/x/main.go", "package main\n\nimport \"ejemplo.com/m/internal/a\"\n\nfunc main() { _ = a.A() }\n")
	poner(t, raiz, "docs/ref.md", "# Ref\n\nEl código vive en `internal/a/a.go`.\nEl CLI en `cmd/x/main.go`.\n")
	poner(t, raiz, "docs/otra.md", "# Otra\n\nNo nombra nada.\n")
	poner(t, raiz, "registro/memoria/x.json", "{\"nota\": \"internal/a/a.go\"}\n")
	git(t, raiz, "add", "-A")
	git(t, raiz, "commit", "-q", "-m", "base")
	reg := &Registro{Fecha: "2026-10-21T10:00:00Z", Head: git(t, raiz, "rev-parse", "HEAD"), Resultado: map[string]string{"./a:TestA": "falló"}}
	fichaJSON(t, raiz, Ficha{ID: "E001", Titulo: "I", Estado: "EN_CURSO", Secuencia: 1})
	fichaJSON(t, raiz, Ficha{ID: "E001/F002", Titulo: "F", Estado: "EN_CURSO", Secuencia: 2})
	fichaJSON(t, raiz, Ficha{ID: "E001/F002/S004", Titulo: "S", Estado: "EN_CURSO", Secuencia: 3,
		Pruebas: []Prueba{{Nombre: "./a:TestA", Intencion: "x", RojoCuando: "r", VerdeCuando: "v"}}, Rojo: reg, Verde: reg})
	fichaJSON(t, raiz, Ficha{ID: "E001/F002/S005", Titulo: "D", Estado: "BORRADOR", Secuencia: 4, BloqueadoPor: []string{"E001/F002/S004"}})
	fichaJSON(t, raiz, Ficha{ID: "E001/F002/S006", Titulo: "H", Estado: "BORRADOR", Secuencia: 5})
	poner(t, raiz, "internal/a/a.go", "package a\n\nfunc A() int { return 2 }\n")
	poner(t, raiz, "internal/b/b.go", "package b\n")
	return raiz
}

func claves(t *testing.T, raiz, id string) map[string]Item {
	t.Helper()
	xs, err := Sello(raiz, id)
	if err != nil {
		t.Fatal(err)
	}
	m := map[string]Item{}
	for _, x := range xs {
		m[x.Clave] = x
	}
	return m
}

func TestSelloArmaListaDelArbol(t *testing.T) {
	raiz := repoSello(t)
	m := claves(t, raiz, "E001/F002/S004")
	for _, k := range []string{"ficha:E001/F002", "ficha:E001", "ficha:E001/F002/S005",
		"pagina:docs/E001/F002/S004.md", "pagina:docs/E001/F002/README.md", "pagina:docs/E001/README.md"} {
		if _, ok := m[k]; !ok {
			t.Errorf("falta %s en el sello; hay %v", k, m)
		}
	}
	if _, ok := m["ficha:E001/F002/S006"]; ok {
		t.Errorf("un hermano sin dependencia ni archivos compartidos no entra")
	}
}

func TestSelloArmaListaDeArchivos(t *testing.T) {
	raiz := repoSello(t)
	m := claves(t, raiz, "E001/F002/S004")
	if it, ok := m["nombra:docs/ref.md>internal/a/a.go"]; !ok || !strings.Contains(it.Linea, "El código vive en `internal/a/a.go`.") {
		t.Errorf("docs/ref.md nombra a un tocado: ítem con la línea relevante; dio %+v (ok=%v)", it, ok)
	}
	if _, ok := m["importa:cmd/x/main.go>internal/a"]; !ok {
		t.Errorf("cmd/x/main.go importa el paquete tocado: ítem; hay %v", m)
	}
	for k := range m {
		if strings.Contains(k, "registro/") || strings.Contains(k, "trabajo/") || strings.Contains(k, "docs/otra.md") || strings.Contains(k, ">cmd/x/main.go") {
			t.Errorf("sobra %s: datos excluidos, archivo que no nombra o archivo no tocado", k)
		}
	}
}

func marcarTodo(t *testing.T, raiz, id string) {
	t.Helper()
	xs, err := Sello(raiz, id)
	if err != nil {
		t.Fatal(err)
	}
	for _, x := range xs {
		if _, err := Marcar(raiz, id, x.Clave, "sin_cambio", "revisado"); err != nil {
			t.Fatal(err)
		}
	}
}

func TestCierreSinSelloCompletoSeNiega(t *testing.T) {
	raiz := repoSello(t)
	id := "E001/F002/S004"
	if _, err := Estado(raiz, id, "CERRADO"); err == nil || !strings.Contains(err.Error(), "sello incompleto") {
		t.Fatalf("sin marcas: CERRADO se niega por el sello; dio %v", err)
	}
	if _, err := Marcar(raiz, id, "ficha:E001", "sin_cambio", ""); err == nil {
		t.Errorf("marca sin motivo: error")
	}
	if _, err := Marcar(raiz, id, "ficha:E001", "listo", "x"); err == nil {
		t.Errorf("marca que no es sin_cambio ni actualizado: error")
	}
	if _, err := Marcar(raiz, id, "ficha:E009", "sin_cambio", "x"); err == nil {
		t.Errorf("clave que no está en el sello: error")
	}
	marcarTodo(t, raiz, id)
	if _, err := Marcar(raiz, id, "ficha:E001", "actualizado", "cambié su estado"); err != nil {
		t.Fatal(err)
	}
	if f, err := Estado(raiz, id, "CERRADO"); err != nil || f.Estado != "CERRADO" {
		t.Fatalf("con todo marcado cierra; dio %v", err)
	}
	f, _, _ := Ver(raiz, id)
	for _, s := range f.Sello {
		if s.Clave == "ficha:E001" && (s.Marca != "actualizado" || s.Motivo != "cambié su estado") {
			t.Errorf("la ficha archivada guarda la última marca de cada ítem; dio %+v", s)
		}
	}
}

func TestSelloSeVuelveAArmarSiCambiaAlgo(t *testing.T) {
	raiz := repoSello(t)
	id := "E001/F002/S004"
	marcarTodo(t, raiz, id)
	poner(t, raiz, "cmd/x/main.go", "package main\n\nfunc main() {}\n")
	_, err := Estado(raiz, id, "CERRADO")
	if err == nil || !strings.Contains(err.Error(), "nombra:docs/ref.md>cmd/x/main.go") {
		t.Fatalf("tocar otro archivo agrega ítems sin marca y el cierre los nombra; dio %v", err)
	}
}

// E001/F002/S014: los derivados del compilador (salidas del sello del compilador y el archivo
// de sellos) no cuentan como tocados: ya los vigila el gate de derivados. Lo demás sigue igual.
func TestSelloIgnoraDerivados(t *testing.T) {
	raiz := repoSello(t)
	poner(t, raiz, "gobierno/config/sellos.json", "{\"salidas\": {\"CLAUDE.md\": {\"sha256\": \"x\"}}}\n")
	poner(t, raiz, "CLAUDE.md", "# Constitución v1\n")
	poner(t, raiz, "docs/derivados.md", "Ver `CLAUDE.md` y `gobierno/config/sellos.json`.\n")
	git(t, raiz, "add", "-A")
	git(t, raiz, "commit", "-q", "-m", "derivados")
	// La ficha apunta al commit nuevo: lo de repoSello vuelve a tocarse después.
	f := Ficha{ID: "E001/F002/S004", Titulo: "S", Estado: "EN_CURSO", Secuencia: 3, Pruebas: []Prueba{{Nombre: "./a:TestA", Intencion: "x", RojoCuando: "r", VerdeCuando: "v"}}}
	reg := &Registro{Fecha: "2026-10-21T10:00:00Z", Head: git(t, raiz, "rev-parse", "HEAD"), Resultado: map[string]string{"./a:TestA": "falló"}}
	f.Rojo, f.Verde = reg, reg
	fichaJSON(t, raiz, f)
	poner(t, raiz, "CLAUDE.md", "# Constitución v2\n")
	poner(t, raiz, "gobierno/config/sellos.json", "{\"salidas\": {\"CLAUDE.md\": {\"sha256\": \"y\"}}}\n")
	poner(t, raiz, "internal/a/a.go", "package a\n\nfunc A() int { return 3 }\n")
	m := claves(t, raiz, "E001/F002/S004")
	for k := range m {
		if strings.HasSuffix(k, ">CLAUDE.md") || strings.HasSuffix(k, ">gobierno/config/sellos.json") {
			t.Errorf("un derivado del compilador no cuenta como tocado; sobra %s", k)
		}
	}
	if _, ok := m["nombra:docs/ref.md>internal/a/a.go"]; !ok {
		t.Errorf("lo que no es derivado sigue en el sello; hay %v", m)
	}
}

// E001/F002/S013: sin commit base (proyecto hijo recién creado, sin commits) el árbol se exige
// igual; solo se omite el nivel archivos, y un aviso que también se marca lo deja a la vista.
func TestSelloSinCommitBaseExigeArbol(t *testing.T) {
	raiz := t.TempDir()
	id := "E001/F002/S004"
	reg := &Registro{Fecha: "2026-10-21T10:00:00Z", Resultado: map[string]string{"./a:TestA": "falló"}}
	fichaJSON(t, raiz, Ficha{ID: "E001", Titulo: "I", Estado: "EN_CURSO", Secuencia: 1})
	fichaJSON(t, raiz, Ficha{ID: "E001/F002", Titulo: "F", Estado: "EN_CURSO", Secuencia: 2})
	fichaJSON(t, raiz, Ficha{ID: id, Titulo: "S", Estado: "EN_CURSO", Secuencia: 3,
		Pruebas: []Prueba{{Nombre: "./a:TestA", Intencion: "x", RojoCuando: "r", VerdeCuando: "v"}}, Rojo: reg, Verde: reg})
	m := claves(t, raiz, id)
	for _, k := range []string{"ficha:E001/F002", "ficha:E001", "pagina:docs/E001/F002/S004.md", "aviso:sin-commit-base"} {
		if _, ok := m[k]; !ok {
			t.Errorf("sin commit base falta %s; hay %v", k, m)
		}
	}
	if it := m["aviso:sin-commit-base"]; !strings.Contains(it.Linea, "nivel archivos omitido") {
		t.Errorf("el aviso dice qué se omitió; dio %q", it.Linea)
	}
	if _, err := Estado(raiz, id, "CERRADO"); err == nil || !strings.Contains(err.Error(), "sello incompleto") {
		t.Fatalf("sin commit base el cierre igual exige el sello; dio %v", err)
	}
	marcarTodo(t, raiz, id)
	if _, err := Estado(raiz, id, "CERRADO"); err != nil {
		t.Fatalf("con el árbol y el aviso marcados cierra; dio %v", err)
	}
}
