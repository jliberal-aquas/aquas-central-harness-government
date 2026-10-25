<!-- GENERADO por `aquas compilar` desde gobierno/config/documentos.json. NO EDITAR: edita los canónicos y recompila. -->

<!-- hueco: titulo -->
# CONSTITUCIÓN · aquas-harness-governance · v0.2
<!-- /hueco -->

## Misión
<!-- hueco: mision -->
- Este repo GOBIERNA y GENERA la estructura inicial de todo proyecto trabajado con Claude Code.
- Fuente única de verdad para: constitución, metodologías, gates y validadores deterministas.
- Proyectos hijos heredan desde aquí; no se inventan reglas locales que contradigan esta constitución.
<!-- /hueco -->

## Principios
1. Gobierno incremental: se construye poco a poco. Regla no escrita en un canónico = regla no vigente.
2. Determinismo primero: toda regla verificable se valida con un programa (gate), no con juicio del modelo.
3. Lo no verificable por programa se marca explícitamente como «criterio humano» o «probabilístico».
4. Go = lenguaje único de gates, validadores y herramientas. Módulo: `github.com/jliberal-aquas/aquas-central-harness-government`.
5. Un cambio coherente = un commit. Git local obligatorio. Mensaje explica el porqué.
6. Evidencia antes de afirmar: no declarar «hecho» sin prueba ejecutada.
7. Contexto permanente en formato Caveman: corto, denso, sin relleno. Detalle extenso va en archivos bajo demanda.
8. Fuente canónica → compilado: se edita solo el canónico; todo derivado se regenera y se sella.
9. Calidad primero; economía de tokens segundo. Token = moneda de todo uso agéntico: medir y justificar cada gasto.

## Canónicos y compilación
<!-- solo-hijo
- Gobierno central sellado: `.aquas/central/` (solo lectura; lo verifica `.aquas/lock.json`). NO editar.
- Lo propio del hijo: `gobierno/canonico/huecos/<nombre>.md`, uno por hueco de la central.
- Compilar: `aquas build .` → regenera este `CLAUDE.md`. NO editarlo a mano.
- Defecto del gobierno → se reporta a la central; no se corrige en el hijo.
-->
<!-- solo-central -->
- Canónicos: `gobierno/canonico/**`. Únicos archivos editables de gobierno.
- Config: `gobierno/config/agentes.json` (agentes) y `gobierno/config/documentos.json` (documentos como este).
- Compilados: `CLAUDE.md`, `.claude/agents/*.md`. Llevan marca «GENERADO»; NO editar a mano.
- Presupuesto: `max_caracteres` obligatorio por salida en config; gate falla si se excede.
- Tope = freno de desborde extremo, con holgura amplia; no es meta. Nunca condensar una regla hasta perder contexto para caber: se sube el tope.
- Sello: `gobierno/config/sellos.json` guarda SHA-256 de cada fuente y cada salida.
- Compilar: `go run ./cmd/aquas compilar`. Verificar (gate): `go build -o .tmp/aquas.exe ./cmd/aquas && ./.tmp/aquas.exe verificar`.
- Disparadores: hook PostToolUse de Claude Code (recompila al editar) + git pre-commit (bloquea si hay desfase).
- Clon nuevo: `git config core.hooksPath .githooks` + `aquas iniciar` (SO y raíz).
<!-- /solo-central -->

## Mapa del repo
<!-- hueco: mapa -->
- `cmd/aquas/` → CLI del harness. `internal/compilador/` → compilador y sellado. `internal/wiki/` → gate wiki.
- `docs/` → wiki para humanos; gate la exige.
- `.claude/settings.json` → local, generado de `gobierno/config/claude_settings.json`.
- `gobierno/canonico/agentes/` → cuerpo de cada agente. `gobierno/canonico/reglas/` → reglas enlazables.
- `gobierno/canonico/comunicacion/` → reglas agente → humano; solo las carga el orquestador.
- `metodologias/` = specs. Futuro: `plantillas/`, más gates.
<!-- /hueco -->
