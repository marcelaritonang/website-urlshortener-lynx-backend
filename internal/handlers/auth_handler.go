package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/interfaces"
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/models"
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/services"
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/types"
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/utils"
	"gorm.io/gorm"
)

type AuthHandler struct {
	authService  interfaces.AuthService
	jwtSecret    string
	db           *gorm.DB
	emailService *services.EmailService
}

func NewAuthHandler(authService interfaces.AuthService, jwtSecret string, db *gorm.DB) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		jwtSecret:    jwtSecret,
		db:           db,
		emailService: services.NewEmailService(),
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

// ForgotPassword handles password reset request
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	ctx := c.Request.Context()
	token, err := h.authService.RequestPasswordReset(ctx, req.Email)
	if err != nil {
		// Log the actual error for debugging
		fmt.Printf("Error generating reset token: %v\n", err)
		utils.SuccessResponse(c, http.StatusOK, "If the email exists, a password reset link has been sent", nil)
		return
	}

	// If token is empty, email doesn't exist (security: don't reveal)
	if token == "" {
		utils.SuccessResponse(c, http.StatusOK, "If the email exists, a password reset link has been sent", nil)
		return
	}

	var user models.User
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		fmt.Printf("Error finding user: %v\n", err)
		utils.SuccessResponse(c, http.StatusOK, "If the email exists, a password reset link has been sent", nil)
		return
	}

	fullName := user.FirstName + " " + user.LastName
	if err := h.emailService.SendResetPasswordEmail(user.Email, fullName, token); err != nil {
		// ✅ Log the actual email error for debugging
		fmt.Printf("SMTP Error: %v\n", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to send email: %v", err))
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Password reset email has been sent successfully", nil)
}

// ResetPasswordConfirm handles the actual password reset with token
func (h *AuthHandler) ResetPasswordConfirm(c *gin.Context) {
	var req models.ResetPasswordConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	ctx := c.Request.Context()
	if err := h.authService.ResetPassword(ctx, req.Token, req.NewPassword); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid or expired reset token"))
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Password has been reset successfully", nil)
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
