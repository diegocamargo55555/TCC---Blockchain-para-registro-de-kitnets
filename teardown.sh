#!/bin/bash
set -e

ROOT_DIR=$(pwd)
TEST_NETWORK_DIR="${ROOT_DIR}/fabric-samples/test-network"

echo "=============================================="
echo "Derrubando a Rede Hyperledger Fabric"
echo "=============================================="

if [ -d "$TEST_NETWORK_DIR" ]; then
    cd "$TEST_NETWORK_DIR"
    ./network.sh down
    echo "Rede derrubada e volumes limpos."
else
    echo "A pasta fabric-samples não foi encontrada."
fi
