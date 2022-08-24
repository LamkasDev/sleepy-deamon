:: Settings
@echo off
set /p SLEEPY_VERSION=<current_version.txt

:: Run
.\%SLEEPY_VERSION%\bin\sleepy-daemon.exe -config="current.json"