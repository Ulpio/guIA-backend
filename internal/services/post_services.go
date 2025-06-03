package services

import (
	"errors"
	"strings"

	"github.com/Ulpio/guIA-backend/internal/models"
	"github.com/Ulpio/guIA-backend/internal/repositories"
)

type PostServiceInterface interface {
	CreatePost(userID uint, req *CreatePostRequest) (*models.PostResponse, error)
	GetFeed(userID uint, limit, offset int) ([]models.PostResponse, error)
	GetPostByID(postID, userID uint) (*models.PostResponse, error)
	UpdatePost(postID, userID uint, req *UpdatePostRequest) (*models.PostResponse, error)
	DeletePost(postID, userID uint) error
	LikePost(userID, postID uint) error
	UnlikePost(userID, postID uint) error
	GetPostsByAuthor(authorID, currentUserID uint, limit, offset int) ([]models.PostResponse, error)
	SearchPosts(query string, currentUserID uint, limit, offset int) ([]models.PostResponse, error)
	GetTrendingPosts(currentUserID uint, limit, offset int) ([]models.PostResponse, error)
}

type CreatePostRequest struct {
	Content   string          `json:"content" binding:"required"`
	PostType  models.PostType `json:"post_type"`
	MediaURLs []string        `json:"media_urls,omitempty"`
	Location  string          `json:"location,omitempty"`
	Latitude  *float64        `json:"latitude,omitempty"`
	Longitude *float64        `json:"longitude,omitempty"`
}

