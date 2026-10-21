# Comunicación entre agentes · v0.1
Aplica a orquestador → worker/auditor y vuelta. No aplica a mensajes al humano.

## Canal
- Entrada: ENCARGO en el prompt. Salida: CIERRE. Otro canal no existe.
- Sin contexto conversacional ni historial: el encargo trae todo. Falta dato → cierre `BLOQUEADO` + motivo. No adivinar.
- Registro caveman: sin artículos innecesarios, cortesía, transiciones ni anuncios de intención. Hechos, no narración.
- Sin analogías ni pedagogía.
- Único texto emitido: el CIERRE. Resto = llamadas a herramientas.
- Cada texto emitido se reenvía en cada turno siguiente → brevedad = menos gasto.
- Gate: hook `SubagentStop` → `aquas cierre-agente` rechaza cierre sin sus campos fijos.

## ENCARGO (orquestador → agente; campos fijos, en este orden)
```
id: E001/F003/S001-T1
contrato: docs/E001/F003/S001.md (o «—» si no hay slice)
objetivo: una línea
rutas: archivos a leer o tocar
incluye: …
excluye: …
prueba: comando que demuestra terminado
terminado: condición verificable
```

## CIERRE worker (primera línea = estado)
```
estado: HECHO | PARCIAL | BLOQUEADO
archivos: rutas tocadas | —
evidencia: comando → resultado, una línea cada uno
bloqueo: motivo | —
fuera_de_alcance: hallazgos no resueltos | —
```

## CIERRE auditor (primera línea = veredicto)
```
veredicto: APROBADO | CON_HALLAZGOS | RECHAZADO
hallazgos: qué · dónde (archivo:línea) · evidencia · efecto · certeza; grave → leve | —
no_revisado: qué y por qué | —
```

## Temporales
- Antes de invocar Go: TMP, TEMP y GOTMPDIR → `<raíz>/.tmp` (ruta absoluta).
- Archivos propios en `.tmp/<id-del-encargo>/`; borrarlos al cerrar. Nunca borrar `.tmp` entero: ahí corren los binarios de los hooks.
