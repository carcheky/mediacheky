#!/bin/bash

# MediaCheky Port Configuration Test Script
# Tests dynamic vs static compose switching based on HostPort field

OUTPUT_LOG="./logs/test-output.log"
API_BASE_URL="http://localhost:80"  # Proxy port exposed to host

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "========================================" | tee "$OUTPUT_LOG"
echo "MediaCheky Port Configuration Test" | tee -a "$OUTPUT_LOG"
echo "========================================" | tee -a "$OUTPUT_LOG"
echo "" | tee -a "$OUTPUT_LOG"

# SETUP: Enable radarr service first
echo -e "${YELLOW}=== SETUP: Habilitando servicio Radarr ===${NC}" | tee -a "$OUTPUT_LOG"
ENABLE_RESPONSE=$(curl -s -X POST ${API_BASE_URL}/api/services/radarr/enable)

echo "$ENABLE_RESPONSE" | jq . | tee -a "$OUTPUT_LOG"

if echo "$ENABLE_RESPONSE" | jq -e '.success == true' > /dev/null; then
  echo -e "${GREEN}✓ Radarr enabled successfully${NC}" | tee -a "$OUTPUT_LOG"
  
  # Verificar en base de datos
  DB_CHECK=$(docker compose exec -T mediacheky sqlite3 /app/data/mediacheky.db "SELECT enabled FROM services WHERE name = 'radarr';")
  if [ "$DB_CHECK" = "1" ]; then
    echo -e "${GREEN}✓ Radarr verified as enabled in database${NC}" | tee -a "$OUTPUT_LOG"
  else
    echo -e "${RED}✗ Radarr NOT enabled in database (DB value: $DB_CHECK)${NC}" | tee -a "$OUTPUT_LOG"
    exit 1
  fi
else
  echo -e "${RED}✗ Failed to enable Radarr${NC}" | tee -a "$OUTPUT_LOG"
  exit 1
fi

echo "" | tee -a "$OUTPUT_LOG"
echo "Waiting 3 seconds for service to initialize..." | tee -a "$OUTPUT_LOG"
sleep 3
echo "" | tee -a "$OUTPUT_LOG"

# TEST 1: Rellenar puerto (usar dynamic compose)
echo "" | tee -a "$OUTPUT_LOG"
echo "" | tee -a "$OUTPUT_LOG"
echo -e "${YELLOW}=== TEST 1: Rellenar puerto 7878 (con HostPort) ===${NC}" | tee -a "$OUTPUT_LOG"
RESPONSE=$(curl -s -X PUT ${API_BASE_URL}/api/services/radarr/config \
  -H "Content-Type: application/json" \
  -d '{
    "Image": "linuxserver/radarr:latest",
    "ContainerName": "radarr",
    "RestartPolicy": "always",
    "Paths": {
      "Config": "./volumes/radarr/config"
    },
    "HostPort": 7878
  }')

echo "$RESPONSE" | jq . | tee -a "$OUTPUT_LOG"

if echo "$RESPONSE" | jq -e '.success == true' > /dev/null; then
  echo -e "${GREEN}✓ Request successful${NC}" | tee -a "$OUTPUT_LOG"
else
  echo -e "${RED}✗ Request failed${NC}" | tee -a "$OUTPUT_LOG"
  echo "Error:" | tee -a "$OUTPUT_LOG"
  echo "$RESPONSE" | jq -r '.error' | tee -a "$OUTPUT_LOG"
fi

echo "" | tee -a "$OUTPUT_LOG"
echo "Waiting 5 seconds..." | tee -a "$OUTPUT_LOG"
sleep 5

echo "" | tee -a "$OUTPUT_LOG"
echo "Estado de radarr:" | tee -a "$OUTPUT_LOG"
docker ps -a --filter name=radarr --format "table {{.Names}}\t{{.Ports}}\t{{.Status}}" | tee -a "$OUTPUT_LOG"

