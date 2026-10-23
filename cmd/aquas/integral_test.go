package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// reFeatureNueva: argumento id= de una feature (sin slice).
var reFeatureNueva = regexp.MustCompile(`^id=(E\d{3}/F\d{3})$`)

// rojoIntegral: si args crean una feature, le anota un rojo a su integral; desde D000023 sus slices lo exigen para arrancar.
func rojoIntegral(t *testing.T, raiz string, args []string) {
	t.Helper()
	if len(args) < 2 || args[0] != "nuevo" {
		return
	}
	m := reFeatureNueva.FindStringSubmatch(args[1])
	if m == nil {
		return
	}
	ruta := filepath.Join(raiz, "trabajo", "activo", strings.ReplaceAll(m[1], "/", "-")+".json")
	b, err := os.ReadFile(ruta)
	if err != nil {
		t.Fatal(err)
	}
	var f map[string]any
	if err := json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	f["rojo"] = map[string]any{"fecha": "2026-10-21T00:00:00Z", "head": "", "resultado": map[string]string{"integral": "falló"}}
	if b, err = json.MarshalIndent(f, "", "  "); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ruta, append(b, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}
