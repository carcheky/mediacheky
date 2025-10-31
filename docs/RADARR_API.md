# Radarr API Integration

Este documento describe la integración completa con la API de Radarr v3 en KeeperCheky.

## Endpoints Disponibles

### 1. System Information
**Endpoint:** `GET /api/radarr/system`

Obtiene información completa del sistema Radarr.

**Respuesta:**
```json
{
  "version": "4.0.0.0",
  "build_time": "2024-01-01T00:00:00Z",
  "branch": "master",
  "os": "Linux",
  "os_version": "5.15.0",
  "runtime": ".NET Core",
  "runtime_version": "6.0.0",
  "is_debug": false,
  "is_production": true,
  "authentication": "forms",
  "url_base": "",
  "startup_path": "/app",
  "app_data": "/config",
  "sqlite_version": "3.36.0"
}
```

**Uso:**
- Se muestra automáticamente en Settings cuando se prueba la conexión a Radarr
- Útil para verificar la versión y configuración del servidor

### 2. Download Queue
**Endpoint:** `GET /api/radarr/queue`

Obtiene la cola de descargas actual con información de progreso.

**Respuesta:**
```json
{
  "total": 2,
  "items": [
    {
      "id": 1,
      "movie_id": 123,
      "title": "Test Movie 2024",
      "size": 1000000000,
      "size_left": 500000000,
      "progress": 50.0,
      "status": "downloading",
      "download_status": "ok",
      "download_state": "downloading",
      "protocol": "torrent",
      "download_client": "qBittorrent",
      "indexer": "Example Indexer",
      "timed_out": false,
      "estimated_completion": "2024-01-01T12:00:00Z"
    }
  ]
}
```

**Uso:**
- Se muestra en el Dashboard con actualización automática cada 30 segundos
- Incluye barra de progreso visual
- Muestra detalles del cliente de descarga, protocolo e indexer

### 3. History
**Endpoint:** `GET /api/radarr/history?pageSize=50`

Obtiene el historial de eventos de Radarr.

**Parámetros de Query:**
- `pageSize` (opcional): Número de resultados (1-100, default: 50)

**Respuesta:**
```json
{
  "total": 2,
  "items": [
    {
      "id": 1,
      "movie_id": 123,
      "source_title": "Test.Movie.2024.1080p.BluRay.x264",
      "quality": "Bluray-1080p",
      "date": "2024-01-01T10:00:00Z",
      "event_type": "grabbed",
      "download_id": "abc123"
    },
    {
      "id": 2,
      "movie_id": 123,
      "source_title": "Test.Movie.2024.1080p.BluRay.x264",
      "quality": "Bluray-1080p",
      "date": "2024-01-01T11:00:00Z",
      "event_type": "downloadFolderImported",
      "download_id": "abc123"
    }
  ]
}
```

**Tipos de Eventos:**
- `grabbed`: Descarga iniciada
- `downloadFolderImported`: Archivo importado exitosamente
- `downloadFailed`: Descarga fallida
- `movieFileDeleted`: Archivo eliminado
- `movieFileRenamed`: Archivo renombrado

**Uso:**
- Útil para auditoría y debugging
- Permite rastrear el ciclo completo de una descarga

### 4. Calendar
**Endpoint:** `GET /api/radarr/calendar?start=2024-01-01&end=2024-01-31`

Obtiene películas próximas a estrenarse.

**Parámetros de Query:**
- `start` (opcional): Fecha de inicio en formato YYYY-MM-DD (default: hoy)
- `end` (opcional): Fecha de fin en formato YYYY-MM-DD (default: +30 días)

**Respuesta:**
```json
{
  "start_date": "2024-01-01",
  "end_date": "2024-01-31",
  "total": 2,
  "items": [
    {
      "id": 1,
      "title": "Upcoming Movie",
      "in_cinemas": "2024-01-15T00:00:00Z",
      "physical_release": "2024-02-01T00:00:00Z",
      "digital_release": "2024-01-20T00:00:00Z",
      "year": 2024,
      "has_file": false,
      "monitored": true
    }
  ]
}
```

**Uso:**
- Ver próximos estrenos
- Planificar espacio en disco
- Identificar películas monitoreadas que aún no tienen archivo

