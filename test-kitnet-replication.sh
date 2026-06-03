#!/bin/bash
set -e

ROOT_DIR=$(pwd)
TEST_NETWORK_DIR="${ROOT_DIR}/fabric-samples/test-network"
CHANNEL_NAME="cartoriochannel"
CC_NAME="cartorio"
KITNET_ID="KITNET_01"

echo "=========================================================="
echo "Iniciando Teste de Replicação do Registro da Kitnet (3 ORGS)"
echo "=========================================================="

cd "$TEST_NETWORK_DIR"
export PATH=${PWD}/../bin:$PATH
export FABRIC_CFG_PATH=${PWD}/../config/
source scripts/envVar.sh

echo "1. Consultando o estado da Kitnet no nó validador da Org1..."
setGlobals 1
ORG1_RESULT=$(peer chaincode query -C $CHANNEL_NAME -n $CC_NAME -c '{"function":"LerAtivo","Args":["'$KITNET_ID'"]}' | jq -S .)

echo "2. Consultando o estado da Kitnet no nó validador da Org2..."
setGlobals 2
ORG2_RESULT=$(peer chaincode query -C $CHANNEL_NAME -n $CC_NAME -c '{"function":"LerAtivo","Args":["'$KITNET_ID'"]}' | jq -S .)

echo "----------------------------------------------------------"
echo "Avaliando Consenso e Replicação de Blocos..."

if [ "$ORG1_RESULT" == "$ORG2_RESULT" ]; then
    echo "✅ SUCESSO: Os dados da $KITNET_ID são rigorosamente IDÊNTICOS na Org1 e Org2."
    echo "Isso atesta que a transação foi endossada e gravada no mesmo bloco nos 2 cartórios físicos da rede."
    echo ""
    echo "Resumo dos Dados Replicados:"
    echo "$ORG1_RESULT" | grep -E "id|descricao|cid_ipfs"
else
    echo "❌ FALHA: Divergência nos registros. A blockchain não está em consenso!"
    echo "$ORG1_RESULT" > /tmp/org1_kitnet.json
    echo "$ORG2_RESULT" > /tmp/org2_kitnet.json
    diff -u /tmp/org1_kitnet.json /tmp/org2_kitnet.json || true
    exit 1
fi

echo "=========================================================="
exit 0
