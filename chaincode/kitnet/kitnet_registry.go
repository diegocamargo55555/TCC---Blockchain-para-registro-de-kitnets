package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type PessoaFisica struct {
	ID            string `json:"id"`
	DocType       string `json:"docType"`
	Nome          string `json:"nome"`
	DocumentoHash string `json:"documento_hash"`
	Contato       string `json:"contato"`
	DataCadastro  string `json:"data_cadastro"`
}

type Kitnet struct {
	ID                  string `json:"id"`
	DocType             string `json:"docType"`
	ProprietarioAtualID string `json:"proprietario_atual_id"`
	EnderecoRua         string `json:"endereco_rua"`
	EnderecoNumero      string `json:"endereco_numero"`
	EnderecoCEP         string `json:"endereco_cep"`
	EnderecoCidade      string `json:"endereco_cidade"`
	MetragemQuadrada    int    `json:"metragem_quadrada"`
	Status              string `json:"status"`
	DataRegistro        string `json:"data_registro"`
}

type ContratoLocacao struct {
	ID                  string `json:"id"`
	DocType             string `json:"docType"`
	KitnetID            string `json:"kitnet_id"`
	LocadorID           string `json:"locador_id"`
	LocatarioID         string `json:"locatario_id"`
	DataInicio          string `json:"data_inicio"`
	DataTermino         string `json:"data_termino"`
	ValorMensalCentavos int    `json:"valor_mensal_centavos"`
	Status              string `json:"status"`
	HashDocumentoPDF    string `json:"hash_documento_pdf"`
}

type ContratoRegistroKitnet struct {
	contractapi.Contract
}

// Helper para obter timestamp formatado
func getTxTimestampString(ctx contractapi.TransactionContextInterface) (string, error) {
	txTimestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return "", err
	}
	t := time.Unix(txTimestamp.Seconds, int64(txTimestamp.Nanos)).UTC()
	return t.Format(time.RFC3339), nil
}

func (s *ContratoRegistroKitnet) RegistrarPessoaFisica(ctx contractapi.TransactionContextInterface, id string, nome string, documentoHash string, contato string) error {
	if id == "" {
		return fmt.Errorf("o ID da pessoa e obrigatorio")
	}
	if nome == "" {
		return fmt.Errorf("o nome da pessoa e obrigatorio")
	}
	if documentoHash == "" {
		return fmt.Errorf("o hash do documento e obrigatorio")
	}
	if contato == "" {
		return fmt.Errorf("o contato e obrigatorio")
	}

	existe, err := s.PessoaExiste(ctx, id)
	if err != nil {
		return err
	}
	if existe {
		return fmt.Errorf("a pessoa %s ja existe", id)
	}

	timestamp, err := getTxTimestampString(ctx)
	if err != nil {
		return err
	}

	pessoa := PessoaFisica{
		ID:            id,
		DocType:       "pessoa",
		Nome:          nome,
		DocumentoHash: documentoHash,
		Contato:       contato, // email ou telefone ou os 2
		DataCadastro:  timestamp,
	}

	pessoaJSON, err := json.Marshal(pessoa)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, pessoaJSON)
}

func (s *ContratoRegistroKitnet) LerPessoaFisica(ctx contractapi.TransactionContextInterface, id string) (*PessoaFisica, error) {
	pessoaJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler do estado mundial: %v", err)
	}
	if pessoaJSON == nil {
		return nil, fmt.Errorf("a pessoa %s nao existe", id)
	}

	var pessoa PessoaFisica
	err = json.Unmarshal(pessoaJSON, &pessoa)
	if err != nil {
		return nil, err
	}

	return &pessoa, nil
}

func (s *ContratoRegistroKitnet) PessoaExiste(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	pessoaJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("falha ao ler do estado mundial: %v", err)
	}
	return pessoaJSON != nil, nil
}

func (s *ContratoRegistroKitnet) RegistrarKitnet(ctx contractapi.TransactionContextInterface, id string, proprietarioID string, rua string, numero string, cep string, cidade string, metragem int) error {
	if id == "" {
		return fmt.Errorf("o ID da kitnet e obrigatorio")
	}
	if proprietarioID == "" {
		return fmt.Errorf("o ID do proprietario e obrigatorio")
	}
	if rua == "" {
		return fmt.Errorf("a rua e obrigatoria")
	}
	if numero == "" {
		return fmt.Errorf("o numero e obrigatorio")
	}
	if cep == "" {
		return fmt.Errorf("o CEP e obrigatorio")
	}
	if cidade == "" {
		return fmt.Errorf("a cidade e obrigatoria")
	}
	if metragem <= 0 {
		return fmt.Errorf("a metragem quadrada deve ser maior que zero")
	}

	existe, err := s.KitnetExiste(ctx, id)
	if err != nil {
		return err
	}
	if existe {
		return fmt.Errorf("a kitnet %s ja existe", id)
	}

	propExiste, err := s.PessoaExiste(ctx, proprietarioID)
	if err != nil {
		return err
	}
	if !propExiste {
		return fmt.Errorf("proprietario %s nao registrado", proprietarioID)
	}

	timestamp, err := getTxTimestampString(ctx)
	if err != nil {
		return err
	}

	kitnet := Kitnet{
		ID:                  id,
		DocType:             "kitnet",
		ProprietarioAtualID: proprietarioID,
		EnderecoRua:         rua,
		EnderecoNumero:      numero,
		EnderecoCEP:         cep,
		EnderecoCidade:      cidade,
		MetragemQuadrada:    metragem,
		Status:              "disponivel",
		DataRegistro:        timestamp,
	}

	kitnetJSON, err := json.Marshal(kitnet)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, kitnetJSON)
}

