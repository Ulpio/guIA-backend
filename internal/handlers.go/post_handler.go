package handlers

import (
	"net/http"
	"strconv"

	"github.com/Ulpio/guIA-backend/internal/services"
	"github.com/gin-gonic/gin"
)

type PostHandler struct {
	postService services.PostServiceInterface
}

func NewPostHandler(postService services.PostServiceInterface) *PostHandler {
	return &PostHandler{
		postService: postService,
	}
}

// CreatePost godoc
// @Summary Create a new post
// @Description Create a new post with text, images or videos
// @Tags posts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.CreatePostRequest true "Post creation data"
// @Success 201 {object} models.PostResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /posts [post]
func (h *PostHandler) CreatePost(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Não autorizado",
			Message: "Token inválido",
		})
		return
	}

	var req services.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Dados inválidos",
			Message: err.Error(),
		})
		return
	}

	post, err := h.postService.CreatePost(userID.(uint), &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMsg := err.Error()

		if contains(errorMsg, "obrigatório") || contains(errorMsg, "inválido") || contains(errorMsg, "deve ter") {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   "Erro ao criar post",
			Message: errorMsg,
		})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Message: "Post criado com sucesso",
		Data:    post,
	})
}

// GetFeed godoc
// @Summary Get user feed
// @Description Get the personalized feed for the authenticated user
// @Tags posts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Number of posts per page" default(20)
// @Param offset query int false "Number of posts to skip" default(0)
// @Success 200 {array} models.PostResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /posts [get]
func (h *PostHandler) GetFeed(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Não autorizado",
			Message: "Token inválido",
		})
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit <= 0 {
		limit = 20
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	posts, err := h.postService.GetFeed(userID.(uint), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Erro ao buscar feed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Feed encontrado",
		Data:    posts,
	})
}

// GetPostByID godoc
// @Summary Get post by ID
// @Description Get a specific post by its ID
// @Tags posts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Post ID"
// @Success 200 {object} models.PostResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /posts/{id} [get]
func (h *PostHandler) GetPostByID(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Não autorizado",
			Message: "Token inválido",
		})
		return
	}

	idParam := c.Param("id")
	postID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "ID inválido",
			Message: "O ID do post deve ser um número válido",
		})
		return
	}

	post, err := h.postService.GetPostByID(uint(postID), userID.(uint))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if contains(err.Error(), "não encontrado") {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   "Erro ao buscar post",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Post encontrado",
		Data:    post,
	})
}

// UpdatePost godoc
// @Summary Update a post
// @Description Update an existing post (only by the author)
// @Tags posts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Post ID"
// @Param request body services.UpdatePostRequest true "Post update data"
// @Success 200 {object} models.PostResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /posts/{id} [put]
func (h *PostHandler) UpdatePost(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Não autorizado",
			Message: "Token inválido",
		})
		return
	}

	idParam := c.Param("id")
	postID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "ID inválido",
			Message: "O ID do post deve ser um número válido",
		})
		return
	}

	var req services.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Dados inválidos",
			Message: err.Error(),
		})
		return
	}

	post, err := h.postService.UpdatePost(uint(postID), userID.(uint), &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMsg := err.Error()

		switch {
		case contains(errorMsg, "não encontrado"):
			statusCode = http.StatusNotFound
		case contains(errorMsg, "não tem permissão"):
			statusCode = http.StatusForbidden
		case contains(errorMsg, "inválido"), contains(errorMsg, "deve ter"):
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   "Erro ao atualizar post",
			Message: errorMsg,
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Post atualizado com sucesso",
		Data:    post,
	})
}

// DeletePost godoc
// @Summary Delete a post
// @Description Delete an existing post (only by the author)
// @Tags posts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Post ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /posts/{id} [delete]
func (h *PostHandler) DeletePost(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Não autorizado",
			Message: "Token inválido",
		})
		return
	}

	idParam := c.Param("id")
	postID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "ID inválido",
			Message: "O ID do post deve ser um número válido",
		})
		return
	}

	err = h.postService.DeletePost(uint(postID), userID.(uint))
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMsg := err.Error()

		switch {
		case contains(errorMsg, "não encontrado"):
			statusCode = http.StatusNotFound
		case contains(errorMsg, "não tem permissão"):
			statusCode = http.StatusForbidden
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   "Erro ao deletar post",
			Message: errorMsg,
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Post deletado com sucesso",
		Data:    nil,
	})
}

