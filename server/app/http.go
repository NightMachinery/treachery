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
	mux.HandleFunc("PUT /api/me", a.handleUpdateProfile)
	mux.HandleFunc("GET /api/healthz", a.handleHealthz)
	mux.HandleFunc("GET /api/games", a.handleListGames)
	mux.HandleFunc("POST /api/games", a.handleCreateGame)
	mux.HandleFunc("GET /api/games/{gameId}/snapshot", a.handleGameSnapshot)
	mux.HandleFunc("GET /api/games/{gameId}/events", a.handleGameEvents)
	mux.HandleFunc("POST /api/games/{gameId}/join", a.handleJoinGame)
	mux.HandleFunc("POST /api/games/{gameId}/participants/{uid}/role", a.handleSetParticipantRole)
	mux.HandleFunc("POST /api/games/{gameId}/scientist/toggle", a.handleToggleScientist)
	mux.HandleFunc("POST /api/games/{gameId}/settings", a.handleUpdateGameSettings)
	mux.HandleFunc("POST /api/games/{gameId}/room-mods", a.handleUpdateRoomMods)
	mux.HandleFunc("POST /api/games/{gameId}/migrate-device", a.handleCreateRoomAuth)
	mux.HandleFunc("POST /api/games/{gameId}/start", a.handleStartGame)
	mux.HandleFunc("POST /api/games/{gameId}/room-timer/start", a.handleStartRoomTimer)
	mux.HandleFunc("POST /api/games/{gameId}/room-timer/pause", a.handlePauseRoomTimer)
	mux.HandleFunc("POST /api/games/{gameId}/room-timer/resume", a.handleResumeRoomTimer)
	mux.HandleFunc("POST /api/games/{gameId}/room-timer/reset", a.handleResetRoomTimer)
	mux.HandleFunc("POST /api/games/{gameId}/room-timer/clear", a.handleClearRoomTimer)
	mux.HandleFunc("POST /api/games/{gameId}/murderer-selection", a.handleSelectMurdererCards)
	mux.HandleFunc("POST /api/games/{gameId}/witness-selection/show", a.handleShowWitnessSelectionPrompt)
	mux.HandleFunc("POST /api/games/{gameId}/witness-selection", a.handleSubmitWitnessSelection)
	mux.HandleFunc("POST /api/games/{gameId}/witness-selection/dismiss", a.handleDismissWitnessSelectionPrompt)
	mux.HandleFunc("POST /api/games/{gameId}/forensic/cause", a.handleSelectCauseCard)
	mux.HandleFunc("POST /api/games/{gameId}/forensic/location", a.handleSelectLocationCard)
	mux.HandleFunc("POST /api/games/{gameId}/forensic/other", a.handleSelectOtherCard)
	mux.HandleFunc("POST /api/games/{gameId}/guess", a.handleMakeGuess)
	mux.HandleFunc("POST /api/games/{gameId}/messages", a.handleSendMessage)
	mux.HandleFunc("POST /api/games/{gameId}/end", a.handleEndGame)
	mux.HandleFunc("POST /api/games/{gameId}/restart", a.handleRestartGame)
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
	existingToken := sessionTokenFromRequest(r)
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