type UpdatePostRequest struct {
	Content   *string  `json:"content,omitempty"`
	Location  *string  `json:"location,omitempty"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
}

type PostService struct {
	postRepo repositories.PostRepositoryInterface
	userRepo repositories.UserRepositoryInterface
}

func NewPostService(postRepo repositories.PostRepositoryInterface) PostServiceInterface {
	return &PostService{
		postRepo: postRepo,
	}
}

func (s *PostService) CreatePost(userID uint, req *CreatePostRequest) (*models.PostResponse, error) {
	// Validações
	if err := s.validateCreatePostRequest(req); err != nil {
		return nil, err
	}

	// Determinar tipo do post baseado na mídia
	postType := models.PostTypeText
	if len(req.MediaURLs) > 0 {
		// Por simplicidade, vamos assumir que se há mídia, é imagem
		// Em um sistema real, você verificaria o tipo do arquivo
		postType = models.PostTypeImage
	}

	if req.PostType != "" {
		postType = req.PostType
	}

	// Criar post
	post := &models.Post{
		AuthorID:  userID,
		Content:   strings.TrimSpace(req.Content),
		PostType:  postType,
		MediaURLs: req.MediaURLs,
		Location:  req.Location,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		IsActive:  true,
	}

	// Para compatibilidade, definir MediaURL como primeira URL se existir
	if len(req.MediaURLs) > 0 {
		post.MediaURL = req.MediaURLs[0]
	}

	if err := s.postRepo.Create(post); err != nil {
		return nil, errors.New("erro ao criar post")
	}

	// Buscar post criado com dados completos
	createdPost, err := s.postRepo.GetByID(post.ID)
	if err != nil {
		return nil, errors.New("erro ao buscar post criado")
	}

	return createdPost.ToResponse(userID), nil
}

func (s *PostService) GetFeed(userID uint, limit, offset int) ([]models.PostResponse, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	posts, err := s.postRepo.GetFeed(userID, limit, offset)
	if err != nil {
		return nil, errors.New("erro ao buscar feed")
	}

	var responses []models.PostResponse
	for _, post := range posts {
		responses = append(responses, *post.ToResponse(userID))
	}

	return responses, nil
}

func (s *PostService) GetPostByID(postID, userID uint) (*models.PostResponse, error) {
	post, err := s.postRepo.GetByID(postID)
	if err != nil {
		return nil, errors.New("post não encontrado")
	}

	return post.ToResponse(userID), nil
}

func (s *PostService) UpdatePost(postID, userID uint, req *UpdatePostRequest) (*models.PostResponse, error) {
	// Buscar post
	post, err := s.postRepo.GetByID(postID)
	if err != nil {
		return nil, errors.New("post não encontrado")
	}

	// Verificar se o usuário é o autor
	if post.AuthorID != userID {
		return nil, errors.New("você não tem permissão para editar este post")
	}

	// Validar e atualizar campos
	if req.Content != nil {
		content := strings.TrimSpace(*req.Content)
		if err := s.validateContent(content); err != nil {
			return nil, err
		}
		post.Content = content
	}

	if req.Location != nil {
		post.Location = *req.Location
	}

	if req.Latitude != nil {
		post.Latitude = req.Latitude
	}

	if req.Longitude != nil {
		post.Longitude = req.Longitude
	}

	if err := s.postRepo.Update(post); err != nil {
		return nil, errors.New("erro ao atualizar post")
	}

	// Buscar post atualizado
	updatedPost, err := s.postRepo.GetByID(postID)
	if err != nil {
		return nil, errors.New("erro ao buscar post atualizado")
	}

	return updatedPost.ToResponse(userID), nil
}

func (s *PostService) DeletePost(postID, userID uint) error {
	// Buscar post
	post, err := s.postRepo.GetByID(postID)
	if err != nil {
		return errors.New("post não encontrado")
	}

	// Verificar se o usuário é o autor
	if post.AuthorID != userID {
		return errors.New("você não tem permissão para deletar este post")
	}

	return s.postRepo.Delete(postID)
}

func (s *PostService) LikePost(userID, postID uint) error {
	// Verificar se o post existe
	_, err := s.postRepo.GetByID(postID)
	if err != nil {
		return errors.New("post não encontrado")
	}

	// Verificar se já curtiu
	isLiked, err := s.postRepo.IsLiked(userID, postID)
	if err != nil {
		return errors.New("erro ao verificar curtida")
	}

	if isLiked {
		return errors.New("você já curtiu este post")
	}

	return s.postRepo.LikePost(userID, postID)
}

func (s *PostService) UnlikePost(userID, postID uint) error {
	// Verificar se o post existe
	_, err := s.postRepo.GetByID(postID)
	if err != nil {
		return errors.New("post não encontrado")
	}

	// Verificar se curtiu
	isLiked, err := s.postRepo.IsLiked(userID, postID)
	if err != nil {
		return errors.New("erro ao verificar curtida")
	}

	if !isLiked {
		return errors.New("você não curtiu este post")
	}

	return s.postRepo.UnlikePost(userID, postID)
}

func (s *PostService) GetPostsByAuthor(authorID, currentUserID uint, limit, offset int) ([]models.PostResponse, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	posts, err := s.postRepo.GetByAuthor(authorID, limit, offset)
	if err != nil {
		return nil, errors.New("erro ao buscar posts do usuário")
	}

	var responses []models.PostResponse
	for _, post := range posts {
		responses = append(responses, *post.ToResponse(currentUserID))
	}

	return responses, nil
}

func (s *PostService) SearchPosts(query string, currentUserID uint, limit, offset int) ([]models.PostResponse, error) {
	if strings.TrimSpace(query) == "" {
		return []models.PostResponse{}, nil
	}

	if limit <= 0 || limit > 50 {
		limit = 20
	}

	posts, err := s.postRepo.SearchPosts(query, limit, offset)
	if err != nil {
		return nil, errors.New("erro ao buscar posts")
	}

	var responses []models.PostResponse
	for _, post := range posts {
		responses = append(responses, *post.ToResponse(currentUserID))
	}

	return responses, nil
}

func (s *PostService) GetTrendingPosts(currentUserID uint, limit, offset int) ([]models.PostResponse, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	posts, err := s.postRepo.GetTrendingPosts(limit, offset)
	if err != nil {
		return nil, errors.New("erro ao buscar posts em alta")
	}

	var responses []models.PostResponse
	for _, post := range posts {
		responses = append(responses, *post.ToResponse(currentUserID))
	}

	return responses, nil
}

// Funções de validação
func (s *PostService) validateCreatePostRequest(req *CreatePostRequest) error {
	if err := s.validateContent(req.Content); err != nil {
		return err
	}

	// Validar tipo de post
	if req.PostType != "" {
		if req.PostType != models.PostTypeText &&
			req.PostType != models.PostTypeImage &&
			req.PostType != models.PostTypeVideo {
			return errors.New("tipo de post inválido")
		}
	}

	// Validar URLs de mídia
	if len(req.MediaURLs) > 10 {
		return errors.New("máximo de 10 mídias por post")
	}

	for _, url := range req.MediaURLs {
		if err := s.validateMediaURL(url); err != nil {
			return err
		}
	}

	// Validar localização
	if req.Location != "" && len(req.Location) > 200 {
		return errors.New("localização deve ter no máximo 200 caracteres")
	}

	// Validar coordenadas
	if req.Latitude != nil && (*req.Latitude < -90 || *req.Latitude > 90) {
		return errors.New("latitude deve estar entre -90 e 90")
	}

	if req.Longitude != nil && (*req.Longitude < -180 || *req.Longitude > 180) {
		return errors.New("longitude deve estar entre -180 e 180")
	}

	return nil
}

func (s *PostService) validateContent(content string) error {
	content = strings.TrimSpace(content)
	if content == "" {
		return errors.New("conteúdo é obrigatório")
	}

	if len(content) > 2000 {
		return errors.New("conteúdo deve ter no máximo 2000 caracteres")
	}

	return nil
}

func (s *PostService) validateMediaURL(url string) error {
	if url == "" {
		return errors.New("URL de mídia não pode ser vazia")
	}

	if len(url) > 500 {
		return errors.New("URL de mídia deve ter no máximo 500 caracteres")
	}

	// Validação básica de URL
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return errors.New("URL de mídia deve começar com http:// ou https://")
	}

	return nil
}
