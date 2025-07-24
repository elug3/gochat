package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/elug3/gochat/pkg/service"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	users *service.UserService
}

// parseToken extracts the bearer token from the Authorization header
func parseToken(r *http.Request) (string, error) {
	token := r.Header.Get("Authorization")
	if token == "" {
		return "", fmt.Errorf("missing Authorization header")
	}
	s := strings.SplitN(token, " ", 2)
	if strings.ToLower(s[0]) != "bearer" || len(s) != 2 {
		return "", fmt.Errorf("invalid Authorization header format")
	}
	return s[1], nil
}

// AuthMiddleware is a middleware that checks for a valid authentication token
// if the token is valid, it sets the userId in the context
func AuthMiddleware(s *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if tokenString, err := parseToken(c.Request); err == nil {
			if token, err := s.Authenticate(c.Request.Context(), tokenString); err == nil {
				c.Set("userId", token.UserId)
			}
		}
		c.Next()
	}
}

func NewAuthHandler(userService *service.UserService) (*AuthHandler, error) {
	return &AuthHandler{users: userService}, nil
}

func (h *AuthHandler) HandleGetAuthorizations(c *gin.Context) {

}

func (h *AuthHandler) HandleLogin(c *gin.Context) {
	username, password, ok := c.Request.BasicAuth()
	if !ok || username == "" || password == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    http.StatusUnauthorized,
			"message": "Unauthorized",
		})
		return
	}
	token, err := h.users.Login(c.Request.Context(), username, password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    http.StatusUnauthorized,
			"message": "Unauthorized",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}
