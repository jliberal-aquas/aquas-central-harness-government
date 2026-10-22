package wiki

// E001/F002/S010: navegación por niveles. Cada página lista solo a sus hijos directos
// (inicio → iniciativas, E → features, F → slices) y un slice lista sus pruebas. Esa lista
// la genera aquas desde las fichas trabajo/ en un bloque GENERADO; el gate falla si se desfasa,
// si una página enlaza saltando niveles, o si una ficha arrancada no tiene página.
// Límite: sin fichas en trabajo/ no se exige bloque (repo sin backlog).

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/backlog"
)

const (
	marcaIni = "<!-- GENERADO por aquas desde trabajo/: no editar a mano; `aquas compilar` lo rehace. -->"
	marcaFin = "<!-- /GENERADO -->"
)

// estadoVisible traduce el estado de la ficha a la palabra que ve el humano.
var estadoVisible = map[string]string{
	"BORRADOR": "Planeado", "LISTO": "Planeado", "EN_CURSO": "En curso", "VERIFICANDO": "En curso",
	"BLOQUEADO": "Bloqueado", "CERRADO": "Hecho", "CANCELADO": "Cancelado",
}

// paginaDe: la página de la wiki que corresponde a un id de ficha.
func paginaDe(id string) string {
	p := strings.Split(id, "/")
	switch len(p) {
	case 1:
		return "docs/" + p[0] + "/README.md"
	case 2:
		return "docs/" + p[0] + "/" + p[1] + "/README.md"
	}
	return "docs/" + p[0] + "/" + p[1] + "/" + p[2] + ".md"
}

func celda(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\n", " "), "|", `\|`)
}

func existe(raiz, rel string) bool {
	_, err := os.Stat(filepath.Join(raiz, filepath.FromSlash(rel)))
	return err == nil
}

// tablaHijos: una fila por hijo directo; enlace solo si su página existe.
func tablaHijos(raiz, padreRel string, hijos []backlog.Ficha) string {
	var b strings.Builder
	b.WriteString("| Hijo | Qué es | Estado |\n|---|---|---|\n")
	if len(hijos) == 0 {
		b.WriteString("| — | sin hijos todavía | — |\n")
	}
	for _, h := range hijos {
		corto := h.ID[strings.LastIndex(h.ID, "/")+1:]
		nombre := corto
		if dest := paginaDe(h.ID); existe(raiz, dest) {
			rel, _ := filepath.Rel(filepath.Dir(filepath.FromSlash(padreRel)), filepath.FromSlash(dest))
			nombre = "[" + corto + "](" + filepath.ToSlash(rel) + ")"
		}
		fmt.Fprintf(&b, "| %s | %s | %s |\n", nombre, celda(h.Titulo), estadoVisible[h.Estado])
	}
	return b.String()
}

func fecha(r *backlog.Registro) string {
	if r == nil || len(r.Fecha) < 10 {
		return "—"
	}
	return r.Fecha[:10]
}

// oGuion: el texto como celda, o «—» si está vacío (pruebas anteriores a E001/F008/S001).
func oGuion(s string) string {
	if strings.TrimSpace(s) == "" {
		return "—"
	}
	return celda(s)
}

// tablaPruebas: las pruebas del slice: qué validan, cuándo quedan en rojo y en verde, y si se vieron en rojo y en verde.
func tablaPruebas(f backlog.Ficha) string {
	var b strings.Builder
	b.WriteString("| Prueba | Valida | Rojo cuando | Verde cuando | Rojo | Verde |\n|---|---|---|---|---|---|\n")
	if len(f.Pruebas) == 0 {
		b.WriteString("| — | ficha anterior al registro de pruebas | — | — | — | — |\n")
	}
	for _, p := range f.Pruebas {
		rojo := "—"
		if f.Rojo != nil && f.Rojo.Resultado[p.Nombre] != "" {
			rojo = fecha(f.Rojo)
		}
		fmt.Fprintf(&b, "| `%s` | %s | %s | %s | %s | %s |\n", p.Nombre[strings.LastIndex(p.Nombre, ":")+1:], celda(p.Intencion), oGuion(p.RojoCuando), oGuion(p.VerdeCuando), rojo, fecha(f.Verde))
	}
	return b.String()
}

type esperado struct {
	seccion, bloque string
}

// esperados: página existente → bloque que debe llevar. Vacío si no hay fichas.
func esperados(raiz string) (map[string]esperado, []backlog.Ficha, error) {
	fichas, err := backlog.Todas(raiz)
	if err != nil || len(fichas) == 0 {
		return nil, fichas, err
	}
	hijos := map[string][]backlog.Ficha{}
	for _, f := range fichas {
		hijos[f.Padre] = append(hijos[f.Padre], f)
	}
	envolver := func(tabla string) string { return marcaIni + "\n\n" + tabla + "\n" + marcaFin }
	exp := map[string]esperado{}
	if existe(raiz, inicio) {
		exp[inicio] = esperado{secNav, envolver(tablaHijos(raiz, inicio, hijos[""]))}
	}
	for _, f := range fichas {
		rel := paginaDe(f.ID)
		if !existe(raiz, rel) {
			continue
		}
		if f.Tipo == "slice" {
			exp[rel] = esperado{secContrato, envolver(tablaPruebas(f) + tablaArreglos(f))}
		} else {
			exp[rel] = esperado{secNav, envolver(tablaHijos(raiz, rel, hijos[f.ID]) + tablaArreglos(f))}
		}
	}
	return exp, fichas, nil
}

