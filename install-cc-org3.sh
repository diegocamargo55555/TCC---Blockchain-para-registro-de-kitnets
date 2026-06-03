#!/bin/bash
set -e

cd /home/heilt/Documents/TCC---Blockchain-para-registro-de-kitnets/fabric-samples/test-network
export PATH=${PWD}/../bin:$PATH
export FABRIC_CFG_PATH=${PWD}/../config/
source scripts/envVar.sh

setGlobals 3
echo "Instalando o chaincode na Org3..."
peer lifecycle chaincode install cartorio.tar.gz

echo "Buscando o Package ID..."
PACKAGE_ID=$(peer lifecycle chaincode queryinstalled | grep "Package ID: cartorio_1.0" | sed -n 's/^Package ID: //; s/, Label:.*$//; p')
echo "Package ID: $PACKAGE_ID"

echo "Aprovando o chaincode para a Org3..."
ORDERER_CA=${PWD}/organizations/ordererOrganizations/example.com/tlsca/tlsca.example.com-cert.pem

peer lifecycle chaincode approveformyorg -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "$ORDERER_CA" --channelID cartoriochannel --name cartorio --version 1.0 --package-id "$PACKAGE_ID" --sequence 1

echo "Chaincode aprovado na Org3!"
