package wiki

// E001/F008/S004: toda función de cmd/ e internal/ lleva comentario; docs/ref/codigo.md las indexa.

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jliberal-aquas/aquas-central-harness-government/internal/backlog"
)

// paginaCodigo: página de referencia que lleva el índice generado.
const paginaCodigo = "docs/ref/codigo.md"

// funcion: una función o método declarado en el código de producción.
type funcion struct {
	archivo, nombre, comentario string
	linea                       int
	firma                       string // parámetros y resultados, sin «func»
}

// funcionesGo recorre cmd/ e internal/ (sin _test.go) y devuelve sus funciones en orden de archivo y línea.
func funcionesGo(raiz string) ([]funcion, error) {
	var todas []funcion
	for _, base := range []string{"cmd", "internal"} {
		dir := filepath.Join(raiz, base)
		if _, err := os.Stat(dir); err != nil {
			continue
		}
		err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
				return err
			}
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, p, nil, parser.ParseComments)
			if err != nil {
				return err
			}
			rel := relativa(raiz, p)
			for _, d := range f.Decls {
				fd, ok := d.(*ast.FuncDecl)
				if !ok {
					continue
				}
				n := fd.Name.Name
				if fd.Recv != nil && len(fd.Recv.List) > 0 {
					t := fd.Recv.List[0].Type
					if s, ok := t.(*ast.StarExpr); ok {
						t = s.X
					}
					if id, ok := t.(*ast.Ident); ok {
						n = id.Name + "." + n
					}
				}
				c := ""
				if fd.Doc != nil {
					c, _, _ = strings.Cut(strings.TrimSpace(fd.Doc.Text()), "\n")
				}
				var firma strings.Builder
				printer.Fprint(&firma, fset, fd.Type)
				todas = append(todas, funcion{rel, n, c, fset.Position(fd.Pos()).Line, strings.TrimPrefix(firma.String(), "func")})
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.SliceStable(todas, func(i, j int) bool {
		if todas[i].archivo != todas[j].archivo {
			return todas[i].archivo < todas[j].archivo
		}
		return todas[i].linea < todas[j].linea
	})
	return todas, nil
}

// revisarCodigo: avisos por cada función sin comentario.
func revisarCodigo(raiz string) ([]string, error) {
	fns, err := funcionesGo(raiz)
	if err != nil {
		return nil, err
	}
	var ps []string
	for _, f := range fns {
		if f.comentario == "" {
			ps = append(ps, fmt.Sprintf("%s:%d: función %s sin comentario: di qué hace (E001/F008/S004)", f.archivo, f.linea, f.nombre))
		}
	}
	return ps, nil
}

// indiceCodigo: tabla archivo → función → parámetros → primera línea de su comentario.
func indiceCodigo(raiz string) (string, error) {
	fns, err := funcionesGo(raiz)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString("| Archivo | Función | Parámetros | Qué hace |\n|---|---|---|---|\n")
	for _, f := range fns {
		fmt.Fprintf(&b, "| [%s](../../%s) | `%s` | `%s` | %s |\n", f.archivo, f.archivo, f.nombre, celda(f.firma), oGuion(celda(f.comentario)))
	}
	return b.String(), nil
}

// tablaCodigo: sección «Código» de un slice: cada archivo tocado con sus funciones, parámetros y qué hacen.
func tablaCodigo(raiz string, f backlog.Ficha) (string, error) {
	if len(f.Codigo) == 0 && f.Estado != "CERRADO" {
		return "", nil // en curso: el código se registra al cerrar
	}
	if len(f.Codigo) == 0 { // D000022: en un slice cerrado la sección vacía queda a la vista
		return "\n### Código\n\nSin código registrado: sus cambios no tocaron código ni configuración, o no nombran el ID (`aquas backlog codigo " + f.ID + " <commit>…` los asigna).\n", nil
	}
	fs, err := funcionesGo(raiz)
	if err != nil {
		return "", err
	}
	subir := strings.Repeat("../", strings.Count(paginaDe(f.ID), "/"))
	var b strings.Builder
	b.WriteString("\n### Código\n\n| Archivo | Función | Parámetros | Qué hace |\n|---|---|---|---|\n")
	for _, a := range f.Codigo {
		if !existe(raiz, a) {
			fmt.Fprintf(&b, "| %s (ya no existe) | — | — | — |\n", a)
			continue
		}
		n := 0
		for _, fn := range fs {
			if fn.archivo == a {
				fmt.Fprintf(&b, "| [%s](%s%s) | `%s` | `%s` | %s |\n", a, subir, a, fn.nombre, celda(fn.firma), oGuion(celda(fn.comentario)))
				n++
			}
		}
		if n == 0 {
			fmt.Fprintf(&b, "| [%s](%s%s) | — | — | — |\n", a, subir, a)
		}
	}
	return b.String(), nil
}
