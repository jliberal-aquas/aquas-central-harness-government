# Hallazgos: insights y fallos agrupados por tema

Registrados como defectos el 2026-10-23: T01–T15 → D000005–D000019 (mismo orden). Fuente: 86 MB de historiales de sesión, filtrados a 232 líneas; agrupación hecha por un worker en Haiku y revisada en T01–T20.

Orden: primero PROPUESTO sin defecto y repetido; luego PROPUESTO sin defecto; luego el resto. Número de línea = línea de crudo.md.

## T01 · Presupuesto: pedido nuevo cuenta contra el plan viejo
- fechas: 2026-10-21 → 2026-10-23
- estado: PROPUESTO (línea 110, 2026-10-22). Líneas de 2026-10-23 sin estado.
- defecto: no. D000003 cubre solo el caso tras /compact.
- repetido: sí (22 líneas, 3 fechas)
- líneas: 90,91,94,96,101,102,107,109,110,111,114,165,167,168,170,172,174,175,179,185,191,193. Plan cumplido sin fallo: 172,174,175,179,185.
- cita: «cerrar un slice no cierra su plan» (110); «el control de presupuesto me frenó» (109)

## T02 · Guardia no bloquea rutas sin resolver
- fechas: 2026-10-21
- estado: PROPUESTO (línea 5). Línea 6 dice «Sigue sin proteger» sin estado.
- defecto: no
- repetido: sí (3 líneas: 4, 5, 6)
- líneas: 4,5,6
- cita: «la guardia no bloqueó» (4); «probar la regla por separado no prueba que el hook esté activo» (5)

## T03 · Guardia bloquea escrituras legítimas (falsos positivos)
- fechas: 2026-10-21 → 2026-10-22
- estado: PROPUESTO (líneas 153, 158, 166; última: 166). Controles ya probados: 17, 119, 146 (IMPLEMENTADO Y PROBADO).
- defecto: no
- repetido: sí (10 líneas)
- líneas: 17,77,119,120,146,147,153,158,159,166
- cita: «el guardia lee como rutas algunos argumentos que no lo son» (158); «la guardia revisa el alcance antes de ejecutar el comando completo» (166)

## T04 · Permisos y temporales (TMP, rm -rf, bloqueos de herramienta)
- fechas: 2026-10-20 → 2026-10-22
- estado: PROPUESTO (línea 23). Resto sin estado.
- defecto: no
- repetido: sí (6 líneas)
- líneas: 3,23,35,135,152,223
- cita: «`TMP=./.tmp` sirve para `go run`, pero rompe `go test`» (23)

## T05 · Hooks por palabra clave confunden un tema con un fallo
- fechas: 2026-10-20 → 2026-10-22
- estado: PROPUESTO (línea 143, 2026-10-22)
- defecto: no
- repetido: sí (7 líneas)
- líneas: 117,142,143,144,203,204,205
- cita: «un control por palabra clave confunde un tema con un incidente» (143); «detectó la palabra «fallas»» (203)

## T06 · Cierre de slice: memoria antes del commit y archivo
- fechas: 2026-10-21 → 2026-10-22
- estado: PROPUESTO (línea 54). Línea 160 dice «ya quedó corregido» sin estado de control. El hook de cierre lo detecta; no lo previene.
- defecto: no
- repetido: sí (9 líneas)
- líneas: 54,55,71,73,75,76,98,99,160
- cita: «actualizar la memoria antes de un commit la deja vieja» (54); «no actualicé la memoria después del último commit» (160)

## T07 · Foco duplicado entre memoria y backlog
- fechas: 2026-10-21
- estado: mixto. Línea 59: «punto 6 IMPLEMENTADO SIN PROBAR», «A PROPUESTO». Línea 57: PROPUESTO.
- defecto: no
- repetido: sí (4 líneas)
- líneas: 57,58,59,60
- cita: «el foco duplicado entre la memoria y el backlog» (57); «la memoria apunta al backlog viejo desde S002» (60)

## T08 · Configuración escrita a mano y compilados sin gate
- fechas: 2026-10-21 → 2026-10-22
- estado: PROPUESTO (línea 140, 2026-10-22). Línea 13: IMPLEMENTADO Y PROBADO para settings.json. Línea 140 trata la herencia de CLAUDE.md, otro caso.
- defecto: no
- repetido: sí (4 líneas)
- líneas: 11,12,13,140
- cita: «un archivo de configuración escrito a mano no pasa por ningún gate» (11)

