package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing a Kitnet
type SmartContract struct {
	contractapi.Contract
}

// ...
// (skipping middle part for now to ensure exact match on main)
// Actually I'll just replace the main part and imports separately if needed.


// Kitnet describes basic details of what constitutes a kitnet registry
type Kitnet struct {
	ID             string `json:"id"`
	EnderecoFisico string `json:"endereco_fisico"`
	Proprietario   string `json:"proprietario"`
	IPFSContratoCID string `json:"ipfs_contrato_cid"`
	Status         string `json:"status"` // ex: "Ativo", "Transferido", "Encerrado"
}

// InitLedger adds a base set of kitnets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	kitnets := []Kitnet{
		{ID: "kitnet1", EnderecoFisico: "Rua A, 123", Proprietario: "Diego", IPFSContratoCID: "QmXoypizjW3WknFiJnKLwHCnL72vedxjQkDDP1mXWo6uco", Status: "Ativo"},
	}

	for _, kitnet := range kitnets {
		kitnetJSON, err := json.Marshal(kitnet)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(kitnet.ID, kitnetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

// CreateKitnet issues a new kitnet to the world state with given details.
func (s *SmartContract) CreateKitnet(ctx contractapi.TransactionContextInterface, id string, endereco string, proprietario string, cid string) error {
	exists, err := s.KitnetExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the kitnet %s already exists", id)
	}

	kitnet := Kitnet{
		ID:             id,
		EnderecoFisico: endereco,
		Proprietario:   proprietario,
		IPFSContratoCID: cid,
		Status:         "Ativo",
	}
	kitnetJSON, err := json.Marshal(kitnet)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, kitnetJSON)
}

// ReadKitnet returns the kitnet stored in the world state with given id.
func (s *SmartContract) ReadKitnet(ctx contractapi.TransactionContextInterface, id string) (*Kitnet, error) {
	kitnetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if kitnetJSON == nil {
		return nil, fmt.Errorf("the kitnet %s does not exist", id)
	}

	var kitnet Kitnet
	err = json.Unmarshal(kitnetJSON, &kitnet)
	if err != nil {
		return nil, err
	}

	return &kitnet, nil
}

// TransferKitnet updates the owner field of kitnet with given id in world state.
func (s *SmartContract) TransferKitnet(ctx contractapi.TransactionContextInterface, id string, newOwner string) error {
	kitnet, err := s.ReadKitnet(ctx, id)
	if err != nil {
		return err
	}

	kitnet.Proprietario = newOwner
	kitnetJSON, err := json.Marshal(kitnet)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, kitnetJSON)
}

// KitnetExists returns true when asset with given ID exists in world state
func (s *SmartContract) KitnetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	kitnetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return kitnetJSON != nil, nil
}

// GetAllKitnets returns all kitnets found in world state
func (s *SmartContract) GetAllKitnets(ctx contractapi.TransactionContextInterface) ([]*Kitnet, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var kitnets []*Kitnet
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var kitnet Kitnet
		err = json.Unmarshal(queryResponse.Value, &kitnet)
		if err != nil {
			return nil, err
		}
		kitnets = append(kitnets, &kitnet)
	}

	return kitnets, nil
}

func main() {
	smartContract := new(SmartContract)

	chaincode, err := contractapi.NewChaincode(smartContract)
	if err != nil {
		log.Panicf("Error creating kitnet chaincode: %v", err)
	}

	server := &shim.ChaincodeServer{
		CCID:    os.Getenv("CHAINCODE_ID"),
		Address: os.Getenv("CHAINCODE_SERVER_ADDRESS"),
		CC:      chaincode,
		TLSProps: shim.TLSProperties{
			Disabled: true,
		},
	}

	if err := server.Start(); err != nil {
		log.Panicf("Error starting kitnet chaincode server: %v", err)
	}
}
