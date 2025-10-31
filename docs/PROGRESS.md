# 📊 Progreso de Desarrollo - KeeperCheky

**Última actualización**: 25 de Octubre de 2025

## ✅ Completado

### Sprint 0: Infraestructura y Configuración
- [x] Estructura de proyecto según arquitectura Go estándar
- [x] Configuración de Docker (desarrollo y producción)
- [x] Configuración de Air para hot-reload
- [x] Actualización a Go 1.25.3
- [x] Integración de Fiber v2.52.9 (con fix de seguridad)
- [x] Actualización de dependencias (golang.org/x/crypto, golang.org/x/net)
- [x] Servidor HTTP funcionando en localhost:8000
- [x] Dashboard básico con Alpine.js y Tailwind CSS
- [x] Endpoints de API iniciales (/api/health, /api/stats)
- [x] Sistema de logging con Zap
- [x] Base de datos SQLite con GORM

### Sprint 1: Clientes de Servicios Externos ✨ **COMPLETADO**
- [x] **Interfaz MediaClient** - Radarr/Sonarr
  - Métodos: TestConnection, GetLibrary, GetItem, DeleteItem, GetTags
- [x] **Interfaz StreamingClient** - Jellyfin
  - Métodos: TestConnection, GetLibrary, GetPlaybackInfo, DeleteItem
- [x] **Interfaz RequestClient** - Jellyseerr
  - Métodos: TestConnection, GetRequests, GetRequest, DeleteRequest
- [x] **RadarrClient** - Cliente completo para Radarr API v3
  - Conversión de películas a modelo interno Media
  - Extracción de posters, tags, calidad
  - Sistema de retry con backoff exponencial
- [x] **SonarrClient** - Cliente completo para Sonarr API v3
  - Conversión de series a modelo interno Media
  - Estadísticas de episodios
  - Sistema de retry con backoff exponencial
- [x] **JellyfinClient** - Cliente completo para Jellyfin API
  - Tracking de reproducción (last watched, play count)
  - Favoritos de Jellyfin
  - Imágenes dinámicas
- [x] **JellyseerrClient** - Cliente completo para Jellyseerr/Overseerr API
  - Gestión de requests (pending, approved, available, denied)
  - Vinculación con Radarr/Sonarr IDs
- [x] **Modelos adicionales**
  - Tag (para tags de Radarr/Sonarr)
  - PlaybackInfo (información de reproducción Jellyfin)
  - Request (requests de Jellyseerr)
- [x] **Características comunes en todos los clientes**
  - Retry automático con backoff exponencial (3 intentos)
  - Timeouts configurables (default: 30s)
  - Logging estructurado con Zap
  - Manejo de contexto para cancelación
  - Validación de respuestas HTTP
  - Headers de autenticación correctos

### Archivos Creados en Sprint 1
```
internal/service/clients/
├── client.go           # Interfaces y constantes comunes
├── radarr.go          # Cliente de Radarr (354 líneas)
├── sonarr.go          # Cliente de Sonarr (345 líneas)
├── jellyfin.go        # Cliente de Jellyfin (282 líneas)
└── jellyseerr.go      # Cliente de Jellyseerr (252 líneas)

internal/models/
├── tag.go             # Modelo Tag
├── playback.go        # Modelo PlaybackInfo
└── request.go         # Modelo Request
```

## 🚧 En Progreso

*Nada actualmente*

## 📋 Pendiente

### Sprint 2: Servicio de Cleanup
- [ ] **CleanupService** - Lógica de limpieza
  - [ ] Evaluación de reglas (tiempo, tags, peticiones)
  - [ ] Modo dry-run
  - [ ] Verificación de exclusiones
  - [ ] Validación de estado de seeding (opcional)
- [ ] **CleanupStrategies**
  - [ ] TimeBasedStrategy (limpieza por antigüedad)
  - [ ] TagBasedStrategy (limpieza por tags)
  - [ ] EpisodeStrategy (limpieza de episodios vistos)
  - [ ] RequestStrategy (limpieza según Jellyseerr)
- [ ] **FilesystemService**
  - [ ] Creación de symlinks para "Leaving Soon"
  - [ ] Validación de paths
  - [ ] Gestión segura de archivos

### Sprint 3: Sistema de Programación
- [ ] **SchedulerService** - Gestión de cron jobs
  - [ ] Integración con robfig/cron v3
  - [ ] Configuración de schedules desde UI
  - [ ] Historial de ejecuciones
  - [ ] Notificaciones de errores
- [ ] **Modelos de Schedule**
  - [ ] Modelo Schedule (CRUD)
  - [ ] Expresiones cron
  - [ ] Estrategias asociadas

### Sprint 4: Repositorios y Persistencia
- [ ] **MediaRepository**
  - [ ] CRUD completo de Media
  - [ ] Filtros avanzados (tipo, tags, antigüedad)
  - [ ] Búsqueda por título
  - [ ] Paginación
- [ ] **HistoryRepository**
  - [ ] Registro de acciones de limpieza
  - [ ] Consultas por fecha/media/acción
- [ ] **ConfigRepository**
  - [ ] Configuración persistente
  - [ ] Versioning de configuración

### Sprint 5: API REST Completa
- [ ] **Handlers de Media**
  - [ ] GET /api/media (lista paginada)
  - [ ] GET /api/media/:id
  - [ ] PUT /api/media/:id/exclude
  - [ ] DELETE /api/media/:id
