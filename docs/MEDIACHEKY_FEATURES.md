# MEDIACHEKY - FUNCIONALIDADES PRINCIPALES

## Panel de Control Principal

MediaCheky es el panel de control centralizado que gestiona todos los servicios multimedia (*arr ecosystem).

---

## 📊 Dashboard

- [ ] **Vista general de servicios**
  - [ ] Tarjetas de estado para cada servicio configurado
  - [ ] Indicadores visuales de estado (Running, Stopped, Error)
  - [ ] Contador de servicios por estado (✅ Running, ⚠️ Issues, 📴 Off)
  - [ ] Quick actions por servicio (Start, Stop, Restart, Configure)
  - [ ] Links directos a la UI de cada servicio

- [ ] **Estadísticas del sistema**
  - [ ] Total de servicios configurados
  - [ ] Información del host Docker (OS, arquitectura)
  - [ ] Uptime de MediaCheky
  - [ ] Métricas opcionales (CPU, memoria)

- [ ] **Actualización en tiempo real**
  - [ ] Polling automático cada 30s
  - [ ] Indicador de última actualización
  - [ ] Botón de refresh manual

---

## ⚙️ Settings - Configuración Global

### Pestaña: Services (Servicios)

- [ ] **Gestión de servicios**
  - [ ] Lista de todos los servicios disponibles
  - [ ] Toggle para habilitar/deshabilitar cada servicio
  - [ ] Botones de control (Start, Stop, Restart)
  - [ ] Botón Configure (redirige a `/services/:name`)
  - [ ] Badge de estado en tiempo real

- [ ] **Acciones globales**
  - [ ] Botón "Start All" (iniciar todos los habilitados)
  - [ ] Botón "Stop All" (detener todos los servicios)
  - [ ] Botón "Restart All" (reiniciar todos los habilitados)
  - [ ] Confirmación para acciones masivas

### Pestaña: Global Variables (Variables Globales)

- [ ] **User & Group**
  - [ ] PUID: User ID (por defecto: 1000)
  - [ ] PGID: Group ID (por defecto: 1000)
  - [ ] Validación de IDs numéricos

- [ ] **Sistema**
  - [ ] Timezone: desplegable con zonas horarias (por defecto: Europe/Madrid)
  - [ ] Language: desplegable de idiomas (por defecto: es-ES)

- [ ] **Rutas base (Base Paths)**
  - [ ] Media: ruta a archivos multimedia (por defecto: /media)
  - [ ] Downloads: ruta a descargas (por defecto: /downloads)
  - [ ] Config: ruta a configuraciones (por defecto: /config)
  - [ ] Validación de rutas absolutas

- [ ] **Red (Network)**
  - [ ] Network name: nombre de la red Docker (por defecto: mediacheky-net)
  - [ ] Subnet: subred opcional (por defecto: 172.20.0.0/16)
  - [ ] Gateway: gateway opcional

- [ ] **Proxy/Reverse Proxy**
  - [ ] Dominio base: dominio principal (ej: example.com)
  - [ ] Habilitar SSL/TLS automático
  - [ ] Email para certificados Let's Encrypt
  - [ ] Puerto HTTP (por defecto: 80)
  - [ ] Puerto HTTPS (por defecto: 443)

- [ ] **Acciones**
  - [ ] Botón "Save Global Settings"
  - [ ] Indicador de servicios afectados
  - [ ] Confirmación antes de aplicar cambios que afecten servicios en ejecución

---

## 🔧 Service Configuration (Configuración Individual)

Página `/services/:name` para configuración específica de cada servicio.

- [ ] **Header de servicio**
  - [ ] Nombre y logo del servicio
  - [ ] Estado actual (Running/Stopped/Error)
  - [ ] Breadcrumb de navegación (Settings > [Service])

- [ ] **Secciones comunes a todos los servicios**
  - [ ] Configuración de Docker
  - [ ] Configuración específica del servicio
  - [ ] Proxy/Acceso externo
  - [ ] Variables de entorno

- [ ] **Botones de control**
  - [ ] Toggle Enable/Disable (en header)
  - [ ] Apply (guardar y aplicar cambios)
  - [ ] Start/Stop (dinámico según estado)
  - [ ] Restart (con --force-recreate)
  - [ ] Reset (borrar config y recrear)
  - [ ] Test Connection (verificar conectividad)

---

## 📝 Logs - Visualización de Logs

