package main

import (
	"encoding/json"
	"testing"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStub is a mock of shim.ChaincodeStubInterface
type MockStub struct {
	shim.ChaincodeStubInterface
	mock.Mock
}

func (m *MockStub) GetState(key string) ([]byte, error) {
	args := m.Called(key)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockStub) PutState(key string, value []byte) error {
	args := m.Called(key, value)
	return args.Error(0)
}

func (m *MockStub) GetTxTimestamp() (*timestamp.Timestamp, error) {
	return &timestamp.Timestamp{Seconds: 1620000000, Nanos: 0}, nil
}

// MockContext is a mock of contractapi.TransactionContextInterface
type MockContext struct {
	contractapi.TransactionContextInterface
	mock.Mock
}

func (m *MockContext) GetStub() shim.ChaincodeStubInterface {
	args := m.Called()
	return args.Get(0).(shim.ChaincodeStubInterface)
}

func TestRegistrarPessoaFisica(t *testing.T) {
	contract := new(ContratoRegistroKitnet)
	ctx := new(MockContext)
	stub := new(MockStub)

	ctx.On("GetStub").Return(stub)
	stub.On("GetState", "pessoa1").Return([]byte(nil), nil)
	stub.On("PutState", "pessoa1", mock.Anything).Return(nil)

	err := contract.RegistrarPessoaFisica(ctx, "pessoa1", "Diego", "hash123", "contato@email.com")
	assert.NoError(t, err)

	stub.AssertExpectations(t)
}

func TestRegistrarKitnet(t *testing.T) {
	contract := new(ContratoRegistroKitnet)
	ctx := new(MockContext)
	stub := new(MockStub)

	ctx.On("GetStub").Return(stub)
	
	// verifica se kitnet existe
	stub.On("GetState", "kitnet1").Return([]byte(nil), nil)
	
	// verifica se proprietário existe
	pessoa := PessoaFisica{ID: "prop1", DocType: "pessoa"}
	pessoaJSON, _ := json.Marshal(pessoa)
	stub.On("GetState", "prop1").Return(pessoaJSON, nil)
	
	stub.On("PutState", "kitnet1", mock.Anything).Return(nil)

	err := contract.RegistrarKitnet(ctx, "kitnet1", "prop1", "Rua A", "123", "Cidade X", 50)
	assert.NoError(t, err)
}

func TestCriarContratoLocacao(t *testing.T) {
	contract := new(ContratoRegistroKitnet)
	ctx := new(MockContext)
	stub := new(MockStub)

	ctx.On("GetStub").Return(stub)

	// Contrato não existe
	stub.On("GetState", "cont1").Return([]byte(nil), nil)
	
	// Kitnet existe e pertence ao locador
	kitnet := Kitnet{ID: "kit1", ProprietarioAtualID: "locador1", Status: "disponivel"}
	kitnetJSON, _ := json.Marshal(kitnet)
	stub.On("GetState", "kit1").Return(kitnetJSON, nil)
	
	// Locatário existe
	locatario := PessoaFisica{ID: "locatario1", DocType: "pessoa"}
	locatarioJSON, _ := json.Marshal(locatario)
	stub.On("GetState", "locatario1").Return(locatarioJSON, nil)

	stub.On("PutState", "cont1", mock.Anything).Return(nil)
	stub.On("PutState", "kit1", mock.Anything).Return(nil)

	err := contract.CriarContratoLocacao(ctx, "cont1", "kit1", "locador1", "locatario1", "2023-01-01", "2023-12-31", 100000, "pdfhash")
	assert.NoError(t, err)
}
