# Resumen Ejecutivo - Proyecto KeeperCheky

## 🎯 Objetivo del Proyecto

Reescribir completamente la funcionalidad de **Janitorr v1.9.0** como una aplicación web moderna con interfaz gráfica, accesible desde el navegador, similar a Jellyseerr, Sonarr, Radarr y otras aplicaciones del ecosistema *arr.

## 📋 Funcionalidad Principal (Janitorr)

### ¿Qué hace Janitorr?

Janitorr es una herramienta de gestión automatizada para bibliotecas de medios que:

1. **Limpieza Automática de Media**
   - Elimina películas/series antiguas según edad y espacio en disco
   - Configuración por porcentajes de espacio libre
   - Diferentes tiempos de expiración según disponibilidad

2. **Gestión Basada en Tags**
   - Limpieza personalizada usando tags de Sonarr/Radarr
   - Exclusión de contenido con tags específicos
   - Schedules personalizados por tag

3. **Limpieza de Episodios**
   - Manejo especial para series semanales/diarias
   - Mantener últimos N episodios
   - Eliminar episodios por edad

4. **Colecciones "Leaving Soon"**
   - Muestra en Jellyfin/Emby el contenido próximo a eliminar
   - Preview de 14 días antes de eliminación
   - Da oportunidad de ver contenido antes de que desaparezca

5. **Integración con Servicios**
   - Radarr (películas)
   - Sonarr (series)
   - Jellyfin/Emby (servidor de media)
   - Jellyseerr (requests)
   - Jellystat/Streamystats (estadísticas de visualización)
   - Bazarr (subtítulos)

## 📊 Análisis del Código Original

### Tecnologías Actuales (Janitorr v1.9.0)
- **Lenguaje**: Kotlin
- **Framework**: Spring Boot 3.5.6
- **Paradigma**: Scheduled jobs + REST clients
- **Deployment**: Docker (imagen JVM ~256MB RAM mínimo)
- **UI**: Ninguna (solo configuración YAML)

### Arquitectura
```
┌─────────────────────────────────────┐
│   Janitorr (Spring Boot Kotlin)     │
│                                     │
│  ┌─────────────────────────────┐   │
│  │  Cleanup Schedules          │   │
│  │  - MediaCleanupSchedule     │   │
│  │  - TagBasedCleanupSchedule  │   │
│  │  - WeeklyEpisodeCleanup     │   │
│  └─────────────────────────────┘   │
│                                     │
│  ┌─────────────────────────────┐   │
│  │  Service Clients            │   │
│  │  - RadarrClient             │   │
│  │  - SonarrClient             │   │
│  │  - JellyfinClient           │   │
│  │  - JellyseerrClient         │   │
│  │  - JellystatClient          │   │
│  └─────────────────────────────┘   │
│                                     │
│  ┌─────────────────────────────┐   │
│  │  Configuration (YAML)       │   │
│  │  - application.yml          │   │
│  └─────────────────────────────┘   │
└─────────────────────────────────────┘
          │         │         │
          ▼         ▼         ▼
    ┌─────────┐ ┌──────────┐ ┌─────────┐
    │ Radarr  │ │ Sonarr   │ │Jellyfin │
    └─────────┘ └──────────┘ └─────────┘
```

## 🎨 Visión del Nuevo Proyecto

### Transformación Propuesta

**De**: Aplicación CLI/backend sin interfaz  
**A**: Aplicación web completa con UI moderna

### Características Nuevas

1. **Dashboard Interactivo**
   - Visualización en tiempo real de estadísticas
   - Gráficos de uso de disco
   - Estado de servicios
   - Próximas eliminaciones

2. **Gestión de Media Visual**
   - Ver toda tu biblioteca con posters
   - Filtros y búsqueda avanzada
   - Excluir/eliminar con un click
   - Ver detalles de cada item

3. **Configuración de Schedules**
   - Crear/editar/eliminar schedules desde la UI
   - Preview de qué se eliminará
   - Ejecutar manualmente
   - Activar/desactivar schedules

4. **Settings Configurables**
   - Formularios para cada servicio
   - Test de conexión en tiempo real
   - Validación de configuración
   - No más editar YAML manualmente

