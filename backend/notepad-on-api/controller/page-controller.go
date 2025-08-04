package controller

import (
	"net/http"
	"sync"
	"time"

	"github.com/DanielGregorini/notepad-on/model"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

var (
	rooms     = make(map[string]map[*websocket.Conn]bool)

	roomsLock sync.RWMutex

	upgrader  = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	db *gorm.DB 
)

type PageController interface {
	FetchAndUpdateText(ctx *gin.Context)
}

type pageController struct{}

func NewPageController(database *gorm.DB) PageController {
	db = database
	return &pageController{}
}

func (ctrl *pageController) FetchAndUpdateText(ctx *gin.Context) {
	slug := ctx.Param("slug")

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	roomsLock.Lock()

	if rooms[slug] == nil {
		rooms[slug] = make(map[*websocket.Conn]bool)
	}

	rooms[slug][conn] = true
	roomsLock.Unlock()

	defer func() {
		roomsLock.Lock()
		delete(rooms[slug], conn)
		roomsLock.Unlock()
		conn.Close()
	}()

	for {

		msgType, payload, err := conn.ReadMessage()
		if err != nil {
			break
		}

		db.Model(&model.Page{}).
			Where("slug = ?", slug).
			Updates(map[string]interface{}{
				"content":    string(payload),
				"updated_at": time.Now(),
			})

		roomsLock.RLock()

		for c := range rooms[slug] {
			if c != conn {
				c.WriteMessage(msgType, payload)
			}
		}

		roomsLock.RUnlock()
	}
}