echo "" | tee -a "$OUTPUT_LOG"
echo "Compose file generado:" | tee -a "$OUTPUT_LOG"
if [ -f "./volumes/mediacheky-data/services/radarr/docker compose.yml" ]; then
  echo -e "${GREEN}✓ Dynamic compose exists${NC}" | tee -a "$OUTPUT_LOG"
  cat ./volumes/mediacheky-data/services/radarr/docker compose.yml | tee -a "$OUTPUT_LOG"
  
  echo "" | tee -a "$OUTPUT_LOG"
  echo "Verificando contenido:" | tee -a "$OUTPUT_LOG"
  
  # Check for version (should not exist)
  if grep -q "^version:" ./volumes/mediacheky-data/services/radarr/docker compose.yml; then
    echo -e "${RED}✗ Tiene 'version:' (obsoleto)${NC}" | tee -a "$OUTPUT_LOG"
  else
    echo -e "${GREEN}✓ Sin 'version:' (correcto)${NC}" | tee -a "$OUTPUT_LOG"
  fi
  
  # Check for <no value>
  if grep -q "<no value>" ./volumes/mediacheky-data/services/radarr/docker compose.yml; then
    echo -e "${RED}✗ Tiene '<no value>' en volumes${NC}" | tee -a "$OUTPUT_LOG"
  else
    echo -e "${GREEN}✓ Sin '<no value>' (correcto)${NC}" | tee -a "$OUTPUT_LOG"
  fi
  
  # Check for correct port
  if grep -q '"7878:7878"' ./volumes/mediacheky-data/services/radarr/docker compose.yml; then
    echo -e "${GREEN}✓ Puerto correcto: 7878:7878${NC}" | tee -a "$OUTPUT_LOG"
  else
    echo -e "${RED}✗ Puerto incorrecto${NC}" | tee -a "$OUTPUT_LOG"
  fi
  
  # Check for correct network
  if grep -q "mediacheky_mediacheky-net" ./volumes/mediacheky-data/services/radarr/docker compose.yml; then
    echo -e "${GREEN}✓ Red correcta: mediacheky_mediacheky-net${NC}" | tee -a "$OUTPUT_LOG"
  else
    echo -e "${RED}✗ Red incorrecta${NC}" | tee -a "$OUTPUT_LOG"
  fi
else
  echo -e "${RED}✗ No dynamic compose (debería existir)${NC}" | tee -a "$OUTPUT_LOG"
fi

# TEST 2: Vaciar puerto (usar static compose)
echo -e "${YELLOW}=== TEST 2: Vaciar puerto (sin HostPort) ===${NC}" | tee -a "$OUTPUT_LOG"
RESPONSE=$(curl -s -X PUT http://localhost:80/api/services/radarr/config \
  -H "Content-Type: application/json" \
  -d '{
    "Image": "linuxserver/radarr:latest",
    "ContainerName": "radarr",
    "RestartPolicy": "always",
    "Paths": {
      "Config": "./volumes/radarr/config"
    }
  }')

echo "$RESPONSE" | jq . | tee -a "$OUTPUT_LOG"

if echo "$RESPONSE" | jq -e '.success == true' > /dev/null; then
  echo -e "${GREEN}✓ Request successful${NC}" | tee -a "$OUTPUT_LOG"
else
  echo -e "${RED}✗ Request failed${NC}" | tee -a "$OUTPUT_LOG"
fi

echo "" | tee -a "$OUTPUT_LOG"
echo "Waiting 5 seconds..." | tee -a "$OUTPUT_LOG"
sleep 5

echo "" | tee -a "$OUTPUT_LOG"
echo "Estado de radarr:" | tee -a "$OUTPUT_LOG"
docker ps -a --filter name=radarr --format "table {{.Names}}\t{{.Ports}}\t{{.Status}}" | tee -a "$OUTPUT_LOG"

echo "" | tee -a "$OUTPUT_LOG"
echo "Compose file generado:" | tee -a "$OUTPUT_LOG"
if [ -f "./volumes/mediacheky-data/services/radarr/docker compose.yml" ]; then
  echo -e "${RED}✗ Dynamic compose exists (debería usar static)${NC}" | tee -a "$OUTPUT_LOG"
  cat ./volumes/mediacheky-data/services/radarr/docker compose.yml | tee -a "$OUTPUT_LOG"
else
  echo -e "${GREEN}✓ No dynamic compose (usando static de services/radarr.yml)${NC}" | tee -a "$OUTPUT_LOG"
fi

# TEST 3: Cambiar RestartPolicy
echo "" | tee -a "$OUTPUT_LOG"
echo "" | tee -a "$OUTPUT_LOG"
echo -e "${YELLOW}=== TEST 3: Cambiar RestartPolicy a 'unless-stopped' ===${NC}" | tee -a "$OUTPUT_LOG"
RESPONSE=$(curl -s -X PUT http://localhost:80/api/services/radarr/config \
  -H "Content-Type: application/json" \
  -d '{
    "Image": "linuxserver/radarr:latest",
    "ContainerName": "radarr",
    "RestartPolicy": "always",
    "Paths": {
      "Config": "./volumes/radarr/config"
    }
  }')

echo "$RESPONSE" | jq . | tee -a "$OUTPUT_LOG"

