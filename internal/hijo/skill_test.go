package hijo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

// TestBuildEscribeSkillBacklog (E001/F007/S003): build deja /backlog en el hijo, sacado de la central sellada y
// apuntando al aquas instalado; update lo renueva; sin fuente no hay skill ni error.
// Rojo cuando: build no escribe el skill, deja el comando del central o Verificar no ve un SKILL.md editado.
// Verde cuando: el skill corre `aquas backlog table`, conserva el frontmatter y Verificar lo cuida como a CLAUDE.md.
func TestBuildEscribeSkillBacklog(t *testing.T) {
	canon, err := os.ReadFile("../../gobierno/canonico/skills/backlog.md")
	if err != nil {
		t.Fatal(err)
	}
	central := func(skill string) fstest.MapFS {
		m := fstest.MapFS{"gobierno/canonico/constitucion.md": {Data: []byte(constitucionConHuecos)}}
		if skill != "" {
			m[FuenteSkill] = &fstest.MapFile{Data: []byte(skill)}
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
	leer := func(dir string) string {
		b, _ := os.ReadFile(filepath.Join(dir, filepath.FromSlash(RutaSkill)))
		return string(b)
	}

	// Hijo recién instalado: build escribe el skill con el aquas instalado.
	dir := instalar(central(string(canon)))
	if err := Build(dir); err != nil {
		t.Fatal(err)
	}
	s := leer(dir)
	fm, resto, ok := strings.Cut(strings.TrimPrefix(s, "---\n"), "\n---\n")
	if !strings.HasPrefix(s, "---\n") || !ok || !strings.HasPrefix(resto, "<!-- GENERADO") {
		t.Fatalf("skill sin frontmatter primero y marca después:\n%s", s)
	}
	if !strings.Contains(fm, "disable-model-invocation: true") {
		t.Errorf("el skill del hijo debe seguir siendo solo del humano:\n%s", fm)
	}
	if !strings.Contains(resto, "!`cd \"${CLAUDE_PROJECT_DIR}\" && aquas backlog table`") || strings.Contains(s, "aquas.sh") {
		t.Errorf("el skill del hijo debe correr el aquas instalado:\n%s", s)
	}
	if m := Verificar(dir); len(m) != 0 {
		t.Fatalf("hijo recién armado: debió pasar; dio %v", m)
	}

	// Hallazgo A3.2: generado editado a mano que conserva la marca: Verificar lo ve y build lo rehace.
	ruta := filepath.Join(dir, filepath.FromSlash(RutaSkill))
	if err := os.WriteFile(ruta, []byte(s+"línea a mano\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if len(Verificar(dir)) == 0 {
		t.Errorf("generado editado con marca: Verificar debió avisar")
	}
	if err := Build(dir); err != nil || leer(dir) != s {
		t.Errorf("generado editado con marca: build debió rehacerlo: %v", err)
	}

	// SKILL.md reemplazado a mano (sin la marca): Verificar lo ve.
	if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(RutaSkill)), []byte("a mano\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if m := strings.Join(Verificar(dir), "\n"); !strings.Contains(m, "SKILL.md") {
		t.Errorf("SKILL.md editado a mano: Verificar debió avisar; dio %q", m)
	}
	// Sin la marca, build lo trata como propio (ver más abajo); se borra para seguir.
	if err := os.Remove(filepath.Join(dir, filepath.FromSlash(RutaSkill))); err != nil {
		t.Fatal(err)
	}
	// Update con una central nueva: build deja el skill nuevo.
	nuevo := strings.Replace(string(canon), "Muestra la tabla", "Muestra ya la tabla", 1)
	if _, err := Update(dir, central(nuevo), "v1.1.0"); err != nil {
		t.Fatal(err)
	}
	if err := Build(dir); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(leer(dir), "Muestra ya la tabla") {
		t.Errorf("tras update, build debió renovar el skill:\n%s", leer(dir))
	}

	// Central sin skill (anterior a S003): build sigue, sin skill; aparece tras update.
	viejo := instalar(central(""))
	if err := Build(viejo); err != nil {
		t.Fatalf("central sin skill: build no debe fallar: %v", err)
	}
	if leer(viejo) != "" || len(Verificar(viejo)) != 0 {
		t.Errorf("central sin skill: no debe haber skill ni avisos")
	}
	if _, err := Update(viejo, central(string(canon)), "v1.1.0"); err != nil {
		t.Fatal(err)
	}
	if err := Build(viejo); err != nil || !strings.Contains(leer(viejo), "aquas backlog table") {
		t.Errorf("tras update, el hijo viejo debió recibir /backlog: %v", err)
	}

	// Hallazgo A1.1: un /backlog propio del hijo no se pisa; build falla sin escribir.
	propio := instalar(central(string(canon)))
	p := filepath.Join(propio, filepath.FromSlash(RutaSkill))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("mi backlog propio\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Build(propio); err == nil || !strings.Contains(err.Error(), "no lo generó aquas") {
		t.Errorf("skill propio: build debió negarse; dio %v", err)
	}
	if leer(propio) != "mi backlog propio\n" {
		t.Errorf("skill propio: build lo pisó")
	}
	// Hallazgo A2.4: Verificar no pide build (build se negaría); da el mismo aviso.
	if m := strings.Join(Verificar(propio), "\n"); !strings.Contains(m, "no lo generó aquas") {
		t.Errorf("skill propio: Verificar debió pedir renombrarlo; dio %q", m)
	}
	// Hallazgo A2.1: install sobre un hijo con /backlog propio falla sin escribir nada.
	previo := t.TempDir()
	pp := filepath.Join(previo, filepath.FromSlash(RutaSkill))
	if err := os.MkdirAll(filepath.Dir(pp), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pp, []byte("mío\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Install(previo, central(string(canon)), "v1.0.0"); err == nil || !strings.Contains(err.Error(), "/backlog propio") {
		t.Errorf("install con skill propio: debió fallar; dio %v", err)
	}
	if _, err := os.Stat(filepath.Join(previo, Carpeta)); err == nil {
		t.Errorf("install con skill propio: no debió copiar la central")
	}
	// Central sin skill: un SKILL.md previo no es conflicto.
	if err := Install(previo, central(""), "v1.0.0"); err != nil {
		t.Errorf("central sin skill: install no debió fallar por un SKILL.md previo: %v", err)
	}

	// Hallazgo A2.2: el sello manda; un skill sellado alterado frena build.
	alterado := instalar(central(string(canon)))
	ps := filepath.Join(alterado, Carpeta, "central", filepath.FromSlash(FuenteSkill))
	if err := os.Chmod(ps, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ps, []byte(strings.Replace(string(canon), "Muestra", "Altera", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Build(alterado); err == nil || !strings.Contains(err.Error(), "sellado alterado") {
		t.Errorf("skill sellado alterado: build debió fallar por el sello; dio %v", err)
	}

	// Fuente CRLF: se normaliza; sale igual que la LF.
	crlf := instalar(central(strings.ReplaceAll(string(canon), "\n", "\r\n")))
	if err := Build(crlf); err != nil || !strings.Contains(leer(crlf), "aquas backlog table") || strings.Contains(leer(crlf), "\r") {
		t.Errorf("fuente CRLF: %v\n%q", err, leer(crlf))
	}
	// Fuente con el comando pero sin frontmatter: build falla.
	sinFM := instalar(central("# B\n!`cd x && sh .claude/hooks/aquas.sh backlog tabla`\n"))
	if err := Build(sinFM); err == nil || !strings.Contains(err.Error(), "frontmatter") {
		t.Errorf("fuente sin frontmatter: build debió fallar; dio %v", err)
	}

	// Fuente sin el comando del central: build falla y no escribe nada.
	roto := instalar(central("---\nname: backlog\n---\n!`otra cosa`\n"))
	if err := Build(roto); err == nil || !strings.Contains(err.Error(), "backlog") {
		t.Errorf("fuente sin el comando: build debió fallar nombrando el skill; dio %v", err)
	}
	if _, err := os.Stat(filepath.Join(roto, "CLAUDE.md")); err == nil || leer(roto) != "" {
		t.Errorf("build fallido no debe escribir CLAUDE.md ni SKILL.md")
	}
}
