package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	server := gin.Default()

	//rota de teste
	server.GET("/", func(context *gin.Context) {
		context.JSON(200, gin.H{"message": "api funcionando!"})
	})

	//inicia o servidor
	server.Run(":8888")
}