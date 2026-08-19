@echo off
setlocal
call "%~dp0config\environment.cmd"
if errorlevel 1 goto error
echo Iniciando Visual Studio Code Portable.
start "" "%BLIND_DEV_ROOT%\tools\vscode\Code.exe" "%BLIND_DEV_ROOT%\workspace"
exit /b 0

:error
echo Nao foi possivel carregar o ambiente.
pause
exit /b 1
