@echo off
REM Script to enable Trust Vault Plugin secrets engine (Windows)

setlocal

REM Variables
set PLUGIN_NAME=trust-vault
set MOUNT_PATH=%MOUNT_PATH%
if "%MOUNT_PATH%"=="" set MOUNT_PATH=trust-vault
set VAULT_ADDR=%VAULT_ADDR%
if "%VAULT_ADDR%"=="" set VAULT_ADDR=http://127.0.0.1:8200

echo Trust Vault Plugin - Enable Secrets Engine
echo ===========================================
echo.

REM Check if vault CLI is installed
where vault >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo Error: vault CLI not found
    echo Please install HashiCorp Vault CLI: https://www.vaultproject.io/downloads
    exit /b 1
)

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

REM Check if plugin is registered
echo Checking if plugin is registered...
vault plugin list secret | findstr /C:"%PLUGIN_NAME%" >nul
if %ERRORLEVEL% NEQ 0 (
    echo Error: Plugin '%PLUGIN_NAME%' is not registered
    echo Please run 'scripts\register-plugin.bat' first
    exit /b 1
)

echo Plugin is registered
echo.

REM Check if already enabled
vault secrets list | findstr /C:"%MOUNT_PATH%/" >nul
if %ERRORLEVEL% EQU 0 (
    echo Warning: Secrets engine already enabled at %MOUNT_PATH%
    set /p REPLY="Do you want to disable and re-enable it? (y/N): "
    if /i "!REPLY!"=="y" (
        echo Disabling existing secrets engine...
        vault secrets disable "%MOUNT_PATH%"
        echo Disabled
        echo.
    ) else (
        echo Keeping existing secrets engine
        exit /b 0
    )
)

REM Enable the secrets engine
echo Enabling secrets engine at %MOUNT_PATH%...
vault secrets enable -path="%MOUNT_PATH%" "%PLUGIN_NAME%"

if %ERRORLEVEL% EQU 0 (
    echo.
    echo Secrets engine enabled successfully
    echo.
    echo Mount Path: %MOUNT_PATH%
    echo Plugin: %PLUGIN_NAME%
    echo.
    echo Example usage:
    echo.
    echo # Create a wallet
    echo vault write %MOUNT_PATH%/wallets/my-eth-wallet coin_type=60
    echo.
    echo # Get wallet info
    echo vault read %MOUNT_PATH%/wallets/my-eth-wallet
    echo.
    echo # Get address for a coin type
    echo vault read %MOUNT_PATH%/wallets/my-eth-wallet/addresses/60
    echo.
    echo # Sign a transaction
    echo vault write %MOUNT_PATH%/wallets/my-eth-wallet/sign tx_data=@transaction.json
    echo.
    echo # List wallets
    echo vault list %MOUNT_PATH%/wallets
    echo.
    echo # Delete a wallet
    echo vault delete %MOUNT_PATH%/wallets/my-eth-wallet
) else (
    echo.
    echo Failed to enable secrets engine
    exit /b 1
)

endlocal
