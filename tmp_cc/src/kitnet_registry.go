package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Kitnet define a estrutura de uma kitnet no ledger
type Kitnet struct {
	ID            string `json:"id"`
	Proprietario  string `json:"proprietario"`
	URIMetadados  string `json:"uriMetadados"`
	EstaAlugada   bool   `json:"estaAlugada"`
}

// Aluguel define a estrutura de um contrato de aluguel
type Aluguel struct {
	IDKitnet    string `json:"idKitnet"`
	Locatario   string `json:"locatario"`
	TempoInicio int64  `json:"tempoInicio"`
	TempoFim    int64  `json:"tempoFim"`
	EstaAtivo   bool   `json:"estaAtivo"`
}

// ContratoRegistroKitnet define o Contrato Inteligente para gerenciar kitnets
type ContratoRegistroKitnet struct {
	contractapi.Contract
}

// RegistrarKitnet adiciona uma nova kitnet ao ledger
func (s *ContratoRegistroKitnet) RegistrarKitnet(ctx contractapi.TransactionContextInterface, id string, uriMetadados string) error {
	existe, err := s.KitnetExiste(ctx, id)
	if err != nil {
		return err
	}
	if existe {
		return fmt.Errorf("a kitnet %s ja existe", id)
	}

	idCliente, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("falha ao obter identidade do cliente: %v", err)
	}

	kitnet := Kitnet{
		ID:           id,
		Proprietario: idCliente,
		URIMetadados: uriMetadados,
		EstaAlugada:  false,
	}

	kitnetJSON, err := json.Marshal(kitnet)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, kitnetJSON)
}

// CriarAluguel registra um contrato de aluguel para uma kitnet
func (s *ContratoRegistroKitnet) CriarAluguel(ctx contractapi.TransactionContextInterface, idKitnet string, locatario string, duracaoSegundos int64) error {
	kitnet, err := s.LerKitnet(ctx, idKitnet)
	if err != nil {
		return err
	}

	idCliente, _ := ctx.GetClientIdentity().GetID()
	if kitnet.Proprietario != idCliente {
		return fmt.Errorf("apenas o proprietario pode criar um aluguel para esta kitnet")
	}

	if kitnet.EstaAlugada {
		return fmt.Errorf("a kitnet %s ja esta alugada", idKitnet)
	}

	timestampTx, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("falha ao obter timestamp da transacao: %v", err)
	}
	tempoInicio := timestampTx.Seconds
	tempoFim := tempoInicio + duracaoSegundos

	aluguel := Aluguel{
		IDKitnet:    idKitnet,
		Locatario:   locatario,
		TempoInicio: tempoInicio,
		TempoFim:    tempoFim,
		EstaAtivo:   true,
	}

	aluguelJSON, err := json.Marshal(aluguel)
	if err != nil {
		return err
	}

	// Salva o aluguel usando uma chave composta ou convenção específica
	chaveAluguel := "ALUGUEL_" + idKitnet
	err = ctx.GetStub().PutState(chaveAluguel, aluguelJSON)
	if err != nil {
		return err
	}

	// Atualiza o status da kitnet
	kitnet.EstaAlugada = true
	kitnetJSON, _ := json.Marshal(kitnet)
	return ctx.GetStub().PutState(idKitnet, kitnetJSON)
}

// EncerrarAluguel termina um contrato de aluguel ativo
func (s *ContratoRegistroKitnet) EncerrarAluguel(ctx contractapi.TransactionContextInterface, idKitnet string) error {
	kitnet, err := s.LerKitnet(ctx, idKitnet)
	if err != nil {
		return err
	}

	chaveAluguel := "ALUGUEL_" + idKitnet
	aluguelJSON, err := ctx.GetStub().GetState(chaveAluguel)
	if err != nil {
		return fmt.Errorf("falha ao ler do estado mundial: %v", err)
	}
	if aluguelJSON == nil {
		return fmt.Errorf("nenhum aluguel ativo encontrado para a kitnet %s", idKitnet)
	}

	var aluguel Aluguel
	err = json.Unmarshal(aluguelJSON, &aluguel)
	if err != nil {
		return err
	}

	if !aluguel.EstaAtivo {
		return fmt.Errorf("o aluguel para a kitnet %s nao esta ativo", idKitnet)
	}

	idCliente, _ := ctx.GetClientIdentity().GetID()
	if idCliente != kitnet.Proprietario && idCliente != aluguel.Locatario {
		return fmt.Errorf("apenas o dono ou o locatario podem encerrar o aluguel")
	}

	aluguel.EstaAtivo = false
	novoAluguelJSON, _ := json.Marshal(aluguel)
	err = ctx.GetStub().PutState(chaveAluguel, novoAluguelJSON)
	if err != nil {
		return err
	}

	kitnet.EstaAlugada = false
	kitnetJSON, _ := json.Marshal(kitnet)
	return ctx.GetStub().PutState(idKitnet, kitnetJSON)
}

// TransferirPropriedade altera o dono de uma kitnet
func (s *ContratoRegistroKitnet) TransferirPropriedade(ctx contractapi.TransactionContextInterface, idKitnet string, novoProprietario string) error {
	kitnet, err := s.LerKitnet(ctx, idKitnet)
	if err != nil {
		return err
	}

	if kitnet.EstaAlugada {
		return fmt.Errorf("nao e possivel transferir a propriedade enquanto a kitnet estiver alugada")
	}

	idCliente, _ := ctx.GetClientIdentity().GetID()
	if kitnet.Proprietario != idCliente {
		return fmt.Errorf("apenas o proprietario pode transferir a propriedade")
	}

	kitnet.Proprietario = novoProprietario
	kitnetJSON, _ := json.Marshal(kitnet)
	return ctx.GetStub().PutState(idKitnet, kitnetJSON)
}

// LerKitnet recupera uma kitnet do ledger
func (s *ContratoRegistroKitnet) LerKitnet(ctx contractapi.TransactionContextInterface, id string) (*Kitnet, error) {
	kitnetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler do estado mundial: %v", err)
	}
	if kitnetJSON == nil {
		return nil, fmt.Errorf("a kitnet %s nao existe", id)
	}

	var kitnet Kitnet
	err = json.Unmarshal(kitnetJSON, &kitnet)
	if err != nil {
		return nil, err
	}

	return &kitnet, nil
}

// KitnetExiste retorna verdadeiro quando o ativo com o ID fornecido existe no estado mundial
func (s *ContratoRegistroKitnet) KitnetExiste(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	kitnetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("falha ao ler do estado mundial: %v", err)
	}

	return kitnetJSON != nil, nil
}

// GetAllKitnets retorna todas as kitnets encontradas no estado mundial
func (s *ContratoRegistroKitnet) GetAllKitnets(ctx contractapi.TransactionContextInterface) ([]*Kitnet, error) {
	// range vazio "" a "" retorna todos os elementos
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
			continue // Ignora entradas que não são Kitnets (como os registros de aluguel)
		}
		kitnets = append(kitnets, &kitnet)
	}

	return kitnets, nil
}

func main() {
	kitnetRegistryChaincode, err := contractapi.NewChaincode(&ContratoRegistroKitnet{})
	if err != nil {
		fmt.Printf("Erro ao criar o chaincode kitnet-registry: %s", err)
		return
	}

	if err := kitnetRegistryChaincode.Start(); err != nil {
		fmt.Printf("Erro ao iniciar o chaincode kitnet-registry: %s", err)
	}
}
