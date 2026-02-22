package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/secrets"
)

// Result contains the authentication result.
type Result struct {
	Credentials *secrets.StoreCredentials
	Error       error
}

// Server handles browser-based credential entry.
type Server struct {
	csrfToken     string
	result        chan Result
	shutdown      chan struct{}
	pendingResult *Result
	server        *http.Server
	validator     CredentialValidator
	Out           io.Writer

	mu       sync.Mutex
	attempts map[string][]time.Time
}

const (
	maxAttempts     = 10
	rateLimitWindow = 15 * time.Minute
)

// NewServer creates a new auth server.
func NewServer() (*Server, error) {
	return NewServerWithValidator(NewAPICredentialValidator())
}

// NewServerWithValidator creates a new auth server with a custom validator.
func NewServerWithValidator(validator CredentialValidator) (*Server, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return nil, fmt.Errorf("failed to generate CSRF token: %w", err)
	}

	return &Server{
		csrfToken: hex.EncodeToString(token),
		result:    make(chan Result, 1),
		shutdown:  make(chan struct{}),
		attempts:  make(map[string][]time.Time),
		validator: validator,
		Out:       os.Stdout,
	}, nil
}

// Run starts the server and opens the browser.
func (s *Server) Run(ctx context.Context) (*secrets.StoreCredentials, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to start listener: %w", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleSetup)
	mux.HandleFunc("/validate", s.handleValidate)
	mux.HandleFunc("/submit", s.handleSubmit)
	mux.HandleFunc("/success", s.handleSuccess)
	mux.HandleFunc("/complete", s.handleComplete)

	s.server = &http.Server{
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		_ = s.server.Serve(listener)
	}()

	// Print URL first so user can open manually if needed
	_, _ = fmt.Fprintf(s.Out, "Open this URL in your browser to authenticate:\n  %s\n", baseURL)
	_, _ = fmt.Fprintln(s.Out, "Attempting to open browser automatically...")
	if err := openBrowser(baseURL); err != nil {
		_, _ = fmt.Fprintf(s.Out, "Could not open browser automatically: %v\n", err)
		_, _ = fmt.Fprintln(s.Out, "Please open the URL manually in your browser.")
	}

	// Wait for result or context cancellation
	select {
	case result := <-s.result:
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.server.Shutdown(shutdownCtx); err != nil {
			_ = s.server.Close()
		}
		return result.Credentials, result.Error
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.server.Shutdown(shutdownCtx); err != nil {
			_ = s.server.Close()
		}
		return nil, ctx.Err()
	case <-s.shutdown:
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.server.Shutdown(shutdownCtx); err != nil {
			_ = s.server.Close()
		}
		if s.pendingResult != nil {
			return s.pendingResult.Credentials, s.pendingResult.Error
		}
		return nil, fmt.Errorf("setup cancelled")
	}
}

