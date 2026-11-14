@echo off
setlocal
title blind-dev-setup

"%~dp0blind-dev-setup-windows-x64.exe"
set "exitCode=%ERRORLEVEL%"
if "%exitCode%"=="0" exit /b 0

echo.
echo Nao foi possivel concluir. Leia a mensagem acima antes de fechar.
echo Pressione qualquer tecla para fechar esta janela.
pause >nul
exit /b %exitCode%