## T09 · Prosa y comandos citados se quedan viejos
- fechas: 2026-10-22
- estado: PROPUESTO (línea 151, última). Línea 149: IMPLEMENTADO Y PROBADO para citas de comandos.
- defecto: no
- repetido: sí (4 líneas)
- líneas: 148,149,150,151
- cita: «la prosa se queda vieja sin que el gate lo note» (149); «marcó `aquas install landing`» (151)

## T10 · Sello de memoria: falso positivo al cambiar campos
- fechas: 2026-10-21
- estado: PROPUESTO (línea 27). Evidencia pendiente de verde.
- defecto: no
- repetido: sí (3 líneas)
- líneas: 26,27,28
- cita: «un sello calculado sobre la estructura del programa se rompe con cada campo nuevo» (27)

## T11 · CRLF al escribir archivos con Python en Windows
- fechas: 2026-10-21
- estado: PROPUESTO (línea 32). Corregido a mano en dos archivos; sin control automático.
- defecto: no
- repetido: sí (2 líneas)
- líneas: 32,33
- cita: «en Windows, Python en modo texto escribe CRLF» (32)

## T12 · Aprobaciones del humano que frenan sin aportar
- fechas: 2026-10-20
- estado: PROPUESTO (línea 201)
- defecto: no
- repetido: sí (3 líneas)
- líneas: 200,201,202
- cita: «el freno crece sin que nadie lo note» (201)

## T13 · Cambios no pedidos o no revisados contra fuentes
- fechas: 2026-10-20 → 2026-10-22
- estado: último estado explícito PROPUESTO (línea 208, 2026-10-20). Línea 118 (2026-10-22) sin estado.
- defecto: no
- repetido: sí (8 líneas)
- líneas: 118,206,207,208,209,226,227,228
- cita: «una regla que depende de que el agente se acuerde no es un control» (208)

## T14 · Compactación: decisiones que se pierden
- fechas: 2026-10-20 → 2026-10-21
- estado: PROPUESTO (línea 31, 2026-10-21). IMPLEMENTADO Y PROBADO (línea 217, 2026-10-20) para guardar decisiones en archivo.
- defecto: no. D000003 (plan perdido tras /compact) se trata en T01.
- repetido: sí (2 líneas)
- líneas: 31,217
- cita: «Lo que solo vive en el chat se borra al compactar» (217)

## T15 · Deuda por excepciones y sello en lote
- fechas: 2026-10-23
- estado: PROPUESTO (línea 187, última; 184 y 176 también). IMPLEMENTADO Y PROBADO para el contador de deuda (línea 178).
- defecto: no
- repetido: sí (7 líneas)
- líneas: 176,177,178,184,186,187,188
- cita: «una regla con excepción para lo viejo necesita un contador visible» (178)

## T16 · Gate revisa la carpeta de trabajo, no lo que entra al commit
- fechas: 2026-10-20
- estado: PROPUESTO (línea 2)
- defecto: no
- repetido: no (1 línea)
- líneas: 2
- cita: «el gate revisa los archivos en la carpeta de trabajo, no lo que entra al commit» (2)

## T17 · Slices de arreglo sin ID propio
- fechas: 2026-10-22
- estado: PROPUESTO (línea 164)
- defecto: no
- repetido: no (1 línea)
- líneas: 164
- cita: ««slice» se estiraba para cubrir arreglos» (164)

## T18 · Insight PROPUESTO sin tarea registrada
- fechas: 2026-10-23. Línea 198 sin fecha.
- estado: PROPUESTO (línea 194)
- defecto: no
- repetido: no (2 líneas, mismo tema)
- líneas: 194,198
- cita: «un insight «PROPUESTO» sin tarea registrada no protege de nada» (194)

## T19 · Medición con hooks que capturan literales
- fechas: 2026-10-20
- estado: PROPUESTO (línea 199)
- defecto: no
- repetido: no (1 línea)
- líneas: 199
- cita: «si el literal lo captura un hook, ningún refinado del agente puede contaminar tu medición futura» (199)

## T20 · Diagramas: la regla y el hook de cierre
- fechas: 2026-10-20
- estado: PROPUESTO (línea 214)
- defecto: no
- repetido: no (1 línea)
- líneas: 214
- cita: «el problema no está en cómo se muestra el diagrama» (214)

