package services

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/Ulpio/guIA-backend/internal/models"
	"github.com/Ulpio/guIA-backend/internal/repositories"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthServiceInterface interface {
	Register(req *RegisterRequest) (*AuthResponse, error)
	Login(req *LoginRequest) (*AuthResponse, error)
	ValidateToken(tokenString string) (*TokenClaims, error)
	RefreshToken(tokenString string) (*AuthResponse, error)
}

type RegisterRequest struct {
	Username    string          `json:"username" binding:"required"`
	Email       string          `json:"email" binding:"required,email"`
	Password    string          `json:"password" binding:"required"`
	FirstName   string          `json:"first_name" binding:"required"`
	LastName    string          `json:"last_name" binding:"required"`
	UserType    models.UserType `json:"user_type"`
	CompanyName string          `json:"company_name,omitempty"`
}

type LoginRequest struct {
	Login    string `json:"login" binding:"required"` // email ou username
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token        string               `json:"token"`
	RefreshToken string               `json:"refresh_token"`
	User         *models.UserResponse `json:"user"`
	ExpiresAt    time.Time            `json:"expires_at"`
}

type TokenClaims struct {
	UserID   uint            `json:"user_id"`
	Username string          `json:"username"`
	UserType models.UserType `json:"user_type"`
	jwt.RegisteredClaims
}

type AuthService struct {
	userRepo  repositories.UserRepositoryInterface
	jwtSecret string
}

func NewAuthService(userRepo repositories.UserRepositoryInterface, jwtSecret string) AuthServiceInterface {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) Register(req *RegisterRequest) (*AuthResponse, error) {
	// Validações
	if err := s.validateRegisterRequest(req); err != nil {
		return nil, err
	}

	// Verificar se email já existe
	if _, err := s.userRepo.GetByEmail(req.Email); err == nil {
		return nil, errors.New("email já está em uso")
	}

	// Verificar se username já existe
	if _, err := s.userRepo.GetByUsername(req.Username); err == nil {
		return nil, errors.New("nome de usuário já está em uso")
	}

	// Hash da senha
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("erro ao processar senha")
	}

	// Criar usuário
	user := &models.User{
		Username:  strings.ToLower(req.Username),
		Email:     strings.ToLower(req.Email),
		Password:  string(hashedPassword),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		UserType:  req.UserType,
		IsActive:  true,
	}

	// Se for empresa, adicionar nome da empresa
	if req.UserType == models.UserTypeCompany && req.CompanyName != "" {
		user.CompanyName = req.CompanyName
	}

	// Definir tipo padrão se não especificado
	if user.UserType == "" {
		user.UserType = models.UserTypeNormal
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, errors.New("erro ao criar usuário")
	}

	// Gerar tokens
	token, refreshToken, expiresAt, err := s.generateTokens(user)
	if err != nil {
		return nil, errors.New("erro ao gerar token de acesso")
	}

	return &AuthResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         user.ToResponse(),
		ExpiresAt:    expiresAt,
	}, nil
}

func (s *AuthService) Login(req *LoginRequest) (*AuthResponse, error) {
	// Validações
	if err := s.validateLoginRequest(req); err != nil {
		return nil, err
	}

	var user *models.User
	var err error

	// Tentar buscar por email primeiro, depois por username
	if s.isEmail(req.Login) {
		user, err = s.userRepo.GetByEmail(strings.ToLower(req.Login))
	} else {
		user, err = s.userRepo.GetByUsername(strings.ToLower(req.Login))
	}

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("credenciais inválidas")
		}
		return nil, errors.New("erro ao buscar usuário")
	}

	// Verificar senha
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("credenciais inválidas")
	}

	// Verificar se conta está ativa
	if !user.IsActive {
		return nil, errors.New("conta desativada")
	}

	// Gerar tokens
	token, refreshToken, expiresAt, err := s.generateTokens(user)
	if err != nil {
		return nil, errors.New("erro ao gerar token de acesso")
	}

	return &AuthResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         user.ToResponse(),
		ExpiresAt:    expiresAt,
	}, nil
}

func (s *AuthService) ValidateToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("token inválido")
}

