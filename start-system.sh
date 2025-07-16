
#!/bin/bash

echo "=== Starting LSCC Blockchain System ==="

# Make the script executable
chmod +x run-nodes.sh

# Clean up any existing processes
echo "Cleaning up existing processes..."
killall lscc-benchmark 2>/dev/null || true
killall server 2>/dev/null || true

# Wait a moment for cleanup
sleep 2

echo "Building the project..."
make

echo "Starting blockchain nodes..."
./run-nodes.sh &
NODES_PID=$!

# Wait for nodes to start
sleep 5

echo "Starting dashboard server..."
cd web
go run server.go &
DASHBOARD_PID=$!
cd ..

echo ""
echo "=== LSCC System Started Successfully ==="
echo "Dashboard: http://0.0.0.0:5000"
echo "Node APIs:"
echo "  Node 1 (PoW):  http://0.0.0.0:8000/status"
echo "  Node 2 (PoS):  http://0.0.0.0:8002/status"
echo "  Node 3 (PoW):  http://0.0.0.0:8003/status"
echo "  Node 4 (PBFT): http://0.0.0.0:8004/status"
echo ""
echo "Press Ctrl+C to stop all services"
echo "========================================"

# Function to cleanup on exit
cleanup() {
    echo ""
    echo "Stopping all services..."
    kill $NODES_PID 2>/dev/null || true
    kill $DASHBOARD_PID 2>/dev/null || true
    killall lscc-benchmark 2>/dev/null || true
    killall server 2>/dev/null || true
    echo "All services stopped."
    exit 0
}

# Set trap for cleanup
trap cleanup INT TERM

# Wait for user interrupt
wait
