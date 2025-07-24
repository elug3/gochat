package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/elug3/gochat/pkg/service"
	"github.com/gin-gonic/gin"
)

type GroupHandler struct {
	Contacts *service.ContactsService
}

func NewGroupHandler(contacts *service.ContactsService) (*GroupHandler, error) {
	return &GroupHandler{Contacts: contacts}, nil
}

// parseGroupId extracts and validates the group ID from the request context.
func parseGroupId(c *gin.Context) (int, error) {
	groupId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return 0, err
	}
	if groupId <= 0 {
		return 0, fmt.Errorf("invalid group ID: %d", groupId)
	}
	return groupId, nil
}

// HandleGetGroups retrieves all groups for the authenticated user.
func (h *GroupHandler) HandleGetGroups(c *gin.Context) {
	userId := c.GetInt("userId")

	groups, err := h.Contacts.GetGroups(userId)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}
	c.IndentedJSON(http.StatusOK, groups)
}

// HandleGetGroup retrieves a specific group by ID for the authenticated user.
func (h *GroupHandler) HandleGetGroup(c *gin.Context) {
	userId := c.GetInt("userId")
	groupId, err := parseGroupId(c)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": fmt.Sprintf("invalid group ID: '%d'", groupId),
		})
		return
	}

	group, err := h.Contacts.GetGroup(groupId, userId)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, group)
}

// HandleCreateGroup creates a new group for the authenticated user.
func (h *GroupHandler) HandleCreateGroup(c *gin.Context) {
	userId := c.GetInt("userId")

	var params struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&params); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "invalid request",
		})
		return
	}

	group, err := h.Contacts.CreateGroup(userId, params.Name)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, group)
}
