# E001 · Gobierno del trabajo · v0.1 · BORRADOR RATIFICADO
Origen: especificación SS-GOV-EFS v0.1.0 podada con el humano (2026-10-20). Estado: decisiones cerradas; implementado F001/S001 (gate de wiki en `internal/wiki/`), F001/S002 (legibilidad + glosario) y F001/S003 (referencia `docs/ref/`), F001/S004 (sección 9, migas y trazabilidad de pruebas en `internal/wiki/contrato.go`); F001 hecha. F005/S001 y S002 hechos (`internal/memoria/`; hooks SessionStart y Stop activos); F005/S003 siguiente.
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
6. Hook `Stop` → Go guarda respuesta del agente en auditoría (campo `last_assistant_message`: PROBABLE; verificar).

## Auditoría
- `auditoria/turnos/AAAA-MM/T000001.md`: literal, respuesta, refinado, diferencia A→B.
- Todo se guarda, sensible o no. Fuera de git (`.gitignore`); respaldo lo hace el humano.

## Backlog
- `trabajo/activo/` y `trabajo/archivo/`; archivo por entidad: `E001.json`, `E001-F001.json`, `E001-F001-S001.json`, `D000001.json`.
- Solo `aquas` escribe. Control: hook PreToolUse niega Edit/Write en `trabajo/**` + pre-commit valida esquema.
- Elemento cerrado se archiva al cerrarse (no espera al padre); referencias siguen resolviendo.
- Prioridad obligatoria: P0 · P1 · P2 · P3 · SIN_TRIAGE, con motivo. Agente asigna; humano puede cambiar.
- Orden determinista de `aquas backlog siguiente`: desbloqueado → prioridad → secuencia de ingreso → ID.
- Estados: BORRADOR · LISTO · EN_CURSO · VERIFICANDO · CERRADO · ARCHIVADO; excepciones BLOQUEADO · CANCELADO.
- Hallazgo fuera de alcance del slice activo → `aquas backlog defecto`; no se toca en el slice.

## Slice y feature
- Contrato slice: id, fuente (turnos), objetivo, alcance (incluye/excluye/rutas), comando de prueba, entrega esperada, docs enlazados.
- `aquas slice rojo`: Go corre la prueba; exige fallo; lo registra. `aquas slice verde`: exige éxito posterior.
- Worker no se declara terminado; Go cierra solo con rojo→verde registrados + sello completo.
- Feature: prueba integral propia, roja al crear. Cierra = todos sus slices cerrados + integral verde.
- Reserva de slice activo: archivo de bloqueo; segundo `slice iniciar` falla.

## Sello (motor de impacto mínimo)
- Al cerrar slice, Go arma lista obligatoria de revisión: iniciativa, feature, slices hermanos, páginas wiki enlazadas.
- F003 agrega usos en código (bibliotecario). Hasta entonces: punto de conexión vacío.
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
- Inicio: mapa de todo E/F/S (hecho y planeado, con estado). Cada página E/F/S: migas «Inicio › E › F › S». Escritos a mano hasta F002; luego `aquas` los genera desde backlog JSON (bloque GENERADO).
- E000 usa el mismo árbol: E000 → F → S.

## Vistas y diagramas en chat
- `aquas ver`: Go genera Mermaid del estado (árbol de backlog, slice, lista de sello) desde JSON; agente no los dibuja.
- Propuesta de proceso/diseño en chat → diagrama ASCII. Hook de cierre: bloque PROPUESTA sin bloque de código → falla. Límite: depende de que el agente marque PROPUESTA.

## Features y slices (backlog provisional hasta F002/S002; orden de ejecución = orden de lista)
- E000 Mantenimiento · F001 Guardia del repo: S001 regla de oro 5 (hecho, 6756b1d) · S002 lista de permitidos (hecho, 9e94e35) · S003 hooks no arrancaban en Windows: `args` lanza `sh` sin shell y el PATH no lo trae → `shell: bash` (hecho, bae7091). Wiki: hecha (docs/E000/). Backlog: pendiente (F002/S002). S004 entorno declarado (hecho; decisiones humanas 2026-10-20: archivo local no versionado, bloquea si falta o no coincide, settings.json compilado entero y no versionado): config declara sistema operativo y raíz del repo · `settings.json` se compila desde esa config con el shell del SO · la guardia valida su raíz contra la declarada (hoy la toma de `os.Getwd()`) · incluye el hueco de PowerShell (matcher no la cubre). Va antes de E001/F001/S004. F002 Compilador y sello (a posteriori, wiki hecha): S001 compilar y sellar (hecho, d2b7f3d) · S002 tope de caracteres (hecho, 458614f).
- E001/F001/S004 contrato y evidencia en la wiki. HECHO. F005 = SIGUIENTE. Tareas en este orden (contenido primero, gate al final para no bloquear el pre-commit):
  - T3 páginas E000 → F001 Guardia → S001–S004, ya con sección 9 y migas (hecho).
  - T4 sección 9 + migas en E001, F001, S001–S004; cada prueba existente nombrada por su slice dueño (hecho; las 8 pruebas de internal/compilador, anteriores al método, pasan a E000/F002 por decisión humana 2026-10-20).
  - T5 inicio: mapa completo + nombre de la metodología + estado de F001 corregido (hecho).
  - T1+T2 gate: sección 9, migas y trazabilidad de pruebas (rojo → verde); un solo commit con las páginas ya en regla (hecho).
