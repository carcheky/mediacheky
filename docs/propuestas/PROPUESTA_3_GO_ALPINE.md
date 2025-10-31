# Propuesta 3: Stack Go + Alpine.js (Rendimiento Extremo)

## 🎯 Visión General

Crear una aplicación ultra-eficiente usando Go en el backend con Alpine.js en el frontend. Esta propuesta maximiza el rendimiento, minimiza el uso de recursos y ofrece tiempos de respuesta extremadamente rápidos.

## 🏗️ Arquitectura

### Stack Tecnológico

#### Backend

- **Lenguaje**: Go 1.22+
- **Framework**: Fiber v2 (Express-like para Go)
- **ORM**: GORM v2
- **Base de datos**: SQLite (embedded) / PostgreSQL (opcional)
- **Templating**: Go html/template
- **Scheduler**: go-cron
- **HTTP Client**: resty

#### Frontend

- **Framework JS**: Alpine.js 3.x (15kb)
- **CSS**: Tailwind CSS
- **Icons**: Heroicons
- **Charts**: Chart.js
- **Components**: Alpine Components

#### DevOps

- **Container**: Multi-stage Docker (scratch base)
- **Binario**: Single static binary
- **Tamaño**: ~15-25MB imagen final

## 📐 Estructura del Proyecto

```plaintext
keepercheky/
├── cmd/
│   └── server/
│       └── main.go            # Entry point
│
├── internal/
│   ├── config/
│   │   ├── config.go
│   │   └── env.go
│   │
│   ├── models/
│   │   ├── media.go
│   │   ├── schedule.go
│   │   └── config.go
│   │
│   ├── repository/
│   │   ├── media_repo.go
│   │   ├── schedule_repo.go
│   │   └── config_repo.go
│   │
│   ├── service/
│   │   ├── clients/
│   │   │   ├── client.go      # Interface
│   │   │   ├── radarr.go
│   │   │   ├── sonarr.go
│   │   │   ├── jellyfin.go
│   │   │   └── jellyseerr.go
│   │   ├── cleanup/
│   │   │   ├── media.go
│   │   │   ├── tags.go
│   │   │   └── episodes.go
│   │   ├── scheduler.go
│   │   └── stats.go
│   │
│   ├── handler/
│   │   ├── dashboard.go
│   │   ├── media.go
│   │   ├── schedules.go
│   │   ├── settings.go
│   │   └── logs.go
│   │
│   └── middleware/
│       ├── logger.go
│       └── error.go
│
├── web/
│   ├── templates/
│   │   ├── layouts/
│   │   │   └── base.html
│   │   ├── pages/
│   │   │   ├── dashboard.html
│   │   │   ├── media.html
│   │   │   ├── schedules.html
│   │   │   ├── settings.html
│   │   │   └── logs.html
│   │   └── components/
│   │       ├── navbar.html
│   │       ├── card.html
│   │       └── table.html
│   │
│   └── static/
│       ├── css/
│       │   └── styles.css
│       ├── js/
│       │   └── app.js
│       └── images/
│
├── pkg/                       # Shared utilities
│   ├── filesystem/
│   │   └── utils.go
│   └── logger/
│       └── logger.go
│
├── migrations/
│   └── 001_initial.sql
│
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
└── Makefile
```

## 🎨 Interfaz de Usuario (Alpine.js)

### Filosofía Frontend

- **Reactive**: Alpine.js para reactividad ligera
- **Server-side rendered**: Go templates
- **Progressive enhancement**: Funciona sin JS
- **API REST**: JSON para datos dinámicos

### Ejemplos de UI

#### 1. Dashboard

