package presence

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	router    *gin.Engine
	Presences *PresenceService
}

func NewHttpHandler(opts *HttpOptions) (*Handler, error) {
	presences, err := NewPresenceService(opts)
	if err != nil {
		return nil, err
	}
	router := gin.Default()
	h := Handler{
		router:    router,
		Presences: presences,
	}

	h.registerRoutes()

	return &h, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func (h *Handler) registerRoutes() {
	h.router.GET("/users/:userId/presence", func(c *gin.Context) {
		userId, err := parseUserId(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		presence, err := h.Presences.GetPresence(c.Request.Context(), userId)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusOK, presence)
	})
}

func parseUserId(c *gin.Context) (int32, error) {
	uidStr, exists := c.Params.Get("userId")
	if !exists {
		return 0, fmt.Errorf("userId is required")
	}
	uid, err := strconv.ParseInt(uidStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid userId: %v", err)
	}
	return int32(uid), nil
}
