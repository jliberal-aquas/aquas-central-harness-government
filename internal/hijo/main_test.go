package hijo

import (
	"os"
	"testing"
)

// TestMain (E000/F009/S001, D000050): git no sube desde la carpeta temporal de una prueba al repo que la contiene.
func TestMain(m *testing.M) {
	if err := os.Setenv("GIT_CEILING_DIRECTORIES", os.TempDir()); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}
