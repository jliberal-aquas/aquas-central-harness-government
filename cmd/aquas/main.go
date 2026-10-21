// Comando aquas: harness de gobierno para proyectos Claude Code.
//
//	aquas compilar   regenera derivados y sello desde los canónicos
//	aquas verificar  gate: falla (exit 1) si algún derivado o sello está desfasado o la wiki docs/ incumple
//	aquas hook       modo hook PostToolUse de Claude Code (lee JSON por stdin)
//	aquas guardia    hook PreToolUse: bloquea (exit 2) toda escritura fuera del repositorio
//	aquas iniciar    declara SO y raíz de esta máquina (entorno.local.json) y genera .claude/settings.json
//	aquas biblioteca leer  bibliotecario: JSON {"rutas":[…]} por stdin → solo lo pedido, numerado
//	aquas cierre-agente  hook SubagentStop: exige el CIERRE fijo de worker y auditor (exit 2)
//	aquas memoria    ver | fijar campo=valor… | inicio | cierre | archivar: memoria caliente del foco (.aquas/memoria.json)
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jliberal/aquas-harness-governance/internal/backlog"
	"github.com/jliberal/aquas-harness-governance/internal/biblioteca"
	"github.com/jliberal/aquas-harness-governance/internal/compilador"
	"github.com/jliberal/aquas-harness-governance/internal/encargo"
	"github.com/jliberal/aquas-harness-governance/internal/entorno"
	"github.com/jliberal/aquas-harness-governance/internal/guardia"
	"github.com/jliberal/aquas-harness-governance/internal/memoria"
	"github.com/jliberal/aquas-harness-governance/internal/wiki"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "uso: aquas iniciar | compilar | verificar | hook | guardia | memoria | backlog | cierre-agente | biblioteca")
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
		fichas := backlog.Validar(raiz)
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
		if len(fichas) > 0 {
			listar("GATE FALLIDO · backlog trabajo/ (fichas):", fichas)
		}
		if len(permitidos) > 0 {
			listar("GATE FALLIDO · guardia (permitidos, entorno o settings):", permitidos)
		}
		if len(derivados)+len(docs)+len(fichas)+len(permitidos) > 0 {
			os.Exit(1)
		}
		fmt.Println("GATE OK · derivados, sello, wiki, backlog y guardia al día")
	case "hook":
		hook(raiz)
	case "memoria":
		memoriaCLI(raiz, os.Args[2:])
	case "backlog":
		backlogCLI(raiz, os.Args[2:])
	case "biblioteca": // bibliotecario: lote de rutas por stdin (JSON) → solo lo pedido
		if len(os.Args) < 3 || os.Args[2] != "leer" {
			fallar(64, fmt.Errorf(`uso: echo '{"rutas":["a.go","b.md:10-40"]}' | aquas biblioteca leer`))
		}
		var p biblioteca.Pedido
		datos, err := io.ReadAll(os.Stdin)
		if err != nil {
			fallar(1, fmt.Errorf("no se pudo leer stdin: %v", err))
		}
		if err := json.Unmarshal(datos, &p); err != nil {
			fallar(64, fmt.Errorf(`se espera JSON {"rutas":[…]} por stdin: %v`, err))
		}
		if len(p.Rutas) == 0 {
			fallar(64, fmt.Errorf(`pedido sin rutas: {"rutas":["a.go:10-40"]}`))
		}
		texto, _, _ := biblioteca.Leer(raiz, p)
		fmt.Print(texto)
	case "cierre-agente": // hook SubagentStop: exit 2 + stderr = el subagente reescribe su cierre
		var e struct {
			Tipo    string `json:"agent_type"`
			Mensaje string `json:"last_assistant_message"`
			Activo  bool   `json:"stop_hook_active"`
		}
		datos, _ := io.ReadAll(os.Stdin)
		_ = json.Unmarshal(datos, &e)
		if m := encargo.Revisar(e.Tipo, e.Mensaje, e.Activo); m != "" {
			fmt.Fprintln(os.Stderr, m)
			os.Exit(2)
		}
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

// memoriaCLI: `ver` imprime la memoria; `fijar campo=valor…` la actualiza y la sella.
func memoriaCLI(raiz string, args []string) {
	switch {
	case len(args) == 0:
	case args[0] == "inicio": // hook SessionStart: stdout entra al contexto; nunca bloquea
		fmt.Print(memoria.Inicio(raiz))
		return
	case args[0] == "cierre": // hook Stop: exit 2 + stderr = seguir trabajando
		var e struct {
			Activo bool `json:"stop_hook_active"`
		}
		datos, _ := io.ReadAll(os.Stdin)
		_ = json.Unmarshal(datos, &e)
		if m := memoria.Cierre(raiz, e.Activo); m != "" {
			fmt.Fprintln(os.Stderr, m)
			os.Exit(2)
		}
		return
	case args[0] == "archivar": // cierre del slice: copia sellada a registro/memoria/
		rel, err := memoria.Archivar(raiz, time.Now())
		if err != nil {
			fallar(1, err)
		}
		fmt.Println("archivada en " + rel)
		m, err := memoria.Leer(raiz)
		if err != nil {
			fallar(1, err)
		}
		fmt.Print(memoria.Texto(m))
		return
	}
	if len(args) == 0 || (args[0] != "ver" && args[0] != "fijar") {
		fallar(64, fmt.Errorf("uso: aquas memoria ver | fijar campo=valor… | inicio | cierre | archivar"))
	}
	var m memoria.Memoria
	var err error
	if args[0] == "fijar" {
		if len(args) == 1 {
			fallar(64, fmt.Errorf("fijar sin campos: aquas memoria fijar tarea=… siguiente=…"))
		}
		m, err = memoria.FijarCon(raiz, args[1:], time.Now(), func(f memoria.Foco) error { return focoEnBacklog(raiz, f) })
	} else {
		m, err = memoria.Leer(raiz)
	}
	if err != nil {
		fallar(1, err)
	}
	fmt.Print(memoria.Texto(m))
}

