# Radarr and Sonarr Configuration Features Documentation

## Executive Summary

This document provides a comprehensive listing of all configuration features available in Radarr (movie management) and Sonarr (TV series management) configuration pages. This documentation is intended to inform the development of MediaCheky's configuration interface for these services.

---

## 🎬 RADARR - Movie Management Configuration

Radarr is a movie collection manager that automatically downloads, organizes, and manages movie files. Below is a complete listing of all configuration sections and their features.

### 1. Media Management

**Purpose**: Controls how Radarr handles movie files, naming, and storage.

#### File Management
- **Auto Unmonitor Previously Downloaded Movies**: Automatically unmonitor movies after they are downloaded
- **Recycle Bin Path**: Path where deleted files are moved instead of permanent deletion
- **Recycle Bin Cleanup Days**: Number of days to keep files in recycle bin before permanent deletion
- **Download Propers and Repacks**: Control how Radarr handles proper/repack releases
  - Prefer and Upgrade
  - Do Not Upgrade Automatically
  - Do Not Prefer
- **Create Empty Movie Folders**: Create folder structure for movies even if not yet downloaded
- **Delete Empty Folders**: Automatically delete empty folders
- **File Date**: What date to use for file timestamps
  - None
  - In Cinemas Date
  - Physical Release Date
- **Rescan After Refresh**: When to rescan movie folder after refresh
  - Always
  - After Manual Refresh
  - Never
- **Auto Rename Folders**: Automatically rename movie folders to match naming scheme

#### Permissions (Linux)
- **Set Permissions**: Enable/disable permission setting for Linux systems
- **chmod Folder**: Folder permission mask (e.g., 755)
- **chown Group**: Group ownership for folders

#### Importing
- **Skip Free Space Check When Importing**: Skip checking available disk space
- **Minimum Free Space When Importing**: Minimum free space required (in MB)
- **Use Hardlinks Instead of Copy**: Use hardlinks when possible to save space
- **Import Using Script**: Use custom script for importing
- **Script Import Path**: Path to import script
- **Import Extra Files**: Import extra files (subtitles, NFO, etc.)
- **Extra File Extensions**: Extensions to import (comma-separated)
- **Enable MediaInfo**: Enable MediaInfo scanning for technical details

#### Root Folders
- **Add/Remove Root Folders**: Manage folders where movies are stored
- **Default Root Folder**: Set default storage location for new movies

#### Naming
- **Rename Movies**: Enable automatic renaming based on naming scheme
- **Replace Illegal Characters**: Replace characters not allowed in filenames
- **Colon Replacement**: How to replace colons in movie titles
- **Standard Movie Format**: Template for movie file naming
- **Movie Folder Format**: Template for movie folder naming
- **Tokens Available**: Variables for use in naming templates
  - {Movie Title}
  - {Movie TitleThe}
  - {Movie CleanTitle}
  - {Movie OriginalTitle}
  - {Movie Year}
  - {Movie TmdbId}
  - {Movie ImdbId}
  - {Quality Full}
  - {Quality Title}
  - {MediaInfo VideoCodec}
  - {MediaInfo AudioCodec}
  - {MediaInfo AudioChannels}
  - {Release Group}
  - {Edition Tags}

### 2. Profiles

**Purpose**: Define quality and upgrade preferences for movies.

#### Quality Profiles
- **Add/Edit/Delete Quality Profiles**: Create custom quality preference profiles
- **Profile Name**: Name for the profile
- **Upgrades Allowed**: Enable automatic quality upgrades
- **Upgrade Until**: Maximum quality to upgrade to
- **Quality Order**: Drag-and-drop priority ordering
- **Available Qualities**:
  - Unknown
  - WORKPRINT
  - CAM
  - TELESYNC
  - TELECINE
  - REGIONAL
  - DVDSCR
  - SDTV
  - DVD
  - DVD-R
  - WEBDL-480p
  - WEBRip-480p
  - Bluray-480p
  - Bluray-576p
  - HDTV-720p
  - WEBDL-720p
  - WEBRip-720p
  - Bluray-720p
  - HDTV-1080p
  - WEBDL-1080p
  - WEBRip-1080p
  - Bluray-1080p
  - Remux-1080p
  - HDTV-2160p
  - WEBDL-2160p
  - WEBRip-2160p
  - Bluray-2160p
  - Remux-2160p

