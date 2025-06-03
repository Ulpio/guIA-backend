package handlers

import (
	"net/http"
	"strconv"

	"github.com/Ulpio/guIA-backend/internal/models"
	"github.com/Ulpio/guIA-backend/internal/services"
	"github.com/gin-gonic/gin"
)

type ItineraryHandler struct {
	itineraryService services.ItineraryServiceInterface
}

func NewItineraryHandler(itineraryService services.ItineraryServiceInterface) *ItineraryHandler {
	return &ItineraryHandler{
		itineraryService: itineraryService,
	}
}

// CreateItinerary godoc
// @Summary Create a new itinerary
// @Description Create a new travel itinerary with days and locations
// @Tags itineraries
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.CreateItineraryRequest true "Itinerary creation data"
// @Success 201 {object} models.ItineraryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /itineraries [post]
func (h *ItineraryHandler) CreateItinerary(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Não autorizado",
			Message: "Token inválido",
		})
		return
	}

	var req services.CreateItineraryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Dados inválidos",
			Message: err.Error(),
		})
		return
	}

	itinerary, err := h.itineraryService.CreateItinerary(userID.(uint), &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMsg := err.Error()

		if contains(errorMsg, "obrigatório") || contains(errorMsg, "inválido") || contains(errorMsg, "deve ter") {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   "Erro ao criar roteiro",
			Message: errorMsg,
		})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Message: "Roteiro criado com sucesso",
		Data:    itinerary,
	})
}

// GetItineraries godoc
// @Summary Get itineraries with filters
// @Description Get a list of itineraries with optional filters
// @Tags itineraries
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param category query string false "Filter by category"
// @Param country query string false "Filter by country"
// @Param city query string false "Filter by city"
// @Param min_duration query int false "Minimum duration in days"
// @Param max_duration query int false "Maximum duration in days"
// @Param difficulty query int false "Filter by difficulty (1-5)"
// @Param featured query bool false "Show only featured itineraries"
// @Param order_by query string false "Order by: recent, popular, rating" default(recent)
// @Param limit query int false "Number of results per page" default(20)
// @Param offset query int false "Number of results to skip" default(0)
// @Success 200 {array} models.ItineraryResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /itineraries [get]
func (h *ItineraryHandler) GetItineraries(c *gin.Context) {
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Não autorizado",
			Message: "Token inválido",
		})
		return
	}

	// Parse filters
	filters := &services.ItineraryFilters{
		Category:   models.ItineraryCategory(c.Query("category")),
		Country:    c.Query("country"),
		City:       c.Query("city"),
		OrderBy:    c.DefaultQuery("order_by", "recent"),
		IsFeatured: c.Query("featured") == "true",
	}

	// Parse numeric filters
	if minDuration := c.Query("min_duration"); minDuration != "" {
		if val, err := strconv.Atoi(minDuration); err == nil {
			filters.MinDuration = val
		}
	}

	if maxDuration := c.Query("max_duration"); maxDuration != "" {
		if val, err := strconv.Atoi(maxDuration); err == nil {
			filters.MaxDuration = val
		}
	}

	if difficulty := c.Query("difficulty"); difficulty != "" {
		if val, err := strconv.Atoi(difficulty); err == nil {
			filters.Difficulty = val
		}
	}

	// Parse pagination
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit <= 0 {
		limit = 20
	}
	filters.Limit = limit

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}
	filters.Offset = offset

	itineraries, err := h.itineraryService.GetItineraries(filters, currentUserID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Erro ao buscar roteiros",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Roteiros encontrados",
		Data:    itineraries,
	})
}

// GetItineraryByID godoc
// @Summary Get itinerary by ID
// @Description Get a specific itinerary by its ID
// @Tags itineraries
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Itinerary ID"
// @Success 200 {object} models.ItineraryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /itineraries/{id} [get]
func (h *ItineraryHandler) GetItineraryByID(c *gin.Context) {
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Não autorizado",
			Message: "Token inválido",
		})
		return
	}

	idParam := c.Param("id")
	itineraryID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "ID inválido",
			Message: "O ID do roteiro deve ser um número válido",
		})
		return
	}

	itinerary, err := h.itineraryService.GetItineraryByID(uint(itineraryID), currentUserID.(uint))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if contains(err.Error(), "não encontrado") {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   "Erro ao buscar roteiro",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Roteiro encontrado",
		Data:    itinerary,
	})
}

