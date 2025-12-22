#!/bin/bash

set -e

echo "=== Testing Jellyfin Reset and Defaults Copy ==="
echo ""

# 1. Reset
echo "[1/4] Executing reset..."
curl -s -X POST http://localhost/api/services/jellyfin/reset | jq . || echo "Reset response: (no JSON)"
sleep 5

# 2. Start
echo ""
echo "[2/4] Executing start..."
curl -s -X POST http://localhost/api/services/jellyfin/start | jq . || echo "Start response: (no JSON)"

# 3. Wait for initialization
echo ""
echo "[3/4] Waiting 60 seconds for Jellyfin to initialize..."
sleep 60

# 4. Verify system.xml
echo ""
echo "[4/4] Verifying system.xml values..."
SYSTEM_XML="/home/user/projects/mediacheky/volumes/mediacheky-data/services-volumes/jellyfin/system.xml"

if [ -f "$SYSTEM_XML" ]; then
    echo "✓ system.xml found at $SYSTEM_XML"
    echo ""
    echo "Key values:"
    grep -o "IsStartupWizardCompleted>[^<]*" "$SYSTEM_XML" || echo "  IsStartupWizardCompleted: NOT FOUND"
    grep -o "PreferredMetadataLanguage>[^<]*" "$SYSTEM_XML" || echo "  PreferredMetadataLanguage: NOT FOUND"
    grep -o "MetadataCountryCode>[^<]*" "$SYSTEM_XML" || echo "  MetadataCountryCode: NOT FOUND"
    echo ""
else
    echo "✗ system.xml NOT FOUND at $SYSTEM_XML"
fi

# 5. Compare directories
echo ""
echo "[5/5] Checking if subdirectories were copied..."
for dir in data cache log .aspnet .cache .jellyfin-config; do
    if [ -d "/home/user/projects/mediacheky/volumes/mediacheky-data/services-volumes/jellyfin/$dir" ]; then
        echo "✓ $dir exists"
    else
        echo "✗ $dir MISSING"
    fi
done

echo ""
echo "=== Test Complete ==="
