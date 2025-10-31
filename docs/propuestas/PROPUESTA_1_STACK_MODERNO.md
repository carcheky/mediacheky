# Propuesta 1: Stack Moderno Full-Stack (TypeScript)

## 🎯 Visión General

Reescribir Janitorr como una aplicación web moderna con arquitectura de microservicios, utilizando tecnologías actuales del ecosistema JavaScript/TypeScript. Esta propuesta prioriza la experiencia de usuario, escalabilidad y facilidad de mantenimiento.

## 🏗️ Arquitectura

### Stack Tecnológico

#### Frontend
- **Framework**: Next.js 15 (App Router) con TypeScript
- **UI Library**: shadcn/ui + Tailwind CSS
- **Estado**: Zustand + React Query (TanStack Query)
- **Formularios**: React Hook Form + Zod
- **Gráficos**: Recharts
- **Notificaciones**: Sonner

#### Backend
- **Framework**: NestJS con TypeScript
- **ORM**: Prisma
- **Base de datos**: PostgreSQL
- **Cache**: Redis
- **Jobs/Scheduler**: Bull + BullMQ
- **Validación**: Class Validator + Class Transformer

#### DevOps
- **Containerización**: Docker + Docker Compose
- **Reverse Proxy**: Traefik (integrado)
- **Logs**: Winston + Pino
- **Monitoreo**: Healthcheck endpoints

## 📐 Estructura del Proyecto

```
keepercheky/
├── apps/
│   ├── web/                    # Next.js Frontend
│   │   ├── app/
│   │   │   ├── (dashboard)/
│   │   │   │   ├── overview/
│   │   │   │   ├── media/
│   │   │   │   ├── schedules/
│   │   │   │   ├── settings/
│   │   │   │   └── logs/
│   │   │   ├── api/           # API Routes (opcional)
│   │   │   └── layout.tsx
│   │   ├── components/
│   │   │   ├── ui/            # shadcn components
│   │   │   ├── dashboard/
│   │   │   ├── media/
│   │   │   └── schedules/
│   │   ├── lib/
│   │   └── hooks/
│   │
│   └── api/                    # NestJS Backend
│       ├── src/
│       │   ├── modules/
│       │   │   ├── media/
│       │   │   ├── cleanup/
│       │   │   ├── clients/    # Radarr, Sonarr, etc.
│       │   │   ├── scheduler/
│       │   │   ├── stats/
│       │   │   └── config/
│       │   ├── common/
│       │   ├── database/
│       │   └── main.ts
│       └── prisma/
│           └── schema.prisma
│
├── packages/                   # Shared libraries
│   ├── types/
│   ├── utils/
│   └── config/
│
├── docker/
│   ├── web.Dockerfile
│   ├── api.Dockerfile
│   └── docker-compose.yml
│
└── scripts/
```

## 🎨 Interfaz de Usuario

### Páginas Principales

#### 1. Dashboard / Overview
- **Estadísticas en tiempo real**
  - Espacio en disco total/usado/libre (gráfico de dona)
  - Media próximo a eliminación (contador con indicador)
  - Últimas ejecuciones de limpieza (timeline)
  - Estado de conexiones con servicios externos (badges)
  
- **Widgets interactivos**
  - Gráfico de tendencia de espacio en disco (últimos 30 días)
  - Media añadido vs eliminado (gráfico de barras)
  - Próximos trabajos programados (lista)

#### 2. Media Management
- **Vista de tabla con filtros avanzados**
  - Búsqueda por título
  - Filtros: tipo (película/serie), estado, tags, edad
  - Ordenamiento multi-columna
  - Vista de tarjetas o lista
  
- **Detalles de cada item**
  - Poster/Banner
  - Información de archivo (tamaño, calidad, fecha de adquisición)
  - Historial de visualización (si Jellystat/Streamystats activo)
  - Estado de seeding (si aplica)
  - Tags aplicados
  - Tiempo restante hasta eliminación
  - Acciones: Excluir, Eliminar manualmente, Ver en Jellyfin/Radarr/Sonarr

