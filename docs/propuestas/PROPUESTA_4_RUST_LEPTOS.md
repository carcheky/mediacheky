# Propuesta 4: Stack Rust + Leptos (Máxima Seguridad y WebAssembly)

## 🎯 Visión General

Reescribir Janitorr usando Rust con Leptos framework, aprovechando WebAssembly para el frontend. Esta propuesta maximiza la seguridad, performance y permite compartir código entre frontend y backend.

## 🏗️ Arquitectura

### Stack Tecnológico

#### Backend

- **Lenguaje**: Rust 1.75+
- **Framework**: Axum
- **ORM**: SeaORM
- **Base de datos**: SQLite / PostgreSQL
- **Async Runtime**: Tokio
- **Serialización**: serde
- **HTTP Client**: reqwest

#### Frontend

- **Framework**: Leptos (SSR + WASM)
- **CSS**: Tailwind CSS
- **State**: Leptos Signals (reactive)
- **Routing**: Leptos Router
- **Icons**: Icondata

#### Shared

- **Types**: Compartidos entre frontend/backend
- **Validation**: validator crate
- **Utils**: Reutilización de lógica

#### DevOps

- **Container**: Multi-stage Docker
- **Build**: Cargo + Trunk
- **Size**: ~30-50MB imagen final

## 📐 Estructura del Proyecto

```plaintext
keepercheky/
├── Cargo.toml              # Workspace root
├── Cargo.lock
│
├── crates/
│   ├── server/             # Backend (Axum)
│   │   ├── Cargo.toml
│   │   └── src/
│   │       ├── main.rs
│   │       ├── routes/
│   │       │   ├── mod.rs
│   │       │   ├── media.rs
│   │       │   ├── schedules.rs
│   │       │   └── settings.rs
│   │       ├── services/
│   │       │   ├── mod.rs
│   │       │   ├── cleanup.rs
│   │       │   ├── scheduler.rs
│   │       │   └── clients/
│   │       │       ├── mod.rs
│   │       │       ├── radarr.rs
│   │       │       ├── sonarr.rs
│   │       │       └── jellyfin.rs
│   │       ├── models/
│   │       │   ├── mod.rs
│   │       │   ├── media.rs
│   │       │   └── schedule.rs
│   │       └── db/
│   │           ├── mod.rs
│   │           └── migrations/
│   │
│   ├── frontend/           # Frontend (Leptos)
│   │   ├── Cargo.toml
│   │   ├── Trunk.toml
│   │   └── src/
│   │       ├── main.rs
│   │       ├── app.rs
│   │       ├── pages/
│   │       │   ├── mod.rs
│   │       │   ├── dashboard.rs
│   │       │   ├── media.rs
│   │       │   ├── schedules.rs
│   │       │   ├── settings.rs
│   │       │   └── logs.rs
│   │       ├── components/
│   │       │   ├── mod.rs
│   │       │   ├── navbar.rs
│   │       │   ├── sidebar.rs
│   │       │   ├── card.rs
│   │       │   └── table.rs
│   │       └── api/
│   │           ├── mod.rs
│   │           └── client.rs
│   │
│   └── shared/             # Shared types & logic
│       ├── Cargo.toml
│       └── src/
│           ├── lib.rs
│           ├── types/
│           │   ├── mod.rs
│           │   ├── media.rs
│           │   ├── schedule.rs
│           │   └── config.rs
│           ├── validation/
│           │   └── mod.rs
│           └── utils/
│               └── mod.rs
│
├── migrations/             # SeaORM migrations
├── static/
│   └── styles.css
├── Dockerfile
├── docker-compose.yml
└── README.md
```

## 🎨 Interfaz de Usuario (Leptos)

### Arquitectura Frontend

**Filosofía**: Full-stack Rust con SSR y hydration

- Server-Side Rendering (SSR) para SEO y performance inicial
- WebAssembly para interactividad en el cliente
- Reactive signals para estado
- Type-safe API calls

### Ejemplos de Componentes

#### 1. Dashboard Page

