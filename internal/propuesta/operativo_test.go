package propuesta

import "testing"

// TestOperativoConPalabrasMedidas (E000/F010/X01): palabras operativas medidas en el registro real (2026-11-02,
// decisión humana): «lo recomendado», «aprobado» y «sube» son operativas.
// Rojo cuando: alguna frena.
// Verde cuando: pasan y un pedido nuevo corto sin ellas sigue siendo nuevo.
func TestOperativoConPalabrasMedidas(t *testing.T) {
	for _, m := range []string{"lo recomendado", "Lo recomendado.", "aprobado", "Aprobado, sube", "sube"} {
		if !Operativo(m) {
			t.Errorf("%q: operativo", m)
		}
	}
	for _, m := range []string{"quiero etiquetas", "subir el tope", "recomiéndame algo"} {
		if Operativo(m) {
			t.Errorf("%q: pedido nuevo", m)
		}
	}
}
