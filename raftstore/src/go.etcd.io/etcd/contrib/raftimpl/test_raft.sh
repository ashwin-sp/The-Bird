#!/bin/bash

echo -e "\nStarting Raft Cluster Nodes - Nodes are differentiated based on Port Numbers"

echo -e "\nSending Data to Node 1"
curl -L http://127.0.0.1:12380/my-key -XPUT -d Hello

sleep 10

echo -e "\nReading data from Node 1"
curl -L http://127.0.0.1:12380/my-key

echo -e "\nKilling Node 1"
$HOME/go/bin/goreman run stop raftimpl1

echo -e "\nReading data from Node 1 after Killing"
curl -L http://127.0.0.1:12380/my-key

echo -e "\nReading data from Node 3"
curl -L http://127.0.0.1:32380/my-key

echo -e "\nWriting World to Node 2"
curl -L http://127.0.0.1:22380/my-key -XPUT -d World

sleep 10

echo -e "\nReading data from Node 2"
curl -L http://127.0.0.1:22380/my-key

echo -e "\nStarting Node 1 again"
$HOME/go/bin/goreman run start raftimpl1

sleep 10

echo -e "\nReading from Node 1 after restarting"
curl -L http://127.0.0.1:12380/my-key

echo -e "\n"