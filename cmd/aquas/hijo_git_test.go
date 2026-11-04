package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// E003/F006: un hijo nuevo queda funcionando al instalar, aunque la carpeta no tenga git.

// carpetaSinRepo: carpeta temporal donde git no ve el repo de la central que la contiene.
func carpetaSinRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(dir))
	return dir
}

// existe: la ruta existe dentro de dir.
func existe(dir, rel string) bool {
	_, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel)))
	return err == nil
}

// bashGuardia: entrada de la guardia para una orden Bash.
func bashGuardia(orden string) string {
	b, _ := json.Marshal(map[string]any{"tool_name": "Bash", "tool_input": map[string]string{"command": orden}})
	return string(b)
}

// TestHijoSinGitQuedaFuncionando (integral E003/F006): install en carpeta sin git deja el hijo con entorno y guardia abierta.
// Rojo cuando: install sin git deja la guardia cerrada.
// Verde cuando: install hace git init y la guardia deja escribir.
func TestHijoSinGitQuedaFuncionando(t *testing.T) {
	dir := carpetaSinRepo(t)
	if s := aquas(t, dir, "", "install"); s.codigo != 0 {
		t.Fatalf("install sin git termina bien; dio %+v", s)
	}
	if !existe(dir, ".git") || !existe(dir, "gobierno/config/entorno.local.json") {
		t.Fatal("install sin git deja .git y el entorno del hijo")
	}
	if s := aquas(t, dir, `{"tool_name":"Write","tool_input":{"file_path":"docs/a.md"}}`, "guardia"); s.codigo != 0 {
		t.Errorf("la guardia del hijo deja escribir dentro; dio %+v", s)
	}
}

