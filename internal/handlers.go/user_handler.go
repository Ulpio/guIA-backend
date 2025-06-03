package handlers

import (
	"net/http"
	"strconv"

	"github.com/Ulpio/guIA-backend/internal/services"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService services.UserServiceInterface
}

func NewUserHandler(userService services.UserServiceInterface) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetProfile godoc
// @Summary Get current user profile
// @Description Get the profile of the authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.UserResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Não autorizado",
			Message: "Token inválido",
		})
		return
	}

	profile, err := h.userService.GetProfile(userID.(uint))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if contains(err.Error(), "não encontrado") {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   "Erro ao buscar perfil",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Perfil encontrado",
		Data:    profile,
	})
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update the profile of the authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.UpdateProfileRequest true "Profile update data"
// @Success 200 {object} models.UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Não autorizado",
			Message: "Token inválido",
		})
		return
	}

	var req services.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Dados inválidos",
			Message: err.Error(),
		})
		return
	}

	updatedProfile, err := h.userService.UpdateProfile(userID.(uint), &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMsg := err.Error()

		switch {
		case contains(errorMsg, "não encontrado"):
			statusCode = http.StatusNotFound
		case contains(errorMsg, "inválido"), contains(errorMsg, "deve ter"):
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   "Erro ao atualizar perfil",
			Message: errorMsg,
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Perfil atualizado com sucesso",
		Data:    updatedProfile,
	})
}

// GetUserByID godoc
// @Summary Get user by ID
// @Description Get a user's public profile by their ID
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} models.UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id} [get]
func (h *UserHandler) GetUserByID(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "ID inválido",
			Message: "O ID do usuário deve ser um número válido",
		})
		return
	}

	user, err := h.userService.GetUserByID(uint(userID))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if contains(err.Error(), "não encontrado") {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   "Erro ao buscar usuário",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Usuário encontrado",
		Data:    user,
	})
}

// SearchUsers godoc
// @Summary Search users
// @Description Search for users by username, name or company name
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param q query string true "Search query"
// @Param limit query int false "Number of results per page" default(20)
// @Param offset query int false "Number of results to skip" default(0)
// @Success 200 {array} models.UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/search [get]
func (h *UserHandler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Parâmetro obrigatório",
			Message: "O parâmetro 'q' (query) é obrigatório",
		})
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit <= 0 {
		limit = 20
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	users, err := h.userService.SearchUsers(query, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Erro na busca",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Busca realizada com sucesso",
		Data:    users,
	})
}

// FollowUser godoc
// @Summary Follow a user
// @Description Follow another user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID to follow"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id}/follow [post]
func (h *UserHandler) FollowUser(c *gin.Context) {
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Não autorizado",
			Message: "Token inválido",
		})
		return
	}

	idParam := c.Param("id")
	followedID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "ID inválido",
			Message: "O ID do usuário deve ser um número válido",
		})
		return
	}

	err = h.userService.FollowUser(currentUserID.(uint), uint(followedID))
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMsg := err.Error()

		switch {
		case contains(errorMsg, "não encontrado"):
			statusCode = http.StatusNotFound
		case contains(errorMsg, "não pode seguir a si mesmo"), contains(errorMsg, "já está seguindo"):
			statusCode = http.StatusConflict
		case contains(errorMsg, "inválido"):
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   "Erro ao seguir usuário",
			Message: errorMsg,
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Usuário seguido com sucesso",
		Data:    nil,
	})
}

// UnfollowUser godoc
// @Summary Unfollow a user
// @Description Stop following a user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID to unfollow"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id}/unfollow [delete]
func (h *UserHandler) UnfollowUser(c *gin.Context) {
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Não autorizado",
			Message: "Token inválido",
		})
		return
	}

	idParam := c.Param("id")
	followedID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "ID inválido",
			Message: "O ID do usuário deve ser um número válido",
		})
		return
	}

	err = h.userService.UnfollowUser(currentUserID.(uint), uint(followedID))
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMsg := err.Error()

		switch {
		case contains(errorMsg, "não está seguindo"):
			statusCode = http.StatusConflict
		case contains(errorMsg, "inválido"):
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   "Erro ao deixar de seguir usuário",
			Message: errorMsg,
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Deixou de seguir o usuário com sucesso",
		Data:    nil,
	})
}

// GetFollowers godoc
// @Summary Get user followers
// @Description Get the list of users following a specific user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param limit query int false "Number of results per page" default(20)
// @Param offset query int false "Number of results to skip" default(0)
// @Success 200 {array} models.UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id}/followers [get]
func (h *UserHandler) GetFollowers(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "ID inválido",
			Message: "O ID do usuário deve ser um número válido",
		})
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit <= 0 {
		limit = 20
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	followers, err := h.userService.GetFollowers(uint(userID), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Erro ao buscar seguidores",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Seguidores encontrados",
		Data:    followers,
	})
}

// GetFollowing godoc
// @Summary Get users that a user is following
// @Description Get the list of users that a specific user is following
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param limit query int false "Number of results per page" default(20)
// @Param offset query int false "Number of results to skip" default(0)
// @Success 200 {array} models.UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id}/following [get]
func (h *UserHandler) GetFollowing(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "ID inválido",
			Message: "O ID do usuário deve ser um número válido",
		})
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit <= 0 {
		limit = 20
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	following, err := h.userService.GetFollowing(uint(userID), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Erro ao buscar usuários seguidos",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Usuários seguidos encontrados",
		Data:    following,
	})
}

// ChangePassword godoc
// @Summary Change user password
// @Description Change the password of the authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ChangePasswordRequest true "Password change data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/change-password [put]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Não autorizado",
			Message: "Token inválido",
		})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Dados inválidos",
			Message: err.Error(),
		})
		return
	}

	err := h.userService.ChangePassword(userID.(uint), req.OldPassword, req.NewPassword)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMsg := err.Error()

		switch {
		case contains(errorMsg, "senha atual incorreta"):
			statusCode = http.StatusBadRequest
		case contains(errorMsg, "deve ter pelo menos"), contains(errorMsg, "deve ter no máximo"):
			statusCode = http.StatusBadRequest
		case contains(errorMsg, "não encontrado"):
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   "Erro ao alterar senha",
			Message: errorMsg,
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Senha alterada com sucesso",
		Data:    nil,
	})
}

// DeactivateAccount godoc
// @Summary Deactivate user account
// @Description Deactivate the authenticated user's account
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/deactivate [delete]
func (h *UserHandler) DeactivateAccount(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Não autorizado",
			Message: "Token inválido",
		})
		return
	}

	err := h.userService.DeactivateAccount(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Erro ao desativar conta",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Conta desativada com sucesso",
		Data:    nil,
	})
}

// Structs auxiliares
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}
