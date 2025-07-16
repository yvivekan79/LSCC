
#!/bin/bash


echo "Launching node1 (PoW) on port 8000"
./lscc-benchmark --config=config/config_node1.json > logs/node1.log 2>&1 &
NODE1_PID=$!
echo "Node1 PID: $NODE1_PID"

sleep 2

echo "Launching node2 (PoS) on port 8002" 
./lscc-benchmark --config=config/config_node2.json > logs/node2.log 2>&1 &
NODE2_PID=$!
echo "Node2 PID: $NODE2_PID"

sleep 2

echo "Launching node3 (PoW) on port 8003"
./lscc-benchmark --config=config/config_node3.json > logs/node3.log 2>&1 &
NODE3_PID=$!
echo "Node3 PID: $NODE3_PID"

sleep 2

echo "Launching node4 (PBFT) on port 8004"
./lscc-benchmark --config=config/config_node4.json > logs/node4.log 2>&1 &
NODE4_PID=$!
echo "Node4 PID: $NODE4_PID"

sleep 2

echo "Starting web dashboard on port 5000"
cd web && go run server.go > ../logs/dashboard.log 2>&1 &
DASHBOARD_PID=$!
echo "Dashboard PID: $DASHBOARD_PID"
cd ..

echo "=== All nodes and dashboard started successfully ==="
echo "🌐 WEB DASHBOARD: http://0.0.0.0:5000"
echo ""
echo "📡 API Endpoints:"
echo "Node1 (PoW): http://0.0.0.0:8000"
echo "Node2 (PoS): http://0.0.0.0:8002" 
echo "Node3 (PoW): http://0.0.0.0:8003"
echo "Node4 (PBFT): http://0.0.0.0:8004"
echo ""
echo "Log files:"
echo "  - Node1: logs/node1.log"
echo "  - Node2: logs/node2.log" 
echo "  - Node3: logs/node3.log"
echo "  - Node4: logs/node4.log"
echo "  - Dashboard: logs/dashboard.log"
echo ""
echo "To stop all nodes: killall lscc-benchmark"
echo "Press Ctrl+C to stop this script (nodes will continue running)"

# Keep script running to show status
wait
