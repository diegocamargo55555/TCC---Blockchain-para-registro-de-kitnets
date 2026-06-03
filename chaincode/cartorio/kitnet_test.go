package main

import (
	"encoding/json"
	"testing"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockStub é um mock simplificado da interface ChaincodeStubInterface
type MockStub struct {
	shim.ChaincodeStubInterface
	mock.Mock
}

func (m *MockStub) GetState(key string) ([]byte, error) {
	args := m.Called(key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockStub) PutState(key string, value []byte) error {
	args := m.Called(key, value)
	return args.Error(0)
}

// MockContext é um mock da interface TransactionContextInterface
type MockContext struct {
	contractapi.TransactionContextInterface
	mock.Mock
}

func (m *MockContext) GetStub() shim.ChaincodeStubInterface {
	args := m.Called()
	return args.Get(0).(*MockStub)
}

func TestRegistrarEntidade(t *testing.T) {
	contract := new(SmartContract)
	ctx := new(MockContext)
	stub := new(MockStub)

	ctx.On("GetStub").Return(stub)

	// Cenário 1: Sucesso ao cadastrar Pessoa Física
	stub.On("GetState", "ENT001").Return(nil, nil).Once()

	expectedEntidade := Entidade{
		ID:               "ENT001",
		Tipo:             "PF",
		Documento:        "123",
		NomeRazaoSocial:  "Teste PF",
		EnderecoCarteira: "0x123",
		RepresentanteID:  "",
	}
	expectedJSON, _ := json.Marshal(expectedEntidade)
	stub.On("PutState", "ENT001", expectedJSON).Return(nil).Once()

	err := contract.RegistrarEntidade(ctx, "ENT001", "PF", "123", "Teste PF", "0x123", "")
	require.NoError(t, err)

	// Cenário 2: Falha ao cadastrar PJ sem representante legal
	stub.On("GetState", "ENT002").Return(nil, nil).Once()
	err = contract.RegistrarEntidade(ctx, "ENT002", "PJ", "123", "Teste PJ", "0x123", "")
	require.Error(t, err)
	require.Equal(t, "uma PJ precisa de um representante_legal_id", err.Error())
}

func TestCriarAtivo(t *testing.T) {
	contract := new(SmartContract)
	ctx := new(MockContext)
	stub := new(MockStub)

	ctx.On("GetStub").Return(stub)

	// Cenário 1: Sucesso ao criar ativo com 100% de posse (RN-AT01)
	stub.On("GetState", "KIT001").Return(nil, nil).Once()

	// Mockando a existência da Entidade
	entidadeJSON, _ := json.Marshal(Entidade{ID: "ENT001"})
	stub.On("GetState", "ENT001").Return(entidadeJSON, nil).Once()

	possesJSON := `[{"entidade_id":"ENT001", "percentual": 100.0}]`

	// Validando se ele salva no banco
	stub.On("PutState", "KIT001", mock.Anything).Return(nil).Once()

	err := contract.CriarAtivo(ctx, "KIT001", "Kitnet 1", "TK1", "2026-06-03", possesJSON)
	require.NoError(t, err)

	// Cenário 2: Falha ao criar ativo com posse inferior a 100% (Segurança RN-AT01)
	stub.On("GetState", "KIT002").Return(nil, nil).Once()
	stub.On("GetState", "ENT001").Return(entidadeJSON, nil).Once()

	possesJSON_Err := `[{"entidade_id":"ENT001", "percentual": 99.0}]`
	err = contract.CriarAtivo(ctx, "KIT002", "Kitnet 2", "TK2", "2026-06-03", possesJSON_Err)
	
	require.Error(t, err)
	require.Contains(t, err.Error(), "soma das posses deve ser exatamente 100")
}
