# E001 · Gobierno del trabajo · v0.1 · BORRADOR RATIFICADO
Origen: especificación SS-GOV-EFS v0.1.0 podada con el humano (2026-10-20). Estado: decisiones cerradas; implementado F001/S001 (gate de wiki en `internal/wiki/`), F001/S002 (legibilidad + glosario) y F001/S003 (referencia `docs/ref/`), F001/S004 (sección 9, migas y trazabilidad de pruebas en `internal/wiki/contrato.go`); F001 hecha. F005 hecha: S001–S003 (`internal/memoria/`; hooks SessionStart y Stop activos; `aquas memoria archivar`).
No compila a ningún agente hasta implementarse; luego se migra a wiki (`docs/`) + resumen Caveman en orquestador.

## Decisiones ratificadas
- Cero aprobaciones formales. Única puerta humana = interrogación cuando el agente tiene duda real.
- Hallazgo que cambia lo pedido = duda → preguntar. Si no, agente decide e informa en el cierre.
- Un solo slice activo por repo. Sin paralelismo de agentes salvo pedido explícito.
- Persistencia: JSON (stdlib Go), un archivo por entidad, versionado en git. Git = historia de revisiones.
- IDs de 3 cifras: `E001`, `E001/F001`, `E001/F001/S001`; turnos `T000001`; defectos `D000001`. Nunca se reutilizan.
- `E000 Mantenimiento`: iniciativa permanente para cambios pequeños; contrato mínimo autogenerado. Solo gobierno de ESTE repo central; productos y proyectos hijos no usan E000 de aquí.
- Contexto del agente en Caveman; explicación extensa en wiki con pedagogía acordada.
- Calidad primero; tokens segundo.
- Nombre: Aquas-driven development. No es spec-driven: contrato mínimo vigilado por gate, no comité. En evolución constante.
- Interrupción (2026-10-20): pedido nuevo con slice activo → clasificar antes de ejecutar. No es trabajo → responder. Trabajo urgente o que protege repo → slice activo PAUSADO (registrado) → slice E000 (contrato mínimo, rojo, verde, commit) → reanudar. Resto → backlog con prioridad; slice sigue.
- Defectos por repo (cada repo, su gobierno; nadie escribe en otros). Ver F004.

## Flujo de un turno
1. Hook `UserPromptSubmit` → Go guarda literal en auditoría (sin pasar por agente).
2. Agente refina: dudas → una ronda de preguntas; sin dudas → sigue.
3. Clasifica: pregunta · queja · información · trabajo. Solo trabajo entra a iniciativa (existente o nueva).
4. `aquas turno refinar` registra «pediste A → acordamos B» + clasificación + vínculo.
5. Trabajo → contrato de slice por CLI → rojo → implementar → verde → sello → cierre → commit.
6. Hook `Stop` → Go guarda respuesta del agente en auditoría (campo `last_assistant_message`: VERIFICADO 2026-10-21 en code.claude.com/docs/en/hooks; Stop se repite si otro hook frena → gana la última respuesta). Hecho en F002/S001 (`internal/turno/`; `aquas turno abrir|cerrar`; estado por sesión en `.aquas/turno.json`; guardia protege `auditoria/`). Límite: UserPromptSubmit también corre con mensajes que no escribe el humano (avisos de otras sesiones); se guardan igual.

## Auditoría
- `auditoria/turnos/AAAA-MM/T000001.md`: literal, respuesta, refinado, diferencia A→B.
- Todo se guarda, sensible o no. Fuera de git (`.gitignore`); respaldo lo hace el humano. Porqué (humano, 2026-10-21): la auditoría es transaccional y operativa, como los datos de una tabla; en Git va solo lo del repositorio y el producto.
- Turno interrumpido (sin Stop) queda con respuesta «—»: el turno siguiente no lo cierra (visto en T000001, 2026-10-21).