func (a *App) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	session, err := a.requireSession(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var body struct {
		DisplayName string `json:"displayName"`
		GameID      string `json:"gameId"`
	}
	if err := decodeBody(r, &body); err != nil {
		writeError(w, fmt.Errorf("%w: %s", ErrBadInput, err.Error()))
		return
	}
	effectiveUID, err := a.resolveEffectiveUID(r, strings.ToUpper(strings.TrimSpace(body.GameID)), session.UID)
	if err != nil {
		writeError(w, err)
		return
	}
	if err := a.UpdateProfile(effectiveUID, body.DisplayName); err != nil {
		writeError(w, err)
		return
	}
	updatedSession, err := a.GetSession(session.Token)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updatedSession)
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
	gameID := strings.ToUpper(strings.TrimSpace(r.PathValue("gameId")))
	viewerUID, err := a.resolveEffectiveUID(r, gameID, session.UID)
	if err != nil {
		writeError(w, err)
		return
	}
	snapshot, err := a.GetSnapshot(gameID, viewerUID)
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
		Role ParticipantRole `json:"role"`
	}
	if err := decodeBody(r, &body); err != nil {
		writeError(w, fmt.Errorf("%w: %s", ErrBadInput, err.Error()))
		return
	}
	gameID := strings.ToUpper(strings.TrimSpace(r.PathValue("gameId")))
	viewerUID, err := a.resolveEffectiveUID(r, gameID, session.UID)
	if err != nil {
		writeError(w, err)
		return
	}
	if err := a.UpsertParticipant(gameID, viewerUID, body.Role); err != nil {
		writeError(w, err)
		return
	}
	a.hub.Publish(gameID)
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

func (a *App) handleSetParticipantRole(w http.ResponseWriter, r *http.Request) {
	a.withGameMutation(w, r, func(viewerUID, gameID string) error {
		var body struct {
			Role ParticipantRole `json:"role"`
		}
		if err := decodeBody(r, &body); err != nil {
			return fmt.Errorf("%w: %s", ErrBadInput, err.Error())
		}
		return a.SetParticipantRole(gameID, viewerUID, r.PathValue("uid"), body.Role)
	})
}

func (a *App) handleToggleScientist(w http.ResponseWriter, r *http.Request) {
	a.withGameMutation(w, r, func(viewerUID, gameID string) error {
		var body struct {
			UID string `json:"uid"`
		}
		if err := decodeBody(r, &body); err != nil {
			return fmt.Errorf("%w: %s", ErrBadInput, err.Error())
		}
		return a.ToggleScientistMark(gameID, viewerUID, body.UID)
	})
}

func (a *App) handleUpdateGameSettings(w http.ResponseWriter, r *http.Request) {
	a.withGameMutation(w, r, func(viewerUID, gameID string) error {
		var body GameSettingsInput
		if err := decodeBody(r, &body); err != nil {
			return fmt.Errorf("%w: %s", ErrBadInput, err.Error())
		}
		return a.UpdateGameSettings(gameID, viewerUID, body)
	})
}

func (a *App) handleUpdateRoomMods(w http.ResponseWriter, r *http.Request) {
	a.withGameMutation(w, r, func(viewerUID, gameID string) error {
		var body GameRoomModsInput
		if err := decodeBody(r, &body); err != nil {
			return fmt.Errorf("%w: %s", ErrBadInput, err.Error())
		}
		return a.UpdateRoomMods(gameID, viewerUID, body)
	})
}

func (a *App) handleCreateRoomAuth(w http.ResponseWriter, r *http.Request) {
	session, err := a.requireSession(r)
	if err != nil {
		writeError(w, err)
		return
	}
	gameID := strings.ToUpper(strings.TrimSpace(r.PathValue("gameId")))
	viewerUID, err := a.resolveEffectiveUID(r, gameID, session.UID)
	if err != nil {
		writeError(w, err)
		return
	}
	snapshot, err := a.GetSnapshot(gameID, viewerUID)
	if err != nil {
		writeError(w, err)
		return
	}
	if !snapshot.Viewer.IsParticipant {
		writeError(w, fmt.Errorf("%w: only room participants can migrate this identity", ErrForbidden))
		return
	}
	token, err := a.CreateOrGetRoomAuthToken(gameID, viewerUID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"token": token.Token, "gameId": gameID, "success": true})
}

func (a *App) handleStartGame(w http.ResponseWriter, r *http.Request) {
	a.withGameMutation(w, r, func(viewerUID, gameID string) error {
		return a.StartGame(gameID, viewerUID)
	})
}

