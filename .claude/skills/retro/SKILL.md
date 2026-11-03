---
name: retro
description: Guía la retrospectiva del gobierno: fuentes, categorías, reparto de cada control y salida al backlog. Úsala cuando el humano pida /retro o cuando aquas backlog next avise «toca retrospectiva».
---
<!-- GENERADO por `aquas compile` desde gobierno/config/documentos.json. NO EDITAR: edita los canónicos y recompila. -->

# /retro · E001/F018/S001

La invoca el humano con `/retro`. También el orquestador, cuando `aquas backlog next` avisa «toca retrospectiva».
Límite: el skill no crea fichas solo. Propone; el humano elige; lo elegido entra por su orden.

## Fuentes

Leer solo desde la última retro, en este orden:
1. Defectos agrupados por familia: `aquas defect show` muestra la familia de cada uno. Familia con varias apariciones = patrón.
2. Corridas de agentes: `aquas measure runs`. Llamadas, vueltas y gasto por encargo.
3. Correcciones del humano en los turnos: cada «no», «otra vez» o rehacer pedido.
4. Registro de sesión, si existe.

Cada hallazgo cita su fuente: id de defecto, turno o corrida. Sin fuente no entra.

## Categorías

Clasificar cada hallazgo en una:
- Control que falta: el fallo no tiene programa que lo detecte.
- Control sin conectar: existe, pero ningún gancho ni gate lo corre.
- Regla de criterio: no se puede verificar por programa; la juzga el auditor.
- Navegación: el agente buscó de más para encontrar algo.
- Economía de herramientas: llamadas o lecturas que sobraron.
- Instrucción que no cambia nada: texto cargado que ninguna conducta usa.
- Acceso a información: dato que el agente necesitó y no tenía.

## Reparto

Dónde vive cada control, de más a menos determinista:
- Mecánico (condición precisa, resultado esperado) → ficha de gate o gancho, con su prueba.
- Criterio → regla del auditor, en su canónico.
- Orientación → puntero corto en el canónico que corresponde.

Preferir siempre el reparto más determinista posible. Instrucción que no cambia nada → proponer quitarla.

En un hijo, lo que toca un canónico sellado de la central no se arregla allí: se reporta con `aquas report`.

## Salida

1. Candidatos ordenados por gravedad, cada uno con fuente, categoría, reparto y recomendación.
2. El humano elige. Cada candidato elegido entra con `aquas backlog new`.
3. Cierre de la retro con `retro close`, nombrando las fichas creadas.
