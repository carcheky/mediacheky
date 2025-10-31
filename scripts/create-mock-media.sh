#!/bin/bash

# Script para crear una biblioteca de medios simulada para desarrollo/testing
# Crea archivos vacíos que simulan películas y series
# Incluye:
#   - downloads/complete/ - Archivos descargados por qBittorrent
#   - library/ - Archivos organizados (algunos hardlinkeados desde downloads)
#   - Archivos huérfanos en downloads (sin hardlink en library)

set -e

# Directorios principales
MEDIA_BASE="./volumes/media-library"
DOWNLOADS_DIR="$MEDIA_BASE/downloads"
DOWNLOADS_INCOMPLETE_DIR="$DOWNLOADS_DIR/incomplete"
DOWNLOADS_MOVIES_DIR="$DOWNLOADS_DIR/movies"
DOWNLOADS_TV_DIR="$DOWNLOADS_DIR/tv"
LIBRARY_DIR="$MEDIA_BASE/library"
MOVIES_DIR="$LIBRARY_DIR/movies"
TVSHOWS_DIR="$LIBRARY_DIR/tv"

echo "🎬 Creando biblioteca de medios simulada..."
echo ""

# Crear directorios si no existen (no borramos nada, solo sobreescribimos)
mkdir -p "$DOWNLOADS_INCOMPLETE_DIR"
mkdir -p "$DOWNLOADS_MOVIES_DIR"
mkdir -p "$DOWNLOADS_TV_DIR"
mkdir -p "$MOVIES_DIR"
mkdir -p "$TVSHOWS_DIR"

# Función para crear un archivo sparse de tamaño específico
create_file() {
    local filepath="$1"
    local size_mb="$2"
    
    # Crear el directorio padre si no existe
    mkdir -p "$(dirname "$filepath")"
    
    # Crear archivo sparse (instantáneo, no ocupa espacio real en disco)
    truncate -s "${size_mb}M" "$filepath"
}

# Función para crear un hardlink
create_hardlink() {
    local source="$1"
    local target="$2"
    
    # Crear el directorio padre del target si no existe
    mkdir -p "$(dirname "$target")"
    
    # Si el target ya existe, eliminarlo primero
    if [ -e "$target" ]; then
        rm -f "$target"
    fi
    
    # Crear hardlink
    ln "$source" "$target"
}

echo ""
echo "🎬 Creando películas en downloads/movies/ (categoría qBittorrent)..."

# Archivos de películas en downloads/movies - cada película en su propia carpeta
# Formato: "nombre_carpeta|nombre_archivo.mkv|tamaño_mb"
declare -a downloaded_movies=(
    "The.Matrix.1999.1080p.BluRay.x264-GROUP|The.Matrix.1999.1080p.BluRay.x264-GROUP.mkv|2048"
    "Inception.2010.1080p.BluRay.x264-SPARKS|Inception.2010.1080p.BluRay.x264-SPARKS.mkv|2560"
    "The.Shawshank.Redemption.1994.1080p.BluRay.x264-YIFY|The.Shawshank.Redemption.1994.1080p.BluRay.x264-YIFY.mkv|1920"
    "Pulp.Fiction.1994.1080p.BluRay.x264-CtrlHD|Pulp.Fiction.1994.1080p.BluRay.x264-CtrlHD.mkv|2304"
    "The.Dark.Knight.2008.1080p.BluRay.x264-SECTOR7|The.Dark.Knight.2008.1080p.BluRay.x264-SECTOR7.mkv|2688"
    "Forrest.Gump.1994.1080p.BluRay.x264-YIFY|Forrest.Gump.1994.1080p.BluRay.x264-YIFY.mkv|2176"
    "Fight.Club.1999.1080p.BluRay.x264-LEVERAGE|Fight.Club.1999.1080p.BluRay.x264-LEVERAGE.mkv|2432"
    "Goodfellas.1990.1080p.BluRay.x264-CiNEFiLE|Goodfellas.1990.1080p.BluRay.x264-CiNEFiLE.mkv|2240"
    "The.Godfather.1972.1080p.BluRay.x264-SiNNERS|The.Godfather.1972.1080p.BluRay.x264-SiNNERS.mkv|2816"
    "The.Lord.of.the.Rings.The.Fellowship.of.the.Ring.2001.EXTENDED.1080p.BluRay.x264-SECTOR7|The.Lord.of.the.Rings.The.Fellowship.of.the.Ring.2001.EXTENDED.1080p.BluRay.x264-SECTOR7.mkv|3072"
    # Películas huérfanas (sin hardlink en library) - Simulan descargas no procesadas
    "Interstellar.2014.1080p.BluRay.x264-SPARKS|Interstellar.2014.1080p.BluRay.x264-SPARKS.mkv|2944"
    "The.Prestige.2006.1080p.BluRay.x264-SECTOR7|The.Prestige.2006.1080p.BluRay.x264-SECTOR7.mkv|2112"
    "Memento.2000.1080p.BluRay.x264-CiNEFiLE|Memento.2000.1080p.BluRay.x264-CiNEFiLE.mkv|1856"
)

