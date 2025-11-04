@echo off
REM Build script for Trust Vault plugin using Docker on Windows

echo Building Trust Vault plugin with Docker...

docker-compose run --rm build

if %ERRORLEVEL% EQU 0 (
    echo Build successful! Binary: trust-vault-plugin
) else (
    echo Build failed!
    exit /b 1
)
