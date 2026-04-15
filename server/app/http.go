package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const sessionCookieName = "treachery_session"

func (a *App) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/session", a.handleCreateSession)
	mux.HandleFunc("GET /api/healthz", a.handleHealthz)
	mux.HandleFunc("GET /api/games", a.handleListGames)
	mux.HandleFunc("POST /api/games", a.handleCreateGame)
	mux.HandleFunc("GET /api/games/{gameId}/snapshot", a.handleGameSnapshot)
	mux.HandleFunc("GET /api/games/{gameId}/events", a.handleGameEvents)
	mux.HandleFunc("POST /api/games/{gameId}/join", a.handleJoinGame)
	mux.HandleFunc("POST /api/games/{gameId}/start", a.handleStartGame)
	mux.HandleFunc("POST /api/games/{gameId}/murderer-selection", a.handleSelectMurdererCards)
	mux.HandleFunc("POST /api/games/{gameId}/forensic/cause", a.handleSelectCauseCard)
	mux.HandleFunc("POST /api/games/{gameId}/forensic/location", a.handleSelectLocationCard)
	mux.HandleFunc("POST /api/games/{gameId}/forensic/other", a.handleSelectOtherCard)
	mux.HandleFunc("POST /api/games/{gameId}/guess", a.handleMakeGuess)
	mux.HandleFunc("POST /api/games/{gameId}/messages", a.handleSendMessage)
	mux.HandleFunc("POST /api/games/{gameId}/end", a.handleEndGame)
	mux.HandleFunc("/", a.handleSPA)
	return withCORSAndLogging(mux)
}

func withCORSAndLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func (a *App) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	var existingToken string
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		existingToken = cookie.Value
	}
	session, err := a.EnsureSession(existingToken)
	if err != nil {
		writeError(w, err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    session.Token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   60 * 60 * 24 * 365,
	})
	writeJSON(w, http.StatusOK, session)
}

func (a *App) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *App) handleListGames(w http.ResponseWriter, r *http.Request) {
	games, err := a.ListGames()
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, games)
}

func (a *App) handleCreateGame(w http.ResponseWriter, r *http.Request) {
	session, err := a.requireSession(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var body struct {
		GameID string `json:"gameId"`
	}
	if err := decodeBody(r, &body); err != nil {
		writeError(w, fmt.Errorf("%w: %s", ErrBadInput, err.Error()))
		return
	}
	if err := a.CreateGame(session.UID, body.GameID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "gameId": strings.ToUpper(strings.TrimSpace(body.GameID))})
}

func (a *App) handleGameSnapshot(w http.ResponseWriter, r *http.Request) {
	session, err := a.requireSession(r)
	if err != nil {
		writeError(w, err)
		return
	}
	snapshot, err := a.GetSnapshot(r.PathValue("gameId"), session.UID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, snapshot)
}

func (a *App) handleGameEvents(w http.ResponseWriter, r *http.Request) {
	if _, err := a.requireSession(r); err != nil {
		writeError(w, err)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, fmt.Errorf("streaming unsupported"))
		return
	}

	gameID := strings.ToUpper(strings.TrimSpace(r.PathValue("gameId")))
	updates, cancel := a.hub.Subscribe(gameID)
	defer cancel()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	_, _ = io.WriteString(w, "data: ready\n\n")
	flusher.Flush()

	pingTicker := time.NewTicker(25 * time.Second)
	defer pingTicker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-pingTicker.C:
			_, _ = io.WriteString(w, ": ping\n\n")
			flusher.Flush()
		case <-updates:
			_, _ = io.WriteString(w, "event: update\ndata: reload\n\n")
			flusher.Flush()
		}
	}
}

