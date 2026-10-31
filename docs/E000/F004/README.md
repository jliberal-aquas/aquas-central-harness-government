# F004 · Vueltas de pruebas y gate

[Inicio](../../README.md) › [E000](../README.md) › F004

[Feature](../../glosario.md#feature) que mide cuántas veces se repiten las pruebas y el [gate](../../glosario.md#gate) antes de pasar. Cada repetición cuesta llamadas y contexto.

## En una frase

Antes de abaratar las vueltas, contarlas y saber por qué ocurren.

## Para qué existe

**El problema.** Una muestra a mano del 2026-10-21 vio que casi la mitad de las corridas fallaban. Nadie sabía cuáles ni por qué.

**La solución.** Un [slice](../../glosario.md#slice) que cuenta corridas, fallas y motivos leyendo el historial de Claude Code.

## Analogía

Piensa en un **taller que anota cada devolución**.

- Sin la libreta, solo hay una sensación de «se devuelven muchos».
- Con la libreta, se ve que casi todas vienen de la misma pieza.

Límite de la analogía: aquí la libreta la llena un programa, no una persona.

## Diagrama

```mermaid
flowchart LR
    H["Historial .jsonl"] --> M["aquas measure runs"]
    M --> C["Corridas y fallas por tipo"]
    C --> T["Cinco motivos principales"]
```

Estados de una vuelta:

```mermaid
stateDiagram-v2
    [*] --> Corrida: go test o verificar
    Corrida --> Verde: pasa
    Corrida --> Falla: falla con motivo
    Falla --> Corrida: otra vuelta
    Verde --> [*]
```

| Slice | Qué entrega |
|---|---|
| [S001 Medir vueltas](S001.md) | `aquas measure runs` con motivos |
| [S002 Por qué cae la wiki](S002.md) | Caídas de la wiki por regla |
| [S003 Enlazar el glosario solo](S003.md) | `aquas wiki link` |
| [S004 Enlazar al guardar](S004.md) | El hook enlaza al guardar |

## Ejemplo de punta a punta

<a id="EJ-E000-F004"></a>

continúa: [EJ-E000](../README.md#EJ-E000)

1. Se corre `aquas measure runs`.
2. Sale `verificar 417/166 (wiki 143, derivados 22, …)`.
3. Se ve que la [wiki](../../glosario.md#wiki) causa casi todas las caídas del gate.

## Paso a paso

1. Medir vueltas de pruebas y gate (S001).
2. Desglosar las caídas de la wiki por regla (S002).
3. Enlazar el glosario sin vueltas (S003).
4. Enlazar al guardar, sin depender del [agente](../../glosario.md#agente) (S004).

## Cómo se usa

- Medir: `aquas measure runs`.
- Ver las fichas: `aquas backlog show E000/F004`.

## Contrato y evidencia

**Objetivo.** Saber, con números, cuántas corridas fallan y por qué.

Prueba integral: `TestCuentaVueltasPorMotivo`, en S001.

## Navegación

<!-- GENERADO por aquas desde trabajo/: no editar a mano; `aquas compile` lo rehace. -->

| Hijo | Qué es | Estado |
|---|---|---|
| [S001](S001.md) | Medir vueltas de pruebas y gate | Hecho |
| [S002](S002.md) | Por qué cae la wiki | Hecho |
| [S003](S003.md) | Enlazar el glosario solo | Hecho |
| [S004](S004.md) | Enlazar al guardar | Hecho |

<!-- /GENERADO -->

- ↑ Subir: [E000 · Mantenimiento del gobierno](../README.md)
- Abajo: [S001 · Medir vueltas](S001.md)
- Abajo: [S002 · Por qué cae la wiki](S002.md)
- Abajo: [S003 · Enlazar el glosario solo](S003.md)
- Abajo: [S004 · Enlazar al guardar](S004.md)
