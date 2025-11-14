/**
 * Shared utilities and constants for MediaCheky frontend
 * @module shared
 */

/**
 * Service logo mapping - official logos for each service
 * @constant {Object.<string, {url: string, alt: string, width: string}>}
 */
const SERVICE_LOGOS = {
    'radarr': {
        url: 'https://radarr.video/img/logo.png',
        alt: 'Radarr Logo',
        width: '40px'
    },
    'sonarr': {
        url: 'https://sonarr.tv/img/logo.png',
        alt: 'Sonarr Logo',
        width: '40px'
    },
    'jellyfin': {
        url: 'https://raw.githubusercontent.com/jellyfin/jellyfin-ux/master/branding/SVG/jellyfin-logo.svg',
        alt: 'Jellyfin Logo',
        width: '40px'
    },
    'qbittorrent': {
        url: 'https://www.qbittorrent.org/images/qbittorrent-nox.svg',
        alt: 'qBittorrent Logo',
        width: '40px'
    },
    'jellyseerr': {
        url: 'https://raw.githubusercontent.com/Fallenbagel/jellyseerr/develop/public/logo_textless.svg',
        alt: 'Jellyseerr Logo',
        width: '40px'
    },
    'jellystat': {
        url: 'https://raw.githubusercontent.com/CyferShepard/Jellystat/main/docker/icon.png',
        alt: 'Jellystat Logo',
        width: '40px'
    },
    'bazarr': {
        url: 'https://raw.githubusercontent.com/morpheus65535/bazarr/master/static/images/favicon.ico',
        alt: 'Bazarr Logo',
        width: '40px'
    },
    'prowlarr': {
        url: 'https://prowlarr.com/img/logo.png',
        alt: 'Prowlarr Logo',
        width: '40px'
    },
    'lidarr': {
        url: 'https://raw.githubusercontent.com/lidarr/Lidarr/develop/Logo/128.png',
        alt: 'Lidarr Logo',
        width: '40px'
    }
};

/**
 * Get service icon by name - returns HTML img tag with fallback emoji
 * @param {string} serviceName - The name of the service
 * @returns {string} HTML img tag or emoji fallback
 */
function getServiceIcon(serviceName) {
    const logo = SERVICE_LOGOS[serviceName.toLowerCase()];
    if (logo) {
        return `<img src="${logo.url}" alt="${logo.alt}" style="width: ${logo.width}; height: ${logo.width}; object-fit: contain;" onerror="this.style.display='none'; this.parentElement.textContent='⚙️'">`;
    }
    return '⚙️';
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