for movie_info in "${downloaded_movies[@]}"; do
    IFS='|' read -r folder_name file_name size <<< "$movie_info"
    filepath="$DOWNLOADS_MOVIES_DIR/$folder_name/$file_name"
    echo "  ✓ Descarga: movies/$folder_name/$file_name ($size MB)"
    create_file "$filepath" $size
done

echo ""
echo "📽️  Creando películas en library/ (con hardlinks)..."

# Mapeo de archivos descargados a su ubicación organizada en library
# Formato: "carpeta_downloads|nombre_archivo|nombre_carpeta_library|nombre_en_library"
declare -a movie_mappings=(
    "The.Matrix.1999.1080p.BluRay.x264-GROUP|The.Matrix.1999.1080p.BluRay.x264-GROUP.mkv|The Matrix (1999)|The Matrix (1999) - 1080p.mkv"
    "Inception.2010.1080p.BluRay.x264-SPARKS|Inception.2010.1080p.BluRay.x264-SPARKS.mkv|Inception (2010)|Inception (2010) - 1080p.mkv"
    "The.Shawshank.Redemption.1994.1080p.BluRay.x264-YIFY|The.Shawshank.Redemption.1994.1080p.BluRay.x264-YIFY.mkv|The Shawshank Redemption (1994)|The Shawshank Redemption (1994) - 1080p.mkv"
    "Pulp.Fiction.1994.1080p.BluRay.x264-CtrlHD|Pulp.Fiction.1994.1080p.BluRay.x264-CtrlHD.mkv|Pulp Fiction (1994)|Pulp Fiction (1994) - 1080p.mkv"
    "The.Dark.Knight.2008.1080p.BluRay.x264-SECTOR7|The.Dark.Knight.2008.1080p.BluRay.x264-SECTOR7.mkv|The Dark Knight (2008)|The Dark Knight (2008) - 1080p.mkv"
    "Forrest.Gump.1994.1080p.BluRay.x264-YIFY|Forrest.Gump.1994.1080p.BluRay.x264-YIFY.mkv|Forrest Gump (1994)|Forrest Gump (1994) - 1080p.mkv"
    "Fight.Club.1999.1080p.BluRay.x264-LEVERAGE|Fight.Club.1999.1080p.BluRay.x264-LEVERAGE.mkv|Fight Club (1999)|Fight Club (1999) - 1080p.mkv"
    "Goodfellas.1990.1080p.BluRay.x264-CiNEFiLE|Goodfellas.1990.1080p.BluRay.x264-CiNEFiLE.mkv|Goodfellas (1990)|Goodfellas (1990) - 1080p.mkv"
)

for mapping in "${movie_mappings[@]}"; do
    IFS='|' read -r download_folder download_file folder_name library_name <<< "$mapping"
    
    source_file="$DOWNLOADS_MOVIES_DIR/$download_folder/$download_file"
    target_file="$MOVIES_DIR/$folder_name/$library_name"
    
    echo "  🔗 Hardlink: $folder_name"
    create_hardlink "$source_file" "$target_file"
done

# Algunas películas solo en library (sin hardlink) - Simulan imports antiguos
declare -A library_only_movies=(
    ["The Godfather (1972)/The Godfather (1972) - 1080p.mkv"]=2816
    ["The.Lord.of.the.Rings.The.Fellowship.of.the.Ring (2001)/The Fellowship of the Ring (2001) - 1080p.mkv"]=3072
)

echo ""
echo "📁 Creando películas solo en library (sin hardlink)..."

for movie in "${!library_only_movies[@]}"; do
    size=${library_only_movies[$movie]}
    filepath="$MOVIES_DIR/$movie"
    echo "  ✓ Solo library: $movie ($size MB)"
    create_file "$filepath" $size
