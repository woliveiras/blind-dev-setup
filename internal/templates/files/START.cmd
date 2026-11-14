@echo off
setlocal
call "%~dp0config\environment.cmd"
if errorlevel 1 goto error
echo Iniciando NVDA Portable.
start "" "%BLIND_DEV_ROOT%\tools\nvda\nvda.exe"
if errorlevel 1 goto error
call "%~dp0config\initialize-toolchain.cmd"
if errorlevel 1 goto error
echo Iniciando Visual Studio Code.
start "" "%BLIND_DEV_ROOT%\tools\vscode\Code.exe" "%BLIND_DEV_ROOT%\workspace"
if errorlevel 1 goto error
exit /b 0

:error
echo Nao foi possivel carregar o ambiente.
if not "%BLIND_DEV_NO_PAUSE%"=="1" pause
exit /b 1
