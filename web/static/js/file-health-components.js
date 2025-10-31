/**
 * Alpine.js Components for File Health Management
 * 
 * This file contains reusable Alpine.js components for the Files Health page.
 * Each component is designed to be modular, maintainable, and follow Alpine.js best practices.
 */

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

/**
 * Formats bytes to human-readable size
 * @param {number} bytes - Size in bytes
 * @returns {string} Formatted size (e.g., "15.2 GB")
 */
function formatBytes(bytes) {
    if (!bytes || bytes === 0) return '0 Bytes';
    
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

/**
 * Formats timestamp to relative date (e.g., "hace 3 días")
 * @param {string|Date} timestamp - Date to format
 * @returns {string} Relative date string
 */
function formatDate(timestamp) {
    if (!timestamp) return 'Nunca';
    
    const date = new Date(timestamp);
    const now = new Date();
    const diffMs = now - date;
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
    
    if (diffDays === 0) return 'Hoy';
    if (diffDays === 1) return 'Ayer';
    if (diffDays < 7) return `Hace ${diffDays} días`;
    if (diffDays < 30) return `Hace ${Math.floor(diffDays / 7)} semanas`;
    if (diffDays < 365) return `Hace ${Math.floor(diffDays / 30)} meses`;
    return `Hace ${Math.floor(diffDays / 365)} años`;
}

/**
 * Gets icon emoji for health status
 * @param {string} status - Health status code
 * @returns {string} Icon emoji
 */
function getStatusIcon(status) {
    const icons = {
        'ok': '✅',
        'orphan_download': '⚠️',
        'only_hardlink': '🔗',
        'dead_torrent': '💀',
        'never_watched': '👁️',
        'missing_metadata': '📝',
        'critical': '🚨',
        'warning': '⚠️',
        'unclassified': '📦'
    };
    return icons[status] || '❓';
}

/**
 * Gets CSS color classes for severity level
 * @param {string} severity - Severity level (ok, warning, critical)
 * @returns {object} Object with border, bg, and text color classes
 */
function getSeverityColor(severity) {
    const colors = {
        'ok': {
            border: 'border-green-600/30',
            bg: 'bg-green-900/20',
            hoverBg: 'hover:bg-green-900/30',
            text: 'text-green-400',
            badge: 'bg-green-900/40 border-green-600/50 text-green-300'
        },
        'warning': {
            border: 'border-yellow-600/30',
            bg: 'bg-yellow-900/20',
            hoverBg: 'hover:bg-yellow-900/30',
            text: 'text-yellow-400',
            badge: 'bg-yellow-900/40 border-yellow-600/50 text-yellow-300'
        },
        'critical': {
            border: 'border-red-600/30',
            bg: 'bg-red-900/20',
            hoverBg: 'hover:bg-red-900/30',
            text: 'text-red-400',
            badge: 'bg-red-900/40 border-red-600/50 text-red-300'
        }
    };
    return colors[severity] || colors['warning'];
}

/**
 * Gets human-readable label for status
 * @param {string} status - Health status code
 * @returns {string} Human-readable label
 */
function getStatusLabel(status) {
    const labels = {
        'ok': 'Saludable',
        'orphan_download': 'Huérfano en Descargas',
        'only_hardlink': 'Solo Hardlink',
        'dead_torrent': 'Torrent Muerto',
        'never_watched': 'Sin Reproducir',
        'missing_metadata': 'Sin Metadata',
        'critical': 'Crítico',
        'warning': 'Atención',
        'unclassified': 'Sin Clasificar'
    };
    return labels[status] || status;
}

/**
 * Checks if a file is an orphan download
 * (in qBittorrent but not in any media management service)
 * @param {object} file - File object to check
 * @returns {boolean} True if file is orphan download
 */
function isOrphanDownload(file) {
    return file.in_qbittorrent && !file.in_jellyfin && !file.in_radarr && !file.in_sonarr;
}

// =============================================================================
// COMPONENT: healthCard(status, count)
// =============================================================================

/**
 * Health Card Component
 * Displays a clickable card showing a health status category with count
 * 
 * @param {string} status - Health status type (ok, orphan_download, etc.)
 * @param {number} count - Number of files in this category
 * @param {string} description - Short description of the category
 * @returns {object} Alpine.js component object
 */
function healthCard(status, count, description) {
    return {
        status: status,
        count: count,
        description: description,
        
        get icon() {
            return getStatusIcon(this.status);
        },
        
        get severity() {
            // Determine severity based on status
            if (this.status === 'ok') return 'ok';
            if (this.status === 'dead_torrent' || this.status === 'critical') return 'critical';
            return 'warning';
        },
        
        get colors() {
            return getSeverityColor(this.severity);
        },
        
        get label() {
            return getStatusLabel(this.status);
        },
        
        handleClick() {
            // Dispatch event to parent component to filter files
            this.$dispatch('filter-by-status', { status: this.status });
        }
    };
}

// =============================================================================
// COMPONENT: fileHealthCard(file, healthReport)
// =============================================================================

/**
 * File Health Card Component
 * Displays detailed information about a single file with health status and actions
 * 
 * @param {object} file - MediaFileInfo object
 * @param {object} healthReport - FileHealthReport object
 * @returns {object} Alpine.js component object
 */
function fileHealthCard(file, healthReport) {
    return {
        file: file,
        healthReport: healthReport || {},
        expanded: false,
        loading: false,
        actionInProgress: null,
        
        init() {
            // If no health report provided, generate basic one
            if (!this.healthReport.status) {
                this.healthReport = this.generateHealthReport();
            }
        },
        
        get status() {
            return this.healthReport.status || 'ok';
        },
        
        get severity() {
            return this.healthReport.severity || 'ok';
        },
        
        get issues() {
            return this.healthReport.issues || [];
        },
        
        get suggestions() {
            return this.healthReport.suggestions || [];
        },
        
        get actions() {
            return this.healthReport.actions || [];
        },
        
        get colors() {
            return getSeverityColor(this.severity);
        },
        
        get statusIcon() {
            return getStatusIcon(this.status);
        },
        
        get statusLabel() {
            return getStatusLabel(this.status);
        },
        
        toggleDetails() {
            this.expanded = !this.expanded;
        },
        
        formatSize(bytes) {
            return formatBytes(bytes);
        },
        
        formatDate(timestamp) {
            return formatDate(timestamp);
        },
        
        /**
         * Execute an action on this file
         * @param {string} action - Action to execute (import, delete, ignore, etc.)
         */
        async executeAction(action) {
            if (this.loading) return;
            
            // Show confirmation for destructive actions
            if (action === 'delete' && !await this.showConfirmation(action)) {
                return;
            }
            
            this.loading = true;
            this.actionInProgress = action;
            
            try {
                let endpoint = '';
                let method = 'POST';
                
                // Map action to API endpoint
                switch (action) {
                    case 'import_radarr':
                        endpoint = `/api/files/${this.file.id}/import-radarr`;
                        break;
                    case 'import_sonarr':
                        endpoint = `/api/files/${this.file.id}/import-sonarr`;
                        break;
                    case 'delete':
                        endpoint = `/api/files/${this.file.id}`;
                        method = 'DELETE';
                        break;
                    case 'ignore':
                        endpoint = `/api/files/${this.file.id}/ignore`;
                        break;
                    case 'clean_hardlink':
                        endpoint = `/api/files/${this.file.id}/clean-hardlink`;
                        break;
                    default:
                        throw new Error(`Acción desconocida: ${action}`);
                }
                
                const response = await fetch(endpoint, { method });
                
                if (!response.ok) {
                    const error = await response.json().catch(() => ({ error: 'Error desconocido' }));
                    throw new Error(error.error || `HTTP ${response.status}`);
                }
                
                // Dispatch success event
                this.$dispatch('file-action-success', {
                    action,
                    file: this.file,
                    message: `✅ Acción "${action}" ejecutada exitosamente`
                });
                
            } catch (error) {
                console.error(`Error executing action ${action}:`, error);
                this.$dispatch('file-action-error', {
                    action,
                    file: this.file,
                    error: error.message
                });
            } finally {
                this.loading = false;
                this.actionInProgress = null;
            }
        },
        
        /**
         * Show confirmation dialog for destructive actions
         * @param {string} action - Action requiring confirmation
         * @returns {Promise<boolean>} True if confirmed
         */
        async showConfirmation(action) {
            const messages = {
                'delete': `⚠️ ¿ELIMINAR "${this.file.title}"?\n\nEsta acción NO se puede deshacer.`,
                'clean_hardlink': `¿Limpiar hardlink de "${this.file.title}"?\n\nEsto eliminará el hardlink sin afectar el archivo original.`
            };
            
            return confirm(messages[action] || `¿Confirmar acción: ${action}?`);
        },
        
        /**
         * Generate basic health report if not provided
         * @returns {object} Basic health report
         */
        generateHealthReport() {
            const report = {
                status: 'ok',
                severity: 'ok',
                issues: [],
                suggestions: [],
                actions: []
            };
            
            // Check for orphan downloads
            if (isOrphanDownload(this.file)) {
                report.status = 'orphan_download';
                report.severity = 'warning';
                report.issues.push('Archivo en descargas pero no en biblioteca');
                report.suggestions.push('Importar a Radarr/Sonarr para agregarlo a Jellyfin');
                report.actions.push('import_radarr', 'import_sonarr', 'delete', 'ignore');
            }
            // Check for hardlink only
            else if (this.file.is_hardlink && !this.file.in_qbittorrent) {
                report.status = 'only_hardlink';
                report.severity = 'warning';
                report.issues.push('Solo queda hardlink, torrent eliminado');
                report.suggestions.push('Limpiar hardlink de descargas sin afectar biblioteca');
                report.actions.push('clean_hardlink', 'ignore');
            }
            // Check for dead torrent
            else if (this.file.torrent_state === 'error' || this.file.torrent_state === 'missingFiles') {
                report.status = 'dead_torrent';
                report.severity = 'critical';
                report.issues.push('Torrent con errores');
                report.suggestions.push('Verificar en qBittorrent o eliminar si ya está en Jellyfin');
                report.actions.push('delete', 'ignore');
            }
            // Check for unwatched
            else if (this.file.in_jellyfin && !this.file.has_been_watched) {
                report.status = 'never_watched';
                report.severity = 'ok';
                report.issues.push('Nunca reproducido');
                report.suggestions.push('Considera eliminar si no te interesa para liberar espacio');
                report.actions.push('delete', 'ignore');
            }
            
            return report;
        }
    };
}

// =============================================================================
// COMPONENT: healthFilters()
// =============================================================================

/**
 * Health Filters Component
 * Manages advanced filtering options for file health
 * 
 * @returns {object} Alpine.js component object
 */
function healthFilters() {
    return {
        // Filter state
        selectedProblem: 'all',
        selectedService: 'all',
        selectedSize: 'all',
        selectedAge: 'all',
        searchQuery: '',
        
        // Filter options
        problemOptions: [
            { value: 'all', label: 'Todos los problemas' },
            { value: 'error', label: 'Errores de torrent' },
            { value: 'missing', label: 'Archivos faltantes' },
            { value: 'orphan', label: 'Descargas huérfanas' },
            { value: 'low_ratio', label: 'Ratio bajo' }
        ],
        
        serviceOptions: [
            { value: 'all', label: 'Todos los servicios' },
            { value: 'jellyfin', label: 'Jellyfin' },
            { value: 'radarr', label: 'Radarr' },
            { value: 'sonarr', label: 'Sonarr' },
            { value: 'qbittorrent', label: 'qBittorrent' }
        ],
        
        sizeOptions: [
            { value: 'all', label: 'Todos los tamaños' },
            { value: 'small', label: 'Pequeño (< 1GB)' },
            { value: 'medium', label: 'Mediano (1-10GB)' },
            { value: 'large', label: 'Grande (10-50GB)' },
            { value: 'huge', label: 'Muy grande (> 50GB)' }
        ],
        
        ageOptions: [
            { value: 'all', label: 'Todas las edades' },
            { value: 'recent', label: '< 1 mes' },
            { value: 'moderate', label: '1-6 meses' },
            { value: 'old', label: '6-12 meses' },
            { value: 'ancient', label: '> 1 año' }
        ],
        
        get hasActiveFilters() {
            return this.selectedProblem !== 'all' ||
                   this.selectedService !== 'all' ||
                   this.selectedSize !== 'all' ||
                   this.selectedAge !== 'all' ||
                   this.searchQuery.trim() !== '';
        },
        
        get activeFilterCount() {
            let count = 0;
            if (this.selectedProblem !== 'all') count++;
            if (this.selectedService !== 'all') count++;
            if (this.selectedSize !== 'all') count++;
            if (this.selectedAge !== 'all') count++;
            if (this.searchQuery.trim() !== '') count++;
            return count;
        },
        
        /**
         * Apply all active filters
         */
        applyFilters() {
            this.$dispatch('filters-changed', {
                problem: this.selectedProblem,
                service: this.selectedService,
                size: this.selectedSize,
                age: this.selectedAge,
                search: this.searchQuery
            });
        },
        
        /**
         * Clear all filters
         */
        clearFilters() {
            if (!this.hasActiveFilters) return;
            
            this.selectedProblem = 'all';
            this.selectedService = 'all';
            this.selectedSize = 'all';
            this.selectedAge = 'all';
            this.searchQuery = '';
            this.applyFilters();
        },
        
        /**
         * Get filtered files based on current filters
         * @param {Array} files - Array of files to filter
         * @returns {Array} Filtered files
         */
        getFilteredFiles(files) {
            let filtered = [...files];
            
            // Filter by problem type
            if (this.selectedProblem !== 'all') {
                filtered = filtered.filter(f => {
                    switch (this.selectedProblem) {
                        case 'error':
                            return f.torrent_state === 'error';
                        case 'missing':
                            return f.torrent_state === 'missingFiles';
                        case 'orphan':
                            return isOrphanDownload(f);
                        case 'low_ratio':
                            return f.in_qbittorrent && (f.seed_ratio || 0) < 1.0;
                        default:
                            return true;
                    }
                });
            }
            
            // Filter by service
            if (this.selectedService !== 'all') {
                filtered = filtered.filter(f => {
                    switch (this.selectedService) {
                        case 'jellyfin':
                            return f.in_jellyfin;
                        case 'radarr':
                            return f.in_radarr;
                        case 'sonarr':
                            return f.in_sonarr;
                        case 'qbittorrent':
                            return f.in_qbittorrent;
                        default:
                            return true;
                    }
                });
            }
            
            // Filter by size
            if (this.selectedSize !== 'all') {
                const GB = 1024 * 1024 * 1024;
                filtered = filtered.filter(f => {
                    const sizeGB = f.size / GB;
                    switch (this.selectedSize) {
                        case 'small':
                            return sizeGB < 1;
                        case 'medium':
                            return sizeGB >= 1 && sizeGB < 10;
                        case 'large':
                            return sizeGB >= 10 && sizeGB < 50;
                        case 'huge':
                            return sizeGB >= 50;
                        default:
                            return true;
                    }
                });
            }
            
            // Filter by age (assuming added_date field exists)
            if (this.selectedAge !== 'all' && filtered.length > 0) {
                const now = new Date();
                filtered = filtered.filter(f => {
                    if (!f.added_date && !f.created_at) return true;
                    const fileDate = new Date(f.added_date || f.created_at);
                    const ageMonths = (now - fileDate) / (1000 * 60 * 60 * 24 * 30);
                    
                    switch (this.selectedAge) {
                        case 'recent':
                            return ageMonths < 1;
                        case 'moderate':
                            return ageMonths >= 1 && ageMonths < 6;
                        case 'old':
                            return ageMonths >= 6 && ageMonths < 12;
                        case 'ancient':
                            return ageMonths >= 12;
                        default:
                            return true;
                    }
                });
            }
            
            // Filter by search query
            if (this.searchQuery.trim() !== '') {
                const query = this.searchQuery.toLowerCase();
                filtered = filtered.filter(f => 
                    (f.title && f.title.toLowerCase().includes(query)) ||
                    (f.file_path && f.file_path.toLowerCase().includes(query))
                );
            }
            
            return filtered;
        }
    };
}

// =============================================================================
// COMPONENT: bulkActions(selectedFiles)
// =============================================================================

/**
 * Bulk Actions Component
 * Manages selection and bulk operations on multiple files
 * 
 * @param {Array} initialSelectedFiles - Initially selected files
 * @returns {object} Alpine.js component object
 */
function bulkActions(initialSelectedFiles = []) {
    return {
        selectedFiles: initialSelectedFiles,
        actionInProgress: false,
        progressCurrent: 0,
        progressTotal: 0,
        
        get selectedCount() {
            return this.selectedFiles.length;
        },
        
        get hasSelection() {
            return this.selectedFiles.length > 0;
        },
        
        /**
         * Toggle select all files in current view
         * @param {Array} files - Files in current view
         */
        toggleSelectAll(files) {
            if (this.selectedFiles.length === files.length) {
                // Deselect all
                this.selectedFiles = [];
            } else {
                // Select all
                this.selectedFiles = [...files];
            }
            this.$dispatch('selection-changed', { selected: this.selectedFiles });
        },
        
        /**
         * Check if all files are selected
         * @param {Array} files - Files in current view
         * @returns {boolean} True if all selected
         */
        isAllSelected(files) {
            return files.length > 0 && this.selectedFiles.length === files.length;
        },
        
        /**
         * Toggle selection of a single file
         * @param {object} file - File to toggle
         */
        toggleSelect(file) {
            const index = this.selectedFiles.findIndex(f => 
                (f.id && f.id === file.id) || f.file_path === file.file_path
            );
            
            if (index >= 0) {
                this.selectedFiles.splice(index, 1);
            } else {
                this.selectedFiles.push(file);
            }
            
            this.$dispatch('selection-changed', { selected: this.selectedFiles });
        },
        
        /**
         * Check if a file is selected
         * @param {object} file - File to check
         * @returns {boolean} True if selected
         */
        isSelected(file) {
            return this.selectedFiles.some(f => 
                (f.id && f.id === file.id) || f.file_path === file.file_path
            );
        },
        
        /**
         * Execute bulk action on selected files
         * @param {string} action - Action to execute
         */
        async executeBulkAction(action) {
            if (!this.hasSelection || this.actionInProgress) return;
            
            // Show confirmation
            if (!await this.confirmBulkAction(action)) return;
            
            this.actionInProgress = true;
            this.progressCurrent = 0;
            this.progressTotal = this.selectedFiles.length;
            
            const results = {
                succeeded: [],
                failed: []
            };
            
            try {
                // Execute action on each file
                for (const file of this.selectedFiles) {
                    try {
                        await this.executeFileAction(file, action);
                        results.succeeded.push(file);
                    } catch (error) {
                        console.error(`Error processing file ${file.id}:`, error);
                        results.failed.push({ file, error: error.message });
                    }
                    this.progressCurrent++;
                }
                
                // Show results
                const message = results.failed.length === 0
                    ? `✅ ${results.succeeded.length} archivo(s) procesados exitosamente`
                    : `⚠️ ${results.succeeded.length} exitosos, ${results.failed.length} fallidos`;
                
                this.$dispatch('bulk-action-complete', {
                    action,
                    results,
                    message
                });
                
                // Clear selection if all succeeded
                if (results.failed.length === 0) {
                    this.selectedFiles = [];
                }
                
            } finally {
                this.actionInProgress = false;
                this.progressCurrent = 0;
                this.progressTotal = 0;
            }
        },
        
        /**
         * Execute action on a single file
         * @param {object} file - File to process
         * @param {string} action - Action to execute
         */
        async executeFileAction(file, action) {
            let endpoint = '';
            let method = 'POST';
            
            switch (action) {
                case 'import_radarr':
                    endpoint = `/api/files/${file.id}/import-radarr`;
                    break;
                case 'import_sonarr':
                    endpoint = `/api/files/${file.id}/import-sonarr`;
                    break;
                case 'delete':
                    endpoint = `/api/files/${file.id}`;
                    method = 'DELETE';
                    break;
                case 'ignore':
                    endpoint = `/api/files/${file.id}/ignore`;
                    break;
                case 'clean_hardlink':
                    endpoint = `/api/files/${file.id}/clean-hardlink`;
                    break;
                case 'tag':
                    endpoint = `/api/files/${file.id}/tag`;
                    break;
                default:
                    throw new Error(`Acción desconocida: ${action}`);
            }
            
            const response = await fetch(endpoint, { method });
            
            if (!response.ok) {
                const error = await response.json().catch(() => ({ error: 'Error desconocido' }));
                throw new Error(error.error || `HTTP ${response.status}`);
            }
        },
        
        /**
         * Show confirmation dialog for bulk action
         * @param {string} action - Action to confirm
         * @returns {Promise<boolean>} True if confirmed
         */
        async confirmBulkAction(action) {
            const count = this.selectedFiles.length;
            const messages = {
                'delete': `⚠️ ¿ELIMINAR ${count} archivo(s)?\n\nEsta acción NO se puede deshacer.`,
                'import_radarr': `¿Importar ${count} archivo(s) a Radarr?`,
                'import_sonarr': `¿Importar ${count} archivo(s) a Sonarr?`,
                'ignore': `¿Marcar ${count} archivo(s) como ignorados?`,
                'clean_hardlink': `¿Limpiar hardlinks de ${count} archivo(s)?`,
                'tag': `¿Etiquetar ${count} archivo(s)?`
            };
            
            return confirm(messages[action] || `¿Confirmar acción ${action} en ${count} archivos?`);
        },
        
        get progressPercent() {
            if (this.progressTotal === 0) return 0;
            return Math.round((this.progressCurrent / this.progressTotal) * 100);
        }
    };
}

// =============================================================================
// COMPONENT: healthStatusBadge(status, severity)
// =============================================================================

/**
 * Health Status Badge Component
 * Displays a badge showing health status with appropriate styling
 * 
 * @param {string} status - Health status code
 * @param {string} severity - Severity level (ok, warning, critical)
 * @param {string} size - Badge size (sm, md, lg)
 * @returns {object} Alpine.js component object
 */
function healthStatusBadge(status, severity = 'ok', size = 'md') {
    return {
        status: status,
        severity: severity,
        size: size,
        
        get icon() {
            return getStatusIcon(this.status);
        },
        
        get label() {
            return getStatusLabel(this.status);
        },
        
        get colors() {
            return getSeverityColor(this.severity);
        },
        
        get sizeClasses() {
            const sizes = {
                'sm': 'px-2 py-0.5 text-xs',
                'md': 'px-3 py-1 text-sm',
                'lg': 'px-4 py-2 text-base'
            };
            return sizes[this.size] || sizes['md'];
        }
    };
}

// =============================================================================
// COMPONENT: serviceStatusIndicator(service, isActive, details)
// =============================================================================

/**
 * Service Status Indicator Component
 * Shows status of a file in a specific service (Radarr, Sonarr, etc.)
 * 
 * @param {string} service - Service name (radarr, sonarr, jellyfin, etc.)
 * @param {boolean} isActive - Whether file is in this service
 * @param {object} details - Additional details (torrent state, ratio, etc.)
 * @returns {object} Alpine.js component object
 */
function serviceStatusIndicator(service, isActive, details = {}) {
    return {
        service: service,
        isActive: isActive,
        details: details,
        
        get icon() {
            const icons = {
                'radarr': '🎬',
                'sonarr': '📺',
                'jellyfin': '📚',
                'jellyseerr': '📋',
                'qbittorrent': '🌊',
                'jellystat': '📊'
            };
            return icons[this.service] || '📦';
        },
        
        get label() {
            const labels = {
                'radarr': 'Radarr',
                'sonarr': 'Sonarr',
                'jellyfin': 'Jellyfin',
                'jellyseerr': 'Jellyseerr',
                'qbittorrent': 'qBittorrent',
                'jellystat': 'Jellystat'
            };
            return labels[this.service] || this.service;
        },
        
        get statusColor() {
            return this.isActive ? 'text-green-400' : 'text-gray-500';
        },
        
        get tooltip() {
            if (!this.isActive) {
                return `No en ${this.label}`;
            }
            
            const parts = [`En ${this.label}`];
            
            // Add service-specific details
            if (this.service === 'qbittorrent' && this.details) {
                if (this.details.torrent_state) {
                    parts.push(`Estado: ${this.details.torrent_state}`);
                }
                if (this.details.seed_ratio !== undefined) {
                    parts.push(`Ratio: ${this.details.seed_ratio.toFixed(2)}`);
                }
                if (this.details.is_seeding) {
                    parts.push('Seeding activo');
                }
            }
            
            if (this.service === 'jellyfin' && this.details) {
                if (this.details.has_been_watched) {
                    parts.push('Reproducido');
                } else {
                    parts.push('Sin reproducir');
                }
            }
            
            return parts.join(' | ');
        }
    };
}

// Export functions for use in other scripts (if using modules)
if (typeof module !== 'undefined' && module.exports) {
    module.exports = {
        // Helper functions
        formatBytes,
        formatDate,
        getStatusIcon,
        getSeverityColor,
        getStatusLabel,
        isOrphanDownload,
        
        // Components
        healthCard,
        fileHealthCard,
        healthFilters,
        bulkActions,
        healthStatusBadge,
        serviceStatusIndicator
    };
}
