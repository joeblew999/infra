#!/bin/bash

# Fly.io Deployment Script for Infrastructure Management System
# This script handles the complete deployment pipeline

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
APP_NAME="${FLY_APP_NAME:-infra-mgmt}"
REGION="${FLY_REGION:-syd}"

echo -e "${BLUE}🚀 Starting Fly.io deployment for ${APP_NAME}${NC}"

# Check if flyctl is available
if ! command -v flyctl &> /dev/null; then
    echo -e "${YELLOW}📦 Installing flyctl...${NC}"
    go run . flyctl version || {
        echo -e "${RED}❌ Failed to install flyctl${NC}"
        exit 1
    }
    export PATH="$PWD/.dep:$PATH"
fi

# Authenticate with Fly.io (if needed)
echo -e "${BLUE}🔐 Checking Fly.io authentication...${NC}"
if ! flyctl auth whoami &> /dev/null; then
    echo -e "${YELLOW}Please authenticate with Fly.io:${NC}"
    flyctl auth login
fi

# Create app if it doesn't exist
echo -e "${BLUE}📱 Checking if app ${APP_NAME} exists...${NC}"
if ! flyctl apps show "${APP_NAME}" &> /dev/null; then
    echo -e "${YELLOW}🏗️  Creating new Fly.io app: ${APP_NAME}${NC}"
    flyctl apps create "${APP_NAME}" --generate-name=false
    
    # Set secrets
    echo -e "${BLUE}🔑 Setting up secrets...${NC}"
    flyctl secrets set ENVIRONMENT=production -a "${APP_NAME}"
fi

# Create volume if it doesn't exist
echo -e "${BLUE}💾 Checking for persistent volume...${NC}"
if ! flyctl volumes list -a "${APP_NAME}" | grep -q "infra_data"; then
    echo -e "${YELLOW}📦 Creating persistent volume...${NC}"
    flyctl volumes create infra_data --size 1 --region "${REGION}" -a "${APP_NAME}"
fi

# Deploy with ko (containerless deployment)
echo -e "${BLUE}🏗️  Building and deploying with ko...${NC}"
export KO_DOCKER_REPO="registry.fly.io/${APP_NAME}"
export ENVIRONMENT=production

# Build and push container image
echo -e "${YELLOW}🔨 Building container image...${NC}"
IMAGE=$(go run . ko build --platform=linux/amd64 github.com/joeblew999/infra)

if [ -z "${IMAGE}" ]; then
    echo -e "${RED}❌ Failed to build container image${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Built image: ${IMAGE}${NC}"

# Update fly.toml with the image
echo -e "${BLUE}📝 Updating fly.toml with image...${NC}"
sed -i.bak "s|# Build configuration - using ko for container builds|image = \"${IMAGE}\"|" fly.toml

# Deploy to Fly.io
echo -e "${BLUE}🚀 Deploying to Fly.io...${NC}"
flyctl deploy -a "${APP_NAME}" --remote-only

# Restore fly.toml
mv fly.toml.bak fly.toml

# Show deployment status
echo -e "${BLUE}📊 Checking deployment status...${NC}"
flyctl status -a "${APP_NAME}"

# Show app URL
APP_URL="https://${APP_NAME}.fly.dev"
echo -e "${GREEN}✅ Deployment complete!${NC}"
echo -e "${GREEN}🌐 Your app is available at: ${APP_URL}${NC}"

# Optional: Open in browser
if command -v open &> /dev/null; then
    read -p "Open app in browser? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        open "${APP_URL}"
    fi
fi

echo -e "${BLUE}📝 Useful commands:${NC}"
echo "  View logs: flyctl logs -a ${APP_NAME}"
echo "  SSH into app: flyctl ssh console -a ${APP_NAME}"
echo "  Scale app: flyctl scale count 2 -a ${APP_NAME}"
echo "  App dashboard: https://fly.io/apps/${APP_NAME}"