# 🎯 Propuesta de Mejora UX - Pestaña Files

**Fecha**: 27 de octubre de 2025  
**Contexto**: Aplicación para limpieza de discos duros de media  
**Objetivo**: Hacer la pestaña Files más intuitiva y orientada a la acción

---

## 📋 Análisis de la Situación Actual

### ¿Qué hace la pestaña Files actualmente?

1. **Escanea archivos físicos** del disco duro desde:
   - Carpeta de descargas completadas de qBittorrent (origen)
   - Bibliotecas de Jellyfin (destino)

2. **Enriquece con datos** de múltiples servicios:
   - **Radarr**: Películas gestionadas, calidad, metadata
   - **Sonarr**: Series/episodios gestionados, temporadas
   - **Jellyfin**: Biblioteca de medios, visionados
   - **Jellyseerr**: Solicitudes de usuarios
   - **Jellystat**: Estadísticas de reproducción
   - **qBittorrent**: Estado de torrents, seeds, ratios

3. **Organiza** por:
   - Tabs: Todos, Películas, Series, Sin clasificar
   - Agrupación: Películas por título, Series por temporada
   - Detección de hardlinks (archivos duplicados que no ocupan espacio extra)

### ❌ Problemas de Usabilidad Identificados

#### 1. **Falta de Propósito Claro**
- No queda claro **para qué sirve** esta pestaña
- No se diferencia claramente de la pestaña "Media"
- El usuario no sabe **qué acción tomar** al ver esta información

#### 2. **Información Densa y Técnica**
- Rutas completas de archivos (muy largas)
- Badges de servicios sin contexto claro
- Tooltips con información técnica (hash, estado torrent)
- No destaca información crítica para **decisiones de limpieza**

#### 3. **Sin Indicadores de Acción**
- No hay alertas visuales para archivos problemáticos
- No se resalta qué archivos **se pueden eliminar de forma segura**
- No hay sugerencias de acción

#### 4. **Navegación Confusa**
- Las "Rutas monitorizadas" están colapsadas por defecto
- No se entiende la diferencia entre "Origen" y "Destino"
- La agrupación por hardlinks no es intuitiva

---

## ✅ Propuestas de Mejora

### 🎯 Propuesta 1: **Enfoque en "Salud del Sistema"** (RECOMENDADA)

Transformar la pestaña Files en un **Dashboard de Salud del Almacenamiento** que muestre el estado de los archivos y guíe al usuario hacia acciones de limpieza.

#### Cambios Principales:

1. **Nuevo Título y Descripción**
   ```
   🏥 Salud del Almacenamiento
   Detecta y resuelve problemas en tus archivos de media
   ```

2. **Tarjetas de Estado (Cards)**
   Reemplazar las estadísticas actuales por indicadores accionables:
   
   ```
   ✅ Archivos Saludables (123)
      - En biblioteca y siendo compartidos
      - Tooltip: "Archivos correctamente gestionados"
   
   ⚠️ Huérfanos en Descargas (45)
      - En qBittorrent pero NO en biblioteca
      - Tooltip: "Archivos descargados que no están en Jellyfin"
      - Acción: "Revisar e importar"
   
   🔗 Solo Hardlinks (78)
      - Archivo original eliminado, solo queda hardlink
      - Tooltip: "Espacio puede liberarse eliminando duplicados"
      - Acción: "Consolidar"
   
   🚫 Torrents Muertos (12)
      - En biblioteca pero torrent pausado/error
      - Tooltip: "Archivos sin seed, pueden eliminarse"
      - Acción: "Eliminar seguros"
   
   📊 Sin Reproducir (234)
      - Nunca visto según Jellystat
      - Tooltip: "Candidatos para limpieza"
      - Acción: "Revisar antiguos"
   ```

