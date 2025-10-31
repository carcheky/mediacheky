# Alpine.js Components for File Health Management

Este documento describe todos los componentes Alpine.js reutilizables creados para la gestión de salud de archivos en KeeperCheky.

## 📁 Archivos

- **`web/static/js/file-health-components.js`**: Archivo principal con todos los componentes y helpers
- **`web/templates/pages/files-example.html`**: Página de demostración de componentes
- **URL de demostración**: `/files-example`

## 🧩 Componentes Disponibles

### 1. `healthCard(status, count, description)`

Card de estadística con acción que muestra un resumen de estado de salud.

**Parámetros:**
- `status` (string): Tipo de health status (ok, orphan_download, only_hardlink, dead_torrent, never_watched)
- `count` (number): Número de archivos en esta categoría
- `description` (string): Descripción corta del estado

**Propiedades computadas:**
- `icon`: Emoji del icono según el status
- `severity`: Nivel de severidad (ok, warning, critical)
- `colors`: Objeto con clases CSS de colores (border, bg, hoverBg, text, badge)
- `label`: Etiqueta legible del status

**Métodos:**
- `handleClick()`: Maneja el clic en el card (puede disparar evento para filtrar)

**Ejemplo de uso:**
```html
<div x-data="healthCard('orphan_download', 45, 'En qBT pero no en biblioteca')" 
    @click="handleClick()"
    :class="colors.border + ' ' + colors.bg + ' ' + colors.hoverBg"
    class="rounded-lg shadow-lg p-4 cursor-pointer transition-all border">
    <div class="flex items-center justify-between mb-2">
        <span class="text-3xl" x-text="icon"></span>
        <span class="text-2xl font-bold text-dark-text" x-text="count"></span>
    </div>
    <h3 class="text-sm font-semibold text-dark-text mb-1" x-text="label"></h3>
    <p class="text-xs text-dark-muted" x-text="description"></p>
</div>
```

---

### 2. `fileHealthCard(file, healthReport)`

Card individual de archivo con toda su información de salud y acciones disponibles.

**Parámetros:**
- `file` (object): Objeto MediaFileInfo con datos del archivo
- `healthReport` (object): Objeto FileHealthReport con status, issues, suggestions, actions

**Propiedades:**
- `expanded`: Boolean para mostrar/ocultar detalles
- `loading`: Boolean durante operaciones async
- `actionInProgress`: String con la acción en curso

**Propiedades computadas:**
- `status`: Estado de salud del archivo
- `severity`: Nivel de severidad
- `issues`: Array de problemas detectados
- `suggestions`: Array de sugerencias
- `actions`: Array de acciones disponibles
- `colors`: Objeto con clases CSS de colores

**Métodos:**
- `toggleDetails()`: Expande/colapsa detalles
- `executeAction(action)`: Ejecuta una acción (import, delete, ignore, etc.)
- `showConfirmation(action)`: Muestra diálogo de confirmación
- `formatSize(bytes)`: Formatea tamaño de archivo
- `formatDate(timestamp)`: Formatea fecha relativa

**Eventos:**
- `file-action-success`: Disparado cuando una acción se completa exitosamente
- `file-action-error`: Disparado cuando una acción falla

**Ejemplo de uso:**
```html
<div x-data="fileHealthCard(file, healthReport)"
    @file-action-success.window="handleSuccess($event.detail)"
    @file-action-error.window="handleError($event.detail)"
    :class="colors.border + ' ' + colors.bg"
    class="bg-dark-surface border rounded-lg shadow-lg p-6">
    <!-- Card content -->
</div>
```

---

### 3. `healthFilters()`

Sistema de filtros inteligentes para archivos de salud.

**Propiedades de estado:**
- `selectedProblem`: Filtro por tipo de problema (all, error, missing, orphan, low_ratio)
- `selectedService`: Filtro por servicio (all, jellyfin, radarr, sonarr, qbittorrent)
- `selectedSize`: Filtro por rango de tamaño (all, small, medium, large, huge)
- `selectedAge`: Filtro por antigüedad (all, recent, moderate, old, ancient)
- `searchQuery`: Query de búsqueda de texto

