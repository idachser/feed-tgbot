#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== Restarting Telegram Bot ===${NC}\n"

# Stop and remove the container
echo -e "${YELLOW}Stopping and removing containers...${NC}"
docker compose down -v bot
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Containers stopped${NC}\n"
else
    echo -e "${RED}✗ Failed to stop containers${NC}\n"
    exit 1
fi

# Build the image
echo -e "${YELLOW}Building image...${NC}"
docker compose build
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Build successful${NC}\n"
else
    echo -e "${RED}✗ Build failed${NC}\n"
    exit 1
fi

# Start the container
echo -e "${YELLOW}Starting container...${NC}"
docker compose up -d
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Container started${NC}\n"
else
    echo -e "${RED}✗ Failed to start container${NC}\n"
    exit 1
fi

# Show logs
echo -e "${YELLOW}Recent logs:${NC}"
docker compose logs --tail=20 telegram-bot

echo -e "\n${GREEN}=== Bot restarted successfully ===${NC}"
echo -e "${YELLOW}Run 'docker compose logs -f' to follow logs${NC}"
