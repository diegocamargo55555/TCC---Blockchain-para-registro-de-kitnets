package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/diegocamargo55555/TCC---Blockchain-para-registro-de-kitnets/backend/fabric"
	"github.com/diegocamargo55555/TCC---Blockchain-para-registro-de-kitnets/backend/ipfs"
	"github.com/gin-gonic/gin"
)

type KitnetController struct {
	Gateway *fabric.GatewayConnection
}

func NewKitnetController(gw *fabric.GatewayConnection) *KitnetController {
	return &KitnetController{Gateway: gw}
}

type DonoRequest struct {
	ID              string  `json:"id" binding:"required"`
	Tipo            string  `json:"tipo" binding:"required"` // "PF" ou "PJ"
	Documento       string  `json:"documento" binding:"required"`
	Nome            string  `json:"nome" binding:"required"`
	Carteira        string  `json:"carteira" binding:"required"`
	RepresentanteID string  `json:"representante_legal_id"` // Opcional, exigido se PJ
	PercentualPosse float64 `json:"percentual_posse" binding:"required"`
}

type RegistrarKitnetRequest struct {
	KitnetID        string        `json:"kitnet_id" binding:"required"`
	KitnetDescricao string        `json:"kitnet_descricao" binding:"required"`
	KitnetTokenID   string        `json:"kitnet_token_id" binding:"required"`
	Donos           []DonoRequest `json:"donos" binding:"required,min=1"`
}

func (kc *KitnetController) Criar(c *gin.Context) {
	var req RegistrarKitnetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var posses []map[string]interface{}

	// 1. Registrar as Entidades (Donos)
	for _, dono := range req.Donos {
		_, err := kc.Gateway.Contract.SubmitTransaction(
			"RegistrarEntidade",
			dono.ID,
			dono.Tipo,
			dono.Documento,
			dono.Nome,
			dono.Carteira,
			dono.RepresentanteID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Falha ao registrar dono %s: %s", dono.ID, err.Error())})
			return
		}

		posses = append(posses, map[string]interface{}{
			"entidade_id": dono.ID,
			"percentual":  dono.PercentualPosse,
		})
	}

	// 2. Criar o Ativo (Kitnet)
	possesBytes, _ := json.Marshal(posses)
	possesJSON := string(possesBytes)
	hoje := time.Now().Format("2006-01-02")

	_, err := kc.Gateway.Contract.SubmitTransaction(
		"CriarAtivo",
		req.KitnetID,
		req.KitnetDescricao,
		req.KitnetTokenID,
		hoje,
		possesJSON,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao criar kitnet: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Kitnet e proprietários registrados com sucesso na Blockchain!"})
}

func (kc *KitnetController) AverbarArquivo(c *gin.Context) {
	kitnetID := c.Param("id")

	// Lê o arquivo do multipart form
	fileHeader, err := c.FormFile("documento")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "O campo 'documento' é obrigatorio"})
		return
	}

	tipoAverbacao := c.PostForm("tipo_averbacao")
	descricao := c.PostForm("descricao")
	averbacaoID := "AVB" + time.Now().Format("150405")

	if tipoAverbacao == "" || descricao == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Os campos 'tipo_averbacao' e 'descricao' são obrigatórios"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao ler arquivo"})
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao converter bytes do arquivo"})
		return
	}

	// 1. Fazer upload para o IPFS
	cidIPFS, err := ipfs.AddFileToIPFS(fileHeader.Filename, fileBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha no upload para o IPFS: " + err.Error()})
		return
	}

	// 2. Submeter transação de averbação no Fabric
	hoje := time.Now().Format("2006-01-02")
	_, err = kc.Gateway.Contract.SubmitTransaction(
		"AverbarImovel",
		kitnetID,
		averbacaoID,
		tipoAverbacao,
		descricao,
		cidIPFS,
		hoje,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao gravar averbação na blockchain: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Averbação realizada com sucesso!",
		"cid_ipfs": cidIPFS,
		"averbacao_id": averbacaoID,
	})
}

func (kc *KitnetController) Ler(c *gin.Context) {
	kitnetID := c.Param("id")

	result, err := kc.Gateway.Contract.EvaluateTransaction("LerAtivo", kitnetID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kitnet não encontrada ou erro na blockchain: " + err.Error()})
		return
	}

	// O retorno do Fabric já é um JSON em bytes
	c.Data(http.StatusOK, "application/json", result)
}
