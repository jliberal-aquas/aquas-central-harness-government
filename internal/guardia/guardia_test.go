package guardia

import (
	"path/filepath"
	"strings"
	"testing"
)

func raizPrueba(t *testing.T) string {
	t.Helper()
	r, err := filepath.Abs(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(r, "repo")
}

func TestArchivoDentroPasa(t *testing.T) {
	raiz := raizPrueba(t)
	for _, p := range []string{
		filepath.Join(raiz, "docs", "x.md"),
		"docs/x.md",
		filepath.Join(raiz, ".tmp", "a.py"),
		strings.ToUpper(raiz[:1]) + raiz[1:] + "/docs/y.md", // mayúsculas de unidad distintas
	} {
		if m := RevisarArchivo(raiz, p); m != "" {
			t.Errorf("%s dentro del repo debe pasar: %s", p, m)
		}
	}
}

func TestArchivoFueraFalla(t *testing.T) {
	raiz := raizPrueba(t)
	for _, p := range []string{
		filepath.Join(filepath.Dir(raiz), "otro.md"),
		filepath.Join(raiz, "..", "otro.md"),
		filepath.Join(raiz+"-gemelo", "x.md"), // prefijo igual, carpeta distinta
		"../fuera.md",
		"",
	} {
		if m := RevisarArchivo(raiz, p); !strings.Contains(m, "fuera del repositorio") {
			t.Errorf("%q fuera del repo debe fallar; dio %q", p, m)
		}
	}
}

func TestBashDentroPasa(t *testing.T) {
	raiz := raizPrueba(t)
	for _, c := range []string{
		`go test ./...`,
		`cat ~/.claude/settings.json`,              // leer fuera está permitido
		`sed -n 1,5p /c/Users/x/a.md && ls ~`,      // leer fuera está permitido
		`echo hola > .tmp/x.txt`,                   // escribir dentro
		`cat > docs/a.md <<'EOF'` + "\nhola\nEOF",  // heredoc dentro
		`go vet ./... 2>&1 | tail -3`,              // 2>&1 no es archivo
		`go run ./cmd/aquas verificar > /dev/null`, // /dev/null permitido
		`f=docs/x.md && sed -i 's/a/b/' $f`,        // variable asignada en el comando
		`cp /c/Users/x/plantilla.md docs/nueva.md`, // copiar desde fuera hacia dentro
		`mkdir -p .tmp && rm -rf .tmp/integral`,    // borrar dentro
		`git commit -q -F - <<'EOF'` + "\nmsg > no es redirección dentro del heredoc\nEOF",
	} {
		if m := RevisarBash(raiz, raiz, c); m != "" {
			t.Errorf("%q debe pasar: %s", c, m)
		}
	}
}

func TestBashFueraFalla(t *testing.T) {
	raiz := raizPrueba(t)
	for _, c := range []string{
		`echo x > ~/nota.txt`,
		`echo x >> /c/Users/jlibe/nota.txt`,
		`echo x > C:/Users/jlibe/nota.txt`,
		`echo x > /tmp/a`,
		`echo x > ../fuera.txt`,
		`echo x>$HOME/a`,
		`cat a | tee ~/.claude/x`,
		`cp docs/a.md ~/a.md`,
		`mv ~/.claude/x docs/x`, // mover desde fuera borra el original fuera
		`rm -rf /c/Users/jlibe/.claude/hooks`,
		`mkdir -p "$TEMP/x"`,
		`touch %USERPROFILE%/x`,
		`sed -i 's/a/b/' ~/.claude/hooks/reglas.py`,
		`cd ~/.claude && echo x > a.txt`, // relativo tras cd fuera
		`cd .. && rm x`,
		`git clone -q . ../copia`,
		`curl -s -o ~/x.bin https://example.com`,
		`echo x > "$DESTINO"`, // variable desconocida: no se puede probar que esté dentro
	} {
		if m := RevisarBash(raiz, raiz, c); !strings.Contains(m, "fuera del repositorio") && !strings.Contains(m, "no se puede resolver") {
			t.Errorf("%q debe fallar; dio %q", c, m)
		}
	}
}

func TestBashConCwdFuera(t *testing.T) {
	raiz := raizPrueba(t)
	// E000/F005/S002: bloquea por el destino relativo al cwd de fuera, no por otra causa.
	if m := RevisarBash(raiz, filepath.Dir(raiz), `echo x > a.txt`); !strings.Contains(m, "escritura fuera del repositorio bloqueada: a.txt") {
		t.Fatalf("con cwd fuera del repo, escribir relativo debe fallar por a.txt; obtuve %q", m)
	}
}

func TestEntradaHook(t *testing.T) {
	raiz := raizPrueba(t)
	// E000/F005/S002: cada caso que falla lo hace por su motivo (espera), no por cualquiera.
	casos := []struct{ json, espera string }{
		{`{"tool_name":"Write","tool_input":{"file_path":"docs/a.md"}}`, ""},
		{`{"tool_name":"Edit","tool_input":{"file_path":"../a.md"}}`, "bloqueada: ../a.md"},
		{`{"tool_name":"NotebookEdit","tool_input":{"notebook_path":"../n.ipynb"}}`, "bloqueada: ../n.ipynb"},
		{`{"tool_name":"Bash","tool_input":{"command":"echo x > ~/a"}}`, "bloqueada: ~/a"},
		{`{"tool_name":"Read","tool_input":{"file_path":"C:/fuera.md"}}`, ""}, // leer no es escribir
		{`no es json`, "entrada del hook ilegible"},
	}
	for _, c := range casos {
		m := Revisar(raiz, []byte(c.json))
		if c.espera == "" && m != "" || c.espera != "" && !strings.Contains(m, c.espera) {
			t.Errorf("%s: se esperaba %q; motivo %q", c.json, c.espera, m)
		}
	}
}

func TestPowerShellSeBloquea(t *testing.T) {
	raiz := t.TempDir()
	e := `{"tool_name":"PowerShell","tool_input":{"command":"Set-Content x.txt hola"}}`
	if m := Revisar(raiz, []byte(e)); !strings.Contains(m, "PowerShell") {
		t.Fatalf("PowerShell no se puede analizar: debe bloquearse; obtuve %q", m)
	}
}
