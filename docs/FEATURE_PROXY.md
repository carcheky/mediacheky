# Proxy Automático para MediaCheky

## Objetivo
Implementar un sistema de proxy/reverse-proxy automático para los servicios gestionados por MediaCheky, permitiendo la asignación de dominios y subdominios por servicio.

---

## Estado de Implementación

### ✅ Completado

#### 1. Selección de tecnología de proxy
- ✅ **Tecnología elegida: Traefik v2.x**
  - **Razones:**
    - Integración nativa con Docker (descubrimiento automático de servicios)
    - Configuración dinámica mediante labels en contenedores
    - Soporte para múltiples dominios y subdominios
    - Let's Encrypt integrado para SSL/TLS automático
    - Liviano y eficiente (~50MB imagen)
  - **Alternativas consideradas:**
    - Nginx Proxy Manager: Requiere configuración manual via UI
    - Caddy: Buena opción, pero menos adopción en ecosistema Docker

#### 2. Configuración global de dominios ✅
- ✅ Interfaz en `/global` para gestión de dominios
- ✅ Base de datos con tabla `domains` para almacenamiento persistente
- ✅ Validación de formato de dominio (regex)
- ✅ Soporte para múltiples dominios
- ✅ Designación de dominio primario
- ✅ API REST completa para CRUD de dominios

#### 3. Configuración de subdominios por servicio ✅
- ✅ Campo `subdomain` en modelo `Service`
- ✅ Auto-generación: si vacío, usa nombre del servicio
- ✅ Override manual permitido
- ✅ Validación de formato y conflictos
- ✅ UI en página de configuración de cada servicio
- ✅ Visualización del endpoint completo

#### 4. Backend y API ✅
- ✅ Modelos de datos:
  - `ProxyConfig`: Configuración global del proxy
  - `Domain`: Dominios configurados
  - Campos `subdomain` y `domain` en `Service`
- ✅ Repositorio para operaciones DB
- ✅ Servicio de negocio con validaciones
- ✅ Endpoints API REST:
  - `GET /api/proxy/config` - Obtener configuración proxy
  - `PUT /api/proxy/config` - Actualizar configuración proxy
  - `GET /api/proxy/domains` - Listar dominios
  - `POST /api/proxy/domains` - Añadir dominio
  - `DELETE /api/proxy/domains/:name` - Eliminar dominio
  - `PUT /api/services/:name/subdomain` - Configurar subdomain servicio
  - `GET /api/services/:name/endpoint` - Obtener endpoint completo

#### 5. Frontend UI/UX ✅
- ✅ Sección de proxy en `/global`:
  - Toggle enable/disable
  - Lista de dominios configurados
  - Formulario para añadir dominios
  - Indicador de dominio primario
- ✅ Sección en configuración de servicio:
  - Campo de subdomain (con auto-generación)
  - Campo de dominio (opcional, usa primario por defecto)
  - Display del endpoint resultante
  - Botón para copiar URL al portapapeles
- ✅ Toast notifications para feedback
- ✅ Validación en cliente y servidor

### 🚧 Pendiente

#### 6. Generación de reglas de proxy
- [ ] Implementar generación de configuración Traefik
- [ ] Crear archivo de configuración dinámica
- [ ] Generar labels de Docker para servicios
- [ ] Actualizar templates de servicios con labels Traefik

#### 7. Integración con Docker Compose
- [ ] Añadir servicio Traefik a docker compose.yml
- [ ] Configurar volúmenes para certificados SSL
- [ ] Configurar red compartida
- [ ] Implementar lógica de reinicio selectivo

#### 8. Testing
- [ ] Unit tests para ProxyService
- [ ] Integration tests para API endpoints
- [ ] Tests de validación de dominios/subdominios
- [ ] Tests de conflicto de subdominios

#### 9. Documentación
- [ ] Guía de usuario
- [ ] Ejemplos de configuración
- [ ] Troubleshooting
- [ ] Guía de seguridad

---

## Arquitectura

### Flujo de Configuración

```
Usuario → UI (Alpine.js) → API (Go/Fiber) → Service Layer → Repository → Database (SQLite)
                                    ↓
                            Docker Compose Templates
                                    ↓
                            Traefik Configuration
```

### Estructura de Base de Datos