## T21 · Plan de llamadas que no se registra
- fechas: 2026-10-21 → 2026-10-23
- estado: PROPUESTO (líneas 181 y 192, 2026-10-23). IMPLEMENTADO SIN PROBAR para el arreglo de conteo (línea 115).
- defecto: D000002 (plan-sin-linea): líneas 123, 124, 125. D000003 (plan-perdido-tras-compact), parcial: líneas 192 y 195. Líneas 180 y 182 («0 agentes») no están cubiertas por D000001–D000004.
- repetido: sí (15 líneas)
- líneas: 92,93,97,112,113,115,116,123,124,125,180,181,182,192,195
- cita: «el plan se registra con mi línea «Plan: N llamadas»» (124); «solo registra un plan cuando escribes la palabra «Apruebo»» (112)

## T22 · Auditor no disponible a mitad de sesión
- fechas: 2026-10-21
- estado: IMPLEMENTADO Y PROBADO (línea 45). Línea 41: PROPUESTO. Línea 43: corregido.
- defecto: no
- repetido: sí (6 líneas)
- líneas: 40,41,42,43,44,45
- cita: «Agent type 'auditor' not found» (40); «el auditor detecta lo que no detecta el gate» (45)

## T23 · Costo de subagentes y medición de ahorro
- fechas: 2026-10-21
- estado: IMPLEMENTADO Y PROBADO (línea 52, script temporal en .tmp, sin control permanente). Líneas 46, 47, 49, 50: PROPUESTO.
- defecto: no
- repetido: sí (8 líneas)
- líneas: 46,47,48,49,50,51,52,53
- cita: «cada subagente tiene un costo fijo de unos 50.000 tokens» (46)

## T24 · Vueltas: la wiki causa casi todas las vueltas del gate
- fechas: 2026-10-22
- estado: IMPLEMENTADO Y PROBADO (línea 121)
- defecto: no
- repetido: no (1 línea)
- líneas: 121
- cita: «en este repo la wiki causa casi todas las vueltas del gate, no el código» (121)

## T25 · Escapes de barras y perl (edición que falla o rompe)
- fechas: 2026-10-21 → 2026-10-22
- estado: IMPLEMENTADO SIN PROBAR (línea 173, última). PROPUESTO: líneas 128, 130, 169.
- defecto: no
- repetido: sí (10 líneas)
- líneas: 29,30,86,122,128,129,130,131,169,173
- cita: «`perl -0pi` falla en silencio cuando el patrón trae caracteres» (173)

## T26 · Archivar vacía el foco de la sesión siguiente
- fechas: 2026-10-21 → 2026-10-23
- estado: IMPLEMENTADO Y PROBADO (línea 36). Línea 34: PROPUESTO. Línea 197 sin estado.
- defecto: no
- repetido: sí (3 líneas)
- líneas: 34,36,197
- cita: «vaciar «siguiente» al archivar borra justo lo que necesita la sesión nueva» (34)

## T27 · Backlog permite ejecutar sin definir y bloqueos cancelados
- fechas: 2026-10-21 → 2026-10-22
- estado: IMPLEMENTADO Y PROBADO (líneas 56 y 106). Línea 104: PROPUESTO (superada por 106).
- defecto: no
- repetido: sí (6 líneas)
- líneas: 56,103,104,105,106,108
- cita: «`aquas backlog siguiente` entregó un slice P2» (103)

## T28 · Wiki desfasada y navegación por niveles
- fechas: 2026-10-20 → 2026-10-23
- estado: IMPLEMENTADO SIN PROBAR (línea 183, 2026-10-23). PROPUESTO (línea 216, 2026-10-20). Línea 81: «Sigue sin arreglar» sin estado.
- defecto: no
- repetido: sí (9 líneas)
- líneas: 80,81,82,84,183,216,230,231,232
- cita: «la wiki no sigue tu regla de navegación por niveles» (80). Línea 82 trata una regla del humano; vecina, no confirmada como wiki.

## T29 · Glosario y enlaces: el gate frena páginas
- fechas: 2026-10-21 → 2026-10-23
- estado: IMPLEMENTADO Y PROBADO (líneas 126, 132, 189)
- defecto: no
- repetido: sí (11 líneas)
- líneas: 16,18,19,20,95,100,126,127,132,133,189
- cita: «el gate exige enlazar «agente» al glosario en su primera aparición» (126)

