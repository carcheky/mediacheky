# MediaCheky

> Panel de control centralizado para servicios multimedia *arr

[![stable](https://img.shields.io/github/actions/workflow/status/carcheky/mediacheky/release.yml?branch=stable&label=stable&logo=github)](https://github.com/carcheky/mediacheky/actions/workflows/release.yml)
[![develop](https://img.shields.io/github/actions/workflow/status/carcheky/mediacheky/release.yml?branch=develop&label=develop&logo=github)](https://github.com/carcheky/mediacheky/actions/workflows/release.yml)
[![Docker Image](https://img.shields.io/badge/docker-ghcr.io-blue?logo=docker)](https://github.com/carcheky/mediacheky/pkgs/container/mediacheky)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/carcheky/mediacheky)](go.mod)

**MediaCheky** es un panel de control web unificado para gestionar, configurar y monitorear toda tu suite de servicios multimedia desde un único lugar. Piensa en ello como un *Portainer* específico para el ecosistema *arr y servicios relacionados.

## ✨ ¿Qué hace MediaCheky?

En lugar de configurar cada servicio por separado, MediaCheky te proporciona:

- 📊 **Dashboard centralizado** - Vista general del estado de todos los servicios
- ⚙️ **Configuración unificada** - Activa, desactiva y configura servicios desde una interfaz
- 🔌 **Gestión de servicios** - Levanta/para contenedores Docker automáticamente
- 🌐 **Variables globales** - Comparte configuración común entre servicios (timezone, PUID/PGID, rutas, etc.)
- 🎯 **Panel por servicio** - Configuración específica para cada servicio con formularios intuitivos

## 🎬 Servicios Soportados

### Actualmente
- **Jellyfin** - Servidor multimedia
- **Sonarr** - Gestión de series
- **Radarr** - Gestión de películas
- **Prowlarr** - Gestión de indexadores
- **qBittorrent** - Cliente torrent
- **Jellyseerr** - Sistema de peticiones
- **Bazarr** - Gestión de subtítulos
- **Jellystat** - Estadísticas de Jellyfin

### Próximamente
- Lidarr (música)
- Readarr (libros)
- Overseerr (alternativa a Jellyseerr)
- Tautulli (estadísticas Plex)
- Y más...

## � Inicio Rápido

### Requisitos
- Docker
- Docker Compose

### Instalación

```bash
git clone https://github.com/carcheky/mediacheky.git
cd mediacheky
make dev
```

Accede a: **http://localhost:7369**

### Con Docker

```bash
docker run -d \
  --name mediacheky \
  -p 7369:7369 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v ./config:/app/config \
  -v ./data:/app/data \
  -v ./services:/app/services \
  ghcr.io/carcheky/mediacheky:latest
```

**⚠️ Importante:** MediaCheky necesita acceso al socket de Docker (`/var/run/docker.sock`) para gestionar contenedores.

## 🎨 Interfaz de Usuario

MediaCheky ofrece 4 pestañas principales:

### 1. � Dashboard
Vista general del estado de todos los servicios:
- ✅ Servicios activos y funcionando
- ⚠️ Servicios con problemas
- 📴 Servicios desactivados
- Estadísticas rápidas

### 2. ⚙️ Settings
Habilitar/deshabilitar servicios de forma global:
- Toggle para cada servicio
- Estado actual (activo/inactivo)
- Acciones rápidas (iniciar/detener)

### 3. 🔧 Services
Panel de configuración individual por servicio:
- Formularios específicos para cada servicio
- Configuración de puertos, rutas, API keys
- Templates predefinidos
- Validación de configuración

### 4. 🌐 Global Variables
Variables compartidas entre todos los servicios:
- Timezone (TZ)
- PUID/PGID
- Rutas base (/media, /downloads, etc.)
- Configuración de red

## 🏗️ Arquitectura

MediaCheky utiliza una arquitectura híbrida para gestionar servicios:

1. **Templates de Docker Compose** - Cada servicio tiene un template predefinido
2. **Generación dinámica** - MediaCheky genera archivos `docker compose.yml` basados en tu configuración
3. **Docker Socket API** - Controla contenedores usando el socket de Docker del host
4. **Base de datos local** - SQLite para almacenar configuraciones

```
mediacheky/
├── volumes/
│   ├── services/          # Archivos docker compose.yml generados
│   │   ├── radarr/
│   │   ├── sonarr/
│   │   └── jellyfin/
│   ├── config/            # Configuración de MediaCheky
│   └── data/              # Base de datos SQLite
```

## 🛠️ Stack Tecnológico

- **Backend**: Go 1.25 + Fiber v2
- **Frontend**: Alpine.js 3.x + Tailwind CSS
- **Base de datos**: GORM v2 + SQLite
- **Docker**: Docker Engine API
- **Imagen**: ~25MB (Alpine-based)
- **Memoria**: ~30-50MB RAM

## 🔒 Seguridad

**⚠️ Advertencia de Seguridad:**

MediaCheky requiere acceso al socket de Docker, lo cual equivale a **acceso root** en el host. 

Recomendaciones:
- ✅ No expongas MediaCheky directamente a internet
- ✅ Usa un reverse proxy con autenticación (Traefik, Nginx)
- ✅ Configura firewalls adecuados
- ✅ Revisa los logs regularmente

## 📚 Documentación

- [Plan de Desarrollo](docs/PROJECT_PLAN.md) - Roadmap y arquitectura
- [Guía de Desarrollo](DEVELOPMENT.md) - Setup para contribuidores
- [Guía para Agentes IA](AGENTS.md) - Instrucciones para GitHub Copilot

## 🤝 Contribuir

¡Las contribuciones son bienvenidas! Áreas donde puedes ayudar:

- 🎨 Mejorar UI/UX
- 🔌 Añadir soporte para más servicios
- 🧪 Tests y validación
- 📝 Documentación
- 🐛 Reportar bugs

## �📝 Licencia

MIT License - Ver [LICENSE](LICENSE) para detalles

## 🎯 Estado del Proyecto

**Fase actual**: Desarrollo inicial - Redefinición completa del proyecto

- [ ] Dashboard básico
- [ ] Sistema de configuración de servicios
- [ ] Integración Docker Compose
- [ ] Templates de servicios principales
- [ ] UI con Alpine.js + Tailwind

---

**Stack**: Go 1.25 + Fiber v2 + Alpine.js 3.x + Tailwind CSS  
**Puerto**: 7369 🦊✨