```rust
// crates/frontend/src/pages/dashboard.rs
use leptos::*;
use shared::types::{Stats, ServiceStatus};
use crate::api::client::ApiClient;

#[component]
pub fn Dashboard() -> impl IntoView {
    let (stats, set_stats) = create_signal(None::<Stats>);
    
    // Fetch stats on mount
    create_effect(move |_| {
        spawn_local(async move {
            if let Ok(data) = ApiClient::get_stats().await {
                set_stats(Some(data));
            }
        });
    });
    
    // Auto-refresh every 30 seconds
    use_interval(30_000, move || {
        spawn_local(async move {
            if let Ok(data) = ApiClient::get_stats().await {
                set_stats(Some(data));
            }
        });
    });
    
    view! {
        <div class="container mx-auto p-6">
            <h1 class="text-3xl font-bold mb-6">"Dashboard"</h1>
            
            <Show
                when=move || stats.get().is_some()
                fallback=|| view! { <LoadingSpinner /> }
            >
                {move || {
                    let stats = stats.get().unwrap();
                    view! {
                        <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
                            <StatsCard
                                title="Espacio en Disco"
                                value=stats.disk_info
                                icon="💾"
                            />
                            
                            <StatsCard
                                title="Media por Eliminar"
                                value=format!("{} items", stats.media_to_delete.count)
                                icon="🗑️"
                            />
                            
                            <StatsCard
                                title="Última Limpieza"
                                value=stats.last_cleanup.time
                                icon="🧹"
                                status=stats.last_cleanup.status
                            />
                        </div>
                        
                        <div class="mt-6">
                            <h2 class="text-xl font-bold mb-4">"Estado de Servicios"</h2>
                            <ServicesStatus services=stats.services />
                        </div>
                        
                        <div class="mt-6">
                            <h2 class="text-xl font-bold mb-4">"Actividad Reciente"</h2>
                            <RecentActivity activities=stats.recent_activity />
                        </div>
                    }
                }}
            </Show>
        </div>
    }
}

#[component]
fn StatsCard(
    title: &'static str,
    value: String,
    icon: &'static str,
    #[prop(optional)] status: Option<String>,
) -> impl IntoView {
    view! {
        <div class="card bg-white rounded-lg shadow p-6">
            <div class="flex items-center justify-between mb-4">
                <h3 class="text-lg font-semibold">{title}</h3>
                <span class="text-2xl">{icon}</span>
            </div>
            <div class="text-2xl font-bold">{value}</div>
            {status.map(|s| view! {
                <span class="badge mt-2" class:badge-success=s == "success" class:badge-error=s == "error">
                    {s}
                </span>
            })}
        </div>
    }
}

#[component]
fn ServicesStatus(services: Vec<ServiceStatus>) -> impl IntoView {
    view! {
        <div class="flex gap-2">
            <For
                each=move || services.clone()
                key=|service| service.name.clone()
                children=|service| {
                    let status_class = match service.status.as_str() {
                        "online" => "badge-success",
                        "degraded" => "badge-warning",
                        _ => "badge-error",
                    };
                    
                    view! {
                        <span class=format!("badge {}", status_class)>
                            {service.name}
                        </span>
                    }
                }
            />
        </div>
    }
}

#[component]
fn LoadingSpinner() -> impl IntoView {
    view! {
        <div class="flex justify-center items-center h-64">
            <div class="spinner"></div>
        </div>
    }
}
```

#### 2. Media Management

