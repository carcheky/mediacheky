# UI Standards and Guidelines

## Toast Notifications

**All feedback messages MUST use floating toast notifications, NOT inline messages.**

### Implementation

Every page with user actions must include:

```html
<!-- Toast Notifications (Floating) -->
<div x-show="toast.show"
     x-transition:enter="transition ease-out duration-300"
     x-transition:enter-start="opacity-0 translate-y-2"
     x-transition:enter-end="opacity-100 translate-y-0"
     x-transition:leave="transition ease-in duration-200"
     x-transition:leave-start="opacity-100 translate-y-0"
     x-transition:leave-end="opacity-0 translate-y-2"
     class="fixed top-4 right-4 z-50 max-w-md"
     style="display: none;">
    <div :class="{
        'bg-green-900/95 border-green-700': toast.type === 'success',
        'bg-red-900/95 border-red-700': toast.type === 'error'
    }" class="backdrop-blur-sm border rounded-lg shadow-2xl p-4 flex items-start gap-3">
        <!-- Icon -->
        <div class="flex-shrink-0">
            <svg x-show="toast.type === 'success'" class="w-6 h-6 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
            </svg>
            <svg x-show="toast.type === 'error'" class="w-6 h-6 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
            </svg>
        </div>
        <!-- Message -->
        <div class="flex-1">
            <p :class="{
                'text-green-100': toast.type === 'success',
                'text-red-100': toast.type === 'error'
            }" class="text-sm font-medium" x-text="toast.message"></p>
        </div>
        <!-- Close Button -->
        <button @click="toast.show = false" class="flex-shrink-0 text-gray-400 hover:text-gray-200">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
            </svg>
        </button>
    </div>
</div>
```

### Alpine.js Data Structure

```javascript
toast: {
    show: false,
    message: '',
    type: 'success'
}
```

### Toast Function

```javascript
showToast(message, type = 'success') {
    this.toast.message = message;
    this.toast.type = type;
    this.toast.show = true;

    // Auto-hide after delay
    setTimeout(() => {
        this.toast.show = false;
    }, type === 'success' ? 5000 : 8000);
}
```

### Timing Rules

- **Success messages**: Auto-dismiss after **5 seconds**
- **Error messages**: Auto-dismiss after **8 seconds**
- User can manually close at any time

### DO NOT Use

❌ **Inline messages:**
```html
<!-- WRONG -->
<div x-show="message" class="p-4 rounded-lg border">
    <p x-text="message"></p>
</div>
```

❌ **Static positioned messages**

❌ **Alert() dialogs**

## Page Structure

### Navigation

All pages must include in Settings:
- **Services Tab**: Enable/disable services with Configure button
- **Global Variables Tab**: PUID, PGID, TZ, paths, network config

**DO NOT create separate `/global` route** - Global Variables is a tab within Settings.

### Tabs Structure

```html
<!-- Main Tabs -->
<div class="border-b border-dark-border">
    <nav class="-mb-px flex space-x-8">
        <button @click="activeTab = 'services'"
            :class="activeTab === 'services' ? 'border-blue-500 text-blue-400' : 'border-transparent text-dark-muted hover:text-dark-text hover:border-slate-600'"
            class="whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm transition-colors">
            Services
        </button>
        <button @click="activeTab = 'global'"
            :class="activeTab === 'global' ? 'border-blue-500 text-blue-400' : 'border-transparent text-dark-muted hover:text-dark-text hover:border-slate-600'"
            class="whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm transition-colors">
            Global Variables
        </button>
    </nav>
</div>
```

## Service Configuration

### Configure Button

Only show when service is enabled:

```html
<a x-show="config.services[service.id]?.enabled" 
   :href="'/services/' + service.id"
   class="px-3 py-1 bg-blue-600 text-white rounded-md hover:bg-blue-700 text-sm">
    Configure
</a>
```

### URL Pattern

Service configuration pages use: `/services/:name`

**NOT**: `/services/:name/config` ❌

## Data Persistence

### Database Naming

- Database file: `mediacheky.db` (NOT `keepercheky.db`)
- Location: `./data/mediacheky.db`

### Environment Variables

Prefix: `MEDIACHEKY_` (NOT `KEEPERCHEKY_`)

Examples:
```bash
MEDIACHEKY_CLIENTS_RADARR_ENABLED=true
MEDIACHEKY_CLIENTS_RADARR_URL=http://radarr:7878
```

## Docker Integration

### Network

All services MUST use: `mediacheky-net` network

**ALWAYS connected, not conditional**

### Port Exposure

Ports are **NOT exposed by default**

Optional exposure via checkbox:
```yaml
{{- if .ExposePort }}
ports:
  - "{{ if .HostPort }}{{ .HostPort }}{{ else }}{{ .Port }}{{ end }}:7878"
{{- end }}
```

## Files and Documentation

### Documentation Location

ALL documentation must be in `/docs` folder.

### Root Files

Only keep in root:
- `README.md` - Project overview
- `QUICKSTART.md` - Quick start guide
- `DEVELOPMENT.md` - Development setup
- `CHANGELOG.md` - Version history
- `LICENSE` - License file
- `AGENTS.md` - AI agent guidelines

### Deprecated Files

Remove from root if found:
- Architecture diagrams
- Detailed implementation guides
- API documentation
- Testing documentation

Move to `/docs` instead.

## Examples

### ✅ Correct Implementation

**settings.html:**
```html
<div x-data="settings()">
    <!-- Tab Navigation -->
    <div class="border-b border-dark-border">
        <nav class="-mb-px flex space-x-8">
            <button @click="activeTab = 'services'">Services</button>
            <button @click="activeTab = 'global'">Global Variables</button>
        </nav>
    </div>
    
    <!-- Services Tab -->
    <div x-show="activeTab === 'services'">
        <!-- Enable/disable with Configure button -->
    </div>
    
    <!-- Global Tab -->
    <div x-show="activeTab === 'global'">
        <!-- PUID, PGID, TZ, paths, network -->
    </div>
    
    <!-- Floating Toast -->
    <div x-show="toast.show" class="fixed top-4 right-4 z-50">
        <!-- Toast content -->
    </div>
</div>
```

### ❌ Incorrect Implementation

**Wrong:**
```html
<!-- Separate /global page -->
<nav>
    <a href="/global">Global</a>
</nav>

<!-- Inline message -->
<div x-show="message" class="p-4">
    <p x-text="message"></p>
</div>
```

## Checklist for New Pages

- [ ] Floating toast notifications (top-right, fixed position)
- [ ] Auto-dismiss (5s success, 8s error)
- [ ] Manual close button
- [ ] No inline feedback messages
- [ ] Correct database path (`mediacheky.db`)
- [ ] Correct environment variable prefix (`MEDIACHEKY_`)
- [ ] Service URLs without `/config` suffix
- [ ] Global Variables in Settings tab (not separate page)
- [ ] Documentation in `/docs` folder

## Related Documents

- [UI Implementation](UI_IMPLEMENTATION.md) - Technical implementation details
- [Project Plan](PROJECT_PLAN.md) - Overall project roadmap
- [API Documentation](API.md) - Backend API reference
