#!/bin/bash
# API Compatibility Checker
# Usage: ./scripts/check-api-compatibility.sh [old-commit] [new-commit]

set -e

OLD_COMMIT="${1:-HEAD~1}"
NEW_COMMIT="${2:-HEAD}"

echo "Checking API compatibility between $OLD_COMMIT and $NEW_COMMIT"

# Create temporary directories
OLD_DIR=$(mktemp -d)
NEW_DIR=$(mktemp -d)

# Cleanup on exit
trap "rm -rf $OLD_DIR $NEW_DIR" EXIT

# Checkout old version
git worktree add "$OLD_DIR" "$OLD_COMMIT" 2>/dev/null || {
    echo "Error: Could not checkout $OLD_COMMIT"
    exit 1
}

# Checkout new version  
git worktree add "$NEW_DIR" "$NEW_COMMIT" 2>/dev/null || {
    echo "Error: Could not checkout $NEW_COMMIT"
    git worktree remove "$OLD_DIR" --force
    exit 1
}

# Find all Go packages
PACKAGES=$(find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" | xargs -I {} dirname {} | sort -u)

BREAKING_CHANGES=0

for pkg in $PACKAGES; do
    if [[ -d "$OLD_DIR/$pkg" && -d "$NEW_DIR/$pkg" ]]; then
        echo "Checking package $pkg..."
        
        # Run apidiff
        if ! apidiff "$OLD_DIR/$pkg" "$NEW_DIR/$pkg" 2>/dev/null; then
            echo "⚠️  Breaking changes detected in $pkg"
            BREAKING_CHANGES=1
        else
            echo "✅ $pkg is API compatible"
        fi
    fi
done

# Cleanup worktrees
git worktree remove "$OLD_DIR" --force 2>/dev/null || true
git worktree remove "$NEW_DIR" --force 2>/dev/null || true

if [[ $BREAKING_CHANGES -eq 1 ]]; then
    echo ""
    echo "❌ Breaking changes detected! Consider:"
    echo "   1. Bumping major version"
    echo "   2. Deprecating instead of removing"
    echo "   3. Adding new functions alongside old ones"
    exit 1
else
    echo ""
    echo "✅ No breaking changes detected"
    exit 0
fi