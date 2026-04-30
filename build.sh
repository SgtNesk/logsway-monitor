#!/usr/bin/env bash
# Logsway — Build Script Multi-Platform
# Produce binari pronti per la distribuzione in dist/
#
# Uso:
#   ./build.sh           # versione "dev"
#   ./build.sh v1.2.0    # versione specifica

set -euo pipefail

VERSION="${1:-dev}"
BUILD_DIR="dist"

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'
step() { echo -e "${YELLOW}▶${NC} $*"; }
ok()   { echo -e "${GREEN}✓${NC} $*"; }

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# ── Clean ────────────────────────────────────────────────────────────────────
step "Pulizia dist/..."
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

# ── Frontend ─────────────────────────────────────────────────────────────────
step "Build frontend (React)..."
cd "$ROOT/frontend"
npm ci --silent
npm run build --silent
ok "Frontend buildato in frontend/dist/"

# ── Copy dist into server for embedding ──────────────────────────────────────
step "Copio frontend/dist nel server per embed..."
rm -rf "$ROOT/server/frontend"
mkdir -p "$ROOT/server/frontend"
cp -r "$ROOT/frontend/dist" "$ROOT/server/frontend/dist"
ok "frontend/dist copiato in server/frontend/dist"

# ── Server (single binary includes frontend) ─────────────────────────────────
step "Build server Go (con frontend embedded)..."
cd "$ROOT/server"
go mod tidy

SERVER_PLATFORMS=("linux/amd64" "linux/arm64")
for PLATFORM in "${SERVER_PLATFORMS[@]}"; do
  OS="${PLATFORM%/*}"
  ARCH="${PLATFORM#*/}"
  OUTPUT="$ROOT/${BUILD_DIR}/logsway-server-${OS}-${ARCH}"
  echo "    → $OS/$ARCH"
  CGO_ENABLED=0 GOOS="$OS" GOARCH="$ARCH" go build \
    -ldflags="-s -w -X main.Version=${VERSION}" \
    -o "$OUTPUT" .
done
ok "Server buildato"

# ── Agent ────────────────────────────────────────────────────────────────────
step "Build agent Go (multi-platform)..."
cd "$ROOT/agent"
go mod tidy

AGENT_PLATFORMS=("linux/amd64" "linux/arm64" "windows/amd64")
for PLATFORM in "${AGENT_PLATFORMS[@]}"; do
  OS="${PLATFORM%/*}"
  ARCH="${PLATFORM#*/}"
  OUTPUT="$ROOT/${BUILD_DIR}/logsway-agent-${OS}-${ARCH}"
  [[ "$OS" == "windows" ]] && OUTPUT="${OUTPUT}.exe"
  echo "    → $OS/$ARCH"
  CGO_ENABLED=0 GOOS="$OS" GOARCH="$ARCH" go build \
    -ldflags="-s -w -X main.Version=${VERSION}" \
    -o "$OUTPUT" .
done
ok "Agent buildato"

# ── Copy example config ───────────────────────────────────────────────────────
cp "$ROOT/agent/config.yaml.example" "$BUILD_DIR/agent-config.yaml.example"

# ── Checksums ─────────────────────────────────────────────────────────────────
step "Genero checksums..."
cd "$ROOT/$BUILD_DIR"
sha256sum * > checksums.txt
ok "checksums.txt creato"

# ── Summary ───────────────────────────────────────────────────────────────────
cd "$ROOT"
echo ""
echo -e "${GREEN}Build completata! Versione: ${VERSION}${NC}"
echo ""
ls -lh "$BUILD_DIR/"
echo ""
echo "Per rilasciare su GitHub:"
echo "  git tag ${VERSION} && git push origin ${VERSION}"
echo "  gh release create ${VERSION} ${BUILD_DIR}/* --title 'Logsway ${VERSION}'"
