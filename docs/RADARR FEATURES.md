# INTEGRACIÓN RADARR

## Página de configuración y funcionalidad

La página de configuración de radar tendrá varias secciones de configuración

- [ ] botones:
  - [ ] Toggle para activar desactivar radarr, desactivado desactiva todos los campos y elimina radarr, si se activa levanta radarr
  - [ ] botón Aplicar (apply): aplica los cambios sin confirmación y reconstruye el contenedor si corresponde
  - [ ] botón start/stop: aparece uno de los dos dinámicamente según el estado del contenedor
  - [ ] botón Reset: borra la configuración (volumen config) y recrea el contenedor con valores por defecto. **Requiere confirmación antes de ejecutar**
- [ ] Configuraciones:
  - [ ] configuración de docker con los campos (cualquiera de estas opciones eliminará el contenedor, guardará la configuración y levantará de nuevo radarr con la nueva configuración) :
    - [ ] tag de la imagen docker (se podrán cargar dinámicamente los últimos tags para seleccionarlos) si está vacío usará latest
    - [ ] puerto expuesto: si está vacío no se expone, si está relleno se expone el puerto expuesto el puerto interno será del de radarr 7878
    - [ ] restart policy: desplegable con las opciones
    - [ ] config path (SOLO LECTURA/INFO): muestra la ruta REAL en el host donde se guardan los datos. Valor hardcodeado: `./volumes/mediacheky-data/services/radarr/config` - NO es configurable por el usuario
  - [ ] Configuración de radarr (estas opciones configurarán directamente radarr [!RADARR API](https://arrapi.kometa.wiki/en/latest/radarr.html))
    - [ ] Revisa la configuración en radarr en el archivo ```` y captura los valores MEDIACHEKY_PLACEHOLDER_*, y crea tantos campos como valores encuentres. el formulario al guardar deberá guardar los valores aquí. esos valores serán los nombres de los campos, y los valores de los campos serán guardados en ese config.yml (haz un backup como referencia antes de que se sobreescriba)
  - [ ] Proxy/Acceso externo: configuración de dominio:
    - [ ] subdominio: si está vacío usará radarr, por defecto vacío
    - [ ] dominio: si está vacío usará la configuración global, por defecto vacío

------

## Notas de implementación

### Comportamiento de Botones

- **Apply**: Aplica cambios **sin confirmación**. Si cambia configuración de Docker, recrea el contenedor. **Solo visible cuando el servicio está habilitado**
- **Reset**: **Requiere confirmación** antes de ejecutar. **Siempre visible**, independientemente del estado del servicio. Al pulsar:
  1. Deshabilita el servicio si está habilitado
  2. Borra el volumen de configuración de Radarr
  3. Actualiza el estado de la UI a "deshabilitado"
- **Test Connection**: **ELIMINADO** - no se muestra en la UI
- **Start/Stop**: Se intercambian dinámicamente según el estado del contenedor. **Solo visibles cuando el servicio está habilitado**
- **Restart**: Se lanza con `--force-recreate`, haciendo kill y down antes. **Solo visible cuando el servicio está habilitado**
- **Enable Toggle**: Al habilitar el servicio, aplica la configuración automáticamente (igual que Apply)

### Configuración de Puerto

- Por defecto el valor del puerto debe estar **vacío** (puerto sin exponer)
- El host port debe poder quedarse **vacío**, no a 0, literalmente vacío
- Si está vacío, no se expone el puerto; si está relleno, se expone

### Configuración de Docker

- Docker image **no se puede cambiar** (siempre linuxserver/radarr)
- El tag **sí es configurable** con desplegable de versiones disponibles
- **latest** es la primera opción y valor por defecto
- Si cambia cualquier configuración de Docker al guardar, se debe recrear el contenedor:
  - Con pull si cambia el tag
  - Sin pull si no cambia el tag

### Configuración de Radarr

- Config path es **SOLO LECTURA**: `./volumes/mediacheky-data/services/radarr/config`
- NO es configurable por el usuario, solo informativo

### Gestión de Permisos

- **PUID/PGID**: MediaCheky aplica automáticamente el UID/GID del usuario del sistema
- **Variables de entorno**: Se configuran en `.env` o se obtienen del usuario actual
- **Permisos de archivos**: Todos los archivos y directorios creados por MediaCheky usan el PUID/PGID configurado
- **Servicios gestionados**: Los servicios (Radarr, Sonarr, etc.) reciben PUID/PGID en sus variables de entorno
- **Configuración por defecto**: 1000:1000 si no se especifica
- **Obtener valores**: Ejecutar `id -u` (PUID) e `id -g` (PGID) en el host
