@echo off
setlocal
call "%~dp0config\environment.cmd"
if errorlevel 1 goto error
echo Iniciando NVDA Portable e Visual Studio Code.
start "" "%BLIND_DEV_ROOT%\tools\nvda\nvda.exe"
start "" "%BLIND_DEV_ROOT%\tools\vscode\Code.exe" "%BLIND_DEV_ROOT%\workspace"
exit /b 0

:error
echo Nao foi possivel carregar o ambiente.
pause
exit /b 1
