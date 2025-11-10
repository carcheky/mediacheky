/**
 * Shared utilities and constants for MediaCheky frontend
 * @module shared
 */

/**
 * Service icon mapping - emoji icons for each service
 * @constant {Object.<string, string>}
 */
const SERVICE_ICONS = {
    'radarr': '🎬',
    'sonarr': '📺',
    'jellyfin': '🍿',
    'qbittorrent': '📡',
    'jellyseerr': '📋',
    'jellystat': '📊',
    'bazarr': '🗣️'
};

/**
 * Get service icon by name
 * @param {string} serviceName - The name of the service
 * @returns {string} The emoji icon for the service, or default gear icon
 */
function getServiceIcon(serviceName) {
    return SERVICE_ICONS[serviceName.toLowerCase()] || '⚙️';
}

/**
 * Create a showMessage function with configurable timeout for Alpine.js components
 * @returns {Function} A function that displays messages with auto-dismiss
 */
function createShowMessage() {
    return function(msg, type, timeout) {
        type = type || 'success';
        timeout = timeout || null;
        this.message = msg;
        this.messageType = type;
        // Set default timeout based on message type if not provided
        const duration = timeout !== null ? timeout : (type === 'error' ? 8000 : 5000);
        setTimeout(() => {
            this.message = '';
        }, duration);
    };
}

/**
 * Format bytes to human readable string
 * @param {number} bytes - The number of bytes to format
 * @returns {string} Formatted string with appropriate unit (B, KB, MB, GB, TB)
 */
function formatBytes(bytes) {
    if (bytes === 0) {
        return '0 B';
    }
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

/**
 * Format date relative to now in Spanish
 * @param {string} dateStr - ISO date string
 * @returns {string} Relative date string in Spanish (e.g., "Hoy", "Ayer", "Hace 3 días")
 */
function formatDate(dateStr) {
    if (!dateStr) {
        return '';
    }
    const date = new Date(dateStr);
    const now = new Date();
    const diffTime = Math.abs(now - date);
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
    
    if (diffDays === 0) {
        return 'Hoy';
    }
    if (diffDays === 1) {
        return 'Ayer';
    }
    if (diffDays < 7) {
        return 'Hace ' + diffDays + ' días';
    }
    if (diffDays < 30) {
        return 'Hace ' + Math.floor(diffDays / 7) + ' semanas';
    }
    return date.toLocaleDateString('es-ES', { month: 'short', day: 'numeric' });
}
