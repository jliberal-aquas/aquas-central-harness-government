// Comando aquas: harness de gobierno para proyectos Claude Code.
//
//	aquas compilar   regenera derivados y sello desde los canónicos
//	aquas verificar  gate: falla (exit 1) si algún derivado o sello está desfasado
//	aquas hook       modo hook PostToolUse de Claude Code (lee JSON por stdin)
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jliberal/aquas-harness-governance/internal/compilador"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "uso: aquas compilar | verificar | hook")
		os.Exit(64)
	}
	raiz, err := os.Getwd()
	if err != nil {
		fallar(1, err)
	}
	switch os.Args[1] {
	case "compilar":
		inf, err := compilador.Compilar(raiz)
		if err != nil {
			fallar(1, err)
		}
		reportar(os.Stdout, inf)
	case "verificar":
		problemas, err := compilador.Verificar(raiz)
		if err != nil {
			fallar(1, err)
		}
		if len(problemas) > 0 {
			fmt.Fprintln(os.Stderr, "GATE FALLIDO · derivados desfasados:")
			for _, p := range problemas {
				fmt.Fprintln(os.Stderr, "  - "+p)
			}
			fmt.Fprintln(os.Stderr, "Corrige con: go run ./cmd/aquas compilar")
			os.Exit(1)
		}
		fmt.Println("GATE OK · derivados y sello al día")
	case "hook":
		hook(raiz)
	default:
		fallar(64, fmt.Errorf("subcomando desconocido: %s", os.Args[1]))
	}
}

// hook: exit 0 = nada que reportar; exit 2 = stderr se muestra a Claude.
func hook(raiz string) {
	var entrada struct {
		ToolInput struct {
			FilePath string `json:"file_path"`
		} `json:"tool_input"`
	}
	datos, _ := io.ReadAll(os.Stdin)
	if err := json.Unmarshal(datos, &entrada); err != nil || entrada.ToolInput.FilePath == "" {
		return
	}
	rel, err := filepath.Rel(raiz, entrada.ToolInput.FilePath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return
	}
	rel = filepath.ToSlash(rel)

	editoSalida := compilador.EsSalida(raiz, rel)
	if !editoSalida && !compilador.EsGobernado(rel) {
		return
	}
	inf, err := compilador.Compilar(raiz)
	if err != nil {
		fallar(2, fmt.Errorf("compilación falló tras editar %s: %w", rel, err))
	}
	if editoSalida {
		fmt.Fprintf(os.Stderr, "%s es GENERADO; tu edición se revirtió. Edita el canónico en gobierno/canonico/ o gobierno/config/.\n", rel)
		os.Exit(2)
	}
	if len(inf.Escritos) > 0 {
		reportar(os.Stderr, inf)
		os.Exit(2)
	}
}

func reportar(w io.Writer, inf compilador.Informe) {
	if len(inf.Escritos) == 0 {
		fmt.Fprintln(w, "Sin cambios · derivados al día")
		return
	}
	fmt.Fprintln(w, "Recompilado · fuentes cambiadas: "+lista(inf.FuentesCambiadas))
	fmt.Fprintln(w, "Escritos: "+lista(inf.Escritos))
}

func lista(xs []string) string {
	if len(xs) == 0 {
		return "—"
	}
	return strings.Join(xs, ", ")
}

func fallar(codigo int, err error) {
	fmt.Fprintln(os.Stderr, "aquas:", err)
	os.Exit(codigo)
}
