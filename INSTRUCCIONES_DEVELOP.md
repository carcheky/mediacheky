# Instrucciones para Crear la Rama Develop

## ✅ Cambios Realizados

Este PR incluye:

1. **Documentación de estrategia de branching** - `docs/BRANCH_STRATEGY.md`
   - Explicación completa del flujo de trabajo con ramas
   - Guías para crear feature branches, hotfixes, etc.
   - Configuración recomendada de protección de ramas en GitHub

2. **Script automatizado** - `scripts/create-develop-branch.sh`
   - Script interactivo para crear la rama `develop` desde `main`
   - Incluye validaciones y confirmaciones
   - Proporciona resumen de acciones realizadas

3. **Actualización de DEVELOPMENT.md**
   - Sección de contribución actualizada con el flujo de branching
   - Referencias a la documentación de estrategia de ramas

## 🚀 Pasos para Crear la Rama Develop

### Opción 1: Usando el Script (Recomendado)

```bash
# 1. Navegar al directorio del proyecto
cd /path/to/mediacheky

# 2. Ejecutar el script
./scripts/create-develop-branch.sh
```

El script te guiará paso a paso y:
- ✅ Verificará que estás en un repositorio git
- ✅ Hará fetch de la última versión de `main`
- ✅ Creará la rama `develop` desde `main`
- ✅ Te pedirá confirmación antes de hacer push
- ✅ Hará push de la rama al repositorio remoto
- ✅ Mostrará un resumen de acciones completadas

### Opción 2: Manualmente

Si prefieres crear la rama manualmente:

```bash
# 1. Asegurarse de tener la última versión de main
git fetch origin
git checkout main
git pull origin main

# 2. Crear la rama develop
git checkout -b develop

# 3. Hacer push al remoto
git push -u origin develop
```

## 📋 Siguiente Paso: Configurar Protección de Ramas

Una vez creada la rama `develop`, es importante configurar protección en GitHub:

### Para la rama `main`:

1. Ve a GitHub → Settings → Branches → Add rule
2. Branch name pattern: `main`
3. Configura:
   - ✅ Require pull request reviews before merging (1-2 reviewers)
   - ✅ Require status checks to pass before merging
   - ✅ Require branches to be up to date before merging
   - ✅ Include administrators
   - ✅ Restrict who can push to matching branches

### Para la rama `develop`:

1. Ve a GitHub → Settings → Branches → Add rule
2. Branch name pattern: `develop`
3. Configura:
   - ✅ Require pull request reviews before merging (1 reviewer)
   - ✅ Require status checks to pass before merging
   - ✅ Require branches to be up to date before merging

## 🎯 Flujo de Trabajo Después de Crear Develop

Una vez que la rama `develop` esté creada:

1. **Nuevas features** se crean desde `develop`:
   ```bash
   git checkout develop
   git pull origin develop
   git checkout -b feature/nueva-funcionalidad
   ```

2. **Pull Requests** se crean hacia `develop` (no hacia `main`)

3. **Releases** se hacen mergeando `develop` → `main`

4. **Hotfixes** se crean desde `main` y se mergean a ambas ramas

## 📖 Documentación Adicional

- **Estrategia completa de branching**: `docs/BRANCH_STRATEGY.md`
- **Guía de desarrollo**: `DEVELOPMENT.md`
- **Convenciones de commits**: Ver `.github/copilot-instructions.md`

## ⚠️ Notas Importantes

1. **No se puede crear la rama directamente desde este PR** debido a limitaciones de autenticación en el entorno de CI/CD
2. **Debes ejecutar el script o crear la rama manualmente** con permisos de push al repositorio
3. **El script es seguro** - pide confirmación antes de hacer push
4. **La rama ya existe localmente** en este entorno de pruebas, pero necesita ser creada en el repositorio remoto

## ✨ Beneficios de Esta Estructura

- 🔒 **Protección de `main`** - Solo código probado y revisado
- 🚀 **Desarrollo ágil en `develop`** - Integración continua de features
- 🐛 **Hotfixes rápidos** - Desde `main` para producción
- 📦 **Releases controlados** - Merge de `develop` a `main` cuando esté listo
- 🔄 **Compatibilidad con CI/CD** - Las ramas `main` y `develop` ya están configuradas en los workflows

---

**Siguiente acción recomendada**: Ejecutar `./scripts/create-develop-branch.sh` después de mergear este PR.