func (a *App) handleJoinGame(w http.ResponseWriter, r *http.Request) {
	session, err := a.requireSession(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var body struct {
		PlayerName string `json:"playerName"`
	}
	if err := decodeBody(r, &body); err != nil {
		writeError(w, fmt.Errorf("%w: %s", ErrBadInput, err.Error()))
		return
	}
	gameID := r.PathValue("gameId")
	if err := a.AddPlayer(gameID, session.UID, body.PlayerName); err != nil {
		writeError(w, err)
		return
	}
	a.hub.Publish(strings.ToUpper(strings.TrimSpace(gameID)))
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

func (a *App) handleStartGame(w http.ResponseWriter, r *http.Request) {
	a.withGameMutation(w, r, func(session *Session, gameID string) error {
		return a.StartGame(gameID, session.UID)
	})
}

func (a *App) handleSelectMurdererCards(w http.ResponseWriter, r *http.Request) {
	a.withGameMutation(w, r, func(session *Session, gameID string) error {
		var body struct {
			ClueCardName  string `json:"clueCardName"`
			MeansCardName string `json:"meansCardName"`
		}
		if err := decodeBody(r, &body); err != nil {
			return fmt.Errorf("%w: %s", ErrBadInput, err.Error())
		}
		return a.SelectMurdererCards(gameID, session.UID, body.ClueCardName, body.MeansCardName)
	})
}

func (a *App) handleSelectCauseCard(w http.ResponseWriter, r *http.Request) {
	a.withGameMutation(w, r, func(session *Session, gameID string) error {
		var body struct {
			Card ForensicCard `json:"card"`
		}
		if err := decodeBody(r, &body); err != nil {
			return fmt.Errorf("%w: %s", ErrBadInput, err.Error())
		}
		return a.SelectForensicCauseCard(gameID, session.UID, body.Card)
	})
}

func (a *App) handleSelectLocationCard(w http.ResponseWriter, r *http.Request) {
	a.withGameMutation(w, r, func(session *Session, gameID string) error {
		var body struct {
			Card ForensicCard `json:"card"`
		}
		if err := decodeBody(r, &body); err != nil {
			return fmt.Errorf("%w: %s", ErrBadInput, err.Error())
		}
		return a.SelectForensicLocationCard(gameID, session.UID, body.Card)
	})
}

func (a *App) handleSelectOtherCard(w http.ResponseWriter, r *http.Request) {
	a.withGameMutation(w, r, func(session *Session, gameID string) error {
		var body struct {
			Card            ForensicCard `json:"card"`
			ReplaceCardName string       `json:"replaceCardName"`
		}
		if err := decodeBody(r, &body); err != nil {
			return fmt.Errorf("%w: %s", ErrBadInput, err.Error())
		}
		return a.SelectForensicOtherCard(gameID, session.UID, body.Card, body.ReplaceCardName)
	})
}

func (a *App) handleMakeGuess(w http.ResponseWriter, r *http.Request) {
	a.withGameMutation(w, r, func(session *Session, gameID string) error {
		var body struct {
			MurdererUID   string `json:"murdererUid"`
			ClueCardName  string `json:"clueCardName"`
			MeansCardName string `json:"meansCardName"`
		}
		if err := decodeBody(r, &body); err != nil {
			return fmt.Errorf("%w: %s", ErrBadInput, err.Error())
		}
		return a.MakeGuess(gameID, session.UID, body.MurdererUID, body.ClueCardName, body.MeansCardName)
	})
}

func (a *App) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	a.withGameMutation(w, r, func(session *Session, gameID string) error {
		var body struct {
			Message string `json:"message"`
		}
		if err := decodeBody(r, &body); err != nil {
			return fmt.Errorf("%w: %s", ErrBadInput, err.Error())
		}
		return a.SendChatMessage(gameID, session.UID, body.Message)
	})
}

func (a *App) handleEndGame(w http.ResponseWriter, r *http.Request) {
	a.withGameMutation(w, r, func(session *Session, gameID string) error {
		return a.EndGame(gameID, session.UID)
	})
}

func (a *App) withGameMutation(w http.ResponseWriter, r *http.Request, fn func(*Session, string) error) {
	session, err := a.requireSession(r)
	if err != nil {
		writeError(w, err)
		return
	}
	gameID := strings.ToUpper(strings.TrimSpace(r.PathValue("gameId")))
	if err := fn(session, gameID); err != nil {
		writeError(w, err)
		return
	}
	a.hub.Publish(gameID)
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

func (a *App) handleSPA(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		http.NotFound(w, r)
		return
	}
	if a.distDir == "" {
		http.Error(w, "missing dist dir", http.StatusServiceUnavailable)
		return
	}

	cleanPath := filepath.Clean(strings.TrimPrefix(r.URL.Path, "/"))
	if cleanPath == "." {
		cleanPath = "index.html"
	}
	candidate := filepath.Join(a.distDir, cleanPath)
	if rel, err := filepath.Rel(a.distDir, candidate); err == nil && !strings.HasPrefix(rel, "..") {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			http.ServeFile(w, r, candidate)
			return
		}
	}
	http.ServeFile(w, r, filepath.Join(a.distDir, "index.html"))
}

func (a *App) requireSession(r *http.Request) (*Session, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return nil, fmt.Errorf("%w: missing session", ErrForbidden)
	}
	return a.GetSession(cookie.Value)
}

func decodeBody(r *http.Request, target any) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, ErrForbidden):
		status = http.StatusForbidden
	case errors.Is(err, ErrBadInput):
		status = http.StatusBadRequest
	}
	writeJSON(w, status, ErrorResponse{Error: err.Error()})
}
