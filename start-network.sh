#!/bin/bash
set -e

# Define directories
ROOT_DIR=$(pwd)
TEST_NETWORK_DIR="${ROOT_DIR}/fabric-samples/test-network"

echo "=============================================="
echo "Iniciando a Rede Hyperledger Fabric de testes"
echo "=============================================="

# Check if fabric-samples exists
if [ ! -d "$TEST_NETWORK_DIR" ]; then
    echo "A pasta fabric-samples não foi encontrada. Aguarde a instalação do Fabric terminar."
    exit 1
fi

cd "$TEST_NETWORK_DIR"

# Derruba qualquer instância anterior para garantir estado limpo
echo "1. Limpando redes antigas..."
./network.sh down

# Sobe a rede com 2 Orgs, 1 Orderer, criando o canal e habilitando CouchDB
echo "2. Subindo a rede e criando o canal 'cartoriochannel' com suporte a CouchDB..."
./network.sh up createChannel -c cartoriochannel -s couchdb

# Faz o deploy do nosso smart contract no canal
echo "3. Fazendo o deploy do chaincode 'cartorio'..."
./network.sh deployCC -ccn cartorio -ccp ../../chaincode/cartorio/ -ccl go -c cartoriochannel

echo "=============================================="
echo "Rede iniciada e Chaincode instalado com sucesso!"
echo "=============================================="
