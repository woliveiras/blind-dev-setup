@echo off
for %%I in ("%~dp0..") do set "BLIND_DEV_ROOT=%%~fI"
if not exist "%BLIND_DEV_ROOT%\READY.txt" (
  echo O ambiente nao possui o marcador READY.txt.
  exit /b 1
)
set "MISE_CONFIG_FILE=%BLIND_DEV_ROOT%\config\mise.toml"
set "MISE_DATA_DIR=%BLIND_DEV_ROOT%\data\mise"
set "MISE_CACHE_DIR=%BLIND_DEV_ROOT%\cache\mise"
set "MISE_STATE_DIR=%BLIND_DEV_ROOT%\data\mise-state"
set "MISE_CONFIG_DIR=%BLIND_DEV_ROOT%\config\mise"
set "UV_CACHE_DIR=%BLIND_DEV_ROOT%\cache\uv"
set "PNPM_HOME=%BLIND_DEV_ROOT%\data\pnpm\home"
set "PNPM_STORE_DIR=%BLIND_DEV_ROOT%\data\pnpm\store"
set "NPM_CONFIG_USERCONFIG=%BLIND_DEV_ROOT%\config\npmrc"
set "GIT_CONFIG_GLOBAL=%BLIND_DEV_ROOT%\config\gitconfig"
set "PATH=%MISE_DATA_DIR%\shims;%BLIND_DEV_ROOT%\tools\git\cmd;%BLIND_DEV_ROOT%\tools\mise;%PNPM_HOME%;%PATH%"
exit /b 0
