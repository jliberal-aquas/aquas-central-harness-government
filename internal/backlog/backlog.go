// Package backlog: el backlog en fichas JSON (E001/F002/S002). Una ficha por elemento;
// solo `aquas` las escribe. Lo cerrado o cancelado pasa a trabajo/archivo/.
package backlog

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	Activo  = "trabajo/activo"
	Archivo = "trabajo/archivo"
)

// Ficha: un elemento del backlog (iniciativa, feature o slice).
type Ficha struct {
	ID           string   `json:"id"`
	Tipo         string   `json:"tipo"`
	Titulo       string   `json:"titulo"`
	Estado       string   `json:"estado"`
	Prioridad    string   `json:"prioridad"`
	Motivo       string   `json:"motivo"`
	Padre        string   `json:"padre"`
	BloqueadoPor []string `json:"bloqueado_por"`
	Secuencia    int      `json:"secuencia"`
	// S003: pruebas nombradas "./paquete:TestNombre"; sin pruebas = ficha anterior a S003.
	Pruebas []Prueba  `json:"pruebas,omitempty"`
	Rojo    *Registro `json:"rojo,omitempty"`
	Verde   *Registro `json:"verde,omitempty"`
	// S004: marcas del sello de cierre (clave, marca, motivo).
	Sello []Marca `json:"sello,omitempty"`
	// S007: rutas declaradas del slice y ampliaciones con motivo.
	Alcance      []string     `json:"alcance,omitempty"`
	Ampliaciones []Ampliacion `json:"ampliaciones,omitempty"`

	Aprobacion *Aprobacion `json:"aprobacion,omitempty"` // E001/F010/S001

	Arreglos []Arreglo `json:"arreglos,omitempty"`
	// E001/F008/S005: código de producción tocado, congelado al cerrar.
	Codigo    []string  `json:"codigo,omitempty"`
	Relacion  string    `json:"relacion,omitempty"`  // E001/F009/S001: nueva, amplia:<id> o corrige:<id>
	Ediciones []Edicion `json:"ediciones,omitempty"` // E001/F009/S003: cambios antes de arrancar
}

// Estados y prioridades del canónico E001 (sección Backlog).
var (
	estados     = []string{"BORRADOR", "LISTO", "EN_CURSO", "VERIFICANDO", "CERRADO", "BLOQUEADO", "CANCELADO"}
	prioridades = []string{"P0", "P1", "P2", "P3", "SIN_TRIAGE"}
	reID        = regexp.MustCompile(`^E\d{3}(/F\d{3}(/S\d{3})?)?$`)
)

// terminal: el estado ya no cambia (CERRADO o CANCELADO).
func terminal(estado string) bool { return estado == "CERRADO" || estado == "CANCELADO" }

// esta: x figura en la lista.
func esta(xs []string, x string) bool {
	for _, y := range xs {
		if x == y {
			return true
		}
	}
	return false
}

// tipoYPadre deriva del id: E001 → iniciativa; E001/F002 → feature de E001; E001/F002/S002 → slice.
func tipoYPadre(id string) (tipo, padre string) {
	partes := strings.Split(id, "/")
	tipo = []string{"iniciativa", "feature", "slice"}[len(partes)-1]
	return tipo, strings.Join(partes[:len(partes)-1], "/")
}

// nombre: archivo JSON de una ficha, con «/» del ID cambiadas por «-».
func nombre(id string) string { return strings.ReplaceAll(id, "/", "-") + ".json" }

// todas lee activo y archivo. Una ficha ilegible es error: Validar la detalla.
func todas(raiz string) (act, arc map[string]Ficha, err error) {
	act, arc = map[string]Ficha{}, map[string]Ficha{}
	for _, c := range []struct {
		dir string
		m   map[string]Ficha
	}{{Activo, act}, {Archivo, arc}} {
		archivos, _ := filepath.Glob(filepath.Join(raiz, filepath.FromSlash(c.dir), "*.json"))
		for _, a := range archivos {
			f, e := leer(a)
			if e != nil {
				return nil, nil, fmt.Errorf("%s/%s ilegible: %v (corre `aquas verificar`)", c.dir, filepath.Base(a), e)
			}
			c.m[f.ID] = f
		}
	}
	return act, arc, nil
}

