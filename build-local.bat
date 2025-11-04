@echo off
REM Local build script for Trust Vault Plugin on Windows (without Docker)

echo Building Trust Vault Plugin...
echo.

REM Variables
set PLUGIN_NAME=trust-vault-plugin
set BUILD_DIR=bin
set VERSION=dev

REM Create build directory
if not exist "%BUILD_DIR%" mkdir "%BUILD_DIR%"

REM Build the plugin
echo Building for Windows...
go build -v -ldflags="-s -w" -o "%BUILD_DIR%\%PLUGIN_NAME%.exe" .\cmd\trust-vault

if %ERRORLEVEL% EQU 0 (
    echo.
    echo Build successful: %BUILD_DIR%\%PLUGIN_NAME%.exe
    echo.
    
    REM Calculate SHA256
    echo Calculating SHA256 checksum...
    certutil -hashfile "%BUILD_DIR%\%PLUGIN_NAME%.exe" SHA256 | findstr /v "hash" | findstr /v "CertUtil" > "%BUILD_DIR%\%PLUGIN_NAME%.exe.sha256"
    
    echo SHA256 checksum saved to: %BUILD_DIR%\%PLUGIN_NAME%.exe.sha256
    echo.
    echo Build complete!
) else (
    echo.
    echo Build failed!
    exit /b 1
)
