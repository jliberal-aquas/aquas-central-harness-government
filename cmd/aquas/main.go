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
//	aquas medir vueltas  cuenta corridas, fallas y motivos de go test y verificar en el historial de Claude Code
//	aquas wiki enlazar docs/x.md…  enlaza al glosario la primera aparición de cada término
//	aquas limpiar <id>  borra .tmp/<id>/ de un encargo (E000/F006/S010-A1) y nada más
//	aquas defecto    nuevo firma=… que=… causa=… arreglo=… cercioro=… | ocurre D000001 T000001 | ver [D000001]
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	gobcentral "github.com/jliberal-aquas/aquas-central-harness-government"
	"github.com/jliberal-aquas/aquas-central-harness-government/internal/backlog"
	"github.com/jliberal-aquas/aquas-central-harness-government/internal/biblioteca"
	"github.com/jliberal-aquas/aquas-central-harness-government/internal/compilador"
	"github.com/jliberal-aquas/aquas-central-harness-government/internal/defecto"
	"github.com/jliberal-aquas/aquas-central-harness-government/internal/encargo"
	"github.com/jliberal-aquas/aquas-central-harness-government/internal/entorno"
	"github.com/jliberal-aquas/aquas-central-harness-government/internal/guardia"
	"github.com/jliberal-aquas/aquas-central-harness-government/internal/hijo"
	"github.com/jliberal-aquas/aquas-central-harness-government/internal/memoria"
	"github.com/jliberal-aquas/aquas-central-harness-government/internal/propuesta"
	"github.com/jliberal-aquas/aquas-central-harness-government/internal/turno"
	"github.com/jliberal-aquas/aquas-central-harness-government/internal/vueltas"
	"github.com/jliberal-aquas/aquas-central-harness-government/internal/wiki"
)

