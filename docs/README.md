# MediaCheky - Documentación

> Panel de control centralizado para servicios multimedia *arr

## 📖 Documentos Principales

- **[Plan de Desarrollo](PROJECT_PLAN.md)** ⭐ **EMPIEZA AQUÍ**
  - Visión completa del proyecto
  - Arquitectura técnica
  - Roadmap de desarrollo
  - UI/UX diseñado
  - Checklist de implementación

## 🚀 Guías Rápidas

- [Instalación Rápida](../QUICKSTART.md)
- [Guía de Desarrollo](../DEVELOPMENT.md)
- [Guía para Agentes IA](../AGENTS.md)

## 🏗️ Stack Tecnológico

- **Backend**: Go 1.25 + Fiber v2
- **Frontend**: Alpine.js 3.x + Tailwind CSS
- **Base de datos**: GORM v2 + SQLite/PostgreSQL
- **Docker**: Socket API + Compose Templates
- **Target**: ~25MB imagen, ~30-50MB RAM

## 🎯 Arquitectura

**Enfoque Híbrido: Docker Socket + Docker Compose Templates**

1. Usuario activa servicio en UI
2. MediaCheky genera `docker-compose.yml` desde template
3. MediaCheky ejecuta `docker compose up -d` vía Socket
4. Dashboard muestra estado en tiempo real

## �� Servicios Soportados

- Jellyfin (media server)
- Radarr (películas)
- Sonarr (series)
- Prowlarr (indexadores)
- qBittorrent (torrents)
- Jellyseerr (peticiones)
- Bazarr (subtítulos)
- Jellystat (estadísticas)

## 🗂️ Documentos Técnicos

### En Desarrollo

- [ ] API Documentation
- [ ] Database Schema
- [ ] Template System Guide
- [ ] Security Best Practices
- [ ] Testing Strategy

---

**Última actualización**: Noviembre 3, 2025  
**Estado**: Redefinición completa del proyecto