// UpdateItinerary godoc
// @Summary Update an itinerary
// @Description Update an existing itinerary (only by the author)
// @Tags itineraries
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Itinerary ID"
// @Param request body services.UpdateItineraryRequest true "Itinerary update data"
// @Success 200 {object} models.ItineraryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /itineraries/{id} [put]
func (h *ItineraryHandler) UpdateItinerary(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Não autorizado",
			Message: "Token inválido",
		})
		return
	}

	idParam := c.Param("id")
	itineraryID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "ID inválido",
			Message: "O ID do roteiro deve ser um número válido",
		})
		return
	}

	var req services.UpdateItineraryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Dados inválidos",
			Message: err.Error(),
		})
		return
	}

	itinerary, err := h.itineraryService.UpdateItinerary(uint(itineraryID), userID.(uint), &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMsg := err.Error()

		switch {
		case contains(errorMsg, "não encontrado"):
			statusCode = http.StatusNotFound
		case contains(errorMsg, "não tem permissão"):
			statusCode = http.StatusForbidden
		case contains(errorMsg, "inválido"), contains(errorMsg, "deve ter"):
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   "Erro ao atualizar roteiro",
			Message: errorMsg,
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Roteiro atualizado com sucesso",
		Data:    itinerary,
	})
}

// DeleteItinerary godoc
// @Summary Delete an itinerary
// @Description Delete an existing itinerary (only by the author)
// @Tags itineraries
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Itinerary ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /itineraries/{id} [delete]
func (h *ItineraryHandler) DeleteItinerary(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Não autorizado",
			Message: "Token inválido",
		})
		return
	}

	idParam := c.Param("id")
	itineraryID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "ID inválido",
			Message: "O ID do roteiro deve ser um número válido",
		})
		return
	}

	err = h.itineraryService.DeleteItinerary(uint(itineraryID), userID.(uint))
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMsg := err.Error()

		switch {
		case contains(errorMsg, "não encontrado"):
			statusCode = http.StatusNotFound
		case contains(errorMsg, "não tem permissão"):
			statusCode = http.StatusForbidden
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   "Erro ao deletar roteiro",
			Message: errorMsg,
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Roteiro deletado com sucesso",
		Data:    nil,
	})
}

// RateItinerary godoc
// @Summary Rate an itinerary
// @Description Rate a specific itinerary (1-5 stars)
// @Tags itineraries
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Itinerary ID"
// @Param request body RateItineraryRequest true "Rating data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /itineraries/{id}/rate [post]
func (h *ItineraryHandler) RateItinerary(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Não autorizado",
			Message: "Token inválido",
		})
		return
	}

	idParam := c.Param("id")
	itineraryID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "ID inválido",
			Message: "O ID do roteiro deve ser um número válido",
		})
		return
	}

	var req RateItineraryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Dados inválidos",
			Message: err.Error(),
		})
		return
	}

	err = h.itineraryService.RateItinerary(userID.(uint), uint(itineraryID), req.Rating, req.Comment)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMsg := err.Error()

		switch {
		case contains(errorMsg, "não encontrado"):
			statusCode = http.StatusNotFound
		case contains(errorMsg, "já avaliou"):
			statusCode = http.StatusConflict
		case contains(errorMsg, "deve estar entre"), contains(errorMsg, "inválido"):
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   "Erro ao avaliar roteiro",
			Message: errorMsg,
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Roteiro avaliado com sucesso",
		Data:    nil,
	})
}

