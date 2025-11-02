# MediaCheky - Plan de Desarrollo del Proyecto

> Panel de control centralizado para servicios multimedia *arr

**Última actualización**: 3 de noviembre, 2025  
**Estado**: Redefinición completa del proyecto

---

## 🎯 Visión del Proyecto

**MediaCheky** es un panel de control web unificado que permite gestionar, configurar y monitorear toda la suite de servicios multimedia (*arr ecosystem) desde una única interfaz centralizada.

### Problema que Resuelve

Configurar y gestionar múltiples servicios multimedia (Jellyfin, Sonarr, Radarr, etc.) requiere:
- Editar múltiples archivos `docker-compose.yml`
- Recordar puertos, rutas y configuraciones de cada servicio
- Configurar variables de entorno repetitivas (PUID, PGID, TZ)
- Monitorear el estado de cada servicio individualmente

### Solución

MediaCheky centraliza todo en una interfaz web donde puedes:
- ✅ Activar/desactivar servicios con un click
- ✅ Configurar cada servicio con formularios intuitivos
- ✅ Compartir variables globales entre servicios
- ✅ Monitorear el estado de todos los servicios en un dashboard

---

## 🏗️ Arquitectura Técnica

### Stack Tecnológico

| Componente | Tecnología | Razón |
|------------|-----------|-------|
| **Backend** | Go 1.25 + Fiber v2 | Rendimiento, binario único, baja memoria |
| **Frontend** | Alpine.js 3.x | Ligero, reactivo, sin build step |
| **CSS** | Tailwind CSS | Utility-first, customizable |
| **Base de datos** | GORM v2 + SQLite | Simple, portátil, sin servidor externo |
| **Docker API** | Docker Engine API | Gestión nativa de contenedores |
| **Templates** | Go html/template | Renderizado server-side |

### Arquitectura de Gestión de Servicios

**Enfoque Híbrido: Docker Socket + Docker Compose Templates**

```
┌─────────────────────────────────────────┐
│         MediaCheky Web UI               │
│  (Alpine.js + Tailwind CSS)             │
└────────────────┬────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────┐
│      MediaCheky Backend (Go + Fiber)    │
│                                          │
│  ┌────────────────────────────────────┐ │
│  │  Service Manager                   │ │
│  │  - Templates Engine                │ │
│  │  - Config Generator                │ │
│  │  - Docker API Client               │ │
│  └────────────────────────────────────┘ │
└────────────────┬────────────────────────┘
                 │
                 ├──────────────────────────┐
                 │                          │
                 ▼                          ▼
┌─────────────────────────┐  ┌──────────────────────────┐
│  Docker Socket          │  │  Generated Configs       │
│  /var/run/docker.sock   │  │  volumes/services/       │
│                         │  │  ├── radarr/             │
│  - Start/Stop services  │  │  │   └── compose.yml     │
│  - Monitor status       │  │  ├── sonarr/             │
│  - Get logs             │  │  │   └── compose.yml     │
└─────────────────────────┘  │  └── jellyfin/           │
                             │      └── compose.yml     │
                             └──────────────────────────┘
```

### Flujo de Trabajo

1. **Usuario activa Radarr en UI** → Click en toggle "Enable Radarr"
2. **MediaCheky genera config** → Usa template + configuración del usuario → Crea `/volumes/services/radarr/docker-compose.yml`
3. **MediaCheky levanta servicio** → Ejecuta `docker compose up -d` vía Docker Socket
4. **MediaCheky monitorea** → Verifica estado del contenedor cada 30s
5. **Dashboard actualiza** → ✅ Radarr (Running)

---

## 📱 Interfaz de Usuario

### Estructura de Pestañas

```
┌────────────────────────────────────────────────┐
│  MediaCheky                    [User] [Logout] │
├────────────────────────────────────────────────┤
│  [Dashboard] [Settings] [Services] [Global]    │
├────────────────────────────────────────────────┤
│                                                 │
│  [Contenido de la pestaña activa]             │
│                                                 │
└────────────────────────────────────────────────┘
```

### 1. 📊 Dashboard (Vista Principal)

**Objetivo**: Vista general del estado de todos los servicios