```rust
// crates/frontend/src/pages/media.rs
use leptos::*;
use leptos_router::*;
use shared::types::{Media, MediaFilter};
use crate::api::client::ApiClient;

#[component]
pub fn MediaPage() -> impl IntoView {
    let (media_list, set_media_list) = create_signal(Vec::<Media>::new());
    let (loading, set_loading) = create_signal(false);
    let (view_mode, set_view_mode) = create_signal("grid");
    
    // Filters
    let (search, set_search) = create_signal(String::new());
    let (media_type, set_media_type) = create_signal("all".to_string());
    let (status_filter, set_status_filter) = create_signal("all".to_string());
    let (page, set_page) = create_signal(1);
    let (total_pages, set_total_pages) = create_signal(1);
    
    // Fetch media
    let fetch_media = create_action(|_: &()| async move {
        set_loading(true);
        
        let filter = MediaFilter {
            search: search.get(),
            media_type: media_type.get(),
            status: status_filter.get(),
            page: page.get(),
        };
        
        match ApiClient::get_media(filter).await {
            Ok(response) => {
                set_media_list(response.items);
                set_total_pages(response.total_pages);
            }
            Err(e) => {
                log::error!("Error fetching media: {:?}", e);
            }
        }
        
        set_loading(false);
    });
    
    // Fetch on mount
    create_effect(move |_| {
        fetch_media.dispatch(());
    });
    
    // Debounced search
    create_effect(move |_| {
        let _ = search.get();
        set_timeout(|| fetch_media.dispatch(()), std::time::Duration::from_millis(500));
    });
    
    view! {
        <div class="container mx-auto p-6">
            <h1 class="text-3xl font-bold mb-6">"Media Management"</h1>
            
            // Filters
            <div class="filters mb-6 flex gap-4">
                <input
                    type="search"
                    placeholder="Buscar media..."
                    class="input flex-1"
                    on:input=move |ev| set_search(event_target_value(&ev))
                    prop:value=search
                />
                
                <select
                    class="select"
                    on:change=move |ev| {
                        set_media_type(event_target_value(&ev));
                        fetch_media.dispatch(());
                    }
                >
                    <option value="all">"Todos"</option>
                    <option value="movie">"Películas"</option>
                    <option value="series">"Series"</option>
                </select>
                
                <select
                    class="select"
                    on:change=move |ev| {
                        set_status_filter(event_target_value(&ev));
                        fetch_media.dispatch(());
                    }
                >
                    <option value="all">"Todos los estados"</option>
                    <option value="leaving-soon">"Leaving Soon"</option>
                    <option value="excluded">"Excluidos"</option>
                </select>
            </div>
            
            // View mode toggle
            <div class="flex justify-between mb-4">
                <div class="btn-group">
                    <button
                        class="btn"
                        class:active=move || view_mode.get() == "grid"
                        on:click=move |_| set_view_mode("grid")
                    >
                        "Grid"
                    </button>
                    <button
                        class="btn"
                        class:active=move || view_mode.get() == "list"
                        on:click=move |_| set_view_mode("list")
                    >
                        "List"
                    </button>
                </div>
                
                <p class="text-gray-600">
                    {move || format!("{} items", media_list.get().len())}
                </p>
            </div>
            
            // Loading state
            <Show
                when=move || !loading.get()
                fallback=|| view! { <LoadingSpinner /> }
            >
                // Grid view
                <Show when=move || view_mode.get() == "grid">
                    <div class="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
                        <For
                            each=move || media_list.get()
                            key=|media| media.id
                            children=|media| view! {
                                <MediaCard
                                    media=media
                                    on_exclude=move || fetch_media.dispatch(())
                                    on_delete=move || fetch_media.dispatch(())
                                />
                            }
                        />
                    </div>
                </Show>
                
                // List view
                <Show when=move || view_mode.get() == "list">
                    <MediaTable
                        media=media_list.get()
                        on_exclude=move || fetch_media.dispatch(())
                        on_delete=move || fetch_media.dispatch(())
                    />
                </Show>
            </Show>
            
            // Pagination
            <Pagination
                current_page=page
                total_pages=total_pages
                on_page_change=move |new_page| {
                    set_page(new_page);
                    fetch_media.dispatch(());
                }
            />
        </div>
    }
}

#[component]
fn MediaCard<F1, F2>(
    media: Media,
    on_exclude: F1,
    on_delete: F2,
) -> impl IntoView 
where
    F1: Fn() + 'static,
    F2: Fn() + 'static,
{
    let exclude_action = create_action(move |id: &i32| {
        let id = *id;
        async move {
            ApiClient::exclude_media(id).await.ok();
        }
    });
    
    let delete_action = create_action(move |id: &i32| {
        let id = *id;
        async move {
            if window().confirm_with_message("¿Eliminar este media?").unwrap_or(false) {
                ApiClient::delete_media(id).await.ok();
            }
        }
    });
    
    view! {
        <div class="card cursor-pointer">
            <img
                src=media.poster_url.clone()
                alt=media.title.clone()
                class="w-full h-auto rounded"
            />
            <h4 class="mt-2 font-medium truncate">{media.title.clone()}</h4>
            <p class="text-sm text-gray-600">
                {media.media_type.clone()} " · " {media.size.clone()}
            </p>
            
            <div class="mt-2 flex gap-2">
                <button
                    class="btn btn-sm"
                    on:click=move |_| {
                        exclude_action.dispatch(media.id);
                        on_exclude();
                    }
                >
                    "Excluir"
                </button>
                
                <button
                    class="btn btn-sm btn-danger"
                    on:click=move |_| {
                        delete_action.dispatch(media.id);
                        on_delete();
                    }
                >
                    "Eliminar"
                </button>
            </div>
        </div>
    }
}

#[component]
fn Pagination<F>(
    current_page: ReadSignal<i32>,
    total_pages: ReadSignal<i32>,
    on_page_change: F,
) -> impl IntoView
where
    F: Fn(i32) + 'static,
{
    view! {
        <div class="flex justify-center mt-6 gap-2">
            <button
                class="btn"
                disabled=move || current_page.get() == 1
                on:click=move |_| on_page_change(current_page.get() - 1)
            >
                "Anterior"
            </button>
            
            <span class="py-2 px-4">
                "Página " {move || current_page.get()} " de " {move || total_pages.get()}
            </span>
            
            <button
                class="btn"
                disabled=move || current_page.get() >= total_pages.get()
                on:click=move |_| on_page_change(current_page.get() + 1)
            >
                "Siguiente"
            </button>
        </div>
    }
}
```