5. **Logs en Vivo**
   - Ver logs en tiempo real
   - Filtrar por nivel (INFO, ERROR, etc.)
   - Búsqueda de texto
   - Export de logs

## 📦 Propuestas Desarrolladas

Se han creado **4 propuestas completas** con diferentes stacks tecnológicos:

### 1️⃣ TypeScript Full-Stack (Next.js + NestJS)
- **Mejor UX**: Interfaz moderna tipo SPA
- **Escalable**: Arquitectura de microservicios
- **Recursos**: 512MB-1GB RAM
- **Desarrollo**: 4-6 semanas

### 2️⃣ Python/HTMX (FastAPI + HTMX)
- **Simple**: Un solo lenguaje
- **Ligero**: 50-150MB RAM
- **Rápido**: 3-4 semanas desarrollo
- **Familiar**: Python para muchos devs

### 3️⃣ Go/Alpine.js (Fiber + Alpine)
- **Performance**: Ultra rápido
- **Mínimo**: 20-50MB RAM, imagen de 15-25MB
- **Eficiente**: Ideal para hardware limitado
- **Desarrollo**: 3-4 semanas

### 4️⃣ Rust/Leptos (Axum + WASM)
- **Seguridad**: Memory-safe
- **Performance**: Comparable a C++
- **Complejo**: Curva de aprendizaje alta
- **Desarrollo**: 5-7 semanas

## 🏆 Recomendación Principal

### **Propuesta 3: Go + Alpine.js** 

**Justificación:**

1. **Balance perfecto**: Performance, recursos, desarrollo
2. **Herencia de Janitorr**: Mantiene filosofía ligera
3. **Deployment simple**: Binario único, fácil distribución
4. **Costo operacional**: Mínimo, ideal para self-hosting
5. **Escalabilidad**: Si crece, Go escala fácilmente

**Comparación con Janitorr original:**
- Janitorr: 256MB RAM (JVM) → KeeperCheky: 20-50MB RAM (Go)
- Janitorr: Sin UI → KeeperCheky: UI moderna
- Janitorr: Config YAML → KeeperCheky: UI de configuración
- Janitorr: Solo logs → KeeperCheky: Dashboard + Logs + Gestión

## 🗂️ Documentación Entregada

### Archivos Creados

```
keepercheky/
├── docs/
│   ├── propuestas/
│   │   ├── PROPUESTA_1_STACK_MODERNO.md     (TypeScript)
│   │   ├── PROPUESTA_2_PYTHON_HTMX.md       (Python/HTMX)
│   │   ├── PROPUESTA_3_GO_ALPINE.md         (Go/Alpine) ⭐
│   │   └── PROPUESTA_4_RUST_LEPTOS.md       (Rust/Leptos)
│   └── COMPARACION_Y_RECOMENDACIONES.md     (Este doc comparativo)
│
└── reference-repos/
    └── janitorr/                             (Código original v1.9.0)
```

### Contenido de Cada Propuesta

Cada documento incluye:

1. **Visión General**: Filosofía y objetivos
2. **Stack Tecnológico**: Detalle de tecnologías
3. **Estructura del Proyecto**: Árbol de archivos completo
4. **Interfaz de Usuario**: Wireframes/código de cada página
5. **Backend**: Arquitectura y código de ejemplo
6. **Docker**: Dockerfiles y docker-compose
7. **Ventajas/Desventajas**: Análisis honesto
8. **Estimación de Recursos**: RAM, CPU, disco
9. **Roadmap**: Plan de implementación por fases

## 📈 Próximos Pasos Recomendados

### Fase de Decisión (Esta semana)

1. ✅ **Revisar propuestas** (completado)
2. ⬜ **Decidir stack** según prioridades
3. ⬜ **Validar requisitos** de infraestructura

### Fase de Setup (Semana 1)

1. ⬜ Crear repositorio GitHub
2. ⬜ Setup estructura de proyecto
3. ⬜ Configurar Docker development
4. ⬜ CI/CD básico

### Fase de Desarrollo (Semanas 2-4)

1. ⬜ Core backend + modelos
2. ⬜ Clientes de servicios
3. ⬜ Lógica de cleanup
4. ⬜ UI base + páginas principales