**Contenido**:
```
┌─────────────────────────────────────────────┐
│  Services Overview                          │
│                                             │
│  ✅ Running (5)    ⚠️ Issues (1)   📴 Off (2) │
│                                             │
│  [Jellyfin]    ✅ Running  Port: 8096       │
│  [Radarr]      ✅ Running  Port: 7878       │
│  [Sonarr]      ✅ Running  Port: 8989       │
│  [qBittorrent] ⚠️ Warning  High CPU usage   │
│  [Prowlarr]    📴 Stopped                   │
│                                             │
│  Quick Stats:                               │
│  - Total services: 8                        │
│  - Docker host: linux/amd64                 │
│  - MediaCheky uptime: 2h 34m                │
└─────────────────────────────────────────────┘
```

**Funcionalidades**:
- Estado en tiempo real (polling cada 30s)
- Quick actions: Start/Stop/Restart
- Links directos a cada servicio
- Alertas si hay problemas
- Métricas básicas (CPU, memoria opcional)

### 2. ⚙️ Settings (Configuración Global)

**Objetivo**: Habilitar/deshabilitar servicios y configuración general

**Contenido**:
```
┌─────────────────────────────────────────────┐
│  Enable/Disable Services                    │
│                                             │
│  [x] Jellyfin     [Start] [Stop] [Config]  │
│  [x] Radarr       [Start] [Stop] [Config]  │
│  [x] Sonarr       [Start] [Stop] [Config]  │
│  [ ] Prowlarr     [Enable]                  │
│  [ ] Bazarr       [Enable]                  │
│  [x] qBittorrent  [Start] [Stop] [Config]  │
│  [x] Jellyseerr   [Start] [Stop] [Config]  │
│  [x] Jellystat    [Start] [Stop] [Config]  │
│                                             │
│  [+ Add Custom Service]                     │
└─────────────────────────────────────────────┘
```

**Funcionalidades**:
- Toggle enable/disable
- Quick start/stop
- Link a configuración específica
- Badge de estado (running/stopped/error)

### 3. 🔧 Services (Configuración Individual)

**Objetivo**: Configurar cada servicio con formularios específicos

**Contenido** (Ejemplo: Radarr):
```
┌─────────────────────────────────────────────┐
│  < Back to Services     [Save] [Cancel]     │
│                                             │
│  Radarr Configuration                       │
│                                             │
│  Basic Settings:                            │
│  Port:         [7878]                       │
│  Image:        [linuxserver/radarr:latest]  │
│  Restart:      [unless-stopped ▼]           │
│                                             │
│  Paths:                                     │
│  Config:       [/config/radarr]             │
│  Movies:       [/media/movies]              │
│  Downloads:    [/downloads]                 │
│                                             │
│  Environment:                               │
│  PUID:         [1000]  (from Global)        │
│  PGID:         [1000]  (from Global)        │
│  TZ:           [Europe/Madrid] (from Global)│
│                                             │
│  Advanced:                                  │
│  [ ] Enable VPN routing                     │
│  [ ] Custom network                         │
│                                             │
│  [Apply Changes] [Reset to Defaults]        │
└─────────────────────────────────────────────┘
```

**Formularios por Servicio**:

| Servicio | Campos Específicos |
|----------|-------------------|
| Jellyfin | Port, paths, transcoding hardware |
| Radarr | Port, API key, paths, quality profiles |
| Sonarr | Port, API key, paths, naming schemes |
| Prowlarr | Port, indexers config |
| qBittorrent | Port, WebUI credentials, download paths |
| Jellyseerr | Port, Jellyfin connection |
| Bazarr | Port, subtitle providers |
| Jellystat | Port, database config |

### 4. 🌐 Global Variables

**Objetivo**: Variables compartidas entre todos los servicios

**Contenido**:
```
┌─────────────────────────────────────────────┐
│  Global Configuration                       │
│                                             │
│  User/Group:                                │
│  PUID:    [1000]                            │
│  PGID:    [1000]                            │
│                                             │
│  System:                                    │
│  Timezone: [Europe/Madrid ▼]                │
│  Language: [es-ES ▼]                        │
│                                             │
│  Base Paths:                                │
│  Media:     [/media]                        │
│  Downloads: [/downloads]                    │
│  Config:    [/config]                       │
│                                             │
│  Network:                                   │
│  Network name: [mediacheky-net]             │
│  Subnet:       [172.20.0.0/16]              │
│                                             │
│  [Save Global Settings]                     │
└─────────────────────────────────────────────┘
```

**Funcionalidades**:
- Valores heredados por todos los servicios
- Override individual si es necesario
- Validación de rutas y permisos

---

## 🗄️ Estructura de Datos

### Modelos de Base de Datos