## Backlog
- `trabajo/activo/` y `trabajo/archivo/`; archivo por entidad: `E001.json`, `E001-F001.json`, `E001-F001-S001.json`, `D000001.json`.
- Solo `aquas` escribe. Control: hook PreToolUse niega Edit/Write en `trabajo/**` + pre-commit valida esquema.
- Elemento cerrado se archiva al cerrarse (no espera al padre); referencias siguen resolviendo.
- Prioridad obligatoria: P0 · P1 · P2 · P3 · SIN_TRIAGE, con motivo. Agente asigna; humano puede cambiar.
- Definir antes de ejecutar (regla del humano, 2026-10-21; E001/F002/S008): el backlog siempre tiene todos sus ítems definidos. Lo único que puede quedar sin definir es lo que se ingresa mientras corre un slice. Al terminar ese slice, la prioridad absoluta es definirlo y repriorizar el backlog. `siguiente` lo exige: iniciativa activa sin features activas o feature activa sin slices activos sale antes que cualquier slice libre, sin mirar prioridad ni bloqueo; iniciativas primero, luego features, cada grupo por secuencia de ingreso.
- Orden determinista de `aquas backlog siguiente`: en curso (EN_CURSO, VERIFICANDO; un solo slice activo) → sin definir (regla anterior) → desbloqueado → prioridad → secuencia de ingreso → ID. Desbloqueado = no BLOQUEADO y cada `bloqueado_por` archivado (CERRADO o CANCELADO: un bloqueador cancelado libera, si no la ficha quedaría trabada).
- Estados: BORRADOR · LISTO · EN_CURSO · VERIFICANDO · CERRADO; excepciones BLOQUEADO · CANCELADO. Archivado no es estado: es la carpeta `trabajo/archivo/` (decisión del agente en S002, 2026-10-21).
- Dependencias (decisión humana 2026-10-21): slice = proceso independiente, finito y testeable. Dependencia circular = un solo slice. No circular → primero el que se usa (`bloqueado_por`). Hoy acíclico por construcción: `bloqueado_por` solo se fija al crear y apunta a fichas existentes; un comando que lo cambie exige chequeo de ciclos.
- Hallazgo fuera de alcance del slice activo → `aquas backlog defecto`; no se toca en el slice.
- Alcance vigilado (F002/S007, decisión humana 2026-10-21: opción B): el slice declara sus rutas al arrancar; tocar otra ruta exige ampliar el alcance con motivo de una línea (queda registrado en la ficha); derivados del compilador siempre permitidos. Rechazadas: A bloqueo duro (traba con archivos imprevisibles: S001 tocó `.gitignore` y la guardia) y C solo aviso al cerrar (no previene).
- Ruido del sello (decisión humana 2026-10-21): los derivados del compilador (salidas en `gobierno/config/sellos.json` y ese archivo) no cuentan como tocados; ya los vigila el gate de derivados.