done

echo ""
echo "📺 Creando series en downloads/tv/ (categoría qBittorrent)..."

# Series descargadas (nombres de carpetas de torrent típicos)
declare -A downloaded_series=(
    ["Breaking.Bad.S01.1080p.BluRay.x264-ROVERS"]=5
    ["Breaking.Bad.S02.1080p.BluRay.x264-ROVERS"]=5
    ["Game.of.Thrones.S01.1080p.BluRay.x264-DEMAND"]=10
    ["The.Office.US.S01.1080p.WEB-DL.DD5.1.H264-NTb"]=6
    ["Friends.S01.1080p.BluRay.x264-PSYCHD"]=24
    ["Stranger.Things.S01.1080p.NF.WEB-DL.DD5.1.x264-NTb"]=8
    # Series huérfanas (carpetas completas sin procesar)
    ["The.Wire.S01.1080p.BluRay.x264-PSYCHD"]=13
    ["Better.Call.Saul.S01.1080p.BluRay.x264-ROVERS"]=10
)

for series_folder in "${!downloaded_series[@]}"; do
    num_episodes=${downloaded_series[$series_folder]}
    series_path="$DOWNLOADS_TV_DIR/$series_folder"
    mkdir -p "$series_path"
    
    echo "  📁 Serie: tv/$series_folder ($num_episodes episodios)"
    
    for episode in $(seq 1 $num_episodes); do
        # Extraer nombre de la serie y temporada del nombre de carpeta
        if [[ $series_folder =~ ^(.+)\.S([0-9]+)\. ]]; then
            series_name="${BASH_REMATCH[1]}"
            season_num="${BASH_REMATCH[2]}"
            
            episode_file="$series_path/${series_name}.S${season_num}E$(printf %02d $episode).1080p.mkv"
            size=$((RANDOM % 400 + 800))
            create_file "$episode_file" $size
        fi
    done
done

echo ""
echo "📺 Creando series en library/ (con hardlinks)..."

# Mapeo de series descargadas a library
# Solo algunas series se procesan a library, otras quedan huérfanas
declare -a series_mappings=(
    "Breaking.Bad.S01.1080p.BluRay.x264-ROVERS|Breaking Bad|1"
    "Breaking.Bad.S02.1080p.BluRay.x264-ROVERS|Breaking Bad|2"
    "Game.of.Thrones.S01.1080p.BluRay.x264-DEMAND|Game of Thrones|1"
    "The.Office.US.S01.1080p.WEB-DL.DD5.1.H264-NTb|The Office|1"
    "Friends.S01.1080p.BluRay.x264-PSYCHD|Friends|1"
)

