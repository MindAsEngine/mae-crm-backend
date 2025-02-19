@echo off
echo Starting cleanup process...

echo Stopping containers...
docker compose down

if %ERRORLEVEL% NEQ 0 echo Failed to stop containers but continuing...

echo Removing containers...
docker rm $(docker ps -aq)

if %ERRORLEVEL% NEQ 0 echo Failed to remove containers but continuing...

echo Removing reporting service image...
docker rmi goldenhouserepo-reporting-service
docker rmi goldenhouserepo-auth-service
docker rmi goldenhouserepo-ads-integration-service
docker rmi goldenhouserepo-frontend
if %ERRORLEVEL% NEQ 0 echo Failed to remove image but continuing...

echo Starting services with docker-compose...
docker-compose up -d

echo Checking service status...
docker-compose ps
if %ERRORLEVEL% NEQ 0 (
    echo Failed to start services
    exit /b 1
)

echo All done!