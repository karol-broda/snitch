#!/bin/bash
# entrypoint script that creates fake network services for demo

set -e

echo "starting demo services..."

# start nginx on port 80
nginx &
sleep 0.5

# start some listening services with socat (stderr silenced)
socat TCP-LISTEN:8080,fork,reuseaddr SYSTEM:"echo HTTP/1.1 200 OK" 2>/dev/null &
socat TCP-LISTEN:3000,fork,reuseaddr SYSTEM:"echo hello" 2>/dev/null &
socat TCP-LISTEN:5432,fork,reuseaddr SYSTEM:"echo postgres" 2>/dev/null &
socat TCP-LISTEN:6379,fork,reuseaddr SYSTEM:"echo redis" 2>/dev/null &
socat TCP-LISTEN:27017,fork,reuseaddr SYSTEM:"echo mongo" 2>/dev/null &

# create some "established" connections by connecting to our own services
sleep 0.5
(while true; do echo "ping" | nc -q 1 localhost 8080 2>/dev/null; sleep 2; done) >/dev/null 2>&1 &
(while true; do echo "ping" | nc -q 1 localhost 3000 2>/dev/null; sleep 2; done) >/dev/null 2>&1 &
(while true; do curl -s http://localhost:80 >/dev/null 2>&1; sleep 3; done) &

# udp listeners
socat UDP-LISTEN:5353,fork,reuseaddr SYSTEM:"echo mdns" 2>/dev/null &
socat UDP-LISTEN:1900,fork,reuseaddr SYSTEM:"echo ssdp" 2>/dev/null &

sleep 1
echo "services started, recording demo..."

# run vhs to record the demo
cd /app
vhs demo.tape

echo "demo recorded, copying output..."

# output will be in /app/demo.gif
cp /app/demo.gif /output/demo.gif 2>/dev/null || echo "output copied"

echo "done!"
