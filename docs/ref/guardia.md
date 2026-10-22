# Guardia: nada se escribe fuera del repo

Referencia de la regla de oro número 5 del [orquestador](../glosario.md#orquestador). Vive en `internal/guardia/` y corre como [hook](../glosario.md#hook) antes de cada herramienta.

## En una frase

Antes de que el [agente](../glosario.md#agente) escriba un archivo o corra un comando, un programa revisa el destino. Si queda fuera del repositorio, lo bloquea.

## Para qué existe

**El problema.** Un agente puede escribir en cualquier carpeta de tu computadora. Una prueba mal hecha puede dejar archivos en tu perfil o pisar tu configuración personal.

**La solución.** Una regla escrita no basta, porque depende de que el agente se acuerde. Por eso es un [gate](../glosario.md#gate): un programa en [Go](../glosario.md#go) que mira cada escritura antes de que ocurra.

Qué revisa:

| Herramienta | Qué mira |
|---|---|
| Write, Edit, NotebookEdit | La ruta del archivo |
| Bash | Redirecciones `>` y `>>`, `cd`, y los destinos de `rm`, `mv`, `cp`, `mkdir`, `touch`, `tee`, `sed -i`, `git clone`, `curl -o` y otros |

Qué deja pasar: leer fuera del repo, `/dev/null` y `2>&1`.

**Lista de permitidos.** Las únicas carpetas de fuera donde sí se puede escribir están en `gobierno/config/guardia.json`. Cada una lleva su motivo. Si una carpeta no está en la lista, la guardia la bloquea.

| Carpeta permitida | Motivo |
|---|---|
| `${LOCALAPPDATA}/go-build` | Caché de compilación de Go en Windows |
| `~/.cache/go-build` | Caché de compilación de Go en Linux |

El gate `aquas verificar` revisa la lista. Falla si el archivo está roto, si falta un motivo o si una ruta permite todo un disco.

**Entorno declarado.** Antes de revisar nada, la guardia mira `gobierno/config/entorno.local.json`. Ese archivo dice el sistema operativo y la raíz del repo en esta máquina. Lo escribe `aquas iniciar` y no se versiona.

- Si el archivo falta, bloquea todo.
- Si el sistema o la raíz no coinciden con la realidad, también bloquea todo. Por ejemplo, si el repo se movió de carpeta.
- Con ese archivo, `aquas compilar` genera `.claude/settings.json` desde `gobierno/config/claude_settings.json`. El shell y las herramientas vigiladas salen del sistema operativo.

**PowerShell.** En Windows la guardia también vigila PowerShell. No sabe leer sus comandos, así que los bloquea siempre. Para escribir, usa Bash.

## Analogía

Piensa en el **portero de un edificio**.

- Puedes mirar por la ventana todo lo que quieras.
- Pero si intentas sacar algo por la puerta, el portero revisa a dónde va.
- Si no puede leer la dirección del paquete, no lo deja salir.

Límite de la analogía: el portero solo ve lo que pasa por la puerta. Si un programa escribe por su cuenta, por ejemplo un script de Python, la guardia no lo ve.

## Diagrama

Qué pasa con cada herramienta.

```mermaid
sequenceDiagram
    participant A as Agente
    participant C as Claude Code
    participant G as aquas guardia
    A->>C: quiere usar Write o Bash
    C->>G: hook PreToolUse con la ruta o el comando
    alt destino dentro del repo
        G-->>C: salida 0, sigue
        C->>A: la herramienta corre
    else destino fuera o ruta no resoluble
        G-->>C: salida 2 y el motivo
        C->>A: bloqueado, con el motivo
    end
```

## Ejemplo de punta a punta

<a id="EJ-REF-guardia"></a>

**Situación.** Jesús pidió: «que no puedas escribir nada fuera de este repositorio».

1. Primero se escribieron las pruebas. Hay 18 comandos que deben fallar, como `echo x > ~/nota.txt`.
2. También hay 11 comandos que deben pasar, como `cat ~/.claude/settings.json`, que solo lee.
3. Para confirmar que las pruebas vigilan, se rompió el programa a propósito. Fallaron 9 casos.
4. Al conectar el hook apareció un error escondido. `go run` cambia la salida 2 por 1, y la salida 1 no bloquea.
5. La solución fue `.claude/hooks/aquas.sh`. Compila el programa en `.tmp/` y lo corre directo, así la salida 2 llega intacta.
6. Ese mismo script guarda los temporales de Go en `.tmp/`. Antes, el hook escribía fuera del repo sin que nadie lo notara.

## Paso a paso

Así decide la guardia con un comando de Bash:

1. Quita el texto de los heredocs, porque es contenido y no comandos.
2. Corta el comando en partes, en cada `&&`, `;` o `|`.
3. Sigue cada `cd` para saber en qué carpeta está.
4. Busca los destinos de escritura de cada parte.
5. Si un destino usa una variable que no conoce, bloquea: no puede probar que esté dentro.
6. Si algún destino queda fuera de la raíz del repo, bloquea con el motivo.

## Cómo se usa

- Con un [slice](../glosario.md#slice) en curso, también frena lo que cae fuera de su alcance declarado (ver E001/F002/S007).
- No hay que llamarla: corre sola antes de cada herramienta.
- Clon nuevo: corre `go run ./cmd/aquas iniciar` en la raíz. Sin ese paso no hay hooks y el [pre-commit](../glosario.md#pre-commit) falla.
- Si te bloquea, el mensaje dice la ruta. Usa una ruta dentro del repo, por ejemplo `.tmp/`.
- Probar: `go test ./internal/guardia/`.
- Agregar una excepción: añade una entrada con `ruta` y `motivo` en `gobierno/config/guardia.json`. Límite: hoy nada impide que el agente edite esa lista. Revisar ese cambio en el [commit](../glosario.md#commit) es criterio humano.
- Ojo: la caché la escribe Go por su cuenta, y la guardia no ve esa escritura. La lista deja la excepción por escrito y permite los comandos explícitos, como `go clean -cache`.
- Límites: no ve programas que escriben por su cuenta, ni `git -C`, `tar -C` o enlaces que salen del repo. Eso es criterio humano.
- Código: [guardia.go](../../internal/guardia/guardia.go), [permitidos.go](../../internal/guardia/permitidos.go) y [aquas.sh](../../.claude/hooks/aquas.sh).

## Navegación

- ↑ Subir: [Inicio](../README.md)