```html
<!-- templates/pages/dashboard.html -->
<div x-data="dashboard()" x-init="init()">
  <!-- Stats Grid -->
  <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
    <!-- Disk Space Card -->
    <div class="card">
      <h3 class="text-lg font-bold">Espacio en Disco</h3>
      <canvas id="diskChart"></canvas>
      <p class="mt-2" x-text="diskInfo"></p>
    </div>
    
    <!-- Media to Delete Card -->
    <div class="card">
      <h3 class="text-lg font-bold">Media por Eliminar</h3>
      <div class="text-4xl font-bold" x-text="mediaToDelete.count"></div>
      <p class="text-gray-600">En los próximos <span x-text="mediaToDelete.days"></span> días</p>
    </div>
    
    <!-- Last Cleanup Card -->
    <div class="card">
      <h3 class="text-lg font-bold">Última Limpieza</h3>
      <p x-text="lastCleanup.time"></p>
      <span 
        class="badge" 
        :class="lastCleanup.success ? 'badge-success' : 'badge-error'"
        x-text="lastCleanup.status">
      </span>
    </div>
  </div>
  
  <!-- Services Status -->
  <div class="mt-6">
    <h3 class="text-lg font-bold mb-4">Estado de Servicios</h3>
    <div class="flex gap-2">
      <template x-for="service in services" :key="service.name">
        <span 
          class="badge"
          :class="{
            'badge-success': service.status === 'online',
            'badge-warning': service.status === 'degraded',
            'badge-error': service.status === 'offline'
          }"
          x-text="service.name">
        </span>
      </template>
    </div>
  </div>
  
  <!-- Recent Activity -->
  <div class="mt-6">
    <h3 class="text-lg font-bold mb-4">Actividad Reciente</h3>
    <div class="space-y-2">
      <template x-for="activity in recentActivity" :key="activity.id">
        <div class="flex items-center justify-between p-3 bg-gray-50 rounded">
          <div>
            <p class="font-medium" x-text="activity.title"></p>
            <p class="text-sm text-gray-600" x-text="activity.description"></p>
          </div>
          <span class="text-sm text-gray-500" x-text="activity.time"></span>
        </div>
      </template>
    </div>
  </div>
</div>

<script>
function dashboard() {
  return {
    diskInfo: '',
    mediaToDelete: { count: 0, days: 14 },
    lastCleanup: { time: '', status: '', success: false },
    services: [],
    recentActivity: [],
    
    async init() {
      await this.fetchStats();
      // Poll every 30 seconds
      setInterval(() => this.fetchStats(), 30000);
    },
    
    async fetchStats() {
      const response = await fetch('/api/stats');
      const data = await response.json();
      
      this.diskInfo = data.disk_info;
      this.mediaToDelete = data.media_to_delete;
      this.lastCleanup = data.last_cleanup;
      this.services = data.services;
      this.recentActivity = data.recent_activity;
      
      this.updateDiskChart(data.disk_usage);
    },
    
    updateDiskChart(data) {
      // Chart.js update logic
    }
  }
}
</script>
```

#### 2. Media Management

