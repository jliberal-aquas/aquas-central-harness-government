# CONSTITUCIÓN · claude-code-scaffolding · v0.1

## Misión
- Este repo GOBIERNA y GENERA la estructura inicial de todo proyecto trabajado con Claude Code.
- Fuente única de verdad para: constitución, metodologías, gates y validadores deterministas.
- Proyectos hijos heredan desde aquí; no se inventan reglas locales que contradigan esta constitución.

## Principios
1. Gobierno incremental: se construye poco a poco. Regla no escrita aquí (o en archivo enlazado) = regla no vigente.
2. Determinismo primero: toda regla verificable se valida con un programa (gate), no con juicio del modelo.
3. Lo no verificable por programa se marca explícitamente como «criterio humano» o «probabilístico».
4. Go = lenguaje único de gates, validadores y herramientas del repo.
5. Un cambio coherente = un commit. Git local obligatorio. Mensaje explica el porqué.
6. Evidencia antes de afirmar: no declarar «hecho» sin prueba ejecutada.
7. Contexto permanente en formato Caveman: corto, denso, sin relleno. Detalle extenso va en archivos bajo demanda.

## Mapa del repo
- `CLAUDE.md` → esta constitución (contexto permanente).
- `.claude/settings.json` → configuración del proyecto; fija al orquestador como agente por defecto.
- `.claude/agents/orchestrator.md` → agente orquestador: único que habla con el humano.
- `gobierno/comunicacion/` → reglas de comunicación agente → humano.
- Futuro: `gates/` (programas Go), `metodologias/`, `plantillas/` (scaffolding de proyectos).

## Alcance de reglas de comunicación
- Aplican SOLO al orquestador (hilo principal que conversa con el humano).
- Subagentes, código, logs y artefactos técnicos: NO aplican; ignorar las secciones siguientes.

@gobierno/comunicacion/01_COMUNICACION_HUMANO.md
@gobierno/comunicacion/02_PEDAGOGIA_CHAT.md
@gobierno/comunicacion/03_INSIGHTS_ANTI_REINCIDENCIA.md
