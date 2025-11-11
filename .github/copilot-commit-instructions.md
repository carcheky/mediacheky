# GitHub Copilot - Commit Message Instructions

## ⚠️ CRITICAL RULES FOR COMMIT MESSAGES

### 1. Language: ALWAYS ENGLISH
- **NEVER** use Spanish or any other language
- **ALL** commit messages MUST be in English
- No exceptions, no matter the context

### 2. Format: Conventional Commits
- **ALWAYS** follow Conventional Commits specification
- Format: `<type>(<scope>): <description>`
- Reference: https://www.conventionalcommits.org/

### 3. Writing Style
- Use **present tense**: "add" not "added"
- Use **imperative mood**: "fix" not "fixes"
- **Capitalize** first letter after type/scope
- **No period** at the end of subject line
- Keep subject line **under 72 characters**

### 4. Commit Types

#### Release-triggering types:
- `feat`: New feature
- `fix`: Bug fix
- `perf`: Performance improvement

#### Non-release types:
- `docs`: Documentation only
- `chore`: Maintenance tasks
- `refactor`: Code restructuring
- `test`: Adding/updating tests
- `style`: Code formatting
- `ci`: CI/CD changes
- `build`: Build system changes

### 5. Examples

#### ✅ CORRECT Examples:
```
feat(services): add Radarr configuration panel
fix(docker): correct container status check
perf(db): optimize query performance
docs(readme): update installation guide
chore(deps): update Go dependencies
refactor(handler): simplify error handling
test(service): add unit tests for template engine
style(format): fix code formatting
ci(github): update workflow to Go 1.25
```

#### ❌ WRONG Examples:
```
❌ feat(servicios): agregar panel de configuración  [Spanish]
❌ feat(services): added configuration panel         [Past tense]
❌ feat(services): adds configuration panel          [Wrong mood]
❌ Add configuration panel                           [No type]
❌ feat(services): Add configuration panel.          [Period at end]
```

### 6. Scope Guidelines
- Use component/module name: `services`, `docker`, `handler`, `config`
- Be specific but concise
- Examples: `(radarr)`, `(template)`, `(database)`, `(api)`

### 7. When to Use Each Type

- **feat**: Adding new functionality that users will see
- **fix**: Fixing a bug that users experienced
- **perf**: Improving performance (faster, less memory)
- **docs**: Updating README, comments, API docs
- **chore**: Updating dependencies, configs (not visible to users)
- **refactor**: Restructuring code without changing behavior
- **test**: Adding or modifying test files
- **style**: Formatting, whitespace, linting fixes
- **ci**: GitHub Actions, CI/CD pipeline changes
- **build**: Makefile, Dockerfile, build scripts

## Priority Rules

1. **Language = English** (no exceptions)
2. **Format = Conventional Commits** (always)
3. **Present tense + Imperative mood** (mandatory)
4. **Subject line < 72 chars** (hard limit)
5. **No period at end** (never)

## Remember

- When in doubt, use `chore` for maintenance tasks
- Use `feat` only for user-visible new features
- Use `fix` only for actual bugs that users experienced
- **NEVER EVER** write commits in Spanish
