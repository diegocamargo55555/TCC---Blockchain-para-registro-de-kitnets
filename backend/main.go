package main

import (
	"log"
	"net/http"

	"github.com/diegocamargo55555/TCC---Blockchain-para-registro-de-kitnets/backend/controllers"
	"github.com/diegocamargo55555/TCC---Blockchain-para-registro-de-kitnets/backend/fabric"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("Conectando ao Hyperledger Fabric...")
	gatewayConn, err := fabric.Connect()
	if err != nil {
		log.Fatalf("Falha ao inicializar o Fabric Gateway: %v", err)
	}
	defer gatewayConn.Gateway.Close()
	log.Println("Conectado ao Fabric com sucesso!")

	r := gin.Default()

	// Configuração de CORS para o Frontend React
	r.Use(cors.Default())

	kc := controllers.NewKitnetController(gatewayConn)

	// Rotas Básicas para validação
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	api := r.Group("/api")
	{
		api.POST("/kitnets", kc.Criar)
		api.GET("/kitnets/:id", kc.Ler)
		api.POST("/kitnets/:id/averbacoes", kc.AverbarArquivo)
	}

	// Inicializa o Servidor
	log.Println("Servidor Backend rodando na porta 8081...")
	if err := r.Run(":8081"); err != nil {
		log.Fatalf("Falha ao iniciar servidor: %v", err)
	}
}