// handleSetup serves the main setup page
func (s *Server) handleSetup(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.New("setup").Parse(setupTemplate)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	data := map[string]string{
		"CSRFToken": s.csrfToken,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.Execute(w, data)
}

// handleValidate tests credentials without saving
func (s *Server) handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Verify CSRF token
	if r.Header.Get("X-CSRF-Token") != s.csrfToken {
		http.Error(w, "Invalid CSRF token", http.StatusForbidden)
		return
	}

	clientIP := r.RemoteAddr
	if !s.checkRateLimit(clientIP) {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"error":   "Too many attempts. Please wait 15 minutes.",
		})
		return
	}

	var req struct {
		Handle      string `json:"handle"`
		AccessToken string `json:"access_token"`
		AppKey      string `json:"app_key"`
		AppSecret   string `json:"app_secret"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	// Normalize handle
	req.Handle = strings.TrimSpace(req.Handle)
	req.AccessToken = strings.TrimSpace(req.AccessToken)

	if req.Handle == "" || req.AccessToken == "" {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"error":   "Store handle and access token are required",
		})
		return
	}

	// Validate credentials
	if err := s.validator.Validate(r.Context(), req.Handle, req.AccessToken); err != nil {
		errorMessage := "Could not validate credentials. Please check your network connection and try again."
		if _, ok := err.(*CredentialValidationError); ok {
			errorMessage = err.Error()
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"error":   errorMessage,
		})
		return
	}

	// Fetch order count for display
	client := api.NewOpenAPIClient(req.AccessToken)
	orderCount := 0
	if resp, err := client.ListOrders(r.Context(), &api.OrdersListOptions{Page: 1, PageSize: 1}); err == nil {
		orderCount = resp.TotalCount
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success":     true,
		"message":     "Connection successful!",
		"store_name":  req.Handle,
		"order_count": orderCount,
	})
}

// handleSubmit saves credentials after validation
func (s *Server) handleSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Verify CSRF token
	if r.Header.Get("X-CSRF-Token") != s.csrfToken {
		http.Error(w, "Invalid CSRF token", http.StatusForbidden)
		return
	}

	clientIP := r.RemoteAddr
	if !s.checkRateLimit(clientIP) {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"error":   "Too many attempts. Please wait 15 minutes.",
		})
		return
	}

	var req struct {
		Handle      string `json:"handle"`
		AccessToken string `json:"access_token"`
		AppKey      string `json:"app_key"`
		AppSecret   string `json:"app_secret"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	// Normalize inputs
	req.Handle = strings.TrimSpace(req.Handle)
	req.AccessToken = strings.TrimSpace(req.AccessToken)
	req.AppKey = strings.TrimSpace(req.AppKey)
	req.AppSecret = strings.TrimSpace(req.AppSecret)

	if req.Handle == "" || req.AccessToken == "" {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"error":   "Store handle and access token are required",
		})
		return
	}

	// Validate first
	if err := s.validator.Validate(r.Context(), req.Handle, req.AccessToken); err != nil {
		errorMessage := "Could not validate credentials. Please check your network connection and try again."
		if _, ok := err.(*CredentialValidationError); ok {
			errorMessage = err.Error()
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"error":   errorMessage,
		})
		return
	}

	creds := &secrets.StoreCredentials{
		Name:        req.Handle, // Use handle as the profile name
		Handle:      req.Handle,
		AccessToken: req.AccessToken,
		AppKey:      req.AppKey,
		AppSecret:   req.AppSecret,
		Region:      "default",
		CreatedAt:   time.Now(),
	}

	// Store pending result
	s.pendingResult = &Result{Credentials: creds}

	writeJSON(w, http.StatusOK, map[string]any{
		"success":    true,
		"store_name": req.Handle + ".myshopline.com",
	})
}

// handleSuccess serves the success page
func (s *Server) handleSuccess(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("success").Parse(successTemplate)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	data := map[string]string{
		"StoreName": r.URL.Query().Get("store"),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.Execute(w, data)
}

// handleComplete signals that setup is done
func (s *Server) handleComplete(w http.ResponseWriter, r *http.Request) {
	if s.pendingResult != nil {
		s.result <- *s.pendingResult
	}
	close(s.shutdown)
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

func (s *Server) checkRateLimit(clientIP string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rateLimitWindow)

	var valid []time.Time
	for _, t := range s.attempts[clientIP] {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= maxAttempts {
		return false
	}

	s.attempts[clientIP] = append(valid, now)
	return true
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// openBrowserFunc is the browser launcher. Tests can override this to prevent real browser opens.
var openBrowserFunc = openBrowserDefault

func openBrowser(url string) error {
	return openBrowserFunc(url)
}

func openBrowserDefault(url string) error {
	// Explicit opt-outs for local automation/CI.
	for _, env := range []string{"SHOPLINE_NO_BROWSER", "NO_BROWSER"} {
		v := strings.TrimSpace(strings.ToLower(os.Getenv(env)))
		if v == "1" || v == "true" || v == "yes" {
			return nil
		}
	}

	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	default:
		return fmt.Errorf("unsupported platform")
	}

	return exec.Command(cmd, args...).Start()
}
