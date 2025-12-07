package services

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/yugo-ibuki/claude-code-prompt-share/models"
)

type SessionService struct {
	claudeDir string
}

func NewSessionService() *SessionService {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return &SessionService{
		claudeDir: filepath.Join(homeDir, ".claude"),
	}
}

// GetAllProjects returns all Claude Code projects
func (s *SessionService) GetAllProjects() ([]models.Project, error) {
	projectsDir := filepath.Join(s.claudeDir, "projects")

	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read projects directory: %w", err)
	}

	var projects []models.Project
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		encodedPath := entry.Name()
		decodedPath := s.decodeProjectPath(encodedPath)

		sessions, err := s.getProjectSessions(encodedPath)
		if err != nil {
			continue
		}

		projects = append(projects, models.Project{
			EncodedPath: encodedPath,
			DecodedPath: decodedPath,
			Sessions:    sessions,
		})
	}

	return projects, nil
}

// GetSessionsByProject returns all sessions for a specific project
func (s *SessionService) getProjectSessions(encodedPath string) ([]models.SessionInfo, error) {
	projectDir := filepath.Join(s.claudeDir, "projects", encodedPath)

	entries, err := os.ReadDir(projectDir)
	if err != nil {
		return nil, err
	}

	var sessions []models.SessionInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}

		// Skip agent sessions
		if strings.HasPrefix(entry.Name(), "agent-") {
			continue
		}

		sessionID := strings.TrimSuffix(entry.Name(), ".jsonl")
		sessionInfo, err := s.getSessionInfo(encodedPath, sessionID)
		if err != nil {
			continue
		}

		sessions = append(sessions, sessionInfo)
	}

	// Sort by start time (newest first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].StartTime.After(sessions[j].StartTime)
	})

	return sessions, nil
}

// GetSessionInfo returns basic information about a session
func (s *SessionService) getSessionInfo(encodedPath, sessionID string) (models.SessionInfo, error) {
	session, err := s.GetSession(encodedPath, sessionID)
	if err != nil {
		return models.SessionInfo{}, err
	}

	firstMessage := ""
	if len(session.Messages) > 0 {
		firstMessage = session.Messages[0].Content
		if len(firstMessage) > 100 {
			firstMessage = firstMessage[:100] + "..."
		}
	}

	return models.SessionInfo{
		ID:          sessionID,
		ProjectPath: session.ProjectPath,
		ProjectName: session.ProjectName,
		StartTime:   session.StartTime,
		EndTime:     session.EndTime,
		MessageCount: len(session.Messages),
		FirstMessage: firstMessage,
	}, nil
}

// GetSession returns a complete session with all messages
func (s *SessionService) GetSession(encodedPath, sessionID string) (models.Session, error) {
	sessionFile := filepath.Join(s.claudeDir, "projects", encodedPath, sessionID+".jsonl")

	file, err := os.Open(sessionFile)
	if err != nil {
		return models.Session{}, fmt.Errorf("failed to open session file: %w", err)
	}
	defer file.Close()

	var messages []models.ConversationMessage
	var startTime, endTime time.Time

	scanner := bufio.NewScanner(file)
	// Increase buffer size for large messages
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		var jsonlMsg models.JSONLMessage
		if err := json.Unmarshal(scanner.Bytes(), &jsonlMsg); err != nil {
			continue
		}

		// Skip non-message types
		if jsonlMsg.Type != "user" && jsonlMsg.Type != "assistant" {
			continue
		}

		if jsonlMsg.Message == nil {
			continue
		}

		content := s.extractContent(jsonlMsg.Message.Content)

		msg := models.ConversationMessage{
			UUID:      jsonlMsg.UUID,
			Role:      jsonlMsg.Message.Role,
			Content:   content,
			Timestamp: jsonlMsg.Timestamp,
			IsAgent:   false,
		}

		messages = append(messages, msg)

		// Track start and end times
		if startTime.IsZero() || jsonlMsg.Timestamp.Before(startTime) {
			startTime = jsonlMsg.Timestamp
		}
		if jsonlMsg.Timestamp.After(endTime) {
			endTime = jsonlMsg.Timestamp
		}
	}

	if err := scanner.Err(); err != nil {
		return models.Session{}, fmt.Errorf("error reading session file: %w", err)
	}

	decodedPath := s.decodeProjectPath(encodedPath)
	projectName := filepath.Base(decodedPath)

	return models.Session{
		ID:          sessionID,
		ProjectPath: decodedPath,
		ProjectName: projectName,
		Messages:    messages,
		StartTime:   startTime,
		EndTime:     endTime,
	}, nil
}

// SearchSessions searches for sessions containing the query in messages
func (s *SessionService) SearchSessions(query string) ([]models.SessionInfo, error) {
	projects, err := s.GetAllProjects()
	if err != nil {
		return nil, err
	}

	var results []models.SessionInfo
	query = strings.ToLower(query)

	for _, project := range projects {
		for _, sessionInfo := range project.Sessions {
			session, err := s.GetSession(project.EncodedPath, sessionInfo.ID)
			if err != nil {
				continue
			}

			// Search in messages
			for _, msg := range session.Messages {
				if strings.Contains(strings.ToLower(msg.Content), query) {
					results = append(results, sessionInfo)
					break
				}
			}
		}
	}

	return results, nil
}

// extractContent extracts text content from various message content formats
func (s *SessionService) extractContent(content interface{}) string {
	switch v := content.(type) {
	case string:
		return v
	case []interface{}:
		var texts []string
		for _, item := range v {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if text, ok := itemMap["text"].(string); ok {
					texts = append(texts, text)
				}
			}
		}
		return strings.Join(texts, "\n")
	default:
		return fmt.Sprintf("%v", v)
	}
}

// decodeProjectPath converts encoded path back to original path
func (s *SessionService) decodeProjectPath(encoded string) string {
	// Replace dashes with slashes, handling special cases
	decoded := strings.ReplaceAll(encoded, "-", "/")

	// Handle leading slash
	if !strings.HasPrefix(decoded, "/") {
		decoded = "/" + decoded
	}

	return decoded
}

// GetProjectSessionsInfo returns session information for a specific project
func (s *SessionService) GetProjectSessionsInfo(encodedPath string) ([]models.SessionInfo, error) {
	return s.getProjectSessions(encodedPath)
}

// walkDir walks through directory and returns file entries
func walkDir(dirPath string) ([]fs.DirEntry, error) {
	return os.ReadDir(dirPath)
}