// Todas: las fichas de activo y archivo, ordenadas por id. Para quien solo lee (wiki, E001/F002/S010).
func Todas(raiz string) ([]Ficha, error) {
	act, arc, err := todas(raiz)
	if err != nil {
		return nil, err
	}
	fs := make([]Ficha, 0, len(act)+len(arc))
	for _, m := range []map[string]Ficha{act, arc} {
		for _, f := range m {
			fs = append(fs, f)
		}
	}
	sort.Slice(fs, func(i, j int) bool { return fs[i].ID < fs[j].ID })
	return fs, nil
}

// leer carga una ficha desde su archivo JSON.
func leer(ruta string) (Ficha, error) {
	b, err := os.ReadFile(ruta)
	if err != nil {
		return Ficha{}, err
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	var f Ficha
	if err := d.Decode(&f); err != nil {
		return Ficha{}, err
	}
	if d.More() {
		return Ficha{}, errors.New("texto después de la ficha")
	}
	var crudo map[string]json.RawMessage // los nueve campos presentes; bloqueado_por nunca null
	if err := json.Unmarshal(b, &crudo); err != nil {
		return Ficha{}, err
	}
	for _, k := range []string{"id", "tipo", "titulo", "estado", "prioridad", "motivo", "padre", "bloqueado_por", "secuencia"} {
		if _, ok := crudo[k]; !ok {
			return Ficha{}, fmt.Errorf("falta el campo %s", k)
		}
	}
	if f.BloqueadoPor == nil {
		return Ficha{}, errors.New("bloqueado_por es null: usa []")
	}
	return f, nil
}

// escribir guarda la ficha como JSON en trabajo/<dir>/, con listas vacías en vez de null.
func escribir(raiz, dir string, f Ficha) error {
	if f.BloqueadoPor == nil {
		f.BloqueadoPor = []string{}
	}
	b, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	d := filepath.Join(raiz, filepath.FromSlash(dir))
	if err := os.MkdirAll(d, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(d, nombre(f.ID)), append(b, '\n'), 0o644)
}

// Nuevo crea una ficha desde campos id=… titulo=… prioridad=… motivo=… relacion=… [bloqueado_por=A,B].
func Nuevo(raiz string, campos []string) (Ficha, error) {
	v := map[string]string{}
	var pruebas []string
	for _, c := range campos {
		k, val, ok := strings.Cut(c, "=")
		if !ok || !esta([]string{"id", "titulo", "prioridad", "motivo", "bloqueado_por", "prueba", "relacion"}, k) {
			return Ficha{}, fmt.Errorf("campo desconocido %q: usa id, titulo, prioridad, motivo, bloqueado_por, prueba, relacion", c)
		}
		if k == "prueba" { // se repite: una por prueba
			pruebas = append(pruebas, val)
			continue
		}
		if _, repetido := v[k]; repetido {
			return Ficha{}, fmt.Errorf("campo %s repetido", k)
		}
		v[k] = strings.TrimSpace(val)
	}
	id := v["id"]
	if !reID.MatchString(id) {
		return Ficha{}, fmt.Errorf("id %q inválido: E001, E001/F001 o E001/F001/S001", id)
	}
	for _, k := range []string{"titulo", "motivo"} {
		if v[k] == "" {
			return Ficha{}, fmt.Errorf("falta %s", k)
		}
	}
	if !esta(prioridades, v["prioridad"]) {
		return Ficha{}, fmt.Errorf("prioridad %q inválida: %s", v["prioridad"], strings.Join(prioridades, " | "))
	}
	ps, err := listaPruebas(pruebas)
	if err != nil {
		return Ficha{}, err
	}
	act, arc, err := todas(raiz)
	if err != nil {
		return Ficha{}, err
	}
	existe := func(x string) bool { _, a := act[x]; _, b := arc[x]; return a || b }
	if existe(id) {
		return Ficha{}, fmt.Errorf("%s ya existe: un id nunca se reutiliza", id)
	}
	tipo, padre := tipoYPadre(id)
	if padre != "" && !existe(padre) {
		return Ficha{}, fmt.Errorf("padre %s no existe: créalo primero", padre)
	}
	bloq := []string{}
	if v["bloqueado_por"] != "" {
		for _, b := range strings.Split(v["bloqueado_por"], ",") {
			if b = strings.TrimSpace(b); !existe(b) {
				return Ficha{}, fmt.Errorf("bloqueado_por: %s no existe", b)
			}
			if esta(bloq, b) {
				return Ficha{}, fmt.Errorf("bloqueado_por: %s repetido", b)
			}
			bloq = append(bloq, b)
		}
	}
	if err := revisarRelacion(raiz, v["relacion"], v["titulo"], act, arc); err != nil {
		return Ficha{}, err
	}
	sec := 0
	for _, m := range []map[string]Ficha{act, arc} {
		for _, f := range m {
			if f.Secuencia > sec {
				sec = f.Secuencia
			}
		}
	}
	f := Ficha{ID: id, Tipo: tipo, Titulo: v["titulo"], Estado: "BORRADOR", Prioridad: v["prioridad"],
		Motivo: v["motivo"], Padre: padre, BloqueadoPor: bloq, Secuencia: sec + 1}
	f.Pruebas, f.Relacion = ps, v["relacion"]
	return f, escribir(raiz, Activo, f)
}

// Ver busca la ficha en activo y luego en archivo.
func Ver(raiz, id string) (Ficha, bool, error) {
	act, arc, err := todas(raiz)
	if err != nil {
		return Ficha{}, false, err
	}
	if f, ok := act[id]; ok {
		return f, false, nil
	}
	if f, ok := arc[id]; ok {
		return f, true, nil
	}
	return Ficha{}, false, fmt.Errorf("%s no existe", id)
}

// activa: la ficha a cambiar; una archivada no cambia.
func activa(raiz, id string) (Ficha, error) {
	f, archivada, err := Ver(raiz, id)
	if err == nil && archivada {
		err = fmt.Errorf("%s está archivada (%s): no cambia", id, f.Estado)
	}
	return f, err
}

// Prioridad cambia la prioridad; el motivo es obligatorio.
func Prioridad(raiz, id, prioridad, motivo string) (Ficha, error) {
	if !esta(prioridades, prioridad) {
		return Ficha{}, fmt.Errorf("prioridad %q inválida: %s", prioridad, strings.Join(prioridades, " | "))
	}
	if strings.TrimSpace(motivo) == "" {
		return Ficha{}, errors.New("falta motivo: motivo=…")
	}
	f, err := activa(raiz, id)
	if err != nil {
		return Ficha{}, err
	}
	f.Prioridad, f.Motivo = prioridad, strings.TrimSpace(motivo)
	return f, escribir(raiz, Activo, f)
}

// Estado cambia el estado. CERRADO o CANCELADO mueve la ficha al archivo.
func Estado(raiz, id, estado string) (Ficha, error) {
	if !esta(estados, estado) {
		return Ficha{}, fmt.Errorf("estado %q inválido: %s", estado, strings.Join(estados, " | "))
	}
	act, _, err := todas(raiz)
	if err != nil {
		return Ficha{}, err
	}
	f, err := activa(raiz, id)
	if err != nil {
		return Ficha{}, err
	}
	if cur, hay := sliceEnCurso(act); f.Tipo == "slice" && hay && cur != id {
		return Ficha{}, fmt.Errorf("%s está en curso: mientras tanto ningún otro slice cambia de estado", cur)
	}
	if terminal(estado) { // E001/F002/S017: un padre no termina con hijos activos
		if err := sinHijosActivos(act, f); err != nil {
			return Ficha{}, err
		}
	}
	if estado == "EN_CURSO" && f.Tipo == "slice" { // E001/F010/S001: contrato aprobado antes de construir
		if err := aprobada(raiz, f); err != nil {
			return Ficha{}, err
		}
	}
	if estado == "EN_CURSO" && f.Tipo == "slice" { // D000023: la integral de la feature va en rojo antes de sus slices
		if p, _, err := Ver(raiz, f.Padre); err == nil && len(p.Pruebas) > 0 && p.Rojo == nil {
			return Ficha{}, fmt.Errorf("%s no arranca: la integral de su feature %s no tiene rojo; regístralo antes con aquas backlog rojo %s", id, p.ID, p.ID)
		}
	}
	if estado == "CERRADO" && f.Tipo == "slice" { // E001/F010/S001: el cierre compara la huella aprobada
		if err := aprobada(raiz, f); err != nil {
			return Ficha{}, err
		}
	}
	if estado == "CERRADO" {
		if err := cierre(f); err != nil {
			return Ficha{}, err
		}
		if err := selloCompleto(raiz, f); err != nil {
			return Ficha{}, err
		}
		if f.Tipo == "slice" { // E001/F008/S005: congela el código tocado
			if f.Codigo, err = codigoTocado(raiz, id, true); err != nil {
				return Ficha{}, err
			}
		}
	}
	f.Estado = estado
	if !terminal(estado) {
		return f, escribir(raiz, Activo, f)
	}
	if err := escribir(raiz, Archivo, f); err != nil {
		return Ficha{}, err
	}
	return f, os.Remove(filepath.Join(raiz, filepath.FromSlash(Activo), nombre(id)))
}

// sinHijosActivos: cerrar o cancelar una feature o iniciativa exige sus hijos directos ya archivados.
// Los hijos son las fichas activas cuyo padre es f.
func sinHijosActivos(act map[string]Ficha, f Ficha) error {
	var hijos []Ficha
	for _, h := range act {
		if h.Padre == f.ID {
			hijos = append(hijos, h)
		}
	}
	if len(hijos) == 0 {
		return nil
	}
	sort.Slice(hijos, func(i, j int) bool { return hijos[i].ID < hijos[j].ID })
	var xs []string
	for _, h := range hijos {
		xs = append(xs, h.ID+" ("+h.Estado+")")
	}
	return fmt.Errorf("%s no termina: tiene hijos sin terminar: %s; termina o cancela cada hijo antes", f.ID, strings.Join(xs, ", "))
}

// orden de `siguiente`: en curso → desbloqueado → prioridad → secuencia → id.
func ordenar(fs []Ficha, arc map[string]Ficha) {
	enCurso := func(f Ficha) bool { return f.Estado == "EN_CURSO" || f.Estado == "VERIFICANDO" }
	libre := func(f Ficha) bool { return libre(f, arc) }
	rango := func(p string) int {
		for i, x := range prioridades {
			if x == p {
				return i
			}
		}
		return len(prioridades)
	}
	sort.SliceStable(fs, func(i, j int) bool {
		a, b := fs[i], fs[j]
		switch {
		case enCurso(a) != enCurso(b):
			return enCurso(a)
		case libre(a) != libre(b):
			return libre(a)
		case rango(a.Prioridad) != rango(b.Prioridad):
			return rango(a.Prioridad) < rango(b.Prioridad)
		case a.Secuencia != b.Secuencia:
			return a.Secuencia < b.Secuencia
		}
		return a.ID < b.ID
	})
}

// Activas: slices en el orden de `siguiente`; después iniciativas y features por secuencia.
func Activas(raiz string) ([]Ficha, error) {
	act, arc, err := todas(raiz)
	if err != nil {
		return nil, err
	}
	var slices, resto []Ficha
	for _, f := range act {
		if f.Tipo == "slice" {
			slices = append(slices, f)
		} else {
			resto = append(resto, f)
		}
	}
	ordenar(slices, arc)
	sort.Slice(resto, func(i, j int) bool { return resto[i].Secuencia < resto[j].Secuencia })
	return append(slices, resto...), nil
}

// Siguiente: el primer slice del orden fijo que esté en curso o desbloqueado.
func Siguiente(raiz string) (Ficha, bool, error) {
	act, arc, err := todas(raiz)
	if err != nil {
		return Ficha{}, false, err
	}
	// E001/F002/S008: el slice en curso no se interrumpe; después, definir antes de ejecutar.
	if id, hay := sliceEnCurso(act); hay {
		return act[id], true, nil
	}
	if f, hay := sinDefinir(act); hay {
		return f, true, nil
	}
	var slices []Ficha
	for _, f := range act {
		if f.Tipo == "slice" {
			slices = append(slices, f)
		}
	}
	ordenar(slices, arc)
	for _, f := range slices {
		if libre(f, arc) {
			return f, true, nil
		}
	}
	return Ficha{}, false, nil
}

// sinDefinir (E001/F002/S008, regla del humano 2026-10-21): iniciativa o feature activa sin hijos
// activos. Va antes que cualquier slice, sin mirar prioridad ni bloqueo: iniciativas primero,
// luego features, cada grupo por secuencia de ingreso.
func sinDefinir(act map[string]Ficha) (Ficha, bool) {
	conHijos := map[string]bool{}
	for _, f := range act {
		conHijos[f.Padre] = true
	}
	var mejor Ficha
	hay := false
	for _, f := range act {
		if f.Tipo == "slice" || conHijos[f.ID] {
			continue
		}
		if !hay || (f.Tipo == "iniciativa") != (mejor.Tipo == "iniciativa") && f.Tipo == "iniciativa" ||
			f.Tipo == mejor.Tipo && f.Secuencia < mejor.Secuencia {
			mejor, hay = f, true
		}
	}
	return mejor, hay
}

// Definir: qué hacer con lo que `siguiente` saca sin definir; "" para un slice.
func Definir(f Ficha) string {
	hijos := map[string]string{"iniciativa": "features", "feature": "slices"}[f.Tipo]
	if hijos == "" {
		return ""
	}
	return fmt.Sprintf("definir %s: crea sus %s y reprioriza el backlog antes de seguir con slices\n", f.ID, hijos)
}

// libre: no está BLOQUEADO y cada bloqueador ya terminó (archivado: CERRADO o CANCELADO).
// Un bloqueador cancelado libera: si no, la ficha quedaría trabada sin comando que la suelte.
func libre(f Ficha, arc map[string]Ficha) bool {
	if f.Estado == "BLOQUEADO" {
		return false
	}
	for _, b := range f.BloqueadoPor {
		if _, terminado := arc[b]; !terminado {
			return false
		}
	}
	return true
}

// Validar: problemas de las fichas en disco (gate). Sin carpeta trabajo/ no hay problemas.
func Validar(raiz string) []string {
	var p []string
	ids := map[string]string{}
	secs := map[int]string{}
	var fichas []Ficha
	primero := "" // único slice en curso permitido
	for _, dir := range []string{Activo, Archivo} {
		archivos, _ := filepath.Glob(filepath.Join(raiz, filepath.FromSlash(dir), "*.json"))
		sort.Strings(archivos)
		for _, a := range archivos {
			rel := dir + "/" + filepath.Base(a)
			f, err := leer(a)
			if err != nil {
				p = append(p, fmt.Sprintf("%s: JSON inválido o campo extra: %v", rel, err))
				continue
			}
			switch {
			case !reID.MatchString(f.ID):
				p = append(p, fmt.Sprintf("%s: id %q inválido", rel, f.ID))
				continue
			case nombre(f.ID) != filepath.Base(a):
				p = append(p, fmt.Sprintf("%s: el nombre no cuadra con el id %s", rel, f.ID))
			}
			tipo, padre := tipoYPadre(f.ID)
			if f.Tipo != tipo || f.Padre != padre {
				p = append(p, fmt.Sprintf("%s: tipo o padre no cuadran con el id (%s, %q)", rel, tipo, padre))
			}
			if !esta(estados, f.Estado) || !esta(prioridades, f.Prioridad) || f.Titulo == "" || f.Motivo == "" || f.Secuencia < 1 {
				p = append(p, fmt.Sprintf("%s: estado, prioridad, titulo, motivo o secuencia inválidos", rel))
			}
			if terminal(f.Estado) != (dir == Archivo) {
				p = append(p, fmt.Sprintf("%s: estado %s no corresponde a %s", rel, f.Estado, dir))
			}
			if otro, ok := ids[f.ID]; ok {
				p = append(p, fmt.Sprintf("%s: id %s repite %s", rel, f.ID, otro))
			}
			ids[f.ID] = rel
			p = append(p, problemasRojoVerde(rel, f)...)
			p = append(p, problemasArreglos(rel, f)...)
			if f.Tipo == "slice" && corriendo(f) && dir == Activo {
				if primero != "" {
					p = append(p, fmt.Sprintf("%s: otro slice en curso además de %s; solo uno a la vez", rel, primero))
				} else {
					primero = f.ID
				}
			}
			if otro, ok := secs[f.Secuencia]; ok {
				p = append(p, fmt.Sprintf("%s: secuencia %d repite %s", rel, f.Secuencia, otro))
			}
			secs[f.Secuencia] = rel
			fichas = append(fichas, f)
		}
	}
	for _, f := range fichas {
		if f.Padre != "" && ids[f.Padre] == "" {
			p = append(p, fmt.Sprintf("%s: padre %s no existe", ids[f.ID], f.Padre))
		}
		if f.Padre != "" && strings.HasPrefix(ids[f.ID], Activo+"/") && strings.HasPrefix(ids[f.Padre], Archivo+"/") { // E001/F002/S017
			p = append(p, fmt.Sprintf("%s: hijo activo con padre %s archivado (%s)", ids[f.ID], f.Padre, ids[f.Padre]))
		}
		for _, b := range f.BloqueadoPor {
			if ids[b] == "" {
				p = append(p, fmt.Sprintf("%s: bloqueado_por %s no existe", ids[f.ID], b))
			}
		}
	}
	return p
}

// Texto: una línea por ficha, para el agente.
func Texto(f Ficha) string {
	s := fmt.Sprintf("%s · %s · %s · %s · %s", f.ID, f.Estado, f.Prioridad, f.Titulo, f.Motivo)
	if len(f.BloqueadoPor) > 0 {
		s += " · bloqueado_por " + strings.Join(f.BloqueadoPor, ",")
	}
	return s + "\n"
}