#### Delay Profiles
- **Protocol Delay**: Prefer Usenet or Torrent
- **Usenet Delay**: Minutes to wait before grabbing from Usenet
- **Torrent Delay**: Minutes to wait before grabbing from Torrent
- **Bypass if Highest Quality**: Skip delay if release is highest quality
- **Tags**: Apply profile to specific movies via tags

#### Release Profiles (Deprecated in newer versions)
- **Must Contain**: Terms that must be in release name
- **Must Not Contain**: Terms that disqualify a release
- **Preferred Words**: Terms that increase release score
- **Include Preferred**: Include preferred terms in filename
- **Tags**: Apply to specific movies

### 3. Quality

**Purpose**: Define quality definitions and file size limits.

#### Quality Definitions
- **Quality Title**: Name of quality tier
- **Minimum Size**: Minimum file size (MB per minute)
- **Maximum Size**: Maximum file size (MB per minute)
- **Preferred Size**: Preferred file size (MB per minute)
- **Available for all qualities listed in Profiles section**

#### Reset Quality Definitions
- **Reset to Defaults**: Restore original quality size settings

### 4. Custom Formats

**Purpose**: Advanced release filtering and scoring based on custom criteria.

#### Custom Format Management
- **Add/Edit/Delete Custom Formats**: Create custom release filters
- **Format Name**: Name for the custom format
- **Specifications**: Conditions to match releases
  - Release Title
  - Edition
  - Language
  - Indexer
  - Size
  - Source
  - Resolution
  - Media Info (Video/Audio codec)
  - Quality Modifier
  - Custom Format Tags

#### Custom Format Scoring
- **Format Score**: Points added/subtracted when format matches
- **Profile Integration**: Assign scores per quality profile
- **Minimum Custom Format Score**: Minimum score required for upgrade

### 5. Indexers

**Purpose**: Configure sources for finding movie releases.

#### Indexer Management
- **Add/Edit/Delete Indexers**: Manage indexer connections
- **Enable/Disable Indexers**: Turn indexers on/off
- **Supported Indexers**: Newznab, Torznab, Torrent RSS, and presets for popular indexers

#### Indexer Configuration (Per Indexer)
- **Name**: Indexer display name
- **Enable RSS**: Enable RSS sync for this indexer
- **Enable Automatic Search**: Enable for automatic searches
- **Enable Interactive Search**: Enable for manual searches
- **URL**: Indexer URL
- **API Path**: API endpoint path
- **API Key**: Authentication key
- **Categories**: Which categories to search
- **Additional Parameters**: Extra parameters for searches
- **Minimum Seeders**: Minimum number of seeders (torrents)
- **Seed Ratio**: Required seed ratio (torrents)
- **Seed Time**: Required seed time in minutes (torrents)
- **Tags**: Apply indexer to specific movies only

#### Indexer Options (Global)
- **Minimum Age**: Minimum age of releases in minutes
- **Retention**: Maximum age of releases in days (Usenet)
- **Maximum Size**: Maximum release size in MB
- **RSS Sync Interval**: How often to sync RSS feeds (minutes)
- **Prefer Indexer Flags**: Prefer releases with indexer-specific flags
- **Availability Delay**: Days to wait after availability date
- **Allow Hardcoded Subs**: Allow releases with hardcoded subtitles
- **Whitelisted Hardcoded Subs**: Languages allowed for hardcoded subs

### 6. Download Clients

**Purpose**: Configure torrent/usenet clients for downloading movies.

#### Download Client Management
- **Add/Edit/Delete Download Clients**: Manage download client connections
- **Enable/Disable Clients**: Turn clients on/off
- **Supported Clients**:
  - **Torrent**: qBittorrent, Deluge, Transmission, rTorrent, uTorrent, Vuze, Hadouken, Flood
  - **Usenet**: SABnzbd, NZBGet, NZBVortex, Pneumatic, Download Station (Synology)

#### Download Client Configuration (Per Client)
- **Name**: Client display name
- **Enable**: Enable this client
- **Host**: Client hostname/IP
- **Port**: Client port
- **URL Base**: Base URL if behind reverse proxy
- **Username**: Client username
- **Password**: Client password
- **Category**: Category to assign downloads
- **Post-Import Category**: Category after successful import
- **Recent Priority**: Priority for recently aired movies
- **Older Priority**: Priority for older movies
- **Initial State**: Start downloads paused or not (torrents)
- **Sequential Order**: Download files in order (torrents)
- **First and Last First**: Prioritize first and last pieces (torrents)
- **Remove Completed**: Remove from client after import
- **Tags**: Apply client to specific movies only

