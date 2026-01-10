#!/bin/bash

# Script to start multiple server instances for horizontal scaling testing
# Usage: ./scripts/start-servers.sh [num_instances]

NUM_INSTANCES=${1:-4}
BASE_PORT=8081

echo "Starting $NUM_INSTANCES server instances on ports $BASE_PORT-$((BASE_PORT + NUM_INSTANCES - 1))"

# Kill any existing servers on these ports
for ((i=0; i<NUM_INSTANCES; i++)); do
    PORT=$((BASE_PORT + i))
    lsof -ti:$PORT | xargs kill -9 2>/dev/null || true
done

# Start servers
for ((i=0; i<NUM_INSTANCES; i++)); do
    PORT=$((BASE_PORT + i))
    echo "Starting server instance $((i+1)) on port $PORT"
    API_PORT=$PORT go run ./cmd/server/main.go > /tmp/tiny-bitly-server-$PORT.log 2>&1 &
    echo $! > /tmp/tiny-bitly-server-$PORT.pid
    sleep 1
done

echo ""
echo "All servers started. PIDs saved to /tmp/tiny-bitly-server-*.pid"
echo "Logs available at /tmp/tiny-bitly-server-*.log"
echo ""
echo "To stop all servers: ./scripts/stop-servers.sh"

