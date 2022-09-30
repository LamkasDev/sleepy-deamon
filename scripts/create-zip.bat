:: Settings
@echo on
set /p SLEEPY_VERSION=<..\cmd\version.txt
set SLEEPY_DAEMONS_PATH=C:\Users\LamkasDev\Desktop\sleepy\data\static\daemons

:: Build
cd ..\cmd
go build -o ..\bin\sleepy-daemon.exe .
cd ..

:: Zip
del %SLEEPY_DAEMONS_PATH%\%SLEEPY_VERSION%.zip
.\tools\windows\7z a %SLEEPY_DAEMONS_PATH%\%SLEEPY_VERSION%.zip .\* -x!bin -x!config