- [ ] **Handlers de Cleanup**
  - [ ] POST /api/cleanup/run (manual)
  - [ ] POST /api/cleanup/preview (dry-run)
  - [ ] GET /api/cleanup/history
- [ ] **Handlers de Configuración**
  - [ ] GET /api/config
  - [ ] PUT /api/config
  - [ ] POST /api/config/test-connection
- [ ] **Handlers de Schedules**
  - [ ] CRUD completo de schedules

### Sprint 6: Interfaz de Usuario
- [ ] **Página de Media Management**
  - [ ] Tabla con filtros
  - [ ] Acciones masivas
  - [ ] Modal de detalles
- [ ] **Página de Configuración**
  - [ ] Formularios de servicios
  - [ ] Test de conexión inline
  - [ ] Configuración de reglas
- [ ] **Página de Schedules**
  - [ ] Calendario visual
  - [ ] Editor de cron
  - [ ] Toggle enable/disable
- [ ] **Página de History**
  - [ ] Timeline de acciones
  - [ ] Filtros por fecha/tipo
  - [ ] Estadísticas visuales

### Sprint 7: Features Avanzados
- [ ] **Sistema de Notificaciones**
  - [ ] Webhook support
  - [ ] Email notifications
  - [ ] Discord/Telegram
- [ ] **Leaving Soon Collection**
  - [ ] Vista dedicada en UI
  - [ ] Actualización automática
  - [ ] Integración con Plex/Jellyfin collections
- [ ] **Import de Janitorr**
  - [ ] Parser de application.yml
  - [ ] Migración de configuración
  - [ ] Validación de datos importados

## 📦 Dependencias Actuales

```go
github.com/gofiber/fiber/v2 v2.52.9
github.com/gofiber/template/html/v2 v2.1.2
github.com/go-resty/resty/v2 v2.16.2       // ✨ Nuevo
github.com/google/uuid v1.6.0
github.com/robfig/cron/v3 v3.0.1
github.com/spf13/viper v1.19.0
go.uber.org/zap v1.27.0
gorm.io/driver/sqlite v1.5.6
gorm.io/gorm v1.25.12
golang.org/x/crypto v0.35.0                // ✨ Actualizado
golang.org/x/net v0.38.0                   // ✨ Actualizado
golang.org/x/sys v0.28.0
```

## 🎯 Próximos Pasos

1. **Implementar CleanupService** (Sprint 2)
   - Crear servicio con inyección de dependencias de clientes
   - Implementar estrategias de limpieza
   - Añadir tests unitarios para cada estrategia

2. **Crear FilesystemService**
   - Manejo seguro de archivos y symlinks
   - Validación de paths
   - Tests con filesystem temporal

3. **Integrar clientes con repositorios**
   - Sincronización de media desde servicios externos
   - Actualización de base de datos
   - Gestión de conflictos

## 📈 Métricas del Proyecto

- **Líneas de código Go**: ~2,500 (estimado)
- **Archivos fuente**: 35+
- **Tests**: 0 (pendiente)
- **Cobertura de tests**: 0% (pendiente)
- **Tiempo de desarrollo**: ~5 días
- **Imagen Docker**: ~25MB (objetivo)
- **Uso de RAM**: ~50MB (objetivo)

## 🔐 Seguridad

- ✅ Fix de seguridad en Fiber v2.52.9 (GHSA-hg3g-gphw-5hhm BodyParser)
- ✅ golang.org/x/crypto actualizado con fixes de SSH
- ✅ golang.org/x/net actualizado con mejoras HTTP/2
- ⏳ Pendiente: Validación de inputs de usuario
- ⏳ Pendiente: Rate limiting en API
- ⏳ Pendiente: Sanitización de file paths

## 📝 Notas

### Decisiones de Diseño

1. **Retry con Backoff Exponencial**: Todos los clientes implementan retry automático con backoff exponencial (1s, 2s, 4s) para manejar fallos temporales de red.

2. **Context Propagation**: Todos los métodos aceptan `context.Context` para permitir cancelación y timeouts desde el caller.

3. **Logging Estructurado**: Uso consistente de Zap con campos estructurados para mejor observabilidad.

4. **Interfaces Separadas**: MediaClient, StreamingClient y RequestClient separados por responsabilidad (diferentes tipos de servicios).

5. **Conversión a Modelo Interno**: Cada cliente convierte sus estructuras específicas al modelo `Media` interno para desacoplar la lógica de negocio de las APIs externas.

### Lecciones Aprendidas

1. **Go 1.25 Compatibility**: Air v1.63+ requiere Go 1.25. Importante mantener versiones consistentes.

2. **Fiber Updates**: Los PRs de Dependabot son valiosos - Fiber v2.52.9 incluía un fix de seguridad importante.

3. **Docker Compose Watch**: Excelente para desarrollo, pero requiere configuración cuidadosa para evitar conflictos con bind mounts.

4. **API Differences**: Radarr/Sonarr usan `X-Api-Key`, Jellyfin usa `X-Emby-Token` - importante documentar estas diferencias.

---

**Mantenido por**: GitHub Copilot  
**Repositorio**: https://github.com/carcheky/keepercheky  
**Branch**: develop