## T30 · Diagnóstico y encargos incompletos de la sesión previa
- fechas: 2026-10-21
- estado: sin estado. Línea 8 dice «ambas cosas están corregidas».
- defecto: no
- repetido: sí (4 líneas)
- líneas: 8,9,10,85
- cita: «La sesión anterior diagnosticó mal la guardia» (8)

## T31 · Canónicos: tope de caracteres y Caveman
- fechas: 2026-10-20 → 2026-10-21
- estado: IMPLEMENTADO como regla escrita, criterio humano sin gate (línea 38). Línea 37: aplicado sin control automático.
- defecto: no
- repetido: sí (9 líneas)
- líneas: 14,37,38,39,210,211,212,213,215
- cita: «un tope ajustado empujaba a recortar reglas y se perdía sentido» (38)

## T32 · Pruebas débiles y rojos sin motivo verificado
- fechas: 2026-10-21 → 2026-10-23
- estado: IMPLEMENTADO Y PROBADO (línea 196, 2026-10-23). PROPUESTO: líneas 15, 65, 74. IMPLEMENTADO SIN PROBAR: línea 25.
- defecto: no
- repetido: sí (22 líneas)
- líneas: 15,25,61,62,63,64,65,66,67,68,69,72,74,78,83,134,138,139,141,154,196
- cita: «11 de 14 errores metidos a propósito sobrevivieron» (65)

## T33 · Revisión independiente y hooks que fallan abiertos
- fechas: 2026-10-22
- estado: IMPLEMENTADO Y PROBADO (línea 145, fallo cerrado con última versión que compiló). PROPUESTO: línea 136.
- defecto: no
- repetido: sí (3 líneas)
- líneas: 136,137,145
- cita: «el más serio, los hooks que fallan abiertos» (136)

## T34 · Cierre de slice sobre gate fallido
- fechas: 2026-10-22
- estado: IMPLEMENTADO SIN PROBAR (líneas 156 y 171)
- defecto: no
- repetido: sí (4 líneas)
- líneas: 155,156,157,171
- cita: «`comando | tail` esconde un fallo» (171)

## T35 · Comunicación: certeza, resúmenes y términos sin explicar
- fechas: 2026-10-21 → 2026-10-22
- estado: IMPLEMENTADO SIN PROBAR (línea 162, regla escrita, aplicada a mano). Línea 21: IMPLEMENTADO Y PROBADO para etiquetas de certeza.
- defecto: no
- repetido: sí (6 líneas)
- líneas: 21,22,161,162,163,224
- cita: «un informe largo sin marcas de certeza mezcla lo probado con lo supuesto» (21)

## T36 · Herramientas: Python, terminal y comandos sin comillas
- fechas: 2026-10-21 → 2026-10-23
- estado: sin estado
- defecto: no
- repetido: sí (3 líneas)
- líneas: 79,190,229
- cita: «Una prueba mía armaba un comando sin comillas en una ruta con espacio» (229)

## T37 · Turno cortado sin causa verificada
- fechas: 2026-10-21
- estado: sin estado. Causa NO CONFIRMADA.
- defecto: no
- repetido: no (1 línea)
- líneas: 88
- cita: «El turno anterior se cortó antes de responder» (88)

## T38 · Go run: código de salida que llega como error
- fechas: 2026-10-21
- estado: corregido, sin probar en sesión viva (línea 225)
- defecto: D000001 (hook-exit-go-run). PROBABLE: la línea 225 menciona «el defecto de `go run`» sin nombrar el código de salida.
- repetido: no (1 línea)
- líneas: 225
- cita: «el defecto de `go run` arriba, ya corregido pero sin probar en una sesión viva» (225)

## T39 · Ejemplo de FALLO en docs de comunicación (falso aviso)
- fechas: 2026-10-20 → 2026-10-21
- estado: corregido (línea 222: «ya corregido»)
- defecto: no
- repetido: sí (3 líneas)
- líneas: 219,221,222
- cita: «el falso aviso de «Go», contado como mensaje de FALLO» (221)

## T40 · descartados (sin aprendizaje)
- líneas: 1,7,24,70,87,89,218,220
- motivo: sin contenido o solo cabecera. Línea 1: «No.» Líneas 7, 218, 220: vacías. Líneas 24, 70, 87, 89: cabecera sin cuerpo.
