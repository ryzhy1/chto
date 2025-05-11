package handlers

import (
	"boton-back/internal/services"
	"context"
	"github.com/gin-gonic/gin"
	"log/slog"
)

type AuthService interface {
	Register(ctx context.Context, login, email, password string) error
	Login(ctx context.Context, input, password string) (string, string, error)
	UpdateUserEmail(ctx context.Context, userId, oldEmail, newEmail string) (string, error)
	UpdateUserPassword(ctx context.Context, userId, oldPassword, newPassword string) (string, error)
}

type AuthHandler struct {
	log         *slog.Logger
	authService *services.AuthService
}

func NewAuthHandler(log *slog.Logger, authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		log:         log,
		authService: authService,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	err := h.authService.Register(c.Request.Context(), input.Username, input.Email, input.Password)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message:": "user created"})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input struct {
		Input    string `json:"input"`
		Password string `json:"password"`
	}
	if err := c.BindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	accessToken, refreshToken, err := h.authService.Login(c.Request.Context(), input.Input, input.Password)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"accessToken": accessToken, "refresh_token": refreshToken})
}

func (h *AuthHandler) UpdateUserEmail(c *gin.Context) {
	var input struct {
		UserID   string `json:"user_id"`
		OldEmail string `json:"old_email"`
		NewEmail string `json:"new_email"`
	}
	if err := c.BindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	message, err := h.authService.UpdateUserEmail(c.Request.Context(), input.UserID, input.OldEmail, input.NewEmail)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": message})
}

func (h *AuthHandler) UpdateUserPassword(c *gin.Context) {
	var input struct {
		UserID      string `json:"user_id"`
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := c.BindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	message, err := h.authService.UpdateUserPassword(c.Request.Context(), input.UserID, input.OldPassword, input.NewPassword)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": message})
}
