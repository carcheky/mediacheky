# Propuesta 2: Stack Minimalista Python (FastAPI + HTMX)

## 🎯 Visión General

Crear una aplicación web ligera y eficiente usando Python con FastAPI en el backend y HTMX en el frontend para interactividad sin JavaScript pesado. Esta propuesta prioriza simplicidad, bajo uso de recursos y facilidad de deployment.

## 🏗️ Arquitectura

### Stack Tecnológico

#### Backend
- **Framework**: FastAPI (Python 3.12+)
- **ORM**: SQLAlchemy 2.0
- **Base de datos**: SQLite (desarrollo) / PostgreSQL (producción)
- **Scheduler**: APScheduler
- **Validación**: Pydantic V2
- **HTTP Client**: httpx (async)
- **Templates**: Jinja2

#### Frontend
- **Hypermedia**: HTMX
- **CSS Framework**: Pico.css o DaisyUI
- **Icons**: Lucide Icons
- **Charts**: Chart.js (minimal JS)
- **Notifications**: HTMX extensions

#### DevOps
- **Container**: Docker single-stage build
- **ASGI Server**: Uvicorn
- **Reverse Proxy**: Caddy (opcional)

## 📐 Estructura del Proyecto

```plaintext
keepercheky/
├── app/
│   ├── main.py                 # FastAPI app entry
│   ├── config.py               # Pydantic settings
│   ├── database.py             # SQLAlchemy setup
│   │
│   ├── models/                 # SQLAlchemy models
│   │   ├── __init__.py
│   │   ├── media.py
│   │   ├── schedule.py
│   │   ├── config.py
│   │   └── history.py
│   │
│   ├── schemas/                # Pydantic schemas
│   │   ├── media.py
│   │   ├── schedule.py
│   │   └── config.py
│   │
│   ├── services/               # Business logic
│   │   ├── clients/
│   │   │   ├── base.py
│   │   │   ├── radarr.py
│   │   │   ├── sonarr.py
│   │   │   ├── jellyfin.py
│   │   │   └── jellyseerr.py
│   │   ├── cleanup/
│   │   │   ├── strategies.py
│   │   │   ├── media_cleanup.py
│   │   │   ├── tag_cleanup.py
│   │   │   └── episode_cleanup.py
│   │   ├── scheduler.py
│   │   └── stats.py
│   │
│   ├── routes/                 # API endpoints + pages
│   │   ├── __init__.py
│   │   ├── dashboard.py
│   │   ├── media.py
│   │   ├── schedules.py
│   │   ├── settings.py
│   │   └── logs.py
│   │
│   ├── templates/              # Jinja2 templates
│   │   ├── base.html
│   │   ├── components/
│   │   │   ├── navbar.html
│   │   │   ├── sidebar.html
│   │   │   ├── card.html
│   │   │   └── table.html
│   │   ├── pages/
│   │   │   ├── dashboard.html
│   │   │   ├── media.html
│   │   │   ├── schedules.html
│   │   │   ├── settings.html
│   │   │   └── logs.html
│   │   └── partials/           # HTMX fragments
│   │       ├── media_row.html
│   │       ├── stats_widget.html
│   │       └── log_entry.html
│   │
│   ├── static/
│   │   ├── css/
│   │   │   └── custom.css
│   │   ├── js/
│   │   │   └── charts.js
│   │   └── images/
│   │
│   └── utils/
│       ├── logger.py
│       ├── filesystem.py
│       └── helpers.py
│
├── migrations/                 # Alembic migrations
│   └── versions/
│
├── tests/
│   ├── test_services/
│   └── test_routes/
│
├── Dockerfile
├── docker-compose.yml
├── requirements.txt
├── pyproject.toml
└── README.md
```

## 🎨 Interfaz de Usuario (HTMX)

### Arquitectura Frontend

**Filosofía**: Hypermedia-Driven Application (HDA)
- El servidor envía HTML completo
- HTMX intercambia fragmentos de HTML sin full page reloads
- Mínimo JavaScript (solo para gráficos)
- Progressive Enhancement

### Páginas Principales

#### 1. Dashboard (`/`)