#### 3. Cleanup Schedules
- **Vista de calendario/cronograma**
  - Visualización de trabajos programados
  - Crear/editar/eliminar schedules
  
- **Tipos de limpieza**
  - Media Cleanup (configuración de expiración por porcentaje de disco)
  - Tag-based Cleanup (reglas por tags)
  - Episode Cleanup (series diarias/semanales)
  
- **Configuración por tipo**
  - Formularios dinámicos según tipo
  - Preview de qué se eliminará
  - Activar/desactivar schedules
  - Ejecutar manualmente con confirmación

#### 4. Settings/Configuración
- **Tabs organizados**
  - **General**: Dry-run, Leaving Soon days, Exclusion tags
  - **Clients**: Configuración de Radarr, Sonarr, Bazarr, Jellyfin/Emby, Jellyseerr
  - **File System**: Paths, validación de seeding, free space check
  - **Notifications**: Email, Discord, Telegram (futuro)
  - **Advanced**: Logs level, cache settings, worker settings

- **Validación en tiempo real**
  - Test de conexión a cada servicio
  - Validación de paths
  - Preview de configuración

#### 5. Logs & History
- **Visor de logs en tiempo real**
  - WebSocket para streaming de logs
  - Filtros por nivel (INFO, WARN, ERROR, DEBUG)
  - Búsqueda de texto
  - Export a archivo
  
- **Historial de acciones**
  - Tabla paginada de todas las eliminaciones
  - Filtros por fecha, tipo, servicio
  - Detalles de cada acción
  - Opción de revertir (si es posible)

#### 6. Leaving Soon Collection
- **Vista dedicada**
  - Media próximo a eliminarse
  - Vista de galería con posters
  - Countdown timer para cada item
  - Opción de excluir directamente

## ⚙️ Funcionalidades Clave

### Backend (NestJS)

#### 1. Módulo de Clientes
```typescript
// Abstracción para todos los servicios externos
interface MediaClient {
  testConnection(): Promise<boolean>;
  getLibrary(): Promise<MediaItem[]>;
  deleteItem(id: string): Promise<void>;
}

// Implementaciones específicas
- RadarrClient
- SonarrClient
- JellyfinClient
- EmbyClient
- JellyseerrClient
- JellystatClient
- StreamystatsClient
- BazarrClient
```

#### 2. Módulo de Cleanup
```typescript
// Estrategias de limpieza
- MediaExpirationStrategy (basado en edad y espacio)
- TagBasedStrategy (por tags)
- EpisodeCleanupStrategy (series semanales)

// Scheduler de trabajos
- Cron-based scheduling
- Manual trigger
- Dry-run mode
- Rollback capability (metadata)
```

#### 3. Módulo de Stats
```typescript
// Recolección de métricas
- Disk usage trends
- Media acquisition/deletion stats
- Service health monitoring
- User activity (watches)
```

#### 4. Módulo de WebSocket
```typescript
// Real-time updates
- Log streaming
- Job progress
- Stats updates
```

### Frontend (Next.js)

#### 1. Server Actions
```typescript
// Para mutations desde el cliente
- updateConfig()
- triggerCleanup()
- excludeMedia()
- testClientConnection()
```

#### 2. React Query Hooks
```typescript
// Para data fetching
- useMedia()
- useSchedules()
- useStats()
- useConfig()
- useLogs()
```

#### 3. Zustand Stores
```typescript
// Estado global
- useAuthStore
- useConfigStore
- useUIStore (theme, sidebar state, etc.)
```

## 🚀 Deployment & DevOps

### Docker Compose Stack
```yaml
services:
  # Frontend
  keepercheky-web:
    image: keepercheky/web:latest
    environment:
      - NEXT_PUBLIC_API_URL=http://keepercheky-api:3001
    ports:
      - "3000:3000"
    
  # Backend API
  keepercheky-api:
    image: keepercheky/api:latest
    environment:
      - DATABASE_URL=postgresql://user:pass@postgres:5432/keepercheky
      - REDIS_URL=redis://redis:6379
    depends_on:
      - postgres
      - redis
    
  # Base de datos
  postgres:
    image: postgres:16-alpine
    volumes:
      - postgres_data:/var/lib/postgresql/data
    
  # Cache & Queue
  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data
```

