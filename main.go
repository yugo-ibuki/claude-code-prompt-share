package main

import (
	"html/template"
	"io"
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/yugo-ibuki/claude-code-prompt-share/handlers"
)

type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Template helper functions
	funcMap := template.FuncMap{
		"getProjectName": func(path string) string {
			if path == "" {
				return "Unknown"
			}
			parts := []rune(path)
			for i := len(parts) - 1; i >= 0; i-- {
				if parts[i] == '/' {
					return string(parts[i+1:])
				}
			}
			return path
		},
	}

	// Template renderer
	renderer := &TemplateRenderer{
		templates: template.Must(template.New("").Funcs(funcMap).ParseGlob("templates/*.html")),
	}
	e.Renderer = renderer

	// Static files
	e.Static("/static", "static")

	// Initialize handlers
	h := handlers.NewHandler()

	// Routes
	e.GET("/", h.IndexHandler)
	e.GET("/search", h.SearchHandler)

	// API Routes
	e.GET("/api/projects", h.GetProjectsAPIHandler)
	e.GET("/api/projects/:encodedPath/sessions", h.GetSessionsAPIHandler)
	e.GET("/api/projects/:encodedPath/sessions/:sessionId/prompts", h.GetPromptsAPIHandler)
	e.GET("/api/projects/:encodedPath/sessions/:sessionId/prompts/:promptIndex", h.GetResponseAPIHandler)
	e.GET("/api/projects/:encodedPath/sessions/:sessionId/full", h.GetSessionFullAPIHandler)

	// Start server
	log.Println("Starting Claude Code Session Viewer on http://localhost:8080")
	e.Logger.Fatal(e.Start(":8080"))
}