// memoriaArchivada: CERRADO exige que la memoria que apunta al slice ya esté archivada (E001/F002/S009).
// Si se cierra antes y la memoria queda vieja, fijar y archivar se bloquean entre sí. Cancelar no lo exige.
func memoriaArchivada(raiz, id string) error {
	m, err := memoria.Leer(raiz)
	if errors.Is(err, memoria.ErrSinMemoria) {
		return nil
	}
	if err != nil {
		return err
	}
	f := m.Foco
	rel := memoria.RutaArchivo(f)
	if rel == "" || f.Iniciativa+"/"+f.Feature+"/"+f.Slice != id {
		return nil
	}
	if _, err := os.Stat(filepath.Join(raiz, filepath.FromSlash(rel))); err == nil {
		return nil
	}
	return fmt.Errorf("la memoria apunta a %s y no está archivada: primero `aquas memoria fijar tarea=… siguiente=…`, luego `aquas memoria archivar`, y entonces cierra", id)
}

// focoEnBacklog: la memoria solo fija el slice que el backlog tiene en curso (S003).
func focoEnBacklog(raiz string, f memoria.Foco) error {
	if f.Slice == "" {
		return nil
	}
	id := f.Iniciativa + "/" + f.Feature + "/" + f.Slice
	cur, hay, err := backlog.SliceEnCurso(raiz)
	switch {
	case err != nil:
		return err
	case !hay:
		if f, archivada, _ := backlog.Ver(raiz, id); archivada {
			return fmt.Errorf("%s ya está archivado (%s): la memoria se archiva antes de cerrar (memoria fijar → memoria archivar → backlog estado CERRADO). Si la memoria sigue fresca, `aquas memoria archivar`; si no, vacía el slice con `aquas memoria fijar slice=` (queda sin archivo de memoria)", id, f.Estado)
		}
		return fmt.Errorf("ningún slice en curso en el backlog: la memoria no fija %s (aquas backlog estado %s EN_CURSO)", id, id)
	case cur != id:
		return fmt.Errorf("el slice en curso es %s: la memoria no fija %s", cur, id)
	}
	return nil
}

// backlogCLI: nuevo | ver [id] | prioridad | estado | siguiente | pruebas | rojo | verde.
func backlogCLI(raiz string, args []string) {
	uso := fmt.Errorf("uso: aquas backlog nuevo id=… titulo=… prioridad=… motivo=… prueba=./paquete:TestNombre=intención… [bloqueado_por=A,B] | ver [id] | prioridad <id> <P0..P3|SIN_TRIAGE> motivo=… | estado <id> <ESTADO> | siguiente | pruebas <id> ./paquete:TestNombre=intención… | rojo <id> | verde <id>")
	if len(args) == 0 {
		fallar(64, uso)
	}
	var f backlog.Ficha
	var err error
	detalle := false
	switch {
	case args[0] == "nuevo" && len(args) > 1:
		f, err = backlog.Nuevo(raiz, args[1:])
	case args[0] == "ver" && len(args) == 1:
		fs, err := backlog.Activas(raiz)
		if err != nil {
			fallar(1, err)
		}
		for _, f := range fs {
			fmt.Print(backlog.Texto(f))
		}
		return
	case args[0] == "ver" && len(args) == 2:
		var archivada bool
		f, archivada, err = backlog.Ver(raiz, args[1])
		if err == nil && archivada {
			fmt.Print("archivada · ")
		}
		detalle = true
	case args[0] == "pruebas" && len(args) > 2:
		f, err = backlog.Pruebas(raiz, args[1], args[2:])
		detalle = true
	case args[0] == "rojo" && len(args) == 2:
		f, err = backlog.Rojo(raiz, args[1], time.Now())
		detalle = true
	case args[0] == "verde" && len(args) == 2:
		f, err = backlog.Verde(raiz, args[1], time.Now())
		detalle = true
	case args[0] == "prioridad" && len(args) == 4 && strings.HasPrefix(args[3], "motivo="):
		f, err = backlog.Prioridad(raiz, args[1], args[2], strings.TrimPrefix(args[3], "motivo="))
	case args[0] == "estado" && len(args) == 3:
		if args[2] == "CERRADO" {
			err = memoriaArchivada(raiz, args[1])
		}
		if err == nil {
			f, err = backlog.Estado(raiz, args[1], args[2])
		}
		if err == nil && (f.Estado == "CERRADO" || f.Estado == "CANCELADO") {
			fmt.Printf("archivada en %s/ · ", backlog.Archivo)
		}
	case args[0] == "siguiente" && len(args) == 1:
		var hay bool
		f, hay, err = backlog.Siguiente(raiz)
		if err == nil && !hay {
			fmt.Println("nada: no hay slice en curso ni desbloqueado")
			return
		}
	default:
		fallar(64, uso)
	}
	if err != nil {
		fallar(1, err)
	}
	fmt.Print(backlog.Texto(f))
	if detalle {
		fmt.Print(backlog.Detalle(f))
	}
}