### Variables de Entorno
```env
# API
NODE_ENV=production
PORT=3001
DATABASE_URL=postgresql://...
REDIS_URL=redis://...
JWT_SECRET=...

# Web
NEXT_PUBLIC_API_URL=http://api:3001
NEXT_PUBLIC_WS_URL=ws://api:3001
```

## 🎯 Ventajas de esta Propuesta

### ✅ Pros
1. **Stack moderno y bien documentado**: TypeScript + Next.js + NestJS
2. **Experiencia de usuario superior**: UI moderna con shadcn/ui
3. **Desarrollo rápido**: Ecosistema maduro con muchas librerías
4. **Type-safety completo**: TypeScript end-to-end
5. **Escalabilidad**: Arquitectura de microservicios
6. **Real-time updates**: WebSockets para logs y stats
7. **Facilidad de deployment**: Docker Compose single-command
8. **Gran comunidad**: Fácil encontrar ayuda y ejemplos
9. **SEO-friendly**: Next.js App Router con SSR
10. **Testing**: Jest + React Testing Library + Supertest

### ⚠️ Contras
1. **Tamaño de imagen Docker**: Mayor que alternativas minimalistas
2. **Memoria**: Requiere ~512MB-1GB RAM
3. **Complejidad inicial**: Más archivos y configuración
4. **Dependencias**: Node modules pueden ser pesados

## 📊 Estimación de Recursos

### Desarrollo
- **Tiempo estimado**: 4-6 semanas (1 desarrollador full-time)
- **Dificultad**: Media-Alta

### Runtime
- **RAM**: 512MB - 1GB
- **CPU**: 1-2 cores
- **Disco**: 500MB (imágenes Docker)
- **Red**: Mínima (solo APIs REST)

## 🔄 Migración desde Janitorr Original

### Estrategia
1. **Import de configuración**: Parser de `application.yml` a PostgreSQL
2. **Compatibilidad de API**: Mantener mismos endpoints
3. **Data migration**: Script para migrar historial (si existe)
4. **Side-by-side deployment**: Posibilidad de correr ambos en paralelo

## 🛣️ Roadmap de Implementación

### Fase 1: Core (2 semanas)
- [ ] Setup proyecto monorepo (Turborepo)
- [ ] Backend: Estructura básica NestJS + Prisma
- [ ] Frontend: Setup Next.js + shadcn/ui
- [ ] Implementar clientes base (Radarr, Sonarr)
- [ ] UI básica: Dashboard + Settings

### Fase 2: Funcionalidades (2 semanas)
- [ ] Cleanup schedules
- [ ] Media management UI
- [ ] Leaving Soon collections
- [ ] WebSocket para logs
- [ ] Sistema de jobs con Bull

### Fase 3: Refinamiento (1-2 semanas)
- [ ] Todos los clientes (Jellyfin, Emby, etc.)
- [ ] Testing completo
- [ ] Docker optimization
- [ ] Documentación

### Fase 4: Polish (1 semana)
- [ ] UI/UX improvements
- [ ] Performance optimization
- [ ] Error handling
- [ ] Beta testing

## 🎓 Conocimientos Requeridos

### Esenciales
- TypeScript avanzado
- React + Next.js 15
- NestJS
- Docker
- PostgreSQL

### Deseables
- Prisma ORM
- Bull/BullMQ
- Redis
- WebSockets
- Tailwind CSS

## 📝 Conclusión

Esta propuesta ofrece la mejor experiencia de usuario y developer experience, utilizando tecnologías modernas y probadas en producción. Es ideal para un proyecto que busca escalabilidad, mantenibilidad a largo plazo, y una interfaz web profesional y pulida.

**Recomendado si**: Priorizas UX, tienes experiencia en TypeScript/React, y recursos de servidor suficientes (512MB+ RAM).

**No recomendado si**: Necesitas mínimo uso de recursos o deployment ultra-simple.
