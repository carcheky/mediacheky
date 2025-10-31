# 🚀 Release Workflow - KeeperCheky

## Descripción General

KeeperCheky utiliza un **workflow unificado** inspirado en [Jellyseerr](https://github.com/seerr-team/seerr), que combina semantic-release y construcción de imágenes Docker en un solo pipeline.

## 🔄 Flujo del Proceso

```mermaid
graph LR
    A[Push a develop/stable] --> B[Semantic Release]
    B --> C{¿Nueva versión?}
    C -->|No| D[Fin]
    C -->|Sí| E[Crear CHANGELOG.md]
    E --> F[Crear commit chore]
    F --> G[Crear tag Git]
    G --> H[Push tag + commit]
    H --> I[Build Docker]
    I --> J[Push a GHCR]
    J --> K[Notificación]
```

## 📋 Workflow Unificado

### Archivo: `.github/workflows/release.yml`

El workflow se activa en:
- Push a `develop` → genera versiones `1.0.0-dev.X`
- Push a `stable` → genera versiones `1.0.0`

### Jobs:

#### 1. **semantic-release**
- Analiza commits convencionales
- Determina nueva versión (si aplica)
- Genera `CHANGELOG.md` actualizado
- Crea commit `chore(release): X.Y.Z`
- Crea tag Git `vX.Y.Z`
- Publica GitHub Release
- **Outputs:**
  - `new_release_published`: `true`/`false`
  - `new_release_version`: `1.0.0-dev.1`
  - `new_release_git_tag`: `v1.0.0-dev.1`

#### 2. **build-and-push**
- **Condición:** Solo si `new_release_published == 'true'`
- Checkout del código en el tag exacto creado
- Construcción multi-arquitectura (`linux/amd64`, `linux/arm64`)
- Push a GitHub Container Registry
- **Tags generados:**
  - Para versión `1.0.0-dev.1`:
    - `ghcr.io/carcheky/keepercheky:1.0.0-dev.1`
    - `ghcr.io/carcheky/keepercheky:develop`
  - Para versión `1.0.0`:
    - `ghcr.io/carcheky/keepercheky:1.0.0`
    - `ghcr.io/carcheky/keepercheky:1.0`
    - `ghcr.io/carcheky/keepercheky:1`
    - `ghcr.io/carcheky/keepercheky:latest`
    - `ghcr.io/carcheky/keepercheky:stable`

#### 3. **notify**
- Siempre se ejecuta (incluso si fallan pasos anteriores)
- Genera resumen del workflow
- Estado de cada job

## 🏷️ Estrategia de Tags

### Rama `develop` (pre-release)
```
Versión: 1.0.0-dev.1
Tags Docker:
  - 1.0.0-dev.1
  - develop
```

### Rama `stable` (producción)
```
Versión: 1.0.0
Tags Docker:
  - 1.0.0
  - 1.0
  - 1
  - latest
  - stable
```

## 📝 Commits Convencionales

El workflow usa [Conventional Commits](https://www.conventionalcommits.org/):

| Tipo       | Release | Descripción                  |
|------------|---------|------------------------------|
| `feat`     | minor   | Nueva funcionalidad          |
| `fix`      | patch   | Corrección de bug            |
| `perf`     | patch   | Mejora de rendimiento        |
| `refactor` | patch   | Refactorización de código    |
| `docs`     | -       | Documentación                |
| `chore`    | -       | Mantenimiento                |
| `test`     | -       | Tests                        |
| `BREAKING` | major   | Cambio incompatible          |

### Ejemplos:

```bash
# Nueva funcionalidad (minor: 1.0.0 → 1.1.0)
git commit -m "feat(sync): add intelligent torrent matching"

# Corrección de bug (patch: 1.0.0 → 1.0.1)
git commit -m "fix(ui): resolve mobile tooltip display"

# Breaking change (major: 1.0.0 → 2.0.0)
git commit -m "feat(api)!: redesign configuration structure

BREAKING CHANGE: Config file format changed from YAML to TOML"
```

## 🔧 Configuración

### `.releaserc.json`

```json
{
  "branches": [
    { "name": "stable" },
    { "name": "develop", "prerelease": "dev" }
  ],
  "plugins": [
    "@semantic-release/commit-analyzer",
    "@semantic-release/release-notes-generator",
    "@semantic-release/changelog",
    "@semantic-release/git",
    "@semantic-release/github"
  ]
}
```

### Secrets necesarios

- `GITHUB_TOKEN`: Automático (GitHub Actions)
- `PAT_TOKEN`: Personal Access Token (opcional, para bypass de protecciones)

## 🎯 Ventajas del Workflow Unificado

### ✅ Resuelve problemas anteriores:

1. **Tag triggering inconsistente**: Ya no hay workflows separados
2. **CHANGELOG desactualizado en Docker**: La imagen siempre contiene el changelog correcto
3. **Race conditions**: Todo es secuencial en un solo workflow
4. **Complejidad**: Un solo archivo vs. múltiples workflows coordinados

### ✅ Beneficios adicionales:

- Construcción solo cuando hay nueva versión (ahorra recursos)
- Workflow más fácil de entender y mantener
- Resumen automático del proceso
- Notificaciones consistentes
- Inspirado en proyectos maduros (Jellyseerr)

## 🧪 Cómo Probar

### Escenario 1: Feature en develop
```bash
git checkout develop
git pull
echo "test" > test.txt
git add test.txt
git commit -m "feat(test): add test feature"
git push origin develop
```

**Resultado esperado:**
- Nueva versión: `1.0.0-dev.1` (o siguiente)
- Tag creado: `v1.0.0-dev.1`
- Docker image: `ghcr.io/carcheky/keepercheky:1.0.0-dev.1` + `develop`

### Escenario 2: Fix en develop
```bash
git commit -m "fix(test): correct test issue"
git push origin develop
```

**Resultado esperado:**
- Nueva versión: `1.0.0-dev.2` (incremento de prerelease)
- Docker image: `ghcr.io/carcheky/keepercheky:1.0.0-dev.2` + `develop`

### Escenario 3: Docs (sin release)
```bash
git commit -m "docs: update README"
git push origin develop
```

**Resultado esperado:**
- **No se genera nueva versión**
- No se construye imagen Docker
- Workflow termina en semantic-release job

## 📊 Monitoreo

### Ver workflow en ejecución:
```
https://github.com/carcheky/keepercheky/actions
```

### Ver releases:
```
https://github.com/carcheky/keepercheky/releases
```

### Ver imágenes Docker:
```
https://github.com/carcheky/keepercheky/pkgs/container/keepercheky
```

## 🐛 Troubleshooting

### El workflow no se ejecuta
- Verificar que el commit no sea `chore(release):` (filtro automático)
- Verificar permisos del token PAT_TOKEN

### No se crea nueva versión
- Verificar que el commit siga Conventional Commits
- Commits tipo `docs`, `chore`, `test` no generan releases

### Docker build falla
- Verificar logs del job `build-and-push`
- Verificar que el Dockerfile sea correcto
- Verificar que GITHUB_TOKEN tenga permisos de packages

### Tag no se crea
- Verificar configuración de `.releaserc.json`
- Verificar que semantic-release se ejecute correctamente

## 📚 Referencias

- [Conventional Commits](https://www.conventionalcommits.org/)
- [Semantic Release](https://semantic-release.gitbook.io/)
- [Jellyseerr Workflow](https://github.com/seerr-team/seerr/blob/develop/.github/workflows/release.yml)
- [Docker Metadata Action](https://github.com/docker/metadata-action)

## 🔄 Migración desde Workflows Antiguos

Los workflows anteriores fueron renombrados:
- `semantic-release.yml` → `semantic-release.yml.old`
- `docker-build.yml` → `docker-build.yml.old`

Están disponibles como referencia pero **no se ejecutarán**.

---

**Última actualización:** 2025-10-28
**Versión workflow:** 1.0.0
