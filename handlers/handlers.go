package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/yugo-ibuki/claude-code-prompt-share/services"
)

type Handler struct {
	sessionService *services.SessionService
}

func NewHandler() *Handler {
	return &Handler{
		sessionService: services.NewSessionService(),
	}
}

// IndexHandler shows the main 3-column layout
func (h *Handler) IndexHandler(c echo.Context) error {
	projects, err := h.sessionService.GetAllProjects()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load projects: "+err.Error())
	}

	return c.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Projects": projects,
	})
}

// GetProjectsAPIHandler returns all projects as JSON
func (h *Handler) GetProjectsAPIHandler(c echo.Context) error {
	projects, err := h.sessionService.GetAllProjects()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, projects)
}

// GetSessionsAPIHandler returns all sessions for a project as JSON
func (h *Handler) GetSessionsAPIHandler(c echo.Context) error {
	encodedPath := c.Param("encodedPath")
	sessions, err := h.sessionService.GetProjectSessionsInfo(encodedPath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, sessions)
}

// GetPromptsAPIHandler returns conversation threads (grouped prompts) for a session as JSON
func (h *Handler) GetPromptsAPIHandler(c echo.Context) error {
	encodedPath := c.Param("encodedPath")
	sessionID := c.Param("sessionId")

	session, err := h.sessionService.GetSession(encodedPath, sessionID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Group messages into conversation threads
	var threads []map[string]interface{}
	var currentThread map[string]interface{}
	var threadPrompts []map[string]interface{}

	for i, msg := range session.Messages {
		if msg.Role == "user" {
			// Skip empty or whitespace-only prompts
			content := strings.TrimSpace(msg.Content)
			if content == "" {
				continue
			}

			// Add prompt to current thread
			threadPrompts = append(threadPrompts, map[string]interface{}{
				"index":     i,
				"uuid":      msg.UUID,
				"content":   content,
				"timestamp": msg.Timestamp,
			})
		} else if msg.Role == "assistant" && len(threadPrompts) > 0 {
			// Create thread when we hit an assistant response
			firstPrompt := threadPrompts[0]
			lastPrompt := threadPrompts[len(threadPrompts)-1]

			// Create thread summary
			summary := firstPrompt["content"].(string)
			if len(summary) > 80 {
				summary = summary[:80] + "..."
			}

			currentThread = map[string]interface{}{
				"id":           fmt.Sprintf("thread-%d", len(threads)),
				"firstIndex":   firstPrompt["index"],
				"promptCount":  len(threadPrompts),
				"prompts":      threadPrompts,
				"summary":      summary,
				"startTime":    firstPrompt["timestamp"],
				"endTime":      lastPrompt["timestamp"],
			}

			threads = append(threads, currentThread)
			threadPrompts = []map[string]interface{}{}
		}
	}

	// Add remaining prompts as a thread if any
	if len(threadPrompts) > 0 {
		firstPrompt := threadPrompts[0]
		lastPrompt := threadPrompts[len(threadPrompts)-1]

		summary := firstPrompt["content"].(string)
		if len(summary) > 80 {
			summary = summary[:80] + "..."
		}

		currentThread = map[string]interface{}{
			"id":          fmt.Sprintf("thread-%d", len(threads)),
			"firstIndex":  firstPrompt["index"],
			"promptCount": len(threadPrompts),
			"prompts":     threadPrompts,
			"summary":     summary,
			"startTime":   firstPrompt["timestamp"],
			"endTime":     lastPrompt["timestamp"],
		}

		threads = append(threads, currentThread)
	}

	// Reverse to show newest first
	for i, j := 0, len(threads)-1; i < j; i, j = i+1, j-1 {
		threads[i], threads[j] = threads[j], threads[i]
	}

	return c.JSON(http.StatusOK, threads)
}

// GetResponseAPIHandler returns the assistant response for a specific prompt
func (h *Handler) GetResponseAPIHandler(c echo.Context) error {
	encodedPath := c.Param("encodedPath")
	sessionID := c.Param("sessionId")
	promptIndex := c.Param("promptIndex")

	session, err := h.sessionService.GetSession(encodedPath, sessionID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Find the prompt and its response
	var promptMsg, responseMsg *map[string]interface{}

	for i, msg := range session.Messages {
		if msg.Role == "user" {
			// Convert index to string and compare
			currentIdx := 0
			for j := 0; j <= i; j++ {
				if session.Messages[j].Role == "user" {
					currentIdx++
				}
			}

			if promptIndex == fmt.Sprintf("%d", i) {
				promptMsg = &map[string]interface{}{
					"uuid":      msg.UUID,
					"content":   msg.Content,
					"timestamp": msg.Timestamp,
				}

				// Find the next assistant message
				for j := i + 1; j < len(session.Messages); j++ {
					if session.Messages[j].Role == "assistant" {
						responseMsg = &map[string]interface{}{
							"uuid":      session.Messages[j].UUID,
							"content":   session.Messages[j].Content,
							"timestamp": session.Messages[j].Timestamp,
						}
						break
					}
				}
				break
			}
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"prompt":   promptMsg,
		"response": responseMsg,
	})
}

// GetSessionFullAPIHandler returns the complete session history (chat view)
func (h *Handler) GetSessionFullAPIHandler(c echo.Context) error {
	encodedPath := c.Param("encodedPath")
	sessionID := c.Param("sessionId")

	session, err := h.sessionService.GetSession(encodedPath, sessionID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Transform messages for frontend
	var chatMessages []map[string]interface{}
	for i, msg := range session.Messages {
		content := strings.TrimSpace(msg.Content)
		if content == "" {
			continue
		}

		chatMessages = append(chatMessages, map[string]interface{}{
			"index":     i,
			"uuid":      msg.UUID,
			"role":      msg.Role, 
			"content":   msg.Content,
			"timestamp": msg.Timestamp,
		})
	}

	return c.JSON(http.StatusOK, chatMessages)
}

// SearchHandler handles search requests
func (h *Handler) SearchHandler(c echo.Context) error {
	query := c.QueryParam("q")

	if query == "" {
		return c.Redirect(http.StatusFound, "/")
	}

	sessions, err := h.sessionService.SearchSessions(query)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Search failed: "+err.Error())
	}

	return c.Render(http.StatusOK, "search.html", map[string]interface{}{
		"Query":    query,
		"Sessions": sessions,
	})
}

func getProjectName(path string) string {
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
}
