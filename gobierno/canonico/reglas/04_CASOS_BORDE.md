# Casos borde · v0.1
Motivo: 5 defectos del 2026-10-23 eran previsibles; cada uno, una categoría que nadie buscó (D000031).
- Contrato de slice → tabla `**Casos borde.**` antes del rojo: `Categoría | Aplica | Prueba o motivo`.
- 7 categorías, nombre exacto, una fila cada una:
  - cada tipo de ficha;
  - cada estado de origen;
  - padre e hijo, activo y archivado;
  - entrada ausente, vacía, mal escrita o duplicada;
  - choque con otro control;
  - quién habla: humano, agente o herramienta;
  - fallo a mitad de camino.
- Aplica `sí` → nombrar prueba (`` `TestX` ``) que cubre el caso. `no` → por qué, sin celda vacía.
- Integral de la feature cubre casos entre slices; pruebas del slice, los suyos.
- Gate (`aquas verificar`) exige tabla completa en página de slice con ficha activa. Archivada: no se revisa.
- Auditor juzga calidad: caso real omitido, `no` sin razón válida, prueba que no cubre el caso → hallazgo.
- Defecto nuevo → causa → buscar el mismo hueco en otros lugares; cada otro caso al backlog.
