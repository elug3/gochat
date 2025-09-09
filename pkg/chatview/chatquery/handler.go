package chatquery

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ChatViewHandler struct {
	chatview *ChatViewService
	router   *gin.Engine
}

func registerRoutes(router *gin.Engine, h *ChatViewHandler) {
	router.GET("/chats/:user_id", h.HandleListChats)
}

func NewHandler(chatview *ChatViewService) *ChatViewHandler {
	router := gin.Default()
	h := &ChatViewHandler{chatview: chatview, router: router}

	registerRoutes(router, h)
	return h
}

func (h *ChatViewHandler) HandleListChats(c *gin.Context) {
	userIdStr, exists := c.Params.Get("user_id")
	if !exists {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}
	chats, err := h.chatview.ListChats(c.Request.Context(), int32(userId))
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, chats)
}

func (h *ChatViewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}
