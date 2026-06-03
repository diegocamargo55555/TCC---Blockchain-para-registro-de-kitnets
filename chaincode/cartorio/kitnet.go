package main

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing a Cartorio
type SmartContract struct {
	contractapi.Contract
}

// Entidade representa uma Pessoa Física (PF) ou Jurídica (PJ)
type Entidade struct {
	ID               string `json:"id"`
	Tipo             string `json:"tipo"` // "PF" ou "PJ"
	Documento        string `json:"documento"` // CPF ou CNPJ
	NomeRazaoSocial  string `json:"nome_razao_social"`
	EnderecoCarteira string `json:"endereco_carteira"`
	RepresentanteID  string `json:"representante_legal_id,omitempty"` // Apenas para PJ
}

// PosseFracionada representa a fração ideal de um dono sobre o imóvel
type PosseFracionada struct {
	EntidadeID string  `json:"entidade_id"`
	Percentual float64 `json:"percentual"` // Ex: 50.0 representa 50%
}

// Gravame representa uma restrição (ex: alienação fiduciária, penhora)
type Gravame struct {
	ID       string `json:"id"`
	Tipo     string `json:"tipo"`
	CredorID string `json:"credor_id"`
	Ativo    bool   `json:"ativo"` // Se true, bloqueia venda
}

// Averbacao representa o historico de modificacoes no imovel
type Averbacao struct {
	ID        string `json:"id"`
	Tipo      string `json:"tipo_averbacao"`
	Descricao string `json:"descricao"`
	CidIPFS   string `json:"cid_ipfs"` // Hash do documento no IPFS
	Data      string `json:"data_averbacao"`
}

// Ativo representa a Kitnet / Imóvel
type Ativo struct {
	ID           string            `json:"id"`
	Tipo         string            `json:"tipo"` // "imovel"
	Descricao    string            `json:"descricao"`
	TokenID      string            `json:"token_id_blockchain"`
	Posses       []PosseFracionada `json:"posses"`
	Gravames     []Gravame         `json:"gravames"`
	Averbacoes   []Averbacao       `json:"averbacoes"` // Histórico
	DataRegistro string            `json:"data_registro"`
}

// ParticipanteTransacao
type ParticipanteTransacao struct {
	EntidadeID string  `json:"entidade_id"`
	Papel      string  `json:"papel"` // "vendedor" ou "comprador"
	Percentual float64 `json:"percentual_fracao"`
	Assinou    bool    `json:"assinou"` // RN-TR01
}

// TransacaoEscrow representa a intencao de compra/venda retida
type TransacaoEscrow struct {
	ID            string                  `json:"id"`
	AtivoID       string                  `json:"ativo_id"`
	Participantes []ParticipanteTransacao `json:"participantes"`
	Status        string                  `json:"status"` // "pendente_assinatura", "pendente_pagamento", "concluida", "cancelada"
	DataInicio    string                  `json:"data_inicio"`
	ITBIPago      bool                    `json:"itbi_pago"` // RN-OR01
}

// InitLedger é usado apenas para inicializar o banco com dados de teste, se necessário
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	return nil
}

// RegistrarEntidade cria um novo cadastro de cidadão ou empresa
func (s *SmartContract) RegistrarEntidade(ctx contractapi.TransactionContextInterface, id string, tipo string, documento string, nome string, carteira string, representanteID string) error {
	exists, err := s.EntidadeExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("a entidade %s já existe", id)
	}

	if tipo == "PJ" && representanteID == "" {
		return fmt.Errorf("uma PJ precisa de um representante_legal_id")
	}

	entidade := Entidade{
		ID:               id,
		Tipo:             tipo,
		Documento:        documento,
		NomeRazaoSocial:  nome,
		EnderecoCarteira: carteira,
		RepresentanteID:  representanteID,
	}

	entidadeJSON, err := json.Marshal(entidade)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, entidadeJSON)
}

// CriarAtivo registra uma nova kitnet no sistema (100% para um dono inicial ou fracionado)
func (s *SmartContract) CriarAtivo(ctx contractapi.TransactionContextInterface, id string, descricao string, tokenID string, dataRegistro string, possesJSON string) error {
	exists, err := s.AtivoExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("o ativo %s já existe", id)
	}

	var posses []PosseFracionada
	err = json.Unmarshal([]byte(possesJSON), &posses)
	if err != nil {
		return fmt.Errorf("erro ao converter possesJSON: %v", err)
	}

	// Regra de Negócio: RN-POS-01 (Validação de 100%)
	totalPosse := 0.0
	for _, p := range posses {
		// Validar se entidade existe
		entExists, _ := s.EntidadeExists(ctx, p.EntidadeID)
		if !entExists {
			return fmt.Errorf("a entidade %s nao existe no sistema", p.EntidadeID)
		}
		totalPosse += p.Percentual
	}

	// Usando uma margem de erro pequena para floats
	if math.Abs(totalPosse-100.0) > 0.001 {
		return fmt.Errorf("soma das posses deve ser exatamente 100%%. Atual: %f", totalPosse)
	}

	ativo := Ativo{
		ID:           id,
		Tipo:         "imovel",
		Descricao:    descricao,
		TokenID:      tokenID,
		Posses:       posses,
		Gravames:     []Gravame{}, // Inicialmente sem gravames
		Averbacoes:   []Averbacao{},
		DataRegistro: dataRegistro,
	}

	ativoBytes, err := json.Marshal(ativo)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, ativoBytes)
}