- F005 Memoria caliente (pedido humano 2026-10-20; va después de E001/F001/S004; alcance mínimo aprobado; no se versiona en Git, solo su archivo al cerrar slice): el agente no pierde el estado del foco entre turnos, compactación o sesión nueva. Contiene foco (E/F/S y tarea), rutas a iniciativa, épica, feature, slice y backlog, y todo lo que no carga solo en contexto. Se archiva al cerrar el slice. Referencia de diseño: constitución HMP v1.0.0 del humano, solo como idea; no se importa.
  - Decisiones humanas 2026-10-20 (aceptó las 4 recomendaciones): vive en `.aquas/memoria.json` (oculto, ignorado por Git); al cerrar slice se archiva en `registro/memoria/<E>-<F>-<S>.json` (versionado). Épica = iniciativa (E); no hay nivel nuevo. Solo `aquas` la escribe; guardia bloquea Write/Edit/redirección hacia ella; sello SHA-256 detecta edición por otra vía (límite: un script puede escribirla; el sello lo delata).
  - Campos fijos: foco {iniciativa, feature, slice, tarea} · siguiente · pendientes[] · decisiones[] · huella_git · actualizado · sello. Rutas (iniciativa, feature, slice, backlog) se derivan del foco, no se guardan.
  - S001 archivo + CLI `aquas memoria ver|fijar` + sello + guardia (hecho). S002 hooks (hecho): SessionStart (startup|resume|clear|compact) → `aquas memoria inicio` imprime la memoria (nunca bloquea; sin memoria o sello roto → aviso); Stop bloquea si hay foco y la huella Git (HEAD + `git status`) cambió desde la última actualización, salvo `stop_hook_active` (límite: turno que solo decide sin tocar archivos no exige). Stop → `aquas memoria cierre`: sin memoria o foco vacío → pasa; sello roto → bloquea una vez. Huella = repo propio (show-toplevel = raíz) · HEAD + `git status --porcelain` excluyendo `.aquas/`; fuera de Git = vacía. Plantilla de settings admite `matcher` para eventos sin herramienta. Sello cubre los bytes escritos, no la estructura: un campo nuevo (huella_git) no hace pasar una memoria vieja por editada (fallo visto 2026-10-21). S003 `aquas memoria archivar` al cerrar slice + prueba de arranque en frío.
- Idea futura (sin slice; pensar más adelante): `aquas` crea repos hijos copiando la versión sellada del gobierno; cambio en el central deja hijos desactualizados → `aquas gobierno update` o similar.
- F001 Wiki + gate: S001 esqueleto + gate de estructura, enlaces, diagramas, ejemplos · S002 glosario + legibilidad · S003 documentar lo existente (constitución, compilador, comunicación) · S004 contrato y evidencia (hecho, arriba).
- F002 Turnos, backlog y sello: S001 auditoría (hooks) · S002 backlog JSON + CLI + prioridad + archivo · S003 slice rojo/verde + bloqueo · S004 sello · S005 `aquas ver` · S006 hook PROPUESTA → diagrama.
- F004 Defectos y reincidencia: registro `D000001` por repo con firma estable (agente elige existente o justifica nueva; elección = probabilística) · campos cortos: qué, causa, arreglo, cómo me cercioro, ocurrencias [turnos] · hook Stop frena si «¿Salió algo mal?» ≠ «No» sin defecto registrado · ocurrencias ≥ umbral (config, 3) → `aquas` crea slice E000 P1. Semilla: D000001 `hook-exit-go-run` (go run convierte exit 2 en 1), ocurrencias previas declaradas por humano: muchas.
- F003 Bibliotecario: S001 leer lista de rutas · S002 buscar por regex · S003 plugin Go (referencias y usos) · S004 conectar usos al sello. Se explora a fondo al llegar.

## Podado (fuera de alcance)
- Cierre transitivo completo y recibos por documento · token de reserva, CAS global, journal, recuperación ante caídas.
- Aislamiento a nivel de SO · 11 estados de turno y 11 categorías · señales de frustración/telemetría emocional.
- Entidades Change Proposal, Review Batch, Seal Receipt, Execution Attempt · similitud semántica en CLI.
- Traductor genérico de Markdown en Go.
