#!/bin/bash

# Cores para o output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo "=========================================================="
echo "Iniciando Testes de Registro e Sincronização Kitnet"
echo "=========================================================="

# 1. Verificar Containers
echo -n "1. Verificando containers... "
containers=("peer0.org1.example.com" "peer1.org1.example.com" "peer2.org1.example.com" "orderer.example.com" "kitnet-chaincode")
for c in "${containers[@]}"; do
    if [ ! "$(docker ps -q -f name=$c)" ]; then echo -e "\n${RED}[ERRO]${NC} $c offline"; exit 1; fi
done
echo -e "${GREEN}OK${NC}"

# 2. Registrar uma nova Kitnet para teste de propagação
ID_TESTE="test_$(date +%s)"
echo -n "2. Registrando nova Kitnet ($ID_TESTE) via Peer 0... "
REGISTER_CMD="peer chaincode invoke -o orderer.example.com:7050 --channelID sys-channel --name kitnet -c '{\"Args\":[\"CreateKitnet\",\"$ID_TESTE\",\"Rua de Teste, 999\",\"Dono Teste\",\"CID_TESTE_123\",\"Disponivel\"]}' --peerAddresses peer0.org1.example.com:7051"

if docker exec cli bash -c "$REGISTER_CMD" 2>&1 | grep -q "status:200"; then
    echo -e "${GREEN}OK${NC}"
else
    echo -e "${RED}FALHA NO REGISTRO${NC}"
    exit 1
fi

# Aguardar um momento para a propagação do bloco
echo "   Aguardando 3 segundos para propagação do bloco..."
sleep 3

# 3. Verificar Sincronização em TODOS os Peers
echo "3. Verificando presença do registro em todos os Peers:"
peers=("peer0.org1.example.com:7051" "peer1.org1.example.com:8051" "peer2.org1.example.com:9051")

for p in "${peers[@]}"; do
    echo -n "   - Consultando $p: "
    QUERY_CMD="CORE_PEER_ADDRESS=$p peer chaincode query -C sys-channel -n kitnet -c '{\"Args\":[\"ReadKitnet\",\"$ID_TESTE\"]}'"
    RESULT=$(docker exec cli bash -c "$QUERY_CMD" 2>/dev/null)
    
    if echo $RESULT | grep -q "$ID_TESTE"; then
        echo -e "${GREEN}DADO SINCRONIZADO${NC}"
    else
        echo -e "${RED}DADO NÃO ENCONTRADO${NC}"
        exit 1
    fi
done

# 4. Verificar Altura do Ledger (Blockchain Height)
echo "4. Verificando altura do Ledger (Blocos) em todos os nós:"
for name in "peer0.org1.example.com" "peer1.org1.example.com" "peer2.org1.example.com"; do
    HEIGHT=$(docker logs $name 2>&1 | grep "Committed block" | tail -n 1 | sed -n 's/.*Committed block \[\([0-9]*\)\].*/\1/p')
    echo -e "   - $name: Altura do Ledger = ${GREEN}Bloco $HEIGHT${NC}"
done

echo "=========================================================="
echo "Sucesso: O registro foi replicado em todos os nós da rede."
echo "=========================================================="