#### Download Client Options (Global)
- **Download Client Working Folders**: Temporary download folder patterns
- **Enable Completed Download Handling**: Process completed downloads
- **Check For Finished Download Interval**: Check interval in minutes

#### Failed Download Handling
- **Auto Redownload Failed**: Automatically retry failed downloads
- **Auto Redownload Failed from Interactive Search**: Retry manual searches

#### Remote Path Mappings
- **Host**: Download client host
- **Remote Path**: Path on download client
- **Local Path**: Corresponding local path on Radarr server
- **Purpose**: Map paths when Radarr and download client on different systems

### 7. Import Lists

**Purpose**: Automatically add movies from external sources.

#### Import List Management
- **Add/Edit/Delete Import Lists**: Manage list sources
- **Supported Lists**:
  - IMDb Lists
  - Trakt Lists
  - TMDb Lists
  - RadarrList (other Radarr instances)
  - CouchPotato
  - Plex Watchlist
  - StevenLu Custom

#### Import List Configuration (Per List)
- **Name**: List display name
- **Enable Automatic Add**: Automatically add movies from list
- **Enable Automatic Search**: Search for movies after adding
- **Monitor**: Monitor status for added movies
  - Movie Only
  - Movie and Collection
  - None
- **Minimum Availability**: When movie should be considered available
  - Announced
  - In Cinemas
  - Released
  - PreDB
- **Quality Profile**: Profile to apply to imported movies
- **Root Folder**: Where to save imported movies
- **Tags**: Tags to apply to imported movies
- **List-Specific Settings**: URL, API Key, List IDs, etc.

#### Import List Options (Global)
- **List Update Interval**: How often to check lists (hours)
- **Clean Library Level**: What to do with movies no longer on lists
  - Disabled
  - Log Only
  - Keep and Unmonitor
  - Remove and Keep Files
  - Remove and Delete Files

### 8. Connect (Notifications)

**Purpose**: Configure notifications and integrations with external services.

#### Connection Management
- **Add/Edit/Delete Connections**: Manage notification services
- **Supported Services**:
  - Kodi/XBMC
  - Plex Media Server
  - Emby/Jellyfin
  - Telegram
  - Discord
  - Slack
  - Pushbullet
  - Pushover
  - Gotify
  - Email (SMTP)
  - Webhook
  - Custom Script
  - Twitter
  - Trakt
  - Synology Indexer
  - Prowlarr

#### Connection Configuration (Per Connection)
- **Name**: Connection display name
- **On Grab**: Trigger when movie is grabbed
- **On Download**: Trigger when movie finishes downloading
- **On Upgrade**: Trigger when movie is upgraded
- **On Rename**: Trigger when movie is renamed
- **On Movie Added**: Trigger when movie is added to library
- **On Movie Delete**: Trigger when movie is deleted
- **On Movie File Delete**: Trigger when movie file is deleted
- **On Health Issue**: Trigger on health check issues
- **On Application Update**: Trigger when Radarr updates
- **Tags**: Apply connection to specific movies only
- **Service-Specific Settings**: URLs, API keys, channels, etc.

### 9. Metadata

**Purpose**: Configure metadata file generation for media servers.

#### Metadata Provider Management
- **Add/Edit/Delete Providers**: Manage metadata consumers
- **Supported Consumers**:
  - Kodi (XBMC) .nfo
  - WDTV .xml
  - Roksbox .xml
  - Emby (Legacy) .xml

#### Metadata Configuration (Per Provider)
- **Name**: Provider display name
- **Enable**: Enable this metadata provider
- **Movie Metadata**: Generate movie .nfo/.xml files
- **Movie Metadata URL**: Include URLs in metadata
- **Movie Images**: Download and save movie images
- **Use Movie.nfo**: Use movie.nfo filename (vs title.nfo)
- **Tags**: Apply to specific movies only

#### Metadata Options (Global)
- **Certification Country**: Country for age ratings (e.g., US, GB)

### 10. Tags

**Purpose**: Organize and filter movies with custom labels.

#### Tag Management
- **Add Tags**: Create new tags
- **Edit Tags**: Rename existing tags
- **Delete Tags**: Remove unused tags
- **Tag Usage**: See which movies/lists/clients use each tag

### 11. General

**Purpose**: Core Radarr settings and system configuration.