**Propiedades computadas:**
- `hasActiveFilters`: Boolean si hay filtros activos
- `activeFilterCount`: Número de filtros activos
- `problemOptions`: Array de opciones de problemas
- `serviceOptions`: Array de opciones de servicios
- `sizeOptions`: Array de opciones de tamaño
- `ageOptions`: Array de opciones de antigüedad

**Métodos:**
- `applyFilters()`: Aplica todos los filtros activos
- `clearFilters()`: Limpia todos los filtros
- `getFilteredFiles(files)`: Retorna array de archivos filtrados

**Eventos:**
- `filters-changed`: Disparado cuando los filtros cambian (detail contiene todos los valores)

**Ejemplo de uso:**
```html
<div x-data="healthFilters()" 
    @filters-changed="handleFiltersChanged($event.detail)">
    <select x-model="selectedProblem" @change="applyFilters()">
        <template x-for="option in problemOptions" :key="option.value">
            <option :value="option.value" x-text="option.label"></option>
        </template>
    </select>
    <!-- Más controles de filtro -->
</div>
```

---

### 4. `bulkActions(initialSelectedFiles = [])`

Gestión de selección múltiple y acciones masivas.

**Parámetros:**
- `initialSelectedFiles` (array): Array inicial de archivos seleccionados

**Propiedades:**
- `selectedFiles`: Array de archivos seleccionados
- `actionInProgress`: Boolean durante ejecución
- `progressCurrent`: Número de archivos procesados
- `progressTotal`: Total de archivos a procesar

**Propiedades computadas:**
- `selectedCount`: Número de archivos seleccionados
- `hasSelection`: Boolean si hay archivos seleccionados
- `progressPercent`: Porcentaje de progreso (0-100)

**Métodos:**
- `toggleSelectAll(files)`: Selecciona/deselecciona todos los archivos
- `isAllSelected(files)`: Verifica si todos están seleccionados
- `toggleSelect(file)`: Selecciona/deselecciona un archivo
- `isSelected(file)`: Verifica si un archivo está seleccionado
- `executeBulkAction(action)`: Ejecuta acción en masa
- `confirmBulkAction(action)`: Muestra confirmación

**Eventos:**
- `selection-changed`: Disparado cuando la selección cambia
- `bulk-action-complete`: Disparado cuando una acción masiva termina

**Ejemplo de uso:**
```html
<div x-data="bulkActions([])"
    @bulk-action-complete.window="handleComplete($event.detail)">
    <label class="flex items-center gap-2">
        <input type="checkbox" 
            :checked="isAllSelected(files)" 
            @change="toggleSelectAll(files)">
        <span>Seleccionar todos</span>
    </label>
    
    <button @click="executeBulkAction('delete')"
        :disabled="!hasSelection || actionInProgress">
        Eliminar Seleccionados
    </button>
    
    <!-- Progress bar -->
    <div x-show="actionInProgress">
        <div :style="'width: ' + progressPercent + '%'"></div>
    </div>
</div>
```

---

### 5. `healthStatusBadge(status, severity, size = 'md')`

Badge reutilizable para mostrar estado de salud.

**Parámetros:**
- `status` (string): Tipo de status
- `severity` (string): ok, warning, critical
- `size` (string): sm, md, lg

**Propiedades computadas:**
- `icon`: Emoji del icono
- `label`: Etiqueta legible
- `colors`: Objeto con clases CSS de colores
- `sizeClasses`: Clases CSS de tamaño

**Ejemplo de uso:**
```html
<div x-data="healthStatusBadge('orphan_download', 'warning', 'md')" 
    :class="colors.badge + ' ' + sizeClasses"
    class="inline-flex items-center gap-1 border rounded">
    <span x-text="icon"></span>
    <span x-text="label"></span>
</div>
```

---

### 6. `serviceStatusIndicator(service, isActive, details = {})`

Indicador de estado de servicio individual.

