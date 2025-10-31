# CI/CD Pipeline - MediaCheky

## 🚀 Overview

This project uses **GitHub Actions** with **semantic-release** to automate versioning, changelog generation, and Docker image builds.

## 📋 Workflow

### 1. **Conventional Commits**
All commits MUST follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Types:**
- `feat`: New feature → **Minor version bump** (1.x.0)
- `fix`: Bug fix → **Patch version bump** (1.0.x)
- `perf`: Performance improvement → **Patch version bump**
- `refactor`: Code refactoring → **Patch version bump**
- `docs`: Documentation changes → **No release**
- `chore`: Maintenance tasks → **No release**
- `style`: Code style/formatting → **No release**
- `test`: Testing changes → **No release**

**Breaking Changes:**
Add `BREAKING CHANGE:` in commit footer or `!` after type/scope → **Major version bump** (x.0.0)

**Examples:**
```bash
feat(sync): implement intelligent torrent matching
fix(ui): resolve tooltip not showing on mobile devices
perf(qbittorrent): optimize bulk torrent fetching
docs(readme): update installation instructions
refactor(models): extract StringSlice type to separate file

# Breaking change
feat(api)!: redesign API endpoints

BREAKING CHANGE: API endpoints have been restructured
```

### 2. **Branch Strategy**

| Branch | Purpose | Pre-release | Docker Tags |
|--------|---------|-------------|-------------|
| `main` | Production releases | No | `latest`, `v1.2.3`, `1.2`, `1` |
| `develop` | Development builds | `dev` | `develop`, `develop-v1.2.3-dev.1` |
| `alpha` | Alpha releases | `alpha` | `alpha`, `alpha-v1.2.3-alpha.1` |
| `beta` | Beta releases | `beta` | `beta`, `beta-v1.2.3-beta.1` |

### 3. **Semantic Versioning**

- **Major (x.0.0)**: Breaking changes
- **Minor (1.x.0)**: New features (backward compatible)
- **Patch (1.0.x)**: Bug fixes and improvements

Examples:
- `v1.0.0` → Initial release
- `v1.1.0` → Added new feature
- `v1.1.1` → Fixed bug
- `v2.0.0` → Breaking API change

### 4. **Docker Image Tagging**

Images are pushed to GitHub Container Registry (ghcr.io):

**Format:** `ghcr.io/{owner}/{repo}:{tag}`

**Tags generated:**
- `latest` (only for `main` branch)
- `{branch}` (e.g., `develop`, `alpha`)
- `v{version}` (e.g., `v1.2.3` - only for `main` branch)
- `{major}.{minor}` (e.g., `1.2` - only for `main` branch)
- `{major}` (e.g., `1` - only for `main` branch)
- `{branch}-{version}` (e.g., `develop-v1.2.3-dev.1`)
- `{branch}-{sha}` (commit tracking)

**Example for a release on `main`:**
```
ghcr.io/carcheky/mediacheky:latest
ghcr.io/carcheky/mediacheky:main
ghcr.io/carcheky/mediacheky:v1.2.3
ghcr.io/carcheky/mediacheky:1.2
ghcr.io/carcheky/mediacheky:1
ghcr.io/carcheky/mediacheky:main-abc1234
```

**Example for a pre-release on `develop`:**
```
ghcr.io/carcheky/mediacheky:develop
ghcr.io/carcheky/mediacheky:develop-v1.2.3-dev.1
ghcr.io/carcheky/mediacheky:develop-abc1234
```

## 🔄 Pipeline Jobs

### 1. **Test**
- Runs Go tests with race detection
- Generates coverage report
- Uploads to Codecov

### 2. **Lint**
- Runs `golangci-lint` with latest rules
- Enforces code quality standards

### 3. **Semantic Release**
- Analyzes commits since last release
- Determines next version based on conventional commits
- Generates CHANGELOG.md
- Creates GitHub release with release notes
- Pushes version tags

**Only runs on:**
- Push to `main`, `develop`, `alpha`, `beta`

### 4. **Docker Build & Push**
- Builds multi-arch images (amd64, arm64)
- Pushes to GitHub Container Registry
- Tags with semantic versions
- Includes build metadata (version, commit SHA, date)

**Only runs on:**
- Push events (after successful tests and lint)

## 📦 Using Docker Images

**Pull latest stable release:**
```bash
docker pull ghcr.io/carcheky/mediacheky:latest
```

**Pull specific version:**
```bash
docker pull ghcr.io/carcheky/mediacheky:v1.2.3
```

**Pull development build:**
```bash
docker pull ghcr.io/carcheky/mediacheky:develop
```

**Run container:**
```bash
docker run -d \
  -p 8000:8000 \
  -v $(pwd)/config:/config \
  -v $(pwd)/data:/data \
  ghcr.io/carcheky/mediacheky:latest
```

## 🔑 Required Secrets

Configure these in GitHub repository settings:

| Secret | Description | Required |
|--------|-------------|----------|
| `GITHUB_TOKEN` | Automatically provided by GitHub | ✅ (auto) |
| `CODECOV_TOKEN` | Token for Codecov coverage uploads | ⚠️ (optional) |

No manual Docker registry tokens needed - uses `GITHUB_TOKEN` for ghcr.io authentication.

## 📝 Changelog

The `CHANGELOG.md` file is automatically generated and updated on each release.

**Format:**
```markdown
# Changelog

## [1.2.0] - 2024-01-15

### ✨ Features
- feat(sync): implement intelligent torrent matching (#123)
- feat(ui): add dark mode support (#124)

### 🐛 Bug Fixes
- fix(files): resolve hardlink detection issue (#125)

### ⚡ Performance Improvements
- perf(db): optimize media query with indexing (#126)
```

## 🚦 Triggering a Release

**Automatic releases:**
1. Make commits following conventional commits format
2. Push to `main`, `develop`, `alpha`, or `beta` branches
3. Pipeline analyzes commits and determines version
4. Release is created automatically if there are releasable commits

**Manual release:**
Not needed - releases are fully automated based on commit messages.

## 🛠️ Local Development

**Test commit message format:**
```bash
# Install commitlint (optional)
npm install -g @commitlint/cli @commitlint/config-conventional

# Test your commit message
echo "feat(files): add new feature" | commitlint
```

**Preview next version (dry-run):**
```bash
npx semantic-release --dry-run
```

## 📚 References

- [Conventional Commits](https://www.conventionalcommits.org/)
- [Semantic Versioning](https://semver.org/)
- [semantic-release](https://github.com/semantic-release/semantic-release)
- [GitHub Actions](https://docs.github.com/en/actions)
- [GitHub Container Registry](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)

---

**Last Updated:** 2024-01-15  
**Pipeline Version:** v1.0.0
