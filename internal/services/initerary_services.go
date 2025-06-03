package services

import (
	"errors"
	"strings"

	"github.com/Ulpio/guIA-backend/internal/models"
	"github.com/Ulpio/guIA-backend/internal/repositories"
)

type ItineraryServiceInterface interface {
	CreateItinerary(userID uint, req *CreateItineraryRequest) (*models.ItineraryResponse, error)
	GetItineraryByID(itineraryID, currentUserID uint) (*models.ItineraryResponse, error)
	UpdateItinerary(itineraryID, userID uint, req *UpdateItineraryRequest) (*models.ItineraryResponse, error)
	DeleteItinerary(itineraryID, userID uint) error
	GetItineraries(filters *ItineraryFilters, currentUserID uint) ([]models.ItineraryResponse, error)
	GetItinerariesByAuthor(authorID, currentUserID uint, limit, offset int) ([]models.ItineraryResponse, error)
	SearchItineraries(query string, currentUserID uint, limit, offset int) ([]models.ItineraryResponse, error)
	RateItinerary(userID, itineraryID uint, rating int, comment string) error
	UpdateRating(userID, itineraryID uint, rating int, comment string) error
	DeleteRating(userID, itineraryID uint) error
	GetSimilarItineraries(itineraryID uint, limit int) ([]models.ItineraryResponse, error)
}

type CreateItineraryRequest struct {
	Title         string                      `json:"title" binding:"required"`
	Description   string                      `json:"description"`
	Category      models.ItineraryCategory    `json:"category" binding:"required"`
	EstimatedCost *float64                    `json:"estimated_cost"`
	Currency      string                      `json:"currency"`
	Duration      int                         `json:"duration" binding:"required"`
	Difficulty    int                         `json:"difficulty"`
	CoverImage    string                      `json:"cover_image"`
	Images        []string                    `json:"images"`
	Country       string                      `json:"country" binding:"required"`
	City          string                      `json:"city"`
	State         string                      `json:"state"`
	IsPublic      bool                        `json:"is_public"`
	Days          []CreateItineraryDayRequest `json:"days"`
}

type CreateItineraryDayRequest struct {
	DayNumber     int                              `json:"day_number" binding:"required"`
	Title         string                           `json:"title"`
	Description   string                           `json:"description"`
	EstimatedCost *float64                         `json:"estimated_cost"`
	Locations     []CreateItineraryLocationRequest `json:"locations"`
}

type CreateItineraryLocationRequest struct {
	Name          string              `json:"name" binding:"required"`
	Description   string              `json:"description"`
	LocationType  models.LocationType `json:"location_type" binding:"required"`
	Address       string              `json:"address"`
	Latitude      *float64            `json:"latitude"`
	Longitude     *float64            `json:"longitude"`
	GooglePlaceID string              `json:"google_place_id"`
	EstimatedCost *float64            `json:"estimated_cost"`
	StartTime     string              `json:"start_time"`
	EndTime       string              `json:"end_time"`
	Order         int                 `json:"order"`
	Images        []string            `json:"images"`
	Website       string              `json:"website"`
	Phone         string              `json:"phone"`
	Rating        *float64            `json:"rating"`
}

type UpdateItineraryRequest struct {
	Title         *string                   `json:"title,omitempty"`
	Description   *string                   `json:"description,omitempty"`
	Category      *models.ItineraryCategory `json:"category,omitempty"`
	EstimatedCost *float64                  `json:"estimated_cost,omitempty"`
	Currency      *string                   `json:"currency,omitempty"`
	Duration      *int                      `json:"duration,omitempty"`
	Difficulty    *int                      `json:"difficulty,omitempty"`
	CoverImage    *string                   `json:"cover_image,omitempty"`
	Images        []string                  `json:"images,omitempty"`
	Country       *string                   `json:"country,omitempty"`
	City          *string                   `json:"city,omitempty"`
	State         *string                   `json:"state,omitempty"`
	IsPublic      *bool                     `json:"is_public,omitempty"`
}

type ItineraryFilters struct {
	Category    models.ItineraryCategory `json:"category"`
	Country     string                   `json:"country"`
	City        string                   `json:"city"`
	MinDuration int                      `json:"min_duration"`
	MaxDuration int                      `json:"max_duration"`
	MinCost     float64                  `json:"min_cost"`
	MaxCost     float64                  `json:"max_cost"`
	Difficulty  int                      `json:"difficulty"`
	IsFeatured  bool                     `json:"is_featured"`
	OrderBy     string                   `json:"order_by"` // "recent", "popular", "rating"
	Limit       int                      `json:"limit"`
	Offset      int                      `json:"offset"`
}

