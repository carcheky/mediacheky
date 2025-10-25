# MediaCheky - Guía de Desarrollo

## Requisitos

- Go 1.22 o superior
- Docker y Docker Compose
- Git

## Inicio Rápido

### 1. Clonar el Repositorio

```bash
git clone https://github.com/carcheky/mediacheky.git
cd mediacheky
```

### 2. Instalar Dependencias

```bash
go mod download
```

### 3. Ejecutar en Desarrollo

```bash
# Opción 1: Usar go run
MEDIACHEKY_APP_ENVIRONMENT=development go run ./cmd/server

# Opción 2: Compilar y ejecutar
make build
./bin/mediacheky

# Opción 3: Usar Make
make run
```

La aplicación estará disponible en `http://localhost:8080`

### 4. Ejecutar con Docker

```bash
# Construir la imagen
docker build -t mediacheky:latest .

# O usar docker-compose
docker-compose up -d
```

## Estructura del Proyecto

```
mediacheky/
├── cmd/
│   └── server/          # Punto de entrada principal
│       └── main.go
├── internal/
│   ├── config/          # Gestión de configuración
│   │   └── config.go
│   ├── handler/         # Manejadores HTTP
│   │   └── handler.go
│   ├── models/          # Modelos de datos (GORM)
│   │   └── service.go
│   └── service/         # Lógica de negocio
│       └── service.go
├── pkg/
│   └── logger/          # Logger compartido
│       └── logger.go
├── web/
│   └── templates/       # Plantillas HTML
│       ├── layout.html
│       ├── index.html
│       └── services.html
├── config/
│   └── config.example.yml  # Ejemplo de configuración
├── Dockerfile           # Dockerfile de producción
├── docker-compose.yml   # Configuración de Docker Compose
├── Makefile            # Comandos de desarrollo
└── README.md           # Documentación principal
```

## Comandos de Make

```bash
make help           # Mostrar ayuda
make build          # Compilar la aplicación
make run            # Ejecutar la aplicación
make test           # Ejecutar tests
make clean          # Limpiar artefactos
make docker-build   # Construir imagen Docker
make docker-run     # Ejecutar con Docker Compose
make fmt            # Formatear código
make deps           # Descargar dependencias
make tidy           # Limpiar dependencias
```

## Variables de Entorno

### Aplicación
- `MEDIACHEKY_APP_ENVIRONMENT`: Entorno (production/development)
- `MEDIACHEKY_APP_LOG_LEVEL`: Nivel de log (debug/info/warn/error)

### Servidor
- `MEDIACHEKY_SERVER_HOST`: Host del servidor (default: 0.0.0.0)
- `MEDIACHEKY_SERVER_PORT`: Puerto del servidor (default: 8080)

### Base de Datos
- `MEDIACHEKY_DATABASE_TYPE`: Tipo de BD (default: sqlite)
- `MEDIACHEKY_DATABASE_PATH`: Ruta de la BD (default: /data/mediacheky.db)

## Desarrollo

### Agregar un Nuevo Servicio

1. Actualizar `internal/service/service.go`:
   ```go
   {
       Name:        "nuevo-servicio",
       DisplayName: "Nuevo Servicio",
       Image:       "imagen/del-servicio:latest",
       Port:        "puerto",
       WebPort:     "puerto",
       Enabled:     false,
   }
   ```

2. Actualizar `internal/config/config.go` si es necesario

3. La UI se actualizará automáticamente

### Modificar Templates

Los templates están en `web/templates/`:
- `layout.html`: Layout base
- `index.html`: Página principal
- `services.html`: Gestión de servicios

### API Endpoints

- `GET /`: Página principal
- `GET /services`: Gestión de servicios
- `POST /api/services/:name/toggle`: Activar/desactivar servicio
- `GET /api/docker-compose`: Obtener configuración
- `GET /docker-compose/download`: Descargar docker-compose.yml
- `GET /health`: Health check

## Testing

### Manual

1. Iniciar la aplicación
2. Navegar a `http://localhost:8080`
3. Activar/desactivar servicios
4. Descargar docker-compose.yml

### Verificar la Configuración Generada

```bash
curl http://localhost:8080/api/docker-compose
```

## Debugging

### Habilitar Logs de Debug

```bash
MEDIACHEKY_APP_LOG_LEVEL=debug go run ./cmd/server
```

### Ver Base de Datos

```bash
sqlite3 data/mediacheky.db "SELECT * FROM services;"
```

## Docker

### Desarrollo con Docker

```bash
# Construir
docker build -t mediacheky:dev .

# Ejecutar
docker run -p 8080:8080 \
  -v $(pwd)/data:/data \
  mediacheky:dev
```

### Producción

```bash
docker-compose up -d
```

## Solución de Problemas

### La aplicación no inicia

1. Verificar que el puerto 8080 esté disponible
2. Comprobar que Go 1.22+ esté instalado
3. Verificar permisos del directorio de datos

### Error de base de datos

```bash
# Eliminar la base de datos y reiniciar
rm -rf data/
mkdir data
```

### Templates no se cargan

Verificar que el directorio `web/templates` exista y contenga los archivos.

## Contribuir

1. Fork el proyecto
2. Crear una rama para tu feature
3. Hacer commit de los cambios
4. Push a la rama
5. Crear un Pull Request

## Recursos

- [Documentación de Fiber](https://docs.gofiber.io/)
- [Documentación de GORM](https://gorm.io/)
- [Alpine.js](https://alpinejs.dev/)
- [Tailwind CSS](https://tailwindcss.com/)

## Licencia

MIT License - Ver [LICENSE](LICENSE) para más detalles