## Slice y feature
- Contrato slice: id, fuente (turnos), objetivo, alcance (incluye/excluye/rutas), comando de prueba, entrega esperada, docs enlazados.
- `aquas backlog rojo <id>`: Go corre cada prueba de la ficha sola (`go test -json`); exige que cada una falle por sí misma; lo registra con el motivo. `aquas backlog verde <id>`: las corre juntas; exige que todas pasen, después del rojo. `aquas backlog pruebas <id> …`: cambia la lista y borra rojo y verde (decisión humana 2026-10-21: comandos bajo `backlog`, no `slice`, porque aplican a E, F y S).
- Worker no se declara terminado; Go cierra solo con rojo→verde registrados + sello completo.
- Feature: prueba integral propia, roja al crear. Cierra = todos sus slices cerrados + integral verde.
- Slice activo único (S003): con un slice EN_CURSO o VERIFICANDO, `aquas backlog estado` niega cambiar otro slice; el gate marca dos en curso. No hay archivo de bloqueo ni `slice iniciar`.
- Decisiones humanas 2026-10-21 para S003 (constitución de la metodología): un solo slice en curso (EN_CURSO o VERIFICANDO); solo la ficha del slice activo cambia de estado (prioridad sí, en cualquiera); la memoria caliente rechaza un foco distinto del slice en curso del backlog. Contrato de prueba en la ficha: `pruebas` (nombre + intención), `rojo` y `verde` (registros); CERRADO exige rojo y luego verde. Rojo (corregido por el humano 2026-10-21 tras el informe de calidad): `go test -json` y CADA prueba nombrada en la ficha falla por sí misma; fallo de compilación u otra prueba no cuenta. Motivo: 15 de 92 pruebas no protegían lo que decían (una fallaba por otra causa) y 11 de 14 mutaciones sobrevivieron.
- Decisiones humanas 2026-10-21 (respuestas al contrato de S003): cada prueba en la ficha = nombre `./paquete:TestNombre` + intención (qué demuestra); el código se escribe al arrancar el trabajo, no al crear la ficha. Fichas anteriores a S003 no se migran; una activa recibe pruebas antes de cerrar. Con un slice en curso, otro slice no cambia de estado; iniciativas y features sí; prioridad en cualquiera; crear fichas sí.
- Rojo por mutación (decisión humana 2026-10-21, E000/F005): si la prueba endurece un código que ya funciona, el rojo se ve rompiendo a propósito la regla que vigila (una mutación por prueba, sin que una tape a otra), `aquas backlog rojo`, y se restaura el código antes del verde. La página del slice nombra cada mutación.
- Cierre con memoria archivada (decisión humana 2026-10-21, E001/F002/S009): `aquas backlog estado <slice> CERRADO` se niega si la memoria caliente apunta a ese slice y su archivo en `registro/memoria/` no existe. Orden: memoria fijar → memoria archivar → CERRADO. CANCELADO no lo exige.
- Directriz humana 2026-10-21: al crear iniciativa, feature o slice se escriben las pruebas de lo que el humano espera de esa etapa (integral en E y F, del slice en S), y su ficha guarda rojo y verde. Objetivo: ante un fallo futuro, saber qué prueba falló o cuál faltó. Git guarda la historia de cada ficha.

## Sello (motor de impacto mínimo)
- Al cerrar slice, Go arma lista obligatoria de revisión: iniciativa, feature, dependientes (fichas abiertas con `bloqueado_por` este slice) y páginas wiki del árbol. Decisión humana 2026-10-21: hermano = parentesco, no dependencia; un hermano sin dependencia ni archivos compartidos no entra.
- Alcance (decisión humana 2026-10-21): F002/S004 = árbol + nivel archivo (tocados = `git diff` desde el HEAD del rojo + nuevos; «nombra» = ruta exacta en texto o import Go del paquete; excluye `trabajo/`, `registro/`, `auditoria/`, `.tmp/`). Nivel trabajo → F002/S012. Sello en cierre de F y E: después.
- Mapa de relaciones (decisión humana 2026-10-21; absorbe F003/S004): dos niveles. Archivo: todo archivo del repo que nombra un archivo tocado por el slice. Trabajo: iniciativas, features y slices que comparten esos archivos o datos (no solo el árbol padre/hermanos). Caso que lo motivó: S002 movió el backlog a `trabajo/` y `internal/memoria/memoria.go` siguió apuntando al canónico; nada lo detectó.
- Cierre bloqueado hasta que cada ítem tenga `sin cambio` o `actualizado` + motivo de una línea.
- Go entrega solo líneas relevantes, no archivos completos (economía de tokens).