```go
// Service - Configuración de un servicio
type Service struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    
    // Identificación
    Name        string    `json:"name" gorm:"unique;not null"` // radarr, sonarr, etc.
    DisplayName string    `json:"display_name"`                 // Radarr, Sonarr, etc.
    Icon        string    `json:"icon"`                         // URL o emoji
    
    // Estado
    Enabled     bool      `json:"enabled" gorm:"default:false"`
    Status      string    `json:"status"`                       // running, stopped, error
    
    // Docker
    Image       string    `json:"image"`                        // linuxserver/radarr:latest
    ContainerID string    `json:"container_id"`                 // ID del contenedor Docker
    Port        int       `json:"port"`                         // Puerto principal
    
    // Configuración (JSON)
    Config      JSON      `json:"config" gorm:"type:json"`      // Configuración específica
    
    // Relaciones
    TemplateID  uint      `json:"template_id"`
    Template    Template  `json:"template" gorm:"foreignKey:TemplateID"`
}

// Template - Plantilla de docker-compose para un servicio
type Template struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    Name        string    `json:"name" gorm:"unique;not null"`
    Version     string    `json:"version"`
    Content     string    `json:"content" gorm:"type:text"`     // Template de docker-compose
    Schema      JSON      `json:"schema" gorm:"type:json"`      // JSON Schema para validación
}

// GlobalConfig - Configuración global compartida
type GlobalConfig struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    Key         string    `json:"key" gorm:"unique;not null"`   // PUID, PGID, TZ, etc.
    Value       string    `json:"value"`
    Category    string    `json:"category"`                     // user, system, paths, network
}

// ServiceLog - Logs de operaciones sobre servicios
type ServiceLog struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    CreatedAt   time.Time `json:"created_at"`
    ServiceID   uint      `json:"service_id"`
    Action      string    `json:"action"`                       // start, stop, restart, config_change
    Status      string    `json:"status"`                       // success, error
    Message     string    `json:"message"`
}
```

---

## 📦 Sistema de Templates

### Template de Docker Compose (Ejemplo: Radarr)

```yaml
# templates/radarr.yml
version: '3.8'
services:
  radarr:
    image: {{ .Image }}
    container_name: {{ .ContainerName }}
    environment:
      - PUID={{ .Global.PUID }}
      - PGID={{ .Global.PGID }}
      - TZ={{ .Global.Timezone }}
      {{- if .Config.UMASK }}
      - UMASK={{ .Config.UMASK }}
      {{- end }}
    volumes:
      - {{ .Paths.Config }}:/config
      - {{ .Paths.Movies }}:/movies
      - {{ .Paths.Downloads }}:/downloads
    ports:
      - "{{ .Port }}:7878"
    networks:
      - {{ .Global.NetworkName }}
    restart: {{ .RestartPolicy }}

networks:
  {{ .Global.NetworkName }}:
    external: true
```

### Schema de Configuración (JSON Schema)

```json
{
  "title": "Radarr Configuration",
  "type": "object",
  "properties": {
    "port": {
      "type": "integer",
      "minimum": 1024,
      "maximum": 65535,
      "default": 7878
    },
    "image": {
      "type": "string",
      "enum": [
        "linuxserver/radarr:latest",
        "linuxserver/radarr:develop",
        "linuxserver/radarr:nightly"
      ],
      "default": "linuxserver/radarr:latest"
    },
    "paths": {
      "type": "object",
      "properties": {
        "config": { "type": "string" },
        "movies": { "type": "string" },
        "downloads": { "type": "string" }
      },
      "required": ["config", "movies", "downloads"]
    },
    "umask": {
      "type": "string",
      "pattern": "^[0-7]{3,4}$",
      "default": "022"
    }
  },
  "required": ["port", "image", "paths"]
}
```

---

## 🔌 API REST

### Endpoints Principales

```
# Servicios
GET    /api/services              # Listar todos los servicios
GET    /api/services/:name        # Obtener servicio específico
POST   /api/services/:name/enable # Habilitar servicio
POST   /api/services/:name/disable# Deshabilitar servicio
POST   /api/services/:name/start  # Iniciar servicio
POST   /api/services/:name/stop   # Detener servicio
POST   /api/services/:name/restart# Reiniciar servicio
PUT    /api/services/:name/config # Actualizar configuración
GET    /api/services/:name/logs   # Obtener logs

# Templates
GET    /api/templates             # Listar templates disponibles
GET    /api/templates/:name       # Obtener template específico

# Configuración Global
GET    /api/config/global         # Obtener config global
PUT    /api/config/global         # Actualizar config global
GET    /api/config/global/:key    # Obtener valor específico
PUT    /api/config/global/:key    # Actualizar valor específico

# Dashboard
GET    /api/dashboard/stats       # Estadísticas generales
GET    /api/dashboard/health      # Health check de servicios

# Docker
GET    /api/docker/info           # Info del host Docker
GET    /api/docker/containers     # Listar contenedores
```

