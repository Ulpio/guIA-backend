package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/Ulpio/guIA-backend/internal/services"
	"github.com/gin-gonic/gin"
)

type MediaHandler struct {
	mediaService services.MediaServiceInterface
}

func NewMediaHandler(mediaService services.MediaServiceInterface) *MediaHandler {
	return &MediaHandler{
		mediaService: mediaService,
	}
}

// UploadImage godoc
// @Summary Upload an image
// @Description Upload an image file for posts or profile
// @Tags media
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "Image file"
// @Success 200 {object} services.MediaUploadResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 413 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /media/upload/image [post]
func (h *MediaHandler) UploadImage(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Não autorizado",
			Message: "Token inválido",
		})
		return
	}

	// Obter arquivo do form
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Arquivo não encontrado",
			Message: "É necessário enviar um arquivo no campo 'file'",
		})
		return
	}

	// Upload do arquivo
	response, err := h.mediaService.UploadFile(file, userID.(uint), services.MediaTypeImage)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMsg := err.Error()

		switch {
		case strings.Contains(errorMsg, "muito grande"):
			statusCode = http.StatusRequestEntityTooLarge
		case strings.Contains(errorMsg, "não permitida"), strings.Contains(errorMsg, "não suportado"):
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   "Erro no upload da imagem",
			Message: errorMsg,
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Imagem enviada com sucesso",
		Data:    response,
	})
}

// UploadVideo godoc
// @Summary Upload a video
// @Description Upload a video file for posts
// @Tags media
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "Video file"
// @Success 200 {object} services.MediaUploadResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 413 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /media/upload/video [post]
func (h *MediaHandler) UploadVideo(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Não autorizado",
			Message: "Token inválido",
		})
		return
	}

	// Obter arquivo do form
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Arquivo não encontrado",
			Message: "É necessário enviar um arquivo no campo 'file'",
		})
		return
	}

	// Upload do arquivo
	response, err := h.mediaService.UploadFile(file, userID.(uint), services.MediaTypeVideo)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMsg := err.Error()

		switch {
		case strings.Contains(errorMsg, "muito grande"):
			statusCode = http.StatusRequestEntityTooLarge
		case strings.Contains(errorMsg, "não permitida"), strings.Contains(errorMsg, "não suportado"):
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   "Erro no upload do vídeo",
			Message: errorMsg,
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Vídeo enviado com sucesso",
		Data:    response,
	})
}

// UploadMultiple godoc
// @Summary Upload multiple files
// @Description Upload multiple image/video files at once
// @Tags media
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param files formData file true "Media files (multiple)"
// @Param type formData string false "Media type filter (image/video)" Enums(image, video)
// @Success 200 {object} MultipleUploadResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 413 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /media/upload/multiple [post]
func (h *MediaHandler) UploadMultiple(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Não autorizado",
			Message: "Token inválido",
		})
		return
	}

	// Obter form
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Erro no formulário",
			Message: err.Error(),
		})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Nenhum arquivo encontrado",
			Message: "É necessário enviar pelo menos um arquivo no campo 'files'",
		})
		return
	}

	// Limitar número de arquivos
	maxFiles := 10
	if len(files) > maxFiles {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Muitos arquivos",
			Message: fmt.Sprintf("Máximo de %d arquivos por vez", maxFiles),
		})
		return
	}

	// Tipo de mídia filtro (opcional)
	mediaTypeFilter := c.PostForm("type")

	var successUploads []services.MediaUploadResponse
	var failedUploads []FailedUpload

	for i, file := range files {
		// Determinar tipo de mídia baseado na extensão
		mediaType := h.determineMediaType(file.Filename)

		// Aplicar filtro se especificado
		if mediaTypeFilter != "" && string(mediaType) != mediaTypeFilter {
			failedUploads = append(failedUploads, FailedUpload{
				FileName: file.Filename,
				Error:    fmt.Sprintf("Tipo de arquivo não permitido para o filtro '%s'", mediaTypeFilter),
				Index:    i,
			})
			continue
		}

		// Tentar upload
		response, err := h.mediaService.UploadFile(file, userID.(uint), mediaType)
		if err != nil {
			failedUploads = append(failedUploads, FailedUpload{
				FileName: file.Filename,
				Error:    err.Error(),
				Index:    i,
			})
			continue
		}

		successUploads = append(successUploads, *response)
	}

	// Resposta
	result := MultipleUploadResponse{
		SuccessCount: len(successUploads),
		FailedCount:  len(failedUploads),
		TotalCount:   len(files),
		Successful:   successUploads,
		Failed:       failedUploads,
	}

	statusCode := http.StatusOK
	message := "Upload concluído"

	if len(failedUploads) > 0 && len(successUploads) == 0 {
		statusCode = http.StatusBadRequest
		message = "Todos os uploads falharam"
	} else if len(failedUploads) > 0 {
		statusCode = http.StatusPartialContent
		message = "Upload parcialmente concluído"
	}

	c.JSON(statusCode, SuccessResponse{
		Message: message,
		Data:    result,
	})
}