#### Host
- **Bind Address**: IP address to bind to (* for all)
- **Port Number**: HTTP port (default 7878)
- **URL Base**: Base URL if behind reverse proxy
- **Instance Name**: Name for this Radarr instance
- **Enable SSL**: Enable HTTPS
- **SSL Port**: HTTPS port
- **SSL Cert Path**: Path to SSL certificate
- **SSL Cert Password**: Certificate password
- **Open Browser on Start**: Launch browser when Radarr starts

#### Security
- **Authentication**: Authentication method
  - None
  - Basic (Browser Popup)
  - Forms (Login Page)
  - External (via reverse proxy)
- **Username**: Login username
- **Password**: Login password
- **API Key**: API key for integrations
- **Certificate Validation**: Validation level for HTTPS certificates
  - Enabled
  - Disabled for Local Addresses
  - Disabled

#### Proxy
- **Use Proxy**: Enable proxy for all connections
- **Proxy Type**: HTTP, SOCKS4, or SOCKS5
- **Hostname**: Proxy hostname
- **Port**: Proxy port
- **Username**: Proxy username
- **Password**: Proxy password
- **Ignored Addresses**: Addresses to bypass proxy
- **Bypass Proxy for Local Addresses**: Don't use proxy for local connections

#### Logging
- **Log Level**: Amount of information to log
  - Trace
  - Debug
  - Info
  - Warn
  - Error
  - Fatal
  - Off

#### Analytics
- **Send Anonymous Usage Data**: Help improve Radarr by sharing analytics

#### Updates
- **Branch**: Update branch to follow
  - Master (Stable)
  - Develop (Beta)
  - Nightly (Alpha)
- **Automatic**: Enable automatic updates
- **Mechanism**: How to install updates
  - Built-in
  - Script
  - External
  - Docker
  - Apt
- **Script Path**: Path to update script (if using Script)

#### Backup
- **Folder**: Where to store backups
- **Interval**: Days between automatic backups
- **Retention**: Number of backups to keep

### 12. UI

**Purpose**: Customize Radarr's user interface.

#### Dates
- **Short Date Format**: Format for short dates
- **Long Date Format**: Format for long dates
- **Time Format**: 12-hour or 24-hour time
- **Show Relative Dates**: Show "Today" instead of actual date
- **Week Column Header**: First day of week

#### Style
- **Theme**: Light or Dark theme
- **Enable Color-Impaired Mode**: Color scheme for color blindness

#### Movies
- **Movie Runtime Format**: How to display runtime
  - Hours and Minutes
  - Minutes Only
- **Show Certification**: Display age rating

#### Language
- **UI Language**: Interface language (30+ languages supported)

---

## 📺 SONARR - TV Series Management Configuration

Sonarr is a TV series collection manager that automatically downloads, organizes, and manages TV show files. Below is a complete listing of all configuration sections and their features.

### 1. Media Management

**Purpose**: Controls how Sonarr handles TV episode files, naming, and storage.

#### File Management
- **Auto Unmonitor Previously Downloaded Episodes**: Automatically unmonitor episodes after download
- **Recycle Bin Path**: Path where deleted files are moved
- **Recycle Bin Cleanup Days**: Days to keep files in recycle bin
- **Download Propers and Repacks**: Control handling of proper/repack releases
  - Prefer and Upgrade
  - Do Not Upgrade Automatically
  - Do Not Prefer
- **Create Empty Series Folders**: Create folder structure even without files
- **Delete Empty Folders**: Automatically delete empty folders
- **File Date**: What date to use for file timestamps
  - None
  - Local Air Date
  - UTC Air Date
- **Rescan After Refresh**: When to rescan series folder
  - Always
  - After Manual Refresh
  - Never

#### Permissions (Linux)
- **Set Permissions**: Enable permission setting for Linux
- **chmod Folder**: Folder permission mask
- **chown Group**: Group ownership for folders

#### Importing
- **Episode Title Required**: When episode title is required
  - Always
  - Only for Bulks
  - Never
- **Skip Free Space Check When Importing**: Skip disk space check
- **Minimum Free Space When Importing**: Required free space in MB
- **Use Hardlinks Instead of Copy**: Use hardlinks to save space
- **Import Using Script**: Use custom import script
- **Script Import Path**: Path to import script
- **Import Extra Files**: Import subtitles, NFO files, etc.
- **Extra File Extensions**: Extensions to import
- **Enable MediaInfo**: Scan for technical details
- **User Rejected Extensions**: Extensions to never import

