:: Settings
@echo off
set /p SLEEPY_VERSION=<../cmd/version.txt
set SSH_KEY=C:\Users\LamkasDev\Documents\id_rsa
set SLEEPY_REMOTE_HOST=root@192.168.0.101
set SLEEPY_REMOTE_PATH=/opt/sleepy-daemon
SET QUIET=
:: Uncomment the next line to disable output and make the daemon continue running in the background
:: SET QUIET=&>/dev/null </dev/null

:: Deploy
ssh -i %SSH_KEY% %SLEEPY_REMOTE_HOST% "rm -rf %SLEEPY_REMOTE_PATH%/%SLEEPY_VERSION%; mkdir -p %SLEEPY_REMOTE_PATH%/%SLEEPY_VERSION%"
scp -r -i %SSH_KEY% ..\go.* %SLEEPY_REMOTE_HOST%:%SLEEPY_REMOTE_PATH%/%SLEEPY_VERSION%
scp -r -i %SSH_KEY% ..\cmd %SLEEPY_REMOTE_HOST%:%SLEEPY_REMOTE_PATH%/%SLEEPY_VERSION%/cmd
scp -r -i %SSH_KEY% ..\misc %SLEEPY_REMOTE_HOST%:%SLEEPY_REMOTE_PATH%/%SLEEPY_VERSION%/misc
scp -r -i %SSH_KEY% ..\tools %SLEEPY_REMOTE_HOST%:%SLEEPY_REMOTE_PATH%/%SLEEPY_VERSION%/tools
scp -r -i %SSH_KEY% ..\scripts %SLEEPY_REMOTE_HOST%:%SLEEPY_REMOTE_PATH%/%SLEEPY_VERSION%/scripts
ssh -i %SSH_KEY% %SLEEPY_REMOTE_HOST% "cd %SLEEPY_REMOTE_PATH%/%SLEEPY_VERSION%; cd scripts; chmod -R a+rx .; ./build-linux.sh"