## Wiki pedagógica (gate en `aquas verificar`)
- Visor: VS Code (Markdown + Mermaid dibujado). Chat no dibuja Mermaid → en chat diagramas ASCII.
- Árbol: `docs/README.md` (inicio: qué es el repo, constitución explicada, mapa) → `docs/E001/README.md` → `F001/README.md` → `S001.md`; `docs/glosario.md`; referencia de lo existente: `docs/ref/<nombre>.md` (8 secciones, ancla `EJ-REF-<nombre>`, ≥1 Mermaid, sin continúa, enlazada desde inicio).
- Secciones fijas por página: En una frase (niño de 5 años) · Para qué existe (problema → solución) · Analogía (con límites) · Diagrama · Ejemplo de punta a punta · Paso a paso · Cómo se usa · Abajo / ↑ Subir.
- Diagramas por nivel: inicio y E = flujo alto nivel; F = proceso + estados; S = datos o secuencia.
- Ejemplos encadenados: un ejemplo hilo por iniciativa; hijo declara `continúa: EJ-<padre>` y enlaza ancla existente.
- Gate falla si: falta página de entidad · falta sección · falta diagrama del tipo exigido · enlace roto · página inalcanzable desde inicio · ejemplo sin `continúa` válido · término del glosario sin enlace en su primera aparición por página (`docs/glosario.md`: `## término` + definición + `Ejemplo:` + `Variantes:` opcional) · legibilidad bajo piso.
- Legibilidad: P = 206,835 − 62,3·sílabas/palabras − palabras/oraciones (Szigriszt-Pazos) ≥ 65 (INFLESZ «bastante fácil») por página + oraciones ≤ 20 palabras. Mide prosa; excluye código, tablas, títulos y Navegación. Fuentes en `internal/wiki/legibilidad.go`.
- Wiki es para humano: extensa permitida; Caveman no aplica. Claridad del ejemplo = criterio humano.
- Manual ≠ contrato: `docs/ref/` explica cómo funciona lo existente; páginas E/F/S son contrato + evidencia de lo hecho y por hacer.
- Sección 9 «Contrato y evidencia» (antes de Navegación) en E/F/S, con diagramas, explicación y ejemplo como el resto:
  - E: resultado esperado · resultado logrado · decisiones · features con estado · prueba integral.
  - F: objetivo · por qué esos slices · decisiones · prueba integral (escenario, resultado, commit) · slices con estado.
  - S: contrato · cada prueba por nombre + qué demuestra + rojo visto · commits · decisiones.
- Trazabilidad de pruebas (gate): toda prueba nombrada en un slice existe en código; toda `func Test…` del repo la nombra algún slice.
- Navegación por niveles (decisión humana 2026-10-21; reemplaza «inicio = mapa de todo E/F/S»): cada página explica completo SU nivel y solo enlaza a sus hijos directos; el mapa entero no cabe en una página porque crece.
  - Inicio: qué hace el repo en su totalidad, constitución, flujo actual del repo, cada iniciativa explicada + enlace. Sin features ni slices.
  - E: detalle completo de la iniciativa (diagramas, ejemplo, pruebas, contrato) + sus features, cada una con qué hace y enlace.
  - F: detalle completo de la feature + sus slices, cada uno con qué hace y enlace.
  - S: hoja; detalle completo del slice + sus pruebas con qué demuestran y rojo/verde.
  - Hijo cambia → padre se actualiza solo: lista de hijos y su estado (inicio, E, F) y pruebas (S) = bloque GENERADO por `aquas` desde las fichas `trabajo/`; gate falla si falta o está desfasado. Prosa y diagramas explicativos = criterio humano.
  - Página nace cuando su ficha arranca (EN_CURSO o CERRADO); gate falla si falta. Antes: fila sin enlace en el padre.
- Cada página E/F/S: migas «Inicio › E › F › S».
- E000 usa el mismo árbol: E000 → F → S.

## Vistas y diagramas en chat
- `aquas ver`: Go genera Mermaid del estado (árbol de backlog, slice, lista de sello) desde JSON; agente no los dibuja.
- Propuesta de proceso/diseño en chat → diagrama ASCII. Hook de cierre: bloque PROPUESTA sin bloque de código → falla. Límite: depende de que el agente marque PROPUESTA.

## Prioridades y estado
- Desde E001/F002/S002 (2026-10-21) estado, prioridad y orden viven en las fichas `trabajo/` (`aquas backlog ver | siguiente`). Las fichas mandan; este canónico guarda diseño y decisiones.
- Migración 2026-10-21 (humano aprobó): todo el backlog anterior pasó a fichas; lo cerrado se creó y archivó al momento por la traza (37 fichas: 23 archivo, 14 activo).

