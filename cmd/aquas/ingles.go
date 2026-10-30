package main

// E000/F006/S006 (humano 2026-10-30, decisiones 1a y 2a): comandos y subcomandos en inglés.
// Estados, campos y mensajes siguen en castellano. Los nombres viejos andan y avisan el nuevo.

// comandos: nombre en inglés → nombre interno del comando de primer nivel.
var comandos = map[string]string{
	"compile": "compilar", "init": "iniciar", "verify": "verificar", "clean": "limpiar", "measure": "medir",
	"turn": "turno", "approval": "aprobacion", "budget": "presupuesto", "proposal": "propuesta",
	"defect": "defecto", "memory": "memoria", "library": "biblioteca",
	"agent-close": "cierre-agente", "guard": "guardia",
	"hook": "hook", "wiki": "wiki", "backlog": "backlog", "build": "build",
	"report": "report", "update": "update", "install": "install",
}

// subcomandos: por comando interno, nombre en inglés → nombre interno del subcomando.
var subcomandos = map[string]map[string]string{
	"backlog": {
		"new": "nuevo", "show": "ver", "table": "tabla", "edit": "editar", "document": "documentar",
		"code": "codigo", "tests": "pruebas", "red": "rojo", "green": "verde", "fix": "arreglo",
		"priority": "prioridad", "state": "estado", "seal": "sello", "scope": "alcance", "next": "siguiente",
	},
	"memoria":     {"start": "inicio", "close": "cierre", "archive": "archivar", "set": "fijar", "show": "ver"},
	"turno":       {"open": "abrir", "close": "cerrar"},
	"presupuesto": {"prompt": "prompt", "tool": "herramienta"},
	"defecto":     {"new": "nuevo", "occur": "ocurre", "show": "ver"},
	"wiki":        {"link": "enlazar"},
	"medir":       {"runs": "vueltas"},
	"biblioteca":  {"read": "leer"},
}

// alcanceAmpliar: la palabra en la cuarta posición de «backlog alcance <id> extend».
var alcanceAmpliar = map[string]string{"extend": "ampliar"}

// traducir devuelve los argumentos con los nombres internos y un aviso por cada nombre viejo usado.
// Solo traduce el comando, su subcomando y la palabra «ampliar» de backlog alcance; nada más.
func traducir(args []string) ([]string, []string) {
	out := append([]string(nil), args...)
	var avisos []string
	if len(out) == 0 {
		return out, nil
	}
	out[0] = traducirUna(out[0], comandos, &avisos)
	if subs, ok := subcomandos[out[0]]; ok && len(out) > 1 {
		out[1] = traducirUna(out[1], subs, &avisos)
	}
	if len(out) > 3 && out[0] == "backlog" && out[1] == "alcance" {
		out[3] = traducirUna(out[3], alcanceAmpliar, &avisos)
	}
	return out, avisos
}

// traducirUna pasa una palabra a su nombre interno; si ya era el nombre viejo, anota el aviso.
func traducirUna(palabra string, tabla map[string]string, avisos *[]string) string {
	if interno, ok := tabla[palabra]; ok {
		return interno
	}
	for nuevo, interno := range tabla {
		if interno == palabra && nuevo != palabra {
			*avisos = append(*avisos, "«"+palabra+"» ahora se escribe «"+nuevo+"»")
			return palabra
		}
	}
	return palabra
}
