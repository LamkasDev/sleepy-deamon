:: Settings
@echo off

:: Build
cd cmd
copy version.txt ..\..\current_version.txt
go build -o ..\bin\sleepy-daemon.exe .
cd ..