// AverbarImovel anexa um novo registro (e seu cid_ipfs) ao histórico do ativo (RN-IM01, RN-IM03)
func (s *SmartContract) AverbarImovel(ctx contractapi.TransactionContextInterface, ativoID string, averbacaoID string, tipo string, descricao string, cidIPFS string, dataAverbacao string) error {
	ativo, err := s.LerAtivo(ctx, ativoID)
	if err != nil {
		return err
	}

	// Histórico Append-Only (não substitui, apenas adiciona)
	averbacao := Averbacao{
		ID:        averbacaoID,
		Tipo:      tipo,
		Descricao: descricao,
		CidIPFS:   cidIPFS,
		Data:      dataAverbacao,
	}

	ativo.Averbacoes = append(ativo.Averbacoes, averbacao)
	ativoBytes, _ := json.Marshal(ativo)
	return ctx.GetStub().PutState(ativoID, ativoBytes)
}


// RegistrarGravame adiciona uma restrição ao ativo (RN-GRV)
func (s *SmartContract) RegistrarGravame(ctx contractapi.TransactionContextInterface, ativoID string, gravameID string, tipoGravame string, credorID string) error {
	ativo, err := s.LerAtivo(ctx, ativoID)
	if err != nil {
		return err
	}

	gravame := Gravame{
		ID:       gravameID,
		Tipo:     tipoGravame,
		CredorID: credorID,
		Ativo:    true,
	}

	ativo.Gravames = append(ativo.Gravames, gravame)
	ativoBytes, _ := json.Marshal(ativo)
	return ctx.GetStub().PutState(ativoID, ativoBytes)
}

// BaixarGravame libera o imóvel da restrição
func (s *SmartContract) BaixarGravame(ctx contractapi.TransactionContextInterface, ativoID string, gravameID string, credorCallerID string) error {
	ativo, err := s.LerAtivo(ctx, ativoID)
	if err != nil {
		return err
	}

	encontrado := false
	for i, g := range ativo.Gravames {
		if g.ID == gravameID {
			// Regra: Apenas o credor pode baixar
			if g.CredorID != credorCallerID {
				return fmt.Errorf("apenas o credor original pode baixar este gravame")
			}
			ativo.Gravames[i].Ativo = false
			encontrado = true
			break
		}
	}

	if !encontrado {
		return fmt.Errorf("gravame nao encontrado")
	}

	ativoBytes, _ := json.Marshal(ativo)
	return ctx.GetStub().PutState(ativoID, ativoBytes)
}

// IniciarTransacaoEscrow cria uma intenção de transferência retida (Escrow)
func (s *SmartContract) IniciarTransacaoEscrow(ctx contractapi.TransactionContextInterface, txID string, ativoID string, participantesJSON string, dataInicio string) error {
	// Validar se Ativo tem gravames (RN-GR01 - Trava de Transferência)
	ativo, err := s.LerAtivo(ctx, ativoID)
	if err != nil {
		return err
	}

	for _, g := range ativo.Gravames {
		if g.Ativo {
			return fmt.Errorf("impossível iniciar transacao: o ativo possui um gravame ativo (ID: %s, Tipo: %s)", g.ID, g.Tipo)
		}
	}

	var participantes []ParticipanteTransacao
	err = json.Unmarshal([]byte(participantesJSON), &participantes)
	if err != nil {
		return fmt.Errorf("erro ao converter participantesJSON: %v", err)
	}

	// Regra de Equilíbrio (RN-AT03): A soma das frações vendidas deve ser igual a soma das compradas
	totalVenda := 0.0
	totalCompra := 0.0

	for _, p := range participantes {
		if p.Papel == "vendedor" {
			// Verifica se vendedor tem saldo (RN-AT02)
			saldoAtual := s.getSaldoPosse(ativo, p.EntidadeID)
			if saldoAtual < p.Percentual {
				return fmt.Errorf("o vendedor %s não possui saldo suficiente (%f < %f)", p.EntidadeID, saldoAtual, p.Percentual)
			}
			totalVenda += p.Percentual
		} else if p.Papel == "comprador" {
			totalCompra += p.Percentual
		} else {
			return fmt.Errorf("papel invalido: %s", p.Papel)
		}
	}

	if math.Abs(totalVenda-totalCompra) > 0.001 {
		return fmt.Errorf("desequilibrio na transacao: total venda %f != total compra %f", totalVenda, totalCompra)
	}

	transacao := TransacaoEscrow{
		ID:            txID,
		AtivoID:       ativoID,
		Participantes: participantes,
		Status:        "pendente_assinatura",
		DataInicio:    dataInicio,
		ITBIPago:      false,
	}

	txBytes, _ := json.Marshal(transacao)
	return ctx.GetStub().PutState(txID, txBytes)
}

