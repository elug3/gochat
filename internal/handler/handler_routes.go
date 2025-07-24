package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(userHandler *UserHandler, authHandler *AuthHandler, contactsHandler *GroupHandler) *gin.Engine {
	r := gin.Default()
	v1 := r.Group("/api/v1")
	v1.Use(AuthMiddleware(userHandler.userService))
	{
		addRoutes(v1, "/users", usersRoutes(userHandler))
		addRoutes(v1, "/auth", authRoutes(authHandler))
		addRoutes(v1, "/groups", groupRoutes(contactsHandler), authRequired)
	}

	return r
}

func authRequired(c *gin.Context) {
	if _, exists := c.Get("userId"); !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"code":    http.StatusUnauthorized,
			"message": "Unauthorized",
		})
		return
	}
	c.Next()
}

func groupRoutes(h *GroupHandler) func(gin.IRouter) {
	return func(r gin.IRouter) {
		r.POST("", h.HandleCreateGroup)
		r.GET("", h.HandleGetGroups)
		r.GET(":id", h.HandleGetGroup)
	}
}

func authRoutes(h *AuthHandler) func(gin.IRouter) {
	return func(r gin.IRouter) {
		r.POST("/login", h.HandleLogin)
	}
}
func usersRoutes(h *UserHandler) func(gin.IRouter) {
	return func(r gin.IRouter) {
		r.POST("", h.HandleCreateUser)
	}
}

func addRoutes(r gin.IRouter, path string, register func(gin.IRouter), middlewares ...gin.HandlerFunc) {
	subRouter := r.Group(path, middlewares...)
	register(subRouter)
}
