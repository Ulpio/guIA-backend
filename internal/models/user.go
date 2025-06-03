package models

import (
	"time"

	"gorm.io/gorm"
)

type UserType string

const (
	UserTypeNormal  UserType = "normal"
	UserTypeCompany UserType = "company"
	UserTypeAdmin   UserType = "admin"
)

type User struct {
	ID               uint           `json:"id" gorm:"primaryKey"`
	Username         string         `json:"username" gorm:"uniqueIndex;not null;size:50"`
	Email            string         `json:"email" gorm:"uniqueIndex;not null;size:100"`
	Password         string         `json:"-" gorm:"not null"`
	FirstName        string         `json:"first_name" gorm:"size:50"`
	LastName         string         `json:"last_name" gorm:"size:50"`
	Bio              string         `json:"bio" gorm:"size:500"`
	ProfilePicture   string         `json:"profile_picture"`
	UserType         UserType       `json:"user_type" gorm:"default:'normal'"`
	IsVerified       bool           `json:"is_verified" gorm:"default:false"`
	IsActive         bool           `json:"is_active" gorm:"default:true"`
	Location         string         `json:"location" gorm:"size:100"`
	Website          string         `json:"website" gorm:"size:200"`
	CompanyName      string         `json:"company_name" gorm:"size:100"`
	CompanyDocument  string         `json:"company_document" gorm:"size:50"`
	FollowersCount   int            `json:"followers_count" gorm:"default:0"`
	FollowingCount   int            `json:"following_count" gorm:"default:0"`
	PostsCount       int            `json:"posts_count" gorm:"default:0"`
	ItinerariesCount int            `json:"itineraries_count" gorm:"default:0"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`

	// Relacionamentos
	Posts       []Post      `json:"posts,omitempty" gorm:"foreignKey:AuthorID"`
	Itineraries []Itinerary `json:"itineraries,omitempty" gorm:"foreignKey:AuthorID"`
	PostLikes   []PostLike  `json:"-" gorm:"foreignKey:UserID"`
	Comments    []Comment   `json:"-" gorm:"foreignKey:AuthorID"`
	Followers   []Follow    `json:"-" gorm:"foreignKey:FollowedID"`
	Following   []Follow    `json:"-" gorm:"foreignKey:FollowerID"`
}

type Follow struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	FollowerID uint      `json:"follower_id" gorm:"not null"`
	FollowedID uint      `json:"followed_id" gorm:"not null"`
	CreatedAt  time.Time `json:"created_at"`

	Follower User `json:"follower" gorm:"foreignKey:FollowerID"`
	Followed User `json:"followed" gorm:"foreignKey:FollowedID"`
}

// UserResponse para retornar dados sem informações sensíveis
type UserResponse struct {
	ID               uint      `json:"id"`
	Username         string    `json:"username"`
	Email            string    `json:"email"`
	FirstName        string    `json:"first_name"`
	LastName         string    `json:"last_name"`
	Bio              string    `json:"bio"`
	ProfilePicture   string    `json:"profile_picture"`
	UserType         UserType  `json:"user_type"`
	IsVerified       bool      `json:"is_verified"`
	Location         string    `json:"location"`
	Website          string    `json:"website"`
	CompanyName      string    `json:"company_name"`
	FollowersCount   int       `json:"followers_count"`
	FollowingCount   int       `json:"following_count"`
	PostsCount       int       `json:"posts_count"`
	ItinerariesCount int       `json:"itineraries_count"`
	CreatedAt        time.Time `json:"created_at"`
}

func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:               u.ID,
		Username:         u.Username,
		Email:            u.Email,
		FirstName:        u.FirstName,
		LastName:         u.LastName,
		Bio:              u.Bio,
		ProfilePicture:   u.ProfilePicture,
		UserType:         u.UserType,
		IsVerified:       u.IsVerified,
		Location:         u.Location,
		Website:          u.Website,
		CompanyName:      u.CompanyName,
		FollowersCount:   u.FollowersCount,
		FollowingCount:   u.FollowingCount,
		PostsCount:       u.PostsCount,
		ItinerariesCount: u.ItinerariesCount,
		CreatedAt:        u.CreatedAt,
	}
}
