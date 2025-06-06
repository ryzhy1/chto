package handlers

import (
	"boton-back/internal/domain/dto"
	"boton-back/internal/services"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
)

type UserService interface {
	GetUser(ctx context.Context, userId uuid.UUID) (*dto.User, error)
}

type UserHandler struct {
	log         *slog.Logger
	userService *services.UserService
}

func NewUserHandler(log *slog.Logger, userService *services.UserService) *UserHandler {
	return &UserHandler{
		log:         log,
		userService: userService,
	}
}

func (h *UserHandler) GetUser(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, err := uuid.Parse(userIDVal.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"user": user})
}
