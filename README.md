# KeeperCheky

> Gestor moderno de limpieza para bibliotecas multimedia - Reescritura completa de Janitorr con interfaz web

[![stable](https://img.shields.io/github/actions/workflow/status/carcheky/keepercheky/release.yml?branch=stable&label=stable&logo=github)](https://github.com/carcheky/keepercheky/actions/workflows/release.yml)
[![stable version](https://img.shields.io/github/v/release/carcheky/keepercheky?label=stable)](https://github.com/carcheky/keepercheky/releases)
[![develop](https://img.shields.io/github/actions/workflow/status/carcheky/keepercheky/release.yml?branch=develop&label=develop&logo=github)](https://github.com/carcheky/keepercheky/actions/workflows/release.yml)
[![develop version](https://img.shields.io/github/v/release/carcheky/keepercheky?include_prereleases&label=develop&filter=*-dev*)](https://github.com/carcheky/keepercheky/releases)
[![Docker Image](https://img.shields.io/badge/docker-ghcr.io-blue?logo=docker)](https://github.com/carcheky/keepercheky/pkgs/container/keepercheky)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/carcheky/keepercheky)](go.mod)

**KeeperCheky** automatiza la limpieza de tu biblioteca multimedia eliminando contenido antiguo o no visto según reglas configurables. Es una reescritura completa de [Janitorr](https://github.com/Schaka/janitorr) con interfaz web moderna, optimizado para mínimo uso de recursos.

## ✨ Características Principales

- 🎨 **Interfaz Web Moderna** - Dashboard intuitivo accesible desde cualquier navegador
- 🧹 **Limpieza Automatizada** - Eliminación inteligente basada en edad y espacio en disco
- 🏷️ **Reglas por Tags** - Programación personalizada usando tags de Radarr/Sonarr
- 📺 **Gestión de Series** - Manejo especial para shows semanales/diarios
- ⏰ **Próximos a Eliminar** - Vista previa en Jellyfin/Emby antes de borrar
- 🔗 **Integración Completa** - Compatible con Radarr, Sonarr, Jellyfin, Jellyseerr, qBittorrent, Bazarr
- 🐳 **Docker Ready** - Despliegue sencillo con Docker Compose
- 🔒 **Modo Seguro** - Dry-run por defecto para prevenir accidentes

## 🚀 Estado del Proyecto

**Fase Actual**: Desarrollo Activo - v1.0.0-dev.17

✅ **90% completado** - La mayoría de características implementadas  
🏗️ **Stack**: Go + Alpine.js para máximo rendimiento y mínimos recursos  
📦 **Docker**: ~25MB imagen, ~30-60MB RAM  
⚡ **Startup**: <2 segundos

## 📚 Documentación

**[👉 Comenzar aquí: Índice de Documentación](docs/README.md)**

### Enlaces Rápidos

- **[Instalación y Uso](quickstart/README.md)** - Guía de inicio rápido
- **[Desarrollo](DEVELOPMENT.md)** - Configuración del entorno de desarrollo
- **[Guía para Agentes IA](AGENTS.md)** - Instrucciones para GitHub Copilot y otros asistentes
- **[Resumen Ejecutivo](docs/RESUMEN_EJECUTIVO.md)** - Visión general del proyecto
- **[Comparación](docs/RESUMEN_COMPARATIVO.md)** - Janitorr vs Maintainerr vs KeeperCheky
- **[Propuestas Técnicas](docs/propuestas/)** - Análisis de 4 stacks diferentes

### Por qué Go + Alpine.js?

**Propuesta 3** seleccionada por balance óptimo:
- ✅ Rendimiento extremo (~30-60MB RAM)
- ✅ Imagen Docker tiny (~25MB)
- ✅ Binario único, sin dependencias
- ✅ Startup instantáneo (<2s)

Ver [análisis completo](docs/COMPARACION_Y_RECOMENDACIONES.md) de las 4 propuestas evaluadas.

## 📦 Instalación Rápida

### Opción 1: Docker Compose (Recomendado)

```bash
cd quickstart
cp .env.example .env
# Editar .env con tus configuraciones
docker-compose up -d

# Acceder a http://localhost:8780
```

Ver [guía completa de instalación](quickstart/README.md).

### Opción 2: Desarrollo

```bash
git clone https://github.com/carcheky/keepercheky.git
cd keepercheky
make init
make dev

# Acceder a http://localhost:8000
```

Ver [guía de desarrollo](DEVELOPMENT.md) para más detalles.

### Opción 3: Docker Manual

```bash
docker run -d \
  --name keepercheky \
  -p 8000:8000 \
  -v ./config:/config \
  -v ./data:/data \
  -v /path/to/media:/media:ro \
  ghcr.io/carcheky/keepercheky:latest
```

## ⚙️ Configuración Básica

```yaml
# config/config.yaml
app:
  dry_run: true              # ⚠️ Mantener en true hasta estar seguro
  leaving_soon_days: 14
  scheduler_enabled: false

clients:
  radarr:
    enabled: true
    url: "http://radarr:7878"
    api_key: "tu-api-key"
  
  jellyfin:
    enabled: true
    url: "http://jellyfin:8096"
    api_key: "tu-api-key"
```

Ver [ejemplo completo de configuración](.env.example).

## 🤝 Contribuir

¿Interesado en ayudar? Revisa:

- **[Guía de desarrollo](DEVELOPMENT.md)** - Configuración y workflows
- **[Guía para agentes IA](AGENTS.md)** - Instrucciones para Copilot
- **[Documentación técnica](docs/)** - Propuestas y arquitectura

### Áreas que necesitan ayuda

- 🧪 Tests unitarios y de integración
- 📝 Documentación y ejemplos
- 🐛 Reportar y corregir bugs
- 💡 Sugerencias de features
- 🌍 Traducciones

## 📝 Licencia

MIT License - Ver [LICENSE](LICENSE) para detalles

## 🙏 Agradecimientos

- **[Janitorr](https://github.com/Schaka/janitorr)** - Proyecto original que inspiró esta reescritura
- **[Maintainerr](https://github.com/jorenn92/Maintainerr)** - Referencia para UI/UX y features avanzadas
- Proyectos *arr (Radarr, Sonarr, etc.)
- Comunidades de Jellyfin y Emby

---

**Estado**: v1.0.0-dev.17 - Desarrollo activo  
**Documentación**: [docs/README.md](docs/README.md) | **Instalación**: [quickstart/README.md](quickstart/README.md) | **Desarrollo**: [DEVELOPMENT.md](DEVELOPMENT.md)


