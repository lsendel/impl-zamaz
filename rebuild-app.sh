#!/bin/bash

# Rebuild and restart the impl-zamaz application

echo "🔨 Rebuilding impl-zamaz application..."
cd /Users/lsendel/IdeaProjects/impl-zamaz

# Build the application
docker-compose build app

# Restart the application
docker-compose up -d app

echo "⏳ Waiting for application to start..."
sleep 5

# Test the new root endpoint
echo "🧪 Testing new root endpoint..."
curl -s http://localhost:8080/ | jq .

echo "✅ Application rebuilt and restarted!"