:: Settings
@echo off

:: Build
cd app\cmd
go build -o ..\bin\sleepy-daemon.exe .
cd ..\..

:: Run
.\app\bin\sleepy-daemon.exe -config="dev.json"