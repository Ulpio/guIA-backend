package repositories

import (
	"github.com/Ulpio/guIA-backend/internal/models"
	"gorm.io/gorm"
)

type ItineraryRepositoryInterface interface {
	Create(itinerary *models.Itinerary) error
	GetByID(id uint) (*models.Itinerary, error)
	Update(itinerary *models.Itinerary) error
	Delete(id uint) error
	GetByAuthor(authorID uint, limit, offset int) ([]models.Itinerary, error)
	GetByCategory(category models.ItineraryCategory, limit, offset int) ([]models.Itinerary, error)
	GetFeatured(limit, offset int) ([]models.Itinerary, error)
	GetTrending(limit, offset int) ([]models.Itinerary, error)
	SearchItineraries(query string, limit, offset int) ([]models.Itinerary, error)
	RateItinerary(userID, itineraryID uint, rating int, comment string) error
	GetUserRating(userID, itineraryID uint) (*models.ItineraryRating, error)
	UpdateRating(userID, itineraryID uint, rating int, comment string) error
	DeleteRating(userID, itineraryID uint) error
	IncrementViews(id uint) error
	GetSimilar(itineraryID uint, limit int) ([]models.Itinerary, error)
}

type ItineraryRepository struct {
	db *gorm.DB
}

func NewItineraryRepository(db *gorm.DB) ItineraryRepositoryInterface {
	return &ItineraryRepository{db: db}
}

func (r *ItineraryRepository) Create(itinerary *models.Itinerary) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Criar o roteiro
		if err := tx.Create(itinerary).Error; err != nil {
			return err
		}

		// Atualizar contador de roteiros do usuário
		return tx.Model(&models.User{}).Where("id = ?", itinerary.AuthorID).
			Update("itineraries_count", gorm.Expr("itineraries_count + 1")).Error
	})
}

func (r *ItineraryRepository) GetByID(id uint) (*models.Itinerary, error) {
	var itinerary models.Itinerary
	err := r.db.Preload("Author").
		Preload("Days").
		Preload("Days.Locations").
		Preload("Ratings").
		Preload("Ratings.User").
		Where("id = ?", id).
		First(&itinerary).Error
	if err != nil {
		return nil, err
	}
	return &itinerary, nil
}

func (r *ItineraryRepository) Update(itinerary *models.Itinerary) error {
	return r.db.Save(itinerary).Error
}

func (r *ItineraryRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Buscar o roteiro para obter o author_id
		var itinerary models.Itinerary
		if err := tx.Where("id = ?", id).First(&itinerary).Error; err != nil {
			return err
		}

		// Soft delete do roteiro
		if err := tx.Delete(&models.Itinerary{}, id).Error; err != nil {
			return err
		}

		// Atualizar contador de roteiros do usuário
		return tx.Model(&models.User{}).Where("id = ?", itinerary.AuthorID).
			Update("itineraries_count", gorm.Expr("itineraries_count - 1")).Error
	})
}

func (r *ItineraryRepository) GetByAuthor(authorID uint, limit, offset int) ([]models.Itinerary, error) {
	var itineraries []models.Itinerary
	err := r.db.Preload("Author").
		Where("author_id = ? AND is_public = ?", authorID, true).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&itineraries).Error
	return itineraries, err
}

func (r *ItineraryRepository) GetByCategory(category models.ItineraryCategory, limit, offset int) ([]models.Itinerary, error) {
	var itineraries []models.Itinerary
	err := r.db.Preload("Author").
		Where("category = ? AND is_public = ?", category, true).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&itineraries).Error
	return itineraries, err
}

func (r *ItineraryRepository) GetFeatured(limit, offset int) ([]models.Itinerary, error) {
	var itineraries []models.Itinerary
	err := r.db.Preload("Author").
		Where("is_featured = ? AND is_public = ?", true, true).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&itineraries).Error
	return itineraries, err
}

