package repositories

import (
	"github.com/Ulpio/guIA-backend/internal/models"
	"gorm.io/gorm"
)

type PostRepositoryInterface interface {
	Create(post *models.Post) error
	GetByID(id uint) (*models.Post, error)
	Update(post *models.Post) error
	Delete(id uint) error
	GetFeed(userID uint, limit, offset int) ([]models.Post, error)
	GetByAuthor(authorID uint, limit, offset int) ([]models.Post, error)
	LikePost(userID, postID uint) error
	UnlikePost(userID, postID uint) error
	IsLiked(userID, postID uint) (bool, error)
	SearchPosts(query string, limit, offset int) ([]models.Post, error)
	GetTrendingPosts(limit, offset int) ([]models.Post, error)
}

type PostRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) PostRepositoryInterface {
	return &PostRepository{db: db}
}

func (r *PostRepository) Create(post *models.Post) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Criar o post
		if err := tx.Create(post).Error; err != nil {
			return err
		}

		// Atualizar contador de posts do usuário
		return tx.Model(&models.User{}).Where("id = ?", post.AuthorID).
			Update("posts_count", gorm.Expr("posts_count + 1")).Error
	})
}

func (r *PostRepository) GetByID(id uint) (*models.Post, error) {
	var post models.Post
	err := r.db.Preload("Author").
		Preload("Likes").
		Preload("Comments").
		Where("id = ? AND is_active = ?", id, true).
		First(&post).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *PostRepository) Update(post *models.Post) error {
	return r.db.Save(post).Error
}

func (r *PostRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Buscar o post para obter o author_id
		var post models.Post
		if err := tx.Where("id = ?", id).First(&post).Error; err != nil {
			return err
		}

		// Soft delete do post
		if err := tx.Delete(&models.Post{}, id).Error; err != nil {
			return err
		}

		// Atualizar contador de posts do usuário
		return tx.Model(&models.User{}).Where("id = ?", post.AuthorID).
			Update("posts_count", gorm.Expr("posts_count - 1")).Error
	})
}

func (r *PostRepository) GetFeed(userID uint, limit, offset int) ([]models.Post, error) {
	var posts []models.Post

	// Buscar posts dos usuários que o usuário segue + próprios posts
	err := r.db.Preload("Author").
		Preload("Likes").
		Where(`author_id IN (
			SELECT followed_id FROM follows WHERE follower_id = ?
			UNION
			SELECT ?
		) AND is_active = ?`, userID, userID, true).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&posts).Error

	return posts, err
}

func (r *PostRepository) GetByAuthor(authorID uint, limit, offset int) ([]models.Post, error) {
	var posts []models.Post
	err := r.db.Preload("Author").
		Preload("Likes").
		Where("author_id = ? AND is_active = ?", authorID, true).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&posts).Error
	return posts, err
}

func (r *PostRepository) LikePost(userID, postID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Verificar se já curtiu
		var existingLike models.PostLike
		err := tx.Where("user_id = ? AND post_id = ?", userID, postID).First(&existingLike).Error
		if err == nil {
			// Já curtiu, não fazer nada
			return nil
		}

		// Criar a curtida
		like := &models.PostLike{
			UserID: userID,
			PostID: postID,
		}
		if err := tx.Create(like).Error; err != nil {
			return err
		}

		// Atualizar contador de curtidas do post
		return tx.Model(&models.Post{}).Where("id = ?", postID).
			Update("likes_count", gorm.Expr("likes_count + 1")).Error
	})
}

func (r *PostRepository) UnlikePost(userID, postID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Deletar a curtida
		result := tx.Where("user_id = ? AND post_id = ?", userID, postID).Delete(&models.PostLike{})
		if result.Error != nil {
			return result.Error
		}

		// Se deletou alguma linha, atualizar contador
		if result.RowsAffected > 0 {
			return tx.Model(&models.Post{}).Where("id = ?", postID).
				Update("likes_count", gorm.Expr("likes_count - 1")).Error
		}

		return nil
	})
}

func (r *PostRepository) IsLiked(userID, postID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.PostLike{}).
		Where("user_id = ? AND post_id = ?", userID, postID).
		Count(&count).Error
	return count > 0, err
}

func (r *PostRepository) SearchPosts(query string, limit, offset int) ([]models.Post, error) {
	var posts []models.Post
	searchQuery := "%" + query + "%"
	err := r.db.Preload("Author").
		Preload("Likes").
		Where("(content ILIKE ? OR location ILIKE ?) AND is_active = ?", searchQuery, searchQuery, true).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&posts).Error
	return posts, err
}

func (r *PostRepository) GetTrendingPosts(limit, offset int) ([]models.Post, error) {
	var posts []models.Post

	// Posts trending baseado em curtidas e comentários recentes
	err := r.db.Preload("Author").
		Preload("Likes").
		Where("is_active = ? AND created_at > NOW() - INTERVAL '7 days'", true).
		Order("(likes_count * 2 + comments_count) DESC, created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&posts).Error

	return posts, err
}