```html
<!-- Template: dashboard.html -->
<div class="container">
  <!-- Auto-refresh cada 30s con HTMX -->
  <div hx-get="/api/stats" 
       hx-trigger="every 30s" 
       hx-swap="innerHTML">
    
    <!-- Stats Cards -->
    <div class="grid">
      <div class="card">
        <h3>Espacio en Disco</h3>
        <canvas id="disk-chart"></canvas>
        <p>1.2TB libre de 4TB</p>
      </div>
      
      <div class="card">
        <h3>Media por Eliminar</h3>
        <h2>23</h2>
        <p>En los próximos 14 días</p>
      </div>
      
      <div class="card">
        <h3>Última Limpieza</h3>
        <p>Hace 2 horas</p>
        <span class="badge success">Exitosa</span>
      </div>
    </div>
    
    <!-- Services Status -->
    <div class="services-status">
      <span class="badge success">Radarr</span>
      <span class="badge success">Sonarr</span>
      <span class="badge warning">Jellyfin</span>
      <span class="badge error">Jellyseerr</span>
    </div>
  </div>
</div>
```

#### 2. Media Management (`/media`)

```html
<!-- Filtros con HTMX -->
<form hx-get="/media/filter" 
      hx-trigger="change, keyup delay:500ms" 
      hx-target="#media-table">
  
  <input type="search" name="q" placeholder="Buscar...">
  
  <select name="type">
    <option value="all">Todos</option>
    <option value="movie">Películas</option>
    <option value="series">Series</option>
  </select>
  
  <select name="status">
    <option value="all">Todos los estados</option>
    <option value="leaving-soon">Leaving Soon</option>
    <option value="excluded">Excluidos</option>
  </select>
</form>

<!-- Tabla de resultados -->
<table id="media-table">
  <thead>
    <tr>
      <th>Poster</th>
      <th>Título</th>
      <th>Tipo</th>
      <th>Tamaño</th>
      <th>Edad</th>
      <th>Eliminar en</th>
      <th>Acciones</th>
    </tr>
  </thead>
  <tbody>
    {% for item in media_items %}
    <tr hx-get="/media/{{ item.id }}/row" 
        hx-trigger="refresh from:body">
      <td><img src="{{ item.poster }}" width="50"></td>
      <td>{{ item.title }}</td>
      <td>{{ item.type }}</td>
      <td>{{ item.size }}</td>
      <td>{{ item.age_days }}d</td>
      <td>{{ item.days_until_deletion }}d</td>
      <td>
        <button hx-post="/media/{{ item.id }}/exclude"
                hx-swap="outerHTML"
                class="btn btn-sm">
          Excluir
        </button>
        <button hx-delete="/media/{{ item.id }}"
                hx-confirm="¿Eliminar este media?"
                class="btn btn-sm btn-danger">
          Eliminar
        </button>
      </td>
    </tr>
    {% endfor %}
  </tbody>
</table>

<!-- Paginación con HTMX -->
<div class="pagination">
  <button hx-get="/media?page={{ page - 1 }}"
          hx-target="#media-table"
          {% if page == 1 %}disabled{% endif %}>
    Anterior
  </button>
  <span>Página {{ page }} de {{ total_pages }}</span>
  <button hx-get="/media?page={{ page + 1 }}"
          hx-target="#media-table"
          {% if page == total_pages %}disabled{% endif %}>
    Siguiente
  </button>
</div>
```

#### 3. Schedules (`/schedules`)

```html
<div class="schedules-container">
  <!-- Crear nuevo schedule -->
  <button hx-get="/schedules/new-form"
          hx-target="#schedule-form"
          hx-swap="innerHTML">
    + Nuevo Schedule
  </button>
  
  <div id="schedule-form"></div>
  
  <!-- Lista de schedules -->
  <div class="schedules-list">
    {% for schedule in schedules %}
    <div class="card schedule-card">
      <h3>{{ schedule.name }}</h3>
      <p>Tipo: {{ schedule.type }}</p>
      <p>Próxima ejecución: {{ schedule.next_run }}</p>
      
      <div class="actions">
        <button hx-post="/schedules/{{ schedule.id }}/run"
                hx-confirm="¿Ejecutar ahora?"
                class="btn btn-primary">
          Ejecutar Ahora
        </button>
        
        <button hx-put="/schedules/{{ schedule.id }}/toggle"
                hx-swap="outerHTML">
          {% if schedule.enabled %}Desactivar{% else %}Activar{% endif %}
        </button>
        
        <button hx-get="/schedules/{{ schedule.id }}/edit"
                hx-target="#schedule-form">
          Editar
        </button>
        
        <button hx-delete="/schedules/{{ schedule.id }}"
                hx-confirm="¿Eliminar schedule?"
                class="btn btn-danger">
          Eliminar
        </button>
      </div>
      
      <!-- Preview de qué se eliminará -->
      <details>
        <summary>Preview ({{ schedule.affected_count }} items)</summary>
        <div hx-get="/schedules/{{ schedule.id }}/preview"
             hx-trigger="click once"
             hx-swap="innerHTML">
          Cargando...
        </div>
      </details>
    </div>
    {% endfor %}
  </div>
</div>
```