#### 3. API Client (Shared)

```rust
// crates/frontend/src/api/client.rs
use shared::types::*;
use serde::{Deserialize, Serialize};
use wasm_bindgen::JsValue;

pub struct ApiClient;

impl ApiClient {
    const BASE_URL: &'static str = "/api";
    
    pub async fn get_stats() -> Result<Stats, JsValue> {
        let url = format!("{}/stats", Self::BASE_URL);
        Self::get(&url).await
    }
    
    pub async fn get_media(filter: MediaFilter) -> Result<MediaResponse, JsValue> {
        let url = format!(
            "{}/media?search={}&type={}&status={}&page={}",
            Self::BASE_URL,
            filter.search,
            filter.media_type,
            filter.status,
            filter.page
        );
        Self::get(&url).await
    }
    
    pub async fn exclude_media(id: i32) -> Result<(), JsValue> {
        let url = format!("{}/media/{}/exclude", Self::BASE_URL, id);
        Self::post(&url, &()).await
    }
    
    pub async fn delete_media(id: i32) -> Result<(), JsValue> {
        let url = format!("{}/media/{}", Self::BASE_URL, id);
        Self::delete(&url).await
    }
    
    async fn get<T: for<'de> Deserialize<'de>>(url: &str) -> Result<T, JsValue> {
        let response = gloo_net::http::Request::get(url)
            .send()
            .await
            .map_err(|e| JsValue::from_str(&e.to_string()))?;
        
        response
            .json()
            .await
            .map_err(|e| JsValue::from_str(&e.to_string()))
    }
    
    async fn post<T: Serialize, R: for<'de> Deserialize<'de>>(
        url: &str,
        body: &T,
    ) -> Result<R, JsValue> {
        let response = gloo_net::http::Request::post(url)
            .json(body)
            .map_err(|e| JsValue::from_str(&e.to_string()))?
            .send()
            .await
            .map_err(|e| JsValue::from_str(&e.to_string()))?;
        
        response
            .json()
            .await
            .map_err(|e| JsValue::from_str(&e.to_string()))
    }
    
    async fn delete<R: for<'de> Deserialize<'de>>(url: &str) -> Result<R, JsValue> {
        let response = gloo_net::http::Request::delete(url)
            .send()
            .await
            .map_err(|e| JsValue::from_str(&e.to_string()))?;
        
        response
            .json()
            .await
            .map_err(|e| JsValue::from_str(&e.to_string()))
    }
}
```

