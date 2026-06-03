#!/bin/bash
set -e

ROOT_DIR=$(pwd)
TEST_NETWORK_DIR="${ROOT_DIR}/fabric-samples/test-network"
CHANNEL_NAME="cartoriochannel"

echo "=========================================================="
echo "Iniciando Teste de Conectividade e Sincronização dos Peers"
echo "=========================================================="

if [ ! -d "$TEST_NETWORK_DIR" ]; then
    echo "Erro: O diretório test-network não foi encontrado."
    exit 1
fi

cd "$TEST_NETWORK_DIR"

# Configurar as variáveis de ambiente necessárias para usar a CLI do Fabric
export PATH=${PWD}/../bin:$PATH
export FABRIC_CFG_PATH=${PWD}/../config/

# Inclui scripts utilitários do test-network
source scripts/envVar.sh

echo "Verificando Org1..."
setGlobals 1
ORG1_INFO=$(peer channel getinfo -c $CHANNEL_NAME 2>&1 | grep "Blockchain info:")
ORG1_HASH=$(echo $ORG1_INFO | grep -o '"currentBlockHash":"[^"]*"' | cut -d'"' -f4)
ORG1_HEIGHT=$(echo $ORG1_INFO | grep -o '"height":[0-9]*' | cut -d':' -f2)
echo "Org1 -> Altura: $ORG1_HEIGHT | Hash Atual: $ORG1_HASH"

echo "Verificando Org2..."
setGlobals 2
ORG2_INFO=$(peer channel getinfo -c $CHANNEL_NAME 2>&1 | grep "Blockchain info:")
ORG2_HASH=$(echo $ORG2_INFO | grep -o '"currentBlockHash":"[^"]*"' | cut -d'"' -f4)
ORG2_HEIGHT=$(echo $ORG2_INFO | grep -o '"height":[0-9]*' | cut -d':' -f2)
echo "Org2 -> Altura: $ORG2_HEIGHT | Hash Atual: $ORG2_HASH"

echo "----------------------------------------------------------"
echo "Avaliando Resultados..."

if [ -z "$ORG1_HASH" ] || [ -z "$ORG2_HASH" ]; then
    echo "❌ FALHA: Não foi possível obter o hash de algum dos peers. Verifique se os containers estão rodando."
    exit 1
fi

if [ "$ORG1_HASH" == "$ORG2_HASH" ]; then
    echo "✅ SUCESSO: Todos os peers estão conectados, conversando via protocolo Gossip e sincronizados no mesmo Bloco ($ORG1_HEIGHT)!"
else
    echo "❌ FALHA: Os peers não estão sincronizados. Os hashes dos blocos ou as alturas divergem."
    exit 1
fi
echo "=========================================================="
exit 0
