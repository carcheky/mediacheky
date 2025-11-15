# GitHub Issue: Implement Radarr and Sonarr Configuration UI in MediaCheky

## Issue Title
Implement comprehensive Radarr and Sonarr configuration interface in MediaCheky

## Issue Type
✨ Feature Request / 📋 Documentation

## Priority
🔥 High - Core functionality for *arr service management

## Labels
- `enhancement`
- `radarr`
- `sonarr`
- `ui`
- `configuration`
- `documentation`

## Assignee
@github-copilot

---

## 📋 Description

This issue tracks the implementation of a comprehensive configuration interface for Radarr (movie management) and Sonarr (TV series management) within MediaCheky. 

A detailed analysis of all configuration features available in both services has been completed and documented in `docs/RADARR_SONARR_FEATURES.md`.

## 🎯 Objectives

Create user-friendly configuration interfaces in MediaCheky that allow users to:

1. **Enable/disable and configure Radarr and Sonarr services**
2. **Set up essential service parameters** (ports, paths, authentication)
3. **Manage quality profiles and monitoring options**
4. **Configure integrations** (download clients, indexers, import lists)
5. **Customize naming schemes and file management**
6. **Set up notifications and metadata providers**

## 📚 Reference Documentation

Complete feature documentation is available in:
- **File**: `docs/RADARR_SONARR_FEATURES.md`
- **Location**: Repository root `/docs/` directory

This documentation includes:
- ✅ Complete listing of all Radarr configuration sections (12 sections)
- ✅ Complete listing of all Sonarr configuration sections (13 sections)
- ✅ Detailed breakdown of every configuration option
- ✅ Key differences between Radarr and Sonarr
- ✅ Implementation recommendations organized by priority
- ✅ UI organization suggestions

## 🏗️ Implementation Plan

### Phase 1: MVP - Essential Features (Priority: Critical)

**Goal**: Get Radarr/Sonarr running with minimal configuration

#### Tasks:
- [ ] Create basic service configuration UI component
- [ ] Implement enable/disable toggle for each service
- [ ] Add port configuration (external port mapping)
- [ ] Add path configuration (config folder, media library)
- [ ] Implement authentication settings (API key management)
- [ ] Add network configuration (Docker network selection)
- [ ] Create service status indicator
- [ ] Add "Test Connection" functionality
- [ ] Generate docker-compose.yml from configuration

**Deliverables**:
- Working Radarr configuration page
- Working Sonarr configuration page
- Services can be started/stopped from MediaCheky
- Basic validation of configuration inputs

**Estimated Effort**: 3-5 days

---

### Phase 2: Quality & Monitoring (Priority: High)

**Goal**: Configure how services monitor and download media

#### Tasks:
- [ ] Implement quality profile selection UI
- [ ] Add monitoring options configuration
  - [ ] Radarr: Movie availability settings
  - [ ] Sonarr: Episode monitoring settings (all/future/missing/etc.)
- [ ] Create download client selector (from configured clients)
- [ ] Add root folder management UI
- [ ] Implement default settings per service

**Deliverables**:
- Quality profile dropdown with descriptions
- Monitoring configuration interface
- Root folder management (add/edit/delete)
- Integration with existing download client configuration

**Estimated Effort**: 3-4 days

---

### Phase 3: Advanced Configuration (Priority: Medium)

**Goal**: Provide access to advanced features

#### Tasks:
- [ ] Create indexer management interface
  - [ ] Add indexer form (Newznab, Torznab)
  - [ ] Test indexer connectivity
  - [ ] Enable/disable indexers
- [ ] Implement import list configuration
  - [ ] Support major list types (Trakt, IMDb, TMDb, Plex)
  - [ ] Configure list sync intervals
- [ ] Add naming template editor
  - [ ] Provide token reference guide
  - [ ] Live preview of naming format
- [ ] Create custom formats UI (basic)
  - [ ] Add/edit/delete custom formats
  - [ ] Assign scores to formats

**Deliverables**:
- Indexer management page
- Import list configuration interface
- Naming template editor with preview
- Basic custom format management

**Estimated Effort**: 5-7 days

---

### Phase 4: Full Integration (Priority: Low)

**Goal**: Deep integration with MediaCheky dashboard

#### Tasks:
- [ ] Add real-time statistics display
  - [ ] Show active downloads
  - [ ] Display queue status
  - [ ] Show disk space usage
- [ ] Implement health monitoring
  - [ ] Display service health checks
  - [ ] Show warnings/errors
  - [ ] Quick fix suggestions
- [ ] Create quick action shortcuts
  - [ ] Search for movies/series from MediaCheky
  - [ ] Add movie/series from MediaCheky
  - [ ] View recent additions
- [ ] Build unified dashboard
  - [ ] Combined view of all *arr services
  - [ ] Cross-service statistics
  - [ ] Unified notification center

**Deliverables**:
- Real-time service monitoring
- Health check integration
- Quick action interface
- Unified *arr dashboard

**Estimated Effort**: 7-10 days

