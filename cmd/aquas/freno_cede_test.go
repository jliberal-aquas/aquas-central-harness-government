package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestFrenoTurnoCedeConFeatureAprobada (E001/F017/S009, D000109): con una feature aprobada y abierta solo cuenta su
// tope. E001/F019/S002 quitó el freno de 60 llamadas por mensaje: sin apruebo, o con la feature cerrada, la 61 pasa.
// Rojo cuando: una feature aprobada con tope vigente frena a las 61 llamadas del turno.
// Verde cuando: con apruebo vigente solo frena el tope y sin apruebo la 61 pasa.
func TestFrenoTurnoCedeConFeatureAprobada(t *testing.T) {
	raiz := repoFeature(t)
	if s := aquas(t, raiz, "", "memory", "set", "iniciativa=E001", "feature=F002"); s.codigo != 0 {
		t.Fatalf("memory set: %+v", s)
	}
	turno := func(prompt string) {
		t.Helper()
		if s := aquas(t, raiz, mensajeHumano(t, prompt), "approval"); s.codigo != 0 {
			t.Fatalf("aprobacion: %+v", s)
		}
	}
	usar := func(agente string) salida {
		t.Helper()
		return aquas(t, raiz, usoHerramienta(t, agente, registroSesion(t)), "budget", "tool")
	}
	llenar := func(caso string) {
		t.Helper()
		for i := range 60 {
			if s := usar(""); s.codigo != 0 {
				t.Fatalf("%s: llamada %d dio %+v", caso, i+1, s)
			}
		}
	}

	// Sin apruebo de feature (uno de slice o de iniciativa no cuenta): no hay tope ni freno por mensaje.
	mensajeHumano(t, "apruebo E001/F002/S001")
	mensajeHumano(t, "apruebo E001") // ni uno de iniciativa (A2.1)
	llenar("sin apruebo")
	if s := usar(""); s.codigo != 0 {
		t.Errorf("sin apruebo, la llamada 61 del turno pasa (E001/F019/S002); dio %+v", s)
	}

	// Apruebo vigente: el turno no corta; el tope de la feature sigue frenando.
	time.Sleep(1100 * time.Millisecond) // la fecha de aprobación va en segundos
	turno("apruebo E001/F002 tope 65")
	llenar("con apruebo")
	for i := range 5 {
		if s := usar(""); s.codigo != 0 {
			t.Errorf("con apruebo vigente la llamada %d del turno pasa; dio %+v", 61+i, s)
		}
	}
	if s := usar(""); s.codigo != 2 || !strings.Contains(s.err, "tope de la feature") {
		t.Errorf("con apruebo, la llamada 66 pasa el tope 65 de la feature; dio %+v", s)
	}

	// Feature cerrada: el apruebo deja de valer; sin tope, la 61 del turno pasa.
	cerrarFeature(t, raiz, "E001/F002")
	turno("sigue")
	llenar("feature cerrada")
	if s := usar(""); s.codigo != 0 {
		t.Errorf("con la feature cerrada, la llamada 61 del turno pasa; dio %+v", s)
	}
}

// cerrarFeature deja en CERRADO la feature y sus slices y los mueve a trabajo/archivo, como el cierre real;
// sin slices, le crea uno cerrado (featureCerrada lo pide).
func cerrarFeature(t *testing.T, raiz, id string) {
	t.Helper()
	fs, _ := filepath.Glob(filepath.Join(raiz, "trabajo", "*", "*.json"))
	hijos := 0
	for _, p := range fs {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		var f map[string]any
		if json.Unmarshal(b, &f) != nil {
			continue
		}
		if f["id"] == id || f["padre"] == id {
			if f["padre"] == id {
				hijos++
			}
			f["estado"] = "CERRADO"
			b, _ = json.Marshal(f)
			archivo := filepath.Join(raiz, "trabajo", "archivo", filepath.Base(p))
			if err := os.MkdirAll(filepath.Dir(archivo), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(archivo, b, 0o644); err != nil {
				t.Fatal(err)
			}
			if archivo != p {
				if err := os.Remove(p); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	if hijos == 0 {
		escribir(t, raiz, "trabajo/archivo/"+strings.ReplaceAll(id, "/", "-")+"-S900.json",
			`{"id": "`+id+`/S900", "tipo": "slice", "titulo": "s", "estado": "CERRADO", "prioridad": "P1", "motivo": "m", "padre": "`+id+`", "bloqueado_por": [], "secuencia": 1}`)
	}
}