// AssinarTransacao registra o consentimento de uma das partes envolvidas (RN-TR01)
func (s *SmartContract) AssinarTransacao(ctx contractapi.TransactionContextInterface, txID string, entidadeID string) error {
	txJSON, err := ctx.GetStub().GetState(txID)
	if err != nil || txJSON == nil {
		return fmt.Errorf("transacao %s nao encontrada", txID)
	}

	var transacao TransacaoEscrow
	json.Unmarshal(txJSON, &transacao)

	todasAssinaram := true
	encontrado := false

	for i, p := range transacao.Participantes {
		if p.EntidadeID == entidadeID {
			transacao.Participantes[i].Assinou = true
			encontrado = true
		}
		if !transacao.Participantes[i].Assinou {
			todasAssinaram = false
		}
	}

	if !encontrado {
		return fmt.Errorf("entidade %s nao faz parte desta transacao", entidadeID)
	}

	if todasAssinaram {
		transacao.Status = "pendente_pagamento"
	}

	txBytes, _ := json.Marshal(transacao)
	return ctx.GetStub().PutState(txID, txBytes)
}

// ConfirmarOraculo processa eventos externos, como confirmação de ITBI (RN-OR01)
func (s *SmartContract) ConfirmarOraculo(ctx contractapi.TransactionContextInterface, txID string, tipoEvento string) error {
	txJSON, err := ctx.GetStub().GetState(txID)
	if err != nil || txJSON == nil {
		return fmt.Errorf("transacao %s nao encontrada", txID)
	}

	var transacao TransacaoEscrow
	json.Unmarshal(txJSON, &transacao)

	if tipoEvento == "Confirmacao_Pagamento_ITBI" {
		transacao.ITBIPago = true
	}

	txBytes, _ := json.Marshal(transacao)
	return ctx.GetStub().PutState(txID, txBytes)
}

// FinalizarTransacao conclui a transferencia atômica se todos assinaram e pagaram (RN-TR02, RN-TR03)
func (s *SmartContract) FinalizarTransacao(ctx contractapi.TransactionContextInterface, txID string) error {
	txJSON, err := ctx.GetStub().GetState(txID)
	if err != nil || txJSON == nil {
		return fmt.Errorf("transacao %s nao encontrada", txID)
	}

	var transacao TransacaoEscrow
	json.Unmarshal(txJSON, &transacao)

	if transacao.Status != "pendente_pagamento" {
		return fmt.Errorf("status da transacao nao permite finalizacao: %s", transacao.Status)
	}

	if !transacao.ITBIPago {
		return fmt.Errorf("ITBI nao foi pago (aguardando oraculo)")
	}

	// Executar a transferência atômica de posses no ativo
	ativo, _ := s.LerAtivo(ctx, transacao.AtivoID)

	for _, p := range transacao.Participantes {
		if p.Papel == "vendedor" {
			s.atualizarSaldoPosse(ativo, p.EntidadeID, -p.Percentual)
		} else if p.Papel == "comprador" {
			s.atualizarSaldoPosse(ativo, p.EntidadeID, p.Percentual)
		}
	}

	// Limpar saldos zerados
	var novasPosses []PosseFracionada
	for _, p := range ativo.Posses {
		if p.Percentual > 0 {
			novasPosses = append(novasPosses, p)
		}
	}
	ativo.Posses = novasPosses

	ativoBytes, _ := json.Marshal(ativo)
	ctx.GetStub().PutState(transacao.AtivoID, ativoBytes)

	transacao.Status = "concluida"
	txBytes, _ := json.Marshal(transacao)
	return ctx.GetStub().PutState(txID, txBytes)
}

func (s *SmartContract) getSaldoPosse(ativo *Ativo, entidadeID string) float64 {
	for _, p := range ativo.Posses {
		if p.EntidadeID == entidadeID {
			return p.Percentual
		}
	}
	return 0.0
}

func (s *SmartContract) atualizarSaldoPosse(ativo *Ativo, entidadeID string, variacao float64) {
	for i, p := range ativo.Posses {
		if p.EntidadeID == entidadeID {
			ativo.Posses[i].Percentual += variacao
			return
		}
	}
	// Se não existe e variação for positiva, cria
	if variacao > 0 {
		ativo.Posses = append(ativo.Posses, PosseFracionada{
			EntidadeID: entidadeID,
			Percentual: variacao,
		})
	}
}


func (s *SmartContract) AtivoExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("falha ao ler o estado mundial: %v", err)
	}
	return assetJSON != nil, nil
}

func (s *SmartContract) EntidadeExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	entJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("falha ao ler o estado: %v", err)
	}
	return entJSON != nil, nil
}

func (s *SmartContract) LerAtivo(ctx contractapi.TransactionContextInterface, id string) (*Ativo, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler do banco: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("o ativo %s nao existe", id)
	}

	var ativo Ativo
	err = json.Unmarshal(assetJSON, &ativo)
	if err != nil {
		return nil, err
	}

	return &ativo, nil
}

// main
func main() {
	chaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		fmt.Printf("Erro ao criar o smart contract cartorio: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Erro ao iniciar o smart contract cartorio: %s", err.Error())
	}
}