**Parámetros:**
- `service` (string): Nombre del servicio (radarr, sonarr, jellyfin, qbittorrent, etc.)
- `isActive` (boolean): Si el archivo está en este servicio
- `details` (object): Detalles adicionales (torrent state, ratio, etc.)

**Propiedades computadas:**
- `icon`: Emoji del servicio
- `label`: Etiqueta del servicio
- `statusColor`: Color según isActive
- `tooltip`: Texto del tooltip con detalles

**Ejemplo de uso:**
```html
<div x-data="serviceStatusIndicator('qbittorrent', true, { 
        torrent_state: 'seeding', 
        seed_ratio: 2.5, 
        is_seeding: true 
    })" 
    :class="statusColor"
    class="flex items-center gap-2 px-3 py-2"
    :title="tooltip">
    <span x-text="icon"></span>
    <span x-text="label"></span>
</div>
```

---

## 🛠️ Funciones Helper

### `formatBytes(bytes)`

Formatea bytes a tamaño legible.

**Parámetros:**
- `bytes` (number): Tamaño en bytes

**Retorna:** String (ej: "15.2 GB")

**Ejemplo:**
```javascript
formatBytes(16318189568) // "15.2 GB"
```

---

### `formatDate(timestamp)`

Formatea timestamp a fecha relativa.

**Parámetros:**
- `timestamp` (string|Date): Fecha a formatear

**Retorna:** String (ej: "Hace 3 días")

**Ejemplo:**
```javascript
formatDate(new Date(Date.now() - 86400000 * 3)) // "Hace 3 días"
```

---

### `getStatusIcon(status)`

Retorna emoji de icono según status.

**Parámetros:**
- `status` (string): Código de status

**Retorna:** String (emoji)

**Ejemplo:**
```javascript
getStatusIcon('orphan_download') // "⚠️"
```

---

### `getSeverityColor(severity)`

Retorna objeto con clases CSS de color según severidad.

**Parámetros:**
- `severity` (string): ok, warning, critical

**Retorna:** Object con border, bg, hoverBg, text, badge

**Ejemplo:**
```javascript
getSeverityColor('warning')
// {
//   border: 'border-yellow-600/30',
//   bg: 'bg-yellow-900/20',
//   hoverBg: 'hover:bg-yellow-900/30',
//   text: 'text-yellow-400',
//   badge: 'bg-yellow-900/40 border-yellow-600/50 text-yellow-300'
// }
```

---

### `getStatusLabel(status)`

Retorna etiqueta legible para un status.

**Parámetros:**
- `status` (string): Código de status

**Retorna:** String

**Ejemplo:**
```javascript
getStatusLabel('orphan_download') // "Huérfano en Descargas"
```

---

### `isOrphanDownload(file)`

Verifica si un archivo es un huérfano en descargas (en qBittorrent pero no en ningún servicio de gestión de medios).

**Parámetros:**
- `file` (object): Objeto de archivo a verificar

**Retorna:** Boolean

**Ejemplo:**
```javascript
const file = {
    in_qbittorrent: true,
    in_jellyfin: false,
    in_radarr: false,
    in_sonarr: false
};
isOrphanDownload(file) // true
```

**Ejemplo:**
```javascript
getStatusLabel('orphan_download') // "Huérfano en Descargas"
```

---

## 🎨 Códigos de Status

### Health Status
- `ok`: Archivo saludable
- `orphan_download`: Huérfano en descargas (en qBT pero no en biblioteca)
- `only_hardlink`: Solo queda hardlink (torrent eliminado)
- `dead_torrent`: Torrent muerto (error o sin seeds)
- `never_watched`: Nunca reproducido
- `missing_metadata`: Sin metadata
- `critical`: Estado crítico
- `warning`: Necesita atención
- `unclassified`: Sin clasificar

### Severity Levels
- `ok`: Todo correcto (verde)
- `warning`: Necesita atención (amarillo)
- `critical`: Problema crítico (rojo)

### Services
- `radarr`: Radarr (películas)
- `sonarr`: Sonarr (series)
- `jellyfin`: Jellyfin (biblioteca)
- `jellyseerr`: Jellyseerr (solicitudes)
- `qbittorrent`: qBittorrent (torrents)
- `jellystat`: Jellystat (estadísticas)

