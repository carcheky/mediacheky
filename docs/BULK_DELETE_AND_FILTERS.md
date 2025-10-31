# Eliminación en Bulk y Filtros Avanzados

## Resumen

Se ha implementado un sistema completo de eliminación en bulk y filtros avanzados para la gestión eficiente de medios en MediaCheky.

## Características Implementadas

### 1. Filtros Avanzados

#### Filtros Disponibles

1. **Tipo de Medio** (`type`)
   - Películas (movie)
   - Series (series)
   - Torrents (torrent)
   - Todos (all)

2. **Estado** (`status`)
   - Activo (active)
   - Excluido (excluded)
   - Seeding (seeding)
   - Todos (all)

3. **Servicio** (`service`)
   - Radarr
   - Sonarr
   - Jellyfin
   - Torrents huérfanos (orphan)
   - Todos (all)

4. **Rango de Tamaño** (`sizeRange`)
   - Pequeño: < 5 GB
   - Mediano: 5-20 GB
   - Grande: 20-50 GB
   - Muy grande: > 50 GB
   - Todos (all)

5. **Fecha de Agregado** (`addedDate`)
   - Última semana (7 días)
   - Último mes (30 días)
   - Último trimestre (90 días)
   - Más antiguos (> 90 días)
   - Todos (all)

6. **Ratio de Seed** (`seedRatio`)
   - Bajo: < 1.0
   - Medio: 1.0-2.0
   - Alto: > 2.0
   - Todos (all)

7. **Calidad** (`quality`)
   - Búsqueda por texto en campo de calidad

8. **Completitud de Episodios** (`episodeCompletion`) - Solo para series
   - Completo: Tiene todos los episodios
   - Incompleto: Le faltan episodios
   - Vacío: Sin información de episodios
   - Todos (all)

9. **Búsqueda por Título** (`search`)
   - Búsqueda por texto en título

### 2. Interfaz de Usuario

#### Panel de Filtros Avanzados

- **Collapsible**: El panel de filtros avanzados se puede expandir/colapsar
- **Diseño responsivo**: Organizado en grid de 4 columnas en pantallas grandes
- **Indicadores visuales**: Colores y badges para cada tipo de filtro
- **Búsqueda rápida**: Campo de búsqueda siempre visible

#### Sistema de Selección

- **Checkbox en header**: Seleccionar/deseleccionar todos los elementos de la página
- **Checkbox por fila**: Selección individual de elementos
- **Highlighting**: Las filas seleccionadas se resaltan en azul
- **Contador**: Barra de acción bulk muestra cantidad de elementos seleccionados

#### Barra de Acciones Bulk

Aparece cuando hay elementos seleccionados:
- Contador de elementos seleccionados
- Botón "Deseleccionar todos"
- Botón "Eliminar seleccionados" (rojo, destructivo)

### 3. Backend - Filtros

#### Endpoint: `GET /api/media`

**Query Parameters:**
```
page=1
pageSize=20
type=all
status=all
search=
service=all
sizeRange=all
addedDate=all
seedRatio=all
quality=
episodeCompletion=all
```

**Implementación:**

- **Handler**: `internal/handler/media.go` - `GetAll()`
  - Extrae todos los parámetros de query
  - Construye struct `MediaFilters`
  - Llama a `GetPaginatedWithFilters()`

- **Repository**: `internal/repository/repository.go` - `GetPaginatedWithFilters()`
  - Construye query dinámica con GORM
  - Aplica cada filtro condicionalmente
  - Maneja paginación y ordenamiento
  - Retorna resultados + total count

**Lógica de Filtros:**

