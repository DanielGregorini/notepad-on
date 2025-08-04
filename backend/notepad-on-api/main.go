package main

import (
	"github.com/gin-gonic/gin"

	"github.com/DanielGregorini/notepad-on/config"
	"github.com/DanielGregorini/notepad-on/db"
	"github.com/DanielGregorini/notepad-on/model"
    "github.com/DanielGregorini/notepad-on/routes"
)

var (
	cfg    = config.Load()
	dbConn = db.Connect(cfg)
)

func main() {

	// migrations db
	dbConn.AutoMigrate(&model.Page{})

	server := gin.Default()

	// aceita qualquer um
	server.SetTrustedProxies([]string{"*"})

	server.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "api funcionando!"})
	})

    //rotas
    routes.PageRoute(server)

	server.Run(":8888")
}