```html
<div x-data="mediaManager()" x-init="init()">
  <!-- Filters -->
  <div class="filters mb-6">
    <input 
      type="search" 
      placeholder="Buscar media..."
      x-model="filters.search"
      @input.debounce.500ms="fetchMedia()"
      class="input">
    
    <select x-model="filters.type" @change="fetchMedia()" class="select">
      <option value="all">Todos</option>
      <option value="movie">Películas</option>
      <option value="series">Series</option>
    </select>
    
    <select x-model="filters.status" @change="fetchMedia()" class="select">
      <option value="all">Todos los estados</option>
      <option value="leaving-soon">Leaving Soon</option>
      <option value="excluded">Excluidos</option>
    </select>
  </div>
  
  <!-- Loading State -->
  <div x-show="loading" class="text-center py-8">
    <div class="spinner"></div>
  </div>
  
  <!-- Media Grid/List Toggle -->
  <div class="flex justify-between mb-4">
    <div class="btn-group">
      <button 
        @click="viewMode = 'grid'" 
        :class="{ 'active': viewMode === 'grid' }">
        Grid
      </button>
      <button 
        @click="viewMode = 'list'" 
        :class="{ 'active': viewMode === 'list' }">
        List
      </button>
    </div>
    
    <p class="text-gray-600">
      <span x-text="media.length"></span> items
    </p>
  </div>
  
  <!-- Grid View -->
  <div 
    x-show="viewMode === 'grid'" 
    class="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
    <template x-for="item in media" :key="item.id">
      <div class="card cursor-pointer" @click="showDetails(item)">
        <img :src="item.poster_url" :alt="item.title" class="w-full h-auto rounded">
        <h4 class="mt-2 font-medium truncate" x-text="item.title"></h4>
        <p class="text-sm text-gray-600">
          <span x-text="item.type"></span> · 
          <span x-text="item.size"></span>
        </p>
        <div class="mt-2 flex gap-2">
          <button 
            @click.stop="excludeItem(item.id)"
            class="btn btn-sm">
            Excluir
          </button>
          <button 
            @click.stop="deleteItem(item.id)"
            class="btn btn-sm btn-danger">
            Eliminar
          </button>
        </div>
      </div>
    </template>
  </div>
  
  <!-- List View -->
  <div x-show="viewMode === 'list'">
    <table class="table">
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
        <template x-for="item in media" :key="item.id">
          <tr>
            <td><img :src="item.poster_url" class="w-12 h-auto"></td>
            <td x-text="item.title"></td>
            <td x-text="item.type"></td>
            <td x-text="item.size"></td>
            <td><span x-text="item.age_days"></span>d</td>
            <td><span x-text="item.days_until_deletion"></span>d</td>
            <td>
              <button @click="excludeItem(item.id)" class="btn btn-sm">Excluir</button>
              <button @click="deleteItem(item.id)" class="btn btn-sm btn-danger">Eliminar</button>
            </td>
          </tr>
        </template>
      </tbody>
    </table>
  </div>
  
  <!-- Pagination -->
  <div class="flex justify-center mt-6 gap-2">
    <button 
      @click="changePage(page - 1)" 
      :disabled="page === 1"
      class="btn">
      Anterior
    </button>
    <span class="py-2 px-4">Página <span x-text="page"></span> de <span x-text="totalPages"></span></span>
    <button 
      @click="changePage(page + 1)" 
      :disabled="page === totalPages"
      class="btn">
      Siguiente
    </button>
  </div>
  
  <!-- Details Modal -->
  <div 
    x-show="detailsModal.show" 
    @click.away="detailsModal.show = false"
    class="modal">
    <div class="modal-content" @click.stop>
      <h2 x-text="detailsModal.item.title"></h2>
      <!-- More details -->
    </div>
  </div>
</div>

<script>
function mediaManager() {
  return {
    media: [],
    loading: false,
    viewMode: 'grid',
    filters: {
      search: '',
      type: 'all',
      status: 'all'
    },
    page: 1,
    totalPages: 1,
    detailsModal: {
      show: false,
      item: null
    },
    
    async init() {
      await this.fetchMedia();
    },
    
    async fetchMedia() {
      this.loading = true;
      const params = new URLSearchParams({
        ...this.filters,
        page: this.page
      });
      
      const response = await fetch(`/api/media?${params}`);
      const data = await response.json();
      
      this.media = data.items;
      this.totalPages = data.total_pages;
      this.loading = false;
    },
    
    async excludeItem(id) {
      if (!confirm('¿Excluir este media de la limpieza?')) return;
      
      await fetch(`/api/media/${id}/exclude`, { method: 'POST' });
      await this.fetchMedia();
    },
    
    async deleteItem(id) {
      if (!confirm('¿Eliminar este media permanentemente?')) return;
      
      await fetch(`/api/media/${id}`, { method: 'DELETE' });
      await this.fetchMedia();
    },
    
    showDetails(item) {
      this.detailsModal.item = item;
      this.detailsModal.show = true;
    },
    
    changePage(newPage) {
      this.page = newPage;
      this.fetchMedia();
    }
  }
}
</script>
```

#### 3. Schedules

