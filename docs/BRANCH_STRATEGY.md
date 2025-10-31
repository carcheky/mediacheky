# Estrategia de Ramas - MediaCheky

## Descripción General

Este documento describe la estrategia de branching (ramas) utilizada en el proyecto MediaCheky.

## Ramas Principales

### `main`
- **Propósito**: Rama de producción estable
- **Protección**: Debe estar protegida contra pushes directos
- **Merges**: Solo desde `develop` después de pruebas exhaustivas
- **Estado**: Siempre debe estar en un estado desplegable

### `develop`
- **Propósito**: Rama de integración para desarrollo
- **Origen**: Creada desde `main`
- **Protección**: Recomendado proteger contra pushes directos
- **Merges**: Recibe features, fixes y mejoras antes de ir a `main`
- **Estado**: Debe mantenerse estable, pero puede contener features en progreso

## Ramas de Trabajo

### Feature Branches (`feature/*`)
```bash
# Crear desde develop
git checkout develop
git pull origin develop
git checkout -b feature/nombre-descriptivo
```

### Fix Branches (`fix/*`)
```bash
# Crear desde develop (o desde main si es hotfix crítico)
git checkout develop
git pull origin develop
git checkout -b fix/descripcion-del-bug
```

### Hotfix Branches (`hotfix/*`)
```bash
# Crear desde main para fixes críticos en producción
git checkout main
git pull origin main
git checkout -b hotfix/descripcion-del-problema
```

## Flujo de Trabajo

1. **Desarrollo de Features**
   ```
   develop <- feature/nueva-funcionalidad
   ```

2. **Integración**
   ```
   feature/nueva-funcionalidad -> develop (via PR)
   ```

3. **Release**
   ```
   develop -> main (via PR después de QA)
   ```

4. **Hotfixes**
   ```
   hotfix/fix-critico -> main (via PR)
   hotfix/fix-critico -> develop (via PR para mantener sincronizado)
   ```

## Creación de la Rama `develop`

La rama `develop` se crea una única vez desde `main`:

```bash
# 1. Asegurarse de estar en main actualizado
git checkout main
git pull origin main

# 2. Crear la rama develop
git checkout -b develop

# 3. Hacer push al remoto
git push -u origin develop
```

## Configuración Recomendada en GitHub

### Protección de Ramas

**Para `main`:**
- ✅ Require pull request reviews before merging (1-2 reviewers)
- ✅ Require status checks to pass before merging
- ✅ Require branches to be up to date before merging
- ✅ Include administrators
- ✅ Restrict who can push to matching branches

**Para `develop`:**
- ✅ Require pull request reviews before merging (1 reviewer)
- ✅ Require status checks to pass before merging
- ✅ Require branches to be up to date before merging

## Sincronización de Ramas

### Actualizar `develop` desde `main` (después de release)
```bash
git checkout develop
git pull origin develop
git merge main
git push origin develop
```

### Actualizar feature branch desde `develop`
```bash
git checkout feature/mi-feature
git pull origin develop
git merge develop
# Resolver conflictos si existen
git push origin feature/mi-feature
```

## Versionado Semántico

El proyecto utiliza [Semantic Versioning](https://semver.org/):

- **MAJOR**: Cambios incompatibles en la API
- **MINOR**: Nueva funcionalidad compatible con versiones anteriores
- **PATCH**: Corrección de bugs compatible con versiones anteriores

Los commits siguen [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` → Incrementa MINOR
- `fix:` → Incrementa PATCH
- `BREAKING CHANGE:` → Incrementa MAJOR

## Referencias

- [Git Flow](https://nvie.com/posts/a-successful-git-branching-model/)
- [GitHub Flow](https://guides.github.com/introduction/flow/)
- [Conventional Commits](https://www.conventionalcommits.org/)

---

**Última actualización**: 2025-10-31
