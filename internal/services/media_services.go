package services

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/uuid"
)

type MediaType string

const (
	MediaTypeImage MediaType = "image"
	MediaTypeVideo MediaType = "video"
)

type MediaServiceInterface interface {
	UploadFile(file *multipart.FileHeader, userID uint, mediaType MediaType) (*MediaUploadResponse, error)
	DeleteFile(filePath string) error
	GetFileURL(filePath string) string
	ValidateFile(file *multipart.FileHeader, mediaType MediaType) error
}

type MediaUploadResponse struct {
	URL       string    `json:"url"`
	FilePath  string    `json:"file_path"`
	FileName  string    `json:"file_name"`
	FileSize  int64     `json:"file_size"`
	MimeType  string    `json:"mime_type"`
	MediaType MediaType `json:"media_type"`
	Width     int       `json:"width,omitempty"`
	Height    int       `json:"height,omitempty"`
}

type MediaConfig struct {
	StorageType     string // "local" or "s3"
	LocalPath       string
	BaseURL         string
	MaxFileSize     int64
	AllowedImageExt []string
	AllowedVideoExt []string
	AWSConfig       *AWSConfig
}

type AWSConfig struct {
	Region    string
	Bucket    string
	AccessKey string
	SecretKey string
	CDNUrl    string
}

type MediaService struct {
	config *MediaConfig
}

func NewMediaService(config *MediaConfig) MediaServiceInterface {
	if config.MaxFileSize == 0 {
		config.MaxFileSize = 50 * 1024 * 1024 // 50MB default
	}

	if len(config.AllowedImageExt) == 0 {
		config.AllowedImageExt = []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	}

	if len(config.AllowedVideoExt) == 0 {
		config.AllowedVideoExt = []string{".mp4", ".avi", ".mov", ".wmv", ".webm"}
	}

	if config.LocalPath == "" {
		config.LocalPath = "./uploads"
	}

	return &MediaService{
		config: config,
	}
}

func (s *MediaService) UploadFile(file *multipart.FileHeader, userID uint, mediaType MediaType) (*MediaUploadResponse, error) {
	// Validar arquivo
	if err := s.ValidateFile(file, mediaType); err != nil {
		return nil, err
	}

	// Gerar nome único do arquivo
	fileName := s.generateFileName(file.Filename, userID)

	// Determinar diretório baseado no tipo de mídia
	var directory string
	switch mediaType {
	case MediaTypeImage:
		directory = "images"
	case MediaTypeVideo:
		directory = "videos"
	default:
		return nil, errors.New("tipo de mídia não suportado")
	}

	// Upload baseado no tipo de storage
	var filePath, url string
	var err error

	switch s.config.StorageType {
	case "s3":
		filePath, url, err = s.uploadToS3(file, fileName, directory)
	default: // local
		filePath, url, err = s.uploadToLocal(file, fileName, directory)
	}

	if err != nil {
		return nil, err
	}

	// Obter metadados do arquivo
	width, height, err := s.getImageDimensions(file, mediaType)
	if err != nil {
		// Log do erro, mas não falha o upload
		width, height = 0, 0
	}

	return &MediaUploadResponse{
		URL:       url,
		FilePath:  filePath,
		FileName:  fileName,
		FileSize:  file.Size,
		MimeType:  file.Header.Get("Content-Type"),
		MediaType: mediaType,
		Width:     width,
		Height:    height,
	}, nil
}

// ============================================================================
// UPLOAD LOCAL
// ============================================================================

func (s *MediaService) uploadToLocal(file *multipart.FileHeader, fileName, directory string) (string, string, error) {
	src, err := file.Open()
	if err != nil {
		return "", "", err
	}
	defer src.Close()

	// Criar diretório se não existir
	fullDir := filepath.Join(s.config.LocalPath, directory)
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		return "", "", err
	}

	// Caminho completo do arquivo
	filePath := filepath.Join(directory, fileName)
	fullPath := filepath.Join(s.config.LocalPath, filePath)

	// Criar arquivo
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", "", err
	}
	defer dst.Close()

	// Copiar dados
	if _, err := io.Copy(dst, src); err != nil {
		return "", "", err
	}

	// Gerar URL
	url := fmt.Sprintf("%s/%s", strings.TrimRight(s.config.BaseURL, "/"), filePath)

	return filePath, url, nil
}

// ============================================================================
// UPLOAD S3 (para uso futuro)
// ============================================================================

