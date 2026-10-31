# Glosario

Palabras técnicas de esta wiki, explicadas en simple. Cada página enlaza aquí la primera vez que usa una de ellas.

## agente

Un programa de IA que recibe un pedido y trabaja con herramientas: lee, escribe y ejecuta.

Variantes: agentes

Ejemplo: el orquestador es el agente que habla contigo.

## ancla

Una marca invisible dentro de una página. Un enlace puede saltar justo a ese punto.

Variantes: anclas

Ejemplo: `<a id="EJ-E001"></a>` es un ancla; el enlace `README.md#EJ-E001` salta ahí.

## backlog

La lista de trabajo pendiente, ordenada por prioridad. Aquí vivirá en un archivo JSON.

Ejemplo: «S003 documentar lo existente» está en el backlog hasta que alguien la toma.

## canónico

El archivo oficial de una regla. Es el único que se edita a mano. Todo lo demás se genera desde él.

Variantes: canónicos, canónica, canónicas

Ejemplo: `gobierno/canonico/constitucion.md` es canónico; `CLAUDE.md` se genera desde él.

## commit

Una foto guardada del repositorio en git. Lleva un mensaje que explica el porqué del cambio.

Variantes: commits

Ejemplo: «E001/F001/S002: el gate mide la legibilidad de la wiki» es un commit.

## compilador

El programa que lee los archivos canónicos y escribe los archivos generados, como `CLAUDE.md`.

Ejemplo: `go run ./cmd/aquas compile` ejecuta el compilador.

## derivado

Un archivo que nadie escribe a mano. El compilador lo genera desde los canónicos.

Variantes: derivados, derivada, derivadas

Ejemplo: `CLAUDE.md` es derivado; sale de `gobierno/canonico/constitucion.md`.

## feature

Una parte grande de una iniciativa. Se divide en slices.

Variantes: features

Ejemplo: F001 Wiki + gate es una feature de la iniciativa E001.

## gate

Un programa que revisa algo y dice sí o no. Si dice no, el trabajo no avanza.

Variantes: gates

Ejemplo: `aquas verify` es un gate; si falla, el commit no entra.

## Go

El lenguaje de programación de todas las herramientas de este repositorio.

Ejemplo: el gate de la wiki está escrito en Go, en `internal/wiki/`.

## hook

Un programa que se ejecuta solo cuando pasa algo. Nadie tiene que acordarse de llamarlo.

Variantes: hooks

Ejemplo: al guardar un canónico, un hook de Claude Code ejecuta el compilador.

## iniciativa

El nivel más grande del trabajo. Agrupa features con una misma meta.

Variantes: iniciativas

Ejemplo: E001 Gobierno del trabajo es una iniciativa.

## Markdown

Una forma simple de escribir texto con formato. Un `#` hace un título y `**así**` hace negrita.

Ejemplo: todas las páginas de esta wiki están en Markdown.

## Mermaid

Un lenguaje para dibujar diagramas con texto. VS Code lo convierte en dibujo.

Ejemplo: `flowchart TD` seguido de `A --> B` dibuja una flecha de A a B.

## orquestador

El agente principal. Habla con el humano, planifica y revisa el trabajo.

Ejemplo: cuando pides un cambio, el orquestador lo refina y lo parte en slices.

## pre-commit

Un hook de git que corre justo antes de guardar un commit. Aquí ejecuta el gate.

Ejemplo: si la wiki tiene un enlace roto, el pre-commit bloquea el commit.

## prueba en rojo

Una prueba escrita antes del código. Debe fallar, porque lo que prueba aún no existe.

Ejemplo: la prueba del enlace roto falló antes de programar el gate.

## prueba en verde

La misma prueba después de programar. Ahora debe pasar.

Ejemplo: después de programar el gate, la prueba del enlace roto pasó.

## sello

Una lista de huellas de cada archivo. Sirve para saber si algo cambió sin avisar.

Variantes: sellos

Ejemplo: `gobierno/config/sellos.json` guarda el sello del compilador.

## SHA-256

Una forma de sacar la huella de un archivo. Si cambia una letra, cambia la huella.

Ejemplo: dos archivos con la misma huella SHA-256 son iguales.

## slice

Una tarea pequeña que se prueba y se entrega sola. Es la parte más chica de una feature.

Variantes: slices

Ejemplo: S001 Esqueleto + gate es un slice de la feature F001.

## token

La unidad con que una IA mide el texto que lee y escribe. Cada token cuesta tiempo y dinero.

Variantes: tokens

Ejemplo: `CLAUDE.md` se carga en cada sesión, así que cada palabra de más gasta tokens siempre.

## wiki

Un conjunto de páginas enlazadas entre sí. Esta wiki explica el repositorio a las personas.

Ejemplo: esta página es parte de la wiki.
