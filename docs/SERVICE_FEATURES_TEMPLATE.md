# INTEGRACIÓN [SERVICE_NAME]

## Página de configuración y funcionalidad

La página de configuración de [service_name] tendrá varias secciones de configuración

- [ ] Botones:
  - [ ] Toggle para activar/desactivar [service_name]: desactivado desactiva todos los campos y elimina [service_name], si se activa levanta [service_name]
  - [ ] Botón Aplicar (Apply): aplica los cambios y reconstruye el contenedor si corresponde
  - [ ] Botón Start/Stop: aparece uno de los dos dinámicamente según el estado del contenedor
  - [ ] Botón Reset: borra la configuración (volumen config) y recrea el contenedor con los valores por defecto

- [ ] Configuración de Docker:
  - Cualquiera de estas opciones eliminará el contenedor, guardará la configuración y levantará de nuevo [service_name] con la nueva configuración:
    - [ ] Tag de la imagen Docker: desplegable con los últimos tags disponibles para seleccionarlos, si está vacío usará `latest`
    - [ ] Puerto expuesto: si está vacío no se expone, si está relleno se expone el puerto. El puerto interno será `[INTERNAL_PORT]`
    - [ ] Restart policy: desplegable con las opciones estándar de Docker
    - [ ] Config path: ruta a la configuración en el host, por defecto `./volumes/[service_name]/config`, esta carpeta debe estar en la raíz de este repositorio

- [ ] Configuración específica de [SERVICE_NAME]:
  - [ ] !TODO: Añadir configuraciones específicas del servicio
  - [ ] !TODO: Estas opciones configurarán directamente [service_name]

- [ ] Proxy/Acceso externo:
  - Configuración de dominio:
    - [ ] Subdominio: si está vacío usará `[service_name]`, por defecto vacío
    - [ ] Dominio: si está vacío usará la configuración global, por defecto vacío

---

## Notas de implementación

### Valores por defecto
- Puerto expuesto: vacío (puerto sin exponer)
- Host port: debe poder quedarse vacío, no a 0, **vacío**
- Tag de imagen: `latest`
- Config path: `./volumes/[service_name]/config`
- Subdominio: vacío (usará `[service_name]`)
- Dominio: vacío (usará configuración global)

### Comportamiento de botones
- **Start/Stop**: se intercambian según si el contenedor está en ejecución o no
- **Restart**: se lanzará con `--force-recreate`, haciendo un `kill` y un `down` antes
- **Reset**: borra el volumen de configuración y recrea el contenedor con valores por defecto

### Gestión de imagen Docker
- La imagen base no se podrá cambiar (ej: `linuxserver/[service_name]`)
- El tag sí tendrá un desplegable con las versiones disponibles
- Opciones del desplegable: `latest` (primera opción y por defecto) + versiones disponibles

### Lógica de recreación
Si cambia cualquier configuración de Docker al guardar:
- **Se debe recrear el contenedor**
- **Con pull**: si cambia el tag de la imagen
- **Sin pull**: si NO cambia el tag

---

## Variables a reemplazar al usar esta plantilla

| Variable | Descripción | Ejemplo |
|----------|-------------|---------|
| `[SERVICE_NAME]` | Nombre del servicio en mayúsculas | RADARR, SONARR, JELLYFIN |
| `[service_name]` | Nombre del servicio en minúsculas | radarr, sonarr, jellyfin |
| `[INTERNAL_PORT]` | Puerto interno del contenedor | 7878, 8989, 8096 |

---

## Servicios pendientes de documentar

- [ ] Sonarr (puerto: 8989)
- [ ] Jellyfin (puerto: 8096)
- [ ] Prowlarr (puerto: 9696)
- [ ] qBittorrent (puerto: 8080)
- [ ] Jellyseerr (puerto: 5055)
- [ ] Bazarr (puerto: 6767)
- [ ] Jellystat (puerto: 3000)

<!-- Variables de plantilla (no son links de Markdown) -->
<!-- [SERVICE_NAME] = Nombre en mayúsculas -->
<!-- [service_name] = Nombre en minúsculas -->
<!-- [INTERNAL_PORT] = Puerto interno del contenedor -->