- [ ] **Filtros**
  - [ ] Selector de servicio (All, MediaCheky, Radarr, Sonarr, etc.)
  - [ ] Nivel de log (All, Info, Warning, Error)
  - [ ] Rango de fechas
  - [ ] Búsqueda por texto

- [ ] **Visualización**
  - [ ] Vista en tiempo real (auto-scroll)
  - [ ] Colorizado por nivel de log
  - [ ] Timestamps
  - [ ] Paginación para logs históricos

- [ ] **Acciones**
  - [ ] Botón "Clear Logs" (con confirmación)
  - [ ] Exportar logs (JSON, TXT)
  - [ ] Botón "Pause/Resume" auto-refresh

---

## 🐳 Docker Integration

### Gestión de contenedores

- [ ] **Docker Socket API**
  - [ ] Conectar a /var/run/docker.sock
  - [ ] Listar contenedores gestionados
  - [ ] Obtener estado de contenedores
  - [ ] Iniciar/detener/reiniciar contenedores
  - [ ] Eliminar contenedores
  - [ ] Leer logs de contenedores

- [ ] **Docker Compose**
  - [ ] Generar archivos docker-compose.yml desde templates
  - [ ] Guardar en ./volumes/services/[service_name]/docker-compose.yml
  - [ ] Ejecutar docker compose up/down/restart
  - [ ] Pull de imágenes

- [ ] **Monitoreo**
  - [ ] Health checks automáticos
  - [ ] Detección de contenedores huérfanos
  - [ ] Alertas de servicios caídos

### Templates de servicios

- [ ] **Sistema de plantillas**
  - [ ] Templates en formato YAML con variables Go template
  - [ ] Schema JSON para validación
  - [ ] Variables globales inyectables
  - [ ] Variables específicas por servicio

- [ ] **Generación de configuración**
  - [ ] Merge de variables globales + específicas
  - [ ] Validación contra JSON Schema
  - [ ] Generación de docker-compose.yml
  - [ ] Backup de configuración anterior

---

## 🌐 Proxy/Reverse Proxy System

- [ ] **Gestión automática de proxy**
  - [ ] Detectar servicios con proxy habilitado
  - [ ] Generar configuración de Traefik/Nginx
  - [ ] Crear reglas de routing
  - [ ] Configurar subdominios automáticamente

- [ ] **SSL/TLS**
  - [ ] Generación automática de certificados Let's Encrypt
  - [ ] Renovación automática de certificados
  - [ ] Redirección HTTP → HTTPS
  - [ ] HSTS headers

- [ ] **Configuración por servicio**
  - [ ] Habilitar/deshabilitar proxy
  - [ ] Subdominio personalizado
  - [ ] Dominio específico (override del global)
  - [ ] Rutas personalizadas (path-based routing)

---

## 🔐 Seguridad

- [ ] **Validación de inputs**
  - [ ] Sanitizar todos los inputs de usuario
  - [ ] Validar rutas (prevenir path traversal)
  - [ ] Whitelist de imágenes Docker permitidas
  - [ ] Validación de puertos (1024-65535)

- [ ] **Autenticación** (Fase 3)
  - [ ] Sistema de login básico
  - [ ] Gestión de usuarios
  - [ ] Roles (admin, viewer)
  - [ ] Tokens de sesión
  - [ ] Logout

- [ ] **Auditoría**
  - [ ] Log de todas las acciones de usuario
  - [ ] Registro de cambios de configuración
  - [ ] Timestamp y usuario en cada acción
  - [ ] Exportar logs de auditoría

- [ ] **Seguridad Docker**
  - [ ] Validar acceso a Docker socket
  - [ ] Limitar operaciones permitidas
  - [ ] Documentar riesgos de seguridad

---

## 💾 Base de Datos

- [ ] **Modelos**
  - [ ] Service: configuración de servicios
  - [ ] Template: plantillas de docker-compose
  - [ ] GlobalConfig: variables globales
  - [ ] ServiceLog: logs de operaciones
  - [ ] User: usuarios (fase 3)

- [ ] **Operaciones**
  - [ ] CRUD completo para todos los modelos
  - [ ] Migraciones automáticas
  - [ ] Backup automático de DB
  - [ ] Importar/exportar configuración

---

## 🎨 UI/UX

- [ ] **Tema oscuro**
  - [ ] Color scheme consistente
  - [ ] Componentes reutilizables
  - [ ] Alpine.js para interactividad
  - [ ] Tailwind CSS para estilos