func (a *App) handleStartRoomTimer(w http.ResponseWriter, r *http.Request) {
	a.withGameMutation(w, r, func(viewerUID, gameID string) error {
		var body struct {
			Seconds *int `json:"seconds"`
		}
		if err := decodeBody(r, &body); err != nil {
			return fmt.Errorf("%w: %s", ErrBadInput, err.Error())
		}
		return a.StartRoomTimer(gameID, viewerUID, body.Seconds)
	})
}

func (a *App) handlePauseRoomTimer(w http.ResponseWriter, r *http.Request) {
	a.withGameMutation(w, r, func(viewerUID, gameID string) error {
		return a.PauseRoomTimer(gameID, viewerUID)
	})
}

func (a *App) handleResumeRoomTimer(w http.ResponseWriter, r *http.Request) {
	a.withGameMutation(w, r, func(viewerUID, gameID string) error {
		return a.ResumeRoomTimer(gameID, viewerUID)
	})
}

func (a *App) handleResetRoomTimer(w http.ResponseWriter, r *http.Request) {
	a.withGameMutation(w, r, func(viewerUID, gameID string) error {
		return a.ResetRoomTimer(gameID, viewerUID)
	})
}

func (a *App) handleClearRoomTimer(w http.ResponseWriter, r *http.Request) {
	a.withGameMutation(w, r, func(viewerUID, gameID string) error {
		return a.ClearRoomTimer(gameID, viewerUID)
	})
}

func (a *App) handleSelectMurdererCards(w http.ResponseWriter, r *http.Request) {
	a.withGameMutation(w, r, func(viewerUID, gameID string) error {
		var body struct {
			ClueCardName  string `json:"clueCardName"`
			MeansCardName string `json:"meansCardName"`
		}
		if err := decodeBody(r, &body); err != nil {
			return fmt.Errorf("%w: %s", ErrBadInput, err.Error())
		}
		return a.SelectMurdererCards(gameID, viewerUID, body.ClueCardName, body.MeansCardName)
	})
}

func (a *App) handleShowWitnessSelectionPrompt(w http.ResponseWriter, r *http.Request) {
	a.withGameMutation(w, r, func(viewerUID, gameID string) error {
		var body struct {
			TargetUID string `json:"targetUid"`
		}
		if err := decodeBody(r, &body); err != nil {
			return fmt.Errorf("%w: %s", ErrBadInput, err.Error())
		}
		return a.ShowWitnessSelectionPrompt(gameID, viewerUID, body.TargetUID)
	})
}

func (a *App) handleSubmitWitnessSelection(w http.ResponseWriter, r *http.Request) {
	a.withGameMutation(w, r, func(viewerUID, gameID string) error {
		var body struct {
			SelectedUIDs []string `json:"selectedUids"`
		}
		if err := decodeBody(r, &body); err != nil {
			return fmt.Errorf("%w: %s", ErrBadInput, err.Error())
		}
		return a.SubmitWitnessSelection(gameID, viewerUID, body.SelectedUIDs)
	})
}

func (a *App) handleDismissWitnessSelectionPrompt(w http.ResponseWriter, r *http.Request) {
	a.withGameMutation(w, r, func(viewerUID, gameID string) error {
		return a.DismissWitnessSelectionPrompt(gameID, viewerUID)
	})
}

func (a *App) handleSelectCauseCard(w http.ResponseWriter, r *http.Request) {
	a.withGameMutation(w, r, func(viewerUID, gameID string) error {
		var body struct {
			Card ForensicCard `json:"card"`
		}
		if err := decodeBody(r, &body); err != nil {
			return fmt.Errorf("%w: %s", ErrBadInput, err.Error())
		}
		return a.SelectForensicCauseCard(gameID, viewerUID, body.Card)
	})
}

func (a *App) handleSelectLocationCard(w http.ResponseWriter, r *http.Request) {
	a.withGameMutation(w, r, func(viewerUID, gameID string) error {
		var body struct {
			Card ForensicCard `json:"card"`
		}
		if err := decodeBody(r, &body); err != nil {
			return fmt.Errorf("%w: %s", ErrBadInput, err.Error())
		}
		return a.SelectForensicLocationCard(gameID, viewerUID, body.Card)
	})
}