---

## 🔌 Integración con API

Los componentes están diseñados para integrarse con los siguientes endpoints:

### GET `/api/files/health`
Obtiene datos de salud de archivos con análisis.

**Response:**
```json
{
  "summary": {
    "healthy": 123,
    "orphan_downloads": 45,
    "only_hardlinks": 78,
    "dead_torrents": 12,
    "never_watched": 234
  },
  "files": [
    {
      "file": { /* MediaFileInfo */ },
      "health": {
        "status": "orphan_download",
        "severity": "warning",
        "issues": ["No está en Jellyfin"],
        "suggestions": ["Importar a Radarr"],
        "actions": ["import_radarr", "delete", "ignore"]
      }
    }
  ]
}
```

### POST `/api/files/:id/import-radarr`
Importa archivo a Radarr.

### POST `/api/files/:id/import-sonarr`
Importa archivo a Sonarr.

### DELETE `/api/files/:id`
Elimina archivo.

### POST `/api/files/:id/ignore`
Marca archivo como ignorado.

### POST `/api/files/:id/clean-hardlink`
Limpia hardlink.

---

## 🧪 Testing

Para probar los componentes:

1. Iniciar el servidor: `make dev` o `go run ./cmd/server`
2. Visitar: `http://localhost:8000/files-example`
3. Interactuar con cada sección de demostración

La página de ejemplo incluye:
1. Health Cards con diferentes estados
2. Health Status Badges en varios tamaños
3. Service Status Indicators
4. Filtros interactivos
5. File Health Card completo
6. Bulk Actions con barra de progreso
7. Demostración de helpers

---

## 📝 Convenciones de Código

### Nomenclatura
- Componentes en camelCase: `healthCard()`, `fileHealthCard()`
- Propiedades computadas con `get`: `get colors()`, `get icon()`
- Métodos en camelCase: `executeAction()`, `toggleDetails()`
- Eventos en kebab-case: `file-action-success`, `filters-changed`

### Estructura de Componentes
```javascript
function componentName(params) {
    return {
        // 1. Propiedades de estado
        prop1: value1,
        prop2: value2,
        
        // 2. init() si es necesario
        init() {
            // Inicialización
        },
        
        // 3. Propiedades computadas
        get computed1() {
            return this.prop1 * 2;
        },
        
        // 4. Métodos
        method1() {
            // Lógica
        },
        
        // 5. Métodos async
        async asyncMethod() {
            // Lógica async
        }
    };
}
```

### Manejo de Errores
- Siempre usar try-catch en operaciones async
- Disparar eventos para comunicar errores: `$dispatch('component-error', { error })`
- Mostrar loading states durante operaciones
- Incluir mensajes de error descriptivos

### Performance
- Evitar re-renders innecesarios con propiedades computadas
- Usar `x-show` para elementos que se ocultan frecuentemente
- Usar `x-if` para elementos que raramente se muestran
- Minimizar watchers y efectos secundarios

---

## 🚀 Futuras Mejoras

- [ ] Agregar tests unitarios con Alpine.js Testing Library
- [ ] Implementar drag & drop para reordenar archivos
- [ ] Agregar soporte para temas personalizados
- [ ] Crear más componentes auxiliares (modals, tooltips, etc.)
- [ ] Mejorar accesibilidad (ARIA labels, keyboard navigation)
- [ ] Agregar animaciones con Alpine.js transitions
- [ ] Implementar virtual scrolling para listas grandes
- [ ] Crear documentación interactiva con Storybook

---

## 📚 Referencias

- [Alpine.js Documentation](https://alpinejs.dev/)
- [Tailwind CSS Documentation](https://tailwindcss.com/docs)
- [Fiber v2 Documentation](https://docs.gofiber.io/)
- [Project Proposal](../../docs/PROPUESTA_MEJORA_FILES_UX.md)

---

**Última actualización**: 30 de octubre de 2025  
**Versión**: 1.0.0  
**Autor**: KeeperCheky Development Team