### Fase de Testing (Semana 5)

1. ⬜ Testing en modo dry-run
2. ⬜ Pruebas con servicios reales
3. ⬜ Bug fixes
4. ⬜ Documentación

### Fase de Release (Semana 6)

1. ⬜ Docker optimization
2. ⬜ Documentación completa
3. ⬜ Release 1.0.0
4. ⬜ Docker Hub publish

## 💡 Consideraciones Importantes

### Compatibilidad

- **Configuración**: Considerar parser para `application.yml` de Janitorr
- **Migración**: Script para usuarios existentes
- **APIs**: Mantener compatibilidad con servicios externos

### Seguridad

- **Dry-run por defecto**: Evitar borrados accidentales
- **Confirmaciones**: Para operaciones destructivas
- **Logs**: Registro completo de todas las acciones
- **Backups**: Metadata antes de eliminar

### Escalabilidad

- **Múltiples instancias**: Considerar para futuro
- **Rate limiting**: Para APIs externas
- **Caching**: Para reducir llamadas a servicios
- **Queue**: Para operaciones en background

## 📞 Soporte y Contacto

### Recursos Disponibles

- **Documentación**: 4 propuestas detalladas + comparación
- **Código de referencia**: Janitorr v1.9.0 descargado
- **Ejemplos**: Código de ejemplo en cada propuesta

### Preguntas Frecuentes

**P: ¿Puedo cambiar de stack después?**  
R: Sí, pero implica reescribir. Mejor decidir bien al inicio.

**P: ¿Funcionará con mis servicios actuales?**  
R: Sí, usa las mismas APIs que Janitorr original.

**P: ¿Puedo usar parte de una propuesta con otra?**  
R: Sí, son modulares. Por ejemplo, backend de Go con frontend de Next.js.

**P: ¿Hay un demo?**  
R: No aún, pero cada propuesta tiene código de ejemplo completo.

## 🎓 Aprendizajes del Análisis

### De Janitorr Original

**Buenas prácticas identificadas:**
- Modo dry-run por defecto
- Exclusión por tags
- Leaving Soon collections
- Múltiples estrategias de limpieza
- Configuración flexible

**Áreas de mejora:**
- Falta de UI (principal motivación)
- Configuración solo en YAML
- Sin preview de qué se eliminará
- Logs solo en archivos
- Sin estadísticas visuales

### Lecciones para KeeperCheky

1. **UX First**: Dashboard y visualización son clave
2. **Safety**: Múltiples confirmaciones y dry-run
3. **Flexibilidad**: Diferentes estrategias de cleanup
4. **Integración**: Soporte para múltiples servicios
5. **Logs**: Visibilidad completa de acciones

## ✨ Valor Agregado

### Vs Janitorr Original

| Característica | Janitorr | KeeperCheky |
|----------------|----------|-------------|
| Interfaz Web | ❌ | ✅ |
| Dashboard | ❌ | ✅ |
| Gestión Visual | ❌ | ✅ |
| Config UI | ❌ | ✅ |
| Preview Eliminaciones | ❌ | ✅ |
| Logs en Vivo | ❌ | ✅ |
| Ejecución Manual | ❌ | ✅ |
| Uso de RAM | 256MB+ | 20-500MB* |
| Tamaño Imagen | ~300MB | 15-500MB* |

*Dependiendo del stack elegido

### Vs Maintainerr (Competencia)

Maintainerr es solo para Plex. KeeperCheky será para Jellyfin/Emby con features similares o mejores.

## 🚀 Conclusión

Se ha completado un **análisis exhaustivo** del proyecto Janitorr y se han desarrollado **4 propuestas técnicas detalladas** para reescribir la funcionalidad con interfaz web moderna.

**Recomendación final**: Proceder con **Propuesta 3 (Go + Alpine.js)** por su óptimo balance entre rendimiento, recursos y tiempo de desarrollo.

El proyecto está listo para comenzar la fase de implementación.

---

**Documento generado**: 25 de Octubre de 2025  
**Versión**: 1.0  
**Estado**: Propuestas completadas, pendiente decisión de stack
