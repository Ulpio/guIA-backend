package repositories

import (
	"github.com/Ulpio/guIA-backend/internal/models"
	"gorm.io/gorm"
)

type UserRepositoryInterface interface {
	Create(user *models.User) error
	GetByID(id uint) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	Update(user *models.User) error
	Delete(id uint) error
	GetFollowers(userID uint, limit, offset int) ([]models.User, error)
	GetFollowing(userID uint, limit, offset int) ([]models.User, error)
	FollowUser(followerID, followedID uint) error
	UnfollowUser(followerID, followedID uint) error
	IsFollowing(followerID, followedID uint) (bool, error)
	SearchUsers(query string, limit, offset int) ([]models.User, error)
	UpdateCounts(userID uint) error
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepositoryInterface {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) GetByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ? AND is_active = ?", id, true).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ? AND is_active = ?", email, true).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ? AND is_active = ?", username, true).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepository) Delete(id uint) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).Update("is_active", false).Error
}

func (r *UserRepository) GetFollowers(userID uint, limit, offset int) ([]models.User, error) {
	var users []models.User
	err := r.db.Table("users").
		Joins("JOIN follows ON users.id = follows.follower_id").
		Where("follows.followed_id = ? AND users.is_active = ?", userID, true).
		Limit(limit).
		Offset(offset).
		Find(&users).Error
	return users, err
}

func (r *UserRepository) GetFollowing(userID uint, limit, offset int) ([]models.User, error) {
	var users []models.User
	err := r.db.Table("users").
		Joins("JOIN follows ON users.id = follows.followed_id").
		Where("follows.follower_id = ? AND users.is_active = ?", userID, true).
		Limit(limit).
		Offset(offset).
		Find(&users).Error
	return users, err
}

func (r *UserRepository) FollowUser(followerID, followedID uint) error {
	if followerID == followedID {
		return gorm.ErrInvalidData
	}

	follow := &models.Follow{
		FollowerID: followerID,
		FollowedID: followedID,
	}

	// Usar transação para garantir consistência
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Criar o follow
		if err := tx.Create(follow).Error; err != nil {
			return err
		}

		// Atualizar contadores
		if err := tx.Model(&models.User{}).Where("id = ?", followerID).
			Update("following_count", gorm.Expr("following_count + 1")).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.User{}).Where("id = ?", followedID).
			Update("followers_count", gorm.Expr("followers_count + 1")).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *UserRepository) UnfollowUser(followerID, followedID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Deletar o follow
		if err := tx.Where("follower_id = ? AND followed_id = ?", followerID, followedID).
			Delete(&models.Follow{}).Error; err != nil {
			return err
		}

		// Atualizar contadores
		if err := tx.Model(&models.User{}).Where("id = ?", followerID).
			Update("following_count", gorm.Expr("following_count - 1")).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.User{}).Where("id = ?", followedID).
			Update("followers_count", gorm.Expr("followers_count - 1")).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *UserRepository) IsFollowing(followerID, followedID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.Follow{}).
		Where("follower_id = ? AND followed_id = ?", followerID, followedID).
		Count(&count).Error
	return count > 0, err
}

func (r *UserRepository) SearchUsers(query string, limit, offset int) ([]models.User, error) {
	var users []models.User
	searchQuery := "%" + query + "%"
	err := r.db.Where("(username ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ? OR company_name ILIKE ?) AND is_active = ?",
		searchQuery, searchQuery, searchQuery, searchQuery, true).
		Limit(limit).
		Offset(offset).
		Find(&users).Error
	return users, err
}

func (r *UserRepository) UpdateCounts(userID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var postsCount, itinerariesCount, followersCount, followingCount int64

		// Contar posts
		tx.Model(&models.Post{}).Where("author_id = ? AND deleted_at IS NULL", userID).Count(&postsCount)

		// Contar roteiros
		tx.Model(&models.Itinerary{}).Where("author_id = ? AND deleted_at IS NULL", userID).Count(&itinerariesCount)

		// Contar seguidores
		tx.Model(&models.Follow{}).Where("followed_id = ?", userID).Count(&followersCount)

		// Contar seguindo
		tx.Model(&models.Follow{}).Where("follower_id = ?", userID).Count(&followingCount)

		// Atualizar usuário
		return tx.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
			"posts_count":       postsCount,
			"itineraries_count": itinerariesCount,
			"followers_count":   followersCount,
			"following_count":   followingCount,
		}).Error
	})
}