```html
<div x-data="schedulesManager()" x-init="init()">
  <!-- Create New Schedule -->
  <button 
    @click="showCreateForm = true" 
    class="btn btn-primary mb-6">
    + Nuevo Schedule
  </button>
  
  <!-- Schedules List -->
  <div class="space-y-4">
    <template x-for="schedule in schedules" :key="schedule.id">
      <div class="card">
        <div class="flex justify-between items-start">
          <div>
            <h3 class="text-lg font-bold" x-text="schedule.name"></h3>
            <p class="text-gray-600">Tipo: <span x-text="schedule.type"></span></p>
            <p class="text-gray-600">Próxima ejecución: <span x-text="schedule.next_run"></span></p>
          </div>
          
          <div class="flex gap-2">
            <button 
              @click="runSchedule(schedule.id)" 
              class="btn btn-sm btn-primary">
              Ejecutar Ahora
            </button>
            
            <button 
              @click="toggleSchedule(schedule.id)" 
              class="btn btn-sm">
              <span x-text="schedule.enabled ? 'Desactivar' : 'Activar'"></span>
            </button>
            
            <button 
              @click="editSchedule(schedule)" 
              class="btn btn-sm">
              Editar
            </button>
            
            <button 
              @click="deleteSchedule(schedule.id)" 
              class="btn btn-sm btn-danger">
              Eliminar
            </button>
          </div>
        </div>
        
        <!-- Preview -->
        <details class="mt-4">
          <summary class="cursor-pointer">
            Preview (<span x-text="schedule.affected_count"></span> items)
          </summary>
          <div 
            x-data="{ preview: null }"
            x-init="fetch(`/api/schedules/${schedule.id}/preview`)
              .then(r => r.json())
              .then(d => preview = d)">
            <template x-if="preview">
              <ul class="mt-2 space-y-1">
                <template x-for="item in preview.items" :key="item.id">
                  <li x-text="item.title"></li>
                </template>
              </ul>
            </template>
          </div>
        </details>
      </div>
    </template>
  </div>
  
  <!-- Create/Edit Modal -->
  <div 
    x-show="showCreateForm || editingSchedule" 
    class="modal">
    <div class="modal-content" @click.stop>
      <h2 x-text="editingSchedule ? 'Editar Schedule' : 'Nuevo Schedule'"></h2>
      
      <form @submit.prevent="saveSchedule()">
        <label>
          Nombre
          <input type="text" x-model="form.name" required class="input">
        </label>
        
        <label>
          Tipo
          <select x-model="form.type" required class="select">
            <option value="media">Media Cleanup</option>
            <option value="tag">Tag-based Cleanup</option>
            <option value="episode">Episode Cleanup</option>
          </select>
        </label>
        
        <!-- Dynamic fields based on type -->
        <template x-if="form.type === 'tag'">
          <label>
            Tag
            <input type="text" x-model="form.tag" class="input">
          </label>
        </template>
        
        <label>
          Cron Expression
          <input type="text" x-model="form.cron" placeholder="0 * * * *" class="input">
        </label>
        
        <div class="flex gap-2 mt-4">
          <button type="submit" class="btn btn-primary">Guardar</button>
          <button 
            type="button" 
            @click="showCreateForm = false; editingSchedule = null"
            class="btn">
            Cancelar
          </button>
        </div>
      </form>
    </div>
  </div>
</div>

<script>
function schedulesManager() {
  return {
    schedules: [],
    showCreateForm: false,
    editingSchedule: null,
    form: {
      name: '',
      type: 'media',
      tag: '',
      cron: '0 * * * *'
    },
    
    async init() {
      await this.fetchSchedules();
    },
    
    async fetchSchedules() {
      const response = await fetch('/api/schedules');
      this.schedules = await response.json();
    },
    
    async runSchedule(id) {
      if (!confirm('¿Ejecutar este schedule ahora?')) return;
      
      await fetch(`/api/schedules/${id}/run`, { method: 'POST' });
      alert('Schedule ejecutado');
    },
    
    async toggleSchedule(id) {
      await fetch(`/api/schedules/${id}/toggle`, { method: 'PUT' });
      await this.fetchSchedules();
    },
    
    editSchedule(schedule) {
      this.editingSchedule = schedule;
      this.form = { ...schedule };
    },
    
    async deleteSchedule(id) {
      if (!confirm('¿Eliminar este schedule?')) return;
      
      await fetch(`/api/schedules/${id}`, { method: 'DELETE' });
      await this.fetchSchedules();
    },
    
    async saveSchedule() {
      const method = this.editingSchedule ? 'PUT' : 'POST';
      const url = this.editingSchedule 
        ? `/api/schedules/${this.editingSchedule.id}`
        : '/api/schedules';
      
      await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(this.form)
      });
      
      this.showCreateForm = false;
      this.editingSchedule = null;
      await this.fetchSchedules();
    }
  }
}
</script>
```

## ⚙️ Backend (Go)

### Main Application

