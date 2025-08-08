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
	rooms     = make(map[string]map[*websocket.Conn]struct{})
	roomsLock sync.RWMutex

	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024 * 4,
		WriteBufferSize: 1024 * 4,
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

// garante sala
func addConn(slug string, c *websocket.Conn) {
	roomsLock.Lock()
	defer roomsLock.Unlock()
	if rooms[slug] == nil {
		rooms[slug] = make(map[*websocket.Conn]struct{})
	}
	rooms[slug][c] = struct{}{}
}

func removeConn(slug string, c *websocket.Conn) {
	roomsLock.Lock()
	defer roomsLock.Unlock()
	if set, ok := rooms[slug]; ok {
		delete(set, c)
		if len(set) == 0 {
			delete(rooms, slug)
		}
	}
}

func broadcast(slug string, msgType int, payload []byte) {
	roomsLock.RLock()
	defer roomsLock.RUnlock()
	for c := range rooms[slug] {
		_ = c.SetWriteDeadline(time.Now().Add(5 * time.Second))
		if err := c.WriteMessage(msgType, payload); err != nil {
			// limpa conexão morta
			go func(conn *websocket.Conn) {
				_ = conn.Close()
				removeConn(slug, conn)
			}(c)
		}
	}
}

func (ctrl *pageController) FetchAndUpdateText(ctx *gin.Context) {
	slug := ctx.Param("slug")
	if slug == "" {
		ctx.Status(http.StatusBadRequest)
		return
	}

	// find-or-create da página (igual notepad: cria vazia se não existe)
	var page model.Page
	if err := db.Where("slug = ?", slug).First(&page).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			page = model.Page{Slug: slug, Content: ""}
			if err := db.Create(&page).Error; err != nil {
				ctx.Status(http.StatusInternalServerError)
				return
			}
		} else {
			ctx.Status(http.StatusInternalServerError)
			return
		}
	}

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// keep-alive básico
	conn.SetReadLimit(1 << 20) // 1MB
	_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	addConn(slug, conn)
	defer removeConn(slug, conn)

	// envia conteúdo inicial
	_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	if err := conn.WriteMessage(websocket.TextMessage, []byte(page.Content)); err != nil {
		return
	}

	// loop principal
	for {
		msgType, payload, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if msgType != websocket.TextMessage && msgType != websocket.BinaryMessage {
			continue
		}

		// persiste (content é texto plano; se quiser JSON, adapte)
		if err := db.Model(&model.Page{}).
			Where("slug = ?", slug).
			Updates(map[string]any{
				"content":    string(payload),
				"updated_at": time.Now(),
			}).Error; err != nil {
			// não derruba conexão por erro de DB; apenas ignora este update
			continue
		}

		// broadcast p/ todos da mesma página (inclui remetente)
		broadcast(slug, msgType, payload)
	}
}
