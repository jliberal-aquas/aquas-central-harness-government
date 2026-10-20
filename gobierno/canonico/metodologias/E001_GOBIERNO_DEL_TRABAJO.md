# E001 · Gobierno del trabajo · v0.1 · BORRADOR RATIFICADO
Origen: especificación SS-GOV-EFS v0.1.0 podada con el humano (2026-10-20). Estado: decisiones cerradas; implementado F001/S001 (gate de wiki en `internal/wiki/`) y F001/S002 (legibilidad + glosario).
No compila a ningún agente hasta implementarse; luego se migra a wiki (`docs/`) + resumen Caveman en orquestador.

## Decisiones ratificadas
- Cero aprobaciones formales. Única puerta humana = interrogación cuando el agente tiene duda real.
- Hallazgo que cambia lo pedido = duda → preguntar. Si no, agente decide e informa en el cierre.
- Un solo slice activo por repo. Sin paralelismo de agentes salvo pedido explícito.
- Persistencia: JSON (stdlib Go), un archivo por entidad, versionado en git. Git = historia de revisiones.
- IDs de 3 cifras: `E001`, `E001/F001`, `E001/F001/S001`; turnos `T000001`; defectos `D000001`. Nunca se reutilizan.
- `E000 Mantenimiento`: iniciativa permanente para cambios pequeños; contrato mínimo autogenerado.
- Contexto del agente en Caveman; explicación extensa en wiki con pedagogía acordada.
- Calidad primero; tokens segundo.

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
- Árbol: `docs/README.md` (inicio: qué es el repo, constitución explicada, mapa) → `docs/E001/README.md` → `F001/README.md` → `S001.md`; `docs/glosario.md`.
- Secciones fijas por página: En una frase (niño de 5 años) · Para qué existe (problema → solución) · Analogía (con límites) · Diagrama · Ejemplo de punta a punta · Paso a paso · Cómo se usa · Abajo / ↑ Subir.
- Diagramas por nivel: inicio y E = flujo alto nivel; F = proceso + estados; S = datos o secuencia.
- Ejemplos encadenados: un ejemplo hilo por iniciativa; hijo declara `continúa: EJ-<padre>` y enlaza ancla existente.
- Gate falla si: falta página de entidad · falta sección · falta diagrama del tipo exigido · enlace roto · página inalcanzable desde inicio · ejemplo sin `continúa` válido · término del glosario sin enlace en su primera aparición por página (`docs/glosario.md`: `## término` + definición + `Ejemplo:` + `Variantes:` opcional) · legibilidad bajo piso.
- Legibilidad: P = 206,835 − 62,3·sílabas/palabras − palabras/oraciones (Szigriszt-Pazos) ≥ 65 (INFLESZ «bastante fácil») por página + oraciones ≤ 20 palabras. Mide prosa; excluye código, tablas, títulos y Navegación. Fuentes en `internal/wiki/legibilidad.go`.
- Wiki es para humano: extensa permitida; Caveman no aplica. Claridad del ejemplo = criterio humano.

## Vistas y diagramas en chat
- `aquas ver`: Go genera Mermaid del estado (árbol de backlog, slice, lista de sello) desde JSON; agente no los dibuja.
- Propuesta de proceso/diseño en chat → diagrama ASCII. Hook de cierre: bloque PROPUESTA sin bloque de código → falla. Límite: depende de que el agente marque PROPUESTA.

## Features y slices iniciales
- F001 Wiki + gate: S001 esqueleto + gate de estructura, enlaces, diagramas, ejemplos · S002 glosario + legibilidad · S003 documentar lo existente (constitución, compilador, comunicación).
- F002 Turnos, backlog y sello: S001 auditoría (hooks) · S002 backlog JSON + CLI + prioridad + archivo · S003 slice rojo/verde + bloqueo · S004 sello · S005 `aquas ver` · S006 hook PROPUESTA → diagrama.
- F003 Bibliotecario: S001 leer lista de rutas · S002 buscar por regex · S003 plugin Go (referencias y usos) · S004 conectar usos al sello. Se explora a fondo al llegar.

## Podado (fuera de alcance)
- Cierre transitivo completo y recibos por documento · token de reserva, CAS global, journal, recuperación ante caídas.
- Aislamiento a nivel de SO · 11 estados de turno y 11 categorías · señales de frustración/telemetría emocional.
- Entidades Change Proposal, Review Batch, Seal Receipt, Execution Attempt · similitud semántica en CLI.
- Traductor genérico de Markdown en Go.
