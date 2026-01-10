#!/bin/bash

# Script to stop all server instances started by start-servers.sh

BASE_PORT=8081
MAX_INSTANCES=10

echo "Stopping all server instances..."

for ((i=0; i<MAX_INSTANCES; i++)); do
    PORT=$((BASE_PORT + i))
    PID_FILE="/tmp/tiny-bitly-server-$PORT.pid"
    
    if [ -f "$PID_FILE" ]; then
        PID=$(cat "$PID_FILE")
        if ps -p $PID > /dev/null 2>&1; then
            echo "Stopping server on port $PORT (PID: $PID)"
            kill $PID 2>/dev/null || true
        fi
        rm -f "$PID_FILE"
    fi
    
    # Also kill any processes on the port
    lsof -ti:$PORT | xargs kill -9 2>/dev/null || true
done

echo "All servers stopped."