// main despacha el subcomando de aquas (compilar, verificar, backlog, hooks…) y sale con su código.
func main() {
	if len(os.Args) < 2 {
		_, _ = fmt.Fprintln(os.Stderr, "uso: aquas iniciar | compilar | verificar | hook | guardia | memoria | backlog | cierre-agente | biblioteca")
		os.Exit(64)
	}
	raiz, err := os.Getwd()
	if err != nil {
		fallar(1, err)
	}
	// E002/F002/S003: dentro de un hijo, avisar si su lock y este ejecutable no van a la par.
	// D000043: AQUAS_VERSION fija solo la versión que compara el aviso (las pruebas corren un binario devel).
	// No toca versionBinario: update sigue sellando el lock con la versión real.
	vAviso := os.Getenv("AQUAS_VERSION")
	if vAviso == "" {
		vAviso = versionBinario()
	}
	avisarVersion(os.Stderr, os.Args[1], raiz, vAviso)
	switch os.Args[1] {
	case "compilar":
		inf, err := compilador.Compilar(raiz)
		if err != nil {
			fallar(1, err)
		}
		reportar(os.Stdout, inf)
		compilarSettings(os.Stdout)
		docs, err := wiki.Generar(raiz) // E001/F002/S010: hijos de cada página desde las fichas
		if err != nil {
			fallar(1, err)
		}
		for _, d := range docs {
			fmt.Println("Escrito: " + d)
		}
	case "iniciar":
		e, err := entorno.Iniciar(raiz)
		if err != nil {
			fallar(1, err)
		}
		fmt.Printf("Entorno declarado en %s · so=%s · raiz=%s\nConfírmalo: si no es esta máquina y esta carpeta, bórralo.\n", entorno.RutaEntorno, e.SO, e.Raiz)
		compilarSettings(os.Stdout)
	case "verificar":
		if hijo.EsHijo(raiz) { // E002/F001/S002/X02: un hijo no tiene la config de la central
			if m := hijo.Verificar(raiz); len(m) > 0 {
				listar("GATE FALLIDO · hijo:", m)
				os.Exit(1)
			}
			fmt.Println("GATE OK · hijo: central sellada y CLAUDE.md al día")
			return
		}
		calidadGo := revisarGo(raiz) // E001/F013/S001: primero, para que se vea aunque otro chequeo corte
		if len(calidadGo) > 0 {
			listar("GATE FALLIDO · calidad Go:", calidadGo)
		}
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
		if exe, err := os.Executable(); err == nil { // D000026: binario viejo = gate falso
			if m := entorno.BinarioViejo(raiz, exe); m != "" {
				permitidos = append(permitidos, m)
			}
		}
		if m := entorno.Validar(raiz); m != "" {
			permitidos = append(permitidos, "entorno: "+m)
		} else if m := entorno.VerificarSettings(raiz); m != "" {
			permitidos = append(permitidos, m)
		}
		if len(derivados) > 0 {
			listar("GATE FALLIDO · derivados desfasados:", derivados)
			_, _ = fmt.Fprintln(os.Stderr, "Desfase → go run ./cmd/aquas compilar · presupuesto → recortar canónico")
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
		if len(derivados)+len(docs)+len(fichas)+len(permitidos)+len(calidadGo) > 0 {
			os.Exit(1)
		}
		fmt.Println("GATE OK · derivados, sello, wiki, backlog y guardia al día")
	case "hook":
		hook(raiz)
	case "turno": // hooks UserPromptSubmit y Stop: nunca bloquean; fallo = exit 1 + stderr
		turnoCLI(raiz, os.Args[2:])
	case "aprobacion": // hook UserPromptSubmit (E001/F010/S001): «apruebo <id>» del humano sella el contrato
		aprobacionCLI(raiz)
	case "presupuesto": // hooks UserPromptSubmit y PreToolUse (E001/F015/S002): tope por feature y freno por turno
		presupuestoCLI(raiz, os.Args[2:])
	case "propuesta": // hook Stop (E001/F002/S006): PROPUESTA sin diagrama → exit 2 + stderr = el agente lo agrega
		var e struct {
			Respuesta string `json:"last_assistant_message"`
		}
		datos, _ := io.ReadAll(os.Stdin)
		if err := json.Unmarshal(datos, &e); err != nil {
			fallar(1, fmt.Errorf("propuesta: entrada del hook ilegible: %w", err))
		}
		// E001/F016/S001: sin entorno legible no hay shell declarado y no se revisa (el cierre no depende del entorno;
		// un entorno roto ya lo frena la guardia). Un shell_humano inválido lo frena RevisarShell.
		en, err := entorno.Leer(raiz)
		if err != nil { // Leer devuelve lo decodificado junto al error: no usar un entorno inválido
			en = entorno.Entorno{}
		}
		for _, m := range []string{propuesta.Revisar(e.Respuesta), propuesta.RevisarInsight(e.Respuesta), propuesta.RevisarShell(e.Respuesta, en.ShellHumano), defecto.RevisarFallo(raiz, e.Respuesta)} { // S016: insight PROPUESTO exige defecto; E001/F004/S002: fallo exige defecto
			if m != "" {
				_, _ = fmt.Fprintln(os.Stderr, m)
				os.Exit(2)
			}
		}
	case "limpiar": // E000/F006/S009: borra .tmp/<id>/ de un encargo y nada más
		msg, err := limpiar(raiz, os.Args[2:])
		if err != nil {
			fallar(1, err)
		}
		fmt.Println(msg)
	case "defecto": // E001/F004/S001: registro de defectos en registro/defectos/
		if err := defectoCLI(raiz, os.Args[2:]); err != nil {
			fallar(1, err)
		}
	case "wiki": // E000/F004/S003: aquas wiki enlazar docs/x.md… pone los enlaces al glosario que pide el gate
		if len(os.Args) < 4 || os.Args[2] != "enlazar" {
			fallar(64, errors.New("uso: aquas wiki enlazar docs/x.md…"))
		}
		for _, rel := range os.Args[3:] {
			n, err := wiki.Enlazar(raiz, filepath.ToSlash(rel))
			if err != nil {
				fallar(1, err)
			}
			fmt.Printf("%s: %d enlaces al glosario\n", rel, n)
		}
	case "medir": // E000/F004/S001: corridas y fallas de pruebas y gate, solo lectura del historial
		if len(os.Args) < 3 || os.Args[2] != "vueltas" {
			fallar(64, errors.New("uso: aquas medir vueltas [historial.jsonl…]"))
		}
		home, err := os.UserHomeDir()
		if err == nil {
			err = vueltas.Medir(os.Stdout, home, raiz, os.Args[3:])
		}
		if err != nil {
			fallar(1, err)
		}
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
	case "build": // arma el CLAUDE.md del hijo: central sellada + huecos locales
		if len(os.Args) != 3 {
			fallar(1, errors.New("uso: aquas build <carpeta-del-hijo>"))
		}
		if err := hijo.Build(os.Args[2]); err != nil {
			fallar(2, err) // exit 2: Claude Code muestra el error al modelo (E002/F002/S006)
		}
		fmt.Printf("CLAUDE.md armado en %s\n", os.Args[2])
	case "report": // borrador de issue para la central; no envía: lo decide el humano
		if len(os.Args) != 5 {
			fallar(1, errors.New("uso: aquas report <carpeta-del-hijo> \"<qué pasó>\" \"<cómo reproducirlo>\""))
		}
		body, err := hijo.Report(os.Args[2], os.Args[3], os.Args[4])
		if err != nil {
			fallar(1, err)
		}
		f := filepath.Join(os.Args[2], hijo.Carpeta, "reporte.md")
		if err := os.WriteFile(f, []byte(body), 0o644); err != nil {
			fallar(1, err)
		}
		fmt.Printf("%s\nBorrador en %s. Nada se envió. Para abrir el issue, revisa y corre:\ngh issue create --repo %s --title %q --body-file %s\n",
			body, f, hijo.RepoCentral, os.Args[3], f)
	case "update": // trae la central de este binario al hijo, sin tocar lo local
		// E002/F002/S001: corre en la carpeta del hijo, sin ruta, como un paquete.
		if len(os.Args) != 2 {
			fallar(1, errors.New("uso: aquas update (en la carpeta del hijo, sin ruta)"))
		}
		rest, err := hijo.Update(raiz, gobcentral.Central, versionBinario())
		if err != nil {
			fallar(1, err)
		}
		for _, p := range rest {
			fmt.Printf("restaurado (estaba alterado; repórtalo a la central): %s\n", p)
		}
		if err := hijo.Listo(raiz); err != nil {
			fallar(1, fmt.Errorf("%w; tras corregirlo repite aquas update", err))
		}
		fmt.Printf("gobierno central %s en %s; CLAUDE.md armado\n", versionBinario(), raiz)
	case "install": // instala el gobierno central en un proyecto hijo
		// E002/F002/S001: corre en la carpeta del hijo, sin ruta, como un paquete.
		if len(os.Args) != 2 {
			fallar(1, errors.New("uso: aquas install (en la carpeta del hijo, sin ruta)"))
		}
		v := "devel"
		if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
			v = bi.Main.Version
		}
		if err := hijo.Install(raiz, gobcentral.Central, v); err != nil {
			fallar(1, err)
		}
		if err := hijo.Listo(raiz); err != nil { // E002/F002/S002: la central ya quedó; reintentar es update
			fallar(1, hijo.AMedias(err))
		}
		fmt.Printf("gobierno central %s instalado en %s; CLAUDE.md armado con huecos PENDIENTE en %s\n", v, raiz, hijo.CarpetaHuecos)
	case "cierre-agente": // hook SubagentStop: exit 2 + stderr = el subagente reescribe su cierre
		var e struct {
			Tipo    string `json:"agent_type"`
			Mensaje string `json:"last_assistant_message"`
			Activo  bool   `json:"stop_hook_active"`
			Agente  string `json:"agent_transcript_path"`
		}
		datos, _ := io.ReadAll(os.Stdin)
		if err := json.Unmarshal(datos, &e); err != nil {
			registrarCierre(raiz, e.Tipo, e.Agente, "", "entrada ilegible") // deja cerrar: un fallo de entrada no traba a los agentes
			return
		}
		if e.Mensaje == "" && e.Agente != "" { // D000053: el cierre real vive en el transcript del subagente
			e.Mensaje = ultimoMensaje(e.Agente)
		}
		primera, _, _ := strings.Cut(strings.TrimSpace(strings.ReplaceAll(e.Mensaje, "\r\n", "\n")), "\n")
		anotar := func(resultado string) { registrarCierre(raiz, e.Tipo, e.Agente, primera, resultado) }
		// E001/F004/S005 (D000068): un tipo que encargo no conoce pasaba en silencio. Revisar con cierre vacío
		// solo da "" para un tipo desconocido; así la lista de tipos sigue en un único lugar (encargo).
		if encargo.Revisar(e.Tipo, "", false) == "" {
			anotar("tipo desconocido")
			if !e.Activo {
				_, _ = fmt.Fprintf(os.Stderr, "aquas cierre-agente: tipo de agente desconocido «%s»; el gancho no sabe revisar su cierre\n", e.Tipo)
				os.Exit(2)
			}
			return
		}
		if m := encargo.Revisar(e.Tipo, e.Mensaje, e.Activo); m != "" {
			// E001/F004/S006 (D000068): tras el rechazo el auditor no reintenta; el veredicto se anota antes de quejarse.
			if id, v, ok := encargo.Veredicto(e.Mensaje); ok && e.Tipo == "auditor" {
				if err := backlog.AnotarAuditoria(raiz, id, v, time.Now()); err != nil {
					anotar("rechazado: " + err.Error())
					_, _ = fmt.Fprintf(os.Stderr, "aquas cierre-agente: el veredicto de %s no quedó anotado: %v\n%s\n", id, err, m)
					os.Exit(2)
				}
				anotar("anotado " + id + " " + v + " · " + m)
				_, _ = fmt.Fprintln(os.Stderr, m)
				os.Exit(2)
			}
			anotar("rechazado: " + m)
			_, _ = fmt.Fprintln(os.Stderr, m)
			os.Exit(2)
		}
		resultado := "aceptado"
		if e.Tipo == "auditor" { // E001/F014/S002: el veredicto queda en la ficha; CERRADO lo exige
			// Sin id y veredicto válidos (stop_hook_active dejó pasar un cierre malo) no se anota nada.
			if id, v, ok := encargo.Veredicto(e.Mensaje); ok {
				if err := backlog.AnotarAuditoria(raiz, id, v, time.Now()); err != nil {
					anotar("rechazado: " + err.Error())
					_, _ = fmt.Fprintf(os.Stderr, "aquas cierre-agente: el veredicto de %s no quedó anotado: %v\n", id, err)
					if !e.Activo { // con stop_hook_active nunca se atrapa al auditor: aviso y sale
						os.Exit(2)
					}
					return
				}
				resultado = "anotado " + id + " " + v
			} else {
				resultado = "sin anotar: el cierre no trae id de tramo y veredicto válidos" // id: — o cierre malo que dejó pasar stop_hook_active
			}
		}
		anotar(resultado)
	case "guardia":
		datos, _ := io.ReadAll(os.Stdin)
		if m := entorno.Validar(raiz); m != "" {
			_, _ = fmt.Fprintln(os.Stderr, "aquas guardia: entorno no verificado, se bloquea todo: "+m)
			os.Exit(2)
		}
		if m := guardia.Revisar(raiz, datos); m != "" {
			_, _ = fmt.Fprintln(os.Stderr, "aquas guardia: "+m)
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

	// E000/F004/S004: página de docs/ → enlaces al glosario que pide el gate, sin depender del agente.
	// Error (p. ej. enlace equivocado previo) no frena aquí: el gate lo reporta.
	if strings.HasPrefix(rel, "docs/") && strings.HasSuffix(rel, ".md") && rel != "docs/glosario.md" {
		if n, err := wiki.Enlazar(raiz, rel); err == nil && n > 0 {
			fmt.Fprintf(os.Stderr, "%s: %d enlaces al glosario puestos por aquas (relee el archivo antes de editarlo otra vez)\n", rel, n)
			os.Exit(2)
		}
		return
	}

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
		_, _ = fmt.Fprintln(w, "Escrito: "+entorno.RutaSettings)
	}
}

// reportar imprime qué derivados escribió la compilación, o que todo estaba al día.
func reportar(w io.Writer, inf compilador.Informe) {
	if len(inf.Escritos) == 0 {
		_, _ = fmt.Fprintln(w, "Sin cambios · derivados al día")
		return
	}
	_, _ = fmt.Fprintln(w, "Recompilado · fuentes cambiadas: "+lista(inf.FuentesCambiadas))
	_, _ = fmt.Fprintln(w, "Escritos: "+lista(inf.Escritos))
}

// listar imprime en stderr un título y cada problema en su viñeta.
func listar(titulo string, problemas []string) {
	_, _ = fmt.Fprintln(os.Stderr, titulo)
	for _, p := range problemas {
		_, _ = fmt.Fprintln(os.Stderr, "  - "+p)
	}
}

// lista une los textos con comas, o «—» si no hay ninguno.
func lista(xs []string) string {
	if len(xs) == 0 {
		return "—"
	}
	return strings.Join(xs, ", ")
}

// fallar imprime el error en stderr y termina con el código dado.
func fallar(codigo int, err error) {
	_, _ = fmt.Fprintln(os.Stderr, "aquas:", err)
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
			_, _ = fmt.Fprintln(os.Stderr, m)
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

// memoriaArchivada: CERRADO exige el archivo de memoria del slice, apunte o no la memoria a él (E001/F002/S009, X01).
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
	if f.Iniciativa+"/"+f.Feature+"/"+f.Slice == id {
		if _, err := os.Stat(filepath.Join(raiz, filepath.FromSlash(memoria.RutaArchivo(f)))); err == nil {
			return nil
		}
		return fmt.Errorf("la memoria apunta a %s y no está archivada: primero `aquas memoria fijar tarea=… siguiente=…`, luego `aquas memoria archivar`, y entonces cierra", id)
	}
	// X01 (D000010): con la memoria en otro foco, el slice igual necesita su archivo de memoria.
	p := strings.Split(id, "/")
	if len(p) != 3 {
		return nil
	}
	rel := memoria.RutaArchivo(memoria.Foco{Iniciativa: p[0], Feature: p[1], Slice: p[2]})
	if _, err := os.Stat(filepath.Join(raiz, filepath.FromSlash(rel))); err == nil {
		return nil
	}
	return fmt.Errorf("%s sin memoria archivada: `aquas memoria fijar slice=%s tarea=… siguiente=…`, luego `aquas memoria archivar`, y entonces cierra", id, p[2])
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

// backlogCLI: nuevo | ver [id] | tabla | prioridad | estado | siguiente | pruebas | rojo | verde.
func backlogCLI(raiz string, args []string) {
	uso := fmt.Errorf("uso: aquas backlog nuevo id=… titulo=… prioridad=… motivo=… que_hace=<una oración, máx. 20 palabras> relacion=nueva|amplia:<id>|corrige:<id> prueba=./paquete:TestNombre=valida|rojo cuando|verde cuando… [bloqueado_por=A,B] | ver [id] | tabla | prioridad <id> <P0..P3|SIN_TRIAGE> motivo=… | estado <id> <ESTADO> | siguiente | pruebas <id> ./paquete:TestNombre=valida|rojo cuando|verde cuando… | documentar <id> ./paquete:TestNombre=valida|rojo cuando|verde cuando… | rojo <id> | verde <id> | sello <id> [<clave> sin_cambio|actualizado motivo=…] | alcance <id> ruta… | alcance <id> ampliar <ruta> motivo=… | arreglo <slice|feature> titulo=… motivo=… [absorbe=<slice>] | codigo <id> [commit…] | editar <id> [titulo=…] [bloqueado_por=A,B] motivo=…")
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
	case args[0] == "tabla" && len(args) == 1: // E001/F007/S002: columnas en el orden de siguiente
		fs, err := backlog.Activas(raiz)
		if err != nil {
			fallar(1, err)
		}
		fmt.Print(backlog.Tabla(fs))
		return
	case args[0] == "ver" && len(args) == 2:
		var archivada bool
		f, archivada, err = backlog.Ver(raiz, args[1])
		if err == nil && archivada {
			fmt.Print("archivada · ")
		}
		detalle = true
	case args[0] == "editar" && len(args) > 2: // E001/F009/S003: solo antes de arrancar
		f, err = backlog.Editar(raiz, args[1], args[2:], time.Now())
		detalle = true
	case args[0] == "documentar" && len(args) > 2: // E001/F008/S007
		f, err = backlog.Documentar(raiz, args[1], args[2:])
		detalle = true
	case args[0] == "codigo" && len(args) >= 2: // E001/F008/S005: recalcula el código tocado
		f, err = backlog.Codigo(raiz, args[1], args[2:])
	case args[0] == "pruebas" && len(args) > 2:
		f, err = backlog.Pruebas(raiz, args[1], args[2:])
		detalle = true
	case args[0] == "rojo" && len(args) == 2:
		f, err = backlog.Rojo(raiz, args[1], time.Now())
		detalle = true
	case args[0] == "verde" && len(args) == 2:
		f, err = backlog.Verde(raiz, args[1], time.Now())
		detalle = true
	case args[0] == "arreglo" && len(args) >= 3: // E001/F008/S006: arreglo con ID bajo su slice o feature
		c := map[string]string{}
		for _, x := range args[2:] {
			k, v, _ := strings.Cut(x, "=")
			c[k] = v
		}
		var a backlog.Arreglo
		if a, err = backlog.Arreglar(raiz, args[1], c["titulo"], c["motivo"], c["absorbe"], time.Now()); err == nil {
			fmt.Printf("%s · %s\n", a.ID, a.Titulo)
			return
		}
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
	case args[0] == "sello" && len(args) == 2: // E001/F002/S004: lista del sello de cierre
		xs, err := backlog.Sello(raiz, args[1])
		if err != nil {
			fallar(1, err)
		}
		fmt.Print(backlog.TextoSello(xs))
		return
	case args[0] == "sello" && len(args) == 5 && strings.HasPrefix(args[4], "motivo="):
		f, err = backlog.Marcar(raiz, args[1], args[2], args[3], strings.TrimPrefix(args[4], "motivo="))
		if err == nil {
			fmt.Println("marcado: " + args[2])
			return
		}
	case args[0] == "alcance" && len(args) == 5 && args[2] == "ampliar" && strings.HasPrefix(args[4], "motivo="): // E001/F002/S007
		f, err = backlog.Ampliar(raiz, args[1], args[3], strings.TrimPrefix(args[4], "motivo="))
	case args[0] == "alcance" && len(args) > 2 && args[2] != "ampliar":
		f, err = backlog.Alcance(raiz, args[1], args[2:])
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
	if args[0] == "siguiente" {
		fmt.Print(backlog.Definir(raiz, f)) // E001/F002/S008
	}
	fmt.Print(backlog.Texto(f))
	if detalle {
		fmt.Print(backlog.Detalle(f))
	}
}

// turnoCLI: auditoría de turnos (E001/F002/S001). Lee el JSON del hook por stdin. No imprime nada
// por stdout: en UserPromptSubmit eso entraría al contexto. Un fallo sale con exit 1, que Claude Code
// trata como error que no bloquea: el mensaje del humano sigue y el agente ve el aviso.
func turnoCLI(raiz string, args []string) {
	if len(args) != 1 || (args[0] != "abrir" && args[0] != "cerrar") {
		fallar(64, fmt.Errorf("uso: aquas turno abrir|cerrar (JSON del hook por stdin)"))
	}
	var e struct {
		Sesion    string `json:"session_id"`
		Prompt    string `json:"prompt"`
		Respuesta string `json:"last_assistant_message"`
	}
	datos, _ := io.ReadAll(os.Stdin)
	if err := json.Unmarshal(datos, &e); err != nil {
		fallar(1, fmt.Errorf("turno %s: entrada del hook ilegible: %w", args[0], err))
	}
	var err error
	if args[0] == "abrir" {
		_, err = turno.Abrir(raiz, e.Sesion, e.Prompt, time.Now())
	} else {
		_, err = turno.Cerrar(raiz, e.Sesion, e.Respuesta, time.Now())
	}
	if err != nil {
		fallar(1, fmt.Errorf("turno %s: %w", args[0], err))
	}
}

// defectoCLI: nuevo | ocurre | ver (E001/F004/S001).
func defectoCLI(raiz string, args []string) error {
	uso := errors.New("uso: aquas defecto nuevo firma=… que=… causa=… arreglo=… cercioro=… relacion=nuevo|repite:<D>|<id de ficha> | ocurre D000001 T000001 | ver [D000001]")
	if len(args) == 0 {
		return uso
	}
	var ds []defecto.Defecto
	switch {
	case args[0] == "nuevo":
		c := map[string]string{}
		for _, a := range args[1:] {
			k, v, ok := strings.Cut(a, "=")
			if !ok {
				return uso
			}
			c[k] = v
		}
		d, err := defecto.Nuevo(raiz, c)
		if err != nil {
			return err
		}
		ds = append(ds, d)
	case args[0] == "ocurre" && len(args) == 3:
		d, err := defecto.Ocurre(raiz, args[1], args[2])
		if err != nil {
			return err
		}
		ds = append(ds, d)
	case args[0] == "ver" && len(args) <= 2:
		todos, err := defecto.Leer(raiz)
		if err != nil {
			return err
		}
		for _, d := range todos {
			if len(args) == 1 || d.ID == args[1] {
				ds = append(ds, d)
			}
		}
		if len(args) == 2 && len(ds) == 0 {
			return fmt.Errorf("no existe el defecto %s", args[1])
		}
	default:
		return uso
	}
	for _, d := range ds {
		fmt.Print(defecto.Texto(d))
	}
	return nil
}

// versionBinario: versión del módulo con que se compiló el binario (= versión del gobierno embebido).
func versionBinario() string {
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		return bi.Main.Version
	}
	return "devel"
}

// registrarCierre (E001/F004/S005, D000068) anota cada llamada del gancho cierre-agente en .tmp/hooks/cierre-agente.jsonl.
// Sin línea tras un subagente = Claude Code no invocó el gancho.
// Un registro que no se puede escribir no cambia la respuesta del gancho.
func registrarCierre(raiz, tipo, transcript, primera, resultado string) {
	dir := filepath.Join(raiz, ".tmp", "hooks")
	if os.MkdirAll(dir, 0o755) != nil {
		return
	}
	l, err := json.Marshal(map[string]string{"fecha": time.Now().UTC().Format(time.RFC3339), "tipo": tipo,
		"transcript": transcript, "primera": primera, "resultado": resultado})
	if err != nil {
		return
	}
	f, err := os.OpenFile(filepath.Join(dir, "cierre-agente.jsonl"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	_, _ = f.Write(append(l, '\n'))
	_ = f.Close()
}
