package handler

import (
	"net/http"

	"github.com/elug3/gochat/pkg/service"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) (*UserHandler, error) {
	return &UserHandler{userService: userService}, nil
}

func (h *UserHandler) HandleCreateUser(c *gin.Context) {
	var params struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&params); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "invalid request",
		})
		return
	}

	user, err := h.userService.Register(c.Request.Context(), params.Username, params.Password)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}
	c.IndentedJSON(http.StatusOK, user)
}

// /user
func (h *UserHandler) HandleGetUser(c *gin.Context) {
	userId := c.GetInt("userId")
	user, err := h.userService.GetUser(c.Request.Context(), userId)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}
	c.IndentedJSON(http.StatusOK, user)
}
