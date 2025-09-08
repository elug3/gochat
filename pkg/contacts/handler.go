package contacts

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ContactsHandler struct {
	contacts *ContactsService
	router   *gin.Engine
}

func (h *ContactsHandler) HandleListGroups(c *gin.Context) {
	groups, err := h.contacts.ListGroups()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return

	}
	c.IndentedJSON(http.StatusOK, groups)
}

func (h *ContactsHandler) HandleGetGroup(c *gin.Context) {
	groupId, err := strconv.Atoi(c.Param("group_id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid group_id parameter"})
		return
	}
	g, err := h.contacts.GetGroup(groupId)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "group not found"})
		return
	}
	c.IndentedJSON(http.StatusOK, g)
}

func (h *ContactsHandler) HandleGetUserGroup(c *gin.Context) {
	userId, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid user_id parameter"})
		return
	}
	groupId, err := strconv.Atoi(c.Param("group_id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid group_id parameter"})
		return
	}
	g, err := h.contacts.GetUserGroup(groupId, userId)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, g)
}

func (h *ContactsHandler) HandleListUserGroup(c *gin.Context) {
	userId, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid user_id parameter"})
		return
	}
	gs, err := h.contacts.ListUserGroups(userId)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gs)
}

func (h *ContactsHandler) HandleCreateUserGroup(c *gin.Context) {
	userId, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid user_id parameter"})
		return
	}
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	g, err := h.contacts.CreateUserGroup(userId, req.Name)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, g)
}

func (h *ContactsHandler) HandleCreateProfile(c *gin.Context) {
	var req struct {
		UserId int    `json:"user_id" binding:"required"`
		Name   string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	p, err := h.contacts.CreateProfile(req.UserId, req.Name)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, p)
}

func (h *ContactsHandler) HandleListProfiles(c *gin.Context) {
	profiles, err := h.contacts.ListProfiles(50)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, profiles)
}
func (h *ContactsHandler) HandleDeleteProfile(c *gin.Context) {
	userId, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid user_id parameter"})
		return
	}
	if err := h.contacts.DeleteProfile(userId); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ContactsHandler) HandleListGroupMembers(c *gin.Context) {
	gid, err := strconv.Atoi(c.Param("group_id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid group_id parameter"})
		return
	}
	uid, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid user_id parameter"})
		return
	}

	ms, err := h.contacts.ListGroupMembers(gid, uid)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, ms)
}

func (h *ContactsHandler) HandleInviteGroupMember(c *gin.Context) {
	gid, err := strconv.Atoi(c.Param("group_id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid group_id parameter"})
		return
	}
	uid, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid user_id parameter"})
		return
	}
	var req struct {
		UserId int `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	m, err := h.contacts.Invite(gid, uid, req.UserId)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, m)
}

func (h *ContactsHandler) HandleRemoveGroupMember(c *gin.Context) {
	gid, err := strconv.Atoi(c.Param("group_id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid group_id parameter"})
		return
	}
	uid, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid user_id parameter"})
		return
	}
	mid, err := strconv.Atoi(c.Param("member_id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid member_id parameter"})
		return
	}
	if err := h.contacts.DeleteMember(gid, uid, mid); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func registerRoutes(router gin.IRouter, h *ContactsHandler) {

	router.GET("/groups", h.HandleListGroups)
	router.GET("/groups/:group_id", h.HandleGetGroup)

	router.GET("/user/:user_id/groups", h.HandleListUserGroup)
	router.POST("/user/:user_id/groups", h.HandleCreateUserGroup)

	router.GET("/user/:user_id/groups/:group_id", h.HandleGetUserGroup)
	router.GET("/user/:user_id/groups/:group_id/members", h.HandleListGroupMembers)
	router.POST("/user/:user_id/groups/:group_id/members", h.HandleInviteGroupMember)
	router.DELETE("/user/:user_id/groups/:group_id/members/:member_id", h.HandleRemoveGroupMember) // TODO: HandleRemoveGroupMember

	router.GET("/profiles", h.HandleListProfiles)
	router.POST("/profiles", h.HandleCreateProfile)
	router.DELETE("/profiles/:user_id", h.HandleDeleteProfile)
}

func newContactsHandler(contacts *ContactsService) *ContactsHandler {
	router := gin.Default()
	h := &ContactsHandler{
		router:   router,
		contacts: contacts,
	}
	registerRoutes(router, h)
	return h
}

func (h *ContactsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}