for mapping in "${series_mappings[@]}"; do
    IFS='|' read -r download_folder library_name season_num <<< "$mapping"
    
    source_dir="$DOWNLOADS_TV_DIR/$download_folder"
    target_dir="$TVSHOWS_DIR/$library_name/Season $(printf %02d $season_num)"
    
    mkdir -p "$target_dir"
    
    # Crear hardlinks para todos los episodios de esa temporada
    episode_count=$(find "$source_dir" -type f -name "*.mkv" 2>/dev/null | wc -l)
    echo "  🔗 Hardlink: $library_name - Temporada $season_num ($episode_count episodios)"
    
    for source_file in "$source_dir"/*.mkv; do
        if [ -f "$source_file" ]; then
            # Extraer número de episodio del nombre del archivo
            if [[ $(basename "$source_file") =~ E([0-9]+) ]]; then
                ep_num="${BASH_REMATCH[1]}"
                target_file="$target_dir/${library_name} - S$(printf %02d $season_num)E${ep_num} - 1080p.mkv"
                create_hardlink "$source_file" "$target_file"
            fi
        fi
    done
done

# Series adicionales solo en library (sin hardlink) - Temporadas adicionales
echo ""
echo "📁 Creando series adicionales solo en library..."

declare -a library_only_series=(
    "Stranger Things|1|8"
    "Stranger Things|2|9"
)

for series_info in "${library_only_series[@]}"; do
    IFS='|' read -r show_name season_num num_episodes <<< "$series_info"
    
    season_dir="$TVSHOWS_DIR/$show_name/Season $(printf %02d $season_num)"
    mkdir -p "$season_dir"
    
    echo "  ✓ $show_name - Temporada $season_num ($num_episodes episodios)"
    
    for episode in $(seq 1 $num_episodes); do
        episode_file="$season_dir/${show_name} - S$(printf %02d $season_num)E$(printf %02d $episode) - 1080p.mkv"
        size=$((RANDOM % 400 + 800))
        create_file "$episode_file" $size
    done
done

echo ""
echo "📊 Estadísticas de la biblioteca:"
echo ""

# Calcular tamaños y archivos
downloads_count=$(find "$DOWNLOADS_DIR" -type f -name "*.mkv" 2>/dev/null | wc -l)
downloads_size=$(du -sh "$DOWNLOADS_DIR" 2>/dev/null | cut -f1)

movies_count=$(find "$MOVIES_DIR" -type f -name "*.mkv" 2>/dev/null | wc -l)
movies_size=$(du -sh "$MOVIES_DIR" 2>/dev/null | cut -f1)

tvshows_count=$(find "$TVSHOWS_DIR" -type f -name "*.mkv" 2>/dev/null | wc -l)
tvshows_size=$(du -sh "$TVSHOWS_DIR" 2>/dev/null | cut -f1)

total_size=$(du -sh "$MEDIA_BASE" 2>/dev/null | cut -f1)

# Contar archivos con múltiples hardlinks (más simple y rápido)
hardlinked_count=$(find "$DOWNLOADS_DIR" -type f -name "*.mkv" -links +1 2>/dev/null | wc -l)
orphaned_downloads=$((downloads_count - hardlinked_count))

echo "  📥 Downloads (downloads/complete/):"
echo "     - Total: $downloads_count archivos ($downloads_size)"
echo "     - Con hardlink en library: $hardlinked_count"
echo "     - Huérfanos (sin hardlink): $orphaned_downloads"
echo ""
echo "  📁 Library (library/):"
echo "     - Películas: $movies_count archivos ($movies_size)"
echo "     - Series: $tvshows_count episodios ($tvshows_size)"
echo ""
echo "  💾 Total: $total_size"

echo ""
echo "✅ Biblioteca de medios simulada creada exitosamente!"
echo ""
echo "📍 Estructura creada:"
echo "   - $DOWNLOADS_MOVIES_DIR/     (categoría movies - películas completas)"
echo "   - $DOWNLOADS_TV_DIR/         (categoría tv - series completas)"
echo "   - $DOWNLOADS_INCOMPLETE_DIR/ (descargas incompletas - vacío)"
echo "   - $LIBRARY_DIR/movies/       (películas organizadas)"
echo "   - $LIBRARY_DIR/tv/           (series organizadas)"
echo ""
echo "💡 Casos de prueba incluidos:"
echo "   ✓ Archivos con hardlink (downloads ↔ library)"
echo "   ✓ Archivos huérfanos en downloads (no procesados)"
echo "   ✓ Archivos solo en library (imports antiguos)"
echo ""

# Leer configuración para obtener URLs y API keys
CONFIG_FILE="./config/config.yaml"

if [ -f "$CONFIG_FILE" ]; then
    echo "� Iniciando escaneo de bibliotecas en servicios..."
    echo ""
    
    # Función para extraer valores del YAML (simple grep)
    get_config_value() {
        local key="$1"
        grep "^\s*${key}:" "$CONFIG_FILE" | sed 's/.*: *"\?\([^"]*\)"\?.*/\1/' | head -1
    }
    
    # Obtener configuración de Radarr
    RADARR_URL=$(get_config_value "url" | sed -n '1p')
    RADARR_API_KEY=$(get_config_value "api_key" | sed -n '1p')
    
    # Obtener configuración de Sonarr (segunda ocurrencia)
    SONARR_URL=$(get_config_value "url" | sed -n '2p')
    SONARR_API_KEY=$(get_config_value "api_key" | sed -n '2p')
    
    # Obtener configuración de Jellyfin (tercera ocurrencia)
    JELLYFIN_URL=$(get_config_value "url" | sed -n '3p')
    JELLYFIN_API_KEY=$(get_config_value "api_key" | sed -n '3p')
    
    # Escanear Radarr
    if [ -n "$RADARR_URL" ] && [ -n "$RADARR_API_KEY" ]; then
        echo "📽️  Escaneando biblioteca de Radarr..."
        RADARR_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
            -H "X-Api-Key: $RADARR_API_KEY" \
            -H "Content-Type: application/json" \
            "$RADARR_URL/api/v3/command" \
            -d '{"name": "RefreshMovie"}' 2>/dev/null)
        
        HTTP_CODE=$(echo "$RADARR_RESPONSE" | tail -n1)
        
        if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "200" ]; then
            echo "  ✓ Escaneo de Radarr iniciado exitosamente (HTTP $HTTP_CODE)"
        else
            echo "  ⚠️  No se pudo conectar con Radarr (HTTP $HTTP_CODE)"
        fi
    else
        echo "  ⚠️  Radarr no configurado en $CONFIG_FILE"
    fi
    
    # Escanear Sonarr
    if [ -n "$SONARR_URL" ] && [ -n "$SONARR_API_KEY" ]; then
        echo "📺 Escaneando biblioteca de Sonarr..."
        SONARR_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
            -H "X-Api-Key: $SONARR_API_KEY" \
            -H "Content-Type: application/json" \
            "$SONARR_URL/api/v3/command" \
            -d '{"name": "RefreshSeries"}' 2>/dev/null)
        
        HTTP_CODE=$(echo "$SONARR_RESPONSE" | tail -n1)
        
        if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "200" ]; then
            echo "  ✓ Escaneo de Sonarr iniciado exitosamente (HTTP $HTTP_CODE)"
        else
            echo "  ⚠️  No se pudo conectar con Sonarr (HTTP $HTTP_CODE)"
        fi
    else
        echo "  ⚠️  Sonarr no configurado en $CONFIG_FILE"
    fi
    
    # Escanear Jellyfin
    if [ -n "$JELLYFIN_URL" ] && [ -n "$JELLYFIN_API_KEY" ]; then
        echo "🟣 Escaneando biblioteca de Jellyfin..."
        
        # Primero verificar conectividad
        if curl -s -f -H "X-Emby-Token: $JELLYFIN_API_KEY" "$JELLYFIN_URL/System/Info" > /dev/null 2>&1; then
            # Iniciar escaneo
            JELLYFIN_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
                -H "X-Emby-Token: $JELLYFIN_API_KEY" \
                "$JELLYFIN_URL/Library/Refresh" 2>/dev/null)
            
            HTTP_CODE=$(echo "$JELLYFIN_RESPONSE" | tail -n1)
            
            if [ "$HTTP_CODE" = "204" ] || [ "$HTTP_CODE" = "200" ]; then
                echo "  ✓ Escaneo de Jellyfin iniciado exitosamente (HTTP $HTTP_CODE)"
                
                # Mostrar conteo actual de items
                ITEMS=$(curl -s -H "X-Emby-Token: $JELLYFIN_API_KEY" "$JELLYFIN_URL/Items/Counts" 2>/dev/null)
                if [ -n "$ITEMS" ]; then
                    MOVIE_COUNT=$(echo "$ITEMS" | grep -o '"MovieCount":[0-9]*' | grep -o '[0-9]*' || echo "0")
                    SERIES_COUNT=$(echo "$ITEMS" | grep -o '"SeriesCount":[0-9]*' | grep -o '[0-9]*' || echo "0")
                    EPISODE_COUNT=$(echo "$ITEMS" | grep -o '"EpisodeCount":[0-9]*' | grep -o '[0-9]*' || echo "0")
                    echo "  📊 Contenido actual: $MOVIE_COUNT películas, $SERIES_COUNT series, $EPISODE_COUNT episodios"
                fi
            else
                echo "  ⚠️  Respuesta inesperada de Jellyfin (HTTP $HTTP_CODE)"
            fi
        else
            echo "  ⚠️  No se pudo conectar con Jellyfin (¿está corriendo?)"
            echo "      URL: $JELLYFIN_URL"
        fi
    else
        echo "  ⚠️  Jellyfin no configurado en $CONFIG_FILE"
    fi
    
    echo ""
    echo "⏳ Los escaneos están en proceso. Esto puede tardar unos minutos."
    echo ""
else
    echo "⚠️  Archivo de configuración no encontrado: $CONFIG_FILE"
    echo "   Los servicios no se escanearon automáticamente."
    echo ""
fi

echo "💡 Próximos pasos:"
echo "   1. Espera a que terminen los escaneos (1-2 minutos)"
echo "   2. Verifica que los medios aparezcan en:"
echo "      - Radarr: $RADARR_URL"
echo "      - Sonarr: $SONARR_URL"
echo "      - Jellyfin: $JELLYFIN_URL"
echo "   3. Ejecuta sincronización en KeeperCheky: POST /api/sync"
echo ""