```go
// cmd/server/main.go
package main

import (
    "log"
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/template/html/v2"
    
    "keepercheky/internal/config"
    "keepercheky/internal/handler"
    "keepercheky/internal/repository"
    "keepercheky/internal/service"
)

func main() {
    // Load config
    cfg := config.Load()
    
    // Initialize database
    db, err := repository.InitDB(cfg.DatabaseURL)
    if err != nil {
        log.Fatal(err)
    }
    
    // Initialize template engine
    engine := html.New("./web/templates", ".html")
    
    // Create Fiber app
    app := fiber.New(fiber.Config{
        Views: engine,
    })
    
    // Static files
    app.Static("/static", "./web/static")
    
    // Initialize repositories
    mediaRepo := repository.NewMediaRepository(db)
    scheduleRepo := repository.NewScheduleRepository(db)
    configRepo := repository.NewConfigRepository(db)
    
    // Initialize services
    radarrClient := service.NewRadarrClient(cfg.Radarr)
    sonarrClient := service.NewSonarrClient(cfg.Sonarr)
    
    cleanupService := service.NewCleanupService(
        mediaRepo,
        radarrClient,
        sonarrClient,
    )
    
    scheduler := service.NewScheduler(cleanupService, scheduleRepo)
    scheduler.Start()
    
    // Initialize handlers
    dashboardHandler := handler.NewDashboardHandler(mediaRepo)
    mediaHandler := handler.NewMediaHandler(mediaRepo, cleanupService)
    schedulesHandler := handler.NewSchedulesHandler(scheduleRepo, scheduler)
    settingsHandler := handler.NewSettingsHandler(configRepo)
    
    // Routes
    setupRoutes(app, dashboardHandler, mediaHandler, schedulesHandler, settingsHandler)
    
    // Start server
    log.Fatal(app.Listen(":8000"))
}

func setupRoutes(
    app *fiber.App,
    dashboard *handler.DashboardHandler,
    media *handler.MediaHandler,
    schedules *handler.SchedulesHandler,
    settings *handler.SettingsHandler,
) {
    // Pages
    app.Get("/", dashboard.Index)
    app.Get("/media", media.Index)
    app.Get("/schedules", schedules.Index)
    app.Get("/settings", settings.Index)
    
    // API
    api := app.Group("/api")
    
    // Stats
    api.Get("/stats", dashboard.GetStats)
    
    // Media
    api.Get("/media", media.GetMedia)
    api.Post("/media/:id/exclude", media.ExcludeMedia)
    api.Delete("/media/:id", media.DeleteMedia)
    
    // Schedules
    api.Get("/schedules", schedules.GetSchedules)
    api.Post("/schedules", schedules.CreateSchedule)
    api.Put("/schedules/:id", schedules.UpdateSchedule)
    api.Delete("/schedules/:id", schedules.DeleteSchedule)
    api.Post("/schedules/:id/run", schedules.RunSchedule)
    api.Put("/schedules/:id/toggle", schedules.ToggleSchedule)
    api.Get("/schedules/:id/preview", schedules.PreviewSchedule)
    
    // Settings
    api.Get("/settings", settings.GetSettings)
    api.Put("/settings", settings.UpdateSettings)
    api.Post("/settings/clients/:client/test", settings.TestClient)
}
```

### Handler Example

```go
// internal/handler/media.go
package handler

import (
    "github.com/gofiber/fiber/v2"
    
    "keepercheky/internal/repository"
    "keepercheky/internal/service"
)

type MediaHandler struct {
    repo    *repository.MediaRepository
    cleanup *service.CleanupService
}

func NewMediaHandler(repo *repository.MediaRepository, cleanup *service.CleanupService) *MediaHandler {
    return &MediaHandler{
        repo:    repo,
        cleanup: cleanup,
    }
}

func (h *MediaHandler) Index(c *fiber.Ctx) error {
    media, err := h.repo.GetAll()
    if err != nil {
        return err
    }
    
    return c.Render("pages/media", fiber.Map{
        "Title": "Media Management",
        "Media": media,
    })
}

func (h *MediaHandler) GetMedia(c *fiber.Ctx) error {
    // Parse query params
    search := c.Query("search")
    mediaType := c.Query("type", "all")
    status := c.Query("status", "all")
    page := c.QueryInt("page", 1)
    
    // Fetch from repository
    media, total, err := h.repo.Filter(search, mediaType, status, page, 20)
    if err != nil {
        return err
    }
    
    return c.JSON(fiber.Map{
        "items":       media,
        "page":        page,
        "total_pages": (total + 19) / 20,
    })
}

func (h *MediaHandler) ExcludeMedia(c *fiber.Ctx) error {
    id, err := c.ParamsInt("id")
    if err != nil {
        return err
    }
    
    if err := h.repo.Exclude(id); err != nil {
        return err
    }
    
    return c.JSON(fiber.Map{"success": true})
}

func (h *MediaHandler) DeleteMedia(c *fiber.Ctx) error {
    id, err := c.ParamsInt("id")
    if err != nil {
        return err
    }
    
    media, err := h.repo.GetByID(id)
    if err != nil {
        return err
    }
    
    // Execute deletion
    if err := h.cleanup.DeleteMedia(media); err != nil {
        return err
    }
    
    return c.JSON(fiber.Map{"success": true})
}
```