---

## 🗓️ Roadmap de Desarrollo

### Fase 1: MVP (Mínimo Producto Viable) - 2 semanas

**Objetivo**: Dashboard básico + configuración de 1 servicio

- [x] Setup del proyecto (Go + Fiber + Alpine.js)
- [ ] Base de datos (modelos + migraciones)
- [ ] UI básica (header + navegación + pestañas)
- [ ] Dashboard simple (lista de servicios)
- [ ] Integración Docker Socket (lectura de contenedores)
- [ ] Template Engine para docker-compose
- [ ] Implementar 1 servicio completo: **Radarr**
  - [ ] Template de docker-compose
  - [ ] Formulario de configuración
  - [ ] Start/Stop/Restart
  - [ ] Mostrar estado en Dashboard
- [ ] Sistema de configuración global básico

**Criterio de éxito**: Puedo activar Radarr desde MediaCheky y verlo funcionando

### Fase 2: Servicios Core - 2 semanas

**Objetivo**: Soportar los 8 servicios principales

- [ ] Implementar templates y formularios para:
  - [ ] Sonarr
  - [ ] Jellyfin
  - [ ] Prowlarr
  - [ ] qBittorrent
  - [ ] Jellyseerr
  - [ ] Bazarr
  - [ ] Jellystat
- [ ] Dashboard mejorado con métricas
- [ ] Sistema de logs por servicio
- [ ] Validación de configuración (JSON Schema)
- [ ] Variables globales funcionales (PUID, PGID, TZ, paths)

**Criterio de éxito**: Puedo gestionar los 8 servicios desde MediaCheky

### Fase 3: Features Avanzadas - 2 semanas

**Objetivo**: Mejorar UX y robustez

- [ ] Health checks automáticos
- [ ] Alertas y notificaciones
- [ ] Exportar/Importar configuración
- [ ] Backup automático de configs
- [ ] Logs centralizados (agregar logs de todos los servicios)
- [ ] Búsqueda en logs
- [ ] Quick links a UIs de servicios
- [ ] Modo oscuro
- [ ] Responsive design (móvil)

**Criterio de éxito**: La aplicación es usable en producción

### Fase 4: Expansión - Futuro

- [ ] Soporte para más servicios (Lidarr, Readarr, etc.)
- [ ] Sistema de plugins para servicios custom
- [ ] API para automatización externa
- [ ] Autenticación (usuarios, roles)
- [ ] Multi-servidor (gestionar múltiples hosts Docker)
- [ ] Estadísticas avanzadas
- [ ] Webhooks
- [ ] Docker Swarm / Kubernetes support

---

## 🎨 Convenciones de Código

### Estructura de Directorios

```
mediacheky/
├── cmd/
│   └── server/
│       └── main.go                 # Entry point
├── internal/
│   ├── config/
│   │   └── config.go               # Configuración de la app
│   ├── models/
│   │   ├── service.go              # Modelo Service
│   │   ├── template.go             # Modelo Template
│   │   └── global_config.go        # Modelo GlobalConfig
│   ├── repository/
│   │   ├── service_repo.go         # CRUD de servicios
│   │   └── config_repo.go          # CRUD de config global
│   ├── service/
│   │   ├── service_manager.go      # Lógica de gestión de servicios
│   │   ├── docker_client.go        # Cliente Docker API
│   │   └── template_engine.go      # Motor de templates
│   ├── handler/
│   │   ├── dashboard.go            # Handlers del dashboard
│   │   ├── service.go              # Handlers de servicios
│   │   └── config.go               # Handlers de config
│   └── middleware/
│       ├── auth.go                 # Autenticación (futuro)
│       └── logger.go               # Request logging
├── pkg/
│   ├── docker/
│   │   └── client.go               # Wrapper Docker API
│   └── validator/
│       └── schema.go               # Validación JSON Schema
├── web/
│   ├── templates/
│   │   ├── layouts/
│   │   │   └── base.html           # Layout base
│   │   ├── pages/
│   │   │   ├── dashboard.html      # Dashboard
│   │   │   ├── settings.html       # Settings
│   │   │   ├── services.html       # Services
│   │   │   └── global.html         # Global config
│   │   └── components/
│   │       ├── service_card.html   # Tarjeta de servicio
│   │       └── form_field.html     # Campo de formulario
│   └── static/
│       ├── css/
│       │   └── tailwind.css
│       └── js/
│           ├── dashboard.js        # Alpine.js components
│           └── services.js
├── templates/
│   ├── radarr.yml                  # Template docker-compose Radarr
│   ├── sonarr.yml                  # Template docker-compose Sonarr
│   └── ...                         # Más templates
├── volumes/
│   ├── services/                   # Generados por MediaCheky
│   │   ├── radarr/
│   │   │   └── docker-compose.yml
│   │   └── ...
│   ├── config/                     # Configuración de MediaCheky
│   └── data/                       # Base de datos SQLite
└── docs/
    ├── PROJECT_PLAN.md             # Este documento
    └── ...
```

