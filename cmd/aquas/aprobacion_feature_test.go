package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// E001/F015 sobre el binario real: una aprobación por feature.

// repoFeature: el repo de repoContrato con la página de la feature E001/F002.
func repoFeature(t *testing.T) string {
	t.Helper()
	raiz := repoContrato(t)
	escribir(t, raiz, "docs/E001/F002/README.md", "# F002\n\n## Para qué existe\n\nContrato de la feature.\n")
	return raiz
}

// aprobacionFeature lee la aprobación guardada en la ficha de la feature.
func aprobacionFeature(t *testing.T, raiz string) (ap struct {
	Push bool
	Foto map[string]json.RawMessage
}) {
	t.Helper()
	b, hay := fichaEnDisco(t, raiz, "trabajo/activo/E001-F002.json")["aprobacion"]
	if !hay {
		t.Fatal("la ficha de la feature no guardó la aprobación")
	}
	if err := json.Unmarshal(b, &ap); err != nil {
		t.Fatal(err)
	}
	return ap
}

// TestApruebaFeatureDesdeElChat (E001/F015/S001): el hook registra «apruebo» de una feature y su marca de push.
// Rojo cuando: apruebo E001/F015 no se registra o con push no queda anotado.
// Verde cuando: el mensaje registra la aprobación de la feature y con push queda en la ficha.
func TestApruebaFeatureDesdeElChat(t *testing.T) {
	raiz := repoFeature(t)
	if s := aquas(t, raiz, mensajeHumano("apruebo E001/F002"), "aprobacion"); s.codigo != 0 {
		t.Fatalf("apruebo de feature: %+v", s)
	}
	if ap := aprobacionFeature(t, raiz); ap.Push || ap.Foto["E001/F002/S003"] == nil {
		t.Errorf("sin «con push»: push falso y foto con S003; dio %+v", ap)
	}
	if s := aquas(t, raiz, mensajeHumano("apruebo E001/F002 con push"), "aprobacion"); s.codigo != 0 {
		t.Fatalf("apruebo con push: %+v", s)
	}
	if ap := aprobacionFeature(t, raiz); !ap.Push {
		t.Errorf("«con push» queda en la ficha; dio %+v", ap)
	}
}

// TestGateHumanoPorFeature (E001/F015, integral): una aprobación de la feature deja avanzar sus slices.
// Rojo cuando: un slice de la feature aprobada exige su propio apruebo o avanza con una prueba aprobada cambiada.
// Verde cuando: los slices aprobados avanzan, agregar pruebas no invalida y cambiar una aprobada sí.
func TestGateHumanoPorFeature(t *testing.T) {
	raiz := repoFeature(t)
	escribir(t, raiz, "p/q_test.go", "package p\n\nimport \"testing\"\n\nfunc TestOtra(t *testing.T) { t.Fatal(\"falta\") }\n")
	backlogOK := func(args ...string) salida {
		t.Helper()
		return aquas(t, raiz, "", append([]string{"backlog"}, args...)...)
	}
	if s := aquas(t, raiz, mensajeHumano("apruebo E001/F002 con push"), "aprobacion"); s.codigo != 0 {
		t.Fatalf("apruebo de feature: %+v", s)
	}
	if s := backlogOK("pruebas", "E001/F002/S003", "./p:TestBandera=pide el arreglo|r|v", "./p:TestOtra=otra|r|v"); s.codigo != 0 {
		t.Fatalf("agregar prueba: %+v", s)
	}
	if s := backlogOK("estado", "E001/F002/S003", "EN_CURSO"); s.codigo != 0 {
		t.Fatalf("slice de feature aprobada arranca sin apruebo propio: %+v", s)
	}
	if s := backlogOK("rojo", "E001/F002/S003"); s.codigo != 0 {
		t.Fatalf("rojo con una prueba agregada: %+v", s)
	}
	backlogOK("pruebas", "E001/F002/S003", "./p:TestBandera=otra intención|r|v", "./p:TestOtra=otra|r|v")
	if s := backlogOK("rojo", "E001/F002/S003"); s.codigo == 0 || !strings.Contains(s.err, "apruebo E001/F002") {
		t.Errorf("prueba aprobada cambiada: rojo rechazado pidiendo «apruebo E001/F002»; dio %+v", s)
	}
}
