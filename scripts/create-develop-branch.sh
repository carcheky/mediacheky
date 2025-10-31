#!/bin/bash

# Script para crear la rama develop desde main
# Este script debe ser ejecutado por alguien con permisos de push al repositorio

set -e

echo "🚀 Creando rama develop desde main..."
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

# Verificar si la rama develop ya existe localmente
if git show-ref --verify --quiet refs/heads/develop; then
    echo -e "${YELLOW}⚠️  La rama 'develop' ya existe localmente${NC}"
    read -p "¿Deseas eliminarla y recrearla? (s/N): " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Ss]$ ]]; then
        git branch -D develop
        echo -e "${GREEN}✅ Rama local 'develop' eliminada${NC}"
    else
        echo "❌ Operación cancelada"
        exit 0
    fi
fi

# Verificar si la rama develop ya existe en el remoto
if git ls-remote --heads origin develop | grep -q develop; then
    echo -e "${YELLOW}⚠️  La rama 'develop' ya existe en el remoto${NC}"
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

# Verificar que existe la rama main en el remoto
if ! git ls-remote --heads origin main | grep -q main; then
    echo "❌ Error: No se encontró la rama 'main' en el remoto"
    exit 1
fi

# Checkout a main
echo -e "${BLUE}🔄 Cambiando a rama main...${NC}"
git checkout main 2>/dev/null || git checkout -b main origin/main

# Hacer pull de main
echo -e "${BLUE}📥 Actualizando rama main...${NC}"
git pull origin main

# Obtener el commit actual
COMMIT=$(git rev-parse HEAD)
COMMIT_SHORT=$(git rev-parse --short HEAD)
COMMIT_MSG=$(git log -1 --pretty=%B)

echo ""
echo -e "${BLUE}📌 Último commit en main:${NC}"
echo -e "   ${COMMIT_SHORT} - ${COMMIT_MSG}"
echo ""

# Crear la rama develop
echo -e "${BLUE}🌿 Creando rama develop...${NC}"
git checkout -b develop

# Confirmar antes de hacer push
echo ""
echo -e "${YELLOW}⚠️  A punto de hacer push de la rama develop al remoto${NC}"
read -p "¿Continuar? (S/n): " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Nn]$ ]]; then
    echo -e "${BLUE}📤 Haciendo push de develop...${NC}"
    git push -u origin develop
    
    echo ""
    echo -e "${GREEN}✅ ¡Rama develop creada exitosamente!${NC}"
    echo ""
    echo -e "${BLUE}📋 Resumen:${NC}"
    echo -e "   • Rama develop creada desde main"
    echo -e "   • Commit base: ${COMMIT_SHORT}"
    echo -e "   • Push realizado a origin/develop"
    echo ""
    echo -e "${BLUE}🎯 Próximos pasos:${NC}"
    echo -e "   1. Configurar protección de rama en GitHub"
    echo -e "   2. Configurar require pull request reviews"
    echo -e "   3. Ver docs/BRANCH_STRATEGY.md para más detalles"
else
    echo "❌ Push cancelado. La rama develop existe localmente pero no en el remoto."
    echo "   Para hacer push manualmente: git push -u origin develop"
fi

echo ""