3. **Vista por Categorías de Acción** (en vez de tabs)
   
   Tabs actuales → **Categorías accionables**:
   
   - **🟢 Todo OK** (Archivos bien gestionados)
   - **🟡 Necesitan Atención** (Huérfanos, hardlinks, etc.)
   - **🔴 Problemas Críticos** (Torrents muertos, archivos corruptos)
   - **📦 Por Clasificar** (Sin metadata de servicios)

4. **Cards de Archivo Mejoradas**

   En vez de mostrar solo la ruta, mostrar:
   
   ```
   ┌─────────────────────────────────────────────────────┐
   │ 🎬 Inception (2010)                                 │
   │                                                     │
   │ Estado: ⚠️ Huérfano en Descargas                   │
   │                                                     │
   │ Ubicación:                                          │
   │ 📥 /downloads/Inception.2010.mkv (15.2 GB)         │
   │                                                     │
   │ Servicios:                                          │
   │ ✅ qBittorrent (Seeding, ratio 2.5)                │
   │ ❌ No en Jellyfin                                   │
   │ ❌ No en Radarr                                     │
   │                                                     │
   │ Sugerencia:                                         │
   │ "Este archivo está siendo compartido pero no está  │
   │  en tu biblioteca. ¿Deseas importarlo a Radarr?"  │
   │                                                     │
   │ [Importar a Radarr] [Eliminar] [Ignorar]          │
   └─────────────────────────────────────────────────────┘
   ```

5. **Filtros Inteligentes**
   
   Barra de filtros superior con opciones:
   - **Por Problema**: Huérfanos, Sin seeds, Duplicados, Sin ver
   - **Por Servicio**: Solo en qBT, Solo en Jellyfin, En ambos
   - **Por Tamaño**: > 20GB, 10-20GB, < 10GB
   - **Por Antigüedad**: > 1 año sin ver, > 6 meses, etc.

6. **Acciones Masivas**
   
   ```
   [✓ Seleccionar todos] [Acciones ▼]
   
   Acciones disponibles:
   - 🗑️ Eliminar seleccionados (solo si seguro)
   - 📥 Importar a Radarr/Sonarr
   - 🏷️ Etiquetar como "No eliminar"
   - 🔗 Convertir hardlinks a copias
   - ⏸️ Pausar torrents
   ```

---

### 🎯 Propuesta 2: **Vista de Flujo de Datos** (ALTERNATIVA)

Mostrar el flujo de archivos desde el origen (qBittorrent) hasta el destino (Jellyfin) de forma visual.

#### Diseño:

```
┌──────────────────────────────────────────────────────────┐
│  FLUJO DE ARCHIVOS                                       │
├──────────────────────────────────────────────────────────┤
│                                                          │
│  [qBittorrent]  →  [Radarr/Sonarr]  →  [Jellyfin]      │
│   Downloads         Gestión             Biblioteca       │
│                                                          │
│  ● 234 archivos    ● 189 gestionados   ● 189 visibles  │
│  ⚠️ 45 huérfanos                       ● 23 vistos     │
│                                         ⚠️ 166 sin ver  │
│                                                          │
└──────────────────────────────────────────────────────────┘

ARCHIVOS POR ETAPA:

📥 Solo en Descargas (45)
   ├─ Inception (2010) - 15.2 GB
   │  Acción: [Importar] [Eliminar]
   └─ ...

🔄 En Gestión (Radarr/Sonarr) pero no en Biblioteca (0)
   └─ (Ninguno - Todo OK)

✅ En Biblioteca (189)
   ├─ Con seed activo (167) 
   └─ Sin seed (22) ⚠️
      Acción: [Eliminar seguros]

📊 Nunca Reproducidos (166)
   └─ Filtrar por antigüedad...
```

---

### 🎯 Propuesta 3: **Asistente de Limpieza Guiada** (COMPLEMENTARIA)

Agregar un modo de "Asistente" que guíe paso a paso:

