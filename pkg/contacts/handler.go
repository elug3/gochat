package contacts

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/elug3/gochat/pkg/contacts/access"
	"github.com/elug3/gochat/pkg/contacts/internal/errs"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type ContactsHandler struct {
	contacts *ContactsService
	router   *gin.Engine
}

func (h *ContactsHandler) HandleListGroups(c *gin.Context) {
	limit := parseLimit(c, DefaultListLimit)

	groups, err := h.contacts.ListGroups(c.Request.Context(), limit)
	if err != nil {
		respondError(c, err)
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
	g, err := h.contacts.GetGroup(c.Request.Context(), groupId)
	if err != nil {
		respondError(c, err)
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
	g, err := h.contacts.GetUserGroup(c.Request.Context(), groupId, userId)
	if err != nil {
		respondError(c, err)
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
	gs, err := h.contacts.ListUserGroups(c.Request.Context(), userId)
	if err != nil {
		respondError(c, err)
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
		Name string `json:"name" binding:"required,min=2,max=50"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	g, err := h.contacts.CreateUserGroup(c.Request.Context(), userId, req.Name)
	if err != nil {
		respondError(c, err)
		return
	}
	c.IndentedJSON(http.StatusCreated, g)
}

func (h *ContactsHandler) HandleCreateProfile(c *gin.Context) {
	uid, err := parseUserId(c)
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

	p, err := h.contacts.CreateProfile(c.Request.Context(), uid, req.Name)
	if err != nil {
		respondError(c, err)
		return
	}
	c.IndentedJSON(http.StatusCreated, p)
}

func (h *ContactsHandler) HandleListProfiles(c *gin.Context) {
	limit := parseLimit(c, DefaultListLimit)

	profiles, err := h.contacts.ListProfiles(c.Request.Context(), limit)
	if err != nil {
		respondError(c, err)
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
	if err := h.contacts.DeleteProfile(c.Request.Context(), userId); err != nil {
		respondError(c, err)
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

	ms, err := h.contacts.ListGroupMembers(c.Request.Context(), gid, uid)
	if err != nil {
		respondError(c, err)
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
	m, err := h.contacts.Invite(c.Request.Context(), gid, uid, req.UserId)
	if err != nil {
		respondError(c, err)
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
	if err := h.contacts.DeleteMember(c.Request.Context(), gid, uid, mid); err != nil {
		respondError(c, err)
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
	err := c.ShouldBindQuery(&req)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid query parameters"})
		return
	}
	var errMsg string
	can, action, err := h.contacts.Can(c.Request.Context(), AccessRequest{
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
	p, err := h.contacts.GetProfile(c.Request.Context(), userId)
	if err != nil {
		respondError(c, err)
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
	contacts, err := h.contacts.ListUserContacts(c.Request.Context(), userId)
	if err != nil {
		respondError(c, err)
		return
	}
	c.IndentedJSON(http.StatusOK, contacts)
}

func (h *ContactsHandler) HandleCreateContact(c *gin.Context) {
	userId, err := parseUserId(c)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid user_id parameter"})
		return
	}
	var req struct {
		TargetId int32 `json:"target_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	contact, err := h.contacts.AddToContacts(c.Request.Context(), userId, req.TargetId)
	if err != nil {
		respondError(c, err)
		return
	}
	c.IndentedJSON(http.StatusCreated, contact)
}

func (h *ContactsHandler) HandleDeleteGroup(c *gin.Context) {
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
	if err := h.contacts.DeleteUserGroup(c.Request.Context(), gid, uid); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ContactsHandler) HandleHealth(c *gin.Context) {
	health := h.contacts.HealthCheck(c.Request.Context())

	// Determine overall status
	status := http.StatusOK
	if dbStatus, ok := health["database"].(map[string]interface{}); ok {
		if dbStatus["status"] != "healthy" {
			status = http.StatusServiceUnavailable
		}
	}

	c.JSON(status, health)
}

func registerRoutes(router gin.IRouter, h *ContactsHandler) {

	router.GET("/groups", h.HandleListGroups)
	router.GET("/groups/:group_id", h.HandleGetGroup)
	router.GET("/health", h.HandleHealth)

	router.GET("/can", h.HandleAccess)

	router.GET("/user/:user_id/groups", h.HandleListUserGroup)
	router.POST("/user/:user_id/groups", h.HandleCreateUserGroup)
	router.GET("/user/:user_id/profile", h.HandleGetProfile)
	router.POST("/user/:user_id/profile", h.HandleCreateProfile)

	router.GET("/user/:user_id/groups/:group_id", h.HandleGetUserGroup)
	router.GET("/user/:user_id/groups/:group_id/members", h.HandleListGroupMembers)
	router.POST("/user/:user_id/groups/:group_id/members", h.HandleInviteGroupMember)
	router.DELETE("/user/:user_id/groups/:group_id", h.HandleDeleteGroup)

	router.DELETE("/user/:user_id/groups/:group_id/members/:member_id", h.HandleRemoveGroupMember) // TODO: HandleRemoveGroupMember

	router.GET("/user/:user_id/contacts", h.HandleListContacts)
	router.POST("/user/:user_id/contacts", h.HandleCreateContact)

	router.GET("/profiles", h.HandleListProfiles)
	router.DELETE("/profiles/:user_id", h.HandleDeleteProfile)
}

func newContactsHandler(contacts *ContactsService) *ContactsHandler {
	router := gin.Default()

	// Add middleware
	router.Use(requestIDMiddleware())
	router.Use(loggingMiddleware())

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

func respondError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, errs.ErrPermissionDenied):
		status = http.StatusForbidden
	case errors.Is(err, errs.ErrNotFound),
		errors.Is(err, errs.ErrGroupNotExists),
		errors.Is(err, errs.ErrUserNotExists):
		status = http.StatusNotFound
	case errors.Is(err, errs.ErrExists):
		status = http.StatusConflict
	case errors.Is(err, errs.ErrSelfContact):
		status = http.StatusBadRequest
	}

	c.IndentedJSON(status, gin.H{"error": err.Error()})
}

func parseInt32(s string) (int32, error) {
	i64, err := strconv.ParseInt(s, 10, 32)
	return int32(i64), err
}

// parseUserId parses user_id from path parameter
func parseUserId(c *gin.Context) (int32, error) {
	return parseInt32(c.Param("user_id"))
}

// parseLimit parses limit query parameter with validation
func parseLimit(c *gin.Context, defaultLimit int) int {
	limitStr := c.DefaultQuery("limit", strconv.Itoa(defaultLimit))
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		return defaultLimit
	}
	if limit > MaxListLimit {
		return MaxListLimit
	}
	return limit
}

// requestIDMiddleware adds a request ID to each request
func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		requestID, _ := c.Get("request_id")
		reqID, _ := requestID.(string)

		log.Info().
			Str("request_id", reqID).
			Str("method", method).
			Str("path", path).
			Int("status", status).
			Dur("latency", latency).
			Msg("HTTP request")
	}
}