type ItineraryService struct {
	itineraryRepo repositories.ItineraryRepositoryInterface
}

func NewItineraryService(itineraryRepo repositories.ItineraryRepositoryInterface) ItineraryServiceInterface {
	return &ItineraryService{
		itineraryRepo: itineraryRepo,
	}
}

func (s *ItineraryService) CreateItinerary(userID uint, req *CreateItineraryRequest) (*models.ItineraryResponse, error) {
	// Validações
	if err := s.validateCreateItineraryRequest(req); err != nil {
		return nil, err
	}

	// Criar roteiro
	itinerary := &models.Itinerary{
		AuthorID:      userID,
		Title:         strings.TrimSpace(req.Title),
		Description:   strings.TrimSpace(req.Description),
		Category:      req.Category,
		EstimatedCost: req.EstimatedCost,
		Currency:      s.getDefaultCurrency(req.Currency),
		Duration:      req.Duration,
		Difficulty:    s.getDefaultDifficulty(req.Difficulty),
		CoverImage:    req.CoverImage,
		Images:        req.Images,
		Country:       strings.TrimSpace(req.Country),
		City:          strings.TrimSpace(req.City),
		State:         strings.TrimSpace(req.State),
		IsPublic:      req.IsPublic,
	}

	if err := s.itineraryRepo.Create(itinerary); err != nil {
		return nil, errors.New("erro ao criar roteiro")
	}

	// Criar dias e localizações se fornecidos
	if len(req.Days) > 0 {
		if err := s.createItineraryDays(itinerary.ID, req.Days); err != nil {
			return nil, err
		}
	}

	// Buscar roteiro criado com dados completos
	createdItinerary, err := s.itineraryRepo.GetByID(itinerary.ID)
	if err != nil {
		return nil, errors.New("erro ao buscar roteiro criado")
	}

	return createdItinerary.ToResponse(), nil
}

func (s *ItineraryService) GetItineraryByID(itineraryID, currentUserID uint) (*models.ItineraryResponse, error) {
	itinerary, err := s.itineraryRepo.GetByID(itineraryID)
	if err != nil {
		return nil, errors.New("roteiro não encontrado")
	}

	// Verificar se o roteiro é público ou se o usuário é o autor
	if !itinerary.IsPublic && itinerary.AuthorID != currentUserID {
		return nil, errors.New("roteiro não encontrado")
	}

	// Incrementar visualizações se não for o autor
	if itinerary.AuthorID != currentUserID {
		s.itineraryRepo.IncrementViews(itineraryID)
	}

	return itinerary.ToResponse(), nil
}

func (s *ItineraryService) UpdateItinerary(itineraryID, userID uint, req *UpdateItineraryRequest) (*models.ItineraryResponse, error) {
	// Buscar roteiro
	itinerary, err := s.itineraryRepo.GetByID(itineraryID)
	if err != nil {
		return nil, errors.New("roteiro não encontrado")
	}

	// Verificar se o usuário é o autor
	if itinerary.AuthorID != userID {
		return nil, errors.New("você não tem permissão para editar este roteiro")
	}

	// Validar e atualizar campos
	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		if err := s.validateTitle(title); err != nil {
			return nil, err
		}
		itinerary.Title = title
	}

	if req.Description != nil {
		itinerary.Description = strings.TrimSpace(*req.Description)
	}

	if req.Category != nil {
		if err := s.validateCategory(*req.Category); err != nil {
			return nil, err
		}
		itinerary.Category = *req.Category
	}

	if req.EstimatedCost != nil {
		itinerary.EstimatedCost = req.EstimatedCost
	}

	if req.Currency != nil {
		itinerary.Currency = *req.Currency
	}

	if req.Duration != nil {
		if err := s.validateDuration(*req.Duration); err != nil {
			return nil, err
		}
		itinerary.Duration = *req.Duration
	}

	if req.Difficulty != nil {
		if err := s.validateDifficulty(*req.Difficulty); err != nil {
			return nil, err
		}
		itinerary.Difficulty = *req.Difficulty
	}

	if req.CoverImage != nil {
		itinerary.CoverImage = *req.CoverImage
	}

	if req.Images != nil {
		itinerary.Images = req.Images
	}

	if req.Country != nil {
		country := strings.TrimSpace(*req.Country)
		if err := s.validateCountry(country); err != nil {
			return nil, err
		}
		itinerary.Country = country
	}

	if req.City != nil {
		itinerary.City = strings.TrimSpace(*req.City)
	}

	if req.State != nil {
		itinerary.State = strings.TrimSpace(*req.State)
	}

	if req.IsPublic != nil {
		itinerary.IsPublic = *req.IsPublic
	}

	if err := s.itineraryRepo.Update(itinerary); err != nil {
		return nil, errors.New("erro ao atualizar roteiro")
	}

	// Buscar roteiro atualizado
	updatedItinerary, err := s.itineraryRepo.GetByID(itineraryID)
	if err != nil {
		return nil, errors.New("erro ao buscar roteiro atualizado")
	}

	return updatedItinerary.ToResponse(), nil
}