### Nomenclatura

- **Archivos**: `snake_case.go` o `kebab-case.html`
- **Tipos/Structs**: `PascalCase`
- **Funciones/Métodos**: `PascalCase` (exportadas), `camelCase` (privadas)
- **Variables**: `camelCase`
- **Constantes**: `UPPER_SNAKE_CASE`
- **Templates**: `kebab-case.yml`

---

## 🧪 Testing

### Estrategia de Testing

1. **Unit Tests** - Lógica de negocio
   - Service Manager
   - Template Engine
   - Validadores

2. **Integration Tests** - API + DB
   - Endpoints REST
   - Persistencia

3. **E2E Tests** (futuro) - UI completa
   - Flujos de usuario con Playwright

### Cobertura Objetivo

- **Fase 1 (MVP)**: 40% mínimo
- **Fase 2**: 60% mínimo
- **Fase 3**: 80% ideal

---

## 🔒 Consideraciones de Seguridad

### Riesgos

1. **Docker Socket = Root Access**
   - MediaCheky tiene acceso root al host
   - Cualquier vulnerabilidad puede comprometer el sistema

2. **Sin autenticación inicial**
   - Cualquiera con acceso a la red puede controlar servicios

3. **Command Injection**
   - Inputs del usuario usados en comandos Docker

### Mitigaciones

1. **Docker Socket**
   - ✅ Documentar claramente los riesgos
   - ✅ No exponer a internet sin reverse proxy
   - ✅ Usar firewall

2. **Autenticación** (Fase 3)
   - ✅ Implementar login básico
   - ✅ Roles (admin, viewer)
   - ✅ Tokens de sesión

3. **Validación**
   - ✅ Whitelist de imágenes permitidas
   - ✅ Sanitizar todos los inputs
   - ✅ Validar paths (evitar `../`)
   - ✅ Usar JSON Schema para configs

4. **Auditoría**
   - ✅ Log de todas las acciones
   - ✅ Registro de cambios de configuración

---

## 📚 Referencias

### Proyectos Similares

- **Portainer** - Panel para gestionar Docker
- **Yacht** - Template manager para Docker
- **Dockge** - Docker Compose stack manager
- **CasaOS** - Home server management

### Documentación Técnica

- [Docker Engine API](https://docs.docker.com/engine/api/)
- [Docker SDK for Go](https://pkg.go.dev/github.com/docker/docker/client)
- [Fiber Framework](https://docs.gofiber.io/)
- [Alpine.js](https://alpinejs.dev/)
- [Tailwind CSS](https://tailwindcss.com/)

---

## ✅ Checklist de Lanzamiento v1.0

### Funcionalidades Core

- [ ] Dashboard con estado de servicios
- [ ] Habilitar/deshabilitar servicios
- [ ] Configurar los 8 servicios principales
- [ ] Variables globales funcionales
- [ ] Start/Stop/Restart servicios
- [ ] Templates predefinidos
- [ ] Validación de configuración
- [ ] Logs por servicio

### Calidad

- [ ] Tests unitarios (60%+ cobertura)
- [ ] Documentación completa
- [ ] README con ejemplos
- [ ] Docker image publicada
- [ ] CI/CD configurado
- [ ] Release notes

### Seguridad

- [ ] Validación de inputs
- [ ] Whitelist de imágenes
- [ ] Logs de auditoría
- [ ] Documentación de riesgos

---

**Última revisión**: 3 de noviembre, 2025  
**Mantenedor**: @carcheky  
**Licencia**: MIT
