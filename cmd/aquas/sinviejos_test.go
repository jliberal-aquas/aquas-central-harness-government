package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// reGenerado: bloque de una página rehecho desde trabajo/; es historia y no se migra.
var reGenerado = regexp.MustCompile(`(?s)<!-- GENERADO por aquas desde trabajo/.*?<!-- /GENERADO -->`)

// reLlamada: una llamada a aquas escrita en texto: el binario, con o sin comilla de cierre, y hasta siete palabras detrás.
var reLlamada = regexp.MustCompile(`(?:\baquas(?:\.exe|\.sh)?"?|\baq\.exe|\./cmd/aquas)[ \t]+([a-z][a-z-]*(?:[ \t]+[^\s` + "`" + `"'«»]+){0,6})`)

// reBarra: separador de alternativas con espacios alrededor.
var reBarra = regexp.MustCompile(`[ \t]*\|[ \t]*`)

// conNombreViejo: cada llamada a aquas del texto que todavía usa un nombre viejo, con su aviso.
func conNombreViejo(texto string) []string {
	var malas []string
	for _, m := range reLlamada.FindAllStringSubmatch(texto, -1) {
		// Cada palabra sin la puntuación final; «a|b|c» lista alternativas y se prueba cada una en su posición.
		pal := strings.Fields(reBarra.ReplaceAllString(m[1], "|")) // «a | b» = «a|b»
		for i := range pal {
			pal[i] = strings.TrimRight(pal[i], ").,;:")
		}
		vistos := map[string]bool{}
		for i, p := range pal {
			for _, alt := range strings.Split(p, "|") {
				args := append([]string(nil), pal...)
				for j := range args {
					args[j], _, _ = strings.Cut(args[j], "|")
				}
				args[i] = alt
				_, avisos := traducir(args)
				for _, a := range avisos {
					if !vistos[a] {
						vistos[a] = true
						malas = append(malas, strings.TrimSpace(m[0])+" ("+a+")")
					}
				}
			}
		}
	}
	return malas
}

// TestSinComandosViejos (E000/F006/S012, humano 2026-10-30, decisión 3a): la wiki, los canónicos, los
// compilados y los mensajes del programa nombran los comandos con su nombre en inglés.
// Rojo cuando: algún archivo vivo cita `aquas <nombre viejo>`.
// Verde cuando: ninguno lo cita, salvo las páginas que explican el cambio de nombre.
func TestSinComandosViejos(t *testing.T) {
	// El detector: encuentra el nombre viejo en cada posición y deja pasar el nuevo.
	casos := map[string]bool{
		"corre `aquas verificar`":                    true,
		"go run ./cmd/aquas compilar":                true,
		`sh .claude/hooks/aquas.sh backlog tabla`:    true,
		"./.tmp/aquas.exe memoria fijar tarea=x":     true,
		"aquas backlog alcance X ampliar a motivo=m": true,
		"corre `aquas verify`":                       false,
		"aquas backlog scope X extend a motivo=m":    false,
		"aquas memory set tarea=x":                   false,
		"el repo aquas-harness-governance":           false,
		"aquas compila los canónicos":                false,
		"(aquas memory fijar)":                       true,
		"`aquas backlog new|prioridad|state`":        true,
		"`aquas turn open|close`":                    false,
		"(`aquas backlog show | siguiente`)":         true,
		"(`aquas backlog show | next`)":              false,
		`sh "x/aquas.sh" turno abrir`:                true,
		`sh "x/aquas.sh" turn open`:                  false,
	}
	bloque := "a <!-- GENERADO por aquas desde trabajo/: x -->\n| aquas verificar |\n<!-- /GENERADO --> aquas verificar"
	if n := len(conNombreViejo(reGenerado.ReplaceAllString(bloque, ""))); n != 1 {
		t.Errorf("el bloque generado desde trabajo/ se salta y lo de fuera se revisa: %d citas; esperado 1", n)
	}
	for texto, viejo := range casos {
		if got := len(conNombreViejo(texto)) > 0; got != viejo {
			t.Errorf("%q: nombre viejo = %v; esperado %v", texto, got, viejo)
		}
	}

	// El repo: archivos vivos, sin historia ni pruebas ni las páginas del cambio de nombre.
	raiz := filepath.Join("..", "..")
	saltar := map[string]bool{".aquas": true, "registro": true, "auditoria": true, "trabajo": true, ".tmp": true, ".git": true, "testdata": true}
	explican := map[string]bool{"docs/E000/F006/S006.md": true, "docs/E000/F006/S012.md": true, "cmd/aquas/ingles.go": true}
	var malas []string
	err := filepath.WalkDir(raiz, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if saltar[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		rel, _ := filepath.Rel(raiz, p)
		rel = filepath.ToSlash(rel)
		ext := filepath.Ext(rel)
		if explican[rel] || strings.HasSuffix(rel, "_test.go") || (ext != ".md" && ext != ".go" && ext != ".json" && ext != ".sh" && rel != ".githooks/pre-commit") {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		for _, m := range conNombreViejo(reGenerado.ReplaceAllString(string(b), "")) {
			malas = append(malas, rel+": "+m)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(malas) > 0 {
		t.Errorf("%d llamadas con nombre viejo:\n%s", len(malas), strings.Join(malas, "\n"))
	}
}