func (s *ItineraryService) DeleteItinerary(itineraryID, userID uint) error {
	// Buscar roteiro
	itinerary, err := s.itineraryRepo.GetByID(itineraryID)
	if err != nil {
		return errors.New("roteiro não encontrado")
	}

	// Verificar se o usuário é o autor
	if itinerary.AuthorID != userID {
		return errors.New("você não tem permissão para deletar este roteiro")
	}

	return s.itineraryRepo.Delete(itineraryID)
}

func (s *ItineraryService) GetItineraries(filters *ItineraryFilters, currentUserID uint) ([]models.ItineraryResponse, error) {
	var itineraries []models.Itinerary
	var err error

	// Definir defaults
	if filters.Limit <= 0 || filters.Limit > 50 {
		filters.Limit = 20
	}

	// Buscar baseado nos filtros
	switch {
	case filters.Category != "":
		itineraries, err = s.itineraryRepo.GetByCategory(filters.Category, filters.Limit, filters.Offset)
	case filters.IsFeatured:
		itineraries, err = s.itineraryRepo.GetFeatured(filters.Limit, filters.Offset)
	case filters.OrderBy == "popular":
		itineraries, err = s.itineraryRepo.GetTrending(filters.Limit, filters.Offset)
	default:
		// Implementar busca mais complexa com múltiplos filtros no futuro
		itineraries, err = s.itineraryRepo.GetTrending(filters.Limit, filters.Offset)
	}

	if err != nil {
		return nil, errors.New("erro ao buscar roteiros")
	}

	var responses []models.ItineraryResponse
	for _, itinerary := range itineraries {
		responses = append(responses, *itinerary.ToResponse())
	}

	return responses, nil
}

func (s *ItineraryService) GetItinerariesByAuthor(authorID, currentUserID uint, limit, offset int) ([]models.ItineraryResponse, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	itineraries, err := s.itineraryRepo.GetByAuthor(authorID, limit, offset)
	if err != nil {
		return nil, errors.New("erro ao buscar roteiros do usuário")
	}

	var responses []models.ItineraryResponse
	for _, itinerary := range itineraries {
		responses = append(responses, *itinerary.ToResponse())
	}

	return responses, nil
}

func (s *ItineraryService) SearchItineraries(query string, currentUserID uint, limit, offset int) ([]models.ItineraryResponse, error) {
	if strings.TrimSpace(query) == "" {
		return []models.ItineraryResponse{}, nil
	}

	if limit <= 0 || limit > 50 {
		limit = 20
	}

	itineraries, err := s.itineraryRepo.SearchItineraries(query, limit, offset)
	if err != nil {
		return nil, errors.New("erro ao buscar roteiros")
	}

	var responses []models.ItineraryResponse
	for _, itinerary := range itineraries {
		responses = append(responses, *itinerary.ToResponse())
	}

	return responses, nil
}

func (s *ItineraryService) RateItinerary(userID, itineraryID uint, rating int, comment string) error {
	// Verificar se o roteiro existe
	itinerary, err := s.itineraryRepo.GetByID(itineraryID)
	if err != nil {
		return errors.New("roteiro não encontrado")
	}

	// Verificar se o roteiro é público
	if !itinerary.IsPublic {
		return errors.New("não é possível avaliar roteiros privados")
	}

	// Validar avaliação
	if err := s.validateRating(rating); err != nil {
		return err
	}

	// Verificar se já avaliou
	if _, err := s.itineraryRepo.GetUserRating(userID, itineraryID); err == nil {
		return errors.New("você já avaliou este roteiro")
	}

	return s.itineraryRepo.RateItinerary(userID, itineraryID, rating, strings.TrimSpace(comment))
}