#### 4. Settings (`/settings`)

```html
<div class="settings-tabs">
  <!-- Tabs con HTMX -->
  <nav>
    <button hx-get="/settings/general"
            hx-target="#settings-content"
            class="active">
      General
    </button>
    <button hx-get="/settings/clients"
            hx-target="#settings-content">
      Clientes
    </button>
    <button hx-get="/settings/filesystem"
            hx-target="#settings-content">
      File System
    </button>
  </nav>
  
  <div id="settings-content">
    <!-- Formulario de settings general -->
    <form hx-post="/settings/general"
          hx-swap="outerHTML">
      
      <label>
        <input type="checkbox" name="dry_run" {% if config.dry_run %}checked{% endif %}>
        Modo Dry Run
      </label>
      
      <label>
        Leaving Soon (días)
        <input type="number" name="leaving_soon_days" value="{{ config.leaving_soon_days }}">
      </label>
      
      <label>
        Tags de Exclusión (separados por coma)
        <input type="text" name="exclusion_tags" value="{{ config.exclusion_tags|join(',') }}">
      </label>
      
      <button type="submit" class="btn btn-primary">
        Guardar Cambios
      </button>
    </form>
  </div>
</div>

<!-- Configuración de clientes -->
<div class="clients-config">
  {% for client in ['radarr', 'sonarr', 'jellyfin', 'jellyseerr'] %}
  <details class="card">
    <summary>{{ client|title }} {% if clients[client].enabled %}✅{% else %}⚪{% endif %}</summary>
    
    <form hx-post="/settings/clients/{{ client }}"
          hx-swap="outerHTML">
      
      <label>
        <input type="checkbox" name="enabled" {% if clients[client].enabled %}checked{% endif %}>
        Habilitar {{ client|title }}
      </label>
      
      <label>
        URL
        <input type="url" name="url" value="{{ clients[client].url }}" required>
      </label>
      
      <label>
        API Key
        <input type="password" name="api_key" value="{{ clients[client].api_key }}" required>
      </label>
      
      <button type="button" 
              hx-post="/settings/clients/{{ client }}/test"
              hx-include="closest form"
              hx-target="next .test-result">
        Probar Conexión
      </button>
      <div class="test-result"></div>
      
      <button type="submit" class="btn btn-primary">Guardar</button>
    </form>
  </details>
  {% endfor %}
</div>
```

#### 5. Logs (`/logs`)

```html
<div class="logs-viewer">
  <!-- Filtros -->
  <div class="filters">
    <select name="level" 
            hx-get="/logs/filter"
            hx-trigger="change"
            hx-target="#logs-container">
      <option value="all">Todos los niveles</option>
      <option value="INFO">INFO</option>
      <option value="WARNING">WARNING</option>
      <option value="ERROR">ERROR</option>
      <option value="DEBUG">DEBUG</option>
    </select>
    
    <input type="search" 
           name="search"
           placeholder="Buscar en logs..."
           hx-get="/logs/filter"
           hx-trigger="keyup changed delay:500ms"
           hx-target="#logs-container">
    
    <button hx-get="/logs/export">Exportar Logs</button>
  </div>
  
  <!-- Logs con auto-refresh -->
  <div id="logs-container"
       hx-get="/logs/stream"
       hx-trigger="every 5s"
       hx-swap="beforeend"
       class="logs-list">
    
    {% for log in logs %}
    <div class="log-entry log-{{ log.level }}">
      <span class="timestamp">{{ log.timestamp }}</span>
      <span class="level">{{ log.level }}</span>
      <span class="message">{{ log.message }}</span>
    </div>
    {% endfor %}
  </div>
</div>

<!-- Auto-scroll to bottom -->
<script>
  document.body.addEventListener('htmx:afterSwap', function(evt) {
    if (evt.detail.target.id === 'logs-container') {
      evt.detail.target.scrollTop = evt.detail.target.scrollHeight;
    }
  });
</script>
```

## ⚙️ Backend (FastAPI)

### Estructura de Rutas

```python
# app/main.py
from fastapi import FastAPI
from fastapi.staticfiles import StaticFiles
from fastapi.templating import Jinja2Templates

app = FastAPI(title="KeeperCheky")

# Montaje de archivos estáticos
app.mount("/static", StaticFiles(directory="app/static"), name="static")

# Templates
templates = Jinja2Templates(directory="app/templates")

# Rutas
from app.routes import dashboard, media, schedules, settings, logs

app.include_router(dashboard.router)
app.include_router(media.router, prefix="/media")
app.include_router(schedules.router, prefix="/schedules")
app.include_router(settings.router, prefix="/settings")
app.include_router(logs.router, prefix="/logs")
```