### Service Example

```go
// internal/service/cleanup/media.go
package service

import (
    "time"
    
    "keepercheky/internal/models"
    "keepercheky/internal/repository"
    "keepercheky/internal/service/clients"
    "keepercheky/pkg/filesystem"
)

type CleanupService struct {
    mediaRepo *repository.MediaRepository
    radarr    clients.MediaClient
    sonarr    clients.MediaClient
    jellyfin  clients.MediaClient
    config    *CleanupConfig
}

func NewCleanupService(
    mediaRepo *repository.MediaRepository,
    radarr clients.MediaClient,
    sonarr clients.MediaClient,
) *CleanupService {
    return &CleanupService{
        mediaRepo: mediaRepo,
        radarr:    radarr,
        sonarr:    sonarr,
    }
}

func (s *CleanupService) GetMediaToDelete() ([]*models.Media, error) {
    // Get free space percentage
    freeSpacePct := filesystem.GetFreeSpacePercentage("/")
    
    // Determine expiration days
    expirationDays := s.determineExpiration(freeSpacePct)
    if expirationDays == 0 {
        return nil, nil
    }
    
    cutoffDate := time.Now().AddDate(0, 0, -expirationDays)
    
    // Get all media
    allMedia, err := s.mediaRepo.GetAll()
    if err != nil {
        return nil, err
    }
    
    var toDelete []*models.Media
    
    for _, media := range allMedia {
        // Check age
        if media.AddedDate.Before(cutoffDate) {
            // Check exclusions
            if !s.isExcluded(media) {
                toDelete = append(toDelete, media)
            }
        }
    }
    
    return toDelete, nil
}

func (s *CleanupService) determineExpiration(freePct float64) int {
    for threshold, days := range s.config.ExpirationMap {
        if freePct < float64(threshold) {
            return days
        }
    }
    return 0
}

func (s *CleanupService) isExcluded(media *models.Media) bool {
    for _, tag := range media.Tags {
        for _, exclusionTag := range s.config.ExclusionTags {
            if tag == exclusionTag {
                return true
            }
        }
    }
    return false
}

func (s *CleanupService) ExecuteCleanup(dryRun bool) error {
    media, err := s.GetMediaToDelete()
    if err != nil {
        return err
    }
    
    for _, item := range media {
        if dryRun {
            log.Printf("[DRY RUN] Would delete: %s", item.Title)
        } else {
            log.Printf("Deleting: %s", item.Title)
            
            if err := s.DeleteMedia(item); err != nil {
                log.Printf("Error deleting %s: %v", item.Title, err)
            }
        }
    }
    
    return nil
}

func (s *CleanupService) DeleteMedia(media *models.Media) error {
    // Delete from Jellyfin
    if s.jellyfin != nil {
        if err := s.jellyfin.DeleteItem(media.JellyfinID); err != nil {
            return err
        }
    }
    
    // Delete from *arr
    if media.Type == "movie" {
        if err := s.radarr.DeleteItem(media.RadarrID); err != nil {
            return err
        }
    } else {
        if err := s.sonarr.DeleteItem(media.SonarrID); err != nil {
            return err
        }
    }
    
    // Update database
    return s.mediaRepo.Delete(media.ID)
}
```

### Client Interface