## Decisiones por feature (el estado vigente está en las fichas)
- E000 Mantenimiento · F003 Hooks globales (inventario; pedido humano 2026-10-21; solo registro, no se migran): viven en `~/.claude/hooks/`, rigen todos los proyectos y este repo no los gobierna, prueba ni documenta. `presupuesto.py` (UserPromptSubmit, PreToolUse): topes de llamadas y agentes por prompt, plan por lotes «Plan: N llamadas, M agentes». `reglas.py` (UserPromptSubmit, PreToolUse, Stop, SessionStart compact): sandbox, no releer, aperturas prohibidas, protocolo de fallo, certeza, contexto medido. `herdr-agent-state.ps1` (SessionStart): función NO CONFIRMADA. `cbm-*`: de codebase-memory-mcp; se van con él. Pruebas propias allá (`probar_*.py`), no corren aquí. Python, no Go (principio 4): migrar = decisión futura del humano.
- E000 Mantenimiento · F001 Guardia del repo: S001 regla de oro 5 (hecho, 6756b1d) · S002 lista de permitidos (hecho, 9e94e35) · S003 hooks no arrancaban en Windows: `args` lanza `sh` sin shell y el PATH no lo trae → `shell: bash` (hecho, bae7091). Wiki: hecha (docs/E000/). Backlog: fichas desde F002/S002. S004 entorno declarado (hecho; decisiones humanas 2026-10-20: archivo local no versionado, bloquea si falta o no coincide, settings.json compilado entero y no versionado): config declara sistema operativo y raíz del repo · `settings.json` se compila desde esa config con el shell del SO · la guardia valida su raíz contra la declarada (hoy la toma de `os.Getwd()`) · incluye el hueco de PowerShell (matcher no la cubre). Va antes de E001/F001/S004. F002 Compilador y sello (a posteriori, wiki hecha): S001 compilar y sellar (hecho, d2b7f3d) · S002 tope de caracteres (hecho, 458614f).
- E001/F001/S004 contrato y evidencia en la wiki. HECHO. F005 = SIGUIENTE. Tareas en este orden (contenido primero, gate al final para no bloquear el pre-commit):
  - T3 páginas E000 → F001 Guardia → S001–S004, ya con sección 9 y migas (hecho).
  - T4 sección 9 + migas en E001, F001, S001–S004; cada prueba existente nombrada por su slice dueño (hecho; las 8 pruebas de internal/compilador, anteriores al método, pasan a E000/F002 por decisión humana 2026-10-20).
  - T5 inicio: mapa completo + nombre de la metodología + estado de F001 corregido (hecho).
  - T1+T2 gate: sección 9, migas y trazabilidad de pruebas (rojo → verde); un solo commit con las páginas ya en regla (hecho).
