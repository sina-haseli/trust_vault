@echo off
REM Script to register Trust Vault Plugin with HashiCorp Vault (Windows)

setlocal enabledelayedexpansion

REM Variables
set PLUGIN_NAME=trust-vault
set PLUGIN_BINARY=trust-vault-plugin.exe
set PLUGIN_DIR=%VAULT_PLUGIN_DIR%
if "%PLUGIN_DIR%"=="" set PLUGIN_DIR=C:\vault\plugins
set BUILD_DIR=bin
set VAULT_ADDR=%VAULT_ADDR%
if "%VAULT_ADDR%"=="" set VAULT_ADDR=http://127.0.0.1:8200

echo Trust Vault Plugin Registration
echo ==================================
echo.

REM Check if vault CLI is installed
where vault >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo Error: vault CLI not found
    echo Please install HashiCorp Vault CLI: https://www.vaultproject.io/downloads
    exit /b 1
)

REM Check if plugin binary exists
set PLUGIN_PATH=%BUILD_DIR%\%PLUGIN_BINARY%
if not exist "%PLUGIN_PATH%" (
    echo Error: Plugin binary not found at %PLUGIN_PATH%
    echo Please run 'build.bat' first
    exit /b 1
)

REM Check if SHA256 file exists
set SHA256_FILE=%PLUGIN_PATH%.sha256
if not exist "%SHA256_FILE%" (
    echo SHA256 file not found, calculating...
    certutil -hashfile "%PLUGIN_PATH%" SHA256 | findstr /v "hash" | findstr /v "CertUtil" > "%SHA256_FILE%"
)

REM Read SHA256
set /p SHA256=<"%SHA256_FILE%"
set SHA256=%SHA256: =%

echo Plugin binary: %PLUGIN_PATH%
echo SHA256: %SHA256%
echo.

REM Check Vault connection
echo Checking Vault connection...
vault status >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo Error: Cannot connect to Vault at %VAULT_ADDR%
    echo Please ensure Vault is running and VAULT_ADDR is correct
    exit /b 1
)

echo Connected to Vault
echo.

REM Copy plugin to plugin directory
echo Copying plugin to %PLUGIN_DIR%...
if not exist "%PLUGIN_DIR%" mkdir "%PLUGIN_DIR%"
copy /Y "%PLUGIN_PATH%" "%PLUGIN_DIR%\" >nul

echo Plugin copied
echo.

REM Register plugin with Vault
echo Registering plugin with Vault...
vault plugin register -sha256="%SHA256%" -command="%PLUGIN_BINARY%" secret "%PLUGIN_NAME%"

if %ERRORLEVEL% EQU 0 (
    echo.
    echo Plugin registered successfully
    echo.
    echo Plugin Name: %PLUGIN_NAME%
    echo Plugin Type: secret
    echo Command: %PLUGIN_BINARY%
    echo.
    echo Next steps:
    echo 1. Enable the plugin: scripts\enable-plugin.bat
    echo 2. Or manually: vault secrets enable -path=%PLUGIN_NAME% %PLUGIN_NAME%
) else (
    echo.
    echo Plugin registration failed
    exit /b 1
)

endlocal