#### Season Pack Upgrades
- **Season Pack Upgrade**: How to handle season pack upgrades
  - Disabled
  - Standard Episode Upgrade
  - Season Pack Upgrade
- **Season Pack Upgrade Threshold**: Percentage threshold (0.0-1.0)

#### Root Folders
- **Add/Remove Root Folders**: Manage TV show storage locations
- **Default Root Folder**: Default location for new series
- **Quality Profile**: Default quality profile
- **Metadata Profile**: Default metadata profile
- **Season Folder**: Create season subfolders
- **Monitor**: Default monitoring option
- **Series Type**: Default series type (Standard, Daily, Anime)

#### Naming
- **Rename Episodes**: Enable automatic renaming
- **Replace Illegal Characters**: Replace invalid filename characters
- **Colon Replacement**: How to replace colons
- **Standard Episode Format**: Template for episode file naming
- **Daily Episode Format**: Template for daily show episodes
- **Anime Episode Format**: Template for anime episodes
- **Series Folder Format**: Template for series folders
- **Season Folder Format**: Template for season folders
- **Specials Folder Format**: Format for specials folder
- **Multi-Episode Style**: How to name multi-episode files
  - Extend
  - Duplicate
  - Repeat
  - Scene
  - Range
  - Prefixed Range
- **Tokens Available**: Variables for naming templates
  - {Series Title}
  - {Series TitleThe}
  - {Series CleanTitle}
  - {Series TitleFirstCharacter}
  - {Series Year}
  - {Series TvdbId}
  - {Series TvMazeId}
  - {Series ImdbId}
  - {Season Number}
  - {Episode Number}
  - {Episode Title}
  - {Episode CleanTitle}
  - {Air Date}
  - {Quality Full}
  - {Quality Title}
  - {MediaInfo VideoCodec}
  - {MediaInfo AudioCodec}
  - {MediaInfo AudioChannels}
  - {Release Group}
  - {Custom Formats}

### 2. Profiles

**Purpose**: Define quality and upgrade preferences for TV series.

#### Quality Profiles
- **Add/Edit/Delete Quality Profiles**: Custom quality preferences
- **Profile Name**: Name for the profile
- **Upgrades Allowed**: Enable automatic upgrades
- **Upgrade Until**: Maximum quality to upgrade to
- **Upgrade Until Custom Format Score**: Score threshold for upgrades
- **Quality Order**: Priority ordering
- **Available Qualities**:
  - Unknown
  - SDTV
  - DVD
  - WEBDL-480p
  - WEBRip-480p
  - Bluray-480p
  - HDTV-720p
  - HDTV-1080p
  - Raw-HD
  - WEBDL-720p
  - WEBRip-720p
  - Bluray-720p
  - WEBDL-1080p
  - WEBRip-1080p
  - Bluray-1080p
  - Bluray-1080p Remux
  - HDTV-2160p
  - WEBDL-2160p
  - WEBRip-2160p
  - Bluray-2160p
  - Bluray-2160p Remux

#### Language Profiles (for multi-audio)
- **Add/Edit/Delete Language Profiles**: Language preferences
- **Languages**: Preferred audio languages
- **Upgrade Language**: Enable language upgrades

#### Delay Profiles
- **Protocol Delay**: Prefer Usenet or Torrent
- **Usenet Delay**: Minutes to wait for Usenet
- **Torrent Delay**: Minutes to wait for Torrent
- **Bypass if Highest Quality**: Skip delay for best quality
- **Bypass if Above Custom Format Score**: Skip delay if score high enough
- **Tags**: Apply to specific series

### 3. Quality

**Purpose**: Define quality settings and file size limits.

#### Quality Definitions
- **Quality Title**: Name of quality tier
- **Minimum Size**: Minimum file size (MB per minute)
- **Maximum Size**: Maximum file size (MB per minute)
- **Preferred Size**: Preferred file size (MB per minute)
- **Available for all qualities listed in Profiles section**

#### Reset Quality Definitions
- **Reset to Defaults**: Restore original settings

### 4. Custom Formats

**Purpose**: Advanced release filtering and scoring.

#### Custom Format Management
- **Add/Edit/Delete Custom Formats**: Create custom filters
- **Format Name**: Name for the format
- **Include Custom Format when Renaming**: Include in filename
- **Specifications**: Conditions to match
  - Release Title
  - Edition
  - Language
  - Indexer
  - Size
  - Source
  - Resolution
  - Media Info (Video/Audio codec)
  - Quality Modifier
  - Custom Format Tags

