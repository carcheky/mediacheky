# Análisis de AGENTS.md y Recomendaciones para MediaCheky

## 📋 Resumen Ejecutivo

**AGENTS.md** es un formato simple y abierto desarrollado por OpenAI para guiar a los agentes de codificación (como GitHub Copilot Coding Agent). Es esencialmente un "README para agentes" que proporciona contexto e instrucciones específicas para ayudar a los agentes de IA a trabajar eficientemente en el proyecto.

**Fuentes consultadas:**
- Repositorio oficial: https://github.com/openai/agents.md
- Documentación de GitHub Copilot sobre instrucciones personalizadas

---

## 🎯 ¿Qué es AGENTS.md?

### Concepto
AGENTS.md es un archivo dedicado y predecible donde se proporciona:
- **Contexto del proyecto**: Información sobre la arquitectura, tecnologías, convenciones
- **Instrucciones específicas**: Cómo ejecutar, probar, construir el proyecto
- **Mejores prácticas**: Convenciones de código, flujos de trabajo recomendados
- **Consejos del entorno de desarrollo**: Comandos útiles, atajos, trucos

### Diferencias con otros archivos de instrucciones

GitHub Copilot soporta múltiples tipos de archivos de instrucciones personalizadas:

| Archivo | Ubicación | Alcance | Soporte Copilot |
|---------|-----------|---------|-----------------|
| **copilot-instructions.md** | `.github/` | Repositorio completo | ✅ Copilot Chat, Coding Agent, Code Review |
| **\*.instructions.md** | `.github/instructions/` | Específico por ruta (con `applyTo`) | ✅ Copilot Chat, Coding Agent |
| **AGENTS.md** | Cualquier lugar del repo | Jerárquico (el más cercano tiene precedencia) | ✅ Coding Agent |
| **CLAUDE.md / GEMINI.md** | Raíz del repo | Repositorio completo | ✅ Agentes específicos |

---

## 🔍 Ejemplos del Mundo Real

### Ejemplo 1: AGENTS.md del repositorio openai/agents.md

```markdown
# Sample AGENTS.md file

## Dev environment tips
- Use `pnpm dlx turbo run where <project_name>` to jump to a package instead of scanning with `ls`.
- Run `pnpm install --filter <project_name>` to add the package to your workspace so Vite, ESLint, and TypeScript can see it.
- Use `pnpm create vite@latest <project_name> -- --template react-ts` to spin up a new React + Vite package with TypeScript checks ready.
- Check the name field inside each package's package.json to confirm the right name—skip the top-level one.

## Testing instructions
- Find the CI plan in the .github/workflows folder.
- Run `pnpm turbo run test --filter <project_name>` to run every check defined for that package.
- From the package root you can just call `pnpm test`. The commit should pass all tests before you merge.
- To focus on one step, add the Vitest pattern: `pnpm vitest run -t "<test name>"`.
- Fix any test or type errors until the whole suite is green.
- After moving files or changing imports, run `pnpm lint --filter <project_name>` to be sure ESLint and TypeScript rules still pass.
- Add or update tests for the code you change, even if nobody asked.

## PR instructions
- Title format: [<project_name>] <Title>
- Always run `pnpm lint` and `pnpm test` before committing.
```

**Características notables:**
- ✅ Comandos específicos y exactos
- ✅ Contexto del flujo de trabajo (pnpm, turbo, Vite)
- ✅ Instrucciones paso a paso para testing
- ✅ Convenciones de commits y PRs
- ✅ Enfoque práctico y orientado a acciones

---

## 🏗️ Estructura Recomendada para AGENTS.md

Basándose en la documentación de GitHub y ejemplos reales, un buen AGENTS.md debe incluir:

### 1. **Información del Proyecto**
- Breve descripción del propósito del proyecto
- Stack tecnológico principal
- Arquitectura de alto nivel

### 2. **Configuración del Entorno de Desarrollo**
- Comandos para iniciar el proyecto
- Variables de entorno requeridas
- Dependencias del sistema
- **IMPORTANTE**: Qué NO ejecutar (ej: `make dev` en nuestro caso)

### 3. **Instrucciones de Construcción y Testing**
- Cómo compilar el proyecto
- Cómo ejecutar tests
- Cómo verificar que todo funciona correctamente

### 4. **Estructura del Proyecto**
- Explicación de directorios principales
- Dónde encontrar qué tipo de código
- Convenciones de nombres de archivos

### 5. **Convenciones de Código**
- Estándares de formateo
- Patrones arquitectónicos a seguir
- Prácticas de seguridad

### 6. **Flujo de Trabajo con Git**
- Formato de commits (Conventional Commits)
- Proceso de creación de PRs
- Revisión de código

### 7. **Comandos Útiles**
- Tabla de referencia rápida de comandos
- Atajos específicos del proyecto
- Scripts del Makefile

---

## 📝 Propuesta: AGENTS.md para MediaCheky

### Opción A: Archivo Único en la Raíz

Crear `/AGENTS.md` con toda la información consolidada.

**Ventajas:**
- ✅ Más fácil de mantener (un solo archivo)
- ✅ Compatible con todos los agentes (OpenAI, Anthropic, Google)
- ✅ Precedencia clara (siempre es el de la raíz)

**Desventajas:**
- ❌ Puede volverse muy grande
- ❌ No permite instrucciones específicas por directorio

### Opción B: AGENTS.md Jerárquicos

Crear múltiples archivos AGENTS.md en diferentes directorios:
- `/AGENTS.md` - Instrucciones generales
- `/internal/AGENTS.md` - Específico para backend
- `/web/AGENTS.md` - Específico para frontend
- `/scripts/AGENTS.md` - Específico para scripts

