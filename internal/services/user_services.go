package services

import (
	"errors"
	"strings"

	"github.com/Ulpio/guIA-backend/internal/models"
	"github.com/Ulpio/guIA-backend/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

type UserServiceInterface interface {
	GetProfile(userID uint) (*models.UserResponse, error)
	UpdateProfile(userID uint, updateData *UpdateProfileRequest) (*models.UserResponse, error)
	GetUserByID(userID uint) (*models.UserResponse, error)
	SearchUsers(query string, limit, offset int) ([]models.UserResponse, error)
	FollowUser(followerID, followedID uint) error
	UnfollowUser(followerID, followedID uint) error
	GetFollowers(userID uint, limit, offset int) ([]models.UserResponse, error)
	GetFollowing(userID uint, limit, offset int) ([]models.UserResponse, error)
	IsFollowing(followerID, followedID uint) (bool, error)
	ChangePassword(userID uint, oldPassword, newPassword string) error
	DeactivateAccount(userID uint) error
}

type UpdateProfileRequest struct {
	FirstName      *string `json:"first_name,omitempty"`
	LastName       *string `json:"last_name,omitempty"`
	Bio            *string `json:"bio,omitempty"`
	Location       *string `json:"location,omitempty"`
	Website        *string `json:"website,omitempty"`
	ProfilePicture *string `json:"profile_picture,omitempty"`
	CompanyName    *string `json:"company_name,omitempty"`
}

type UserService struct {
	userRepo repositories.UserRepositoryInterface
}

func NewUserService(userRepo repositories.UserRepositoryInterface) UserServiceInterface {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) GetProfile(userID uint) (*models.UserResponse, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("usuário não encontrado")
	}

	return user.ToResponse(), nil
}

func (s *UserService) UpdateProfile(userID uint, updateData *UpdateProfileRequest) (*models.UserResponse, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("usuário não encontrado")
	}

	// Validar e atualizar campos
	if updateData.FirstName != nil {
		if err := s.validateName(*updateData.FirstName); err != nil {
			return nil, err
		}
		user.FirstName = *updateData.FirstName
	}

	if updateData.LastName != nil {
		if err := s.validateName(*updateData.LastName); err != nil {
			return nil, err
		}
		user.LastName = *updateData.LastName
	}

	if updateData.Bio != nil {
		if err := s.validateBio(*updateData.Bio); err != nil {
			return nil, err
		}
		user.Bio = *updateData.Bio
	}

	if updateData.Location != nil {
		user.Location = *updateData.Location
	}

	if updateData.Website != nil {
		if err := s.validateWebsite(*updateData.Website); err != nil {
			return nil, err
		}
		user.Website = *updateData.Website
	}

	if updateData.ProfilePicture != nil {
		user.ProfilePicture = *updateData.ProfilePicture
	}

	if updateData.CompanyName != nil && user.UserType == models.UserTypeCompany {
		if err := s.validateCompanyName(*updateData.CompanyName); err != nil {
			return nil, err
		}
		user.CompanyName = *updateData.CompanyName
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, errors.New("erro ao atualizar perfil")
	}

	return user.ToResponse(), nil
}

func (s *UserService) GetUserByID(userID uint) (*models.UserResponse, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("usuário não encontrado")
	}

	return user.ToResponse(), nil
}

func (s *UserService) SearchUsers(query string, limit, offset int) ([]models.UserResponse, error) {
	if strings.TrimSpace(query) == "" {
		return []models.UserResponse{}, nil
	}

	if limit <= 0 || limit > 50 {
		limit = 20
	}

	users, err := s.userRepo.SearchUsers(query, limit, offset)
	if err != nil {
		return nil, errors.New("erro ao buscar usuários")
	}

	var responses []models.UserResponse
	for _, user := range users {
		responses = append(responses, *user.ToResponse())
	}

	return responses, nil
}