## ⚙️ Backend (Axum + Rust)

### Main Application

```rust
// crates/server/src/main.rs
use axum::{
    Router,
    routing::{get, post, put, delete},
    extract::Extension,
};
use sea_orm::Database;
use std::sync::Arc;
use tower_http::services::ServeDir;

mod routes;
mod services;
mod models;
mod db;

use services::{CleanupService, SchedulerService};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Initialize logging
    tracing_subscriber::fmt::init();
    
    // Connect to database
    let db = Database::connect("sqlite://data/keepercheky.db").await?;
    
    // Run migrations
    db::run_migrations(&db).await?;
    
    // Initialize services
    let cleanup_service = Arc::new(CleanupService::new(db.clone()));
    let scheduler_service = Arc::new(SchedulerService::new(cleanup_service.clone()));
    
    // Start scheduler
    scheduler_service.start().await;
    
    // Build router
    let app = Router::new()
        // Serve static files
        .nest_service("/static", ServeDir::new("static"))
        
        // API routes
        .route("/api/stats", get(routes::dashboard::get_stats))
        
        .route("/api/media", get(routes::media::get_media))
        .route("/api/media/:id/exclude", post(routes::media::exclude_media))
        .route("/api/media/:id", delete(routes::media::delete_media))
        
        .route("/api/schedules", get(routes::schedules::get_schedules))
        .route("/api/schedules", post(routes::schedules::create_schedule))
        .route("/api/schedules/:id", put(routes::schedules::update_schedule))
        .route("/api/schedules/:id", delete(routes::schedules::delete_schedule))
        .route("/api/schedules/:id/run", post(routes::schedules::run_schedule))
        .route("/api/schedules/:id/toggle", put(routes::schedules::toggle_schedule))
        
        .route("/api/settings", get(routes::settings::get_settings))
        .route("/api/settings", put(routes::settings::update_settings))
        
        // Serve frontend
        .fallback_service(ServeDir::new("frontend/dist"))
        
        // Add services to extensions
        .layer(Extension(db))
        .layer(Extension(cleanup_service))
        .layer(Extension(scheduler_service));
    
    // Start server
    let addr = "0.0.0.0:8000".parse()?;
    tracing::info!("Listening on {}", addr);
    
    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await?;
    
    Ok(())
}
```

### Route Handler Example

```rust
// crates/server/src/routes/media.rs
use axum::{
    extract::{Extension, Path, Query},
    http::StatusCode,
    Json,
};
use sea_orm::DatabaseConnection;
use serde::{Deserialize, Serialize};
use shared::types::{Media, MediaFilter, MediaResponse};

use crate::services::CleanupService;
use crate::models::media::Entity as MediaEntity;

pub async fn get_media(
    Query(filter): Query<MediaFilter>,
    Extension(db): Extension<DatabaseConnection>,
) -> Result<Json<MediaResponse>, StatusCode> {
    // Query database
    let (items, total) = MediaEntity::find_filtered(
        &db,
        &filter.search,
        &filter.media_type,
        &filter.status,
        filter.page,
        20,
    )
    .await
    .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;
    
    Ok(Json(MediaResponse {
        items,
        page: filter.page,
        total_pages: (total + 19) / 20,
    }))
}

pub async fn exclude_media(
    Path(id): Path<i32>,
    Extension(db): Extension<DatabaseConnection>,
) -> Result<StatusCode, StatusCode> {
    MediaEntity::exclude(&db, id)
        .await
        .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;
    
    Ok(StatusCode::OK)
}

pub async fn delete_media(
    Path(id): Path<i32>,
    Extension(db): Extension<DatabaseConnection>,
    Extension(cleanup): Extension<Arc<CleanupService>>,
) -> Result<StatusCode, StatusCode> {
    let media = MediaEntity::find_by_id(&db, id)
        .await
        .map_err(|_| StatusCode::NOT_FOUND)?;
    
    cleanup
        .delete_media(&media)
        .await
        .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;
    
    Ok(StatusCode::OK)
}
```

### Cleanup Service

