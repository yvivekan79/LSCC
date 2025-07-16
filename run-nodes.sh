#!/bin/bash

echo Running nodes
# Launch 4 LSCC nodes with different configs
echo "Launching node1 on port 8000"
./lscc-benchmark --config=config/config_node1.json > logs/node1.log 2>&1 &

echo "Launching node2 on port 8001"
./lscc-benchmark --config=config/config_node2.json > logs/node2.log 2>&1 &

echo "Launching node3 on port 8002"
./lscc-benchmark --config=config/config_node3.json > logs/node3.log 2>&1 &


echo "Launching node3 on port 8002"
./lscc-benchmark --config=config/config_node4.json > logs/node4.log 2>&1 &

