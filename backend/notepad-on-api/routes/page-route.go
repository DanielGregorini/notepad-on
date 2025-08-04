package routes

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// defina o upgrader como uma variável de pacote, com CheckOrigin liberado
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // aceita qualquer origin
	},
}

func wsHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("upgrade error:", err)
		return
	}
	defer conn.Close()

	for {
		msgType, payload, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("read error:", err)
			break
		}
		fmt.Printf("Recebi (%d bytes): %s\n", len(payload), string(payload))
		resposta := "mensagem: oi"
		if err := conn.WriteMessage(msgType, []byte(resposta)); err != nil {
			fmt.Println("write error:", err)
			break
		}
	}
}


func PageRoute(rg *gin.Engine) {

	rg.GET("/page", wsHandler)
	
}