func (a *App) handleSelectOtherCard(w http.ResponseWriter, r *http.Request) {
	a.withGameMutation(w, r, func(viewerUID, gameID string) error {
		var body struct {
			Card            ForensicCard `json:"card"`
			ReplaceCardName string       `json:"replaceCardName"`
		}
		if err := decodeBody(r, &body); err != nil {
			return fmt.Errorf("%w: %s", ErrBadInput, err.Error())
		}
		return a.SelectForensicOtherCard(gameID, viewerUID, body.Card, body.ReplaceCardName)
	})
}

func (a *App) handleMakeGuess(w http.ResponseWriter, r *http.Request) {
	a.withGameMutation(w, r, func(viewerUID, gameID string) error {
		var body struct {
			MurdererUID   string `json:"murdererUid"`
			ClueCardName  string `json:"clueCardName"`
			MeansCardName string `json:"meansCardName"`
		}
		if err := decodeBody(r, &body); err != nil {
			return fmt.Errorf("%w: %s", ErrBadInput, err.Error())
		}
		return a.MakeGuess(gameID, viewerUID, body.MurdererUID, body.ClueCardName, body.MeansCardName)
	})
}

func (a *App) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	a.withGameMutation(w, r, func(viewerUID, gameID string) error {
		var body struct {
			Message string `json:"message"`
		}
		if err := decodeBody(r, &body); err != nil {
			return fmt.Errorf("%w: %s", ErrBadInput, err.Error())
		}
		return a.SendChatMessage(gameID, viewerUID, body.Message)
	})
}

func (a *App) handleEndGame(w http.ResponseWriter, r *http.Request) {
	a.withGameMutation(w, r, func(viewerUID, gameID string) error {
		return a.EndGame(gameID, viewerUID)
	})
}

func (a *App) handleRestartGame(w http.ResponseWriter, r *http.Request) {
	a.withGameMutation(w, r, func(viewerUID, gameID string) error {
		return a.RestartGame(gameID, viewerUID)
	})
}

func (a *App) withGameMutation(w http.ResponseWriter, r *http.Request, fn func(string, string) error) {
	session, err := a.requireSession(r)
	if err != nil {
		writeError(w, err)
		return
	}
	gameID := strings.ToUpper(strings.TrimSpace(r.PathValue("gameId")))
	viewerUID, err := a.resolveEffectiveUID(r, gameID, session.UID)
	if err != nil {
		writeError(w, err)
		return
	}
	if err := fn(viewerUID, gameID); err != nil {
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
	token := sessionTokenFromRequest(r)
	if token == "" {
		return nil, fmt.Errorf("%w: missing session", ErrForbidden)
	}
	return a.GetSession(token)
}

func (a *App) resolveEffectiveUID(r *http.Request, gameID, sessionUID string) (string, error) {
	gameID = strings.ToUpper(strings.TrimSpace(gameID))
	if gameID == "" {
		return sessionUID, nil
	}
	roomAuth := roomAuthFromRequest(r)
	if roomAuth == "" {
		return sessionUID, nil
	}
	uid, err := a.ResolveRoomAuthToken(gameID, roomAuth)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return sessionUID, nil
		}
		return "", err
	}
	return uid, nil
}

func sessionTokenFromRequest(r *http.Request) string {
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return strings.TrimSpace(authHeader[7:])
	}
	if headerToken := strings.TrimSpace(r.Header.Get("X-Treachery-Session")); headerToken != "" {
		return headerToken
	}
	if queryToken := strings.TrimSpace(r.URL.Query().Get("sessionToken")); queryToken != "" {
		return queryToken
	}
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		return cookie.Value
	}
	return ""
}

func roomAuthFromRequest(r *http.Request) string {
	if headerToken := strings.TrimSpace(r.Header.Get("X-Treachery-Room-Auth")); headerToken != "" {
		return headerToken
	}
	return strings.TrimSpace(r.URL.Query().Get("roomAuth"))
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
