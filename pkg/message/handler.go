package message

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	Messages *MessageService
	router   *gin.Engine
}

func NewMessageHandler(messages *MessageService) *MessageHandler {
	router := gin.Default()
	h := MessageHandler{
		Messages: messages,
		router:   router,
	}
	RegisterRoutes(router, &h)

	return &h
}

func (h *MessageHandler) HandleListMessages(c *gin.Context) {
	chatIdStr := c.Param("chat_id")
	chatId, err := strconv.Atoi(chatIdStr)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid chat_id"})
		return
	}

	messages, err := h.Messages.ListMessages(c.Request.Context(), chatId)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "failed to list messages"})
		return
	}

	c.IndentedJSON(http.StatusOK, messages)
}

func (h *MessageHandler) HandleUserSendMessage(c *gin.Context) {
	userIdStr := c.Param("user_id")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	var req struct {
		ChatId  int    `json:"chat_id" binding:"required"`
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	msg, err := h.Messages.Send(c.Request.Context(), userId, req.ChatId, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, msg)
}

func (h *MessageHandler) HandleUserListMessage(c *gin.Context) {
	userIdStr := c.Param("user_id")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	chatIdStr := c.Param("chat_id")
	chatId, err := strconv.Atoi(chatIdStr)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid chat_id"})
		return
	}

	msgs, err := h.Messages.ListUserChatMessages(c.Request.Context(), chatId, userId)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, msgs)
}

func RegisterRoutes(router *gin.Engine, h *MessageHandler) {
	router.GET("/chat/:chat_id/messages", h.HandleListMessages)

	router.GET("/user/:user_id/chat/:chat_id/messages", h.HandleUserListMessage)
	router.POST("/user/:user_id/messages", h.HandleUserSendMessage)
}

func (h *MessageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}
