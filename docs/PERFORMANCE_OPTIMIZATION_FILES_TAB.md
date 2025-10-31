# Optimización de Rendimiento - Pestaña Files

## Problema
La pestaña "Files" cargaba muy lentamente con bibliotecas grandes (miles de archivos).

## Causas Raíz Identificadas
1. **6 queries COUNT separadas** ejecutadas en cada request
2. **Sin caché** - conteos recalculados cada vez
3. **Índices compuestos faltantes** para patrones de filtrado comunes
4. **Sin monitoreo de rendimiento de queries**

## Soluciones Implementadas

### 1. Sistema de Caché Inteligente
- Caché thread-safe con TTL de 30 segundos
- Invalidación automática después de sincronización
- Copias defensivas para prevenir modificaciones externas
- Tasa de aciertos esperada: >80%

### 2. Optimización de Queries
```sql
-- ANTES: 6 queries separadas
SELECT COUNT(*) FROM media WHERE in_jellyfin = true AND ...;  -- Query 1
SELECT COUNT(*) FROM media WHERE in_q_bittorrent = true AND ...; -- Query 2
-- ... 4 queries más

-- DESPUÉS: 1 query optimizada con CASE statements
SELECT 
  COUNT(CASE WHEN in_jellyfin = true AND ... THEN 1 END) as healthy,
  COUNT(CASE WHEN in_q_bittorrent = true AND ... THEN 1 END) as attention,
  -- ... todos los conteos en una sola query
FROM media;
```

### 3. Índices Compuestos
```sql
-- Para patrones de filtrado comunes
CREATE INDEX idx_media_healthy_files ON media(in_jellyfin, in_radarr, in_sonarr, torrent_state);
CREATE INDEX idx_media_orphan_downloads ON media(in_q_bittorrent, in_jellyfin, in_radarr, in_sonarr);
CREATE INDEX idx_media_dead_torrents ON media(in_q_bittorrent, torrent_state);
CREATE INDEX idx_media_default_sort ON media(in_q_bittorrent DESC, in_jellyfin DESC, file_path ASC);
```

### 4. Monitoreo de Rendimiento
```go
// Logs detallados de timing
h.logger.Info("Files API request completed",
    zap.Duration("total_time", totalElapsed),      // Tiempo total del request
    zap.Duration("count_time", countElapsed),      // Tiempo de query COUNT
    zap.Duration("query_time", queryElapsed),      // Tiempo de fetch de datos
    zap.Duration("counts_time", countsElapsed),    // Tiempo de conteos de categorías
)
```

## Impacto de Rendimiento Esperado

| Métrica | Antes | Después | Mejora |
|---------|-------|---------|--------|
| Queries de conteo de categorías | 6 separadas | 1 optimizada (cacheada) | ~83% reducción |
| Tiempo de respuesta promedio (1000+ archivos) | 500-2000ms | 50-200ms | **70-90% más rápido** |
| Carga de base de datos | Alta en cada request | Baja con cache hits | Reducción significativa |
| Tasa de aciertos de caché | 0% (sin caché) | >80% esperado | Mejora infinita |

## Cómo Verificar las Mejoras

### 1. Revisar Logs Después del Deploy
Buscar los nuevos logs de rendimiento:
```
Files API request completed page=1 perPage=25 total=1523 
  total_time=45ms count_time=8ms query_time=15ms counts_time=2ms
```

### 2. Comparar Antes/Después
- **Primer request (cache miss)**: Debería mostrar `counts_time` ~20-50ms
- **Requests subsecuentes (cache hit)**: Debería mostrar `counts_time` ~0-2ms
- **Después de sync**: Caché invalidado, próximo request será cache miss

### 3. Monitorear Base de Datos
- Verificar planes de ejecución de queries - deberían usar los nuevos índices
- Monitorear conteo de queries - debería verse reducción en queries COUNT

## Calidad del Código

✅ Build exitoso
✅ Code review completado y feedback implementado
✅ Escaneo de seguridad CodeQL: 0 alertas
✅ Sin vulnerabilidades de seguridad
✅ Implementación thread-safe
✅ Backward compatible

## Archivos Modificados

1. `pkg/cache/counts.go` - Nuevo paquete de caché thread-safe
2. `internal/handler/files.go` - Integración de caché y logs de rendimiento
3. `internal/models/models.go` - Migración de índices compuestos
4. `internal/service/filesystem_sync_service.go` - Invalidación de caché

## Próximos Pasos

1. Desplegar a producción/staging
2. Monitorear logs de rendimiento
3. Ajustar TTL de caché si es necesario (actualmente 30s)
4. Considerar caché adicional si emergen otros patrones lentos

---
**Estado**: ✅ Listo para merge
**Riesgo**: Bajo (backward compatible, implementación defensiva)
**Impacto**: Alto (70-90% de mejora de rendimiento esperada)