func (r *ItineraryRepository) GetTrending(limit, offset int) ([]models.Itinerary, error) {
	var itineraries []models.Itinerary

	// Roteiros trending baseado em visualizações, curtidas e avaliações recentes
	err := r.db.Preload("Author").
		Where("is_public = ? AND created_at > NOW() - INTERVAL '30 days'", true).
		Order("(views_count + likes_count * 2 + ratings_count * 3) DESC, average_rating DESC, created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&itineraries).Error

	return itineraries, err
}

func (r *ItineraryRepository) SearchItineraries(query string, limit, offset int) ([]models.Itinerary, error) {
	var itineraries []models.Itinerary
	searchQuery := "%" + query + "%"
	err := r.db.Preload("Author").
		Where("(title ILIKE ? OR description ILIKE ? OR city ILIKE ? OR country ILIKE ?) AND is_public = ?",
			searchQuery, searchQuery, searchQuery, searchQuery, true).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&itineraries).Error
	return itineraries, err
}

func (r *ItineraryRepository) RateItinerary(userID, itineraryID uint, rating int, comment string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Criar a avaliação
		itineraryRating := &models.ItineraryRating{
			ItineraryID: itineraryID,
			UserID:      userID,
			Rating:      rating,
			Comment:     comment,
		}

		if err := tx.Create(itineraryRating).Error; err != nil {
			return err
		}

		// Recalcular média e contador de avaliações
		return r.updateItineraryRatingStats(tx, itineraryID)
	})
}

func (r *ItineraryRepository) GetUserRating(userID, itineraryID uint) (*models.ItineraryRating, error) {
	var rating models.ItineraryRating
	err := r.db.Where("user_id = ? AND itinerary_id = ?", userID, itineraryID).First(&rating).Error
	if err != nil {
		return nil, err
	}
	return &rating, nil
}

func (r *ItineraryRepository) UpdateRating(userID, itineraryID uint, rating int, comment string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Atualizar a avaliação
		err := tx.Model(&models.ItineraryRating{}).
			Where("user_id = ? AND itinerary_id = ?", userID, itineraryID).
			Updates(map[string]interface{}{
				"rating":  rating,
				"comment": comment,
			}).Error

		if err != nil {
			return err
		}

		// Recalcular média e contador de avaliações
		return r.updateItineraryRatingStats(tx, itineraryID)
	})
}

func (r *ItineraryRepository) DeleteRating(userID, itineraryID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Deletar a avaliação
		err := tx.Where("user_id = ? AND itinerary_id = ?", userID, itineraryID).
			Delete(&models.ItineraryRating{}).Error

		if err != nil {
			return err
		}

		// Recalcular média e contador de avaliações
		return r.updateItineraryRatingStats(tx, itineraryID)
	})
}

func (r *ItineraryRepository) IncrementViews(id uint) error {
	return r.db.Model(&models.Itinerary{}).Where("id = ?", id).
		Update("views_count", gorm.Expr("views_count + 1")).Error
}

func (r *ItineraryRepository) GetSimilar(itineraryID uint, limit int) ([]models.Itinerary, error) {
	// Buscar roteiro original para obter categoria e localização
	var originalItinerary models.Itinerary
	if err := r.db.Where("id = ?", itineraryID).First(&originalItinerary).Error; err != nil {
		return nil, err
	}

	var itineraries []models.Itinerary
	err := r.db.Preload("Author").
		Where("id != ? AND (category = ? OR city = ? OR country = ?) AND is_public = ?",
			itineraryID, originalItinerary.Category, originalItinerary.City, originalItinerary.Country, true).
		Order("average_rating DESC, views_count DESC").
		Limit(limit).
		Find(&itineraries).Error

	return itineraries, err
}

// Função auxiliar para recalcular estatísticas de avaliação
func (r *ItineraryRepository) updateItineraryRatingStats(tx *gorm.DB, itineraryID uint) error {
	var avgRating float64
	var ratingsCount int64

	// Calcular média e contagem
	err := tx.Model(&models.ItineraryRating{}).
		Where("itinerary_id = ?", itineraryID).
		Count(&ratingsCount).Error
	if err != nil {
		return err
	}

	if ratingsCount > 0 {
		err = tx.Model(&models.ItineraryRating{}).
			Where("itinerary_id = ?", itineraryID).
			Select("AVG(rating)").
			Row().Scan(&avgRating)
		if err != nil {
			return err
		}
	}

	// Atualizar roteiro
	return tx.Model(&models.Itinerary{}).Where("id = ?", itineraryID).
		Updates(map[string]interface{}{
			"average_rating": avgRating,
			"ratings_count":  ratingsCount,
		}).Error
}
