package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/naufalrafianto/lynx-api/internal/interfaces"
	"github.com/naufalrafianto/lynx-api/internal/models"
	"github.com/naufalrafianto/lynx-api/internal/types"
	"github.com/naufalrafianto/lynx-api/internal/utils"
)

type AuthHandler struct {
	authService interfaces.AuthService
	jwtSecret   string
}

func NewAuthHandler(authService interfaces.AuthService, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		jwtSecret:   jwtSecret,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	ctx := c.Request.Context()
	user := &models.User{
		ID:        uuid.New(),
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	if err := h.authService.Register(ctx, user); err != nil {
		if err == types.ErrUserExists {
			utils.ErrorResponse(c, http.StatusConflict, err)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "User registered successfully", types.RegisterResponse{
		User: user,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	ctx := c.Request.Context()
	user, err := h.authService.Login(ctx, req.Email, req.Password)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, types.ErrInvalidCredentials)
		return
	}

	token, refresh, err := h.generateTokenPair(user.ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, types.ErrInvalidToken)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Login successful", types.LoginResponse{
		Token:        token,
		RefreshToken: refresh,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, types.ErrInvalidUUID)
		return
	}

	ctx := c.Request.Context()
	if err := h.authService.InvalidateUserSessions(ctx, userID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Logged out successfully", nil)
}

func (h *AuthHandler) GetUserDetails(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, types.ErrInvalidUUID)
		return
	}

	ctx := c.Request.Context()
	user, err := h.authService.GetUserByID(ctx, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, types.ErrUserNotFound)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User details retrieved successfully", user)
}
func (h *AuthHandler) generateTokenPair(userID uuid.UUID) (token, refresh string, err error) {
	token, err = h.generateToken(userID, 24*time.Hour)
	if err != nil {
		return "", "", err
	}

	refresh, err = h.generateToken(userID, 7*24*time.Hour)
	if err != nil {
		return "", "", err
	}

	return token, refresh, nil
}

func (h *AuthHandler) generateToken(userID uuid.UUID, expiration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(expiration).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
}