```go
// internal/service/clients/client.go
package clients

import "keepercheky/internal/models"

type MediaClient interface {
    TestConnection() error
    GetLibrary() ([]*models.Media, error)
    DeleteItem(id int) error
}

// internal/service/clients/radarr.go
package clients

import (
    "github.com/go-resty/resty/v2"
    
    "keepercheky/internal/models"
)

type RadarrClient struct {
    client *resty.Client
    apiKey string
    url    string
}

func NewRadarrClient(url, apiKey string) *RadarrClient {
    client := resty.New()
    client.SetBaseURL(url)
    client.SetHeader("X-Api-Key", apiKey)
    
    return &RadarrClient{
        client: client,
        apiKey: apiKey,
        url:    url,
    }
}

func (c *RadarrClient) TestConnection() error {
    _, err := c.client.R().Get("/api/v3/system/status")
    return err
}

func (c *RadarrClient) GetLibrary() ([]*models.Media, error) {
    var movies []RadarrMovie
    
    _, err := c.client.R().
        SetResult(&movies).
        Get("/api/v3/movie")
    
    if err != nil {
        return nil, err
    }
    
    var media []*models.Media
    for _, movie := range movies {
        media = append(media, &models.Media{
            Title:      movie.Title,
            Type:       "movie",
            RadarrID:   movie.ID,
            AddedDate:  movie.Added,
            Size:       movie.SizeOnDisk,
            Tags:       movie.Tags,
        })
    }
    
    return media, nil
}

func (c *RadarrClient) DeleteItem(id int) error {
    _, err := c.client.R().
        SetQueryParam("deleteFiles", "true").
        Delete("/api/v3/movie/" + strconv.Itoa(id))
    
    return err
}

type RadarrMovie struct {
    ID         int       `json:"id"`
    Title      string    `json:"title"`
    Added      time.Time `json:"added"`
    SizeOnDisk int64     `json:"sizeOnDisk"`
    Tags       []string  `json:"tags"`
}
```

## 🐳 Docker

### Multi-stage Dockerfile

```dockerfile
# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o keepercheky ./cmd/server

# Final stage
FROM scratch

# Copy binary
COPY --from=builder /app/keepercheky /keepercheky

# Copy templates and static files
COPY --from=builder /app/web /web

# Copy CA certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Expose port
EXPOSE 8000

# Run
ENTRYPOINT ["/keepercheky"]
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
      - DRY_RUN=true
    volumes:
      - ./data:/data
      - /share_media:/data/media
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8000/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

## 🎯 Ventajas de esta Propuesta

### ✅ Pros

1. **Rendimiento extremo**: Go es ultra rápido
2. **Consumo mínimo de recursos**: 20-50MB RAM
3. **Imagen Docker tiny**: 15-25MB total
4. **Binario único**: No dependencias externas
5. **Concurrencia**: Goroutines para tareas paralelas
6. **Type-safe**: Go's type system
7. **Alpine.js ligero**: Solo 15kb de JS
8. **Fast startup**: < 1 segundo
9. **Cross-platform**: Compilable para cualquier OS
10. **Fácil deployment**: Single binary

### ⚠️ Contras

1. **Curva de aprendizaje**: Go puede ser nuevo
2. **Menos librerías UI**: Comparado con React
3. **Templating básico**: html/template es simple
4. **Menos "mágico"**: Más código boilerplate
5. **Ecosistema Go**: Menos familiar para web devs

## 📊 Estimación de Recursos

### Desarrollo

- **Tiempo estimado**: 3-4 semanas
- **Dificultad**: Media-Alta (si no sabes Go)

### Runtime

- **RAM**: 20-50MB
- **CPU**: 0.25-0.5 cores
- **Disco**: 15-25MB (imagen)
- **Startup**: < 1s

## 🛣️ Roadmap

### Fase 1 (1.5 semanas)

- [ ] Setup Go project
- [ ] Fiber app + GORM
- [ ] Clientes básicos
- [ ] UI base con Alpine.js

### Fase 2 (1.5 semanas)

- [ ] Cleanup logic
- [ ] Scheduler
- [ ] Todas las páginas
- [ ] Todos los clientes

### Fase 3 (1 semana)

- [ ] Testing
- [ ] Docker optimization
- [ ] Documentación
- [ ] Performance tuning

## 📝 Conclusión

**Ideal para**: Máxima eficiencia, mínimos recursos, deployments en hardware limitado.

**No recomendado si**: No tienes experiencia con Go o necesitas una UI muy compleja.