// TestInstallIniciaGit (E003/F006/S001): install inicia git si falta; la guardia del hijo sin entorno pide aquas update.
// Rojo cuando: install sin git no crea .git.
// Verde cuando: crea .git, falla dentro de otro repo y la guardia nombra aquas update.
func TestInstallIniciaGit(t *testing.T) {
	guardiaHijoSinEntorno(t)

	// Carpeta con archivos y sin git: git init y entorno.
	dir := carpetaSinRepo(t)
	if err := os.WriteFile(filepath.Join(dir, "notas.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if s := aquas(t, dir, "", "install"); s.codigo != 0 || !existe(dir, ".git") || strings.Contains(s.out, "AVISO: entorno sin declarar") {
		t.Fatalf("carpeta con archivos: install inicia git sin aviso; dio %+v", s)
	}

	// Raíz git previa: igual que hoy.
	raizGit := carpetaSinRepo(t)
	git(t, raizGit, "init", "-q")
	if s := aquas(t, raizGit, "", "install"); s.codigo != 0 || !existe(raizGit, "gobierno/config/entorno.local.json") {
		t.Errorf("raíz git previa: install como hoy; dio %+v", s)
	}

	// Dentro de otro repo: falla sin copiar nada y nombra la raíz de ese repo.
	padre := carpetaSinRepo(t)
	git(t, padre, "init", "-q")
	anidado := filepath.Join(padre, "hijo")
	if err := os.MkdirAll(anidado, 0o755); err != nil {
		t.Fatal(err)
	}
	if s := aquas(t, anidado, "", "install"); s.codigo == 0 || !strings.Contains(s.err, "dentro del repo") || !strings.Contains(strings.ToLower(filepath.ToSlash(s.err)), strings.ToLower(filepath.ToSlash(padre))+":") || existe(anidado, ".aquas") {
		t.Errorf("dentro de otro repo: falla sin copiar y lo nombra; dio %+v", s)
	}

	// A1.2: hijo ya instalado sin git (el caso real de D000090): update inicia git y repone el entorno.
	viejo := carpetaSinRepo(t)
	if s := aquas(t, viejo, "", "install"); s.codigo != 0 {
		t.Fatalf("install: %+v", s)
	}
	for _, r := range []string{".git", "gobierno/config/entorno.local.json"} {
		if err := os.RemoveAll(filepath.Join(viejo, filepath.FromSlash(r))); err != nil {
			t.Fatal(err)
		}
	}
	if s := aquas(t, viejo, "", "update"); s.codigo != 0 || !existe(viejo, ".git") || !existe(viejo, "gobierno/config/entorno.local.json") {
		t.Errorf("update en hijo sin git inicia git y repone el entorno; dio %+v", s)
	}

	// .git roto: git falla a mitad de camino; install no copia la central.
	roto := carpetaSinRepo(t)
	if err := os.WriteFile(filepath.Join(roto, ".git"), []byte("basura"), 0o644); err != nil {
		t.Fatal(err)
	}
	if s := aquas(t, roto, "", "install"); s.codigo == 0 || !strings.Contains(s.err, "git no pudo leer la carpeta") || existe(roto, ".aquas") {
		t.Errorf(".git roto: install falla sin copiar; dio %+v", s)
	}

	// git ausente del PATH: falla nombrándolo, sin copiar.
	sinGit := carpetaSinRepo(t)
	t.Setenv("PATH", t.TempDir())
	if s := aquas(t, sinGit, "", "install"); s.codigo == 0 || !strings.Contains(s.err, "git no está en el PATH") || existe(sinGit, ".aquas") {
		t.Errorf("sin git en el PATH: falla nombrándolo; dio %+v", s)
	}
}

// guardiaHijoSinEntorno (S001): la guardia del hijo sin entorno pide aquas update y solo deja pasar esa orden.
func guardiaHijoSinEntorno(t *testing.T) {
	dir := carpetaSinRepo(t)
	if s := aquas(t, dir, "", "install"); s.codigo != 0 {
		t.Fatalf("install: %+v", s)
	}
	if err := os.Remove(filepath.Join(dir, "gobierno", "config", "entorno.local.json")); err != nil {
		t.Fatal(err)
	}
	s := aquas(t, dir, `{"tool_name":"Write","tool_input":{"file_path":"docs/a.md"}}`, "guardia")
	if s.codigo != 2 || !strings.Contains(s.err, "aquas update` por Bash") || strings.Contains(s.err, "go run ./cmd/aquas") {
		t.Errorf("hijo sin entorno: frena pidiendo aquas update, no el comando de la central; dio %+v", s)
	}
	if s := aquas(t, dir, bashGuardia("aquas update"), "guardia"); s.codigo != 0 {
		t.Errorf("el agente del hijo corre aquas update por Bash; dio %+v", s)
	}
	ps, _ := json.Marshal(map[string]any{"tool_name": "PowerShell", "tool_input": map[string]string{"command": "aquas update"}})
	if s := aquas(t, dir, string(ps), "guardia"); s.codigo != 2 || !strings.Contains(s.err, "por Bash") {
		t.Errorf("A1.5: por PowerShell se frena y pide Bash; dio %+v", s)
	}
	wr, _ := json.Marshal(map[string]any{"tool_name": "Write", "tool_input": map[string]string{"command": "aquas update", "file_path": "docs/a.md"}})
	if s := aquas(t, dir, string(wr), "guardia"); s.codigo != 2 {
		t.Errorf("A1.5: otra herramienta con el mismo texto se frena; dio %+v", s)
	}
	for _, o := range []string{"aquas update; rm x", "aquas update && ls", "aquas updatex", "ls", "aquas init"} {
		if s := aquas(t, dir, bashGuardia(o), "guardia"); s.codigo != 2 {
			t.Errorf("%q se frena en el hijo sin entorno; dio %+v", o, s)
		}
	}
	// aquas update repone el entorno y la guardia se abre.
	if s := aquas(t, dir, "", "update"); s.codigo != 0 {
		t.Fatalf("update: %+v", s)
	}
	if s := aquas(t, dir, `{"tool_name":"Write","tool_input":{"file_path":"docs/a.md"}}`, "guardia"); s.codigo != 0 {
		t.Errorf("tras update, la guardia deja escribir; dio %+v", s)
	}
	// En la central, el arreglo sigue siendo aquas init.
	central := carpetaSinRepo(t)
	git(t, central, "init", "-q")
	if s := aquas(t, central, bashGuardia("aquas update"), "guardia"); s.codigo != 2 || !strings.Contains(s.err, "init") {
		t.Errorf("central sin entorno: frena y pide init, también a aquas update; dio %+v", s)
	}
}
