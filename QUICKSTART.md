# MediaCheky - Guía de Inicio Rápido

## 🚀 Comenzar en 1 Minuto

### Prerrequisitos

MediaCheky solo requiere **Docker** para funcionar. El comando `make dev` verificará automáticamente las dependencias y te guiará si falta algo.

- **Docker** (incluye Docker Compose v2)
  - 📥 [Instalar Docker Desktop](https://docs.docker.com/get-docker/) (Recomendado para Mac/Windows)
  - 📥 [Instalar Docker Engine](https://docs.docker.com/engine/install/) (Linux)
- **Go** (opcional) - Solo necesario si quieres ejecutar el código fuera de Docker
  - 📥 [Descargar Go](https://golang.org/dl/)

### Instalación y Ejecución

```bash
# 1. Clonar el repositorio
git clone https://github.com/carcheky/mediacheky.git
cd mediacheky

# 2. Configurar permisos de usuario (recomendado)
./scripts/setup-permissions.sh

# 3. ¡Ejecutar! (hace todo automáticamente)
make dev
```

**¡Eso es todo!** 🎉

El comando `make dev` automáticamente:
- ✅ Verifica que Docker y Docker Compose estén instalados
- ✅ Crea el archivo `.env` desde `.env.example` si no existe
- ✅ Crea todos los directorios necesarios (`volumes/`, `logs/`, `data/`, etc.)
- ✅ Hace los scripts ejecutables
- ✅ Construye e inicia los contenedores Docker
- ✅ Habilita hot-reload (los cambios en el código se recargan automáticamente)

**💡 Sobre permisos:**

El script `setup-permissions.sh` configura automáticamente `PUID` y `PGID` en el archivo `.env` con los valores de tu usuario. Esto evita problemas de permisos con archivos creados por MediaCheky y los servicios que gestiona.

Si no ejecutas el script, se usarán los valores por defecto (1000:1000). Si encuentras problemas de permisos más adelante, ejecuta el script y reinicia con `make restart`.

### Acceder a la Aplicación

Una vez iniciado, abre tu navegador en:

```
http://localhost:7369
```

Verás el dashboard de MediaCheky con tres páginas principales:
- **Dashboard** - Vista general del sistema
- **Settings** - Configuración de servicios
- **Logs** - Logs de la aplicación

## 📝 Comandos Útiles

### Durante el Desarrollo

```bash
# Ver todos los comandos disponibles
make help

# Ver logs en tiempo real
make logs

# Abrir shell en el contenedor
make shell

# Detener el servidor
make stop
# o presionar Ctrl+C en la terminal donde corre make dev

# Reiniciar (si cambiaste configuración)
make stop
make dev
```

### Verificar Dependencias Manualmente

```bash
# Verificar que todo esté instalado
make check-deps
```

Este comando te mostrará:
- ✅ Docker y su versión
- ✅ Docker Compose y su versión
- ✅ Go y su versión (opcional)
- ❌ Qué falta instalar y el enlace oficial

### Compilación y Tests

```bash
# Compilar binario de producción (requiere Go)
make build

# Ejecutar tests (requiere Go)
make test

# Ejecutar tests con coverage
make test-coverage

# Formatear código
make fmt

# Ejecutar linter (requiere golangci-lint)
make lint
```

### Docker

```bash
# Construir imagen de producción
make docker-build

# Construir imagen de desarrollo
make docker-build-dev

# Ejecutar imagen de producción
make docker-run
```

### Limpieza

```bash
# Limpiar artefactos de compilación
make clean

# Detener y limpiar contenedores y volúmenes
make stop-clean

# Limpiar biblioteca de medios de prueba
make clean-media
```

## 🔧 Configuración

### Archivo .env

El archivo `.env` se crea automáticamente desde `.env.example` la primera vez que ejecutas `make dev`. Puedes editarlo para cambiar:

```bash
# Configuración de la aplicación
MEDIACHEKY_APP_ENVIRONMENT=development
MEDIACHEKY_APP_LOG_LEVEL=debug
MEDIACHEKY_SERVER_PORT=7369

# Base de datos
MEDIACHEKY_DATABASE_TYPE=sqlite
MEDIACHEKY_DATABASE_PATH=/app/data/mediacheky.db
```

### Estructura de Directorios

Después de ejecutar `make dev`, tendrás:

```
mediacheky/
├── .env                    # Configuración (auto-creado)
├── volumes/                # Datos persistentes Docker
│   ├── mediacheky-data/    # Base de datos SQLite
│   ├── mediacheky-config/  # Archivos de configuración
│   └── media-library/      # Biblioteca de medios de prueba
├── logs/                   # Logs de la aplicación
│   └── mediacheky-dev.json # Log principal (auto-rotado)
├── data/                   # Datos locales
└── config/                 # Configuración local
```

## 🐛 Resolución de Problemas

### Docker no está instalado

```bash
$ make dev
❌ Docker is not installed
📥 Install from: https://docs.docker.com/get-docker/
```

**Solución**: Instala Docker Desktop o Docker Engine desde el enlace proporcionado.

### Docker Compose no disponible

```bash
$ make dev
❌ Docker Compose is not available
📥 Install Docker Desktop or Docker Compose plugin
```

**Solución**: 
- **Mac/Windows**: Instala Docker Desktop (incluye Compose)
- **Linux**: Instala el plugin de Docker Compose

### Puerto 7369 ya en uso

Si ves un error como "port is already allocated":

```bash
# Opción 1: Detener el proceso que usa el puerto 7369
lsof -ti:7369 | xargs kill -9

# Opción 2: Cambiar el puerto en .env
MEDIACHEKY_SERVER_PORT=7370
```

### El contenedor no inicia

```bash
# Ver logs completos
docker compose logs mediacheky

# Reconstruir desde cero
make stop-clean
make dev
```

### Permisos en Linux

Si tienes problemas de permisos con volúmenes:

```bash
# Dar permisos a las carpetas
sudo chown -R $USER:$USER volumes/ logs/ data/
```

## 📚 Siguiente Paso

Una vez que tengas MediaCheky ejecutándose:

1. **Explora la interfaz**: Abre http://localhost:7369
2. **Lee la documentación**: Revisa [DEVELOPMENT.md](DEVELOPMENT.md) para entender la arquitectura
3. **Configura servicios**: Ve a Settings para conectar servicios multimedia
4. **Revisa el código**: El hot-reload está activado, ¡haz cambios y obsérvalos en acción!

## 🆘 ¿Necesitas Ayuda?

- 📖 [Documentación completa](docs/README.md)
- 🐛 [Reportar un problema](https://github.com/carcheky/mediacheky/issues)
- 💬 [Discusiones](https://github.com/carcheky/mediacheky/discussions)

---

**¡Feliz desarrollo!** 🚀
