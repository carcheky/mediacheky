#!/bin/bash

# Script para crear la rama stable desde develop
# Este script debe ser ejecutado por alguien con permisos de push al repositorio
# La rama stable se creará vacía/inicial y se actualizará cuando la aplicación esté lista para producción

set -e

echo "🚀 Creando rama stable desde develop..."
echo ""

# Colores para output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Verificar que estamos en un repositorio git
if [ ! -d .git ]; then
    echo "❌ Error: No se encuentra el directorio .git"
    echo "   Por favor, ejecuta este script desde la raíz del repositorio"
    exit 1
fi

# Obtener el repositorio remoto
REMOTE=$(git remote get-url origin 2>/dev/null || echo "")
if [ -z "$REMOTE" ]; then
    echo "❌ Error: No se encontró el remoto 'origin'"
    exit 1
fi

echo -e "${BLUE}📍 Repositorio: ${NC}$REMOTE"
echo ""

# Verificar si la rama stable ya existe localmente
if git show-ref --verify --quiet refs/heads/stable; then
    echo -e "${YELLOW}⚠️  La rama 'stable' ya existe localmente${NC}"
    read -p "¿Deseas eliminarla y recrearla? (s/N): " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Ss]$ ]]; then
        git branch -D stable
        echo -e "${GREEN}✅ Rama local 'stable' eliminada${NC}"
    else
        echo "❌ Operación cancelada"
        exit 0
    fi
fi

# Verificar si la rama stable ya existe en el remoto
if git ls-remote --heads origin stable | grep -q stable; then
    echo -e "${YELLOW}⚠️  La rama 'stable' ya existe en el remoto${NC}"
    read -p "¿Deseas continuar de todos modos? (s/N): " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Ss]$ ]]; then
        echo "❌ Operación cancelada"
        exit 0
    fi
fi

# Hacer fetch para asegurarnos de tener la última versión
echo -e "${BLUE}📥 Obteniendo última versión del remoto...${NC}"
git fetch origin

# Verificar que existe la rama develop en el remoto
if ! git ls-remote --heads origin develop | grep -q develop; then
    echo "❌ Error: No se encontró la rama 'develop' en el remoto"
    exit 1
fi

# Checkout a develop
echo -e "${BLUE}🔄 Cambiando a rama develop...${NC}"
git checkout develop 2>/dev/null || git checkout -b develop origin/develop

# Hacer pull de develop
echo -e "${BLUE}📥 Actualizando rama develop...${NC}"
git pull origin develop

# Obtener el commit actual
COMMIT=$(git rev-parse HEAD)
COMMIT_SHORT=$(git rev-parse --short HEAD)
COMMIT_MSG=$(git log -1 --pretty=%B)

echo ""
echo -e "${BLUE}📌 Último commit en develop:${NC}"
echo -e "   ${COMMIT_SHORT} - ${COMMIT_MSG}"
echo ""

# Crear la rama stable
echo -e "${BLUE}🌿 Creando rama stable...${NC}"
git checkout -b stable

# Confirmar antes de hacer push
echo ""
echo -e "${YELLOW}⚠️  A punto de hacer push de la rama stable al remoto${NC}"
echo -e "${YELLOW}ℹ️  Esta rama se usará para releases estables en el futuro${NC}"
read -p "¿Continuar? (S/n): " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Nn]$ ]]; then
    echo -e "${BLUE}📤 Haciendo push de stable...${NC}"
    git push -u origin stable
    
    echo ""
    echo -e "${GREEN}✅ ¡Rama stable creada exitosamente!${NC}"
    echo ""
    echo -e "${BLUE}📋 Resumen:${NC}"
    echo -e "   • Rama stable creada desde develop"
    echo -e "   • Commit base: ${COMMIT_SHORT}"
    echo -e "   • Push realizado a origin/stable"
    echo ""
    echo -e "${BLUE}🎯 Próximos pasos:${NC}"
    echo -e "   1. Configurar protección de rama en GitHub"
    echo -e "   2. La rama stable se actualizará cuando la app esté lista para producción"
    echo -e "   3. semantic-release ahora puede funcionar correctamente"
    echo -e "   4. Ver .releaserc.json para la configuración de branches"
else
    echo "❌ Push cancelado. La rama stable existe localmente pero no en el remoto."
    echo "   Para hacer push manualmente: git push -u origin stable"
fi

echo ""
