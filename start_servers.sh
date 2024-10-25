#! /bin/bash

# build server binary
go build -o server servers/main.go

# Start three server instances
./server -port 8081 > server1.log 2>&1 &
./server -port 8082 > server2.log 2>&1 &
./server -port 8083 > server3.log 2>&1 &

echo "Started servers on ports 8081, 8082, and 8083"
echo "Check server1.log, server2.log, and server3.log for output"

# Save PIDs for later cleanup
echo $! > server3.pid
echo $(($!-1)) > server2.pid
echo $(($!-2)) > server1.pid

echo "To stop servers, run: kill \$(cat server*.pid)"
