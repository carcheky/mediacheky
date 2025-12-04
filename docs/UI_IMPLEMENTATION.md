# UI Implementation Documentation

## Overview

This document describes the UI implementation using Alpine.js and Tailwind CSS for MediaCheky.

## Implementation Date

November 9, 2025

## Implemented Features

### 1. Page Structure

#### Base Layout (`web/templates/layouts/main.html`)

- **Enhanced Navigation**: Top navigation bar with links to all main pages
  - **Services Dropdown**: Hover-activated menu showing enabled/disabled services with quick access to configuration pages
  - Dashboard, Settings, and Logs links

- **Mobile Responsive**: Collapsible mobile menu for small screens

- **Footer**: Version information and GitHub link

- **Theme**: Dark theme with custom color scheme

- **Technologies**: Tailwind CSS (CDN) + Alpine.js 3.x

#### Dashboard (`/`)

- Already implemented with comprehensive widgets

- Displays stats from various *arr services

- Real-time data with auto-refresh

#### Settings (`/settings`)

- Already implemented

- Global Variables management only (PUID, PGID, TZ, paths, network)

- Connection testing for global configuration

- Single section: Global Variables (services management moved to `/services`)

#### Global Variables (Settings) - **UPDATED**

NOTE: Global Variables is a section within Settings (only section). There is no separate `/global` route.

- User & Group configuration (PUID, PGID)

- System settings (Timezone, Language)

- Base paths (Media, Downloads, Config)

- Network configuration

- **Floating toast notifications** for all feedback

#### Service Config (`/services/:name`)

- Service-specific configuration page

- Basic settings (Port, Image, Restart policy)

- Path configuration (Config, Media, Downloads)

- Environment variables preview from global config

- Advanced settings (VPN routing, Custom network)

- Real-time field validation

- Connection testing

##### Sonarr Configuration

- Default port: `8989`
- Paths: `config`, `series` (TV shows), `downloads`
- Image options: `linuxserver/sonarr:latest|develop|nightly`
- Optional port exposure via checkbox per `UI_STANDARDS` rules
- Follows same pattern as Radarr (PR #37)

### 2. Components

#### Service Card (`web/templates/components/service_card.html`)

- Reusable service display component

- Status badges

- Quick actions (Start, Stop, Restart, Configure)

- Icon and description display

#### Form Field (`web/templates/components/form_field.html`)

- Generic form field component

- Supports: text, number, password, url, select, textarea, checkbox

- Validation error display

- Help text support

- Environment variable source indication

- Disabled state styling

### 3. Styling

#### Custom CSS (`web/static/css/custom.css`)

- Custom scrollbar for dark theme

- Loading animations

- Card hover effects

- Status badge animations

- Form focus states

- Toast notifications

- Service icon colors

- Skeleton loading animations

- Print-friendly styles

- Responsive utilities

### 4. Backend Routes

#### Page Routes

```go
app.Get("/", h.Dashboard.Index)
app.Get("/settings", h.Settings.Index)          // Global Variables only
app.Get("/services", h.Service.ListPage)        // Services overview + enable/disable + Configure
app.Get("/services/:name", h.Service.ConfigPage)
app.Get("/logs", h.Logs.Index)
```

#### API Routes (existing)

```go
api.Get("/config/global", h.Config.GetGlobalConfig)
api.Put("/config/global", h.Config.UpdateGlobalConfig)
api.Get("/services", h.Service.List)
api.Get("/services/:name", h.Service.GetService)
api.Put("/services/:name/config", h.Service.UpdateServiceConfig)
api.Post("/services/:name/enable", h.Service.Enable)
api.Post("/services/:name/disable", h.Service.Disable)
api.Post("/config/test/:service", h.Settings.TestConnection)
```

### 5. Alpine.js Components

#### Global Config (`globalConfig()`)

- Load and save global configuration

- Track affected services

- Real-time form validation

- Success/error messaging

#### Service Config (`serviceConfig(serviceName)`)

- Load service-specific configuration

- Load global config for environment variables

- Field validation (port, paths)

- Connection testing

- Save configuration with validation

## File Structure

**Note**: Current workspace shows templates in `web/static/templates/`. The structure below represents the planned organization. Future work should either move files to `web/templates/` or update this documentation to reflect `web/static/templates/` as canonical.

```text
web/
├── static/
│   ├── css/
│   │   └── custom.css                 # Custom styles
│   ├── js/
│   │   └── file-health-components.js  # Existing JS components
│   └── templates/                     # Current actual location in workspace
│       └── ...                         # Template files (to be organized)
└── templates/                         # Planned/standard location
    ├── components/                     # Reusable UI components
    │   ├── form_field.html            # Generic form field
    │   └── service_card.html          # Service display card
    ├── layouts/
    │   └── main.html                  # Base layout template
    └── pages/
        ├── dashboard.html             # Main dashboard
        ├── settings.html              # Global Variables (only section)
        ├── services.html              # Services management
        └── logs.html                  # Logs viewer
```

## Design Principles

### Color Scheme

- **Background**: `#0f172a` (dark-bg)

- **Surface**: `#1e293b` (dark-surface)

- **Border**: `#334155` (dark-border)

- **Text**: `#e2e8f0` (dark-text)

- **Muted**: `#94a3b8` (dark-muted)

- **Accent**: Blue/Cyan for positive actions

- **Destructive**: Red for dangerous actions

### Responsive Design

- Mobile-first approach

- Collapsible navigation on small screens

- Grid layouts adapt to screen size

- Touch-friendly button sizes

### Accessibility

- Semantic HTML

- ARIA labels where appropriate

- Keyboard navigation support

- Focus indicators

- Sufficient color contrast

## Usage Examples

### Creating a New Service Configuration

1. Navigate to `/services`

2. Enable the service (toggle)

3. Click "Configure" or navigate to `/services/servicename`

4. Fill in required fields (marked with *)

5. Test connection

6. Save configuration

### Modifying Global Variables

1. Navigate to `/settings`

2. Global Variables is the only section

3. Update desired values

4. Save changes (toast notification will appear)

5. Restart affected services if needed

## Testing

### Manual Testing Checklist

- [x] Build successful (`go build`)

- [x] All handler tests pass

- [ ] Dashboard page renders

- [ ] Settings page renders

- [ ] Global page renders and functions

- [ ] Service config page renders and functions

- [ ] Mobile responsive layout works

- [ ] Alpine.js interactivity works

- [ ] Form validation works

- [ ] API endpoints respond correctly

### Browser Testing

- [ ] Chrome/Chromium

- [ ] Firefox

- [ ] Safari

- [ ] Mobile browsers

## Future Enhancements

### Potential Improvements

1. **Additional Components**

   - Notification toast system
   - Modal dialogs
   - Confirmation dialogs
   - Progress indicators

2. **Enhanced Validation**

   - Real-time path existence checking
   - Port availability checking
   - Network name validation

3. **User Experience**

   - Auto-save drafts
   - Undo/redo functionality
   - Keyboard shortcuts
   - Dark/light theme toggle

4. **Advanced Features**

   - Bulk service configuration
   - Configuration import/export
   - Configuration templates
   - Service dependency visualization

## Notes

- All pages use server-side rendering with Go templates

- Alpine.js provides client-side reactivity

- Tailwind CSS is loaded via CDN for simplicity

- Custom CSS extends Tailwind with project-specific styles

- All API calls use fetch with proper error handling

- Forms include real-time validation feedback

## Related Issues

- Issue: carcheky/mediacheky#[number] - UI Frontend con Alpine.js y Tailwind

- Parent: carcheky/mediacheky#9 - Fase 1: MVP
