# Hyperledger Fabric Network for Kitnet Registry

This directory contains the Docker configuration for a standard Hyperledger Fabric network (v2.5) with:
- 2 Organizations (Org1, Org2)
- 1 Peer per Org
- 1 Raft Orderer
- Fabric CA for each Org

## Prerequisites
- Docker and Docker Compose
- Fabric Binaries (cryptogen, configtxgen)

## Setup Instructions
1. **Generate Crypto Materials**: Use `cryptogen` with `crypto-config.yaml`.
2. **Generate Genesis Block and Channel Artifacts**: Use `configtxgen` with `configtx.yaml`.
3. **Start the Network**:
   ```bash
   cd docker
   docker-compose -f docker-compose-test-net.yaml up -d
   ```

## Chaincode
The chaincode is located in `/chaincode/kitnet`. You can package and install it using the peer CLI once the network is up.
