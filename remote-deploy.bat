:: Settings
@echo off
set /p SLEEPY_VERSION=<app/cmd/version.txt
set SSH_KEY=C:\Users\LamkasDev\Documents\id_rsa
set SLEEPY_REMOTE_HOST=root@192.168.0.101
set SLEEPY_REMOTE_PATH=/opt/sleepy-daemon
SET QUIET=
:: Uncomment the next line to disable output and make the daemon continue running in the background
:: SET QUIET=&>/dev/null </dev/null

:: Deploy
ssh -i %SSH_KEY% %SLEEPY_REMOTE_HOST% "rm %SLEEPY_REMOTE_PATH%/*; rm -rf %SLEEPY_REMOTE_PATH%/misc %SLEEPY_REMOTE_PATH%/tools %SLEEPY_REMOTE_PATH%/%SLEEPY_VERSION%"
scp -i %SSH_KEY% .gitignore %SLEEPY_REMOTE_HOST%:%SLEEPY_REMOTE_PATH%
scp -i %SSH_KEY% *.sh %SLEEPY_REMOTE_HOST%:%SLEEPY_REMOTE_PATH%
scp -i %SSH_KEY% *.bat %SLEEPY_REMOTE_HOST%:%SLEEPY_REMOTE_PATH%
scp -r -i %SSH_KEY% misc %SLEEPY_REMOTE_HOST%:%SLEEPY_REMOTE_PATH%/misc
scp -r -i %SSH_KEY% tools %SLEEPY_REMOTE_HOST%:%SLEEPY_REMOTE_PATH%/tools
scp -r -i %SSH_KEY% app %SLEEPY_REMOTE_HOST%:%SLEEPY_REMOTE_PATH%/%SLEEPY_VERSION%
ssh -i %SSH_KEY% %SLEEPY_REMOTE_HOST% "cd %SLEEPY_REMOTE_PATH%; chmod -R a+rx .; cd %SLEEPY_VERSION%; ./build-linux.sh; cd ..; sh ./launch-linux.sh %QUIET%"