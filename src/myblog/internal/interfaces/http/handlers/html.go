// myblog/internal/interfaces/http/handler/html.go
package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/lonmouth/myblog/internal/domain/entities"
	"github.com/lonmouth/myblog/internal/interfaces/http/middleware"
	"go.uber.org/zap"
)

func (h *PostHandler) HandleHTMLGetPosts(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	posts, pagination, err := h.postUseCase.GetPosts(page)
	if err != nil {
		h.logger.Error("Failed to retrieve posts", zap.Error(err))
		http.Error(w, "Failed to retrieve posts", http.StatusInternalServerError)
		return
	}

	data := TemplateData{
		Title:         "Главная страница блога", // Устанавливаем заголовок
		Posts:         posts,
		Pagination:    pagination,
		Authenticated: false,
		Username:      "",
		Role:          "",
		CSRFToken:     middleware.CSRFTokenFromContext(r.Context()),
		CurrentPage:   page, // Передаем текущую страницу в шаблон
	}

	if token, ok := r.Context().Value("user").(*jwt.Token); ok {
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			data.Authenticated = true
			data.Username = claims["username"].(string)
			data.Role = claims["role"].(string) // Добавляем получение роли
		}
	}

	tmpl := template.Must(template.New("base.html").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
	}).ParseFiles(
		"templates/base.html",
		"templates/index.html",
	))

	if err := tmpl.Execute(w, data); err != nil {
		h.logger.Error("Failed to execute template", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *PostHandler) HandleHTMLGetPost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		h.logger.Error("Failed conversion currentPage in int", zap.Error(err))
	}
	post, err := h.postUseCase.GetPostByID(id)
	if err != nil {
		h.logger.Error("Failed to get post", zap.Error(err))
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// Получаем текущую страницу из параметров URL
	currentPage, err := strconv.Atoi(r.URL.Query().Get("fromPage"))
	if err != nil {
		h.logger.Error("Failed conversion currentPage in int", zap.Error(err))
	}
	if currentPage == 0 {
		currentPage = 1 // Значение по умолчанию
	}

	// Создаем структуру с преобразованным HTML
	data := TemplateData{
		Post:          post,
		ContentHTML:   template.HTML(post.HTMLContent),
		CurrentPage:   currentPage,
		Authenticated: false,
		Role:          "",
		CSRFToken:     middleware.CSRFTokenFromContext(r.Context()),
	}

	if token, ok := r.Context().Value("user").(*jwt.Token); ok {
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			data.Authenticated = true
			data.Username = claims["username"].(string)
			data.Role = claims["role"].(string) // Добавляем получение роли
		}
	}

	// Регистрируем функцию safeHTML
	tmpl := template.Must(template.New("post.html").Funcs(template.FuncMap{
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
	}).ParseFiles(
		"templates/base.html",
		PathToPost,
	))

	if err := tmpl.Execute(w, data); err != nil {
		h.logger.Error("Failed to execute template", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *PostHandler) HandleAdminCreatePostForm(w http.ResponseWriter, r *http.Request) {
	// currentPage := r.URL.Query().Get("fromPage")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	currentPage, err := strconv.Atoi(r.URL.Query().Get("fromPage"))
	if err != nil {
		h.logger.Error("Failed conversion currentPage in int", zap.Error(err))
	}
	if currentPage == 0 {
		currentPage = 1 // Значение по умолчанию
	}

	data := TemplateData{
		CurrentPage:   currentPage,
		Authenticated: false,
		Role:          "",
		CSRFToken:     middleware.CSRFTokenFromContext(r.Context()),
	}

	if token, ok := r.Context().Value("user").(*jwt.Token); ok {
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			data.Authenticated = true
			data.Username = claims["username"].(string)
			data.Role = claims["role"].(string) // Добавляем получение роли
		}
	}

	tmpl := template.Must(template.ParseFiles("templates/admin_create.html"))
	tmpl.Execute(w, data)
}

func (h *PostHandler) HandleAdminCreatePost(w http.ResponseWriter, r *http.Request) {
	// Логирование входящего запроса
	h.logger.Info("CreatePost request",
		zap.String("method", r.Method),
		zap.Any("form", r.Form),
	)

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		h.logger.Error("Failed to parse form", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")
	contentDescr := r.FormValue("content_description")

	h.logger.Info("Form data",
		zap.String("title", title),
		zap.String("content", content),
		zap.String("content_description", contentDescr),
	)

	postID, currentPage, err := h.postUseCase.CreatePost(&entities.Post{
		Title:        title,
		Content:      content,
		ContentDescr: contentDescr,
		Author:       "admin",
	})
	if err != nil {
		h.logger.Error("Post creation failed", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if currentPage == 0 {
		currentPage = 1 // Значение по умолчанию
	}

	// Перенаправляем на страницу созданного поста с параметром fromPage
	http.Redirect(w, r, fmt.Sprintf("/posts/%d?fromPage=%d", postID, currentPage), http.StatusSeeOther)
}
