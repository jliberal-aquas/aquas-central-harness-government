package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// copiarFeatureCerrada copia al repo de prueba la feature archivada E003/F001 y sus slices con otro id;
// cambiar, si no es nil, decide el estado de cada slice copiado (por su número 1…5).
func copiarFeatureCerrada(t *testing.T, raiz, id string, conSlices bool, cambiar func(n int) string) {
	t.Helper()
	nombres := []string{"E003-F001.json"}
	if conSlices {
		for _, n := range []string{"1", "2", "3", "4", "5"} {
			nombres = append(nombres, "E003-F001-S00"+n+".json")
		}
	}
	for i, nombre := range nombres {
		b, err := os.ReadFile(filepath.Join("..", "..", "trabajo", "archivo", nombre))
		if err != nil {
			t.Fatal(err)
		}
		txt := strings.ReplaceAll(string(b), "E003/F001", id)
		if i > 0 && cambiar != nil {
			txt = strings.Replace(txt, `"CERRADO"`, `"`+cambiar(i)+`"`, 1)
		}
		destino := strings.ReplaceAll(id, "/", "-") + strings.TrimPrefix(nombre, "E003-F001")
		escribir(t, raiz, "trabajo/archivo/"+destino, txt)
	}
}

// TestFrenoSigueAlTrabajoActivo (E001/F017/S001, D000091, D000101): solo suelta el tope una feature CERRADO con todos
// sus slices CERRADO; activa, cancelada, sin slices, con un slice cancelado o sin ficha siguen frenando; los
// comandos exentos pasan con el tope agotado.
// Rojo cuando: el apruebo de una feature cerrada frena o un comando exento se bloquea con el tope agotado.
// Verde cuando: solo frena el tope de una feature no cerrada y los exentos pasan siempre.
func TestFrenoSigueAlTrabajoActivo(t *testing.T) {
	raiz := repoFeature(t)
	copiarFeatureCerrada(t, raiz, "E003/F001", true, nil)
	copiarFeatureCerrada(t, raiz, "E004/F001", true, func(n int) string { // cancelar un slice no pide auditoría
		if n == 3 {
			return "CANCELADO"
		}
		return "CERRADO"
	})
	copiarFeatureCerrada(t, raiz, "E005/F001", false, nil) // feature CERRADO sin slices
	b, _ := os.ReadFile(filepath.Join(raiz, "trabajo", "archivo", "E005-F001.json"))
	escribir(t, raiz, "trabajo/archivo/E006-F001.json",
		strings.Replace(strings.ReplaceAll(string(b), "E005/F001", "E006/F001"), `"CERRADO"`, `"CANCELADO"`, 1))
	reg := registroSesion(t)
	apruebo := func(texto string) {
		t.Helper()
		time.Sleep(1100 * time.Millisecond) // la hora del apruebo va en segundos
		anexar(t, reg, lineaHumana(texto))
	}
	usar := func(comando string) salida {
		t.Helper()
		return aquas(t, raiz, usoComando(t, "Bash", comando), "budget", "tool")
	}
	frena := func(id string) {
		t.Helper()
		apruebo("apruebo " + id + " tope 1")
		if s := usar("git status"); s.codigo != 0 {
			t.Fatalf("%s: la primera llamada está dentro del tope; dio %+v", id, s)
		}
		if s := usar("git status"); s.codigo != 2 || !strings.Contains(s.err, "tope de la feature "+id) {
			t.Errorf("%s: la segunda llamada pasa el tope 1 y frena; dio %+v", id, s)
		}
	}

	// Feature CERRADO con todos sus slices CERRADO: su tope viejo no frena.
	apruebo("apruebo E003/F001 tope 1")
	for i := range 3 {
		if s := usar("git status"); s.codigo != 0 {
			t.Fatalf("el tope de una feature CERRADO no frena; llamada %d dio %+v", i+1, s)
		}
	}
	// Activa (BORRADOR), slice CANCELADO, sin slices, feature CANCELADO o sin ficha: el tope rige (hallazgo A1.2).
	for _, id := range []string{"E001/F002", "E004/F001", "E005/F001", "E006/F001", "E009/F009"} {
		frena(id)
	}
	// Ficha ilegible en el backlog: no se inventa salida, el tope de la feature cerrada vuelve a regir.
	escribir(t, raiz, "trabajo/activo/E007-F001.json", "{roto")
	frena("E003/F001")

	// Tope agotado (E003/F001 con backlog ilegible): los exentos pasan, también con cd a la raíz y comillas.
	raizBash := filepath.ToSlash(raiz)
	exentos := []string{
		`aquas memory set tarea=x`,
		`./.tmp/aquas.exe memory set "tarea=dos palabras" 'siguiente=otra cosa'`,
		`cd "` + raizBash + `" && ./.tmp/aquas.exe memory set "tarea=E001/F017: S001 en curso"`,
		`aquas defect occur D000101 T000598`,
		`./.tmp/aquas.exe defect new firma=x "que=algo pasó" relacion=E001/F017`,
		"aquas memory set tarea=x\n", // un salto final no trae otro comando
	}
	if runtime.GOOS == "windows" && len(raizBash) > 2 && raizBash[1] == ':' { // forma de Git Bash: /d/… (hallazgo A2.4)
		exentos = append(exentos, `cd "/`+strings.ToLower(raizBash[:1])+raizBash[2:]+`" && aquas memory set a=b`)
	}
	for _, c := range exentos {
		if s := usar(c); s.codigo != 0 {
			t.Errorf("exento con el tope agotado debe pasar: %q dio %+v", c, s)
		}
	}
	noExentos := []string{
		`cd /otro && aquas memory set tarea=x`,
		`aquas memory set tarea=x; rm -rf docs`,
		`aquas memory set "tarea=$(whoami)"`,
		`aquas defect show D000101`,
		"aquas defect occur D000101 T000598\ngit push", // D000105
		"aquas\nmemory set tarea=x",                    // hallazgo A1.3: salto de línea en cualquier hueco
		"aquas memory\nset tarea=x",
		"aquas defect\nnew firma=x",
		"go run ./cmd/aquas\ndefect new firma=x",
		"cd\n\"" + raizBash + "\" && aquas memory set a=b",
		"cd \"" + raizBash + "\"\n&& aquas memory set a=b",
	}
	for _, c := range noExentos {
		if s := usar(c); s.codigo != 2 {
			t.Errorf("no exento con el tope agotado debe frenar: %q dio %+v", c, s)
		}
	}

	// Las formas exentas tampoco suman a la cuenta del registro (hallazgo A2.5): con tope 1, tras cada exento
	// la primera llamada común pasa y la segunda frena.
	frena("E001/F002")
	apruebo("apruebo E001/F002 tope 1")
	for _, c := range exentos {
		_ = usar(c)
	}
	if s := usar("git status"); s.codigo != 0 {
		t.Errorf("los exentos no cuentan: la primera llamada común está dentro del tope 1; dio %+v", s)
	}
}