- F005 Memoria caliente (pedido humano 2026-10-20; va después de E001/F001/S004; alcance mínimo aprobado; no se versiona en Git, solo su archivo al cerrar slice): el agente no pierde el estado del foco entre turnos, compactación o sesión nueva. Contiene foco (E/F/S y tarea), rutas a iniciativa, épica, feature, slice y backlog, y todo lo que no carga solo en contexto. Se archiva al cerrar el slice. Referencia de diseño: constitución HMP v1.0.0 del humano, solo como idea; no se importa.
  - Decisiones humanas 2026-10-20 (aceptó las 4 recomendaciones): vive en `.aquas/memoria.json` (oculto, ignorado por Git); al cerrar slice se archiva en `registro/memoria/<E>-<F>-<S>.json` (versionado). Épica = iniciativa (E); no hay nivel nuevo. Solo `aquas` la escribe; guardia bloquea Write/Edit/redirección hacia ella; sello SHA-256 detecta edición por otra vía (límite: un script puede escribirla; el sello lo delata).
  - Campos fijos: foco {iniciativa, feature, slice, tarea} · siguiente · pendientes[] · decisiones[] · huella_git · actualizado · sello. Rutas (iniciativa, feature, slice, backlog) se derivan del foco, no se guardan.
  - S001 archivo + CLI `aquas memoria ver|fijar` + sello + guardia (hecho). S002 hooks (hecho): SessionStart (startup|resume|clear|compact) → `aquas memoria inicio` imprime la memoria (nunca bloquea; sin memoria o sello roto → aviso); Stop bloquea si hay foco y la huella Git (HEAD + `git status`) cambió desde la última actualización, salvo `stop_hook_active` (límite: turno que solo decide sin tocar archivos no exige). Stop → `aquas memoria cierre`: sin memoria o foco vacío → pasa; sello roto → bloquea una vez. Huella = repo propio (show-toplevel = raíz) · HEAD + `git status --porcelain` excluyendo `.aquas/`; fuera de Git = vacía. Plantilla de settings admite `matcher` para eventos sin herramienta. Sello cubre los bytes escritos, no la estructura: un campo nuevo (huella_git) no hace pasar una memoria vieja por editada (fallo visto 2026-10-21). S003 `aquas memoria archivar` al cerrar slice + prueba de arranque en frío (hecho 2026-10-21; humano aceptó las 4 recomendaciones, a revisar con el uso): copia los bytes sellados de la memoria a `registro/memoria/<E>-<F>-<S>.json`; exige foco con slice, sello bien y memoria al día; no pisa un archivo existente; después conserva E, F, siguiente, pendientes y decisiones y vacía slice y tarea (siguiente se vaciaba: fallo visto en el primer archivo real 2026-10-21); guardia bloquea Write/Edit/Bash hacia `registro/memoria/`. Arranque en frío = prueba integral de F005 en `cmd/aquas/arranque_test.go`: sesión 1 fija y el cierre exige; sesión 2 recibe el foco solo por la orden del hook SessionStart de la plantilla real; tras archivar arranca en E › F.