---

## 🎨 UI/UX Design Recommendations

### Layout Structure

```
┌─────────────────────────────────────────────────┐
│ MediaCheky > Services > Radarr                  │
├─────────────────────────────────────────────────┤
│                                                 │
│  [Enable Service] ◯────────────●  Port: 7878   │
│                                                 │
│  ┌─ Basic Settings ────────────────────────┐  │
│  │ • Service Port: 7878                     │  │
│  │ • Config Path: /config/radarr            │  │
│  │ • Media Path: /media/movies              │  │
│  │ • API Key: ••••••••••••• [Generate]      │  │
│  │ • Network: mediacheky-net                │  │
│  └──────────────────────────────────────────┘  │
│                                                 │
│  ┌─ Quality & Profiles ──────────────────┐    │
│  │ • Default Quality Profile: HD-1080p     │    │
│  │ • Monitor: Movie Only                   │    │
│  │ • Minimum Availability: Released        │    │
│  │ • Root Folder: /media/movies           │    │
│  └────────────────────────────────────────┘    │
│                                                 │
│  ┌─ Integration ─────────────────────────┐    │
│  │ • Download Clients: [qBittorrent]       │    │
│  │ • Indexers: [2 configured]              │    │
│  │ • Import Lists: [1 configured]          │    │
│  └────────────────────────────────────────┘    │
│                                                 │
│  ┌─ Advanced ───────────────────────────┐      │
│  │ • Naming Format: {Movie Title} ({Y...   │    │
│  │ • Custom Formats: [3 configured]        │    │
│  │ • Permissions: PUID=1000 PGID=1000     │    │
│  └────────────────────────────────────────┘    │
│                                                 │
│  [Test Connection] [Save Configuration]        │
│                                                 │
└─────────────────────────────────────────────────┘
```

### Design Principles

1. **Progressive Disclosure**: Show basic settings first, advanced settings in collapsible sections
2. **Validation Feedback**: Real-time validation with clear error messages
3. **Smart Defaults**: Pre-populate sensible defaults for new users
4. **Test Before Save**: Allow users to test connections before saving
5. **Visual Feedback**: Show service status with color-coded indicators
6. **Help Text**: Include tooltips and help icons for complex settings
7. **Responsive Design**: Works on desktop and mobile devices

### Color Coding

- 🟢 **Green**: Service running and healthy
- 🟡 **Yellow**: Service running with warnings
- 🔴 **Red**: Service stopped or unhealthy
- ⚪ **Gray**: Service disabled

---

## 🔧 Technical Specifications

### Backend Requirements

1. **Database Schema Updates**
   ```go
   type ServiceConfig struct {
       ID              uint
       ServiceName     string // "radarr" or "sonarr"
       Enabled         bool
       Port            int
       ConfigPath      string
       MediaPath       string
       APIKey          string
       Network         string
       QualityProfile  string
       MonitoringType  string
       RootFolder      string
       // ... additional fields
   }
   ```

2. **API Endpoints**
   - `GET /api/services/radarr/config` - Get Radarr configuration
   - `PUT /api/services/radarr/config` - Update Radarr configuration
   - `POST /api/services/radarr/test` - Test Radarr connection
   - `GET /api/services/radarr/status` - Get Radarr service status
   - (Same endpoints for Sonarr)

3. **Docker Compose Generation**
   - Use Go templates to generate docker-compose.yml
   - Validate configuration before generation
   - Support hot-reload of configuration

4. **Service Integration**
   - Use Radarr/Sonarr APIs to validate settings
   - Fetch available quality profiles from service
   - Sync configuration between MediaCheky and service

### Frontend Requirements

1. **Components**
   - `ServiceConfigForm` - Main configuration form
   - `ServiceToggle` - Enable/disable switch
   - `PathSelector` - File path picker
   - `APIKeyGenerator` - Generate and manage API keys
   - `ServiceStatusBadge` - Visual status indicator
   - `ConfigSection` - Collapsible configuration section

2. **State Management**
   - Store service configuration in component state
   - Validate inputs before submission
   - Show loading states during API calls
   - Display success/error messages

3. **Validation Rules**
   - Port: 1024-65535, not already in use
   - Paths: Valid directory paths, writable
   - API Key: Valid format, test connectivity
   - Network: Must exist in Docker

---

## 📊 Success Criteria

### Phase 1 (MVP)
- [ ] Users can enable Radarr from MediaCheky UI
- [ ] Users can enable Sonarr from MediaCheky UI
- [ ] Services start successfully with configured settings
- [ ] Services are accessible on configured ports
- [ ] Configuration persists across MediaCheky restarts
- [ ] Basic validation prevents invalid configurations

### Phase 2
- [ ] Users can select quality profiles
- [ ] Users can configure monitoring options
- [ ] Users can manage root folders
- [ ] Default settings work out-of-the-box for most users

### Phase 3
- [ ] Users can add indexers through MediaCheky
- [ ] Users can configure import lists
- [ ] Users can customize naming templates
- [ ] Custom formats can be created and managed

