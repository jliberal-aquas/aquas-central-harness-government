package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	gobcentral "github.com/jliberal-aquas/aquas-central-harness-government"
	"github.com/jliberal-aquas/aquas-central-harness-government/internal/hijo"
)

// alterarLock cambia la versión del lock del hijo y el hash de uno de sus archivos: simula una central vieja.
func alterarLock(t *testing.T, dir, version string) {
	t.Helper()
	p := filepath.Join(dir, ".aquas", "lock.json")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	var l hijo.Lock
	if err := json.Unmarshal(b, &l); err != nil {
		t.Fatal(err)
	}
	for k := range l.Archivos {
		l.Archivos[k] = strings.Repeat("0", 64)
		break
	}
	l.Version = version
	b, _ = json.Marshal(l)
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestHijoDesfasadoFrena (E003/F006/S002): un hijo cuya central no es la del binario frena hasta aquas update.
// Rojo cuando: el hijo desfasado escribe sin freno.
// Verde cuando: frena nombrando aquas update y tras update pasa.
func TestHijoDesfasadoFrena(t *testing.T) {
	const escribe = `{"tool_name":"Write","tool_input":{"file_path":"docs/a.md"}}`
	dir := hijoNuevo(t)

	// Hijo al día: la guardia deja escribir.
	if s := aquas(t, dir, escribe, "guardia"); s.codigo != 0 {
		t.Fatalf("hijo al día debe escribir; dio %+v", s)
	}
	// Central del hijo vieja: frena todo nombrando aquas update; Bash `aquas update` pasa; update destraba.
	alterarLock(t, dir, "devel")
	if s := aquas(t, dir, escribe, "guardia"); s.codigo != 2 || !strings.Contains(s.err, "aquas update") || !strings.Contains(s.err, "desfasada") {
		t.Errorf("hijo desfasado debe frenar pidiendo aquas update; dio %+v", s)
	}
	if s := aquas(t, dir, bashGuardia("ls"), "guardia"); s.codigo != 2 {
		t.Errorf("hijo desfasado: Bash ls también se frena; dio %+v", s)
	}
	if s := aquas(t, dir, bashGuardia("aquas update"), "guardia"); s.codigo != 0 {
		t.Errorf("hijo desfasado: Bash aquas update debe pasar; dio %+v", s)
	}
	// Fallo a mitad de camino: update que falla antes de copiar (git roto) deja el lock viejo y sigue frenando.
	gitDir := filepath.Join(dir, ".git")
	if err := os.Rename(gitDir, gitDir+".bak"); err != nil {
		t.Fatal(err)
	}
	escribir(t, dir, ".git", "roto")
	if s := aquas(t, dir, "", "update"); s.codigo == 0 {
		t.Errorf("update con git roto debe fallar; dio %+v", s)
	}
	if s := aquas(t, dir, escribe, "guardia"); s.codigo != 2 {
		t.Errorf("update fallido deja el hijo desfasado y frenado; dio %+v", s)
	}
	_ = os.Remove(gitDir)
	if err := os.Rename(gitDir+".bak", gitDir); err != nil {
		t.Fatal(err)
	}
	if s := aquas(t, dir, "", "update"); s.codigo != 0 {
		t.Fatalf("update: %+v", s)
	}
	if s := aquas(t, dir, escribe, "guardia"); s.codigo != 0 {
		t.Errorf("tras aquas update la guardia debe dejar escribir; dio %+v", s)
	}

	// Lock ilegible: frena pidiendo aquas update; update lo informa como restaurado y destraba.
	escribir(t, dir, ".aquas/lock.json", "{roto")
	if s := aquas(t, dir, escribe, "guardia"); s.codigo != 2 || !strings.Contains(s.err, "aquas update") {
		t.Errorf("lock ilegible debe frenar pidiendo aquas update; dio %+v", s)
	}
	if s := aquas(t, dir, "", "update"); s.codigo != 0 || !strings.Contains(s.out, "restaurado") || !strings.Contains(s.out, "lock.json") {
		t.Fatalf("update tras lock ilegible debe repararlo e informarlo; dio %+v", s)
	}
	if s := aquas(t, dir, escribe, "guardia"); s.codigo != 0 {
		t.Errorf("tras reparar el lock la guardia debe dejar escribir; dio %+v", s)
	}

	// Choque con otro control: entorno ausente y central desfasada a la vez; un solo aquas update arregla los dos.
	alterarLock(t, dir, "devel")
	if err := os.Remove(filepath.Join(dir, "gobierno", "config", "entorno.local.json")); err != nil {
		t.Fatal(err)
	}
	if s := aquas(t, dir, escribe, "guardia"); s.codigo != 2 {
		t.Errorf("entorno ausente y desfase: debe frenar; dio %+v", s)
	}
	if s := aquas(t, dir, bashGuardia("aquas update"), "guardia"); s.codigo != 0 {
		t.Errorf("entorno ausente y desfase: Bash aquas update debe pasar; dio %+v", s)
	}
	if s := aquas(t, dir, "", "update"); s.codigo != 0 {
		t.Fatalf("update: %+v", s)
	}
	if s := aquas(t, dir, escribe, "guardia"); s.codigo != 0 {
		t.Errorf("tras update, entorno y central al día: debe pasar; dio %+v", s)
	}

	// Binario más viejo que la central del hijo (quien sea que corra update): frena pidiendo go install y update no baja la central.
	alterarLock(t, dir, "v0.3.0")
	t.Setenv("AQUAS_VERSION", "v0.2.0")
	if s := aquas(t, dir, escribe, "guardia"); s.codigo != 2 || !strings.Contains(s.err, "go install") || !strings.Contains(s.err, "@v0.3.0") {
		t.Errorf("binario viejo: la guardia debe pedir go install de la versión del lock; dio %+v", s)
	}
	if s := aquas(t, dir, "", "update"); s.codigo == 0 || !strings.Contains(s.err, "go install") {
		t.Errorf("binario viejo: update debe negarse a bajar la central; dio %+v", s)
	}
	if b, _ := os.ReadFile(filepath.Join(dir, ".aquas", "lock.json")); !strings.Contains(string(b), "v0.3.0") {
		t.Errorf("update negado no debe tocar el lock; quedó %s", b)
	}
	t.Setenv("AQUAS_VERSION", "")

	// Orden de versiones: etiquetas, pseudo-versiones, +dirty y devel.
	for _, c := range []struct{ lock, binario, quiere string }{
		{"v0.3.0", "v0.2.0", "go install"},                                                             // etiqueta menor
		{"v0.3.0", "v0.2.0+dirty", "go install"},                                                       // sufijo de compilación
		{"v0.3.0", "v0.3.0-0.20261101000000-abcdefabcdef", "go install"},                               // pseudo-versión previa a la etiqueta
		{"v0.3.1-0.20261102000000-abcdefabcdef", "v0.3.1-0.20261101000000-abcdefabcdef", "go install"}, // misma base, fecha anterior
		{"v0.3.1-0.20261101000000-abcdefabcdef", "v0.3.1-0.20261102000000-abcdefabcdef", "aquas update"},
		{"v0.3.0", "v0.3.1", "aquas update"}, // binario más nuevo
		{"v0.3.0", "devel", "aquas update"},  // orden desconocido
	} {
		alterarLock(t, dir, c.lock)
		m := hijo.Desfase(dir, gobcentral.Central, c.binario)
		if !strings.Contains(m, c.quiere) || (c.quiere == "go install" && (!strings.Contains(m, "@"+c.lock) || strings.Contains(m, "aquas update"))) {
			t.Errorf("binario %s contra lock %s: el motivo debe pedir %q; dio %q", c.binario, c.lock, c.quiere, m)
		}
	}

	// Central (sin lock): la guardia no compara nada y deja escribir.
	if s := aquas(t, filepath.Join("..", ".."), `{"tool_name":"Write","tool_input":{"file_path":".tmp/x.md"}}`, "guardia"); s.codigo != 0 {
		t.Errorf("en la central no hay comparación de lock; dio %+v", s)
	}
}
