:: Settings
@echo off
set SSH_KEY=C:\Users\LamkasDev\Documents\id_rsa
set SLEEPY_REMOTE_HOST=root@192.168.0.101
set SLEEPY_REMOTE_PATH=/opt/sleepy-daemon
SET QUIET=
:: Uncomment the next line to disable output and make the daemon continue running in the background
:: SET QUIET=&>/dev/null </dev/null

:: Launch service
ssh -i %SSH_KEY% %SLEEPY_REMOTE_HOST% "cd %SLEEPY_REMOTE_PATH%; sh ./service-linux.sh %QUIET%"