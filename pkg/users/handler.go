package users

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	users  *UserService
	router *gin.Engine
}

func (h *UserHandler) HandleGetUser(c *gin.Context) {
	idStr, exists := c.Params.Get("user_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id parameter is required"})
		return
	}
	userId, err := parseId(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid userid parameter"})
		return
	}

	user, err := h.users.GetUser(c.Request.Context(), userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func registerHandlers(r *gin.Engine, h *UserHandler) {
	r.GET("/users/:user_id", h.HandleGetUser)
}

func newUserHandler(userService *UserService) *UserHandler {
	router := gin.Default()
	h := &UserHandler{
		users:  userService,
		router: router,
	}
	registerHandlers(router, h)
	return h
}

func (h *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func parseId(idStr string) (int32, error) {
	i32, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(i32), nil
}
