# Work Queue System - Sequential Job Processing

## Overview

Sistema de **cola de trabajo secuencial** que procesa acciones en orden (FIFO - First In, First Out) y muestra progreso visual persistente aunque el usuario navegue entre páginas.

### ¿Qué problema resuelve?

Si el usuario activa/desactiva/prune múltiples servicios rápidamente, las acciones se ejecutarán **en orden**, evitando:

- Conflictos de recursos (Docker socket)
- Errores por operaciones concurrentes
- Sobrecarga del backend

### Funcionalidades Principales

- ✅ **Cola FIFO**: Las acciones se procesan en el orden en que se añaden
- ✅ **Un trabajo a la vez**: Solo se ejecuta una acción simultáneamente
- ✅ **Persistencia**: El estado de la cola se guarda en `localStorage`

## Características

- ✅ **Persistencia**: Las acciones en progreso se guardan en `localStorage`
- ✅ **Multi-pestaña**: Sincronización automática entre pestañas del navegador
- ✅ **Visual consistente**: Mismo tooltip de progreso en todas las páginas
- ✅ **Auto-limpieza**: Elimina progreso obsoleto (>5 minutos)
- ✅ **Integración Alpine.js**: Componente reactivo global

## Arquitectura

```text
┌─────────────────────────────────────────┐
│  localStorage (mediacheky_work_queue)   │
│  - queue: [job1, job2, job3, ...]       │
│  - current: { job executing now }       │
└────────────────┬────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────┐
│  window.WorkQueue (JavaScript API)      │
│  - enqueue(job) → adds to queue         │
│  - processNext() → executes FIFO        │
│  - updateStep(index) → progress update  │
│  - getCurrentProgress() → current job   │
└────────────────┬────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────┐
│  Alpine.js globalProgress() component   │
│  - Renderiza tooltip flotante           │
│  - Auto-actualiza en tiempo real        │
└─────────────────────────────────────────┘
```

## Uso

### 1. Añadir trabajo a la cola

```javascript
window.WorkQueue.enqueue({
    id: 'prune-radarr',              // ID único (opcional)
    title: 'Pruning Radarr',         // Título para mostrar
    steps: [                         // Pasos del progreso
        'Stopping service',
        'Removing volumes',
        'Cleaning configuration',
        'Reloading data'
    ],
    warning: 'Service cleanup in progress...',  // Advertencia
    execute: async (queue) => {      // Función que ejecuta el trabajo
        // Tu código aquí
        await fetch('/api/services/radarr/reset', { method: 'POST' });
        
        // Actualizar paso
        queue.updateStep(3);  // Saltar a paso 4
        
        // Más código...
    }
});
```

### 2. Actualizar paso dentro de execute

```javascript
execute: async (queue) => {
    // Paso 1 (automático al iniciar)
    await doSomething();
    
    // Avanzar al siguiente paso
    queue.updateStep();
    
    // O saltar a un paso específico (0-indexed)
    queue.updateStep(2);  // Saltar a paso 3
}
```

### 3. La cola procesa automáticamente

**No necesitas llamar a `.end()`** - WorkQueue finaliza automáticamente cuando `execute()` termina.

## Implementación

### Archivos

- **`/web/static/js/global-progress.js`**: Sistema de cola y API
- **`/web/templates/layouts/main.html`**: Componente visual global
- **`/web/templates/pages/services.html`**: Ejemplo de uso en listado
- **`/web/templates/pages/service_config.html`**: Ejemplo de uso en configuración

### Componente Visual

El tooltip flotante aparece en **bottom-left** de todas las páginas:

```html
<div x-data="globalProgress()" x-show="showProgress">
    <!-- Warning Banner (amarillo/amber) -->
    <!-- Progress Content (azul) -->
    <!-- Steps con checkmarks verdes -->
</div>
```

### Estados de Pasos

- ✅ **Completado**: Círculo verde con checkmark
- 🔵 **En progreso**: Círculo azul pulsante
- ⚪ **Pendiente**: Círculo vacío (border azul)

## Ejemplo Completo

```javascript
async function pruneService(serviceId) {
    // Añadir trabajo a la cola
    window.WorkQueue.enqueue({
        id: `prune-${serviceId}`,
        title: `Pruning ${serviceId}`,
        steps: [
            'Stopping service',
            'Removing volumes',
            'Cleaning configuration',
            'Reloading data'
        ],
        warning: 'Service cleanup in progress...',
        execute: async (queue) => {
            try {
                // Paso 1: Stopping service (automático)
                const response = await fetch(`/api/services/${serviceId}/reset`, {
                    method: 'POST'
                });

                if (response.ok) {
                    // Saltar a paso 4: Reloading data
                    queue.updateStep(3);
                    
                    await new Promise(resolve => setTimeout(resolve, 2000));
                    await this.reloadData();
                    
                    this.showMessage('Success!', 'success');
                } else {
                    this.showMessage('Failed', 'error');
                }
            } catch (error) {
                this.showMessage('Error: ' + error.message, 'error');
            }
        }
    });
}
```

## Ejemplo: Múltiples Acciones en Secuencia

```javascript
// Usuario pulsa: Enable Radarr, Enable Sonarr, Prune Jellyfin
// Se ejecutan EN ORDEN:

// 1. Primera acción se ejecuta inmediatamente
window.WorkQueue.enqueue({ id: 'enable-radarr', ... });  // ✅ Ejecutando

// 2. Segunda acción espera en la cola
window.WorkQueue.enqueue({ id: 'enable-sonarr', ... });  // ⏳ En cola

// 3. Tercera acción también espera
window.WorkQueue.enqueue({ id: 'prune-jellyfin', ... }); // ⏳ En cola

// Resultado: Se ejecutan una tras otra, sin conflictos
```

## Beneficios

1. **UX Mejorada**: Usuario siempre ve el estado de operaciones largas
2. **Navegación Libre**: Puede cambiar de página sin perder progreso visual
3. **Feedback Claro**: Pasos detallados y advertencias visibles
4. **Prevención de Errores**: Usuario sabe que no debe cerrar la pestaña
5. **Multi-pestaña**: Si abre otra pestaña, verá el mismo progreso

## Casos de Uso

- ✅ Prune de servicios (docker compose down -v)
- ✅ Enable/Disable servicios
- ✅ Aplicar configuración
- ✅ Test de conexión (opcional)
- ✅ Restart de contenedores
- ✅ Cualquier operación larga (>2s)

## Notas

- El progreso se limpia automáticamente después de 5 minutos
- Solo se muestra un progreso a la vez (el más reciente)
- Si hay error, llamar siempre a `end()` en el `catch`
- El componente visual se renderiza en el layout base, visible en todas las páginas

## Related

- [UI Standards](UI_STANDARDS.md)
- [UI Implementation](UI_IMPLEMENTATION.md)
- [Project Plan](PROJECT_PLAN.md)