```rust
// crates/server/src/services/cleanup.rs
use sea_orm::DatabaseConnection;
use shared::types::Media;
use std::sync::Arc;

use super::clients::{RadarrClient, SonarrClient, JellyfinClient};

pub struct CleanupService {
    db: DatabaseConnection,
    radarr: Arc<RadarrClient>,
    sonarr: Arc<SonarrClient>,
    jellyfin: Option<Arc<JellyfinClient>>,
}

impl CleanupService {
    pub fn new(db: DatabaseConnection) -> Self {
        // Initialize clients from config
        let radarr = Arc::new(RadarrClient::new());
        let sonarr = Arc::new(SonarrClient::new());
        let jellyfin = Some(Arc::new(JellyfinClient::new()));
        
        Self {
            db,
            radarr,
            sonarr,
            jellyfin,
        }
    }
    
    pub async fn get_media_to_delete(&self) -> Result<Vec<Media>, anyhow::Error> {
        // Implementation similar to Go/Python versions
        todo!()
    }
    
    pub async fn delete_media(&self, media: &Media) -> Result<(), anyhow::Error> {
        // Delete from Jellyfin
        if let Some(jellyfin) = &self.jellyfin {
            if let Some(jellyfin_id) = media.jellyfin_id {
                jellyfin.delete_item(jellyfin_id).await?;
            }
        }
        
        // Delete from *arr
        match media.media_type.as_str() {
            "movie" => {
                if let Some(radarr_id) = media.radarr_id {
                    self.radarr.delete_item(radarr_id).await?;
                }
            }
            "series" => {
                if let Some(sonarr_id) = media.sonarr_id {
                    self.sonarr.delete_item(sonarr_id).await?;
                }
            }
            _ => {}
        }
        
        // Update database
        MediaEntity::delete(&self.db, media.id).await?;
        
        Ok(())
    }
}
```

## 🐳 Docker

### Multi-stage Dockerfile

```dockerfile
# Build frontend (WASM)
FROM rust:1.75-alpine AS frontend-builder

WORKDIR /app

RUN apk add --no-cache musl-dev

# Install trunk
RUN cargo install trunk wasm-bindgen-cli

# Copy workspace
COPY Cargo.toml Cargo.lock ./
COPY crates/ ./crates/

# Build frontend
WORKDIR /app/crates/frontend
RUN trunk build --release

# Build backend
FROM rust:1.75-alpine AS backend-builder

WORKDIR /app

RUN apk add --no-cache musl-dev

COPY Cargo.toml Cargo.lock ./
COPY crates/ ./crates/

# Build server
RUN cargo build --release --bin keepercheky-server

# Final image
FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy binary
COPY --from=backend-builder /app/target/release/keepercheky-server /app/keepercheky

# Copy frontend dist
COPY --from=frontend-builder /app/crates/frontend/dist /app/frontend/dist

EXPOSE 8000

CMD ["/app/keepercheky"]
```

## 🎯 Ventajas

### ✅ Pros

1. **Memory safety**: Zero null pointers, data races
2. **Performance**: Comparable a C/C++
3. **Type safety extrema**: Compile-time guarantees
4. **Code sharing**: Frontend/backend share types
5. **WebAssembly**: Performance en el browser
6. **Concurrency**: Tokio async runtime
7. **Small binary**: 30-50MB
8. **Low memory**: 30-80MB RAM

### ⚠️ Contras

1. **Curva de aprendizaje**: Rust es complejo
2. **Compile times**: Más lentos que Go
3. **Ecosistema web**: Menos maduro que JS/Python
4. **Leptos**: Framework relativamente nuevo
5. **WASM size**: Puede ser grande (MB)

## 📊 Recursos

### Desarrollo

- **Tiempo**: 5-7 semanas
- **Dificultad**: Alta

### Runtime

- **RAM**: 30-80MB
- **CPU**: 0.5-1 core
- **Disco**: 30-50MB

## 📝 Conclusión

**Ideal para**: Máxima seguridad, performance, proyectos a largo plazo con requisitos estrictos.

**No recomendado si**: No tienes experiencia con Rust o necesitas desarrollo rápido.
