# Build
cd cmd
cp version.txt ../../current_version.txt
go build -o ../bin/sleepy-daemon .
cd ..