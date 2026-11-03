package hijo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

// todosLosSkills (E001/F018/S001, subprueba de TestBuildEscribeSkillBacklog): cada skill de la central viaja al
// hijo, no solo el primero. Falla si build o Verificar miran solo /backlog, o un skill roto deja escrito parte del hijo.
func todosLosSkills(t *testing.T) {
	backlog := "---\nname: backlog\n---\n!`cd x && " + comandoCentral + "backlog table`\n"
	retro := "---\nname: retro\n---\n## Reparto\n"
	const rutaRetro = ".claude/skills/retro/SKILL.md"
	central := func(skills map[string]string) fstest.MapFS {
		m := fstest.MapFS{"gobierno/canonico/constitucion.md": {Data: []byte(constitucionConHuecos)}}
		for n, s := range skills {
			m[carpetaSkills+"/"+n+".md"] = &fstest.MapFile{Data: []byte(s)}
		}
		return m
	}
	instalar := func(c fstest.MapFS) string {
		t.Helper()
		dir := filepath.Join(t.TempDir(), "hijo")
		if err := Install(dir, c, "v1.0.0"); err != nil {
			t.Fatal(err)
		}
		p := filepath.Join(dir, CarpetaHuecos, "mision.md")
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("- Landing.\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		return dir
	}
	leer := func(dir, ruta string) string {
		b, _ := os.ReadFile(filepath.Join(dir, filepath.FromSlash(ruta)))
		return string(b)
	}
	escribir := func(dir, ruta, txt string) {
		t.Helper()
		p := filepath.Join(dir, filepath.FromSlash(ruta))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(txt), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	ambos := map[string]string{"backlog": backlog, "retro": retro}

	// Build escribe los dos; Verificar ve un /retro desfasado aunque /backlog esté al día.
	dir := instalar(central(ambos))
	if err := Build(dir); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(leer(dir, RutaSkill), "aquas backlog table") || !strings.Contains(leer(dir, rutaRetro), "## Reparto") {
		t.Fatalf("build debió escribir /backlog y /retro")
	}
	if m := Verificar(dir); len(m) != 0 {
		t.Fatalf("recién armado: debió pasar; dio %v", m)
	}
	escribir(dir, rutaRetro, leer(dir, rutaRetro)+"a mano\n")
	if m := strings.Join(Verificar(dir), "\n"); !strings.Contains(m, rutaRetro) {
		t.Errorf("/retro editado: Verificar debió nombrarlo; dio %q", m)
	}

	// Hijo viejo con /retro propio: update trae /retro de la central; build se niega y no lo pisa.
	viejo := instalar(central(map[string]string{"backlog": backlog}))
	escribir(viejo, rutaRetro, "mi retro\n")
	if _, err := Update(viejo, central(ambos), "v1.1.0"); err != nil {
		t.Fatal(err)
	}
	if err := Build(viejo); err == nil || !strings.Contains(err.Error(), "/retro propio") {
		t.Errorf("/retro propio: build debió negarse; dio %v", err)
	}
	if leer(viejo, rutaRetro) != "mi retro\n" {
		t.Errorf("/retro propio: build lo pisó")
	}
	if m := strings.Join(Verificar(viejo), "\n"); !strings.Contains(m, "/retro propio") {
		t.Errorf("/retro propio: Verificar debió pedir renombrarlo; dio %q", m)
	}

	// Fallo a mitad: el segundo skill roto (sin frontmatter) frena build sin escribir nada.
	roto := instalar(central(map[string]string{"backlog": backlog, "retro": "## sin frontmatter\n"}))
	if err := Build(roto); err == nil || !strings.Contains(err.Error(), "/retro") {
		t.Errorf("/retro roto: build debió fallar nombrándolo; dio %v", err)
	}
	if leer(roto, "CLAUDE.md") != "" || leer(roto, RutaSkill) != "" {
		t.Errorf("/retro roto: build no debió escribir CLAUDE.md ni /backlog")
	}

	// Cada línea que corre un comando usa el del central, aunque otra línea sí lo use (hallazgo A1.7).
	mixto := instalar(central(map[string]string{"backlog": backlog + "!`otra cosa`\n"}))
	if err := Build(mixto); err == nil || !strings.Contains(err.Error(), "/backlog") {
		t.Errorf("comando ajeno: build debió fallar nombrando el skill; dio %v", err)
	}
}
