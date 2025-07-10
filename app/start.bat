@echo off
echo Starting Go Message Server...

REM Set database environment variables
set POSTGRES_HOST=localhost
set POSTGRES_PORT=5432
set POSTGRES_USER=postgres
set POSTGRES_PASSWORD=password
set POSTGRES_DB=gomsg
set API_PORT=8080

echo Database settings:
echo HOST: %POSTGRES_HOST%
echo PORT: %POSTGRES_PORT%
echo USER: %POSTGRES_USER%
echo DB: %POSTGRES_DB%
echo API_PORT: %API_PORT%

echo.
echo Starting server...
go run main.go

pause 