```sql
-- Configuración global del proxy
CREATE TABLE proxy_config (
    id INTEGER PRIMARY KEY,
    proxy_type VARCHAR(50) DEFAULT 'traefik',
    enabled BOOLEAN DEFAULT FALSE,
    domains TEXT,  -- Comma-separated legacy field
    ssl_enabled BOOLEAN DEFAULT FALSE,
    ssl_email VARCHAR(255),
    container_id VARCHAR(255),
    status VARCHAR(50) DEFAULT 'stopped',
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

-- Dominios configurados
CREATE TABLE domains (
    id INTEGER PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    is_primary BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

-- Servicios (campos relacionados con proxy)
ALTER TABLE services ADD COLUMN subdomain VARCHAR(255);
ALTER TABLE services ADD COLUMN domain VARCHAR(255);
```

### Modelos Go

```go
type ProxyConfig struct {
    ProxyType  string // "traefik", "nginx", "caddy"
    Enabled    bool
    SSLEnabled bool
    SSLEmail   string
    // ...
}

type Domain struct {
    Name      string
    IsPrimary bool
    // ...
}

type Service struct {
    // ... existing fields
    Subdomain string // e.g., "radarr"
    Domain    string // e.g., "media.local"
}
```

---

## Uso

### Configurar Dominios

1. Ir a **Global Variables** (`/global`)
2. Scroll a la sección **Proxy & Domains**
3. Habilitar el proxy con el toggle
4. Añadir dominio:
   - Ejemplo: `media.local`, `mediacheky.home`, `example.com`
   - Marcar como primario si es el dominio principal
5. Guardar

### Configurar Subdomain de Servicio

1. Ir a **Settings** y seleccionar un servicio
2. En la sección **Proxy & Domain Access**:
   - **Subdomain**: Dejar vacío para auto-generar, o especificar custom
   - **Domain**: Dejar vacío para usar dominio primario
3. Click **Update Proxy**
4. El endpoint resultante se muestra automáticamente

### Ejemplo de Configuración

**Global:**
- Dominio primario: `media.local`

**Servicios:**
- Radarr: subdomain vacío → `radarr.media.local`
- Sonarr: subdomain `tv` → `tv.media.local`
- Jellyfin: subdomain vacío, domain `streaming.local` → `jellyfin.streaming.local`

---

## Validaciones y Seguridad

### Validación de Dominios
- Formato: `^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)*[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`
- Permitidos: `example.com`, `media.local`, `localhost`
- No permitidos: `-example.com`, `example..com`, `exam ple.com`

### Validación de Subdominios
- Formato: `^[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`
- Permitidos: `radarr`, `tv-shows`, `media123`
- No permitidos: `-radarr`, `radarr-`, `rad arr`

### Prevención de Conflictos
- El sistema verifica antes de guardar si otro servicio ya usa el mismo subdomain+domain
- Si hay conflicto, muestra error indicando qué servicio lo está usando

### Sanitización
- Todas las entradas se limpian de espacios extras
- Se convierten a minúsculas para consistencia
- Se validan contra regex antes de almacenar

---

## Próximos Pasos

### Sprint 1: Integración Traefik
1. Crear template de servicio Traefik
2. Generar configuración dinámica
3. Implementar labels de Docker
4. Testing básico

### Sprint 2: SSL/TLS
1. Configurar Let's Encrypt
2. Soporte para certificados custom
3. Renovación automática
4. Testing SSL

### Sprint 3: Documentación
1. Guía de usuario completa
2. Video tutorial
3. Ejemplos avanzados
4. FAQ y troubleshooting

---

## Notas Técnicas

### Por qué Traefik sobre Nginx
- **Pro Traefik:**
  - Auto-discovery de contenedores Docker
  - Configuración mediante labels (no requiere archivos)
  - SSL automático con Let's Encrypt
  - Reload sin downtime
  - Dashboard integrado
  
- **Contra Nginx:**
  - Requiere regenerar archivos de configuración
  - Requiere reload manual
  - Let's Encrypt via Certbot externo
  - Más complejo de automatizar

### Seguridad
⚠️ **Importante:**
- El proxy NO añade autenticación por sí mismo
- Los servicios siguen siendo accesibles directamente por puerto si están expuestos
- Recomendamos:
  - Usar firewall para bloquear puertos directos
  - Configurar autenticación en cada servicio
  - Usar VPN o autenticación en Traefik
  - No exponer MediaCheky a internet sin protección

---

**Última actualización:** 2025-11-11 19:30 UTC  
**Estado:** Fase de backend y frontend completada ✅  
**Siguiente:** Integración con Traefik 🚧