func (s *MediaService) uploadToS3(file *multipart.FileHeader, fileName, directory string) (string, string, error) {
	if s.config.AWSConfig == nil {
		return "", "", fmt.Errorf("configuração AWS não encontrada")
	}

	// Abrir arquivo
	src, err := file.Open()
	if err != nil {
		return "", "", err
	}
	defer src.Close()

	// Criar sessão AWS
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(s.config.AWSConfig.Region),
		Credentials: credentials.NewStaticCredentials(
			s.config.AWSConfig.AccessKey,
			s.config.AWSConfig.SecretKey,
			"",
		),
	})
	if err != nil {
		return "", "", err
	}

	// Criar uploader
	uploader := s3manager.NewUploader(sess)

	// Caminho do arquivo no S3
	s3Key := fmt.Sprintf("%s/%s", directory, fileName)

	// Determinar Content-Type
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = s.getContentTypeFromExtension(fileName)
	}

	// Upload
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(s.config.AWSConfig.Bucket),
		Key:         aws.String(s3Key),
		Body:        src,
		ContentType: aws.String(contentType),
		ACL:         aws.String("public-read"),
	})
	if err != nil {
		return "", "", err
	}

	return s3Key, result.Location, nil
}

// ============================================================================
// DELETE FILES
// ============================================================================

func (s *MediaService) DeleteFile(filePath string) error {
	switch s.config.StorageType {
	case "s3":
		return s.deleteFromS3(filePath)
	default: // local
		return s.deleteFromLocal(filePath)
	}
}

func (s *MediaService) deleteFromLocal(filePath string) error {
	fullPath := filepath.Join(s.config.LocalPath, filePath)
	return os.Remove(fullPath)
}

func (s *MediaService) deleteFromS3(filePath string) error {
	if s.config.AWSConfig == nil {
		return fmt.Errorf("configuração AWS não encontrada")
	}

	// Criar sessão AWS
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(s.config.AWSConfig.Region),
		Credentials: credentials.NewStaticCredentials(
			s.config.AWSConfig.AccessKey,
			s.config.AWSConfig.SecretKey,
			"",
		),
	})
	if err != nil {
		return err
	}

	// Criar cliente S3
	svc := s3.New(sess)

	// Deletar objeto
	_, err = svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.config.AWSConfig.Bucket),
		Key:    aws.String(filePath),
	})

	return err
}

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

func (s *MediaService) GetFileURL(filePath string) string {
	switch s.config.StorageType {
	case "s3":
		if s.config.AWSConfig != nil && s.config.AWSConfig.CDNUrl != "" {
			return fmt.Sprintf("%s/%s", strings.TrimRight(s.config.AWSConfig.CDNUrl, "/"), filePath)
		}
		return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s",
			s.config.AWSConfig.Bucket, s.config.AWSConfig.Region, filePath)
	default: // local
		return fmt.Sprintf("%s/%s", strings.TrimRight(s.config.BaseURL, "/"), filePath)
	}
}

func (s *MediaService) ValidateFile(file *multipart.FileHeader, mediaType MediaType) error {
	// Validar tamanho
	if file.Size > s.config.MaxFileSize {
		return fmt.Errorf("arquivo muito grande. Tamanho máximo: %d MB", s.config.MaxFileSize/(1024*1024))
	}

	// Validar extensão
	ext := strings.ToLower(filepath.Ext(file.Filename))

	var allowedExtensions []string
	switch mediaType {
	case MediaTypeImage:
		allowedExtensions = s.config.AllowedImageExt
	case MediaTypeVideo:
		allowedExtensions = s.config.AllowedVideoExt
	default:
		return errors.New("tipo de mídia não suportado")
	}

	for _, allowedExt := range allowedExtensions {
		if ext == allowedExt {
			return nil
		}
	}

	return fmt.Errorf("extensão de arquivo não permitida: %s. Extensões permitidas: %v",
		ext, allowedExtensions)
}

func (s *MediaService) generateFileName(originalName string, userID uint) string {
	ext := filepath.Ext(originalName)
	timestamp := time.Now().Unix()
	uuid := uuid.New().String()[:8]

	return fmt.Sprintf("%d_%d_%s%s", userID, timestamp, uuid, ext)
}

func (s *MediaService) getImageDimensions(file *multipart.FileHeader, mediaType MediaType) (int, int, error) {
	if mediaType != MediaTypeImage {
		return 0, 0, nil
	}

	// Por enquanto retorna 0,0 - pode ser implementado com bibliotecas de processamento de imagem
	return 0, 0, nil
}

func (s *MediaService) getContentTypeFromExtension(fileName string) string {
	ext := filepath.Ext(fileName)

	contentTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".mp4":  "video/mp4",
		".avi":  "video/x-msvideo",
		".mov":  "video/quicktime",
		".wmv":  "video/x-ms-wmv",
		".webm": "video/webm",
	}

	if contentType, exists := contentTypes[ext]; exists {
		return contentType
	}

	return "application/octet-stream"
}