// bloqueActual: posición del bloque en c; ini < 0 = no hay; fin < 0 = sin cierre.
func bloqueActual(c string) (ini, fin int) {
	ini = strings.Index(c, "<!-- GENERADO")
	if ini < 0 {
		return -1, -1
	}
	j := strings.Index(c[ini:], marcaFin)
	if j < 0 {
		return ini, -1
	}
	return ini, ini + j + len(marcaFin)
}

// Generar reescribe el bloque GENERADO de cada página desde las fichas; devuelve lo escrito.
func Generar(raiz string) ([]string, error) {
	exp, _, err := esperados(raiz)
	if err != nil {
		return nil, err
	}
	var escritos []string
	for rel, e := range exp {
		p := filepath.Join(raiz, filepath.FromSlash(rel))
		b, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		c := string(b)
		nuevo := c
		if ini, fin := bloqueActual(c); ini >= 0 && fin >= 0 {
			nuevo = c[:ini] + e.bloque + c[fin:]
		} else if ini >= 0 {
			return nil, fmt.Errorf("%s: bloque GENERADO sin cierre %s: arréglalo a mano", rel, marcaFin)
		} else {
			loc := regexp.MustCompile(`(?m)^## ` + regexp.QuoteMeta(e.seccion) + `[ \t]*\r?\n`).FindStringIndex(c)
			if loc == nil {
				continue // falta la sección: el gate lo reporta
			}
			nuevo = c[:loc[1]] + "\n" + e.bloque + "\n" + c[loc[1]:]
		}
		if nuevo != c {
			if err := os.WriteFile(p, []byte(nuevo), 0o644); err != nil {
				return nil, err
			}
			escritos = append(escritos, rel)
		}
	}
	sort.Strings(escritos)
	return escritos, nil
}

// revisarHijos: bloque presente y al día, enlaces sin saltar niveles, ficha arrancada con página.
func revisarHijos(raiz string, paginas map[string]*pagina, orden []string) ([]string, error) {
	exp, fichas, err := esperados(raiz)
	if err != nil {
		return nil, err
	}
	var ps []string
	for _, rel := range orden {
		e, ok := exp[rel]
		if !ok {
			continue
		}
		b, err := os.ReadFile(filepath.Join(raiz, filepath.FromSlash(rel)))
		if err != nil {
			return nil, err
		}
		c := strings.ReplaceAll(string(b), "\r\n", "\n")
		switch ini, fin := bloqueActual(c); {
		case ini < 0:
			ps = append(ps, rel+": falta el bloque GENERADO de hijos (corre `aquas compilar`)")
		case fin < 0:
			ps = append(ps, rel+": bloque GENERADO sin cierre "+marcaFin)
		case c[ini:fin] != e.bloque:
			ps = append(ps, rel+": hijos desfasados con trabajo/ (corre `aquas compilar`)")
		}
	}
	for _, rel := range orden {
		pg := paginas[rel]
		if pg.nivel > nS {
			continue
		}
		visto := map[string]bool{}
		for _, en := range pg.enlaces {
			d, _ := resolver(rel, en.destino)
			if dst := clasificar(d); dst != nil && dst.nivel >= nE && dst.nivel <= nS && dst.nivel-pg.nivel >= 2 && !visto[d] {
				visto[d] = true
				ps = append(ps, fmt.Sprintf("%s: enlace salta niveles → %s (cada página enlaza solo hasta sus hijos directos)", rel, d))
			}
		}
	}
	for _, f := range fichas {
		if f.Estado == "EN_CURSO" || f.Estado == "VERIFICANDO" || f.Estado == "CERRADO" {
			if rel := paginaDe(f.ID); paginas[rel] == nil {
				ps = append(ps, fmt.Sprintf("%s: falta la página de %s (ficha %s)", rel, f.ID, f.Estado))
			}
		}
	}
	return ps, nil
}

// tablaArreglos: sección «Arreglos» de la ficha; vacía si no tiene.
// Cada fila: ID corto, qué arregla (y el slice que absorbió), por qué, sus pruebas y la fecha.
func tablaArreglos(f backlog.Ficha) string {
	if len(f.Arreglos) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("\n### Arreglos\n\n| Arreglo | Qué arregla | Por qué | Pruebas | Fecha |\n|---|---|---|---|---|\n")
	for _, a := range f.Arreglos {
		titulo := a.Titulo
		if a.Antes != "" {
			titulo += " (antes " + a.Antes[strings.LastIndex(a.Antes, "/")+1:] + ")"
		}
		var ps []string
		for _, p := range a.Pruebas {
			ps = append(ps, "`"+p.Nombre[strings.LastIndex(p.Nombre, ":")+1:]+"`")
		}
		fecha := a.Fecha
		if len(fecha) >= 10 {
			fecha = fecha[:10]
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s |\n", a.ID[strings.LastIndex(a.ID, "/")+1:], celda(titulo), celda(a.Motivo), oGuion(strings.Join(ps, ", ")), fecha)
	}
	return b.String()
}