func (s *ItineraryService) UpdateRating(userID, itineraryID uint, rating int, comment string) error {
	// Verificar se já avaliou
	if _, err := s.itineraryRepo.GetUserRating(userID, itineraryID); err != nil {
		return errors.New("você ainda não avaliou este roteiro")
	}

	// Validar avaliação
	if err := s.validateRating(rating); err != nil {
		return err
	}

	return s.itineraryRepo.UpdateRating(userID, itineraryID, rating, strings.TrimSpace(comment))
}

func (s *ItineraryService) DeleteRating(userID, itineraryID uint) error {
	// Verificar se já avaliou
	if _, err := s.itineraryRepo.GetUserRating(userID, itineraryID); err != nil {
		return errors.New("você ainda não avaliou este roteiro")
	}

	return s.itineraryRepo.DeleteRating(userID, itineraryID)
}

func (s *ItineraryService) GetSimilarItineraries(itineraryID uint, limit int) ([]models.ItineraryResponse, error) {
	if limit <= 0 || limit > 20 {
		limit = 5
	}

	itineraries, err := s.itineraryRepo.GetSimilar(itineraryID, limit)
	if err != nil {
		return nil, errors.New("erro ao buscar roteiros similares")
	}

	var responses []models.ItineraryResponse
	for _, itinerary := range itineraries {
		responses = append(responses, *itinerary.ToResponse())
	}

	return responses, nil
}

// Funções auxiliares e validações
func (s *ItineraryService) createItineraryDays(itineraryID uint, daysReq []CreateItineraryDayRequest) error {
	// Implementação simplificada - em um sistema real, usaria transação
	// e salvaria os dias no banco de dados
	return nil
}

func (s *ItineraryService) getDefaultCurrency(currency string) string {
	if currency == "" {
		return "BRL"
	}
	return currency
}

func (s *ItineraryService) getDefaultDifficulty(difficulty int) int {
	if difficulty == 0 {
		return 1
	}
	return difficulty
}

// Funções de validação
func (s *ItineraryService) validateCreateItineraryRequest(req *CreateItineraryRequest) error {
	if err := s.validateTitle(req.Title); err != nil {
		return err
	}

	if err := s.validateCategory(req.Category); err != nil {
		return err
	}

	if err := s.validateDuration(req.Duration); err != nil {
		return err
	}

	if req.Difficulty != 0 {
		if err := s.validateDifficulty(req.Difficulty); err != nil {
			return err
		}
	}

	if err := s.validateCountry(req.Country); err != nil {
		return err
	}

	return nil
}

func (s *ItineraryService) validateTitle(title string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return errors.New("título é obrigatório")
	}
	if len(title) > 200 {
		return errors.New("título deve ter no máximo 200 caracteres")
	}
	return nil
}

func (s *ItineraryService) validateCategory(category models.ItineraryCategory) error {
	validCategories := []models.ItineraryCategory{
		models.CategoryAdventure, models.CategoryCultural, models.CategoryGastronomic,
		models.CategoryNature, models.CategoryUrban, models.CategoryBeach,
		models.CategoryMountain, models.CategoryBusiness, models.CategoryFamily,
		models.CategoryRomantic,
	}

	for _, valid := range validCategories {
		if category == valid {
			return nil
		}
	}

	return errors.New("categoria inválida")
}

func (s *ItineraryService) validateDuration(duration int) error {
	if duration <= 0 {
		return errors.New("duração deve ser maior que zero")
	}
	if duration > 365 {
		return errors.New("duração não pode ser maior que 365 dias")
	}
	return nil
}

func (s *ItineraryService) validateDifficulty(difficulty int) error {
	if difficulty < 1 || difficulty > 5 {
		return errors.New("dificuldade deve estar entre 1 e 5")
	}
	return nil
}

func (s *ItineraryService) validateCountry(country string) error {
	country = strings.TrimSpace(country)
	if country == "" {
		return errors.New("país é obrigatório")
	}
	if len(country) > 100 {
		return errors.New("país deve ter no máximo 100 caracteres")
	}
	return nil
}

func (s *ItineraryService) validateRating(rating int) error {
	if rating < 1 || rating > 5 {
		return errors.New("avaliação deve estar entre 1 e 5")
	}
	return nil
}
