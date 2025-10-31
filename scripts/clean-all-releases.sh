#!/bin/bash

# Script para limpiar completamente todos los releases, tags y packages de GitHub
# Requiere: gh CLI instalado y autenticado

REPO="carcheky/keepercheky"

echo "🧹 Limpiando todos los releases, tags y packages de GitHub..."
echo ""

# 0. Cancelar workflows en ejecución
echo "🛑 Cancelando todos los workflows en ejecución..."
gh run list --limit 100 --json databaseId,status --jq '.[] | select(.status=="in_progress") | .databaseId' 2>/dev/null | while read -r id; do
  if [ -n "$id" ]; then
    echo "  Cancelando workflow: $id"
    gh run cancel "$id" 2>/dev/null || true
  fi
done
echo "✅ Workflows cancelados"
echo ""

# 1. Borrar todos los GitHub Releases
echo "📦 Borrando GitHub Releases..."
gh release list --repo "$REPO" --limit 1000 2>/dev/null | while read -r line; do
    tag=$(echo "$line" | awk '{print $1}')
    if [ -n "$tag" ]; then
        echo "  Borrando release: $tag"
        gh release delete "$tag" --repo "$REPO" --yes 2>/dev/null || true
    fi
done
echo "✅ Releases borrados"
echo ""

# 2. Borrar todos los tags remotos
echo "🏷️  Borrando tags remotos..."
git ls-remote --tags origin 2>/dev/null | awk '{print $2}' | sed 's/refs\/tags\///' | while read -r tag; do
    if [ -n "$tag" ]; then
        echo "  Borrando tag remoto: $tag"
        git push --delete origin "$tag" 2>/dev/null || true
    fi
done
echo "✅ Tags remotos borrados"
echo ""

# 3. Borrar todos los tags locales
echo "🏷️  Borrando tags locales..."
git tag -l 2>/dev/null | while read -r tag; do
    if [ -n "$tag" ]; then
        echo "  Borrando tag local: $tag"
        git tag -d "$tag" 2>/dev/null || true
    fi
done
echo "✅ Tags locales borrados"
echo ""

# 4. Borrar package completo de GitHub Container Registry
echo "🐳 Borrando package de GitHub Container Registry..."
if gh api /user/packages/container/keepercheky >/dev/null 2>&1; then
    echo "  Borrando package completo: keepercheky"
    gh api --method DELETE /user/packages/container/keepercheky 2>/dev/null || true
    echo "✅ Package Docker borrado"
else
    echo "  No se encontró package Docker para borrar"
fi
echo ""

# 5. Verificar limpieza
echo "🔍 Verificando limpieza..."
echo ""
echo "Tags locales restantes:"
git tag -l 2>/dev/null | wc -l | xargs echo "  "
echo ""
echo "Tags remotos restantes:"
git ls-remote --tags origin 2>/dev/null | wc -l | xargs echo "  "
echo ""
echo "Releases restantes:"
gh release list --repo "$REPO" --limit 10 2>/dev/null | wc -l | xargs echo "  "
echo ""
echo "Packages Docker restantes:"
gh api /user/packages/container/keepercheky 2>/dev/null >/dev/null && echo "  1" || echo "  0"
echo ""

echo "✅ Limpieza completada!"
echo ""

# 6. Pull, commit y push para disparar workflow
echo "� Haciendo pull, commit y push..."
git pull
rm -fr CHANGELOG.md
git add .
git commit -m "chore: 1 apply writerOpts to avoid Date.prototype.toString error" --allow-empty
git push
sleep 20
git commit -m "fix: 2 apply writerOpts to avoid Date.prototype.toString error" --allow-empty
git push
echo "✅ Commit y push realizados"
echo ""

# 7. Esperar 1 minuto para que el workflow se ejecute
echo "⏳ Esperando a que el workflow se ejecute..."
echo "   Opciones:"
echo "   - Espera automática de 60 segundos"
echo "   - Presiona cualquier tecla para continuar inmediatamente"
echo ""
read -t 300 -n 1 -s -r && echo "✅ Continuando..." || echo "✅ Tiempo de espera completado"
echo ""

# 8. Ejecutar check-last-workflow y guardar en logl.log
echo "🔍 Ejecutando check-last-workflow.sh..."
bash scripts/check-last-workflow.sh > logl.log
echo "✅ Resultados guardados en logl.log"
echo ""

echo "📝 Próximos pasos:"
echo "   1. Revisa logl.log para ver el resultado del workflow"
echo "   2. Verifica packages en: https://github.com/$REPO/pkgs/container/keepercheky"
echo ""
echo "⚠️  NOTA: Si hay errores con packages, asegúrate de tener permisos:"
echo "   gh auth refresh -h github.com -s delete:packages,write:packages,read:packages"
