#!/bin/bash
# scripts/validate.sh - Validación completa pre-commit para MediaCheky
# Este script ejecuta todas las verificaciones necesarias antes de hacer commit

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Unicode symbols
CHECK="✅"
CROSS="❌"
WARN="⚠️"
INFO="ℹ️"
ROCKET="🚀"
TOOLS="🔧"

# Error counter
ERRORS=0

echo -e "${BLUE}${ROCKET} Running MediaCheky validation suite...${NC}\n"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"

# Function to print section header
print_section() {
    echo -e "${BOLD}${YELLOW}$1${NC}"
}

# Function to print success
print_success() {
    echo -e "${GREEN}${CHECK} $1${NC}"
}

# Function to print error
print_error() {
    echo -e "${RED}${CROSS} $1${NC}"
    ERRORS=$((ERRORS + 1))
}

# Function to print warning
print_warning() {
    echo -e "${YELLOW}${WARN} $1${NC}"
}

# Function to print info
print_info() {
    echo -e "${CYAN}${INFO} $1${NC}"
}

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 1. GO FORMAT CHECK
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
print_section "📝 Checking Go code format..."

UNFORMATTED=$(gofmt -s -l . 2>&1 | grep -v '^vendor/' | grep -v '^volumes/' | grep -v '^reference-repos/' | grep '.go$' || true)
if [ -n "$UNFORMATTED" ]; then
    print_error "The following files need formatting:"
    echo "$UNFORMATTED" | while read -r file; do
        echo "    - $file"
    done
    echo ""
    print_info "💡 Run 'make lint-fix' or 'gofmt -s -w .' to fix"
    echo ""
else
    print_success "All files properly formatted"
fi
echo ""

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 2. GO VET
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
print_section "🔍 Running go vet..."

# Get packages, excluding volumes/ and reference-repos/
GO_PACKAGES=$(cd $(dirname $0)/.. && find . -name "*.go" -not -path "./volumes/*" -not -path "./reference-repos/*" -not -path "./vendor/*" -exec dirname {} \; | sort -u | sed 's|^\./|./|' | grep -v "^\.$")

if [ -z "$GO_PACKAGES" ]; then
    print_error "No Go packages found"
else
    if go vet $GO_PACKAGES 2>&1; then
        print_success "Go vet passed"
    else
        print_error "Go vet found issues"
    fi
fi
echo ""

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 3. TESTS WITH COVERAGE
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
print_section "🧪 Running tests with coverage and race detector..."

# Get test packages using Go's module system, excluding volumes/ and reference-repos/
TEST_PACKAGES=$(go list ./... 2>/dev/null | grep -v '/volumes/' | grep -v '/reference-repos/' | grep -v '/vendor/' || echo "")

if [ -z "$TEST_PACKAGES" ]; then
    print_warning "No test packages found"
else
    if go test -v -race -coverprofile=coverage.out $TEST_PACKAGES 2>&1; then
        # Calculate coverage
        if [ -f coverage.out ]; then
            COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
            print_success "Tests passed - Coverage: ${COVERAGE}"
            
            # Warn if coverage is low
            COVERAGE_NUM=$(echo $COVERAGE | sed 's/%//')
            if (( $(echo "$COVERAGE_NUM < 50" | bc -l) )); then
                print_warning "Coverage is below 50% - consider adding more tests"
            fi
        else
            print_success "Tests passed"
        fi
    else
        print_error "Tests failed"
    fi
fi
echo ""

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 4. BUILD CHECK
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
print_section "🔨 Checking build..."

if CGO_ENABLED=1 go build -o /tmp/mediacheky-test ./cmd/server 2>&1; then
    rm -f /tmp/mediacheky-test
    print_success "Build successful"
else
    print_error "Build failed"
fi
echo ""

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 5. GO MOD TIDY CHECK
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
print_section "📦 Checking go.mod and go.sum..."

# Create backup
cp go.mod go.mod.backup
cp go.sum go.sum.backup

# Run go mod tidy
go mod tidy 2>&1 > /dev/null

# Check for differences
if ! diff -q go.mod go.mod.backup > /dev/null 2>&1 || ! diff -q go.sum go.sum.backup > /dev/null 2>&1; then
    print_error "go.mod or go.sum needs tidying"
    print_info "💡 Run 'go mod tidy' and commit changes"
    echo ""
    print_info "Differences found:"
    diff -u go.mod.backup go.mod || true
    
    # Restore backup
    mv go.mod.backup go.mod
    mv go.sum.backup go.sum