// DeleteMedia godoc
// @Summary Delete a media file
// @Description Delete an uploaded media file
// @Tags media
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body DeleteMediaRequest true "File path to delete"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /media/delete [delete]
func (h *MediaHandler) DeleteMedia(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Não autorizado",
			Message: "Token inválido",
		})
		return
	}

	var req DeleteMediaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Dados inválidos",
			Message: err.Error(),
		})
		return
	}

	if req.FilePath == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Caminho do arquivo obrigatório",
			Message: "O campo 'file_path' é obrigatório",
		})
		return
	}

	err := h.mediaService.DeleteFile(req.FilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Erro ao deletar arquivo",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Arquivo deletado com sucesso",
		Data:    nil,
	})
}

// GetMediaInfo godoc
// @Summary Get media file information
// @Description Get information about a media file
// @Tags media
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param file_path query string true "File path"
// @Success 200 {object} MediaInfoResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /media/info [get]
func (h *MediaHandler) GetMediaInfo(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Não autorizado",
			Message: "Token inválido",
		})
		return
	}

	filePath := c.Query("file_path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Parâmetro obrigatório",
			Message: "O parâmetro 'file_path' é obrigatório",
		})
		return
	}

	url := h.mediaService.GetFileURL(filePath)

	response := MediaInfoResponse{
		FilePath:  filePath,
		URL:       url,
		MediaType: h.determineMediaType(filePath),
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Informações do arquivo",
		Data:    response,
	})
}

// Funções auxiliares
func (h *MediaHandler) determineMediaType(filename string) services.MediaType {
	ext := strings.ToLower(filepath.Ext(filename))

	imageExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true,
		".gif": true, ".webp": true,
	}

	videoExts := map[string]bool{
		".mp4": true, ".avi": true, ".mov": true,
		".wmv": true, ".webm": true,
	}

	if imageExts[ext] {
		return services.MediaTypeImage
	} else if videoExts[ext] {
		return services.MediaTypeVideo
	}

	// Default para imagem se não conseguir determinar
	return services.MediaTypeImage
}

// Structs auxiliares
type MultipleUploadResponse struct {
	SuccessCount int                            `json:"success_count"`
	FailedCount  int                            `json:"failed_count"`
	TotalCount   int                            `json:"total_count"`
	Successful   []services.MediaUploadResponse `json:"successful"`
	Failed       []FailedUpload                 `json:"failed"`
}

type FailedUpload struct {
	FileName string `json:"file_name"`
	Error    string `json:"error"`
	Index    int    `json:"index"`
}

type DeleteMediaRequest struct {
	FilePath string `json:"file_path" binding:"required"`
}

type MediaInfoResponse struct {
	FilePath  string             `json:"file_path"`
	URL       string             `json:"url"`
	MediaType services.MediaType `json:"media_type"`
}
