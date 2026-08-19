@echo off
setlocal
title blind-dev-setup

echo Procurando pendrives conectados ao computador...
echo.
"%~dp0blind-dev-setup-windows-x64.exe" list-targets
set "exitCode=%ERRORLEVEL%"

echo.
if not "%exitCode%"=="0" echo Nao foi possivel concluir. Leia a mensagem acima antes de fechar.
echo Pressione qualquer tecla para fechar esta janela.
pause >nul
exit /b %exitCode%