- [ ] **Notificaciones**
  - [ ] Toast notifications (floating, top-right)
  - [ ] Auto-dismiss (5s success, 8s error)
  - [ ] Close button manual
  - [ ] Tipos: success, error, warning, info

- [ ] **Responsive design**
  - [ ] Mobile-first approach
  - [ ] Navegación colapsable en móvil
  - [ ] Grid layouts adaptativos
  - [ ] Touch-friendly buttons

- [ ] **Accesibilidad**
  - [ ] HTML semántico
  - [ ] ARIA labels
  - [ ] Navegación por teclado
  - [ ] Contraste de color suficiente

---

## 🔄 Sistema de Actualización

- [ ] **Auto-update**
  - [ ] Verificar versiones disponibles
  - [ ] Notificar al usuario de nuevas versiones
  - [ ] Changelog integrado
  - [ ] Botón "Update MediaCheky"

- [ ] **Actualizaciones de servicios**
  - [ ] Verificar nuevas versiones de imágenes Docker
  - [ ] Notificar actualizaciones disponibles
  - [ ] Botón "Update All Services"
  - [ ] Update individual por servicio

---

## 📦 Exportar/Importar

- [ ] **Exportar configuración**
  - [ ] Exportar todo a archivo JSON/YAML
  - [ ] Exportar servicio específico
  - [ ] Exportar solo variables globales
  - [ ] Incluir/excluir credenciales

- [ ] **Importar configuración**
  - [ ] Importar desde archivo
  - [ ] Validar formato antes de importar
  - [ ] Preview de cambios
  - [ ] Confirmación antes de aplicar

- [ ] **Backup automático**
  - [ ] Backup antes de cambios importantes
  - [ ] Programar backups periódicos
  - [ ] Retención configurable
  - [ ] Restaurar desde backup

---

## 🧪 Health Checks & Monitoring

- [ ] **Health checks**
  - [ ] Endpoint `/health` para MediaCheky
  - [ ] Verificar estado de DB
  - [ ] Verificar acceso a Docker socket
  - [ ] Verificar servicios críticos

- [ ] **Monitoreo de servicios**
  - [ ] Ping periódico a servicios habilitados
  - [ ] Detectar servicios caídos
  - [ ] Intentos de auto-restart
  - [ ] Alertas configurables

- [ ] **Métricas**
  - [ ] Uptime de cada servicio
  - [ ] Número de reinicios
  - [ ] Historial de cambios de estado
  - [ ] Estadísticas de uso (opcional)

---

## 🚀 Deployment

- [ ] **Docker image**
  - [ ] Multi-stage build
  - [ ] Imagen optimizada (<25MB)
  - [ ] Tags versionados
  - [ ] Publicar en Docker Hub

- [ ] **Requisitos**
  - [ ] Docker 20.10+
  - [ ] Docker Compose 2.0+
  - [ ] Acceso a Docker socket
  - [ ] Puertos disponibles

- [ ] **Instalación**
  - [ ] docker-compose.yml ejemplo
  - [ ] Script de instalación rápida
  - [ ] Documentación completa
  - [ ] Troubleshooting guide

---

## 📚 Documentación

- [ ] **README**
  - [ ] Descripción del proyecto
  - [ ] Quick start
  - [ ] Screenshots
  - [ ] Links a documentación

- [ ] **Guías**
  - [ ] Installation guide
  - [ ] Configuration guide
  - [ ] Service-specific guides
  - [ ] Troubleshooting

- [ ] **API**
  - [ ] Documentación de endpoints
  - [ ] Ejemplos de uso
  - [ ] Códigos de error
  - [ ] Swagger/OpenAPI (opcional)

---

## 🎯 Prioridades de Desarrollo

### MVP (Fase 1)
- Dashboard básico
- Settings con servicios
- Configuración de 1 servicio (Radarr)
- Variables globales
- Docker integration

### Core Features (Fase 2)
- Los 8 servicios principales
- Logs centralizados
- Health checks
- Templates completos

### Advanced Features (Fase 3)
- Autenticación
- Proxy/Reverse proxy automático
- Backup/Restore
- Auto-updates
- Métricas avanzadas

---

## Servicios soportados

- [x] MediaCheky (core)
- [ ] Radarr (movies)
- [ ] Sonarr (TV shows)
- [ ] Jellyfin (media server)
- [ ] Prowlarr (indexer manager)
- [ ] qBittorrent (torrent client)
- [ ] Jellyseerr (request management)
- [ ] Bazarr (subtitles)
- [ ] Jellystat (statistics)