func (s *ContratoRegistroKitnet) CriarContratoLocacao(ctx contractapi.TransactionContextInterface, idContrato string, idKitnet string, locadorID string, locatarioID string, dataInicio string, dataTermino string, valorMensal int, hashPDF string) error {
	// Verificar se contrato existe
	contratoJSON, err := ctx.GetStub().GetState(idContrato)
	if err != nil {
		return err
	}
	if contratoJSON != nil {
		return fmt.Errorf("o contrato %s ja existe", idContrato)
	}

	kitnet, err := s.LerKitnet(ctx, idKitnet)
	if err != nil {
		return err
	}

	if kitnet.ProprietarioAtualID != locadorID {
		return fmt.Errorf("apenas o proprietario atual (%s) pode criar um aluguel para esta kitnet", kitnet.ProprietarioAtualID)
	}

	if kitnet.Status == "alugada" {
		return fmt.Errorf("a kitnet %s ja esta alugada", idKitnet)
	}

	// Verificar locatario
	locExiste, err := s.PessoaExiste(ctx, locatarioID)
	if err != nil {
		return err
	}
	if !locExiste {
		return fmt.Errorf("locatario %s nao registrado", locatarioID)
	}

	contrato := ContratoLocacao{
		ID:                  idContrato,
		DocType:             "contrato",
		KitnetID:            idKitnet,
		LocadorID:           locadorID,
		LocatarioID:         locatarioID,
		DataInicio:          dataInicio,
		DataTermino:         dataTermino,
		ValorMensalCentavos: valorMensal,
		Status:              "ativo",
		HashDocumentoPDF:    hashPDF,
	}

	contratoBytes, err := json.Marshal(contrato)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(idContrato, contratoBytes)
	if err != nil {
		return err
	}

	kitnet.Status = "alugada"
	kitnetJSON, _ := json.Marshal(kitnet)
	return ctx.GetStub().PutState(idKitnet, kitnetJSON)
}

func (s *ContratoRegistroKitnet) EncerrarContratoLocacao(ctx contractapi.TransactionContextInterface, idContrato string) error {
	contratoJSON, err := ctx.GetStub().GetState(idContrato)
	if err != nil {
		return fmt.Errorf("falha ao ler do estado mundial: %v", err)
	}
	if contratoJSON == nil {
		return fmt.Errorf("nenhum contrato encontrado para %s", idContrato)
	}

	var contrato ContratoLocacao
	err = json.Unmarshal(contratoJSON, &contrato)
	if err != nil {
		return err
	}

	if contrato.Status != "ativo" {
		return fmt.Errorf("o contrato %s nao esta ativo", idContrato)
	}

	contrato.Status = "encerrado"
	novoContratoJSON, _ := json.Marshal(contrato)
	err = ctx.GetStub().PutState(idContrato, novoContratoJSON)
	if err != nil {
		return err
	}

	// Atualiza a kitnet
	kitnet, err := s.LerKitnet(ctx, contrato.KitnetID)
	if err != nil {
		return err
	}
	kitnet.Status = "disponivel"
	kitnetJSON, _ := json.Marshal(kitnet)
	return ctx.GetStub().PutState(contrato.KitnetID, kitnetJSON)
}

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

func (s *ContratoRegistroKitnet) KitnetExiste(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	kitnetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("falha ao ler do estado mundial: %v", err)
	}

	return kitnetJSON != nil, nil
}

func (s *ContratoRegistroKitnet) GetAllKitnets(ctx contractapi.TransactionContextInterface) ([]*Kitnet, error) {
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

		var temp map[string]interface{}
		json.Unmarshal(queryResponse.Value, &temp)
		if temp["docType"] == "kitnet" {
			var kitnet Kitnet
			err = json.Unmarshal(queryResponse.Value, &kitnet)
			if err != nil {
				continue
			}
			kitnets = append(kitnets, &kitnet)
		}
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
