---
name: backlog
description: Muestra el backlog de este repo en columnas, en el orden en que se trabajará (aquas backlog table).
disable-model-invocation: true
allowed-tools: Bash(cd *) Bash(sh .claude/hooks/aquas.sh backlog table)
---
<!-- GENERADO por `aquas compile` desde gobierno/config/documentos.json. NO EDITAR: edita los canónicos y recompila. -->

# /backlog · E001/F007/S002

Backlog de este repo: primero los slices, en el orden en que se trabajarán; luego features e iniciativas.

!`cd "${CLAUDE_PROJECT_DIR}" && sh .claude/hooks/aquas.sh backlog table`

Muestra la tabla de arriba tal cual, dentro de un bloque de código. No agregues comentarios ni la reordenes.