func (s *UserService) FollowUser(followerID, followedID uint) error {
	if followerID == followedID {
		return errors.New("você não pode seguir a si mesmo")
	}

	// Verificar se o usuário a ser seguido existe
	_, err := s.userRepo.GetByID(followedID)
	if err != nil {
		return errors.New("usuário não encontrado")
	}

	// Verificar se já está seguindo
	isFollowing, err := s.userRepo.IsFollowing(followerID, followedID)
	if err != nil {
		return errors.New("erro ao verificar se já está seguindo")
	}

	if isFollowing {
		return errors.New("você já está seguindo este usuário")
	}

	return s.userRepo.FollowUser(followerID, followedID)
}

func (s *UserService) UnfollowUser(followerID, followedID uint) error {
	if followerID == followedID {
		return errors.New("operação inválida")
	}

	// Verificar se está seguindo
	isFollowing, err := s.userRepo.IsFollowing(followerID, followedID)
	if err != nil {
		return errors.New("erro ao verificar se está seguindo")
	}

	if !isFollowing {
		return errors.New("você não está seguindo este usuário")
	}

	return s.userRepo.UnfollowUser(followerID, followedID)
}

func (s *UserService) GetFollowers(userID uint, limit, offset int) ([]models.UserResponse, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	users, err := s.userRepo.GetFollowers(userID, limit, offset)
	if err != nil {
		return nil, errors.New("erro ao buscar seguidores")
	}

	var responses []models.UserResponse
	for _, user := range users {
		responses = append(responses, *user.ToResponse())
	}

	return responses, nil
}

func (s *UserService) GetFollowing(userID uint, limit, offset int) ([]models.UserResponse, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	users, err := s.userRepo.GetFollowing(userID, limit, offset)
	if err != nil {
		return nil, errors.New("erro ao buscar usuários seguidos")
	}

	var responses []models.UserResponse
	for _, user := range users {
		responses = append(responses, *user.ToResponse())
	}

	return responses, nil
}

func (s *UserService) IsFollowing(followerID, followedID uint) (bool, error) {
	return s.userRepo.IsFollowing(followerID, followedID)
}

func (s *UserService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("usuário não encontrado")
	}

	// Verificar senha atual
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("senha atual incorreta")
	}

	// Validar nova senha
	if err := s.validatePassword(newPassword); err != nil {
		return err
	}

	// Hash da nova senha
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("erro ao processar nova senha")
	}

	user.Password = string(hashedPassword)
	return s.userRepo.Update(user)
}

func (s *UserService) DeactivateAccount(userID uint) error {
	return s.userRepo.Delete(userID)
}

// Funções de validação
func (s *UserService) validateName(name string) error {
	name = strings.TrimSpace(name)
	if len(name) < 2 {
		return errors.New("nome deve ter pelo menos 2 caracteres")
	}
	if len(name) > 50 {
		return errors.New("nome deve ter no máximo 50 caracteres")
	}
	return nil
}

func (s *UserService) validateBio(bio string) error {
	if len(bio) > 500 {
		return errors.New("bio deve ter no máximo 500 caracteres")
	}
	return nil
}

func (s *UserService) validateWebsite(website string) error {
	if website == "" {
		return nil
	}

	if !strings.HasPrefix(website, "http://") && !strings.HasPrefix(website, "https://") {
		return errors.New("website deve começar com http:// ou https://")
	}

	if len(website) > 200 {
		return errors.New("website deve ter no máximo 200 caracteres")
	}

	return nil
}

func (s *UserService) validateCompanyName(companyName string) error {
	companyName = strings.TrimSpace(companyName)
	if len(companyName) < 2 {
		return errors.New("nome da empresa deve ter pelo menos 2 caracteres")
	}
	if len(companyName) > 100 {
		return errors.New("nome da empresa deve ter no máximo 100 caracteres")
	}
	return nil
}

func (s *UserService) validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("senha deve ter pelo menos 8 caracteres")
	}
	if len(password) > 100 {
		return errors.New("senha deve ter no máximo 100 caracteres")
	}
	return nil
}
