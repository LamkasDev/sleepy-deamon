# Build
cd app/cmd
go build -o ../bin/sleepy-daemon .
cd ../..

# Run
killall sleepy-daemon
./app/bin/sleepy-daemon -config="dev.json"