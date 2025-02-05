#!/bin/bash

# Stop all docker-compose services
docker-compose down

# Remove all images except for database data
docker rmi mae-crm-backend-reporting-service
docker rmi mae-crm-backend-frontend
docker rmi mae-crm-backend-ads-integration-service
docker rmi mae-crm-backend-auth-service

# Fetch and pull the latest changes from the main repository
git fetch --all
git pull

# Fetch and pull the latest changes from the second repository
cd mae-crm-frontend
git fetch --all
git pull
cd ..

# Start docker-compose services
docker-compose up -d