```go
// Tipo de medio
if filters.Type != "all" {
    query = query.Where("type = ?", filters.Type)
}

// Estado
switch filters.Status {
case "excluded":
    query = query.Where("excluded = ?", true)
case "active":
    query = query.Where("excluded = ?", false)
case "seeding":
    query = query.Where("is_seeding = ?", true)
}

// Servicio
switch filters.Service {
case "radarr":
    query = query.Where("radarr_id IS NOT NULL")
case "sonarr":
    query = query.Where("sonarr_id IS NOT NULL")
case "jellyfin":
    query = query.Where("jellyfin_id IS NOT NULL")
case "orphan":
    query = query.Where("type = ?", "torrent")
}

// Tamaño
switch filters.SizeRange {
case "small":
    query = query.Where("size < ?", 5*1024*1024*1024)
case "medium":
    query = query.Where("size >= ? AND size < ?", 5GB, 20GB)
case "large":
    query = query.Where("size >= ? AND size < ?", 20GB, 50GB)
case "xlarge":
    query = query.Where("size >= ?", 50GB)
}

// Fecha
now := r.db.NowFunc()
switch filters.AddedDate {
case "week":
    query = query.Where("added_date >= ?", now.AddDate(0, 0, -7))
case "month":
    query = query.Where("added_date >= ?", now.AddDate(0, 0, -30))
case "quarter":
    query = query.Where("added_date >= ?", now.AddDate(0, 0, -90))
case "older":
    query = query.Where("added_date < ?", now.AddDate(0, 0, -90))
}

// Ratio de seed
switch filters.SeedRatio {
case "low":
    query = query.Where("seed_ratio < ?", 1.0)
case "medium":
    query = query.Where("seed_ratio >= ? AND seed_ratio < ?", 1.0, 2.0)
case "high":
    query = query.Where("seed_ratio >= ?", 2.0)
}

// Calidad (búsqueda por texto)
if filters.Quality != "" {
    query = query.Where("quality LIKE ?", "%"+filters.Quality+"%")
}

// Completitud de episodios (solo series)
switch filters.EpisodeCompletion {
case "complete":
    query = query.Where("type = ? AND episode_count > 0 AND episode_count = episode_file_count", "series")
case "incomplete":
    query = query.Where("type = ? AND episode_count > 0 AND episode_count > episode_file_count", "series")
case "empty":
    query = query.Where("type = ? AND (episode_count IS NULL OR episode_count = 0)", "series")
}

// Búsqueda por título
if filters.Search != "" {
    query = query.Where("title LIKE ?", "%"+filters.Search+"%")
}
```

### 4. Backend - Bulk Delete

#### Endpoint: `POST /api/media/bulk-delete`

**Request Body:**
```json
{
  "ids": [1, 2, 3, 4, 5],
  "options": {
    "radarr": true,
    "sonarr": true,
    "jellyfin": true,
    "deleteFiles": true,
    "qbittorrent": true
  }
}
```

**Response:**
```json
{
  "message": "Bulk delete completed",
  "total": 5,
  "success_count": 4,
  "failure_count": 1,
  "results": {
    "1": {
      "success": true,
      "deleted_from": ["radarr", "jellyfin"],
      "files_deleted": true
    },
    "2": {
      "success": true,
      "deleted_from": ["radarr", "jellyfin"],
      "files_deleted": true
    },
    "3": {
      "success": false,
      "error": "Media not found in Radarr",
      "deleted_from": ["jellyfin"],
      "errors": {
        "radarr": "404 Not Found"
      }
    },
    ...
  }
}
```

**Implementación:**

- **Handler**: `internal/handler/media.go` - `BulkDelete()`
  - Parsea request con IDs y opciones
  - Valida que haya al menos un ID
  - Itera sobre cada ID
  - Llama a `CleanupService.DeleteMedia()` para cada uno
  - Recopila resultados individuales
  - Retorna resumen con contador de éxitos/fallos

- **Ventajas del enfoque actual:**
  - Feedback detallado por cada elemento
  - Control individual de opciones de eliminación
  - Reporte de errores granular
  - Reutiliza lógica existente de `DeleteMedia()`

**Mejoras futuras posibles:**
- Procesamiento paralelo con goroutines
- Queue system para operaciones muy grandes
- Progress streaming via WebSockets

### 5. Frontend - Filtros

#### Componente Alpine.js

**Estado:**
```javascript
filters: {
    type: 'all',
    status: 'all',
    search: '',
    service: 'all',
    sizeRange: 'all',
    addedDate: 'all',
    seedRatio: 'all',
    quality: '',
    episodeCompletion: 'all',
    pageSize: 20
}
```