- Idea futura (sin slice; pensar más adelante): `aquas` crea repos hijos copiando la versión sellada del gobierno; cambio en el central deja hijos desactualizados → `aquas gobierno update` o similar.
- F001 Wiki + gate: S001 esqueleto + gate de estructura, enlaces, diagramas, ejemplos · S002 glosario + legibilidad · S003 documentar lo existente (constitución, compilador, comunicación) · S004 contrato y evidencia (hecho, arriba).
- F002 Turnos, backlog y sello: S001 auditoría (hooks) · S002 backlog JSON + CLI + prioridad + archivo · S003 rojo/verde + un solo slice en curso · S004 sello · S005 `aquas ver` · S006 hook PROPUESTA → diagrama.
- F004 Defectos y reincidencia: registro `D000001` por repo con firma estable (agente elige existente o justifica nueva; elección = probabilística) · campos cortos: qué, causa, arreglo, cómo me cercioro, ocurrencias [turnos] · hook Stop frena si «¿Salió algo mal?» ≠ «No» sin defecto registrado · ocurrencias ≥ umbral (config, 3) → `aquas` crea slice E000 P1. Semilla: D000001 `hook-exit-go-run` (go run convierte exit 2 en 1), ocurrencias previas declaradas por humano: muchas.
- F006 Agentes worker y auditor (pedido humano, reiterado 2026-10-21; no estaba en backlog): `worker-sonnet` (sonnet), `worker-haiku` (haiku), `auditor` (opus, effort high). HECHO (S001): un cuerpo `worker.md` para ambos workers (sin reglas de comunicación con el humano); `auditor.md` solo lee (`tools: Read, Grep, Glob, Bash`; límite: Bash puede escribir, lo prohíbe su cuerpo = criterio); compilador admite `tools`; orquestador elige: mecánico → haiku, código → sonnet, slice terminado → auditor. S002 (hecho 2026-10-21; humano aceptó los 3 ajustes y el hook): `reglas/02_COMUNICACION_ENTRE_AGENTES.md` (ENCARGO y CIERRE fijos, caveman; temporales propios en `.tmp/<id>/`, nunca `.tmp` entero) para los 4 agentes; hook SubagentStop → `aquas cierre-agente` (`internal/encargo`) exige el CIERRE (exit 2 → stderr llega al subagente; `stop_hook_active` deja cerrar). Hook exige: primera línea, cada campo con valor («—» si no aplica), nada fuera del cierre; tolera cerca ``` y sangría. JSON ilegible → deja cerrar. Límite: revisa forma, no verdad; eso es del auditor. Auditoría 2026-10-21: CON_HALLAZGOS, 5 corregidos.
- F003 Bibliotecario: S001 leer lista de rutas · S002 buscar por regex (DESCARTADO 2026-10-21) · S003 esquema (firmas Go, títulos Markdown) · S004 relaciones (referencias y usos) conectadas al sello. Se explora a fondo al llegar.
  - Decisión humana 2026-10-21: herramienta propia en Go; NO codebase-memory-mcp (repo ajeno, sin confianza; el humano lo desinstala). Meta: agente pasa arreglo de rutas (o regex) y recibe solo lo útil, mínimo y con ahorro grande de tokens. Plugin Go de relaciones entre código (S003) va después de que la biblioteca funcione. Diseño: analizar a fondo antes del contrato de S001.
  - Decisiones humanas 2026-10-21: interfaz = comando `aquas biblioteca …` por Bash, arreglo de rutas en JSON por stdin (MCP propio solo si el uso lo pide). Esquema (firmas Go con `go/ast`, títulos Markdown) en S003, no en S001. Ahorro = prueba determinista: misma tarea por bibliotecario vs lectura completa, comparando caracteres. Ahorro real esperado: lote en una llamada, solo lo útil, tope + cuenta «devolví X de Y».
  - S001 HECHO 2026-10-21: `internal/biblioteca` + `aquas biblioteca leer` (stdin `{"rutas":["a.go:10-40","b.md"],"tope":0}`; bloques `== ruta · d-h de N` numerados; error por ruta sin frenar el lote; solo rutas relativas al repo; tope 200000 = desborde extremo, avisa corte; última línea `== devueltos X de Y caracteres`). Ahorro medido en prueba: 435 vs 10472 caracteres (96 %) frente a lectura completa; frente a Read con rango solo ahorra llamadas y numeración (auditor 2026-10-21). Auditoría: CON_HALLAZGOS, 5 corregidos (vacío, carpeta, grafía repetida, mensajes CLI, fixture); enlaces simbólicos = límite, sin uso hoy. Worker y orquestador lo usan para lotes.
  - S002 DESCARTADO 2026-10-21 (humano aceptó con datos): telemetría de las 3 sesiones del repo (576 vueltas, 610 herramientas): rachas seguidas de buscar/leer = 65, máximo 3 vueltas, 6 mezclan buscar y leer; fundirlas ahorraría ≤ 14 vueltas (≈ 2 % de la caché leída). Grep y Bash con tuberías ya buscan y leen en una vuelta. Métrica correcta del ahorro = vueltas al modelo × contexto, no caracteres. Script: `.tmp/telemetria/` (no versionado).
  - F003 CERRADA 2026-10-21 con S001 (humano aceptó). S003 esquema y S004 relaciones aplazados: telemetría mide leer/buscar ≈ 18 % de las llamadas; pruebas+gate 28 %, escribir/editar 22 %.

## Podado (fuera de alcance)
- Cierre transitivo completo y recibos por documento · token de reserva, CAS global, journal, recuperación ante caídas.
- Aislamiento a nivel de SO · 11 estados de turno y 11 categorías · señales de frustración/telemetría emocional.
- Entidades Change Proposal, Review Batch, Seal Receipt, Execution Attempt · similitud semántica en CLI.
- Traductor genérico de Markdown en Go.
