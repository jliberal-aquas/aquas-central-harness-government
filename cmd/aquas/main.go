// Comando aquas: harness de gobierno para proyectos Claude Code.
//
//	aquas compilar   regenera derivados y sello desde los canónicos
//	aquas verificar  gate: falla (exit 1) si algún derivado o sello está desfasado o la wiki docs/ incumple
//	aquas hook       modo hook PostToolUse de Claude Code (lee JSON por stdin)
//	aquas guardia    hook PreToolUse: bloquea (exit 2) toda escritura fuera del repositorio
//	aquas iniciar    declara SO y raíz de esta máquina (entorno.local.json) y genera .claude/settings.json
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jliberal/aquas-harness-governance/internal/compilador"
	"github.com/jliberal/aquas-harness-governance/internal/entorno"
	"github.com/jliberal/aquas-harness-governance/internal/guardia"
	"github.com/jliberal/aquas-harness-governance/internal/wiki"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "uso: aquas iniciar | compilar | verificar | hook | guardia")
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
		compilarSettings(os.Stdout)
	case "iniciar":
		e, err := entorno.Iniciar(raiz)
		if err != nil {
			fallar(1, err)
		}
		fmt.Printf("Entorno declarado en %s · so=%s · raiz=%s\nConfírmalo: si no es esta máquina y esta carpeta, bórralo.\n", entorno.RutaEntorno, e.SO, e.Raiz)
		compilarSettings(os.Stdout)
	case "verificar":
		derivados, err := compilador.Verificar(raiz)
		if err != nil {
			fallar(1, err)
		}
		docs, err := wiki.Verificar(raiz)
		if err != nil {
			fallar(1, err)
		}
		permitidos := guardia.ValidarConfig(raiz)
		if m := entorno.Validar(raiz); m != "" {
			permitidos = append(permitidos, "entorno: "+m)
		} else if m := entorno.VerificarSettings(raiz); m != "" {
			permitidos = append(permitidos, m)
		}
		if len(derivados) > 0 {
			listar("GATE FALLIDO · derivados desfasados:", derivados)
			fmt.Fprintln(os.Stderr, "Desfase → go run ./cmd/aquas compilar · presupuesto → recortar canónico")
		}
		if len(docs) > 0 {
			listar("GATE FALLIDO · wiki docs/ incompleta:", docs)
		}
		if len(permitidos) > 0 {
			listar("GATE FALLIDO · guardia (permitidos, entorno o settings):", permitidos)
		}
		if len(derivados)+len(docs)+len(permitidos) > 0 {
			os.Exit(1)
		}
		fmt.Println("GATE OK · derivados, sello, wiki y guardia al día")
	case "hook":
		hook(raiz)
	case "guardia":
		datos, _ := io.ReadAll(os.Stdin)
		if m := entorno.Validar(raiz); m != "" {
			fmt.Fprintln(os.Stderr, "aquas guardia: entorno no verificado, se bloquea todo: "+m)
			os.Exit(2)
		}
		if m := guardia.Revisar(raiz, datos); m != "" {
			fmt.Fprintln(os.Stderr, "aquas guardia: "+m)
			os.Exit(2)
		}
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

// compilarSettings regenera .claude/settings.json; sin entorno declarado no inventa uno.
func compilarSettings(w io.Writer) {
	raiz, _ := os.Getwd()
	if m := entorno.Validar(raiz); m != "" {
		fallar(1, fmt.Errorf("settings.json no generado: %s", m))
	}
	cambio, err := entorno.CompilarSettings(raiz)
	if err != nil {
		fallar(1, err)
	}
	if cambio {
		fmt.Fprintln(w, "Escrito: "+entorno.RutaSettings)
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

func listar(titulo string, problemas []string) {
	fmt.Fprintln(os.Stderr, titulo)
	for _, p := range problemas {
		fmt.Fprintln(os.Stderr, "  - "+p)
	}
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