else
    print_success "Dependencies are clean"
    rm go.mod.backup go.sum.backup
fi
echo ""

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 6. LINTING (golangci-lint in CI, native tools in local)
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
print_section "🔎 Running linter checks..."

# Check if we're in CI environment (GitHub Actions)
if [ -n "$CI" ] || [ -n "$GITHUB_ACTIONS" ]; then
    # In CI: golangci-lint is MANDATORY
    print_info "🤖 CI environment detected - using golangci-lint"
    if command -v golangci-lint &> /dev/null; then
        if golangci-lint run ./... 2>&1; then
            print_success "golangci-lint passed"
        else
            print_error "golangci-lint found issues"
        fi
    else
        print_error "golangci-lint not found in CI environment!"
        print_info "💡 Install in CI: https://golangci-lint.run/docs/welcome/install/#github-actions"
    fi
else
    # In local: Try golangci-lint, fallback to native Go tools
    if command -v golangci-lint &> /dev/null; then
        print_info "✨ Using golangci-lint (recommended)"
        if golangci-lint run ./... 2>&1; then
            print_success "golangci-lint passed"
        else
            print_error "golangci-lint found issues"
        fi
    else
        print_warning "golangci-lint not installed locally"
        print_info "🔧 Using native Go tools as fallback:"
        echo ""
        
        # Run basic linting with native tools
        LINT_ERRORS=0
        
        # 1. Check for common issues with go vet (already done, but mention it)
        print_info "  • go vet: ✓ (already validated)"
        
        # 2. Check for shadowed variables
        print_info "  • Checking for shadowed variables..."
        if command -v shadow &> /dev/null; then
            if go vet -vettool=$(which shadow) ./... 2>&1; then
                echo "    ${CHECK} No shadowed variables"
            else
                echo "    ${WARN} Shadowed variables found"
            fi
        else
            echo "    ${INFO} shadow not installed (optional)"
            echo "    ${INFO} Install: go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@latest"
        fi
        
        # 3. Check for unused code with staticcheck (if available)
        print_info "  • Checking for unused code..."
        if command -v staticcheck &> /dev/null; then
            if staticcheck ./... 2>&1; then
                echo "    ${CHECK} No unused code"
            else
                echo "    ${WARN} Unused code found"
                LINT_ERRORS=$((LINT_ERRORS + 1))
            fi
        else
            echo "    ${INFO} staticcheck not installed (optional)"
            echo "    ${INFO} Install: go install honnef.co/go/tools/cmd/staticcheck@latest"
        fi
        
        # 4. Check for inefficient assignments
        print_info "  • Checking code with go vet composites..."
        if go vet -composites ./... 2>&1 | grep -v "no packages" > /dev/null; then
            echo "    ${WARN} Composite literal issues found"
        else
            echo "    ${CHECK} No composite literal issues"
        fi
        
        echo ""
        if [ $LINT_ERRORS -eq 0 ]; then
            print_success "Native linting checks passed"
        else
            print_warning "Some linting issues found (not critical for local development)"
        fi
        
        echo ""
        print_info "💡 For complete linting, install golangci-lint:"
        print_info "   Linux:  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$(go env GOPATH)/bin"
        print_info "   macOS:  brew install golangci-lint"
        print_info "   Other:  https://golangci-lint.run/usage/install/"
    fi
fi
echo ""

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# SUMMARY
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

if [ $ERRORS -eq 0 ]; then
    echo -e "${GREEN}${BOLD}"
    echo "╔════════════════════════════════════════════╗"
    echo "║  ✅ ALL VALIDATION CHECKS PASSED!         ║"
    echo "╚════════════════════════════════════════════╝"
    echo -e "${NC}"
    echo -e "${BLUE}👍 Ready to commit and push${NC}"
    echo ""
    exit 0
else
    echo -e "${RED}${BOLD}"
    echo "╔════════════════════════════════════════════╗"
    echo "║  ❌ VALIDATION FAILED ($ERRORS error(s))          ║"
    echo "╚════════════════════════════════════════════╝"
    echo -e "${NC}"
    echo -e "${YELLOW}Please fix the issues above before committing.${NC}"
    echo ""
    echo -e "${CYAN}Quick fixes:${NC}"
    echo "  • Format issues:  make lint-fix"
    echo "  • Auto-fix all:   make check-and-fix"
    echo "  • Run tests:      make test"
    echo ""
    exit 1
fi