func (s *AuthService) RefreshToken(tokenString string) (*AuthResponse, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, errors.New("token de refresh inválido")
	}

	// Buscar usuário atual
	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		return nil, errors.New("usuário não encontrado")
	}

	if !user.IsActive {
		return nil, errors.New("conta desativada")
	}

	// Gerar novos tokens
	token, refreshToken, expiresAt, err := s.generateTokens(user)
	if err != nil {
		return nil, errors.New("erro ao gerar novo token")
	}

	return &AuthResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         user.ToResponse(),
		ExpiresAt:    expiresAt,
	}, nil
}

// Funções auxiliares
func (s *AuthService) generateTokens(user *models.User) (string, string, time.Time, error) {
	expiresAt := time.Now().Add(24 * time.Hour)            // 24 horas
	refreshExpiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 dias

	// Token principal
	claims := &TokenClaims{
		UserID:   user.ID,
		Username: user.Username,
		UserType: user.UserType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "guia-backend",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", "", time.Time{}, err
	}

	// Refresh token
	refreshClaims := &TokenClaims{
		UserID:   user.ID,
		Username: user.Username,
		UserType: user.UserType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "guia-backend-refresh",
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", "", time.Time{}, err
	}

	return tokenString, refreshTokenString, expiresAt, nil
}

func (s *AuthService) validateRegisterRequest(req *RegisterRequest) error {
	// Validar username
	if err := s.validateUsername(req.Username); err != nil {
		return err
	}

	// Validar email
	if err := s.validateEmail(req.Email); err != nil {
		return err
	}

	// Validar senha
	if err := s.validatePassword(req.Password); err != nil {
		return err
	}

	// Validar nomes
	if err := s.validateName(req.FirstName, "primeiro nome"); err != nil {
		return err
	}

	if err := s.validateName(req.LastName, "sobrenome"); err != nil {
		return err
	}

	// Validar tipo de usuário
	if req.UserType != "" && req.UserType != models.UserTypeNormal && req.UserType != models.UserTypeCompany {
		return errors.New("tipo de usuário inválido")
	}

	// Se for empresa, validar nome da empresa
	if req.UserType == models.UserTypeCompany {
		if req.CompanyName == "" {
			return errors.New("nome da empresa é obrigatório para contas empresariais")
		}
		if err := s.validateCompanyName(req.CompanyName); err != nil {
			return err
		}
	}

	return nil
}

func (s *AuthService) validateLoginRequest(req *LoginRequest) error {
	if strings.TrimSpace(req.Login) == "" {
		return errors.New("email ou nome de usuário é obrigatório")
	}

	if strings.TrimSpace(req.Password) == "" {
		return errors.New("senha é obrigatória")
	}

	return nil
}

func (s *AuthService) validateUsername(username string) error {
	username = strings.TrimSpace(username)
	if len(username) < 3 {
		return errors.New("nome de usuário deve ter pelo menos 3 caracteres")
	}
	if len(username) > 50 {
		return errors.New("nome de usuário deve ter no máximo 50 caracteres")
	}

	// Apenas letras, números e underscore
	matched, _ := regexp.MatchString("^[a-zA-Z0-9_]+$", username)
	if !matched {
		return errors.New("nome de usuário deve conter apenas letras, números e underscore")
	}

	return nil
}

func (s *AuthService) validateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return errors.New("email é obrigatório")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("formato de email inválido")
	}

	return nil
}

func (s *AuthService) validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("senha deve ter pelo menos 8 caracteres")
	}
	if len(password) > 100 {
		return errors.New("senha deve ter no máximo 100 caracteres")
	}
	return nil
}

func (s *AuthService) validateName(name, fieldName string) error {
	name = strings.TrimSpace(name)
	if len(name) < 2 {
		return errors.New(fieldName + " deve ter pelo menos 2 caracteres")
	}
	if len(name) > 50 {
		return errors.New(fieldName + " deve ter no máximo 50 caracteres")
	}
	return nil
}

func (s *AuthService) validateCompanyName(companyName string) error {
	companyName = strings.TrimSpace(companyName)
	if len(companyName) < 2 {
		return errors.New("nome da empresa deve ter pelo menos 2 caracteres")
	}
	if len(companyName) > 100 {
		return errors.New("nome da empresa deve ter no máximo 100 caracteres")
	}
	return nil
}

func (s *AuthService) isEmail(login string) bool {
	return strings.Contains(login, "@")
}