**Ventajas:**
- ✅ Instrucciones más relevantes según el contexto
- ✅ Archivos más pequeños y manejables
- ✅ Mejor organización modular

**Desventajas:**
- ❌ Más archivos que mantener
- ❌ Puede haber confusión sobre cuál tiene precedencia
- ❌ Requiere habilitación en VS Code (actualmente deshabilitado por defecto)

### Opción C: Combinar .github/copilot-instructions.md + /AGENTS.md

Usar **ambos** archivos de forma complementaria:

**`.github/copilot-instructions.md`** (ya existente) para:
- Instrucciones generales de Copilot Chat
- Configuración de código
- Estándares del proyecto

**`/AGENTS.md`** (nuevo) para:
- Instrucciones específicas para Copilot Coding Agent
- Comandos de desarrollo
- Flujos de trabajo
- Consejos prácticos

**Ventajas:**
- ✅ Separación clara de responsabilidades
- ✅ Aprovecha las fortalezas de cada formato
- ✅ Compatible con todos los features de Copilot

**Desventajas:**
- ❌ Dos archivos que mantener
- ❌ Posible duplicación de información

---

## ✅ Recomendación para MediaCheky

### **Recomiendo la Opción C: Complementar con AGENTS.md**

**Razones:**

1. **Ya tenemos `.github/copilot-instructions.md` robusto** con 400+ líneas de instrucciones detalladas
2. **AGENTS.md puede ser más práctico y orientado a acciones** para el Coding Agent
3. **Separación de responsabilidades clara**:
   - `copilot-instructions.md`: Qué hacer y por qué (filosofía del proyecto)
   - `AGENTS.md`: Cómo hacerlo (comandos, flujos de trabajo)

### Estructura Propuesta

```
mediacheky/
├── .github/
│   └── copilot-instructions.md       # YA EXISTE - Instrucciones generales
├── AGENTS.md                          # NUEVO - Instrucciones para Coding Agent
└── README.md                          # Documentación para usuarios
```

---

## 🚀 Próximos Pasos

### 1. Crear AGENTS.md para MediaCheky

Basado en el contenido actual de `.github/copilot-instructions.md`, crear un AGENTS.md que incluya:

**Secciones propuestas:**

```markdown
# AGENTS Guidelines for MediaCheky

## Project Overview
[Breve descripción: qué es MediaCheky, stack tecnológico]

## CRITICAL: Do NOT Start Services
[Explicación de que NUNCA debe ejecutar make dev, docker-compose, etc.]

## Development Environment Setup
[Comandos para configurar el entorno inicial]

## Building and Testing
[Cómo compilar, ejecutar tests]

## Project Structure Quick Reference
[Tabla con directorios principales y su propósito]

## Common Development Tasks
[Tabla de comandos útiles del Makefile]

## Debugging
[Cómo leer logs, inspeccionar containers, ejecutar comandos en containers]

## Code Conventions
[Resumen de patrones arquitectónicos: Repository, Service, Handler]

## Git Workflow
[Conventional Commits, tipos que triggean builds vs. los que no]

## Useful Commands Reference
[Tabla de referencia rápida]
```

### 2. Actualizar .github/copilot-instructions.md

- Mantener las instrucciones filosóficas y de alto nivel
- Referenciar a AGENTS.md para comandos específicos
- Reducir duplicación

### 3. Testing

- Probar con Copilot Coding Agent en diferentes escenarios
- Verificar que respeta las reglas críticas (no iniciar servicios)
- Ajustar según sea necesario

---

## 🔗 Referencias

- **Repositorio oficial AGENTS.md**: https://github.com/openai/agents.md
- **Sitio web**: https://agents.md
- **Documentación GitHub Copilot - Custom Instructions**: 
  - https://docs.github.com/en/copilot/customizing-copilot/adding-repository-custom-instructions-for-github-copilot
  - https://docs.github.com/en/copilot/using-github-copilot/using-github-copilot-coding-agent-to-work-on-tasks/get-the-best-results-from-github-copilot-coding-agent

---

## 📊 Comparación de Formatos

| Característica | .github/copilot-instructions.md | AGENTS.md | \*.instructions.md |
|----------------|--------------------------------|-----------|-------------------|
| **Soporte** | Copilot Chat, Coding Agent, Code Review | Coding Agent | Copilot Chat, Coding Agent |
| **Alcance** | Repositorio completo | Jerárquico (más cercano gana) | Específico por patrón glob |
| **Ubicación** | `.github/` (fija) | Cualquier lugar | `.github/instructions/` |
| **Formato** | Markdown simple | Markdown simple | Markdown con frontmatter YAML |
| **Precedencia** | Media | Alta (si está cerca del código) | Alta (si coincide el patrón) |
| **Propósito** | Instrucciones generales del proyecto | Guía práctica para agentes | Instrucciones específicas por tipo de archivo |

---

## 💡 Conclusión

**AGENTS.md es una excelente adición para MediaCheky** porque:

1. ✅ Proporciona instrucciones prácticas y accionables para Copilot Coding Agent
2. ✅ Complementa (no reemplaza) las instrucciones existentes en `.github/copilot-instructions.md`
3. ✅ Ayuda a evitar errores críticos (como iniciar servicios sin permiso)
4. ✅ Mejora la eficiencia del agente con comandos específicos y flujos de trabajo claros
5. ✅ Es un estándar abierto adoptado por la comunidad de OpenAI

**Siguiente acción recomendada:** Crear `/AGENTS.md` siguiendo la estructura propuesta en este documento.

---

**Documento creado:** $(date)  
**Autor:** Análisis basado en documentación oficial de GitHub Copilot y repositorio openai/agents.md  
**Versión:** 1.0  