```
┌─────────────────────────────────────────────────┐
│ 🧙 Asistente de Limpieza                       │
├─────────────────────────────────────────────────┤
│                                                 │
│ Paso 1/5: Huérfanos en Descargas              │
│                                                 │
│ Encontré 45 archivos descargados que no están  │
│ en tu biblioteca de Jellyfin.                   │
│                                                 │
│ ¿Qué deseas hacer?                             │
│                                                 │
│ ○ Importarlos automáticamente a Radarr/Sonarr │
│ ○ Revisarlos uno por uno                       │
│ ○ Omitir este paso                             │
│                                                 │
│ [Anterior] [Siguiente] [Salir del Asistente]  │
└─────────────────────────────────────────────────┘
```

---

## 🎨 Mockups Visuales de la Propuesta Recomendada

### Vista Principal - Salud del Almacenamiento

```
┌────────────────────────────────────────────────────────────────┐
│ 🏥 Salud del Almacenamiento                    [Sincronizar]  │
│ Detecta y resuelve problemas en tus archivos de media         │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│ ESTADO GENERAL                                                 │
│                                                                │
│ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐          │
│ │ ✅ 123   │ │ ⚠️ 45    │ │ 🔗 78    │ │ 📊 234   │          │
│ │ Saludables│ │ Huérfanos│ │ Hardlinks│ │ Sin Ver  │          │
│ │          │ │          │ │          │ │          │          │
│ │ [Ver →] │ │ [Actuar]│ │ [Revisar]│ │ [Filtrar]│          │
│ └──────────┘ └──────────┘ └──────────┘ └──────────┘          │
│                                                                │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│ FILTROS:  [⚠️ Con Problemas ▼] [📁 Por Ubicación ▼]           │
│           [🎬 Por Tipo ▼] [📅 Por Antigüedad ▼]               │
│                                                                │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│ CATEGORÍAS:                                                    │
│                                                                │
│ [🟢 Todo OK (123)] [🟡 Necesitan Atención (123)]              │
│ [🔴 Problemas Críticos (12)] [📦 Sin Clasificar (34)]        │
│                                                                │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│ 🟡 NECESITAN ATENCIÓN (123)              [☑ Seleccionar todos]│
│                                                                │
│ ╔═══════════════════════════════════════════════════════════╗ │
│ ║ ⚠️ HUÉRFANO EN DESCARGAS                                  ║ │
│ ║                                                           ║ │
│ ║ 🎬 Inception (2010)                                       ║ │
│ ║                                                           ║ │
│ ║ 📍 Ubicación:                                             ║ │
│ ║    📥 /downloads/completed/Inception.2010.mkv             ║ │
│ ║    💾 15.2 GB                                             ║ │
│ ║                                                           ║ │
│ ║ 🔌 Estado de Servicios:                                   ║ │
│ ║    ✅ qBittorrent: Seeding (ratio 2.5, 12 seeds)         ║ │
│ ║    ❌ Radarr: No gestionado                               ║ │
│ ║    ❌ Jellyfin: No en biblioteca                          ║ │
│ ║                                                           ║ │
│ ║ 💡 Sugerencia:                                            ║ │
│ ║    Este archivo está siendo compartido activamente pero   ║ │
│ ║    no está en tu biblioteca. Puedes importarlo a Radarr   ║ │
│ ║    para agregarlo a Jellyfin automáticamente.             ║ │
│ ║                                                           ║ │
│ ║ [✨ Importar a Radarr] [🗑️ Eliminar] [👁️ Ignorar]        ║ │
│ ╚═══════════════════════════════════════════════════════════╝ │
│                                                                │
│ ╔═══════════════════════════════════════════════════════════╗ │
│ ║ 🔗 SOLO HARDLINK (Puede Liberar Espacio)                  ║ │
│ ║                                                           ║ │
│ ║ 📺 Breaking Bad - S01E01                                  ║ │
│ ║                                                           ║ │
│ ║ 📍 Hardlinks detectados:                                  ║ │
│ ║    📚 /jellyfin/series/Breaking.Bad/S01/E01.mkv          ║ │
│ ║    📥 /downloads/completed/Breaking.Bad.S01E01.mkv       ║ │
│ ║    💾 2.1 GB (ocupando solo una vez)                      ║ │
│ ║                                                           ║ │
│ ║ 🔌 Estado:                                                ║ │
│ ║    ✅ Jellyfin: En biblioteca                             ║ │
│ ║    ❌ qBittorrent: Torrent eliminado                      ║ │
│ ║                                                           ║ │
│ ║ 💡 Sugerencia:                                            ║ │
│ ║    El torrent original fue eliminado. Puedes eliminar el  ║ │
│ ║    hardlink de descargas sin perder el archivo.           ║ │
│ ║                                                           ║ │
│ ║ [🧹 Limpiar Hardlink de Descargas] [👁️ Ignorar]          ║ │
│ ╚═══════════════════════════════════════════════════════════╝ │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

---

## 🎯 Plan de Implementación

### Fase 1: Backend - Lógica de Categorización (1-2 días)

1. **Crear servicio de análisis de salud**
   ```go
   // internal/service/health_analyzer.go
   type HealthAnalyzer struct {
       // ...
   }
   
   type FileHealthStatus string
   const (
       HealthStatusOK               FileHealthStatus = "ok"
       HealthStatusOrphanDownload   FileHealthStatus = "orphan_download"
       HealthStatusOnlyHardlink     FileHealthStatus = "only_hardlink"
       HealthStatusDeadTorrent      FileHealthStatus = "dead_torrent"
       HealthStatusNeverWatched     FileHealthStatus = "never_watched"
       HealthStatusUnclassified     FileHealthStatus = "unclassified"
   )
   
   type FileHealthReport struct {
       File          *MediaFileInfo
       Status        FileHealthStatus
       Severity      string // "ok", "warning", "critical"
       Issues        []string
       Suggestions   []string
       Actions       []string // Available actions
   }
   
   func (ha *HealthAnalyzer) AnalyzeFile(file *MediaFileInfo) *FileHealthReport {
       // Lógica de análisis
   }
   ```

2. **Implementar reglas de negocio**
   - Detectar huérfanos: `InQBittorrent && !InJellyfin`
   - Detectar solo hardlinks: `IsHardlink && len(HardlinkPaths) > 1 && !InQBittorrent`
   - Detectar torrents muertos: `InQBittorrent && (TorrentState == "error" || !IsSeeding)`
   - Detectar sin reproducir: `InJellyfin && !HasBeenWatched && AgeDays > 180`

3. **Endpoint API mejorado**
   ```go
   GET /api/files/health
   
   Response:
   {
     "summary": {
       "healthy": 123,
       "orphan_downloads": 45,
       "only_hardlinks": 78,
       "dead_torrents": 12,
       "never_watched": 234,
       "unclassified": 34
     },
     "files": [
       {
         "file": {...},
         "health": {
           "status": "orphan_download",
           "severity": "warning",
           "issues": ["No está en Jellyfin", "No gestionado por Radarr"],
           "suggestions": ["Importar a Radarr para agregarlo automáticamente a Jellyfin"],
           "actions": ["import_radarr", "delete", "ignore"]
         }
       }
     ]
   }
   ```

### Fase 2: Frontend - UI Mejorada (2-3 días)

1. **Componentes Alpine.js**
   - `healthCard()` - Tarjetas de estado
   - `fileHealthCard()` - Card de archivo con sugerencias
   - `healthFilters()` - Filtros inteligentes
   - `bulkActions()` - Acciones masivas

2. **Actualizar template HTML**
   - Reemplazar estadísticas simples por cards accionables
   - Implementar categorías por severidad
   - Agregar sugerencias y botones de acción
   - Mejorar visual de rutas (solo mostrar nombre + tooltip con ruta completa)

3. **Estilos y UX**
   - Colores semánticos: Verde (OK), Amarillo (Warning), Rojo (Critical)
   - Iconos claros y consistentes
   - Tooltips informativos
   - Animaciones suaves para transiciones

### Fase 3: Acciones (3-4 días)

1. **Implementar handlers de acciones**
   ```go
   POST /api/files/{id}/import-to-radarr
   POST /api/files/{id}/delete
   POST /api/files/{id}/ignore
   POST /api/files/{id}/cleanup-hardlink
   POST /api/files/bulk-action
   ```

2. **Integración con servicios**
   - Importar a Radarr/Sonarr vía API
   - Eliminar torrents de qBittorrent
   - Eliminar archivos del sistema (con confirmación)
   - Marcar como ignorados en base de datos

3. **Confirmaciones y validaciones**
   - Modales de confirmación
   - Dry-run mode
   - Logs de acciones
   - Rollback si es posible

### Fase 4: Testing y Refinamiento (1-2 días)

1. Tests de lógica de categorización
2. Tests de integración con servicios
3. Pruebas de UX con usuarios
4. Ajustes según feedback

---

## 🚀 Beneficios Esperados

### Para el Usuario:

1. **Claridad**: Sabe exactamente qué significa cada estado
2. **Acción Guiada**: Ve sugerencias concretas de qué hacer
3. **Confianza**: Confirmaciones claras antes de eliminar
4. **Eficiencia**: Acciones masivas para tareas repetitivas
5. **Aprendizaje**: Tooltips explican conceptos técnicos

### Para el Objetivo de la App (Limpieza de Discos):

1. **Identificación Rápida** de archivos problemáticos
2. **Acciones Seguras**: Solo permite eliminar cuando es seguro
3. **Optimización de Espacio**: Detecta hardlinks innecesarios
4. **Mantenimiento Proactivo**: Detecta torrents muertos
5. **Gestión de Biblioteca**: Ayuda a mantener Jellyfin organizado

---

## 📊 Comparación: Antes vs. Después

### ANTES (Situación Actual):

```
❓ Usuario piensa:
- "¿Qué es todo esto?"
- "¿Qué significa 'hardlink'?"
- "¿Debería eliminar algo?"
- "¿Qué diferencia hay con la pestaña Media?"
```

### DESPUÉS (Con Mejoras):

```
✅ Usuario piensa:
- "¡Tengo 45 archivos huérfanos que puedo importar!"
- "Ah, estos hardlinks puedo limpiarlos sin perder archivos"
- "Estos 12 torrents están muertos, puedo eliminarlos de forma segura"
- "Esta vista me ayuda a mantener mi sistema limpio y organizado"
```

---

## 🎯 Decisión Recomendada

**Implementar Propuesta 1 (Salud del Almacenamiento)** como prioridad principal, con elementos de la Propuesta 3 (Asistente) como feature opcional para usuarios novatos.

### Justificación:

1. ✅ **Alineado con el objetivo**: Limpieza de discos duros
2. ✅ **Accionable**: Cada estado tiene acciones claras
3. ✅ **Educativo**: Explica conceptos técnicos
4. ✅ **Seguro**: No permite acciones destructivas sin confirmación
5. ✅ **Escalable**: Fácil agregar nuevas categorías de salud

---

## 📝 Notas Adicionales

### Diferenciación con Pestaña "Media":

- **Media**: Catálogo completo de contenido (qué tienes)
- **Files**: Salud del almacenamiento (cómo está organizado)

### Integración con Features Existentes:

- Puede enlazar a "Leaving Soon" para archivos antiguos sin ver
- Puede usar reglas de limpieza para sugerir eliminaciones
- Puede aprovechar dry-run mode para simular acciones

### Consideraciones Técnicas:

- Caché de análisis de salud (no recalcular en cada carga)
- Background jobs para análisis pesados
- Rate limiting en APIs de servicios externos
- Logs detallados de todas las acciones

---

**¿Aprobamos esta propuesta para implementación?** 🚀
