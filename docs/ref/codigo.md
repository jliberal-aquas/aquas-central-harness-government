# Código

Índice de cada función del harness, con lo que dice su comentario. Lo genera `aquas compilar` desde el código en `cmd/` e `internal/`.

## En una frase

Quien quiere saber qué hay en el repo lo lee aquí, sin abrir el código.

## Para qué existe

**El problema.** Para saber qué hace una pieza había que abrir el archivo y leer [Go](../glosario.md#go). Un tercio de las funciones no tenía comentario.

**La solución.** El [gate](../glosario.md#gate) falla si una función no lleva comentario y nombra archivo y línea. Este índice se rehace desde el código; si se desfasa, el gate también falla.

## Analogía

Piensa en el **índice de un manual de taller**.

- Cada pieza aparece con su ubicación y una línea de para qué sirve.
- Una pieza sin descripción no entra al manual.

Límite de la analogía: aquí el índice lo escribe un programa a partir del comentario de cada función.

## Diagrama

```mermaid
sequenceDiagram
    participant C as Código Go
    participant A as aquas compilar
    participant G as aquas verificar
    participant P as docs/ref/codigo.md
    A->>C: lee cada función y su comentario
    A->>P: reescribe la tabla
    G->>C: ¿alguna función sin comentario?
    G-->>G: falla y nombra archivo, línea y función
```

## Ejemplo de punta a punta

<a id="EJ-REF-codigo"></a>

1. Al activar el gate, `aquas verificar` nombró 69 funciones sin comentario, por ejemplo `internal/wiki/hijos.go:43: función celda sin comentario`.
2. Cada una recibió una línea que dice qué hace.
3. `aquas compilar` llenó la tabla de abajo y el gate volvió a pasar.

## Paso a paso

1. Recorrer los `.go` de `cmd/` e `internal/`, sin los de prueba.
2. Tomar cada función y método con la primera línea de su comentario.
3. Avisar cada función sin comentario.
4. Escribir la tabla, ordenada por archivo y línea.

## Cómo se usa

Busca el archivo o la función; la última columna dice qué hace.

<!-- GENERADO por aquas desde trabajo/: no editar a mano; `aquas compilar` lo rehace. -->

| Archivo | Función | Parámetros | Qué hace |
|---|---|---|---|
| [cmd/aquas/aprobacion.go](../../cmd/aquas/aprobacion.go) | `idAprobado` | `(prompt string) (string, bool)` | idAprobado: id de la PRIMERA línea no vacía si es «apruebo <id>»; una línea que termina en «?» es pregunta. |
| [cmd/aquas/aprobacion.go](../../cmd/aquas/aprobacion.go) | `aprobacionCLI` | `(raiz string)` | aprobacionCLI: solo un evento UserPromptSubmit cuenta como mensaje del humano. PreToolUse, PostToolUse, |
| [cmd/aquas/aviso.go](../../cmd/aquas/aviso.go) | `avisarVersion` | `(w io.Writer, comando, raiz, version string)` | avisarVersion (E002/F002/S003) escribe el aviso de versión del hijo en w, salvo en install y update. |
| [cmd/aquas/calidad.go](../../cmd/aquas/calidad.go) | `revisarGo` | `(raiz string) []string` | revisarGo (E001/F013/S001) corre la calidad Go de la central para el gate. |
| [cmd/aquas/limpiar.go](../../cmd/aquas/limpiar.go) | `limpiar` | `(raiz string, args []string) (string, error)` | limpiar (E000/F006/S009): borra .tmp/<id>/ de un encargo y nada más. |
| [cmd/aquas/main.go](../../cmd/aquas/main.go) | `main` | `()` | main despacha el subcomando de aquas (compilar, verificar, backlog, hooks…) y sale con su código. |
| [cmd/aquas/main.go](../../cmd/aquas/main.go) | `hook` | `(raiz string)` | hook: exit 0 = nada que reportar; exit 2 = stderr se muestra a Claude. |
| [cmd/aquas/main.go](../../cmd/aquas/main.go) | `compilarSettings` | `(w io.Writer)` | compilarSettings regenera .claude/settings.json; sin entorno declarado no inventa uno. |
| [cmd/aquas/main.go](../../cmd/aquas/main.go) | `reportar` | `(w io.Writer, inf compilador.Informe)` | reportar imprime qué derivados escribió la compilación, o que todo estaba al día. |
| [cmd/aquas/main.go](../../cmd/aquas/main.go) | `listar` | `(titulo string, problemas []string)` | listar imprime en stderr un título y cada problema en su viñeta. |
| [cmd/aquas/main.go](../../cmd/aquas/main.go) | `lista` | `(xs []string) string` | lista une los textos con comas, o «—» si no hay ninguno. |
| [cmd/aquas/main.go](../../cmd/aquas/main.go) | `fallar` | `(codigo int, err error)` | fallar imprime el error en stderr y termina con el código dado. |
| [cmd/aquas/main.go](../../cmd/aquas/main.go) | `memoriaCLI` | `(raiz string, args []string)` | memoriaCLI: `ver` imprime la memoria; `fijar campo=valor…` la actualiza y la sella. |
| [cmd/aquas/main.go](../../cmd/aquas/main.go) | `memoriaArchivada` | `(raiz, id string) error` | memoriaArchivada: CERRADO exige el archivo de memoria del slice, apunte o no la memoria a él (E001/F002/S009, X01). |
| [cmd/aquas/main.go](../../cmd/aquas/main.go) | `focoEnBacklog` | `(raiz string, f memoria.Foco) error` | focoEnBacklog: la memoria solo fija el slice que el backlog tiene en curso (S003). |
| [cmd/aquas/main.go](../../cmd/aquas/main.go) | `backlogCLI` | `(raiz string, args []string)` | backlogCLI: nuevo \\| ver [id] \\| prioridad \\| estado \\| siguiente \\| pruebas \\| rojo \\| verde. |
| [cmd/aquas/main.go](../../cmd/aquas/main.go) | `turnoCLI` | `(raiz string, args []string)` | turnoCLI: auditoría de turnos (E001/F002/S001). Lee el JSON del hook por stdin. No imprime nada |
| [cmd/aquas/main.go](../../cmd/aquas/main.go) | `defectoCLI` | `(raiz string, args []string) error` | defectoCLI: nuevo \\| ocurre \\| ver (E001/F004/S001). |
| [cmd/aquas/main.go](../../cmd/aquas/main.go) | `versionBinario` | `() string` | versionBinario: versión del módulo con que se compiló el binario (= versión del gobierno embebido). |
| [cmd/aquas/presupuesto.go](../../cmd/aquas/presupuesto.go) | `exentaDelFreno` | `(raiz string, e entradaPresupuesto) bool` | exentaDelFreno: true si la llamada es Bash y su comando entero es `aquas memoria fijar campo=valor…`. |
| [cmd/aquas/presupuesto.go](../../cmd/aquas/presupuesto.go) | `presupuestoCLI` | `(raiz string, args []string)` | presupuestoCLI: «prompt» reinicia el turno; «herramienta» cuenta la llamada y bloquea con exit 2 al pasar un tope. |
| [cmd/aquas/presupuesto.go](../../cmd/aquas/presupuesto.go) | `frenoFeature` | `(raiz, dir string) string` | frenoFeature cuenta la llamada contra la feature del foco si está aprobada; "" si puede seguir. |
| [cmd/aquas/presupuesto.go](../../cmd/aquas/presupuesto.go) | `anotar` | `(ruta string) int` | anotar suma una línea al archivo y devuelve el total. |
| [cmd/aquas/presupuesto.go](../../cmd/aquas/presupuesto.go) | `contextoMedido` | `(ruta string) int` | contextoMedido: tokens de entrada de la última llamada del modelo en el transcript; 0 si después |
| [cmd/aquas/transcript.go](../../cmd/aquas/transcript.go) | `ultimoMensaje` | `(ruta string) string` | ultimoMensaje (D000053): SubagentStop no trae last_assistant_message en esta versión de Claude Code |
| [internal/backlog/alcance.go](../../internal/backlog/alcance.go) | `limpia` | `(r string) string` | limpia normaliza una ruta: barras «/», sin «./» inicial ni segmentos redundantes. |
| [internal/backlog/alcance.go](../../internal/backlog/alcance.go) | `Alcance` | `(raiz, id string, rutas []string) (Ficha, error)` | Alcance declara (reemplaza) las rutas del slice. |
| [internal/backlog/alcance.go](../../internal/backlog/alcance.go) | `Ampliar` | `(raiz, id, ruta, motivo string) (Ficha, error)` | Ampliar agrega una ruta al alcance; el motivo es obligatorio y de una línea. |
| [internal/backlog/alcance.go](../../internal/backlog/alcance.go) | `cubre` | `(patron, rel string) bool` | cubre: «dir/**» cubre todo lo de dentro; si no, patrón de path.Match o ruta exacta. |
| [internal/backlog/alcance.go](../../internal/backlog/alcance.go) | `FueraDeAlcance` | `(raiz, rel string) string` | FueraDeAlcance: "" si rel (relativa a la raíz) se puede tocar; si no, el motivo y cómo seguir. |
| [internal/backlog/aprobacion.go](../../internal/backlog/aprobacion.go) | `seccionContrato` | `(texto string) (string, bool)` | seccionContrato: texto de «## Para qué existe» de la página, sin los bloques GENERADO de aquas. |
| [internal/backlog/aprobacion.go](../../internal/backlog/aprobacion.go) | `Huella` | `(raiz string, f Ficha) (string, error)` | Huella: SHA-256 de las pruebas, el alcance (con sus ampliaciones) y la sección «Para qué existe» de la página. |
| [internal/backlog/aprobacion.go](../../internal/backlog/aprobacion.go) | `Aprobar` | `(raiz, id, mensaje string, ahora time.Time) (Ficha, error)` | Aprobar registra la aprobación del contrato de un slice o de una feature (con la foto de sus slices) con la huella |
| [internal/backlog/aprobacion.go](../../internal/backlog/aprobacion.go) | `aprobada` | `(raiz string, f Ficha) error` | aprobada: nil si el slice puede avanzar (EN_CURSO, rojo, verde, cierre); si no, el paso que falta. |
| [internal/backlog/aprobacion.go](../../internal/backlog/aprobacion.go) | `cubiertaPorFeature` | `(raiz string, p, f Ficha, aprobadas []Prueba) error` | cubiertaPorFeature (E001/F015/S001): el slice avanza si el contrato de su feature no cambió y cada prueba |
| [internal/backlog/arreglo.go](../../internal/backlog/arreglo.go) | `idArreglo` | `(padre string, n int) string` | idArreglo: ID del n-ésimo arreglo (1, 2…) de padre: padre/X01. |
| [internal/backlog/arreglo.go](../../internal/backlog/arreglo.go) | `dirDe` | `(archivada bool) string` | dirDe: carpeta de trabajo/ donde vive la ficha, según esté archivada o no. |
| [internal/backlog/arreglo.go](../../internal/backlog/arreglo.go) | `Arreglar` | `(raiz, padre, titulo, motivo, absorbe string, ahora time.Time) (Arreglo, error)` | Arreglar anota un arreglo en la ficha de su slice o feature, abierta o archivada. |
| [internal/backlog/arreglo.go](../../internal/backlog/arreglo.go) | `problemasArreglos` | `(rel string, f Ficha) []string` | problemasArreglos: chequeo del gate: IDs en secuencia bajo su ficha, con título y motivo. |
| [internal/backlog/arreglo_pruebas.go](../../internal/backlog/arreglo_pruebas.go) | `padreDeArreglo` | `(id string) (string, bool)` | padreDeArreglo: el ID del slice o feature dueño, si id es de un arreglo (padre/Xnn). |
| [internal/backlog/arreglo_pruebas.go](../../internal/backlog/arreglo_pruebas.go) | `cambiarArreglo` | `(raiz, id string, cambio func(a *Arreglo) error) (Ficha, error)` | cambiarArreglo carga la ficha dueña, aplica cambio al arreglo id y la guarda donde vive (activa o archivada). |
| [internal/backlog/arreglo_pruebas.go](../../internal/backlog/arreglo_pruebas.go) | `rojoDe` | `(raiz string, ps []Prueba, ahora time.Time) (*Registro, error)` | rojoDe corre cada prueba sola; exige que cada una falle por sí misma. |
| [internal/backlog/arreglo_pruebas.go](../../internal/backlog/arreglo_pruebas.go) | `verdeDe` | `(raiz, id string, ps []Prueba, rojo *Registro, ahora time.Time) (*Registro, error)` | verdeDe corre las pruebas juntas tras el rojo; exige que todas pasen. |
| [internal/backlog/arreglo_pruebas.go](../../internal/backlog/arreglo_pruebas.go) | `pruebasArreglo` | `(raiz, id string, ps []Prueba) (Ficha, error)` | pruebasArreglo fija las pruebas del arreglo; cambiarlas borra su rojo y su verde. |
| [internal/backlog/arreglo_pruebas.go](../../internal/backlog/arreglo_pruebas.go) | `rojoArreglo` | `(raiz, id string, ahora time.Time) (Ficha, error)` | rojoArreglo registra el rojo del arreglo. |
| [internal/backlog/arreglo_pruebas.go](../../internal/backlog/arreglo_pruebas.go) | `verdeArreglo` | `(raiz, id string, ahora time.Time) (Ficha, error)` | verdeArreglo registra el verde del arreglo. |
| [internal/backlog/auditoria.go](../../internal/backlog/auditoria.go) | `AnotarAuditoria` | `(raiz, id, veredicto string, ahora time.Time) error` | AnotarAuditoria anota el veredicto del auditor en la ficha de un slice EN_CURSO; gana el último. |
| [internal/backlog/auditoria.go](../../internal/backlog/auditoria.go) | `veredictoVigente` | `(f Ficha) error` | veredictoVigente: CERRADO de un slice exige el último veredicto APROBADO, posterior al verde y sobre el mismo código. |
| [internal/backlog/auditoria.go](../../internal/backlog/auditoria.go) | `derefHuella` | `(h *map[string]string) map[string]string` | derefHuella: la huella apuntada, o nil si no hay. |
| [internal/backlog/backlog.go](../../internal/backlog/backlog.go) | `terminal` | `(estado string) bool` | terminal: el estado ya no cambia (CERRADO o CANCELADO). |
| [internal/backlog/backlog.go](../../internal/backlog/backlog.go) | `esta` | `(xs []string, x string) bool` | esta: x figura en la lista. |
| [internal/backlog/backlog.go](../../internal/backlog/backlog.go) | `tipoYPadre` | `(id string) (tipo, padre string)` | tipoYPadre deriva del id: E001 → iniciativa; E001/F002 → feature de E001; E001/F002/S002 → slice. |
| [internal/backlog/backlog.go](../../internal/backlog/backlog.go) | `nombre` | `(id string) string` | nombre: archivo JSON de una ficha, con «/» del ID cambiadas por «-». |
| [internal/backlog/backlog.go](../../internal/backlog/backlog.go) | `todas` | `(raiz string) (act, arc map[string]Ficha, err error)` | todas lee activo y archivo. Una ficha ilegible es error: Validar la detalla. |
| [internal/backlog/backlog.go](../../internal/backlog/backlog.go) | `Todas` | `(raiz string) ([]Ficha, error)` | Todas: las fichas de activo y archivo, ordenadas por id. Para quien solo lee (wiki, E001/F002/S010). |
| [internal/backlog/backlog.go](../../internal/backlog/backlog.go) | `leer` | `(ruta string) (Ficha, error)` | leer carga una ficha desde su archivo JSON. |
| [internal/backlog/backlog.go](../../internal/backlog/backlog.go) | `escribir` | `(raiz, dir string, f Ficha) error` | escribir guarda la ficha como JSON en trabajo/<dir>/, con listas vacías en vez de null. |
| [internal/backlog/backlog.go](../../internal/backlog/backlog.go) | `Nuevo` | `(raiz string, campos []string) (Ficha, error)` | Nuevo crea una ficha desde campos id=… titulo=… prioridad=… motivo=… relacion=… [bloqueado_por=A,B]. |
| [internal/backlog/backlog.go](../../internal/backlog/backlog.go) | `Ver` | `(raiz, id string) (Ficha, bool, error)` | Ver busca la ficha en activo y luego en archivo. |
| [internal/backlog/backlog.go](../../internal/backlog/backlog.go) | `activa` | `(raiz, id string) (Ficha, error)` | activa: la ficha a cambiar; una archivada no cambia. |
| [internal/backlog/backlog.go](../../internal/backlog/backlog.go) | `Prioridad` | `(raiz, id, prioridad, motivo string) (Ficha, error)` | Prioridad cambia la prioridad; el motivo es obligatorio. |
| [internal/backlog/backlog.go](../../internal/backlog/backlog.go) | `Estado` | `(raiz, id, estado string) (Ficha, error)` | Estado cambia el estado. CERRADO o CANCELADO mueve la ficha al archivo. |
| [internal/backlog/backlog.go](../../internal/backlog/backlog.go) | `sinHijosActivos` | `(act map[string]Ficha, f Ficha) error` | sinHijosActivos: cerrar o cancelar una feature o iniciativa exige sus hijos directos ya archivados. |
| [internal/backlog/backlog.go](../../internal/backlog/backlog.go) | `ordenar` | `(fs []Ficha, arc map[string]Ficha)` | orden de `siguiente`: en curso → desbloqueado → prioridad → secuencia → id. |
| [internal/backlog/backlog.go](../../internal/backlog/backlog.go) | `Activas` | `(raiz string) ([]Ficha, error)` | Activas: slices en el orden de `siguiente`; después iniciativas y features por secuencia. |
| [internal/backlog/backlog.go](../../internal/backlog/backlog.go) | `Siguiente` | `(raiz string) (Ficha, bool, error)` | Siguiente: el primer slice del orden fijo que esté en curso o desbloqueado. |
| [internal/backlog/backlog.go](../../internal/backlog/backlog.go) | `sinDefinir` | `(act map[string]Ficha) (Ficha, bool)` | sinDefinir (E001/F002/S008, regla del humano 2026-10-21): iniciativa o feature activa sin hijos |
| [internal/backlog/backlog.go](../../internal/backlog/backlog.go) | `Definir` | `(raiz string, f Ficha) string` | Definir: qué hacer con lo que `siguiente` saca sin definir; "" para un slice. |
| [internal/backlog/backlog.go](../../internal/backlog/backlog.go) | `libre` | `(f Ficha, arc map[string]Ficha) bool` | libre: no está BLOQUEADO y cada bloqueador ya terminó (archivado: CERRADO o CANCELADO). |
| [internal/backlog/backlog.go](../../internal/backlog/backlog.go) | `Validar` | `(raiz string) []string` | Validar: problemas de las fichas en disco (gate). Sin carpeta trabajo/ no hay problemas. |
| [internal/backlog/backlog.go](../../internal/backlog/backlog.go) | `Texto` | `(f Ficha) string` | Texto: una línea por ficha, para el agente. |
| [internal/backlog/codigo.go](../../internal/backlog/codigo.go) | `codigoTocado` | `(raiz, id string, sinCommit bool) ([]string, error)` | codigoTocado: .go de producción en cmd/ e internal/ que tocaron los commits del slice y, si sinCommit, los cambios pendientes (E001/F008/S005). |
| [internal/backlog/codigo.go](../../internal/backlog/codigo.go) | `soloProduccion` | `(log []string) []string` | soloProduccion: .go de cmd/ e internal/ sin pruebas y archivos de configuración (settings, .githooks, gobierno/config sin sellos), sin repetir y ordenados. |
| [internal/backlog/codigo.go](../../internal/backlog/codigo.go) | `Codigo` | `(raiz, id string, commits []string) (Ficha, error)` | Codigo recalcula y guarda el código tocado de un slice desde los commits que nombran su ID y los commits dados. |
| [internal/backlog/documentar.go](../../internal/backlog/documentar.go) | `Documentar` | `(raiz, id string, lista []string) (Ficha, error)` | Documentar reescribe valida, rojo cuando y verde cuando de pruebas ya registradas. |
| [internal/backlog/editar.go](../../internal/backlog/editar.go) | `Editar` | `(raiz, id string, campos []string, ahora time.Time) (Ficha, error)` | Editar cambia título o bloqueado_por de una ficha en BORRADOR o LISTO y guarda campo, antes, después y motivo. |
| [internal/backlog/huella_codigo.go](../../internal/backlog/huella_codigo.go) | `huellaVerde` | `(raiz string, f Ficha) (*map[string]string, error)` | huellaVerde: huella para el registro del verde. Nunca nil: sin archivos Go queda vacía, |
| [internal/backlog/huella_codigo.go](../../internal/backlog/huella_codigo.go) | `baseFija` | `(p string) string` | baseFija: carpeta del patrón antes del primer tramo con comodín («internal/*/x.go» → «internal»). |
| [internal/backlog/huella_codigo.go](../../internal/backlog/huella_codigo.go) | `fueraDelRepo` | `(p string) bool` | fueraDelRepo: patrón absoluto, con unidad («C:») o con un tramo «..»; la huella no lee fuera de la raíz. |
| [internal/backlog/huella_codigo.go](../../internal/backlog/huella_codigo.go) | `huellaCodigo` | `(raiz string, f Ficha) (map[string]string, error)` | huellaCodigo: ruta relativa con «/» → SHA-256 de cada .go cubierto por el alcance (con ampliaciones). |
| [internal/backlog/huella_codigo.go](../../internal/backlog/huella_codigo.go) | `codigoCambio` | `(raiz string, f Ficha) ([]string, error)` | codigoCambio: rutas (ordenadas) que difieren entre la huella del verde y el código de ahora. |
| [internal/backlog/relacion.go](../../internal/backlog/relacion.go) | `palabras` | `(s string) map[string]bool` | palabras: conjunto de palabras de tres o más letras, en minúscula. |
| [internal/backlog/relacion.go](../../internal/backlog/relacion.go) | `parecidas` | `(raiz, titulo string, fichas ...map[string]Ficha) []parecida` | parecidas: hasta maxParecidas fichas (activas y archivadas) o defectos con más palabras en común con titulo. |
| [internal/backlog/relacion.go](../../internal/backlog/relacion.go) | `revisarRelacion` | `(raiz, rel, titulo string, act, arc map[string]Ficha) error` | revisarRelacion valida el campo relacion contra las fichas existentes; sin él, lista las parecidas. |
| [internal/backlog/relacion.go](../../internal/backlog/relacion.go) | `Parecidas` | `(raiz, texto string) ([]string, error)` | Parecidas: líneas «id · texto» de las fichas y defectos con más palabras en común con texto (E001/F009/S002). |
| [internal/backlog/rojo.go](../../internal/backlog/rojo.go) | `problemaPrueba` | `(p Prueba) string` | problemaPrueba: por qué una prueba de la ficha está mal escrita ("" si está bien). |
| [internal/backlog/rojo.go](../../internal/backlog/rojo.go) | `listaPruebas` | `(xs []string) ([]Prueba, error)` | listaPruebas convierte "./paquete:TestNombre=intención" en pruebas válidas, sin repetidas. |
| [internal/backlog/rojo.go](../../internal/backlog/rojo.go) | `corriendo` | `(f Ficha) bool` | corriendo: la ficha está EN_CURSO o VERIFICANDO. |
| [internal/backlog/rojo.go](../../internal/backlog/rojo.go) | `sliceEnCurso` | `(act map[string]Ficha) (string, bool)` | sliceEnCurso: el primer slice EN_CURSO o VERIFICANDO entre las activas, por id. |
| [internal/backlog/rojo.go](../../internal/backlog/rojo.go) | `SliceEnCurso` | `(raiz string) (string, bool, error)` | SliceEnCurso: el único slice EN_CURSO o VERIFICANDO, si hay. |
| [internal/backlog/rojo.go](../../internal/backlog/rojo.go) | `Pruebas` | `(raiz, id string, lista []string) (Ficha, error)` | Pruebas fija las pruebas de la ficha; cambiarlas borra rojo y verde. |
| [internal/backlog/rojo.go](../../internal/backlog/rojo.go) | `corrible` | `(raiz, id string) (Ficha, error)` | corrible: ficha activa con pruebas; un slice además debe estar en curso. |
| [internal/backlog/rojo.go](../../internal/backlog/rojo.go) | `Rojo` | `(raiz, id string, ahora time.Time) (Ficha, error)` | Rojo corre cada prueba sola con `go test -json`; exige que cada una falle por sí misma. |
| [internal/backlog/rojo.go](../../internal/backlog/rojo.go) | `rojoHeredado` | `(raiz string, f Ficha) *Registro` | rojoHeredado (X02): rojo que un slice cerrado de la feature registró para todas sus pruebas |
| [internal/backlog/rojo.go](../../internal/backlog/rojo.go) | `Verde` | `(raiz, id string, ahora time.Time) (Ficha, error)` | Verde corre las pruebas juntas tras el rojo; exige que todas pasen. |
| [internal/backlog/rojo.go](../../internal/backlog/rojo.go) | `resultado.motivo` | `(siPasa string) string` | motivo: texto que explica el resultado de una prueba; siPasa se usa cuando pasó. |
| [internal/backlog/rojo.go](../../internal/backlog/rojo.go) | `correr` | `(raiz string, ps []Prueba) (map[string]resultado, error)` | correr: `go test -json -count=1 -run ^(A\\|B)$ ./paquete` por paquete; resultado por prueba. |
| [internal/backlog/rojo.go](../../internal/backlog/rojo.go) | `motivoDelFallo` | `(ls []string) string` | motivoDelFallo: la última línea «archivo.go:N: …» (la que hizo fallar); si no hay, la primera. |
| [internal/backlog/rojo.go](../../internal/backlog/rojo.go) | `head` | `(raiz string) string` | head: commit actual del repo, o "" fuera de Git. |
| [internal/backlog/rojo.go](../../internal/backlog/rojo.go) | `cierre` | `(raiz string, f Ficha) error` | cierre: CERRADO exige pruebas, rojo, verde y, si el verde trae huella, el mismo código Go. |
| [internal/backlog/rojo.go](../../internal/backlog/rojo.go) | `problemasRojoVerde` | `(rel string, f Ficha) []string` | problemasRojoVerde: chequeos del gate sobre pruebas, rojo y verde de una ficha. |
| [internal/backlog/rojo.go](../../internal/backlog/rojo.go) | `Detalle` | `(f Ficha) string` | Detalle: pruebas con su intención, y rojo y verde, para `ver <id>`, `pruebas`, `rojo` y `verde`. |
| [internal/backlog/sello.go](../../internal/backlog/sello.go) | `excluida` | `(p string) bool` | excluida: la ruta cae bajo una carpeta que el sello no revisa. |
| [internal/backlog/sello.go](../../internal/backlog/sello.go) | `gitLineas` | `(raiz string, args ...string) ([]string, error)` | gitLineas corre git en la raíz y devuelve su salida línea por línea. |
| [internal/backlog/sello.go](../../internal/backlog/sello.go) | `pagina` | `(id string) string` | pagina: ruta de la página wiki de una ficha según su nivel en el ID. |
| [internal/backlog/sello.go](../../internal/backlog/sello.go) | `arbol` | `(raiz string, f Ficha) ([]Item, error)` | arbol: padre, abuelo, dependientes abiertos y páginas. |
| [internal/backlog/sello.go](../../internal/backlog/sello.go) | `moduloGo` | `(raiz string) string` | moduloGo: ruta del módulo declarada en go.mod, o vacío si no hay. |
| [internal/backlog/sello.go](../../internal/backlog/sello.go) | `lineaCon` | `(c, s string) string` | lineaCon: primera línea de c que contiene s, recortada; sirve de evidencia en el sello. |
| [internal/backlog/sello.go](../../internal/backlog/sello.go) | `derivados` | `(raiz string) map[string]bool` | derivados: salidas del compilador según su archivo de sellos, más ese archivo (S014). Ya los vigila |
| [internal/backlog/sello.go](../../internal/backlog/sello.go) | `archivos` | `(raiz, head string) ([]Item, error)` | archivos: quién nombra o importa lo tocado desde el HEAD del rojo. |
| [internal/backlog/sello.go](../../internal/backlog/sello.go) | `Sello` | `(raiz, id string) ([]Item, error)` | Sello arma la lista del slice y le pega las marcas guardadas en su ficha. |
| [internal/backlog/sello.go](../../internal/backlog/sello.go) | `sello` | `(raiz string, f Ficha) ([]Item, error)` | sello arma la lista de piezas que el cambio de un slice obliga a revisar. |
| [internal/backlog/sello.go](../../internal/backlog/sello.go) | `selloCompleto` | `(raiz string, f Ficha) error` | selloCompleto: CERRADO exige marca en cada ítem, haya o no commit base (S013). |
| [internal/backlog/sello.go](../../internal/backlog/sello.go) | `Marcar` | `(raiz, id, clave, marca, motivo string) (Ficha, error)` | Marcar guarda la revisión de un ítem del sello de un slice activo. |
| [internal/backlog/sello.go](../../internal/backlog/sello.go) | `TextoSello` | `(xs []Item) string` | TextoSello: una línea por ítem; [ ] pendiente, [x] marcado con su marca y motivo. |
| [internal/biblioteca/biblioteca.go](../../internal/biblioteca/biblioteca.go) | `Leer` | `(raiz string, p Pedido) (texto string, devueltos, completos int)` | Leer devuelve el texto para el agente y cuántos caracteres devolvió de cuántos tiene |
| [internal/calidad/calidad.go](../../internal/calidad/calidad.go) | `Revisar` | `(dir, config string) ([]string, error)` | Revisar corre golangci-lint con config sobre dir y devuelve un hallazgo por línea. |
| [internal/calidad/calidad.go](../../internal/calidad/calidad.go) | `chequearVersion` | `(salida string) error` | chequearVersion falla si la salida de `golangci-lint version` no es la versión fija. |
| [internal/calidad/calidad.go](../../internal/calidad/calidad.go) | `RevisarCentral` | `(raiz string) []string` | RevisarCentral corre la calidad Go de raiz con la config del gate; sin go.mod no hay código que revisar. |
| [internal/calidad/dependencias.go](../../internal/calidad/dependencias.go) | `RevisarDependencias` | `(dir, tabla string) ([]string, error)` | RevisarDependencias (E001/F013/S002) compara los imports internos de dir, de código y de prueba, |
| [internal/calidad/dependencias.go](../../internal/calidad/dependencias.go) | `rel` | `(imp, mod string) string` | rel devuelve la ruta de import relativa al módulo ("." para la raíz). |
| [internal/compilador/compilador.go](../../internal/compilador/compilador.go) | `marca` | `(config string) string` | marca: comentario «GENERADO» que encabeza cada derivado y nombra su config. |
| [internal/compilador/compilador.go](../../internal/compilador/compilador.go) | `Planificar` | `(raiz string) ([]Salida, error)` | Planificar lee la config y calcula todas las salidas sin tocar el disco. |
| [internal/compilador/compilador.go](../../internal/compilador/compilador.go) | `compilarAgente` | `(raiz string, a Agente) (Salida, error)` | compilarAgente arma el .md de un agente desde su config y sus canónicos, y valida campos y tope. |
| [internal/compilador/compilador.go](../../internal/compilador/compilador.go) | `Compilar` | `(raiz string) (Informe, error)` | Compilar escribe las salidas que cambiaron y actualiza el sello. |
| [internal/compilador/compilador.go](../../internal/compilador/compilador.go) | `Verificar` | `(raiz string) ([]string, error)` | Verificar es el gate: devuelve un problema por cada salida o sello desfasado. |
| [internal/compilador/compilador.go](../../internal/compilador/compilador.go) | `EsGobernado` | `(rel string) bool` | EsGobernado indica si una ruta relativa (con /) afecta la compilación. |
| [internal/compilador/compilador.go](../../internal/compilador/compilador.go) | `EsSalida` | `(raiz, rel string) bool` | EsSalida indica si una ruta relativa es un archivo compilado. |
| [internal/compilador/compilador.go](../../internal/compilador/compilador.go) | `sellar` | `(raiz string, salidas []Salida) (Sello, error)` | sellar calcula la huella SHA-256 de cada salida y de sus fuentes. |
| [internal/compilador/compilador.go](../../internal/compilador/compilador.go) | `diferencias` | `(anterior, nuevo Sello) []string` | diferencias: rutas cuya huella cambió o que aparecen o desaparecen entre dos sellos. |
| [internal/compilador/compilador.go](../../internal/compilador/compilador.go) | `leerCanonicos` | `(raiz string, rutas []string) ([]string, error)` | leerCanonicos lee las fuentes pedidas; rechaza las que no están bajo gobierno/canonico/. |
| [internal/compilador/compilador.go](../../internal/compilador/compilador.go) | `leerJSON` | `(raiz, rel string, v any) error` | leerJSON lee y decodifica un JSON relativo a la raíz. |
| [internal/compilador/compilador.go](../../internal/compilador/compilador.go) | `leerSello` | `(raiz string) (Sello, error)` | leerSello carga gobierno/config/sellos.json. |
| [internal/compilador/compilador.go](../../internal/compilador/compilador.go) | `codificarSello` | `(s Sello) ([]byte, error)` | codificarSello serializa el sello como JSON indentado con salto final. |
| [internal/compilador/compilador.go](../../internal/compilador/compilador.go) | `citarYAML` | `(s string) string` | citarYAML produce un escalar YAML entre comillas dobles (JSON es YAML válido). |
| [internal/compilador/compilador.go](../../internal/compilador/compilador.go) | `normalizar` | `(b []byte) []byte` | normalizar cambia CRLF por LF para que la huella no dependa del sistema. |
| [internal/compilador/compilador.go](../../internal/compilador/compilador.go) | `huella` | `(b []byte) string` | huella: SHA-256 en hexadecimal. |
| [internal/compilador/compilador.go](../../internal/compilador/compilador.go) | `escribir` | `(raiz, rel string, datos []byte) error` | escribir guarda datos en la ruta relativa, creando sus carpetas. |
| [internal/defecto/defecto.go](../../internal/defecto/defecto.go) | `Nuevo` | `(raiz string, c map[string]string) (Defecto, error)` | Nuevo crea el siguiente D000NNN. Todos los campos son obligatorios; la firma es kebab-case y única. |
| [internal/defecto/defecto.go](../../internal/defecto/defecto.go) | `Ocurre` | `(raiz, id, turno string) (Defecto, error)` | Ocurre suma un turno T000NNN al defecto; el mismo turno no se cuenta dos veces. |
| [internal/defecto/defecto.go](../../internal/defecto/defecto.go) | `Leer` | `(raiz string) ([]Defecto, error)` | Leer devuelve todos los defectos, ordenados por ID. Sin carpeta = registro vacío. |
| [internal/defecto/defecto.go](../../internal/defecto/defecto.go) | `leerUno` | `(raiz, id string) (Defecto, error)` | leerUno carga el defecto con ese ID; valida el formato D000001. |
| [internal/defecto/defecto.go](../../internal/defecto/defecto.go) | `escribir` | `(raiz string, d Defecto) error` | escribir guarda el defecto como JSON en su carpeta. |
| [internal/defecto/defecto.go](../../internal/defecto/defecto.go) | `Texto` | `(d Defecto) string` | Texto: una línea por defecto para `aquas defecto ver`. |
| [internal/defecto/defecto.go](../../internal/defecto/defecto.go) | `revisarRelacion` | `(raiz, rel, que string, todos []Defecto) error` | revisarRelacion (E001/F009/S002): relacion=nuevo o el id de una ficha; sin ella, lista defectos y fichas parecidos. |
| [internal/encargo/encargo.go](../../internal/encargo/encargo.go) | `Revisar` | `(tipo, mensaje string, stopHookActive bool) string` | Revisar decide el hook SubagentStop: "" deja cerrar; si no, qué falta en el cierre. |
| [internal/encargo/encargo.go](../../internal/encargo/encargo.go) | `Veredicto` | `(mensaje string) (id, veredicto string, ok bool)` | Veredicto extrae del cierre de un auditor ya válido el id del tramo («E001/F014/S002») y su veredicto. |
| [internal/encargo/encargo.go](../../internal/encargo/encargo.go) | `falla` | `(motivo string) string` | falla: mensaje de cierre inválido que pide reescribirlo según el canónico. |
| [internal/entorno/binario.go](../../internal/entorno/binario.go) | `BinarioViejo` | `(raiz, exe string) string` | BinarioViejo (D000026): aviso si algún .go de producción es más nuevo que el binario exe; "" si está al día. |
| [internal/entorno/entorno.go](../../internal/entorno/entorno.go) | `Leer` | `(raiz string) (Entorno, error)` | Leer carga el entorno declarado. |
| [internal/entorno/entorno.go](../../internal/entorno/entorno.go) | `Validar` | `(raiz string) string` | Validar compara el entorno declarado con el real; "" = coincide, si no, el motivo del bloqueo. |
| [internal/entorno/entorno.go](../../internal/entorno/entorno.go) | `Iniciar` | `(raiz string) (Entorno, error)` | Iniciar detecta SO y raíz (git) y escribe entorno.local.json; exige correr desde la raíz. |
| [internal/entorno/entorno.go](../../internal/entorno/entorno.go) | `mismaRuta` | `(a, b string) bool` | mismaRuta compara dos rutas limpias; en Windows sin distinguir mayúsculas. |
| [internal/entorno/entorno.go](../../internal/entorno/entorno.go) | `generarSettings` | `(fuente []byte, so string) ([]byte, error)` | generarSettings arma settings.json para un SO desde la plantilla canónica. Bytes estables. |
| [internal/entorno/entorno.go](../../internal/entorno/entorno.go) | `Settings` | `(raiz string) ([]byte, error)` | Settings genera el settings.json esperado para el entorno declarado. |
| [internal/entorno/entorno.go](../../internal/entorno/entorno.go) | `CompilarSettings` | `(raiz string) (bool, error)` | CompilarSettings escribe .claude/settings.json; devuelve true si cambió. |
| [internal/entorno/entorno.go](../../internal/entorno/entorno.go) | `VerificarSettings` | `(raiz string) string` | VerificarSettings: "" si .claude/settings.json coincide con lo regenerado. |
| [internal/guardia/guardia.go](../../internal/guardia/guardia.go) | `Revisar` | `(raiz string, entrada []byte) string` | Revisar recibe el JSON del hook PreToolUse; "" = permitido, si no, el motivo del bloqueo. |
| [internal/guardia/guardia.go](../../internal/guardia/guardia.go) | `RevisarArchivo` | `(raiz, p string) string` | RevisarArchivo valida una ruta de Write/Edit (absoluta o relativa a la raíz). |
| [internal/guardia/guardia.go](../../internal/guardia/guardia.go) | `protegida` | `(raiz, abs, p string) string` | protegida: carpetas que solo escribe `aquas` (E001/F005, E001/F002). Las rutas salen de los |
| [internal/guardia/guardia.go](../../internal/guardia/guardia.go) | `resolver` | `(base, p string) (string, bool)` | resolver lleva p a ruta absoluta del sistema. Entiende ~, /c/… (Git Bash) y relativas a base. |
| [internal/guardia/guardia.go](../../internal/guardia/guardia.go) | `dentro` | `(raiz, abs string) bool` | dentro: la ruta absoluta queda dentro de la raíz del repo. |
| [internal/guardia/guardia.go](../../internal/guardia/guardia.go) | `sinHeredocs` | `(c string) string` | sinHeredocs quita el cuerpo de cada heredoc: su contenido es texto, no comandos. |
| [internal/guardia/guardia.go](../../internal/guardia/guardia.go) | `lexear` | `(c string) []palabra` | lexear parte un comando de shell en palabras, respetando comillas y operadores. |
| [internal/guardia/guardia.go](../../internal/guardia/guardia.go) | `variable` | `(r []rune, i int) (string, int)` | variable reconoce $NOMBRE, ${NOMBRE}, $(…), `…` y %NOMBRE%; devuelve nombre ("?" = desconocida) y largo. |
| [internal/guardia/guardia.go](../../internal/guardia/guardia.go) | `esNombre` | `(c rune) bool` | esNombre: el carácter vale en un nombre de variable de shell. |
| [internal/guardia/guardia.go](../../internal/guardia/guardia.go) | `RevisarBash` | `(raiz, cwd, comando string) string` | RevisarBash analiza un comando de Bash; cwd es el directorio donde arranca. |
| [internal/guardia/guardia.go](../../internal/guardia/guardia.go) | `segmento` | `(raiz string, cwd *string, cwdOK *bool, vars map[string]string, ps []palabra) string` | segmento revisa un comando simple: sigue cd y variables y devuelve el primer destino fuera del repo. |
| [internal/guardia/guardia.go](../../internal/guardia/guardia.go) | `posicionales` | `(args []string) []string` | posicionales: argumentos que no son opciones (los que empiezan con «-», salvo «-» solo). |
| [internal/guardia/guardia.go](../../internal/guardia/guardia.go) | `revisarDestino` | `(raiz, cwd string, cwdOK bool, vars map[string]string, d string) string` | revisarDestino: motivo de bloqueo si la ruta no se resuelve o cae fuera del repo; vacío si vale. |
| [internal/guardia/guardia.go](../../internal/guardia/guardia.go) | `expandir` | `(s string, vars map[string]string, cwd string) (string, bool)` | expandir sustituye variables conocidas; HOME/USERPROFILE/TEMP/TMP salen del entorno real. |
| [internal/guardia/guardia.go](../../internal/guardia/guardia.go) | `legible` | `(s string) string` | legible muestra las variables internas como $NOMBRE en los mensajes. |
| [internal/guardia/permitidos.go](../../internal/guardia/permitidos.go) | `cargar` | `(raiz string) ([]string, []string)` | cargar devuelve las carpetas permitidas ya resueltas y los problemas de la config. |
| [internal/guardia/permitidos.go](../../internal/guardia/permitidos.go) | `expandirConfig` | `(s string) (string, bool)` | expandirConfig sustituye variables de entorno ${X} de la config; ok falso si falta alguna. |
| [internal/guardia/permitidos.go](../../internal/guardia/permitidos.go) | `permitido` | `(raiz, abs string) bool` | permitido: abs cae dentro de alguna carpeta de la lista. |
| [internal/guardia/permitidos.go](../../internal/guardia/permitidos.go) | `ValidarConfig` | `(raiz string) []string` | ValidarConfig es el gate de la lista de permitidos (lo llama `aquas verificar`). |
| [internal/guardia/permitidos.go](../../internal/guardia/permitidos.go) | `itoa` | `(i int) string` | itoa convierte un entero a texto. |
| [internal/hijo/ajustes.go](../../internal/hijo/ajustes.go) | `AsegurarHook` | `(dir string) error` | AsegurarHook (E002/F002/S006) agrega a .claude/settings.json del hijo un hook PostToolUse |
| [internal/hijo/aviso.go](../../internal/hijo/aviso.go) | `AvisoVersion` | `(dir, binario string) string` | AvisoVersion (E002/F002/S003) devuelve el aviso de versión del hijo, o "" si no hay nada que avisar. |
| [internal/hijo/build.go](../../internal/hijo/build.go) | `Build` | `(dir string) error` | Build arma el CLAUDE.md del hijo: central sellada + huecos locales. |
| [internal/hijo/build.go](../../internal/hijo/build.go) | `armar` | `(dir string) (string, error)` | armar devuelve el CLAUDE.md que Build escribiría, sin escribirlo. |
| [internal/hijo/build.go](../../internal/hijo/build.go) | `verificarSello` | `(dir, central string) error` | verificarSello compara cada archivo de la central con su hash del lock. |
| [internal/hijo/build.go](../../internal/hijo/build.go) | `leerHuecos` | `(dir string) (map[string]string, error)` | leerHuecos devuelve nombre → contenido de cada <nombre>.md local. |
| [internal/hijo/build.go](../../internal/hijo/build.go) | `Verificar` | `(dir string) []string` | Verificar (X02, E002/F002/S004): gate del hijo; nil si su central sellada, sus huecos sin PENDIENTE y su CLAUDE.md están al día. |
| [internal/hijo/build.go](../../internal/hijo/build.go) | `EsHijo` | `(dir string) bool` | EsHijo: dir tiene un gobierno central instalado (lock.json de aquas install). |
| [internal/hijo/build.go](../../internal/hijo/build.go) | `pendientes` | `(dir string) ([]string, error)` | pendientes (E002/F002/S004): un problema por hueco que conserva la marca de relleno, en orden. |
| [internal/hijo/hijo.go](../../internal/hijo/hijo.go) | `Install` | `(dest string, central fs.FS, version string) error` | Install copia la central sellada en dest/.aquas/central y escribe dest/.aquas/lock.json. |
| [internal/hijo/hijo.go](../../internal/hijo/hijo.go) | `copiarCentral` | `(dest string, central fs.FS, version string) error` | copiarCentral escribe la central como solo lectura y su lock con versión y hashes. |
| [internal/hijo/hijo.go](../../internal/hijo/hijo.go) | `AMedias` | `(err error) error` | AMedias (E002/F002/S002): error de install cuando la central ya quedó copiada y falló lo siguiente. |
| [internal/hijo/huecos.go](../../internal/hijo/huecos.go) | `PrepararHuecos` | `(dir string) error` | PrepararHuecos (E002/F002/S001): crea con texto PENDIENTE los huecos de la central |
| [internal/hijo/huecos.go](../../internal/hijo/huecos.go) | `Listo` | `(dir string) error` | Listo deja al hijo con sus huecos, su hook de build (S006) y su CLAUDE.md armado. |
| [internal/hijo/report.go](../../internal/hijo/report.go) | `Report` | `(dir, quePaso, repro string) (string, error)` | Report arma el borrador de issue para la central: firma del hijo, versión |
| [internal/hijo/update.go](../../internal/hijo/update.go) | `Update` | `(dest string, central fs.FS, version string) ([]string, error)` | Update reemplaza la central sellada por otra versión y reescribe el lock. |
| [internal/hijo/update.go](../../internal/hijo/update.go) | `alterados` | `(dest string) ([]string, error)` | alterados lista los archivos sellados que ya no coinciden con el lock. |
| [internal/memoria/archivar.go](../../internal/memoria/archivar.go) | `RutaArchivo` | `(f Foco) string` | RutaArchivo: registro/memoria/<E>-<F>-<S>.json; "" si el foco no llega a slice. |
| [internal/memoria/archivar.go](../../internal/memoria/archivar.go) | `Archivar` | `(raiz string, ahora time.Time) (string, error)` | Archivar copia los bytes sellados de la memoria a registro/memoria/<E>-<F>-<S>.json y deja |
| [internal/memoria/archivar.go](../../internal/memoria/archivar.go) | `LeerArchivo` | `(ruta string) (Memoria, error)` | LeerArchivo carga una memoria archivada y comprueba su sello. |
| [internal/memoria/memoria.go](../../internal/memoria/memoria.go) | `Fijar` | `(raiz string, pares []string, ahora time.Time) (Memoria, error)` | Fijar aplica pares campo=valor sobre la memoria actual y la sella. |
| [internal/memoria/memoria.go](../../internal/memoria/memoria.go) | `FijarCon` | `(raiz string, pares []string, ahora time.Time, validar func(Foco) error) (Memoria, error)` | FijarCon: como Fijar, pero validar (si no es nil) revisa el foco resultante antes de escribir. |
| [internal/memoria/memoria.go](../../internal/memoria/memoria.go) | `Leer` | `(raiz string) (Memoria, error)` | Leer carga la memoria y comprueba su sello. |
| [internal/memoria/memoria.go](../../internal/memoria/memoria.go) | `decodificar` | `(b []byte) (Memoria, error)` | decodificar: campos fijos y sello sobre los bytes tal como se escribieron. |
| [internal/memoria/memoria.go](../../internal/memoria/memoria.go) | `Rutas` | `(f Foco) Mapa` | Rutas: páginas wiki de la iniciativa, feature y slice del foco, más el backlog. |
| [internal/memoria/memoria.go](../../internal/memoria/memoria.go) | `Texto` | `(m Memoria) string` | Texto: lo que imprime `aquas memoria ver` (y, en S002, el hook de inicio de sesión). |
| [internal/memoria/memoria.go](../../internal/memoria/memoria.go) | `guion` | `(s string) string` | guion devuelve «—» si el texto está vacío. |
| [internal/memoria/memoria.go](../../internal/memoria/memoria.go) | `lista` | `(b *strings.Builder, titulo string, xs []string)` | lista escribe un título y sus viñetas, o «—» si no hay ninguna. |
| [internal/memoria/memoria.go](../../internal/memoria/memoria.go) | `codificar` | `(m Memoria) []byte` | codificar serializa la memoria como JSON indentado con salto final. |
| [internal/memoria/memoria.go](../../internal/memoria/memoria.go) | `huella` | `(m Memoria) string` | huella: SHA-256 de la memoria codificada; detecta cambios. |
| [internal/memoria/memoria.go](../../internal/memoria/memoria.go) | `escribir` | `(raiz string, m Memoria) error` | escribir: archivo temporal + rename, para no dejar una memoria a medias. |
| [internal/memoria/memoria.go](../../internal/memoria/memoria.go) | `HuellaGit` | `(raiz string) string` | HuellaGit resume el estado del repo: HEAD + cada ruta de `git status` con el contenido de su |
| [internal/memoria/memoria.go](../../internal/memoria/memoria.go) | `mismaCarpeta` | `(a, b string) bool` | mismaCarpeta: las dos rutas apuntan a la misma carpeta en disco. |
| [internal/memoria/memoria.go](../../internal/memoria/memoria.go) | `Cierre` | `(raiz string, stopHookActive bool) string` | Cierre decide el hook Stop: "" deja cerrar el turno; si no, el motivo para seguir. |
| [internal/memoria/memoria.go](../../internal/memoria/memoria.go) | `Inicio` | `(raiz string) string` | Inicio: texto que el hook SessionStart agrega al contexto. Nunca bloquea. |
| [internal/propuesta/insight.go](../../internal/propuesta/insight.go) | `RevisarInsight` | `(msg string) string` | RevisarInsight devuelve "" si pasa; si no, el motivo del freno: insight PROPUESTO sin defecto registrado (E001/F002/S016). |
| [internal/propuesta/propuesta.go](../../internal/propuesta/propuesta.go) | `marca` | `(l string) (string, bool)` | marca: texto de la propuesta si la línea empieza con PROPUESTA: (admite viñeta, título, cita o negrita). |
| [internal/propuesta/propuesta.go](../../internal/propuesta/propuesta.go) | `diagrama` | `(c string) bool` | diagrama: el texto trae algún trazo de diagrama (flechas o cajas). |
| [internal/propuesta/propuesta.go](../../internal/propuesta/propuesta.go) | `cercaDe` | `(t string) (byte, int)` | cercaDe: el carácter y el largo de una cerca de bloque (``` o ~~~), o largo 0 si no es cerca. |
| [internal/propuesta/propuesta.go](../../internal/propuesta/propuesta.go) | `Revisar` | `(msg string) string` | Revisar devuelve "" si pasa; si no, el motivo del freno con las propuestas sin diagrama. |
| [internal/propuesta/shell.go](../../internal/propuesta/shell.go) | `RevisarShell` | `(msg, shell string) string` | RevisarShell (E001/F016/S001): "" si ningún bloque de código marcado es de un shell ajeno al declarado; |
| [internal/turno/turno.go](../../internal/turno/turno.go) | `siguiente` | `(raiz string) (int, error)` | siguiente: el mayor número existente + 1; un hueco no se rellena (los ID no se reutilizan). |
| [internal/turno/turno.go](../../internal/turno/turno.go) | `leerEstado` | `(raiz string) (map[string]abierto, error)` | leerEstado carga los turnos abiertos; vacío si el archivo no existe. |
| [internal/turno/turno.go](../../internal/turno/turno.go) | `escribir` | `(raiz, rel string, b []byte) error` | escribir guarda bytes en la ruta relativa, creando sus carpetas. |
| [internal/turno/turno.go](../../internal/turno/turno.go) | `cerca` | `(s string) string` | cerca: bloque de código cuya cerca es más larga que cualquier racha de backticks del texto, |
| [internal/turno/turno.go](../../internal/turno/turno.go) | `render` | `(a abierto, respuesta, cerrado string) []byte` | render arma el Markdown del registro de un turno: literal del humano y respuesta. |
| [internal/turno/turno.go](../../internal/turno/turno.go) | `Abrir` | `(raiz, sesion, literal string, ahora time.Time) (string, error)` | Abrir crea el turno siguiente con el literal y lo deja como el abierto de la sesión. |
| [internal/turno/turno.go](../../internal/turno/turno.go) | `Cerrar` | `(raiz, sesion, respuesta string, ahora time.Time) (string, error)` | Cerrar guarda la respuesta en el último turno de la sesión; sin turno abierto, no escribe. |
| [internal/vueltas/vueltas.go](../../internal/vueltas/vueltas.go) | `motivos` | `(tipo, salida string) []string` | motivos de una salida para un tipo; vacío = sin señal propia de fallo. |
| [internal/vueltas/vueltas.go](../../internal/vueltas/vueltas.go) | `regla` | `(txt string) string` | regla: el mensaje del gate sin lo que cambia de página a página (citas, paréntesis, números). |
| [internal/vueltas/vueltas.go](../../internal/vueltas/vueltas.go) | `Contar` | `(r io.Reader) (map[string]Conteo, error)` | Contar lee un historial .jsonl y agrupa por tipo ("go test", "verificar"). |
| [internal/vueltas/vueltas.go](../../internal/vueltas/vueltas.go) | `texto` | `(raw json.RawMessage) string` | texto: el contenido de un resultado viene como cadena o como lista de bloques {type,text}. |
| [internal/vueltas/vueltas.go](../../internal/vueltas/vueltas.go) | `Carpeta` | `(home, raiz string) string` | Carpeta del historial de Claude Code para este repo: ~/.claude/projects/<raíz con :\/ y espacios → ->. |
| [internal/vueltas/vueltas.go](../../internal/vueltas/vueltas.go) | `Medir` | `(w io.Writer, home, raiz string, archivos []string) error` | Medir imprime una línea por historial y el total. Sin archivos: todos los .jsonl de Carpeta. |
| [internal/vueltas/vueltas.go](../../internal/vueltas/vueltas.go) | `Linea` | `(nombre string, c map[string]Conteo) string` | Linea: «nombre · go test 92/47 (TestA 3, compilación 2) · verificar 102/37 (wiki 20)»; motivos de más a menos. |
| [internal/wiki/casos_borde.go](../../internal/wiki/casos_borde.go) | `revisarCasosBorde` | `(raiz, rel string, enCodigo map[string]string) []string` | revisarCasosBorde: problemas de la tabla «Casos borde» de la página rel; nada si no es un slice con ficha activa. |
| [internal/wiki/casos_borde.go](../../internal/wiki/casos_borde.go) | `filasCasosBorde` | `(contenido string) (string, [][3]string, bool)` | filasCasosBorde: encabezado y celdas [categoría, aplica, texto] de la tabla que sigue a la marca; false si no hay marca. |
| [internal/wiki/casos_borde.go](../../internal/wiki/casos_borde.go) | `celdasTabla` | `(linea string) []string` | celdasTabla parte una fila Markdown por «\\|» sin escapar y recorta cada celda. |
| [internal/wiki/casos_borde.go](../../internal/wiki/casos_borde.go) | `lineaDeMarca` | `(lineas []string) int` | lineaDeMarca: índice de la marca fuera de bloques de código y comentarios HTML; -1 si no está. |
| [internal/wiki/codigo.go](../../internal/wiki/codigo.go) | `funcionesGo` | `(raiz string) ([]funcion, error)` | funcionesGo recorre cmd/ e internal/ (sin _test.go ni testdata) y devuelve sus funciones en orden de archivo y línea. |
| [internal/wiki/codigo.go](../../internal/wiki/codigo.go) | `revisarCodigo` | `(raiz string) ([]string, error)` | revisarCodigo: avisos por cada función sin comentario. |
| [internal/wiki/codigo.go](../../internal/wiki/codigo.go) | `indiceCodigo` | `(raiz string) (string, error)` | indiceCodigo: tabla archivo → función → parámetros → primera línea de su comentario. |
| [internal/wiki/codigo.go](../../internal/wiki/codigo.go) | `tablaCodigo` | `(raiz string, f backlog.Ficha) (string, error)` | tablaCodigo: sección «Código» de un slice: cada archivo tocado con sus funciones, parámetros y qué hacen. |
| [internal/wiki/comandos.go](../../internal/wiki/comandos.go) | `RevisarComandos` | `(raiz string) ([]string, error)` | RevisarComandos: todo `aquas X [Y]` citado entre comillas invertidas en docs/ y en |
| [internal/wiki/contrato.go](../../internal/wiki/contrato.go) | `seccionesDe` | `(n nivel) []string` | seccionesDe: orden fijo de secciones según el nivel de la página. |
| [internal/wiki/contrato.go](../../internal/wiki/contrato.go) | `migasEsperadas` | `(pg *pagina) string` | migasEsperadas: migas «Inicio › E001 › F001», cada nivel con enlace relativo a la página. |
| [internal/wiki/contrato.go](../../internal/wiki/contrato.go) | `revisarMigas` | `(pg *pagina) []string` | revisarMigas: aviso si las migas de pan no coinciden con la ruta de la página. |
| [internal/wiki/contrato.go](../../internal/wiki/contrato.go) | `revisarTrazabilidad` | `(raiz string, paginas map[string]*pagina, orden []string) ([]string, error)` | revisarTrazabilidad cruza las pruebas nombradas en los slices con las func Test del repo. |
| [internal/wiki/contrato.go](../../internal/wiki/contrato.go) | `prevista` | `(raiz, rel, n string) bool` | prevista (E001/F010/S001): la prueba n aún no existe en el código pero la ficha del slice de la página rel |
| [internal/wiki/contrato.go](../../internal/wiki/contrato.go) | `integralesDe` | `(raiz, rel string) []string` | integralesDe: nombres de las func Test que declara la ficha de la feature o iniciativa de la página rel; ninguno si no hay ficha. |
| [internal/wiki/contrato.go](../../internal/wiki/contrato.go) | `pruebasDelRepo` | `(raiz string) (map[string]string, error)` | pruebasDelRepo: nombre de cada func Test → archivo; salta carpetas ocultas (.git, .tmp) y testdata. |
| [internal/wiki/enlazar.go](../../internal/wiki/enlazar.go) | `Enlazar` | `(raiz, rel string) (int, error)` | Enlazar pone el enlace al glosario en la primera aparición en prosa de cada término (E000/F004/S003). |
| [internal/wiki/enlazar.go](../../internal/wiki/enlazar.go) | `envolver` | `(raiz, rel string, f falta) (bool, error)` | envolver enlaza la primera aparición en prosa del término en el archivo. |
| [internal/wiki/glosario.go](../../internal/wiki/glosario.go) | `terminos` | `(pg *pagina) []termino` | terminos extrae las entradas del glosario con sus formas (nombre + variantes). |
| [internal/wiki/glosario.go](../../internal/wiki/glosario.go) | `revisarGlosario` | `(pg *pagina) []string` | revisarGlosario: cada término del glosario lleva definición y ejemplo. |
| [internal/wiki/glosario.go](../../internal/wiki/glosario.go) | `revisarTerminos` | `(pg *pagina, ts []termino) []string` | revisarTerminos exige que la primera aparición en prosa de cada término enlace a su entrada. |
| [internal/wiki/glosario.go](../../internal/wiki/glosario.go) | `faltantes` | `(pg *pagina, ts []termino) []falta` | faltantes: la detección que comparten el gate (revisarTerminos) y el arreglo (Enlazar). |
| [internal/wiki/glosario.go](../../internal/wiki/glosario.go) | `relDesde` | `(desde, destino string) string` | relDesde da la ruta relativa de destino vista desde la carpeta de la página. |
| [internal/wiki/hijos.go](../../internal/wiki/hijos.go) | `paginaDe` | `(id string) string` | paginaDe: la página de la wiki que corresponde a un id de ficha. |
| [internal/wiki/hijos.go](../../internal/wiki/hijos.go) | `celda` | `(s string) string` | celda escapa saltos de línea y «\\|» para una celda de tabla Markdown. |
| [internal/wiki/hijos.go](../../internal/wiki/hijos.go) | `existe` | `(raiz, rel string) bool` | existe: la ruta relativa existe en disco. |
| [internal/wiki/hijos.go](../../internal/wiki/hijos.go) | `tablaHijos` | `(raiz, padreRel string, hijos []backlog.Ficha) string` | tablaHijos: una fila por hijo directo; enlace solo si su página existe. |
| [internal/wiki/hijos.go](../../internal/wiki/hijos.go) | `fecha` | `(r *backlog.Registro) string` | fecha: día AAAA-MM-DD de un registro, o «—» si no hay. |
| [internal/wiki/hijos.go](../../internal/wiki/hijos.go) | `oGuion` | `(s string) string` | oGuion: el texto como celda, o «—» si está vacío (pruebas anteriores a E001/F008/S001). |
| [internal/wiki/hijos.go](../../internal/wiki/hijos.go) | `tablaPruebas` | `(f backlog.Ficha) string` | tablaPruebas: las pruebas del slice: qué validan, cuándo quedan en rojo y en verde, y si se vieron en rojo y en verde. |
| [internal/wiki/hijos.go](../../internal/wiki/hijos.go) | `esperados` | `(raiz string) (map[string]esperado, []backlog.Ficha, error)` | esperados: página existente → bloque que debe llevar. Vacío si no hay fichas. |
| [internal/wiki/hijos.go](../../internal/wiki/hijos.go) | `bloqueActual` | `(c string) (ini, fin int)` | bloqueActual: posición del bloque en c; ini < 0 = no hay; fin < 0 = sin cierre. |
| [internal/wiki/hijos.go](../../internal/wiki/hijos.go) | `Generar` | `(raiz string) ([]string, error)` | Generar reescribe el bloque GENERADO de cada página desde las fichas; devuelve lo escrito. |
| [internal/wiki/hijos.go](../../internal/wiki/hijos.go) | `revisarHijos` | `(raiz string, paginas map[string]*pagina, orden []string) ([]string, error)` | revisarHijos: bloque presente y al día, enlaces sin saltar niveles, ficha arrancada con página. |
| [internal/wiki/hijos.go](../../internal/wiki/hijos.go) | `tablaArreglos` | `(f backlog.Ficha) string` | tablaArreglos: sección «Arreglos» de la ficha; vacía si no tiene. |
| [internal/wiki/legibilidad.go](../../internal/wiki/legibilidad.go) | `Silabas` | `(palabra string) int` | Silabas cuenta las sílabas de una palabra en español por núcleos vocálicos, |
| [internal/wiki/legibilidad.go](../../internal/wiki/legibilidad.go) | `esVocalLetra` | `(r rune) bool` | esVocalLetra: la runa es vocal castellana, con o sin tilde. |
| [internal/wiki/legibilidad.go](../../internal/wiki/legibilidad.go) | `hiato` | `(a, b rune) bool` | hiato: las dos vocales seguidas se cuentan como sílabas separadas. |
| [internal/wiki/legibilidad.go](../../internal/wiki/legibilidad.go) | `oraciones` | `(lineas []string) [][]string` | oraciones parte la prosa en oraciones de palabras. Línea vacía o ítem de lista |
| [internal/wiki/legibilidad.go](../../internal/wiki/legibilidad.go) | `Perspicuidad` | `(ors [][]string) float64` | Perspicuidad calcula el índice Szigriszt-Pazos; 0 si no hay palabras. |
| [internal/wiki/legibilidad.go](../../internal/wiki/legibilidad.go) | `revisarLegibilidad` | `(pg *pagina) []string` | revisarLegibilidad: avisos por oraciones largas o prosa difícil de leer. |
| [internal/wiki/legibilidad.go](../../internal/wiki/legibilidad.go) | `coma` | `(x float64) string` | coma formatea un número con un decimal y coma decimal. |
| [internal/wiki/mermaid.go](../../internal/wiki/mermaid.go) | `revisarMermaid` | `(rel, c string) []string` | revisarMermaid: avisa cada «;» dentro de un bloque mermaid; Mermaid lo toma como fin de sentencia (D000021). |
| [internal/wiki/portadas.go](../../internal/wiki/portadas.go) | `faltanIniciativas` | `(ids []string, readme, inicioWiki string) []string` | faltanIniciativas: problemas del gate si alguna iniciativa no está enlazada en una portada. |
| [internal/wiki/portadas.go](../../internal/wiki/portadas.go) | `revisarPortadas` | `(raiz string) ([]string, error)` | revisarPortadas: lee las iniciativas del backlog y las dos portadas; solo exige las iniciativas que ya tienen página. |
| [internal/wiki/wiki.go](../../internal/wiki/wiki.go) | `Verificar` | `(raiz string) ([]string, error)` | Verificar devuelve los problemas de la wiki bajo raiz/docs; vacío = gate OK. |
| [internal/wiki/wiki.go](../../internal/wiki/wiki.go) | `relativa` | `(raiz, p string) string` | relativa: ruta de p respecto de la raíz, con barras «/». |
| [internal/wiki/wiki.go](../../internal/wiki/wiki.go) | `clasificar` | `(rel string) *pagina` | clasificar asigna nivel, padre e id de ejemplo según la ruta; nil = fuera del árbol. |
| [internal/wiki/wiki.go](../../internal/wiki/wiki.go) | `analizar` | `(raiz string, pg *pagina) error` | analizar lee la página: secciones H2, diagramas, anclas (id HTML y títulos) y enlaces fuera de código. |
| [internal/wiki/wiki.go](../../internal/wiki/wiki.go) | `slug` | `(s string) string` | slug al estilo GitHub/VS Code: minúsculas, letras/dígitos/-/_ se conservan, espacio → guion. |
| [internal/wiki/wiki.go](../../internal/wiki/wiki.go) | `pagina.seccion` | `(titulo string) *seccion` | seccion busca una sección de la página por su título; nil si no está. |
| [internal/wiki/wiki.go](../../internal/wiki/wiki.go) | `revisarSecciones` | `(pg *pagina) []string` | revisarSecciones: faltan secciones fijas o están fuera de orden. |
| [internal/wiki/wiki.go](../../internal/wiki/wiki.go) | `revisarDiagramas` | `(pg *pagina) []string` | revisarDiagramas: la sección Diagrama trae el Mermaid que pide su nivel. |
| [internal/wiki/wiki.go](../../internal/wiki/wiki.go) | `revisarEjemplo` | `(pg *pagina, paginas map[string]*pagina) []string` | revisarEjemplo: el ejemplo lleva su ancla EJ- y encadena con el del padre. |
| [internal/wiki/wiki.go](../../internal/wiki/wiki.go) | `revisarNavegacion` | `(pg *pagina, hijos []string) []string` | revisarNavegacion: hay ↑ Subir y un enlace Abajo a cada hija. |
| [internal/wiki/wiki.go](../../internal/wiki/wiki.go) | `revisarEnlaces` | `(raiz string, pg *pagina, anclas map[string]map[string]bool) []string` | revisarEnlaces: cada enlace interno apunta a un archivo y ancla que existen. |
| [internal/wiki/wiki.go](../../internal/wiki/wiki.go) | `revisarAlcance` | `(paginas map[string]*pagina, orden []string) []string` | revisarAlcance: toda página del árbol debe alcanzarse desde docs/README.md siguiendo enlaces. |
| [internal/wiki/wiki.go](../../internal/wiki/wiki.go) | `resolver` | `(desde, destino string) (string, string)` | resolver convierte un destino relativo a la página en ruta relativa a la raíz + fragmento. |

<!-- /GENERADO -->

## Navegación

- ↑ Subir: [Inicio](../README.md)
