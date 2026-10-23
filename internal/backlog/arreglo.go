package backlog

// E001/F008/S006: un arreglo no es un slice: cuelga de su slice (o de su feature) con ID propio.

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Arreglo: corrección posterior a un slice o feature. ID = padre/X01, X02…
// Antes guarda el ID del slice que el arreglo absorbió, si lo hubo.
type Arreglo struct {
	ID      string   `json:"id"`
	Titulo  string   `json:"titulo"`
	Motivo  string   `json:"motivo"`
	Fecha   string   `json:"fecha"`
	Antes   string   `json:"antes,omitempty"`
	Pruebas []Prueba `json:"pruebas,omitempty"`
}

// idArreglo: ID del n-ésimo arreglo (1, 2…) de padre: padre/X01.
func idArreglo(padre string, n int) string { return fmt.Sprintf("%s/X%02d", padre, n) }

// dirDe: carpeta de trabajo/ donde vive la ficha, según esté archivada o no.
func dirDe(archivada bool) string {
	if archivada {
		return Archivo
	}
	return Activo
}

// Arreglar anota un arreglo en la ficha de su slice o feature, abierta o archivada.
// absorbe (opcional): slice cerrado hermano que en realidad era un arreglo; hereda sus pruebas y se borra.
func Arreglar(raiz, padre, titulo, motivo, absorbe string, ahora time.Time) (Arreglo, error) {
	titulo, motivo = strings.TrimSpace(titulo), strings.TrimSpace(motivo)
	if titulo == "" || motivo == "" {
		return Arreglo{}, errors.New("arreglo sin titulo o motivo: aquas backlog arreglo <slice|feature> titulo=… motivo=… [absorbe=<slice>]")
	}
	f, archivada, err := Ver(raiz, padre)
	if err != nil {
		return Arreglo{}, fmt.Errorf("arreglo sin padre: %w", err)
	}
	if f.Tipo != "slice" && f.Tipo != "feature" {
		return Arreglo{}, fmt.Errorf("%s es %s: un arreglo cuelga de un slice o de un feature", padre, f.Tipo)
	}
	a := Arreglo{ID: idArreglo(padre, len(f.Arreglos)+1), Titulo: titulo, Motivo: motivo, Fecha: ahora.UTC().Format(time.RFC3339)}
	borrar := ""
	if absorbe != "" {
		g, gArch, err := Ver(raiz, absorbe)
		switch {
		case err != nil:
			return Arreglo{}, err
		case g.Tipo != "slice" || g.Estado != "CERRADO" || g.Padre != f.Padre || g.ID == f.ID:
			return Arreglo{}, fmt.Errorf("%s no se absorbe: debe ser un slice CERRADO hermano de %s", absorbe, padre)
		case len(g.Arreglos) > 0:
			return Arreglo{}, fmt.Errorf("%s tiene arreglos propios: muévelos antes", absorbe)
		}
		a.Antes, a.Pruebas = g.ID, g.Pruebas
		borrar = filepath.Join(raiz, filepath.FromSlash(dirDe(gArch)), nombre(g.ID))
	}
	f.Arreglos = append(f.Arreglos, a)
	if err := escribir(raiz, dirDe(archivada), f); err != nil {
		return Arreglo{}, err
	}
	if borrar != "" {
		return a, os.Remove(borrar)
	}
	return a, nil
}

// problemasArreglos: chequeo del gate: IDs en secuencia bajo su ficha, con título y motivo.
func problemasArreglos(rel string, f Ficha) []string {
	var p []string
	for i, a := range f.Arreglos {
		if quiere := idArreglo(f.ID, i+1); a.ID != quiere {
			p = append(p, fmt.Sprintf("%s: arreglo %s fuera de secuencia (se esperaba %s)", rel, a.ID, quiere))
		}
		if strings.TrimSpace(a.Titulo) == "" || strings.TrimSpace(a.Motivo) == "" {
			p = append(p, fmt.Sprintf("%s: arreglo %s sin título o motivo", rel, a.ID))
		}
		for _, x := range a.Pruebas {
			if m := problemaPrueba(x); m != "" {
				p = append(p, fmt.Sprintf("%s: arreglo %s: %s", rel, a.ID, m))
			}
		}
	}
	return p
}
