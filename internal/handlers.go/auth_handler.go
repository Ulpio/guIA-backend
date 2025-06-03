package handlers

import (
	"net/http"

	"github.com/Ulpio/guIA-backend/internal/services"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService services.AuthServiceInterface
}

func NewAuthHandler(authService services.AuthServiceInterface) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body services.RegisterRequest true "User registration data"
// @Success 201 {object} services.AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req services.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Dados inválidos",
			Message: err.Error(),
		})
		return
	}

	response, err := h.authService.Register(&req)
	if err != nil {
		statusCode := http.StatusInternalServerError

		// Determinar código de status baseado no erro
		errorMsg := err.Error()
		switch {
		case contains(errorMsg, "já está em uso"), contains(errorMsg, "já existe"):
			statusCode = http.StatusConflict
		case contains(errorMsg, "inválido"), contains(errorMsg, "obrigatório"):
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   "Erro no registro",
			Message: errorMsg,
		})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Message: "Usuário registrado com sucesso",
		Data:    response,
	})
}

// Login godoc
// @Summary User login
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body services.LoginRequest true "User login credentials"
// @Success 200 {object} services.AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req services.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Dados inválidos",
			Message: err.Error(),
		})
		return
	}

	response, err := h.authService.Login(&req)
	if err != nil {
		statusCode := http.StatusInternalServerError

		// Determinar código de status baseado no erro
		errorMsg := err.Error()
		switch {
		case contains(errorMsg, "credenciais inválidas"), contains(errorMsg, "conta desativada"):
			statusCode = http.StatusUnauthorized
		case contains(errorMsg, "obrigatório"):
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   "Erro no login",
			Message: errorMsg,
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Login realizado com sucesso",
		Data:    response,
	})
}

// RefreshToken godoc
// @Summary Refresh JWT token
// @Description Refresh an expired JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token"
// @Success 200 {object} services.AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Dados inválidos",
			Message: err.Error(),
		})
		return
	}

	response, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		statusCode := http.StatusInternalServerError

		errorMsg := err.Error()
		switch {
		case contains(errorMsg, "inválido"), contains(errorMsg, "expirado"):
			statusCode = http.StatusUnauthorized
		case contains(errorMsg, "não encontrado"), contains(errorMsg, "desativada"):
			statusCode = http.StatusUnauthorized
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   "Erro ao renovar token",
			Message: errorMsg,
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Token renovado com sucesso",
		Data:    response,
	})
}

// Logout godoc
// @Summary User logout
// @Description Logout user (client-side token removal)
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Como estamos usando JWT stateless, o logout é feito no frontend
	// removendo o token do storage local
	// Aqui podemos adicionar lógica adicional como blacklist de tokens se necessário

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Logout realizado com sucesso",
		Data:    nil,
	})
}

// ValidateToken godoc
// @Summary Validate JWT token
// @Description Validate if the provided JWT token is valid
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} TokenValidationResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/validate [get]
func (h *AuthHandler) ValidateToken(c *gin.Context) {
	// O token já foi validado pelo middleware, então só retornamos as informações
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Token inválido",
			Message: "Não foi possível extrair informações do token",
		})
		return
	}

	username, _ := c.Get("username")
	userType, _ := c.Get("user_type")

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Token válido",
		Data: TokenValidationResponse{
			Valid:    true,
			UserID:   userID.(uint),
			Username: username.(string),
			UserType: userType.(string),
		},
	})
}

// Structs auxiliares
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type TokenValidationResponse struct {
	Valid    bool   `json:"valid"`
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	UserType string `json:"user_type"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Função auxiliar para verificar se uma string contém uma substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 1; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}