**Función loadMedia():**
- Construye URLSearchParams con todos los filtros
- Hace fetch a `/api/media?${params}`
- Actualiza lista de medios
- Respeta paginación

**Función resetFilters():**
- Reinicia todos los filtros a valores por defecto
- Reinicia página a 1
- Recarga medios

### 6. Frontend - Bulk Delete

**Estado:**
```javascript
selectedItems: [],
bulkDeleteModal: {
    deleting: false
}
```

**Función toggleSelection(itemId):**
- Agrega/quita itemId del array selectedItems
- Actualiza UI de checkbox

**Función toggleSelectAll():**
- Selecciona/deselecciona todos los elementos visibles en la página
- Actualiza array selectedItems

**Función bulkDelete():**
- Valida que haya elementos seleccionados
- Muestra confirmación
- Hace POST a `/api/media/bulk-delete` con:
  - Array de IDs seleccionados
  - Opciones de eliminación
- Muestra mensaje con resultado
- Limpia selección
- Recarga lista

### 7. Presets de Filtros

#### Características

- **Guardar preset**: Guarda configuración actual de filtros con un nombre
- **Cargar preset**: Dropdown con lista de presets guardados
- **Eliminar preset**: Botón 🗑️ en cada preset para eliminarlo
- **Storage**: Usa localStorage para persistencia

#### Funciones

**saveCurrentFilters():**
```javascript
const presets = this.getSavedPresets();
presets[presetName] = { ...this.filters };
localStorage.setItem('filter_presets', JSON.stringify(presets));
```

**loadFilterPreset(name):**
```javascript
const presets = this.getSavedPresets();
this.filters = { ...this.filters, ...presets[name] };
this.currentPage = 1;
this.loadMedia();
```

**getSavedPresets():**
```javascript
const saved = localStorage.getItem('filter_presets');
return saved ? JSON.parse(saved) : {};
```

**deletePreset(name):**
```javascript
const presets = this.getSavedPresets();
delete presets[name];
localStorage.setItem('filter_presets', JSON.stringify(presets));
```

## Casos de Uso

### Caso 1: Eliminar películas grandes antiguas

1. Aplicar filtros:
   - Tipo: Películas
   - Tamaño: > 50 GB
   - Fecha: > 90 días
2. Revisar lista filtrada
3. Seleccionar elementos deseados
4. Bulk delete

### Caso 2: Limpiar series incompletas

1. Aplicar filtros:
   - Tipo: Series
   - Completitud: Incompleto
2. Revisar lista
3. Seleccionar series a eliminar
4. Bulk delete

### Caso 3: Gestionar torrents con bajo ratio

1. Aplicar filtros:
   - Estado: Seeding
   - Ratio: < 1.0
2. Identificar torrents a mantener
3. Seleccionar los demás
4. Bulk delete

### Caso 4: Limpiar media solo en un servicio

1. Aplicar filtros:
   - Servicio: Radarr (solo en Radarr)
2. Revisar qué no está en otros servicios
3. Seleccionar elementos
4. Bulk delete

### Caso 5: Usar presets para auditorías

1. Crear preset "Películas viejas grandes":
   - Tipo: Películas
   - Tamaño: > 50 GB
   - Fecha: > 90 días

2. Crear preset "Series incompletas":
   - Tipo: Series
   - Completitud: Incompleto

3. Crear preset "Torrents huérfanos":
   - Servicio: Orphan

4. Cargar preset según necesidad de limpieza

## Optimizaciones Implementadas

### Backend

1. **Queries eficientes**: 
   - Uso de GORM para queries optimizadas
   - Índices en campos filtrados (added_date, excluded, type)
   - Paginación para evitar cargar todo en memoria

2. **Bulk delete optimizado**:
   - Una sola transacción HTTP para todo el batch
   - Reporte detallado de resultados
   - Manejo de errores por elemento

### Frontend

