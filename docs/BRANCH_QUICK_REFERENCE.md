# 🌿 Guía Rápida de Branching - MediaCheky

## Crear Nueva Feature

```bash
git checkout develop
git pull origin develop
git checkout -b feature/mi-nueva-funcionalidad

# ... hacer cambios ...

git add .
git commit -m "feat: agregar nueva funcionalidad"
git push origin feature/mi-nueva-funcionalidad

# Crear PR en GitHub hacia develop
```

## Crear Fix

```bash
git checkout develop
git pull origin develop
git checkout -b fix/corregir-bug

# ... hacer cambios ...

git add .
git commit -m "fix: corregir bug en ..."
git push origin fix/corregir-bug

# Crear PR en GitHub hacia develop
```

## Crear Hotfix (urgente en producción)

```bash
git checkout main
git pull origin main
git checkout -b hotfix/fix-critico

# ... hacer cambios ...

git add .
git commit -m "fix!: corregir problema crítico"
git push origin hotfix/fix-critico

# Crear PR en GitHub hacia main
# Después, mergear también a develop
```

## Actualizar Feature Branch desde Develop

```bash
git checkout feature/mi-feature
git pull origin develop
git merge develop

# Resolver conflictos si existen

git push origin feature/mi-feature
```

## Tipos de Commit (Conventional Commits)

| Tipo | Cuándo Usar | Ejemplo |
|------|-------------|---------|
| `feat:` | Nueva funcionalidad | `feat: add media filters` |
| `fix:` | Corrección de bug | `fix: correct API endpoint` |
| `docs:` | Solo documentación | `docs: update README` |
| `style:` | Formato de código | `style: format with gofmt` |
| `refactor:` | Refactorización | `refactor: extract function` |
| `perf:` | Mejora de rendimiento | `perf: optimize query` |
| `test:` | Agregar tests | `test: add unit tests` |
| `chore:` | Tareas de mantenimiento | `chore: update dependencies` |
| `ci:` | Cambios en CI/CD | `ci: update workflow` |

## Estructura de Ramas

```
main (producción)
  └── hotfix/* (fixes urgentes)

develop (integración)
  ├── feature/* (nuevas funcionalidades)
  ├── fix/* (corrección de bugs)
  ├── docs/* (documentación)
  └── refactor/* (refactorización)
```

## Flujo de Release

```bash
# Cuando develop está listo para producción:

# 1. Crear PR de develop → main
# 2. Revisar y aprobar PR
# 3. Mergear a main
# 4. El CI/CD creará automáticamente el release
```

## Enlaces Útiles

- 📖 Documentación completa: `docs/BRANCH_STRATEGY.md`
- 🛠️ Guía de desarrollo: `DEVELOPMENT.md`
- 📋 Instrucciones develop: `INSTRUCCIONES_DEVELOP.md`
- 🤖 Guía para Copilot: `.github/copilot-instructions.md`

## Comandos Útiles

```bash
# Ver estado actual
git status
git branch -a

# Ver diferencias
git diff
git diff develop

# Ver commits
git log --oneline -10
git log --graph --oneline --all -15

# Limpiar ramas locales ya mergeadas
git branch --merged | grep -v "\*" | grep -v "main" | grep -v "develop" | xargs -n 1 git branch -d

# Actualizar lista de ramas remotas
git fetch --prune
```

## ⚠️ Reglas Importantes

1. ❌ NUNCA hacer push directo a `main` o `develop`
2. ✅ SIEMPRE crear PR para mergear cambios
3. ✅ SIEMPRE usar commits convencionales
4. ✅ SIEMPRE hacer pull antes de crear nueva rama
5. ✅ SIEMPRE probar localmente antes de push
6. ✅ SIEMPRE pedir code review antes de mergear

---

**Última actualización**: 2025-10-31