#### Custom Format Scoring
- **Format Score**: Points when format matches
- **Profile Integration**: Assign scores per profile
- **Minimum Custom Format Score**: Minimum score for upgrade

### 5. Indexers

**Purpose**: Configure sources for finding TV releases.

#### Indexer Management
- **Add/Edit/Delete Indexers**: Manage indexer connections
- **Enable/Disable Indexers**: Turn indexers on/off
- **Supported Indexers**: Newznab, Torznab, Torrent RSS, and popular presets

#### Indexer Configuration (Per Indexer)
- **Name**: Indexer display name
- **Enable RSS**: Enable RSS sync
- **Enable Automatic Search**: Enable for automatic searches
- **Enable Interactive Search**: Enable for manual searches
- **URL**: Indexer URL
- **API Path**: API endpoint
- **API Key**: Authentication key
- **Categories**: Search categories
- **Anime Categories**: Categories for anime
- **Additional Parameters**: Extra search parameters
- **Minimum Seeders**: Minimum seeders for torrents
- **Seed Ratio**: Required seed ratio
- **Seed Time**: Required seed time in minutes
- **Season Pack Seed Time**: Seed time for season packs
- **Tags**: Apply to specific series

#### Indexer Options (Global)
- **Minimum Age**: Minimum release age in minutes
- **Retention**: Maximum release age in days (Usenet)
- **Maximum Size**: Maximum release size in MB
- **RSS Sync Interval**: RSS sync frequency in minutes

### 6. Download Clients

**Purpose**: Configure torrent/usenet clients for downloading.

#### Download Client Management
- **Add/Edit/Delete Download Clients**: Manage clients
- **Enable/Disable Clients**: Turn clients on/off
- **Supported Clients**:
  - **Torrent**: qBittorrent, Deluge, Transmission, rTorrent, uTorrent, Vuze, Hadouken, Flood
  - **Usenet**: SABnzbd, NZBGet, NZBVortex, Pneumatic, Download Station

#### Download Client Configuration (Per Client)
- **Name**: Client display name
- **Enable**: Enable this client
- **Host**: Client hostname/IP
- **Port**: Client port
- **URL Base**: Base URL
- **Username**: Client username
- **Password**: Client password
- **Category**: Download category
- **Post-Import Category**: Category after import
- **Recent Priority**: Priority for recent episodes
- **Older Priority**: Priority for older episodes
- **Season Pack Priority**: Priority for season packs
- **Initial State**: Start paused or not (torrents)
- **Sequential Order**: Download in order (torrents)
- **First and Last First**: Prioritize endpoints (torrents)
- **Remove Completed**: Remove after import
- **Tags**: Apply to specific series

#### Download Client Options (Global)
- **Download Client Working Folders**: Temporary folder patterns
- **Enable Completed Download Handling**: Process completed downloads
- **Check For Finished Download Interval**: Check interval in minutes

#### Failed Download Handling
- **Auto Redownload Failed**: Retry failed downloads
- **Auto Redownload Failed from Interactive Search**: Retry manual searches

#### Remote Path Mappings
- **Host**: Download client host
- **Remote Path**: Path on client
- **Local Path**: Corresponding local path
- **Purpose**: Map paths between different systems

### 7. Import Lists

**Purpose**: Automatically add series from external sources.

#### Import List Management
- **Add/Edit/Delete Import Lists**: Manage list sources
- **Supported Lists**:
  - Trakt Lists (Popular, Trending, User Lists)
  - SonarrList (other Sonarr instances)
  - Plex Watchlist
  - AniList
  - Simkl
  - IMDb Lists

#### Import List Configuration (Per List)
- **Name**: List display name
- **Enable Automatic Add**: Auto-add series from list
- **Enable Automatic Search**: Search after adding
- **Monitor**: Monitor setting for added series
  - All Episodes
  - Future Episodes
  - Missing Episodes
  - Existing Episodes
  - First Season
  - Latest Season
  - Pilot Episode
  - None
- **Season Folder**: Create season folders
- **Quality Profile**: Profile for imported series
- **Language Profile**: Language profile for imported series
- **Series Type**: Type for imported series
  - Standard
  - Daily
  - Anime
- **Root Folder**: Where to save imported series
- **Tags**: Tags for imported series
- **List-Specific Settings**: URLs, API keys, list IDs

#### Import List Options (Global)
- **List Update Interval**: Check frequency in hours
- **Clean Library Level**: Action for removed series
  - Disabled
  - Log Only
  - Keep and Unmonitor
  - Remove and Keep Files
  - Remove and Delete Files

