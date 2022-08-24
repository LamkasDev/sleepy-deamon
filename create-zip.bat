:: Settings
@echo on
set /p SLEEPY_VERSION=<app\cmd\version.txt
set SLEEPY_DAEMONS_PATH=C:\Users\LamkasDev\Desktop\sleepy\static\daemons

:: Build
cd app\cmd
go build -o ..\bin\sleepy-daemon.exe .
cd ..\..

:: Zip
del %SLEEPY_DAEMONS_PATH%\%SLEEPY_VERSION%.zip
del %SLEEPY_DAEMONS_PATH%\%SLEEPY_VERSION%-root.zip
tools\windows\7z a %SLEEPY_DAEMONS_PATH%\%SLEEPY_VERSION%.zip .\app\*
tools\windows\7z a %SLEEPY_DAEMONS_PATH%\%SLEEPY_VERSION%-root.zip .\* -x!app -x!config -x!dump -x!temp -x!current_version.txt