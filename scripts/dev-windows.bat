:: Settings
@echo off

:: Build
cd ..\cmd
go build -o ..\bin\sleepy-daemon.exe .
cd ..\..

:: Run
.\dev\bin\sleepy-daemon.exe -config="dev.json"