### 8. Connect (Notifications)

**Purpose**: Configure notifications and external integrations.

#### Connection Management
- **Add/Edit/Delete Connections**: Manage notification services
- **Supported Services**:
  - Kodi
  - Plex Media Server
  - Emby/Jellyfin
  - Telegram
  - Discord
  - Slack
  - Pushbullet
  - Pushover
  - Gotify
  - Email (SMTP)
  - Webhook
  - Custom Script
  - Twitter
  - Trakt
  - Synology Indexer
  - Prowlarr

#### Connection Configuration (Per Connection)
- **Name**: Connection display name
- **On Grab**: Trigger on episode grab
- **On Import**: Trigger on episode import
- **On Upgrade**: Trigger on episode upgrade
- **On Rename**: Trigger on episode rename
- **On Series Add**: Trigger when series added
- **On Series Delete**: Trigger when series deleted
- **On Episode File Delete**: Trigger on file deletion
- **On Health Issue**: Trigger on health issues
- **On Application Update**: Trigger on Sonarr update
- **Tags**: Apply to specific series
- **Service-Specific Settings**: URLs, tokens, channels

### 9. Metadata

**Purpose**: Generate metadata files for media servers.

#### Metadata Provider Management
- **Add/Edit/Delete Providers**: Manage metadata consumers
- **Supported Consumers**:
  - Kodi (XBMC) .nfo
  - WDTV .xml
  - Roksbox .xml
  - Emby (Legacy) .xml

#### Metadata Configuration (Per Provider)
- **Name**: Provider display name
- **Enable**: Enable this provider
- **Series Metadata**: Generate series metadata files
- **Series Metadata URL**: Include URLs
- **Episode Metadata**: Generate episode metadata
- **Series Images**: Download series images
- **Season Images**: Download season posters
- **Episode Images**: Download episode thumbnails
- **Tags**: Apply to specific series

### 10. Metadata Source

**Purpose**: Configure TV metadata providers (unique to Sonarr).

#### TheTVDB Settings
- **API User**: TheTVDB API username
- **API Key**: TheTVDB API key

#### Preferences
- **Prefer TheTVDB Numbering**: Use TVDB episode numbers over network
- **Specials Folder Format**: Where to place specials

### 11. Tags

**Purpose**: Organize and filter series with custom labels.

#### Tag Management
- **Add Tags**: Create new tags
- **Edit Tags**: Rename tags
- **Delete Tags**: Remove unused tags
- **Tag Usage**: See which series/lists/clients use tags

### 12. General

**Purpose**: Core Sonarr settings and system configuration.

#### Host
- **Bind Address**: IP address to bind to
- **Port Number**: HTTP port (default 8989)
- **URL Base**: Base URL for reverse proxy
- **Instance Name**: Name for this instance
- **Enable SSL**: Enable HTTPS
- **SSL Port**: HTTPS port
- **SSL Cert Path**: SSL certificate path
- **SSL Cert Password**: Certificate password
- **Open Browser on Start**: Launch browser on start

#### Security
- **Authentication**: Authentication method
  - None
  - Basic (Browser Popup)
  - Forms (Login Page)
  - External (via reverse proxy)
- **Username**: Login username
- **Password**: Login password
- **API Key**: API key for integrations
- **Certificate Validation**: HTTPS certificate validation
  - Enabled
  - Disabled for Local Addresses
  - Disabled

#### Proxy
- **Use Proxy**: Enable proxy for connections
- **Proxy Type**: HTTP, SOCKS4, or SOCKS5
- **Hostname**: Proxy hostname
- **Port**: Proxy port
- **Username**: Proxy username
- **Password**: Proxy password
- **Ignored Addresses**: Bypass proxy for these
- **Bypass Proxy for Local Addresses**: Skip proxy locally

#### Logging
- **Log Level**: Logging detail level
  - Trace
  - Debug
  - Info
  - Warn
  - Error
  - Fatal
  - Off

#### Analytics
- **Send Anonymous Usage Data**: Share analytics to improve Sonarr

#### Updates
- **Branch**: Update branch
  - Main (Stable)
  - Develop (Beta)
  - Nightly (Alpha)
- **Automatic**: Enable automatic updates
- **Mechanism**: Update installation method
  - Built-in
  - Script
  - External
  - Docker
  - Apt
- **Script Path**: Path to update script

