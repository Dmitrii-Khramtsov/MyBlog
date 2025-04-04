// myblog/internal/interfaces/http/handler/api.go
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"go.uber.org/zap"

	"github.com/lonmouth/myblog/internal/application/usecases"
	"github.com/lonmouth/myblog/internal/domain/entities"
	"github.com/lonmouth/myblog/internal/infrastructure/logger"
)

const (
	PathToIndex = "templates/index.html"
	PathToAdmin = "templates/admin.html"
	PathToPost  = "templates/post.html"
)

type PostHandler struct {
	postUseCase usecases.PostUseCase
	logger      *logger.AppLogger
}

func NewPostHandler(postUseCase usecases.PostUseCase, logger *logger.AppLogger) *PostHandler {
	return &PostHandler{
		postUseCase: postUseCase,
		logger:      logger.WithModule("post_handler"),
	}
}

// API Endpoints
func (h *PostHandler) HandleAPICreatePost(w http.ResponseWriter, r *http.Request) {
	log := h.logger.With(zap.String("handler", "HandleAPICreatePost"))

	var post entities.Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		log.Error("Failed to decode request", zap.Error(err))
		errorResponse(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if _, _, err := h.postUseCase.CreatePost(&post); err != nil {
		log.Error("Failed to create post", zap.Error(err))
		errorResponse(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}

func (h *PostHandler) HandleAPIGetPosts(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	posts, _, err := h.postUseCase.GetPosts(page)
	if err != nil {
		h.logger.Error("Failed to get posts", zap.Error(err))
		errorResponse(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func (h *PostHandler) getPostIDParameter(r *http.Request) (int, error) {
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		return 0, fmt.Errorf("missing post ID")
	}

	return strconv.Atoi(idParam)
}

func (h *PostHandler) renderJSONResponse(w http.ResponseWriter, data interface{}) {
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON", zap.Error(err))
		errorResponse(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *PostHandler) HandleAPIGetPostByID(w http.ResponseWriter, r *http.Request) {
	log := h.logger.With(zap.String("handler", "HandleGetPostByID"))
	w.Header().Set("Content-Type", "application/json")

	id, err := h.getPostIDParameter(r)
	if err != nil {
		log.Warn("Invalid post ID", zap.Error(err))
		errorResponse(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	post, err := h.postUseCase.GetPostByID(id)
	if err != nil {
		log.Error("Failed to retrieve post by ID", zap.Error(err), zap.Int("id", id))
		errorResponse(w, "Failed to retrieve post", http.StatusInternalServerError)
		return
	}

	h.renderJSONResponse(w, post)
}

func errorResponse(w http.ResponseWriter, message string, httpStatusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusCode)
	resp := InlineResponse400{
		Error_: message,
	}

	json.NewEncoder(w).Encode(resp)
}
