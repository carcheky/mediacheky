# Settings Page Refactoring

## Overview

This document describes the refactoring of the settings page to eliminate code duplication and improve maintainability by using reusable Alpine.js components, plus the transition from a tabbed interface to a unified single-page view.

## Problem

The original `settings.html` file had 986 lines with significant duplication:
- Each service (Radarr, Sonarr, Jellyfin, Jellyseerr, Jellystat, qBittorrent) had nearly identical HTML structure
- Connection test UI was duplicated 6 times
- System info display was duplicated with slight variations
- Adding a new service required ~140 lines of mostly duplicated code
- Tab-based navigation required switching views to see different services

## Solution

Refactored the page to use data-driven rendering with Alpine.js:
- Reduced from 986 lines to 456 lines (-54%)
- Single reusable panel template for all services
- Service configuration defined as metadata
- Dynamic field generation based on service definitions
- **Unified single-page view**: All services visible simultaneously without tabs
- **Test All functionality**: Test all enabled services with one click

## Architecture

### Service Metadata Structure

```javascript
services: [
    {
        id: 'radarr',                    // Unique service identifier
        label: 'Radarr',                 // Display name
        fields: [                         // Form fields configuration
            { 
                name: 'url',              // Field name (maps to config.services.radarr.url)
                label: 'URL',             // Display label
                type: 'url',              // Input type
                placeholder: 'http://...' // Placeholder text
            },
            // ... more fields
        ],
        systemInfoFields: [               // System info display configuration
            { 
                key: 'version',           // Key in system_info response
                label: 'Version',         // Display label
                mono: false,              // Use monospace font?
                span: false               // Span 2 columns?
            },
            // ... more fields
        ]
    },
    // ... more services
]
```

### Component Hierarchy

```
settings()                              // Main Alpine.js component
├── services[]                          // Service metadata array
├── config.services.*                   // Service configurations
├── connectionResults.*                 // Connection test results
└── envSources.*                        // Environment variable sources

systemInfoDisplay()                     // System info display component
├── systemInfo                          // System info data
├── fieldsDef                           // Field definitions
└── displayFields                       // Computed display fields
```

### Dynamic Rendering Flow

1. **Panel Generation** (All visible simultaneously)
   ```html
   <template x-for="service in services" :key="service.id">
       <div class="bg-dark-surface ...">
           <!-- Panel content always visible -->
       </div>
   </template>
   ```

2. **Field Generation**
   ```html
   <template x-for="field in service.fields" :key="field.name">
       <input 
           :type="field.type" 
           x-model="config.services[service.id][field.name]"
           :placeholder="field.placeholder">
   </template>
   ```

3. **System Info Display**
   ```html
   <template x-for="field in getSystemInfoDisplay(service.id)" :key="field.key">
       <p x-show="field.html" x-html="field.value"></p>
       <p x-show="!field.html" x-text="field.value"></p>
   </template>
   ```

4. **Test All Connections**
   ```javascript
   async testAllConnections() {
       this.testing = true;
       this.connectionResults = {};
       
       // Test all enabled services sequentially
       for (const service of this.services) {
           if (this.config.services[service.id]?.enabled) {
               await this.testConnection(service.id);
           }
       }
       
       this.testing = false;
   }
   ```

## Key Improvements

### 1. Security

**XSS Prevention**: All user-controlled content is properly escaped
- HTML escaping for string values using `escapeHtml()`
- Separate rendering paths for HTML vs text content
- `x-text` used by default (safe), `x-html` only for controlled boolean badges
- `needsHtml()` flag explicitly controls HTML rendering

**Before** (vulnerable):
```javascript
formatValue(key, value) {
    return String(value);  // No escaping - XSS risk!
}
```

**After** (secure):
```javascript
formatValue(key, value) {
    if (typeof value === 'boolean') {
        return '<span>...</span>';  // Controlled HTML for badges
    }
    return this.escapeHtml(String(value));  // All other values escaped
}
```

### 2. Performance

**Avoided Re-initialization**: System info display computed once in parent component
- Removed nested `x-data` that re-initialized on visibility changes
- `getSystemInfoDisplay()` method called from parent scope
- Improved performance with all panels visible simultaneously

### 3. User Experience

**Unified Single-Page View**: All configurations visible without navigation
- **Before**: Tab-based interface, one service visible at a time
- **After**: All services visible on one scrollable page
- **Benefit**: Complete overview of system configuration
- **Test All**: Single button to test all enabled services at once

### 4. Maintainability

**Before**: Adding a new service
```html
<!-- ~140 lines of HTML -->
<div x-show="activeTab === 'newservice'">
    <div class="flex items-center...">
        <h2>New Service Configuration</h2>
        <!-- ... -->
    </div>
    <div class="space-y-4">
        <div>
            <label>URL</label>
            <input type="url" x-model="config.services.newservice.url" ...>
        </div>
        <!-- ... more fields ... -->
        <button @click="testConnection('newservice')">Test</button>
        <!-- ... connection results ... -->
    </div>
</div>
```

**After**: Adding a new service
```javascript
// ~15 lines of metadata
{
    id: 'newservice',
    label: 'New Service',
    fields: [
        { name: 'url', label: 'URL', type: 'url', placeholder: 'http://...' },
        { name: 'api_key', label: 'API Key', type: 'password', placeholder: '...' }
    ],
    systemInfoFields: [
        { key: 'version', label: 'Version' }
    ]
}
```

### 5. Consistency

All services automatically have:
- Same visual styling
- Same behavior
- Same validation
- Same error handling

### 6. Testing Workflow

**Individual Tests**: Test specific services independently
```html
<button @click="testConnection(service.id)">Test Connection</button>
```

