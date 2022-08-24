rm /etc/init.d/sleepy-daemon
rc-service sleepy-daemon stop
rc-update del sleepy-daemon default