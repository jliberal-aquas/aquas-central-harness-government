package hijo

import (
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestReportArmaIssueConVersion(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "landing")
	central := fstest.MapFS{"gobierno/canonico/constitucion.md": {Data: []byte(constitucionConHuecos)}}
	if err := Install(dir, central, "v1.2.3"); err != nil {
		t.Fatal(err)
	}
	// Pasa: el borrador trae firma del hijo, qué pasó, reproducción y versión del lock.
	body, err := Report(dir, "build no avisa X", "1. aquas build .\n2. ver Y")
	if err != nil {
		t.Fatalf("report falló: %v", err)
	}
	for _, debe := range []string{"«landing»", "v1.2.3", "build no avisa X", "aquas build .", "Sellados alterados: ninguno"} {
		if !strings.Contains(body, debe) {
			t.Errorf("borrador sin %q:\n%s", debe, body)
		}
	}
	// Falla: sin qué pasó, sin reproducción o sin instalación.
	if _, err := Report(dir, "", "x"); err == nil {
		t.Error("report sin qué pasó debió fallar")
	}
	if _, err := Report(dir, "x", ""); err == nil {
		t.Error("report sin reproducción debió fallar")
	}
	if _, err := Report(t.TempDir(), "x", "y"); err == nil || !strings.Contains(err.Error(), "install") {
		t.Errorf("report sin instalación debió pedir install; err=%v", err)
	}
}
