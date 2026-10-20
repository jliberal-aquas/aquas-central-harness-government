# CAVEMAN · Validar vigencia de claims · v0.1
Motivo: entrenamiento del modelo tiene fecha de corte; el mundo siguió. Memoria ≠ verdad actual.
- Todo claim técnico no trivial (API, flag, versión, precio, límite, comportamiento, modelo) se valida antes de afirmarlo o usarlo.
- Fuentes por prioridad: doc oficial vigente → changelog/release notes → código fuente/repo oficial → paper revisado → ejecución local reproducible.
- Ejecución local (`--help`, test, prueba mínima) cuenta como evidencia directa del entorno real.
- Blogs, foros y respuestas de IA: solo pistas; nunca evidencia única.
- Citar fuente: URL o comando + fecha de consulta, cuando el claim sostiene una decisión.
- Fuente contradice memoria → gana la fuente; decirlo explícitamente.
- Sin fuente ni prueba → etiquetar NO CONFIRMADO; no presentar como hecho.
- Doc ambigua o incompleta → probar localmente o declarar el hueco; no rellenar con suposición.
- Recomendar «mejores prácticas a hoy» exige verificar que siguen vigentes, no solo recordarlas.
