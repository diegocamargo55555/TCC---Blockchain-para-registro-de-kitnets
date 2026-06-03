#!/bin/bash
set -e

ROOT_DIR=$(pwd)
TEST_NETWORK_DIR="${ROOT_DIR}/fabric-samples/test-network"
CHANNEL_NAME="cartoriochannel"
CC_NAME="cartorio"

echo "=========================================================="
echo "Iniciando Transações no Smart Contract (Registro de Kitnet)"
echo "=========================================================="

cd "$TEST_NETWORK_DIR"
export PATH=${PWD}/../bin:$PATH
export FABRIC_CFG_PATH=${PWD}/../config/
source scripts/envVar.sh

# Usa Org1 para submeter as transações
setGlobals 1

ORDERER_CA=${PWD}/organizations/ordererOrganizations/example.com/tlsca/tlsca.example.com-cert.pem
PEER0_ORG1_CA=${PWD}/organizations/peerOrganizations/org1.example.com/tlsca/tlsca.org1.example.com-cert.pem
PEER0_ORG2_CA=${PWD}/organizations/peerOrganizations/org2.example.com/tlsca/tlsca.org2.example.com-cert.pem

# As transações precisam ser endossadas por Org1 e Org2, então passamos os endereços e certificados de ambos
PEER_CONN_PARMS="--peerAddresses localhost:7051 --tlsRootCertFiles $PEER0_ORG1_CA --peerAddresses localhost:9051 --tlsRootCertFiles $PEER0_ORG2_CA"

echo "1. Registrando uma nova Entidade (Dono da Kitnet)..."
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "$ORDERER_CA" -C $CHANNEL_NAME -n $CC_NAME $PEER_CONN_PARMS -c '{"function":"RegistrarEntidade","Args":["ENT001","PF","123.456.789-00","João da Silva","0xCarteiraJoao",""]}'
sleep 3

echo "2. Registrando a Kitnet na Blockchain (100% de posse para ENT001)..."
# PossesJSON: [{"entidade_id":"ENT001","percentual":100.0}]
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "$ORDERER_CA" -C $CHANNEL_NAME -n $CC_NAME $PEER_CONN_PARMS -c '{"function":"CriarAtivo","Args":["KITNET_01","Kitnet 101 - Centro","TOKEN_KN01","2026-06-01","[{\"entidade_id\":\"ENT001\",\"percentual\":100.0}]"]}'
sleep 3

echo "3. Inserindo o PDF da Planta Baixa no Histórico (Averbação com Hash IPFS)..."
# CID Falso simulando o IPFS: QmHashIPFSExemploPlantaBaixa123
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "$ORDERER_CA" -C $CHANNEL_NAME -n $CC_NAME $PEER_CONN_PARMS -c '{"function":"AverbarImovel","Args":["KITNET_01","AVB001","construcao","Registro da Planta Baixa da Kitnet","QmHashIPFSExemploPlantaBaixa123","2026-06-01"]}'
sleep 3

echo "4. Consultando o estado final da Kitnet no Ledger..."
peer chaincode query -C $CHANNEL_NAME -n $CC_NAME -c '{"function":"LerAtivo","Args":["KITNET_01"]}' | jq .

echo "=========================================================="
echo "Operações realizadas com sucesso na Blockchain!"
echo "=========================================================="