### Phase 4
- [ ] Real-time service statistics displayed
- [ ] Health checks integrated into dashboard
- [ ] Quick actions work from MediaCheky
- [ ] Unified dashboard shows all *arr services

---

## 🧪 Testing Requirements

### Unit Tests
- [ ] Configuration validation logic
- [ ] Docker Compose template generation
- [ ] API key generation
- [ ] Path validation

### Integration Tests
- [ ] Service start/stop with configuration
- [ ] Configuration persistence in database
- [ ] API endpoint functionality
- [ ] Service communication (MediaCheky ↔ Radarr/Sonarr)

### E2E Tests
- [ ] Complete configuration flow
- [ ] Service enable → configure → start → verify
- [ ] Configuration update → service restart → verify
- [ ] Service disable → stop → cleanup

### Manual Testing Checklist
- [ ] Configure Radarr with various options
- [ ] Configure Sonarr with various options
- [ ] Verify services accessible after start
- [ ] Test configuration changes while service running
- [ ] Verify data persistence
- [ ] Test with different Docker networks
- [ ] Test with different port configurations

---

## 📖 Documentation Requirements

### User Documentation
- [ ] How to configure Radarr in MediaCheky
- [ ] How to configure Sonarr in MediaCheky
- [ ] Common configuration scenarios
- [ ] Troubleshooting guide
- [ ] FAQ

### Developer Documentation
- [ ] Configuration schema documentation
- [ ] API endpoint documentation
- [ ] Template system documentation
- [ ] Service integration patterns

### Screenshots/Videos
- [ ] Configuration UI screenshots
- [ ] Video walkthrough of setup process
- [ ] Before/after comparison

---

## 🔗 Related Resources

- **Radarr Official Documentation**: https://wiki.servarr.com/radarr
- **Sonarr Official Documentation**: https://wiki.servarr.com/sonarr
- **Radarr API Documentation**: https://radarr.video/docs/api/
- **Sonarr API Documentation**: https://sonarr.tv/docs/api/
- **MediaCheky Development Guide**: `DEVELOPMENT.md`
- **MediaCheky Architecture**: `docs/PROJECT_PLAN.md`

---

## 🐛 Known Issues & Considerations

### Security
- ⚠️ Store API keys securely (encrypted in database)
- ⚠️ Validate all user inputs to prevent injection
- ⚠️ Don't expose sensitive configuration in logs
- ⚠️ Use HTTPS for API communication when possible

### Performance
- ⚠️ Configuration changes may require service restart
- ⚠️ Large media libraries may slow initial sync
- ⚠️ API calls to services should be rate-limited

### Compatibility
- ⚠️ Support multiple Radarr/Sonarr versions
- ⚠️ Handle API changes between versions
- ⚠️ Test with different Docker configurations

### User Experience
- ⚠️ Provide clear feedback during long operations
- ⚠️ Save configuration drafts (don't lose user work)
- ⚠️ Confirm before potentially destructive actions
- ⚠️ Show what will change before applying

---

## 💬 Questions & Discussion

### Open Questions
1. Should we support multiple Radarr/Sonarr instances?
2. How do we handle configuration migration from existing installations?
3. Should we provide import/export of configurations?
4. Do we need a configuration wizard for first-time users?
5. How do we handle version compatibility issues?

### Discussion Topics
- Configuration UI organization
- Default settings strategy
- Migration path for existing users
- Multi-instance support priority

---

## 📝 Notes

- This feature is foundational for MediaCheky's core value proposition
- User feedback should be collected during Phase 1 implementation
- Consider creating a configuration template library for common setups
- May need to coordinate with Docker Compose v2 requirements
- Should align with overall MediaCheky design language

---

## ✅ Acceptance Criteria Summary

**This issue is complete when:**

1. ✅ Radarr and Sonarr can be fully configured through MediaCheky UI
2. ✅ Services start successfully with user-defined configurations
3. ✅ Configuration changes are persisted and survive restarts
4. ✅ All critical features from Phase 1 and 2 are implemented
5. ✅ Documentation is complete and tested
6. ✅ All tests pass (unit, integration, E2E)
7. ✅ User feedback has been incorporated

---

## 🏷️ Issue Metadata

- **Created**: 2025-11-15
- **Updated**: 2025-11-15
- **Repository**: carcheky/mediacheky
- **Related Issues**: None yet
- **Blocked By**: None
- **Blocks**: Future *arr service integrations (Prowlarr, Bazarr, etc.)

---

## 👥 Stakeholders

- **Product Owner**: @carcheky
- **Assigned Developer**: @github-copilot
- **Reviewers**: To be assigned
- **Testers**: To be assigned

---

**Priority**: 🔥 High  
**Effort**: 🏋️ Large (20-30 days across all phases)  
**Impact**: 🎯 High - Core feature for MediaCheky

---

*This issue is part of the MediaCheky service configuration initiative. For more context, see the full feature documentation in `docs/RADARR_SONARR_FEATURES.md`.*