// UpdateRating godoc
// @Summary Update itinerary rating
// @Description Update an existing rating for an itinerary
// @Tags itineraries
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Itinerary ID"
// @Param request body RateItineraryRequest true "Updated rating data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /itineraries/{id}/rate [put]
func (h *ItineraryHandler) UpdateRating(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Não autorizado",
			Message: "Token inválido",
		})
		return
	}

	idParam := c.Param("id")
	itineraryID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "ID inválido",
			Message: "O ID do roteiro deve ser um número válido",
		})
		return
	}

	var req RateItineraryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Dados inválidos",
			Message: err.Error(),
		})
		return
	}

	err = h.itineraryService.UpdateRating(userID.(uint), uint(itineraryID), req.Rating, req.Comment)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMsg := err.Error()

		switch {
		case contains(errorMsg, "ainda não avaliou"):
			statusCode = http.StatusNotFound
		case contains(errorMsg, "deve estar entre"), contains(errorMsg, "inválido"):
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   "Erro ao atualizar avaliação",
			Message: errorMsg,
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Avaliação atualizada com sucesso",
		Data:    nil,
	})
}

// DeleteRating godoc
// @Summary Delete itinerary rating
// @Description Delete an existing rating for an itinerary
// @Tags itineraries
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Itinerary ID"
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /itineraries/{id}/rate [delete]
func (h *ItineraryHandler) DeleteRating(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Não autorizado",
			Message: "Token inválido",
		})
		return
	}

	idParam := c.Param("id")
	itineraryID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "ID inválido",
			Message: "O ID do roteiro deve ser um número válido",
		})
		return
	}

	err = h.itineraryService.DeleteRating(userID.(uint), uint(itineraryID))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if contains(err.Error(), "ainda não avaliou") {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   "Erro ao deletar avaliação",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Avaliação removida com sucesso",
		Data:    nil,
	})
}

// SearchItineraries godoc
// @Summary Search itineraries
// @Description Search for itineraries by title, description, city or country
// @Tags itineraries
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param q query string true "Search query"
// @Param limit query int false "Number of results per page" default(20)
// @Param offset query int false "Number of results to skip" default(0)
// @Success 200 {array} models.ItineraryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /itineraries/search [get]
func (h *ItineraryHandler) SearchItineraries(c *gin.Context) {
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Não autorizado",
			Message: "Token inválido",
		})
		return
	}

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

	itineraries, err := h.itineraryService.SearchItineraries(query, currentUserID.(uint), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Erro na busca de roteiros",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Busca realizada com sucesso",
		Data:    itineraries,
	})
}

// GetItinerariesByAuthor godoc
// @Summary Get itineraries by author
// @Description Get all itineraries from a specific author
// @Tags itineraries
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param authorId query int true "Author ID"
// @Param limit query int false "Number of results per page" default(20)
// @Param offset query int false "Number of results to skip" default(0)
// @Success 200 {array} models.ItineraryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /itineraries/author [get]
func (h *ItineraryHandler) GetItinerariesByAuthor(c *gin.Context) {
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Não autorizado",
			Message: "Token inválido",
		})
		return
	}

	authorIDParam := c.Query("authorId")
	if authorIDParam == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Parâmetro obrigatório",
			Message: "O parâmetro 'authorId' é obrigatório",
		})
		return
	}

	authorID, err := strconv.ParseUint(authorIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "ID inválido",
			Message: "O ID do autor deve ser um número válido",
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

	itineraries, err := h.itineraryService.GetItinerariesByAuthor(uint(authorID), currentUserID.(uint), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Erro ao buscar roteiros do autor",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Roteiros encontrados",
		Data:    itineraries,
	})
}

// GetSimilarItineraries godoc
// @Summary Get similar itineraries
// @Description Get itineraries similar to a specific one
// @Tags itineraries
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Itinerary ID"
// @Param limit query int false "Number of results" default(5)
// @Success 200 {array} models.ItineraryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /itineraries/{id}/similar [get]
func (h *ItineraryHandler) GetSimilarItineraries(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Não autorizado",
			Message: "Token inválido",
		})
		return
	}

	idParam := c.Param("id")
	itineraryID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "ID inválido",
			Message: "O ID do roteiro deve ser um número válido",
		})
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "5"))
	if err != nil || limit <= 0 {
		limit = 5
	}

	itineraries, err := h.itineraryService.GetSimilarItineraries(uint(itineraryID), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Erro ao buscar roteiros similares",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Roteiros similares encontrados",
		Data:    itineraries,
	})
}

// Structs auxiliares
type RateItineraryRequest struct {
	Rating  int    `json:"rating" binding:"required,min=1,max=5"`
	Comment string `json:"comment"`
}