### Ejemplo de Ruta (Media)

```python
# app/routes/media.py
from fastapi import APIRouter, Request, Depends
from fastapi.responses import HTMLResponse
from sqlalchemy.orm import Session

from app.database import get_db
from app.services.media import MediaService
from app.schemas.media import MediaFilter

router = APIRouter()

@router.get("/", response_class=HTMLResponse)
async def media_page(request: Request, db: Session = Depends(get_db)):
    """Renderiza la página completa de media"""
    media_service = MediaService(db)
    media_items = await media_service.get_all()
    
    return templates.TemplateResponse(
        "pages/media.html",
        {
            "request": request,
            "media_items": media_items,
            "page": 1,
            "total_pages": 10
        }
    )

@router.get("/filter", response_class=HTMLResponse)
async def filter_media(
    request: Request,
    q: str = "",
    type: str = "all",
    status: str = "all",
    page: int = 1,
    db: Session = Depends(get_db)
):
    """Retorna solo la tabla filtrada (para HTMX)"""
    media_service = MediaService(db)
    
    filters = MediaFilter(query=q, type=type, status=status)
    media_items = await media_service.filter(filters, page=page)
    
    return templates.TemplateResponse(
        "partials/media_table.html",
        {
            "request": request,
            "media_items": media_items,
            "page": page,
            "total_pages": media_items.total_pages
        }
    )

@router.post("/{media_id}/exclude")
async def exclude_media(media_id: int, db: Session = Depends(get_db)):
    """Excluir media de limpieza"""
    media_service = MediaService(db)
    await media_service.exclude(media_id)
    
    # Retornar el row actualizado
    item = await media_service.get_by_id(media_id)
    return templates.TemplateResponse(
        "partials/media_row.html",
        {"item": item}
    )

@router.delete("/{media_id}")
async def delete_media(media_id: int, db: Session = Depends(get_db)):
    """Eliminar media manualmente"""
    media_service = MediaService(db)
    await media_service.delete(media_id)
    
    # Retornar vacío para remover el row
    return HTMLResponse("")
```

### Servicios

```python
# app/services/cleanup/media_cleanup.py
from datetime import datetime, timedelta
from typing import List

from app.models.media import MediaItem
from app.services.clients.base import MediaClient
from app.utils.filesystem import get_free_space_percentage

class MediaCleanupService:
    def __init__(
        self,
        radarr_client: MediaClient,
        sonarr_client: MediaClient,
        jellyfin_client: MediaClient,
        config: CleanupConfig
    ):
        self.radarr = radarr_client
        self.sonarr = sonarr_client
        self.jellyfin = jellyfin_client
        self.config = config
    
    async def get_media_to_delete(self) -> List[MediaItem]:
        """Determina qué media debe eliminarse"""
        free_space_pct = get_free_space_percentage(self.config.check_dir)
        
        # Determinar expiration time según espacio libre
        expiration_days = self._determine_expiration(free_space_pct)
        
        if not expiration_days:
            return []
        
        cutoff_date = datetime.now() - timedelta(days=expiration_days)
        
        # Obtener media de Radarr/Sonarr
        movies = await self.radarr.get_library()
        series = await self.sonarr.get_library()
        
        to_delete = []
        
        for item in movies + series:
            # Verificar edad
            if item.added_date < cutoff_date:
                # Verificar exclusiones
                if not self._is_excluded(item):
                    to_delete.append(item)
        
        return to_delete
    
    def _determine_expiration(self, free_pct: float) -> int | None:
        """Determina días de expiración según espacio libre"""
        for threshold, days in sorted(
            self.config.expiration_map.items(),
            key=lambda x: x[0]
        ):
            if free_pct < threshold:
                return days
        return None
    
    def _is_excluded(self, item: MediaItem) -> bool:
        """Verifica si el item tiene tags de exclusión"""
        return any(
            tag in self.config.exclusion_tags
            for tag in item.tags
        )
    
    async def execute_cleanup(self, dry_run: bool = True):
        """Ejecuta la limpieza"""
        items_to_delete = await self.get_media_to_delete()
        
        for item in items_to_delete:
            if dry_run:
                logger.info(f"[DRY RUN] Would delete: {item.title}")
            else:
                logger.info(f"Deleting: {item.title}")
                
                # Eliminar en Jellyfin
                if self.jellyfin:
                    await self.jellyfin.delete_item(item.jellyfin_id)
                
                # Eliminar en *arr
                if item.type == "movie":
                    await self.radarr.delete_item(item.radarr_id)
                else:
                    await self.sonarr.delete_item(item.sonarr_id)
```