if echo "$RESPONSE" | jq -e '.success == true' > /dev/null; then
  echo -e "${GREEN}✓ Request successful${NC}" | tee -a "$OUTPUT_LOG"
else
  echo -e "${RED}✗ Request failed${NC}" | tee -a "$OUTPUT_LOG"
  echo "Error:" | tee -a "$OUTPUT_LOG"
  echo "$RESPONSE" | jq -r '.error' | tee -a "$OUTPUT_LOG"
fi

echo "" | tee -a "$OUTPUT_LOG"
echo "Waiting 5 seconds..." | tee -a "$OUTPUT_LOG"
sleep 5

echo "" | tee -a "$OUTPUT_LOG"
echo "Estado de radarr:" | tee -a "$OUTPUT_LOG"
docker ps -a --filter name=radarr --format "table {{.Names}}\t{{.Ports}}\t{{.Status}}" | tee -a "$OUTPUT_LOG"

echo "" | tee -a "$OUTPUT_LOG"
echo "Verificando RestartPolicy:" | tee -a "$OUTPUT_LOG"
if [ -f "./volumes/mediacheky-data/services/radarr/docker compose.yml" ]; then
  # Check restart policy in compose file
  RESTART_POLICY=$(grep "restart:" ./volumes/mediacheky-data/services/radarr/docker compose.yml | awk '{print $2}')
  if [ "$RESTART_POLICY" = "unless-stopped" ]; then
    echo -e "${GREEN}✓ RestartPolicy en compose: unless-stopped${NC}" | tee -a "$OUTPUT_LOG"
  else
    echo -e "${RED}✗ RestartPolicy en compose: $RESTART_POLICY (esperado: unless-stopped)${NC}" | tee -a "$OUTPUT_LOG"
  fi
  
  echo "" | tee -a "$OUTPUT_LOG"
  echo "Inspeccionando contenedor Docker:" | tee -a "$OUTPUT_LOG"
  CONTAINER_RESTART=$(docker inspect radarr --format='{{.HostConfig.RestartPolicy.Name}}' 2>/dev/null)
  if [ "$CONTAINER_RESTART" = "unless-stopped" ]; then
    echo -e "${GREEN}✓ Contenedor con RestartPolicy: $CONTAINER_RESTART${NC}" | tee -a "$OUTPUT_LOG"
  else
    echo -e "${RED}✗ Contenedor con RestartPolicy: $CONTAINER_RESTART (esperado: unless-stopped)${NC}" | tee -a "$OUTPUT_LOG"
  fi
  
  echo "" | tee -a "$OUTPUT_LOG"
  echo "Compose completo:" | tee -a "$OUTPUT_LOG"
  cat ./volumes/mediacheky-data/services/radarr/docker compose.yml | tee -a "$OUTPUT_LOG"
else
  echo -e "${RED}✗ No existe compose dinámico${NC}" | tee -a "$OUTPUT_LOG"
fi

# Template verification
echo "" | tee -a "$OUTPUT_LOG"
echo "" | tee -a "$OUTPUT_LOG"
echo -e "${YELLOW}=== Verificación de Templates ===${NC}" | tee -a "$OUTPUT_LOG"

echo "" | tee -a "$OUTPUT_LOG"
echo "Template en filesystem (primeras 15 líneas):" | tee -a "$OUTPUT_LOG"
head -15 ./templates/radarr.yml | tee -a "$OUTPUT_LOG"

echo "" | tee -a "$OUTPUT_LOG"
echo "Template en DB:" | tee -a "$OUTPUT_LOG"
docker compose exec mediacheky sqlite3 /app/data/mediacheky.db \
  "SELECT name, CASE WHEN length(content) > 200 THEN substr(content, 1, 200) || '...' ELSE content END FROM templates WHERE name='radarr';" \
  2>&1 | tee -a "$OUTPUT_LOG"

# Logs
echo "" | tee -a "$OUTPUT_LOG"
echo "" | tee -a "$OUTPUT_LOG"
echo -e "${YELLOW}=== Logs recientes (últimas 20 líneas con 'radarr') ===${NC}" | tee -a "$OUTPUT_LOG"
docker compose logs mediacheky --tail 50 | grep -i radarr | tail -20 | tee -a "$OUTPUT_LOG"

echo "" | tee -a "$OUTPUT_LOG"
echo "========================================" | tee -a "$OUTPUT_LOG"
echo -e "${GREEN}Test completado${NC}" | tee -a "$OUTPUT_LOG"
echo "Log guardado en: $OUTPUT_LOG" | tee -a "$OUTPUT_LOG"
echo "========================================" | tee -a "$OUTPUT_LOG"

# Open log in VS Code
code "$OUTPUT_LOG"