#### Backup
- **Folder**: Backup storage location
- **Interval**: Days between backups
- **Retention**: Number of backups to keep

### 13. UI

**Purpose**: Customize Sonarr's user interface.

#### Dates
- **Short Date Format**: Format for short dates
- **Long Date Format**: Format for long dates
- **Time Format**: 12-hour or 24-hour
- **Show Relative Dates**: Show "Today" vs actual date
- **Week Column Header**: First day of week

#### Style
- **Theme**: Light or Dark theme
- **Enable Color-Impaired Mode**: Accessibility colors

#### Series
- **Season Folder Format**: Display format for seasons
- **Show Certification**: Display age ratings

#### Language
- **UI Language**: Interface language (30+ languages)

---

## 🔄 Key Differences Between Radarr and Sonarr

While Radarr and Sonarr share similar architectures and many common features, here are the main differences:

### Radarr-Specific Features
1. **Movie-specific metadata**: TMDb IDs, IMDb IDs, cinema release dates
2. **Availability options**: Announced, In Cinemas, Released, PreDB
3. **Single-file focus**: Movies are typically single files
4. **Collection management**: Movie collections (e.g., Marvel Cinematic Universe)
5. **No season/episode handling**: Flat movie structure

### Sonarr-Specific Features
1. **TV-specific metadata**: TVDB IDs, TVMaze IDs, air dates
2. **Episode title requirements**: Configuration for when episode titles are needed
3. **Season pack handling**: Special handling for season packs
4. **Series type options**: Standard, Daily, Anime
5. **Season folder structure**: Nested folder organization
6. **Metadata Source settings**: TheTVDB configuration
7. **Multi-episode naming**: Various multi-episode file naming styles
8. **Specials handling**: Special episodes folder management

### Common Features
- Quality profiles and custom formats
- Indexer management
- Download client integration
- Import lists
- Notifications (Connect)
- Tags system
- General settings (Host, Security, Proxy, Logging, Updates, Backup)
- UI customization

---

## 💡 Recommendations for MediaCheky Implementation

Based on this comprehensive feature analysis, here are recommendations for implementing Radarr and Sonarr configuration in MediaCheky:

### Phase 1: Essential Features (MVP)
1. **Basic Service Enable/Disable**: On/off toggle
2. **Port Configuration**: External port mapping
3. **Path Configuration**: Root folders and config paths
4. **Authentication**: API key management
5. **Network Configuration**: Docker network integration

### Phase 2: Quality & Monitoring
1. **Quality Profiles**: Basic quality profile selection
2. **Monitoring Options**: What to monitor by default
3. **Download Client Selection**: Link to configured clients

### Phase 3: Advanced Configuration
1. **Indexer Management**: Configure indexers within MediaCheky
2. **Import Lists**: Manage lists from MediaCheky interface
3. **Naming Templates**: Custom naming schemes
4. **Custom Formats**: Advanced release filtering

### Phase 4: Full Integration
1. **Real-time Statistics**: Show active downloads, queue
2. **Health Monitoring**: Display service health checks
3. **Quick Actions**: Search, add movies/series from MediaCheky
4. **Unified Dashboard**: Combined view of all *arr services

### UI Organization Suggestion
Group settings into collapsible sections:
- **Basic Settings** (Port, Paths, Enable/Disable)
- **Quality & Profiles** (Quality profiles, upgrade settings)
- **Integration** (API keys, download clients, indexers)
- **Advanced** (Naming, custom formats, permissions)
- **Notifications** (Connect settings)

---

## 📚 Additional Resources

- **Radarr Wiki**: https://wiki.servarr.com/radarr
- **Sonarr Wiki**: https://wiki.servarr.com/sonarr
- **Radarr GitHub**: https://github.com/Radarr/Radarr
- **Sonarr GitHub**: https://github.com/Sonarr/Sonarr
- **Radarr API Docs**: https://radarr.video/docs/api/
- **Sonarr API Docs**: https://sonarr.tv/docs/api/

---

## 📝 Document Metadata

- **Created**: 2025-11-15
- **Purpose**: Document Radarr and Sonarr features for MediaCheky development
- **Sources**: Official Radarr and Sonarr repositories (commit: 2025-11-15)
- **Maintainer**: MediaCheky Development Team
- **Last Updated**: 2025-11-15

---

**Note**: This documentation reflects the current state of Radarr and Sonarr as of November 2025. Features may change in future versions. Always consult official documentation for the most up-to-date information.
