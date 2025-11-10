# MediaCheky Web Frontend

This directory contains all web frontend assets for MediaCheky.

## Structure

```text
web/
├── static/              # Static assets (CSS, JS, images)
│   ├── css/            # Stylesheets
│   │   └── custom.css  # Custom theme and utilities
│   └── js/             # JavaScript files
│       └── file-health-components.js
│
└── templates/           # HTML templates (Go template engine)
    ├── components/      # Reusable components
    │   ├── form_field.html    # Generic form field
    │   └── service_card.html  # Service status card
    │
    ├── layouts/         # Layout templates
    │   └── main.html    # Base layout with nav and footer
    │
    └── pages/           # Page templates
        ├── dashboard.html      # Main dashboard
        ├── global.html         # Global configuration
        ├── logs.html           # System logs
        ├── service_config.html # Service configuration
        └── settings.html       # Settings page
```text

## Technologies

- **Templating**: Go html/template

- **CSS Framework**: Tailwind CSS 3.x (CDN)

- **JavaScript**: Alpine.js 3.x (CDN)

- **Theme**: Custom dark theme

- **Icons**: Emoji-based icons

## Pages

### Dashboard (`/`)

Main overview page showing:

- System statistics

- Service status cards

- Download queues (Radarr, Sonarr)

- Active streaming sessions (Jellyfin)

- Quick actions

### Settings (`/settings`)

Service management page with:

- Enable/disable services

- Service configuration overview

- Connection testing

- Cleanup rules configuration

### Global Variables (`/global`)

Global configuration page for:

- User & Group IDs (PUID, PGID)

- System settings (Timezone, Language)

- Base paths (Media, Downloads, Config)

- Network configuration

- Affected services preview

### Service Config (`/services/:name`)

Service-specific configuration:

- Basic settings (Port, Image, Restart policy)

- Path configuration

- Environment variables (from global)

- Advanced settings (VPN, Custom network)

- Connection testing

### Logs (`/logs`)

System logs viewer

## Components

### Service Card (`components/service_card.html`)

Displays service information with:

- Service icon and name

- Status badge

- Quick action buttons (Start, Stop, Restart, Configure)

**Usage in template:**
```html
{{template "service_card" .Service}}
```text

### Form Field (`components/form_field.html`)

Generic form field with validation support:

- Text, number, password, URL inputs

- Select dropdowns

- Textareas

- Checkboxes/toggles

- Error and help text display

- Environment variable source indication

**Usage in template:**
```html
{{template "form_field" .Field}}
```text

## Styling

### Color Scheme (Dark Theme)

```css
--dark-bg: #0f172a        /* Main background */
--dark-surface: #1e293b   /* Cards and panels */
--dark-border: #334155    /* Borders */
--dark-text: #e2e8f0      /* Primary text */
--dark-muted: #94a3b8     /* Secondary text */
```text

### Utility Classes

- `.card-hover` - Card hover effect

- `.status-pulse` - Pulsing animation for status indicators

- `.skeleton` - Loading placeholder animation

- `.toast-notification` - Toast animation

- `.grid-auto-fit` - Responsive grid layout

### Service Icon Colors

- Radarr: Yellow (`#fbbf24`)

- Sonarr: Blue (`#3b82f6`)

- Jellyfin: Purple (`#a855f7`)

- qBittorrent: Cyan (`#06b6d4`)

- Jellyseerr: Indigo (`#6366f1`)

- Jellystat: Green (`#10b981`)

- Bazarr: Orange (`#f97316`)

## Alpine.js Components

### Dashboard

```javascript
dashboard() {
  // Main dashboard component
  // Fetches and displays system stats
  // Auto-refreshes every 30 seconds
}
```text

### Global Config

```javascript
globalConfig() {
  // Global variables management
  // Tracks affected services
  // Validates and saves configuration
}
```text

### Service Config

```javascript
serviceConfig(serviceName) {
  // Service-specific configuration
  // Real-time validation
  // Connection testing
  // Inherits global environment variables
}
```text

## Development

### Adding a New Page

1. Create template in `web/templates/pages/newpage.html`

2. Add route in `cmd/server/main.go`:

   ```go
   app.Get("/newpage", h.NewPage.Index)
   ```

3. Create handler method:

   ```go
   func (h *Handler) Index(c *fiber.Ctx) error {
       return c.Render("pages/newpage", fiber.Map{
           "Title": "New Page",
       }, "layouts/main")
   }
   ```

4. Add navigation link in `layouts/main.html`

### Adding a New Component

1. Create template in `web/templates/components/newcomponent.html`

2. Use in pages with:

   ```html
   {{template "newcomponent" .Data}}
   ```

### Modifying Styles

1. Edit `web/static/css/custom.css`

2. Changes apply immediately (no build step for CSS)

3. Tailwind utilities available via CDN

### Adding JavaScript Functionality

1. Add Alpine.js component in page script section

2. Or create new file in `web/static/js/`

3. Include in layout with:

   ```html
   <script src="/static/js/newfile.js"></script>
   ```

## Best Practices

### Templates

- Use semantic HTML

- Keep components small and focused

- Use Go template conditionals for dynamic content

- Include ARIA labels for accessibility

### Styling

- Prefer Tailwind utility classes

- Use custom CSS only for complex patterns

- Maintain dark theme consistency

- Ensure responsive design (mobile-first)

### JavaScript

- Use Alpine.js for interactivity

- Keep components isolated

- Handle errors gracefully

- Show loading states

- Validate user input

### Performance

- Minimize HTTP requests

- Use CDN for frameworks

- Optimize images

- Lazy load when appropriate

- Use browser caching

## Browser Support

- Chrome/Chromium 90+

- Firefox 88+

- Safari 14+

- Edge 90+

- Mobile browsers (iOS Safari, Chrome Mobile)

## Accessibility

- Semantic HTML structure

- ARIA labels and roles

- Keyboard navigation

- Focus indicators

- Color contrast ratios meet WCAG 2.1 AA

- Form labels and error messages

## Future Enhancements

### Planned Features

- [ ] Dark/light theme toggle

- [ ] Configuration import/export

- [ ] Bulk operations

- [ ] Advanced search/filtering

- [ ] Keyboard shortcuts

- [ ] Toast notification system

- [ ] Modal dialogs

- [ ] Drag and drop support

- [ ] Real-time updates via WebSocket

### Potential Improvements

- Progressive Web App (PWA) support

- Offline functionality

- Custom theme builder

- Multi-language support

- Advanced dashboard customization

- Service dependency visualization

## Troubleshooting

### Styles not loading

1. Check browser console for errors

2. Verify `/static/css/custom.css` is accessible

3. Clear browser cache

4. Check server static file serving

### Alpine.js not working

1. Verify CDN is accessible

2. Check browser console for JavaScript errors

3. Ensure `x-data` is properly initialized

4. Check component syntax

### Page not rendering

1. Verify template file exists

2. Check template syntax

3. Verify route is registered

4. Check handler implementation

5. Review server logs

## License

Part of MediaCheky project - See root LICENSE file
