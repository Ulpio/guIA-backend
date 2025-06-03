package models

import (
	"time"

	"gorm.io/gorm"
)

type ItineraryCategory string

const (
	CategoryAdventure   ItineraryCategory = "adventure"
	CategoryCultural    ItineraryCategory = "cultural"
	CategoryGastronomic ItineraryCategory = "gastronomic"
	CategoryNature      ItineraryCategory = "nature"
	CategoryUrban       ItineraryCategory = "urban"
	CategoryBeach       ItineraryCategory = "beach"
	CategoryMountain    ItineraryCategory = "mountain"
	CategoryBusiness    ItineraryCategory = "business"
	CategoryFamily      ItineraryCategory = "family"
	CategoryRomantic    ItineraryCategory = "romantic"
)

type Itinerary struct {
	ID            uint              `json:"id" gorm:"primaryKey"`
	AuthorID      uint              `json:"author_id" gorm:"not null"`
	Title         string            `json:"title" gorm:"not null;size:200"`
	Description   string            `json:"description" gorm:"type:text"`
	Category      ItineraryCategory `json:"category" gorm:"not null"`
	EstimatedCost *float64          `json:"estimated_cost"`
	Currency      string            `json:"currency" gorm:"size:3;default:'BRL'"`
	Duration      int               `json:"duration"` // em dias
	Difficulty    int               `json:"difficulty" gorm:"check:difficulty >= 1 AND difficulty <= 5"`
	CoverImage    string            `json:"cover_image"`
	Images        []string          `json:"images" gorm:"serializer:json"`
	Country       string            `json:"country" gorm:"size:100"`
	City          string            `json:"city" gorm:"size:100"`
	State         string            `json:"state" gorm:"size:100"`
	IsPublic      bool              `json:"is_public" gorm:"default:true"`
	IsFeatured    bool              `json:"is_featured" gorm:"default:false"`
	ViewsCount    int               `json:"views_count" gorm:"default:0"`
	LikesCount    int               `json:"likes_count" gorm:"default:0"`
	RatingsCount  int               `json:"ratings_count" gorm:"default:0"`
	AverageRating float64           `json:"average_rating" gorm:"default:0"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	DeletedAt     gorm.DeletedAt    `json:"-" gorm:"index"`

	// Relacionamentos
	Author  User              `json:"author" gorm:"foreignKey:AuthorID"`
	Days    []ItineraryDay    `json:"days,omitempty" gorm:"foreignKey:ItineraryID;constraint:OnDelete:CASCADE"`
	Ratings []ItineraryRating `json:"ratings,omitempty" gorm:"foreignKey:ItineraryID"`
}

type ItineraryDay struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	ItineraryID   uint      `json:"itinerary_id" gorm:"not null"`
	DayNumber     int       `json:"day_number" gorm:"not null"`
	Title         string    `json:"title" gorm:"size:200"`
	Description   string    `json:"description" gorm:"type:text"`
	EstimatedCost *float64  `json:"estimated_cost"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Relacionamentos
	Itinerary Itinerary           `json:"itinerary" gorm:"foreignKey:ItineraryID"`
	Locations []ItineraryLocation `json:"locations,omitempty" gorm:"foreignKey:DayID;constraint:OnDelete:CASCADE"`
}

type LocationType string

const (
	LocationTypeHotel      LocationType = "hotel"
	LocationTypeRestaurant LocationType = "restaurant"
	LocationTypeAttraction LocationType = "attraction"
	LocationTypeTransport  LocationType = "transport"
	LocationTypeShopping   LocationType = "shopping"
	LocationTypeOther      LocationType = "other"
)

type ItineraryLocation struct {
	ID            uint         `json:"id" gorm:"primaryKey"`
	DayID         uint         `json:"day_id" gorm:"not null"`
	Name          string       `json:"name" gorm:"not null;size:200"`
	Description   string       `json:"description" gorm:"type:text"`
	LocationType  LocationType `json:"location_type" gorm:"not null"`
	Address       string       `json:"address" gorm:"size:300"`
	Latitude      *float64     `json:"latitude"`
	Longitude     *float64     `json:"longitude"`
	GooglePlaceID string       `json:"google_place_id" gorm:"size:100"`
	EstimatedCost *float64     `json:"estimated_cost"`
	StartTime     *time.Time   `json:"start_time"`
	EndTime       *time.Time   `json:"end_time"`
	Order         int          `json:"order" gorm:"default:0"`
	Images        []string     `json:"images" gorm:"serializer:json"`
	Website       string       `json:"website" gorm:"size:200"`
	Phone         string       `json:"phone" gorm:"size:20"`
	Rating        *float64     `json:"rating"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`

	// Relacionamentos
	Day ItineraryDay `json:"day" gorm:"foreignKey:DayID"`
}

type ItineraryRating struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	ItineraryID uint      `json:"itinerary_id" gorm:"not null"`
	UserID      uint      `json:"user_id" gorm:"not null"`
	Rating      int       `json:"rating" gorm:"not null;check:rating >= 1 AND rating <= 5"`
	Comment     string    `json:"comment" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relacionamentos
	Itinerary Itinerary `json:"itinerary" gorm:"foreignKey:ItineraryID"`
	User      User      `json:"user" gorm:"foreignKey:UserID"`
}

type ItineraryResponse struct {
	ID            uint              `json:"id"`
	AuthorID      uint              `json:"author_id"`
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	Category      ItineraryCategory `json:"category"`
	EstimatedCost *float64          `json:"estimated_cost"`
	Currency      string            `json:"currency"`
	Duration      int               `json:"duration"`
	Difficulty    int               `json:"difficulty"`
	CoverImage    string            `json:"cover_image"`
	Images        []string          `json:"images"`
	Country       string            `json:"country"`
	City          string            `json:"city"`
	State         string            `json:"state"`
	IsFeatured    bool              `json:"is_featured"`
	ViewsCount    int               `json:"views_count"`
	LikesCount    int               `json:"likes_count"`
	RatingsCount  int               `json:"ratings_count"`
	AverageRating float64           `json:"average_rating"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	Author        *UserResponse     `json:"author,omitempty"`
	Days          []ItineraryDay    `json:"days,omitempty"`
}

func (i *Itinerary) ToResponse() *ItineraryResponse {
	response := &ItineraryResponse{
		ID:            i.ID,
		AuthorID:      i.AuthorID,
		Title:         i.Title,
		Description:   i.Description,
		Category:      i.Category,
		EstimatedCost: i.EstimatedCost,
		Currency:      i.Currency,
		Duration:      i.Duration,
		Difficulty:    i.Difficulty,
		CoverImage:    i.CoverImage,
		Images:        i.Images,
		Country:       i.Country,
		City:          i.City,
		State:         i.State,
		IsFeatured:    i.IsFeatured,
		ViewsCount:    i.ViewsCount,
		LikesCount:    i.LikesCount,
		RatingsCount:  i.RatingsCount,
		AverageRating: i.AverageRating,
		CreatedAt:     i.CreatedAt,
		UpdatedAt:     i.UpdatedAt,
		Days:          i.Days,
	}

	if i.Author.ID != 0 {
		response.Author = i.Author.ToResponse()
	}

	return response
}
