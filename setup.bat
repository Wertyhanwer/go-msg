@echo off
chcp 65001 >nul 2>&1
echo ==========================================
echo      LAUNCHING MESSAGE SYSTEM
echo ==========================================

echo Checking Docker...
docker --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Docker is not installed or running!
    echo Please install Docker Desktop and restart
    pause
    exit /b 1
)

echo Checking Docker Compose...
docker-compose --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Docker Compose not found!
    pause
    exit /b 1
)

echo.
echo Checking for running containers...
docker-compose ps -q >nul 2>&1
if %errorlevel% equ 0 (
    for /f %%i in ('docker-compose ps -q') do (
        if not "%%i"=="" (
            echo Stopping previous containers...
            docker-compose down
            goto :build
        )
    )
)
echo No running containers found.

:build
echo.
echo Building images...
docker-compose build --no-cache

echo.
echo Starting system...
docker-compose up -d

echo.
echo Waiting for services to start...
timeout /t 10 /nobreak >nul

echo.
echo ==========================================
echo      SYSTEM LAUNCHED SUCCESSFULLY!
echo ==========================================
echo.
echo   Web Interface: http://localhost:8080
echo   API Test:      http://localhost:8080/api/
echo   DB Test:       http://localhost:8080/api/testdb
echo.
echo Commands:
echo   View logs:  docker-compose logs -f
echo   Stop:       docker-compose down
echo   Cleanup:    docker-compose down -v
echo.

pause