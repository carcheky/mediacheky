# Copilot Commit Message Generation Rules

## ⚠️ CRITICAL: Commit Type Selection

**Only `feat`, `fix`, and `perf` trigger releases and Docker builds!**

Use these types strategically to avoid unnecessary builds.

## Decision Tree

**Does it change runtime behavior?**

- **YES** → Does it add functionality?
  - **YES** → `feat`
  - **NO** → Does it fix a bug?
    - **YES** → `fix`
    - **NO** → Does it improve performance?
      - **YES** → `perf`
      - **NO** → `refactor`

- **NO** → What does it change?
  - Documentation → `docs`
  - Tests → `test`
  - Dependencies/Config → `chore`
  - CI/CD → `ci`
  - Code formatting → `style`

## Types that TRIGGER builds (use sparingly):

### `feat` - New user-facing feature or significant functionality
- ✅ New API endpoint
- ✅ New UI component or page
- ✅ New integration with external service
- ❌ Adding a comment to code
- ❌ Updating .env.example

### `fix` - Bug fix that affects runtime behavior
- ✅ Fix crash or error in application
- ✅ Fix incorrect data processing
- ✅ Fix API response issue
- ❌ Fix typo in comment
- ❌ Fix README formatting

### `perf` - Performance improvement that affects runtime
- ✅ Optimize database query
- ✅ Reduce memory usage
- ✅ Improve API response time
- ❌ Code cleanup without measurable impact

## Types that DO NOT trigger builds (use for maintenance):

### `docs` - Documentation-only changes
- ✅ Update README.md
- ✅ Update .env.example
- ✅ Add code comments
- ✅ Update documentation files in /docs
- ✅ Update copilot-instructions.md

### `chore` - Maintenance tasks, config changes, dependencies
- ✅ Update dependencies in go.mod
- ✅ Update .gitignore
- ✅ Update Makefile
- ✅ Update docker-compose.yml (non-functional)
- ✅ Cleanup temporary files

### `refactor` - Code restructuring without changing behavior
- ✅ Extract function/method
- ✅ Rename variables for clarity
- ✅ Move code between files
- ❌ If it changes behavior, use `feat` or `fix`

### `test` - Adding or updating tests only
- ✅ Add unit tests
- ✅ Add integration tests
- ✅ Update test fixtures

### `style` - Code style/formatting changes
- ✅ Run gofmt
- ✅ Fix linting issues
- ✅ Adjust indentation

### `ci` - CI/CD configuration changes
- ✅ Update GitHub Actions workflows
- ✅ Update release configuration

## Format

`<type>(<scope>): <description>`

**ALL commit messages MUST be in English** following Conventional Commits specification.

## Examples

### TRIGGERS BUILD (runtime changes)
```
feat(api): add endpoint for bulk media deletion
fix(sync): correct torrent hash matching algorithm
perf(db): add index on media.created_at for faster queries
```

### DOES NOT TRIGGER BUILD (maintenance)
```
docs(config): update .env.example with Bazarr configuration
chore(deps): update Go dependencies to latest versions
refactor(handler): extract validation logic to separate function
test(repository): add unit tests for media queries
style(models): format code with gofmt
ci(release): update semantic-release configuration
```
