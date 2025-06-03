package models

import (
	"time"

	"gorm.io/gorm"
)

type PostType string

const (
	PostTypeText  PostType = "text"
	PostTypeImage PostType = "image"
	PostTypeVideo PostType = "video"
)

type Post struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	AuthorID      uint           `json:"author_id" gorm:"not null"`
	Content       string         `json:"content" gorm:"type:text"`
	PostType      PostType       `json:"post_type" gorm:"default:'text'"`
	MediaURL      string         `json:"media_url"`
	MediaURLs     []string       `json:"media_urls" gorm:"serializer:json"`
	Location      string         `json:"location" gorm:"size:200"`
	Latitude      *float64       `json:"latitude"`
	Longitude     *float64       `json:"longitude"`
	LikesCount    int            `json:"likes_count" gorm:"default:0"`
	CommentsCount int            `json:"comments_count" gorm:"default:0"`
	SharesCount   int            `json:"shares_count" gorm:"default:0"`
	IsActive      bool           `json:"is_active" gorm:"default:true"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	// Relacionamentos
	Author   User       `json:"author" gorm:"foreignKey:AuthorID"`
	Likes    []PostLike `json:"likes,omitempty" gorm:"foreignKey:PostID"`
	Comments []Comment  `json:"comments,omitempty" gorm:"foreignKey:PostID"`
}

type PostLike struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	PostID    uint      `json:"post_id" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`

	User User `json:"user" gorm:"foreignKey:UserID"`
	Post Post `json:"post" gorm:"foreignKey:PostID"`
}

type Comment struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	PostID    uint           `json:"post_id" gorm:"not null"`
	AuthorID  uint           `json:"author_id" gorm:"not null"`
	Content   string         `json:"content" gorm:"type:text;not null"`
	ParentID  *uint          `json:"parent_id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relacionamentos
	Post    Post      `json:"post" gorm:"foreignKey:PostID"`
	Author  User      `json:"author" gorm:"foreignKey:AuthorID"`
	Parent  *Comment  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Replies []Comment `json:"replies,omitempty" gorm:"foreignKey:ParentID"`
}

type PostResponse struct {
	ID            uint          `json:"id"`
	AuthorID      uint          `json:"author_id"`
	Content       string        `json:"content"`
	PostType      PostType      `json:"post_type"`
	MediaURL      string        `json:"media_url"`
	MediaURLs     []string      `json:"media_urls"`
	Location      string        `json:"location"`
	Latitude      *float64      `json:"latitude"`
	Longitude     *float64      `json:"longitude"`
	LikesCount    int           `json:"likes_count"`
	CommentsCount int           `json:"comments_count"`
	SharesCount   int           `json:"shares_count"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
	Author        *UserResponse `json:"author,omitempty"`
	IsLiked       bool          `json:"is_liked"`
}

func (p *Post) ToResponse(currentUserID uint) *PostResponse {
	response := &PostResponse{
		ID:            p.ID,
		AuthorID:      p.AuthorID,
		Content:       p.Content,
		PostType:      p.PostType,
		MediaURL:      p.MediaURL,
		MediaURLs:     p.MediaURLs,
		Location:      p.Location,
		Latitude:      p.Latitude,
		Longitude:     p.Longitude,
		LikesCount:    p.LikesCount,
		CommentsCount: p.CommentsCount,
		SharesCount:   p.SharesCount,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}

	if p.Author.ID != 0 {
		response.Author = p.Author.ToResponse()
	}

	// Verificar se o usuário atual curtiu o post
	for _, like := range p.Likes {
		if like.UserID == currentUserID {
			response.IsLiked = true
			break
		}
	}

	return response
}
