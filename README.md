# MediaCheky

> Sistema de gestión de servidores multimedia fácil de configurar mediante interfaz web dockerizada

[![Status](https://img.shields.io/badge/status-active-success)](https://github.com/carcheky/mediacheky)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.22+-00ADD8?logo=go)](https://go.dev/)

## 📖 ¿Qué es MediaCheky?

MediaCheky es un sistema multimedia completo que permite configurar y gestionar fácilmente servicios multimedia mediante una interfaz web intuitiva. Basado en la arquitectura de [KeeperCheky](https://github.com/carcheky/keepercheky), proporciona una manera simple de levantar y configurar un servidor multimedia completo usando Docker.

### Características Principales

- 🎨 **Interfaz Web Moderna** - Dashboard intuitivo accesible desde cualquier navegador
- 🐳 **Dockerizado** - Todos los servicios corriendo en contenedores Docker
- ⚙️ **Configuración Simple** - Activa y desactiva servicios con un clic
- 📦 **Generación Automática** - Genera automáticamente docker-compose.yml
- 🔧 **Servicios Integrados** - Gestión de múltiples servicios multimedia
- 💾 **Base de Datos SQLite** - Configuración persistente y ligera
- 🚀 **Mínimos Recursos** - Optimizado para bajo consumo de memoria

## 🎬 Servicios Incluidos

MediaCheky permite gestionar los siguientes servicios multimedia:

| Servicio | Descripción | Puerto por defecto |
|----------|-------------|-------------------|
| **Jellyfin** | Servidor multimedia (alternativa a Plex) | 8096 |
| **Sonarr** | Gestión de series de TV | 8989 |
| **Radarr** | Gestión de películas | 7878 |
| **Prowlarr** | Indexador de torrents | 9696 |
| **qBittorrent** | Cliente de torrents | 8080 |
| **Jellyseerr** | Sistema de peticiones para Jellyfin | 5055 |
| **qBit Manager** | Gestor automático de torrents | - |
| **Docker Controller Bot** | Bot de Telegram para controlar Docker | - |
| **Bazarr** | Gestión de subtítulos | 6767 |
| **Jellystat** | Estadísticas de Jellyfin | 3000 |

## 🚀 Inicio Rápido

### Requisitos Previos

- Docker y Docker Compose instalados
- Go 1.22+ (solo para desarrollo)
- Puerto 8080 disponible

### Instalación con Docker (Recomendado)

```bash
# Clonar el repositorio
git clone https://github.com/carcheky/mediacheky.git
cd mediacheky

# Levantar el servicio
docker-compose up -d

# Acceder a la interfaz web
# Abrir http://localhost:8080
```

### Compilación desde Código Fuente

```bash
# Clonar el repositorio
git clone https://github.com/carcheky/mediacheky.git
cd mediacheky

# Descargar dependencias
go mod download

# Compilar
go build -o mediacheky ./cmd/server

# Ejecutar
./mediacheky
```

## 💻 Uso

1. **Accede a la interfaz web** en `http://localhost:8080`

2. **Activa los servicios** que deseas utilizar desde la página de "Servicios"

3. **Descarga la configuración** generada haciendo clic en "Descargar Docker Compose"

4. **Levanta los servicios** usando el docker-compose.yml generado:
   ```bash
   docker-compose -f docker-compose-generated.yml up -d
   ```

## 🏗️ Arquitectura

MediaCheky está construido siguiendo las mejores prácticas de desarrollo en Go:

```
mediacheky/
├── cmd/
│   └── server/          # Punto de entrada de la aplicación
│       └── main.go
├── internal/
│   ├── config/          # Gestión de configuración
│   ├── handler/         # Handlers HTTP
│   ├── models/          # Modelos de datos
│   └── service/         # Lógica de negocio
├── pkg/
│   └── logger/          # Logger compartido
├── web/
│   ├── templates/       # Plantillas HTML
│   └── static/          # Archivos estáticos
├── Dockerfile           # Imagen de producción
├── docker-compose.yml   # Configuración de despliegue
└── go.mod              # Dependencias Go
```

### Stack Tecnológico

**Backend:**
- Go 1.22+
- Fiber v2 (framework web)
- GORM v2 (ORM)
- SQLite (base de datos)
- Viper (configuración)
- Zap (logging)

**Frontend:**
- Alpine.js 3.x (framework reactivo ligero)
- Tailwind CSS (estilos)
- Font Awesome (iconos)

## ⚙️ Configuración

### Variables de Entorno

```bash
# Aplicación
MEDIACHEKY_APP_ENVIRONMENT=production
MEDIACHEKY_APP_LOG_LEVEL=info

# Servidor
MEDIACHEKY_SERVER_HOST=0.0.0.0
MEDIACHEKY_SERVER_PORT=8080

# Base de Datos
MEDIACHEKY_DATABASE_TYPE=sqlite
MEDIACHEKY_DATABASE_PATH=/data/mediacheky.db
```

### Archivo de Configuración (config.yml)

También puedes usar un archivo YAML para la configuración:

```yaml
app:
  environment: production
  loglevel: info

server:
  host: 0.0.0.0
  port: 8080

database:
  type: sqlite
  path: /data/mediacheky.db

services:
  jellyfin:
    enabled: true
    image: jellyfin/jellyfin:latest
    port: "8096"
```

## 📁 Estructura de Datos

Los datos se almacenan en `/data`:

```
data/
├── mediacheky.db        # Base de datos SQLite
└── services/            # Configuraciones de servicios
```

## 🔧 Desarrollo

### Requisitos de Desarrollo

- Go 1.22+
- Docker y Docker Compose
- Make (opcional)

### Comandos de Desarrollo

```bash
# Instalar dependencias
go mod download

# Ejecutar en modo desarrollo
go run cmd/server/main.go

# Compilar
go build -o mediacheky ./cmd/server

# Ejecutar tests
go test ./...

# Compilar para producción
CGO_ENABLED=1 go build -ldflags="-w -s" -o mediacheky ./cmd/server
```

### Ejecutar con Docker en Desarrollo

```bash
# Construir imagen
docker build -t mediacheky:dev .

# Ejecutar contenedor
docker run -p 8080:8080 -v $(pwd)/data:/data mediacheky:dev
```

## 🤝 Contribuir

Las contribuciones son bienvenidas. Por favor:

1. Haz fork del proyecto
2. Crea una rama para tu feature (`git checkout -b feature/AmazingFeature`)
3. Commit tus cambios (`git commit -m 'Add some AmazingFeature'`)
4. Push a la rama (`git push origin feature/AmazingFeature`)
5. Abre un Pull Request

## 📝 Licencia

Este proyecto está bajo la Licencia MIT - ver el archivo [LICENSE](LICENSE) para más detalles.

## 🙏 Agradecimientos

- **[KeeperCheky](https://github.com/carcheky/keepercheky)** - Proyecto base e inspiración
- Todos los proyectos *arr (Radarr, Sonarr, etc.)
- Jellyfin y la comunidad de código abierto

## 📞 Enlaces

- **Documentación**: [Wiki](https://github.com/carcheky/mediacheky/wiki)
- **Issues**: [GitHub Issues](https://github.com/carcheky/mediacheky/issues)
- **KeeperCheky**: [github.com/carcheky/keepercheky](https://github.com/carcheky/keepercheky)

---

**Status**: ✅ Activo - Listo para usar