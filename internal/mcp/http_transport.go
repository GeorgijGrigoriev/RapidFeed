package mcp

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/localrivet/gomcp/transport"
)

const (
	defaultMCPEndpoint     = "/mcp"
	defaultShutdownTimeout = 10 * time.Second
)

type httpTransport struct {
	transport.BaseTransport
	addr           string
	server         *http.Server
	pathPrefix     string
	mcpEndpoint    string
	enableSessions bool
	sessions       map[string]*sessionInfo
	sessionsMu     sync.Mutex
}

type sessionInfo struct {
	ID        string
	CreatedAt time.Time
	LastSeen  time.Time
	ClientID  string
	Token     string
}

func newHTTPTransport(addr string) *httpTransport {
	return &httpTransport{
		addr:           addr,
		pathPrefix:     "",
		mcpEndpoint:    defaultMCPEndpoint,
		enableSessions: true,
		sessions:       make(map[string]*sessionInfo),
	}
}

func (t *httpTransport) Initialize() error {
	return nil
}

func (t *httpTransport) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc(t.getFullMCPEndpoint(), t.handleMCPRequest)

	t.server = &http.Server{
		Addr:    t.addr,
		Handler: mux,
	}

	go func() {
		if err := t.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.GetLogger().Error("MCP HTTP server error", "error", err)
		}
	}()

	return nil
}

func (t *httpTransport) Stop() error {
	if t.server == nil {
		return nil
	}

	ctx, cancel := contextWithTimeout(defaultShutdownTimeout)
	defer cancel()

	return t.server.Shutdown(ctx)
}

func (t *httpTransport) Send(message []byte) error {
	// Server-initiated messages would be delivered via SSE streams.
	// For now, keep parity with gomcp's HTTP transport behavior.
	t.sessionsMu.Lock()
	defer t.sessionsMu.Unlock()

	for sessionID, session := range t.sessions {
		if time.Since(session.LastSeen) > 5*time.Minute {
			delete(t.sessions, sessionID)
			continue
		}

		t.GetLogger().Info("MCP SSE message queued", "sessionID", sessionID, "size", len(message))
	}

	return nil
}

func (t *httpTransport) Receive() ([]byte, error) {
	return nil, errors.New("receive not supported for HTTP transport")
}

func (t *httpTransport) handleMCPRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		t.handleClientMessage(w, r)
	case http.MethodGet:
		t.handleSSEStream(w, r)
	case http.MethodDelete:
		t.handleSessionTermination(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (t *httpTransport) handleClientMessage(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(strings.ToLower(contentType), "application/json") {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var session *sessionInfo
	sessionID := r.Header.Get("MCP-Session-ID")
	if t.enableSessions {
		if sessionID == "" {
			sessionID = t.generateSessionID()
			w.Header().Set("MCP-Session-ID", sessionID)
			t.sessionsMu.Lock()
			session = &sessionInfo{
				ID:        sessionID,
				CreatedAt: time.Now(),
				LastSeen:  time.Now(),
				ClientID:  r.RemoteAddr,
			}
			t.sessions[sessionID] = session
			t.sessionsMu.Unlock()
		} else {
			t.sessionsMu.Lock()
			if existing, exists := t.sessions[sessionID]; exists {
				session = existing
				session.LastSeen = time.Now()
				w.Header().Set("MCP-Session-ID", sessionID)
			} else {
				sessionID = t.generateSessionID()
				w.Header().Set("MCP-Session-ID", sessionID)
				session = &sessionInfo{
					ID:        sessionID,
					CreatedAt: time.Now(),
					LastSeen:  time.Now(),
					ClientID:  r.RemoteAddr,
				}
				t.sessions[sessionID] = session
			}
			t.sessionsMu.Unlock()
		}
	}

	token := tokenFromHeaders(r)
	if token == "" && session != nil {
		token = session.Token
	}
	if token != "" {
		if session != nil {
			session.Token = token
		}
		if patched, err := injectToken(body, token); err == nil {
			body = patched
		}
	}

	response, err := t.HandleMessage(body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Message handling failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(response); err != nil {
		t.GetLogger().Error("Failed to write response", "error", err)
	}
}

func (t *httpTransport) handleSSEStream(w http.ResponseWriter, r *http.Request) {
	accept := r.Header.Get("Accept")
	if !strings.Contains(accept, "text/event-stream") {
		http.Error(w, "text/event-stream not accepted", http.StatusNotAcceptable)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	sessionID := r.Header.Get("MCP-Session-ID")
	if t.enableSessions && sessionID != "" {
		t.sessionsMu.Lock()
		if session, exists := t.sessions[sessionID]; exists {
			session.LastSeen = time.Now()
			w.Header().Set("MCP-Session-ID", sessionID)
		} else {
			http.Error(w, "Session not found", http.StatusNotFound)
			t.sessionsMu.Unlock()
			return
		}
		t.sessionsMu.Unlock()
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(1 * time.Second)
			fmt.Fprintf(w, "data: {\"type\":\"heartbeat\",\"timestamp\":\"%s\"}\n\n", time.Now().Format(time.RFC3339))
			flusher.Flush()
		}
	}
}

func (t *httpTransport) handleSessionTermination(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("MCP-Session-ID")
	if sessionID == "" {
		http.Error(w, "Missing MCP-Session-ID header", http.StatusBadRequest)
		return
	}

	t.sessionsMu.Lock()
	delete(t.sessions, sessionID)
	t.sessionsMu.Unlock()

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"session_terminated"}`))
}

func (t *httpTransport) generateSessionID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (t *httpTransport) getFullMCPEndpoint() string {
	if t.pathPrefix == "" {
		return t.mcpEndpoint
	}
	return t.pathPrefix + t.mcpEndpoint
}

func tokenFromHeaders(r *http.Request) string {
	if token := strings.TrimSpace(r.Header.Get("X-MCP-Token")); token != "" {
		return token
	}

	if auth := strings.TrimSpace(r.Header.Get("Authorization")); auth != "" {
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			if token := strings.TrimSpace(parts[1]); token != "" {
				return token
			}
		}
	}

	return ""
}

func injectToken(message []byte, token string) ([]byte, error) {
	trimmed := bytes.TrimSpace(message)
	if len(trimmed) == 0 {
		return message, nil
	}

	if trimmed[0] == '[' {
		var batch []map[string]interface{}
		if err := json.Unmarshal(message, &batch); err != nil {
			return message, err
		}
		for i := range batch {
			batch[i] = injectTokenToObject(batch[i], token)
		}
		return json.Marshal(batch)
	}

	var obj map[string]interface{}
	if err := json.Unmarshal(message, &obj); err != nil {
		return message, err
	}
	obj = injectTokenToObject(obj, token)
	return json.Marshal(obj)
}

func injectTokenToObject(obj map[string]interface{}, token string) map[string]interface{} {
	method, _ := obj["method"].(string)
	if method != "tools/call" {
		return obj
	}

	params, ok := obj["params"].(map[string]interface{})
	if !ok {
		params = make(map[string]interface{})
		obj["params"] = params
	}

	args, ok := params["arguments"].(map[string]interface{})
	if !ok {
		args = make(map[string]interface{})
		params["arguments"] = args
	}

	if existing, ok := args["token"]; ok {
		switch v := existing.(type) {
		case string:
			if strings.TrimSpace(v) != "" {
				return obj
			}
		default:
			if v != nil {
				return obj
			}
		}
	}

	args["token"] = token
	return obj
}

func contextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}
