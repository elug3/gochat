package contacts

import (
	"net/http"
	"strconv"

	"github.com/elug3/gochat/pkg/contacts/access"
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
	userId, err := parseInt32(c.Param("user_id"))
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
	userId, err := parseInt32(c.Param("user_id"))
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
	userId, err := parseInt32(c.Param("user_id"))
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
		UserId int32  `json:"user_id" binding:"required"`
		Name   string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	p, err := h.contacts.CreateProfile(c.Request.Context(), req.UserId, req.Name)
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
	userId, err := parseInt32(c.Param("user_id"))
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
	uid, err := parseInt32(c.Param("user_id"))
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
	uid, err := parseInt32(c.Param("user_id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid user_id parameter"})
		return
	}
	var req struct {
		UserId int32 `json:"user_id" binding:"required"`
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
	uid, err := parseInt32(c.Param("user_id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid user_id parameter"})
		return
	}
	mid, err := parseInt32(c.Param("member_id"))
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

func (h *ContactsHandler) HandleAccess(c *gin.Context) {
	var req struct {
		UserId   int32  `form:"user_id" binding:"required"`
		ChatId   int    `form:"chat_id"`
		TargetId int32  `form:"target_id"`
		Action   string `form:"action"`
	}
	err := c.BindQuery(&req)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid query parameters"})
		return
	}
	var errMsg string
	can, action, err := h.contacts.Can(AccessRequest{
		UserId:   req.UserId,
		ChatId:   req.ChatId,
		TargetId: req.TargetId,
		Action:   access.Action(req.Action),
	})
	if err != nil {
		errMsg = err.Error()
	}
	c.IndentedJSON(http.StatusOK, gin.H{
		"can":    can,
		"action": action,
		"error":  errMsg,
	})
}

func (h *ContactsHandler) HandleGetProfile(c *gin.Context) {
	userId, err := parseInt32(c.Param("user_id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid user_id parameter"})
		return
	}
	p, err := h.contacts.GetProfile(userId)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "profile not found"})
		return
	}
	c.IndentedJSON(http.StatusOK, p)
}

func (h *ContactsHandler) HandleListContacts(c *gin.Context) {
	userId, err := parseUserId(c)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid user_id parameter"})
		return
	}
	contacts, err := h.contacts.ListUserContacts(userId)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, contacts)
}

func registerRoutes(router gin.IRouter, h *ContactsHandler) {

	router.GET("/groups", h.HandleListGroups)
	router.GET("/groups/:group_id", h.HandleGetGroup)

	router.GET("/can", h.HandleAccess)

	router.GET("/user/:user_id/groups", h.HandleListUserGroup)
	router.POST("/user/:user_id/groups", h.HandleCreateUserGroup)
	router.GET("user/:user_id/profile", h.HandleGetProfile)

	router.GET("/user/:user_id/groups/:group_id", h.HandleGetUserGroup)
	router.GET("/user/:user_id/groups/:group_id/members", h.HandleListGroupMembers)
	router.POST("/user/:user_id/groups/:group_id/members", h.HandleInviteGroupMember)
	router.DELETE("/user/:user_id/groups/:group_id/members/:member_id", h.HandleRemoveGroupMember) // TODO: HandleRemoveGroupMember

	router.GET("/profiles", h.HandleListProfiles)
	router.POST("/profiles", h.HandleCreateProfile)
	router.DELETE("/profiles/:user_id", h.HandleDeleteProfile)

	router.GET("/contacts", h.HandleListContacts)
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

func parseInt32(s string) (int32, error) {
	i64, err := strconv.ParseInt(s, 10, 32)
	return int32(i64), err
}

// parseUserId parses user_id from path parameter
func parseUserId(c *gin.Context) (int32, error) {
	return parseInt32(c.Param("user_id"))
}
