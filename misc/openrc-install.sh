cp openrc /etc/init.d/sleepy-daemon
rc-update add sleepy-daemon default
# rc-service sleepy-daemon start