package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type AuthHandler struct {
	router *gin.Engine
	auth   *AuthService
}

func registerRoutes(r gin.IRouter, h *AuthHandler) {

	r.GET("/login", h.HandleLogin)
	r.POST("/register", h.HandleRegister)
	r.GET("/.well-known/jwks.json", h.HandleJwks)
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		log.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Int("size", c.Writer.Size()).
			Send()
	}
}

func newAuthHandler(authService *AuthService) *AuthHandler {
	router := gin.New()
	router.Use(Logger())
	h := AuthHandler{
		router: router,
		auth:   authService,
	}

	registerRoutes(router, &h)

	return &h
}

func (h *AuthHandler) HandleLogin(c *gin.Context) {
	username, paasword, ok := c.Request.BasicAuth()
	if !ok {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid authorization header"})
		return
	}
	token, err := h.auth.Login(c.Request.Context(), username, paasword)
	if err != nil {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, token)
}

func (h *AuthHandler) HandleRegister(c *gin.Context) {
	var req struct {
		Username string `json:"username" bind:"required"`
		Password string `json:"password" bind:"required"`
	}
	err := c.BindJSON(&req)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	userId, err := h.auth.RegisterUser(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"user_id": userId})
}

func (h *AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func (h *AuthHandler) HandleJwks(c *gin.Context) {
	jwksData, err := h.auth.GetJwksData()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Header("Content-Type", "application/json")
	c.Status(http.StatusOK)
	c.Writer.Write(jwksData)
}
