#!/bin/bash

# Script para revisar el estado del último workflow de manera automática
# Requiere: gh CLI instalado y autenticado

REPO="carcheky/keepercheky"

# Función para obtener y mostrar información del workflow
check_workflow() {
    echo "🔍 Revisando último workflow..."
    echo ""

    # Obtener información del último workflow
    LAST_RUN=$(gh run list --repo "$REPO" --limit 1 --json databaseId,status,conclusion,name,headBranch,event,createdAt,url 2>/dev/null)

    if [ -z "$LAST_RUN" ] || [ "$LAST_RUN" = "[]" ]; then
        echo "❌ No se encontraron workflows"
        exit 1
    fi

    # Parsear información
    RUN_ID=$(echo "$LAST_RUN" | jq -r '.[0].databaseId')
    STATUS=$(echo "$LAST_RUN" | jq -r '.[0].status')
    CONCLUSION=$(echo "$LAST_RUN" | jq -r '.[0].conclusion')
    NAME=$(echo "$LAST_RUN" | jq -r '.[0].name')
    BRANCH=$(echo "$LAST_RUN" | jq -r '.[0].headBranch')
    EVENT=$(echo "$LAST_RUN" | jq -r '.[0].event')
    CREATED=$(echo "$LAST_RUN" | jq -r '.[0].createdAt')
    URL=$(echo "$LAST_RUN" | jq -r '.[0].url')
}

# Bucle principal: repetir mientras haya jobs en progreso
while true; do
    check_workflow
    
    # Limpiar pantalla para actualización (opcional, comenta si no quieres)
    # clear

# Mostrar información básica
echo "📊 Información del Workflow"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Nombre:       $NAME"
echo "  ID:           $RUN_ID"
echo "  Branch:       $BRANCH"
echo "  Evento:       $EVENT"
echo "  Creado:       $CREATED"
echo "  URL:          $URL"
echo ""

# Determinar estado con emojis
if [ "$STATUS" = "completed" ]; then
    if [ "$CONCLUSION" = "success" ]; then
        echo "✅ Estado:      COMPLETADO EXITOSAMENTE"
    elif [ "$CONCLUSION" = "failure" ]; then
        echo "❌ Estado:      FALLÓ"
    elif [ "$CONCLUSION" = "cancelled" ]; then
        echo "🚫 Estado:      CANCELADO"
    elif [ "$CONCLUSION" = "skipped" ]; then
        echo "⏭️  Estado:      OMITIDO"
    else
        echo "❓ Estado:      COMPLETADO ($CONCLUSION)"
    fi
elif [ "$STATUS" = "in_progress" ]; then
    echo "⏳ Estado:      EN PROGRESO"
elif [ "$STATUS" = "queued" ]; then
    echo "⏸️  Estado:      EN COLA"
else
    echo "❓ Estado:      $STATUS"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Mostrar TODOS los jobs del workflow
echo ""
echo "📋 Jobs del workflow:"
echo ""

gh run view "$RUN_ID" --repo "$REPO" --json jobs --jq '.jobs[] | "  \(if .conclusion == "success" then "✅" elif .conclusion == "failure" then "❌" elif .conclusion == "cancelled" then "🚫" elif .conclusion == "skipped" then "⏭️" elif .status == "in_progress" then "⏳" elif .status == "queued" then "⏸️" else "❓" end) \(.name)\n     Status: \(.status) | Conclusion: \(.conclusion // "N/A")"' 2>/dev/null

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Mostrar logs de TODOS los jobs (no solo los fallidos)
echo ""
echo "📄 Logs de TODOS los jobs:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

gh run view "$RUN_ID" --repo "$REPO" --json jobs --jq '.jobs[] | .name' 2>/dev/null | while read -r job_name; do
    echo "📦 Job: $job_name"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    gh run view "$RUN_ID" --repo "$REPO" --job "$job_name" --log 2>/dev/null | tail -50 || echo "   (No se pudieron obtener logs)"
    echo ""
done

# Si falló, mostrar jobs que fallaron con más detalle
if [ "$CONCLUSION" = "failure" ]; then
    echo ""
    echo "🔍 Revisando jobs fallidos..."
    echo ""
    
    gh run view "$RUN_ID" --repo "$REPO" --json jobs --jq '.jobs[] | select(.conclusion == "failure") | "  ❌ Job: \(.name)\n     Conclusión: \(.conclusion)\n"' 2>/dev/null
    
    echo ""
    echo "📋 Logs de errores:"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    # Obtener logs de los steps que fallaron
    gh api "/repos/$REPO/actions/runs/$RUN_ID/jobs" 2>/dev/null | jq -r '.jobs[] | select(.conclusion == "failure") | .steps[] | select(.conclusion == "failure") | .name' | while read -r step_name; do
        echo ""
        echo "🔴 Step fallido: $step_name"
        echo ""
    done
    
    # Intentar obtener logs completos del job fallido
    echo ""
    echo "📄 Logs completos del workflow fallido:"
    echo ""
    gh run view "$RUN_ID" --repo "$REPO" --log-failed 2>/dev/null | tail -100 || echo "   (No se pudieron obtener logs detallados)"
    
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "🔗 Ver completo en navegador:"
    echo "   gh run view $RUN_ID --web"
fi

# Si está en progreso, mostrar progreso
if [ "$STATUS" = "in_progress" ]; then
    echo ""
    echo "⏳ Jobs en progreso..."
    echo ""
    
    gh run view "$RUN_ID" --repo "$REPO" --json jobs --jq '.jobs[] | "  \(if .status == "completed" then "✅" elif .status == "in_progress" then "⏳" else "⏸️" end) \(.name) - \(.status)"' 2>/dev/null
    
    echo ""
    echo "🔄 Para monitorear en tiempo real:"
    echo "   gh run watch $RUN_ID"
fi

# Comprobar si hay jobs en progreso
JOBS_IN_PROGRESS=$(gh run view "$RUN_ID" --repo "$REPO" --json jobs --jq '[.jobs[] | select(.status == "in_progress")] | length' 2>/dev/null)

# Si hay jobs en progreso, esperar y repetir
if [ "$JOBS_IN_PROGRESS" -gt 0 ]; then
    echo ""
    echo "⏳ Esperando 5 minutos antes de volver a comprobar..."
    echo "   (Presiona Ctrl+C para cancelar)"
    sleep 300
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "🔄 ACTUALIZANDO ESTADO..."
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    continue
fi

# Si no hay jobs en progreso, salir del bucle
break

done

# Si fue exitoso, mostrar releases/tags creados
if [ "$CONCLUSION" = "success" ] && [ "$NAME" = "Release" ]; then
    echo ""
    echo "🎉 Release exitoso!"
    echo ""
    echo "📦 Últimas releases:"
    gh release list --repo "$REPO" --limit 3 2>/dev/null | while read -r line; do
        echo "   $line"
    done
    
    echo ""
    echo "🏷️  Últimos tags:"
    git tag --sort=-creatordate | head -3 | while read -r tag; do
        echo "   $tag"
    done
    
    echo ""
    echo "🐳 Paquetes Docker:"
    DOCKER_PACKAGES=$(gh api /user/packages/container/keepercheky/versions 2>/dev/null | jq -r 'if type == "array" then .[] | .metadata.container.tags | join(", ") else empty end' 2>/dev/null | head -3)
    if [ -n "$DOCKER_PACKAGES" ]; then
        echo "$DOCKER_PACKAGES" | while read -r tags; do
            echo "   $tags"
        done
    else
        echo "   (No disponible o sin permisos)"
    fi
fi

echo ""