// LikePost godoc
// @Summary Like a post
// @Description Like a specific post
// @Tags posts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Post ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /posts/{id}/like [post]
func (h *PostHandler) LikePost(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Não autorizado",
			Message: "Token inválido",
		})
		return
	}

	idParam := c.Param("id")
	postID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "ID inválido",
			Message: "O ID do post deve ser um número válido",
		})
		return
	}

	err = h.postService.LikePost(userID.(uint), uint(postID))
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMsg := err.Error()

		switch {
		case contains(errorMsg, "não encontrado"):
			statusCode = http.StatusNotFound
		case contains(errorMsg, "já curtiu"):
			statusCode = http.StatusConflict
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   "Erro ao curtir post",
			Message: errorMsg,
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Post curtido com sucesso",
		Data:    nil,
	})
}

// UnlikePost godoc
// @Summary Unlike a post
// @Description Remove like from a specific post
// @Tags posts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Post ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /posts/{id}/like [delete]
func (h *PostHandler) UnlikePost(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Não autorizado",
			Message: "Token inválido",
		})
		return
	}

	idParam := c.Param("id")
	postID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "ID inválido",
			Message: "O ID do post deve ser um número válido",
		})
		return
	}

	err = h.postService.UnlikePost(userID.(uint), uint(postID))
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMsg := err.Error()

		switch {
		case contains(errorMsg, "não encontrado"):
			statusCode = http.StatusNotFound
		case contains(errorMsg, "não curtiu"):
			statusCode = http.StatusConflict
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   "Erro ao descurtir post",
			Message: errorMsg,
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Like removido com sucesso",
		Data:    nil,
	})
}

// GetPostsByAuthor godoc
// @Summary Get posts by author
// @Description Get all posts from a specific author
// @Tags posts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param authorId query int true "Author ID"
// @Param limit query int false "Number of posts per page" default(20)
// @Param offset query int false "Number of posts to skip" default(0)
// @Success 200 {array} models.PostResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /posts/author [get]
func (h *PostHandler) GetPostsByAuthor(c *gin.Context) {
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Não autorizado",
			Message: "Token inválido",
		})
		return
	}

	authorIDParam := c.Query("authorId")
	if authorIDParam == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Parâmetro obrigatório",
			Message: "O parâmetro 'authorId' é obrigatório",
		})
		return
	}

	authorID, err := strconv.ParseUint(authorIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "ID inválido",
			Message: "O ID do autor deve ser um número válido",
		})
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit <= 0 {
		limit = 20
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	posts, err := h.postService.GetPostsByAuthor(uint(authorID), currentUserID.(uint), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Erro ao buscar posts do autor",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Posts encontrados",
		Data:    posts,
	})
}

// SearchPosts godoc
// @Summary Search posts
// @Description Search for posts by content or location
// @Tags posts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param q query string true "Search query"
// @Param limit query int false "Number of results per page" default(20)
// @Param offset query int false "Number of results to skip" default(0)
// @Success 200 {array} models.PostResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /posts/search [get]
func (h *PostHandler) SearchPosts(c *gin.Context) {
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Não autorizado",
			Message: "Token inválido",
		})
		return
	}

	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Parâmetro obrigatório",
			Message: "O parâmetro 'q' (query) é obrigatório",
		})
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit <= 0 {
		limit = 20
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	posts, err := h.postService.SearchPosts(query, currentUserID.(uint), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Erro na busca de posts",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Busca realizada com sucesso",
		Data:    posts,
	})
}

// GetTrendingPosts godoc
// @Summary Get trending posts
// @Description Get posts that are currently trending
// @Tags posts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Number of posts per page" default(20)
// @Param offset query int false "Number of posts to skip" default(0)
// @Success 200 {array} models.PostResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /posts/trending [get]
func (h *PostHandler) GetTrendingPosts(c *gin.Context) {
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Não autorizado",
			Message: "Token inválido",
		})
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit <= 0 {
		limit = 20
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	posts, err := h.postService.GetTrendingPosts(currentUserID.(uint), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Erro ao buscar posts em alta",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Posts em alta encontrados",
		Data:    posts,
	})
}