### 5. Quality Profiles
**Endpoint:** `GET /api/radarr/quality-profiles`

Obtiene los perfiles de calidad configurados en Radarr.

**Respuesta:**
```json
{
  "total": 5,
  "profiles": [
    {
      "id": 1,
      "name": "Any"
    },
    {
      "id": 4,
      "name": "HD-1080p"
    },
    {
      "id": 5,
      "name": "Ultra-HD"
    }
  ]
}
```

**Uso:**
- Útil para configuración
- Puede usarse para filtrar o categorizar películas por perfil de calidad

## Integración en la UI

### Dashboard
La página del Dashboard incluye una sección de "Radarr Download Queue" que muestra:
- Lista de descargas activas
- Barra de progreso con porcentaje
- Tamaño descargado vs. tamaño total
- Cliente de descarga, protocolo e indexer
- Estado actual de cada descarga
- Actualización automática cada 30 segundos

### Settings
Cuando se prueba la conexión a Radarr:
- Se muestra información completa del sistema
- Versión, rama, OS, runtime
- Rutas de configuración
- Estado de autenticación

## Funciones del Cliente Radarr

### GetSystemInfo
```go
info, err := radarrClient.GetSystemInfo(ctx)
if err != nil {
    // Manejar error
}
// info contiene toda la información del sistema
```

### GetQueue
```go
queue, err := radarrClient.GetQueue(ctx)
if err != nil {
    // Manejar error
}
// queue es un slice de RadarrQueueItem
for _, item := range queue {
    fmt.Printf("Descargando: %s (%.1f%%)\n", item.Title, item.Progress)
}
```

### GetHistory
```go
history, err := radarrClient.GetHistory(ctx, 50)
if err != nil {
    // Manejar error
}
// history es un slice de RadarrHistoryItem
```

### GetCalendar
```go
startDate := time.Now()
endDate := time.Now().AddDate(0, 0, 30)
calendar, err := radarrClient.GetCalendar(ctx, startDate, endDate)
if err != nil {
    // Manejar error
}
// calendar es un slice de RadarrCalendarItem
```

### GetQualityProfiles
```go
profiles, err := radarrClient.GetQualityProfiles(ctx)
if err != nil {
    // Manejar error
}
// profiles es un slice de RadarrQualityProfile
```

## Manejo de Errores

Todos los endpoints devuelven errores apropiados:

- **503 Service Unavailable**: Radarr no está configurado
- **401 Unauthorized**: API Key inválida
- **500 Internal Server Error**: Error al comunicarse con Radarr
- **Timeout**: Después de reintentos con backoff exponencial

## Retry Logic

Todas las llamadas a la API de Radarr implementan lógica de reintentos:
- Máximo 3 intentos
- Backoff exponencial (1s, 2s, 4s)
- Respeta cancelación de contexto

## Tests

Los tests unitarios cubren:
- ✅ GetSystemInfo con datos completos
- ✅ GetQueue con múltiples items
- ✅ GetQueue con cola vacía
- ✅ GetHistory con eventos
- ✅ GetCalendar con filtro de fechas
- ✅ GetQualityProfiles
- ✅ Manejo de errores (401, 500, timeout)

Ejecutar tests:
```bash
go test ./internal/service/clients -v -run TestRadarr*
```

## Configuración

Asegúrate de tener Radarr configurado en `config.yaml`:

```yaml
clients:
  radarr:
    enabled: true
    url: "http://localhost:7878"
    api_key: "tu-api-key-aqui"
```

O mediante variables de entorno:
```bash
KEEPERCHEKY_CLIENTS_RADARR_ENABLED=true
KEEPERCHEKY_CLIENTS_RADARR_URL=http://localhost:7878
KEEPERCHEKY_CLIENTS_RADARR_API_KEY=tu-api-key-aqui
```

## Referencia de la API Oficial

Documentación oficial de Radarr API v3:
- https://radarr.video/docs/api/
- https://wiki.servarr.com/radarr

## Próximas Mejoras

Pendientes de implementar:
- [ ] Mostrar historial reciente en el Dashboard
- [ ] Integrar calendar para ver próximos estrenos
- [ ] Mostrar quality profiles en Files para categorización
- [ ] Permitir cambiar quality profile desde la UI
- [ ] Estadísticas de descargas completadas vs. fallidas
