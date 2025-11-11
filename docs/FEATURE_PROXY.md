# Proxy Automático para MediaCheky

## Objetivo
Implementar un sistema de proxy/reverse-proxy automático para los servicios gestionados por MediaCheky, permitiendo la asignación de dominios y subdominios por servicio.

---

## Checklist de Implementación

### 1. Selección de tecnología de proxy
- [ ] Analizar opciones: Traefik, Nginx Proxy Manager, Caddy, etc.
- [ ] Elegir la mejor opción para integración automática con Docker y configuración dinámica.
- [ ] Documentar pros/contras y decisión final.

### 2. Configuración global de dominios
- [ ] Añadir apartado en la página de configuración global para definir uno o varios dominios (ej: localcheky, mediacheky.com).
- [ ] Guardar dominios en la configuración global (DB y/o .env).
- [ ] Validar formato de dominio y evitar duplicados.

### 3. Configuración de subdominios por servicio
- [ ] Añadir campo en la configuración de cada servicio para definir el subdominio (ej: radarr.localcheky).
- [ ] Si el campo está vacío, generar automáticamente subdominio: {servicio}.{dominio}.
- [ ] Permitir override manual del subdominio.
- [ ] Mostrar subdominio resultante en la UI.

### 4. Generación de reglas de proxy
- [ ] Generar reglas de proxy dinámicamente al habilitar/deshabilitar servicios.
- [ ] Apuntar subdominios al contenedor y puerto interno correspondiente.
- [ ] Para desarrollo local, usar dominio docker.internal.
- [ ] Actualizar reglas al cambiar configuración.

### 5. Integración con Docker Compose
- [ ] Añadir servicio de proxy al docker-compose si no existe.
- [ ] Montar configuración dinámica desde MediaCheky.
- [ ] Reiniciar proxy sólo si cambian reglas.

### 6. UI/UX
- [ ] En configuración global: sección para dominios.
- [ ] En configuración de servicio: campo para subdominio y vista del endpoint final.
- [ ] Validación y feedback con toast notifications.

### 7. Seguridad
- [ ] Validar entradas para evitar inyección de configuración.
- [ ] Documentar riesgos y recomendaciones.

### 8. Documentación y ejemplos
- [ ] Documentar el flujo completo en este archivo.
- [ ] Añadir ejemplos de configuración y reglas generadas.
- [ ] Incluir troubleshooting y preguntas frecuentes.

---

## Notas
- El proxy debe funcionar tanto en desarrollo local como en producción.
- La configuración debe ser persistente y fácil de modificar desde la UI.
- El sistema debe ser extensible para futuros servicios y dominios.

---

**Última actualización:** 2025-11-11