### Scheduler

```python
# app/services/scheduler.py
from apscheduler.schedulers.asyncio import AsyncIOScheduler
from apscheduler.triggers.cron import CronTrigger

class CleanupScheduler:
    def __init__(self):
        self.scheduler = AsyncIOScheduler()
    
    def start(self):
        """Inicia el scheduler"""
        # Media cleanup cada hora
        self.scheduler.add_job(
            self.run_media_cleanup,
            CronTrigger(hour="*"),
            id="media_cleanup"
        )
        
        # Tag cleanup según schedules configurados
        self.scheduler.add_job(
            self.run_tag_cleanup,
            CronTrigger(hour="2"),
            id="tag_cleanup"
        )
        
        self.scheduler.start()
    
    async def run_media_cleanup(self):
        """Ejecuta limpieza de media"""
        from app.services.cleanup.media_cleanup import MediaCleanupService
        
        service = MediaCleanupService(...)
        await service.execute_cleanup(dry_run=False)
    
    def add_custom_schedule(self, schedule_id: str, cron: str, callback):
        """Añade un schedule personalizado"""
        self.scheduler.add_job(
            callback,
            CronTrigger.from_crontab(cron),
            id=schedule_id
        )
```

## 🐳 Docker

### Dockerfile

```dockerfile
FROM python:3.12-slim

WORKDIR /app

# Instalar dependencias del sistema
RUN apt-get update && apt-get install -y \
    gcc \
    && rm -rf /var/lib/apt/lists/*

# Copiar requirements
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copiar aplicación
COPY app/ ./app/

# Usuario no-root
RUN useradd -m -u 1000 keepercheky && \
    chown -R keepercheky:keepercheky /app
USER keepercheky

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s \
  CMD python -c "import requests; requests.get('http://localhost:8000/health')"

# Exponer puerto
EXPOSE 8000

# Comando
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]
```

### docker-compose.yml

```yaml
version: '3.8'

services:
  keepercheky:
    image: keepercheky:latest
    container_name: keepercheky
    ports:
      - "8000:8000"
    environment:
      - DATABASE_URL=sqlite:///data/keepercheky.db
      - LOG_LEVEL=INFO
    volumes:
      - ./data:/app/data
      - ./logs:/app/logs
      - /share_media:/data/media  # Mismo mount que Radarr/Sonarr
    restart: unless-stopped
```

## 🎯 Ventajas de esta Propuesta

### ✅ Pros

1. **Extremadamente ligero**: ~100-150MB de imagen Docker
2. **Bajo uso de RAM**: 50-150MB en runtime
3. **Simplicidad**: Un solo lenguaje (Python) para todo
4. **Rápido desarrollo**: FastAPI es muy productivo
5. **Sin JavaScript complejo**: HTMX maneja toda la interactividad
6. **Performance**: FastAPI + async es muy rápido
7. **Type hints**: Python 3.12+ con type checking
8. **Fácil deployment**: Single container
9. **Progressive enhancement**: Funciona sin JS
10. **SEO-friendly**: Server-side rendering

### ⚠️ Contras

1. **UI menos "moderna"**: No SPA, recargas parciales
2. **HTMX menos conocido**: Curva de aprendizaje para equipo
3. **Limitaciones de interactividad**: Comparado con React/Vue
4. **Ecosistema Python**: Menos opciones de UI components
5. **SQLite limitations**: Para producción necesita PostgreSQL

## 📊 Estimación de Recursos

### Desarrollo
- **Tiempo estimado**: 3-4 semanas (1 desarrollador)
- **Dificultad**: Media

### Runtime
- **RAM**: 50-150MB
- **CPU**: 0.5-1 core
- **Disco**: 150MB (imagen Docker)
- **DB**: 10-50MB (SQLite) o PostgreSQL

## 🛣️ Roadmap

### Fase 1: Core (1.5 semanas)
- [ ] Setup FastAPI + SQLAlchemy
- [ ] Modelos de datos
- [ ] Clientes básicos (Radarr, Sonarr)
- [ ] UI base con HTMX

### Fase 2: Features (1.5 semanas)
- [ ] Cleanup logic
- [ ] Scheduler
- [ ] Todas las páginas
- [ ] Todos los clientes

### Fase 3: Polish (1 semana)
- [ ] Testing
- [ ] Docker optimization
- [ ] Documentación
- [ ] UI refinement

## 📝 Conclusión

**Ideal para**: Deployments con recursos limitados, usuarios que prefieren simplicidad, y equipos con experiencia en Python.

**No recomendado si**: Necesitas una SPA moderna con mucha interactividad en el frontend.