**Bulk Testing**: Test all enabled services at once
```html
<button @click="testAllConnections()">Test All Connections</button>
```

The `testAllConnections()` method:
- Tests all enabled services sequentially
- Displays individual results for each service
- Provides complete validation in one operation

### 7. Bug Fixes

Fixed once → applies to all services
- Before: Had to fix 6 times
- After: Fix in one place

### 4. Extensibility

Easy to add features globally:
- New field type? Add to field template
- New validation? Add to input binding
- New UI element? Add to panel template

## Implementation Details

### Environment Variable Handling

The refactored version maintains compatibility with environment variable overrides:

```html
<input 
    x-model="config.services[service.id][field.name]"
    :disabled="envSources[service.id]?.[field.name]"
    :class="envSources[service.id]?.[field.name] ? 'bg-slate-800/50 cursor-not-allowed' : 'bg-dark-bg'">
```

### Dynamic System Info

The `systemInfoDisplay()` helper provides secure formatting for different value types with XSS prevention:

```javascript
formatValue(key, value) {
    if (value === null || value === undefined) return 'N/A';
    if (typeof value === 'boolean') {
        // Return HTML for boolean values (styled badges)
        return value ? 
            '<span class="px-2 py-1 bg-green-900/30 border border-green-600/50 text-green-300 rounded text-xs">Yes</span>' :
            '<span class="px-2 py-1 bg-gray-900/30 border border-gray-600/50 text-gray-300 rounded text-xs">No</span>';
    }
    // Return escaped plain text for all other values
    return this.escapeHtml(String(value));
}

escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}
```

The component uses a dual rendering approach for security:
- `x-text` for plain text values (most cases) - safe from XSS
- `x-html` only for boolean badges - controlled HTML output
- `needsHtml()` method determines which renderer to use

### Performance Optimization

System info display is computed in the parent component rather than using nested `x-data`:

```javascript
// In settings component
getSystemInfoDisplay(serviceId) {
    const service = this.services.find(s => s.id === serviceId);
    const systemInfo = this.connectionResults[serviceId]?.system_info;
    
    if (!systemInfo || !service) {
        return [];
    }
    
    // Use the systemInfoDisplay helper to generate fields
    const helper = systemInfoDisplay(systemInfo, service.systemInfoFields || []);
    return helper.displayFields;
}
```

This avoids re-initialization of the Alpine component on every visibility change.

### Connection Testing

Connection testing works the same way for all services:

```javascript
async testConnection(service) {
    this.testing = true;
    this.connectionResults[service] = null;

    const response = await fetch(`/api/config/test/${service}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(this.config.services[service])
    });

    // ... handle response ...
}
```

## Testing

### Automated Tests
- ✅ Go build succeeds
- ✅ Go fmt passes
- ✅ Go vet passes
- ✅ All unit tests pass

### Manual Testing Required
Since services need to be started by the user, manual testing should verify:
- [ ] All tabs render correctly
- [ ] Tab switching works
- [ ] Form fields display properly
- [ ] Connection tests work for each service
- [ ] System info displays after successful connection
- [ ] Configuration saves correctly
- [ ] Environment variable overrides work

## Future Enhancements

Possible improvements for future iterations:

1. **Field Validation**
   - Add validation rules to field metadata
   - Display validation errors inline

2. **Conditional Fields**
   - Show/hide fields based on other field values
   - Example: Advanced settings toggle

3. **Field Groups**
   - Group related fields together
   - Collapsible sections

4. **Custom Field Types**
   - File picker
   - Color picker
   - Multi-select

5. **Service Templates**
   - Define service types (arr-based, media-server, etc.)
   - Inherit common configuration

## Migration Guide

For developers working with this code:

### Adding a New Service

1. Add service metadata to `services` array:
```javascript
{
    id: 'bazarr',
    label: 'Bazarr',
    fields: [
        { name: 'url', label: 'URL', type: 'url', placeholder: 'http://localhost:6767' },
        { name: 'api_key', label: 'API Key', type: 'password', placeholder: 'Enter Bazarr API Key' }
    ],
    systemInfoFields: [
        { key: 'version', label: 'Version' },
        { key: 'branch', label: 'Branch' }
    ]
}
```

2. Add backend configuration in `internal/config/config.go`:
```go
type ClientsConfig struct {
    // ... existing clients ...
    Bazarr    BazarrConfig    `mapstructure:"bazarr"`
}

type BazarrConfig struct {
    Enabled bool   `mapstructure:"enabled"`
    URL     string `mapstructure:"url"`
    APIKey  string `mapstructure:"api_key"`
}
```

3. Implement connection test in `internal/handler/settings.go`

### Modifying Field Display

To change how a field is rendered, modify the field template:

```html
<template x-for="field in service.fields" :key="field.name">
    <div>
        <!-- Add custom rendering logic here -->
        <input 
            :type="field.type || 'text'" 
            x-model="config.services[service.id][field.name]"
            :placeholder="field.placeholder">
    </div>
</template>
```

### Customizing System Info

To customize system info display for a specific service, define `systemInfoFields`:

```javascript
systemInfoFields: [
    { key: 'version', label: 'Version' },
    { key: 'server_id', label: 'Server ID', mono: true },      // Monospace font
    { key: 'local_address', label: 'Address', mono: true, span: true }  // Full width
]
```

## Conclusion

This refactoring significantly improves the maintainability and extensibility of the settings page while reducing code size by 54%. The data-driven approach makes it easy to add new services and modify behavior globally.

## References

- Original file: `web/templates/pages/settings-original-backup.html`
- Refactored file: `web/templates/pages/settings.html`
- Alpine.js documentation: https://alpinejs.dev/