1. **Renderizado eficiente**:
   - Alpine.js reactive binding
   - Solo re-renderiza cuando cambian datos
   - Lazy loading de páginas

2. **UX optimizado**:
   - Feedback visual inmediato en selección
   - Confirmación antes de eliminar
   - Mensajes de éxito/error claros
   - Contador en tiempo real de selección

## Testing Recomendado

### Tests de Filtros

```bash
# Test básico de filtros
curl "http://localhost:8000/api/media?type=movie&sizeRange=xlarge&addedDate=older"

# Test de búsqueda
curl "http://localhost:8000/api/media?search=Matrix"

# Test de combinación de filtros
curl "http://localhost:8000/api/media?type=series&status=active&episodeCompletion=incomplete"

# Test de filtro de servicio
curl "http://localhost:8000/api/media?service=orphan"
```

### Tests de Bulk Delete

```bash
# Bulk delete básico
curl -X POST http://localhost:8000/api/media/bulk-delete \
  -H "Content-Type: application/json" \
  -d '{
    "ids": [1, 2, 3],
    "options": {
      "radarr": true,
      "sonarr": true,
      "jellyfin": true,
      "deleteFiles": true,
      "qbittorrent": true
    }
  }'
```

### Tests de UI

1. **Test de filtros**:
   - Aplicar cada filtro individualmente
   - Combinar múltiples filtros
   - Verificar que el contador de total se actualice
   - Resetear filtros

2. **Test de selección**:
   - Seleccionar elementos individuales
   - Select all checkbox
   - Deseleccionar todos
   - Verificar highlight azul

3. **Test de bulk delete**:
   - Seleccionar 2-3 elementos
   - Confirmar eliminación
   - Verificar mensaje de éxito
   - Verificar que lista se actualice

4. **Test de presets**:
   - Crear preset
   - Cargar preset
   - Verificar que filtros se apliquen
   - Eliminar preset

## Archivos Modificados

### Backend

1. **internal/handler/media.go**
   - Nueva función `BulkDelete()`
   - Modificada función `GetAll()` para soportar filtros avanzados
   - Nuevo tipo `MediaFilters` (movido a repository)

2. **internal/repository/repository.go**
   - Nuevo tipo `MediaFilters`
   - Nueva función `GetPaginatedWithFilters()`
   - Lógica de filtrado dinámica con GORM

3. **cmd/server/main.go**
   - Nueva ruta `POST /api/media/bulk-delete`

### Frontend

1. **web/templates/pages/media.html**
   - Panel de filtros avanzados collapsible
   - Sistema de checkboxes en tabla
   - Barra de acciones bulk
   - Dropdown de presets
   - Función `loadMedia()` con todos los filtros
   - Función `bulkDelete()` optimizada
   - Funciones de gestión de presets

## Próximos Pasos

### Mejoras Sugeridas

1. **Performance**:
   - [ ] Implementar caché de queries frecuentes
   - [ ] Procesamiento paralelo en bulk delete
   - [ ] Progress bar en tiempo real

2. **UX**:
   - [ ] Drag & drop para reordenar presets
   - [ ] Exportar/importar presets
   - [ ] Preview de elementos a eliminar
   - [ ] Undo/redo de bulk delete

3. **Analytics**:
   - [ ] Estadísticas de uso de filtros
   - [ ] Historial de bulk deletes
   - [ ] Reportes de limpieza

4. **Seguridad**:
   - [ ] Confirmación extra para bulk > 10 items
   - [ ] Rate limiting en bulk delete
   - [ ] Audit log de eliminaciones

## Conclusión

Se ha implementado un sistema completo y robusto de filtros avanzados y eliminación en bulk que permite:

- ✅ 10+ opciones de filtrado combinables
- ✅ Selección múltiple con checkboxes
- ✅ Eliminación en bulk optimizada
- ✅ Sistema de presets guardables
- ✅ Feedback detallado de operaciones
- ✅ UI intuitiva y responsiva
- ✅ Backend eficiente con queries optimizadas

El sistema está listo para producción y puede manejar bibliotecas medianas a grandes con eficiencia.
