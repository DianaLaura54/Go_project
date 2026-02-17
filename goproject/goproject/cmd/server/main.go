package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"goproject/internal/auth"
	"goproject/internal/notes"
)

var (
	users  = auth.NewUserStore()
	store  = notes.NewStore()
	tokens = auth.NewTokenManager("super-secret-change-me")
)

// ─── helpers ──────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func readJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func errJSON(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func getUser(r *http.Request) (*auth.Claims, bool) {
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return nil, false
	}
	claims, err := tokens.ValidateToken(strings.TrimPrefix(header, "Bearer "))
	if err != nil {
		return nil, false
	}
	return claims, true
}

func withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := getUser(r)
		if !ok {
			errJSON(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		ctx := context.WithValue(r.Context(), claimsKey, claims)
		next(w, r.WithContext(ctx))
	}
}

// ─── context helpers ──────────────────────────────────────────────────────────

type contextKey string

const claimsKey contextKey = "claims"

// ─── handlers ─────────────────────────────────────────────────────────────────

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := readJSON(r, &body); err != nil {
		errJSON(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if body.Username == "" || body.Password == "" {
		errJSON(w, http.StatusBadRequest, "username and password required")
		return
	}

	user, err := users.Register(body.Username, body.Password)
	if err == auth.ErrUserExists {
		errJSON(w, http.StatusConflict, "username already taken")
		return
	}
	if err != nil {
		errJSON(w, http.StatusInternalServerError, "registration failed")
		return
	}

	token, _ := tokens.CreateToken(user, 24*time.Hour)
	writeJSON(w, http.StatusCreated, map[string]any{
		"message": "registered successfully",
		"token":   token,
		"user":    map[string]string{"id": user.ID, "username": user.Username},
	})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := readJSON(r, &body); err != nil {
		errJSON(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	user, err := users.Login(body.Username, body.Password)
	if err != nil {
		errJSON(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, _ := tokens.CreateToken(user, 24*time.Hour)
	writeJSON(w, http.StatusOK, map[string]any{
		"token": token,
		"user":  map[string]string{"id": user.ID, "username": user.Username},
	})
}

func handleNotes(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(claimsKey).(*auth.Claims)

	switch r.Method {
	case http.MethodGet:
		noteList := store.List(claims.UserID)
		if noteList == nil {
			noteList = []*notes.Note{}
		}
		writeJSON(w, http.StatusOK, map[string]any{"notes": noteList, "count": len(noteList)})

	case http.MethodPost:
		var input notes.CreateInput
		if err := readJSON(r, &input); err != nil {
			errJSON(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if input.Title == "" {
			errJSON(w, http.StatusBadRequest, "title is required")
			return
		}
		note := store.Create(claims.UserID, input)
		writeJSON(w, http.StatusCreated, note)

	default:
		errJSON(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func handleNote(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(claimsKey).(*auth.Claims)
	noteID := strings.TrimPrefix(r.URL.Path, "/notes/")
	if noteID == "" {
		errJSON(w, http.StatusBadRequest, "note id required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		note, err := store.Get(claims.UserID, noteID)
		if err == notes.ErrNotFound {
			errJSON(w, http.StatusNotFound, "note not found")
			return
		}
		if err == notes.ErrForbidden {
			errJSON(w, http.StatusForbidden, "access denied")
			return
		}
		writeJSON(w, http.StatusOK, note)

	case http.MethodPut:
		var input notes.UpdateInput
		if err := readJSON(r, &input); err != nil {
			errJSON(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		note, err := store.Update(claims.UserID, noteID, input)
		if err == notes.ErrNotFound {
			errJSON(w, http.StatusNotFound, "note not found")
			return
		}
		if err == notes.ErrForbidden {
			errJSON(w, http.StatusForbidden, "access denied")
			return
		}
		writeJSON(w, http.StatusOK, note)

	case http.MethodDelete:
		err := store.Delete(claims.UserID, noteID)
		if err == notes.ErrNotFound {
			errJSON(w, http.StatusNotFound, "note not found")
			return
		}
		if err == notes.ErrForbidden {
			errJSON(w, http.StatusForbidden, "access denied")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})

	default:
		errJSON(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "time": time.Now().Format(time.RFC3339)})
}

// ─── main ─────────────────────────────────────────────────────────────────────

func main() {
	port := flag.String("port", "8080", "Port to listen on")
	flag.Parse()

	mux := http.NewServeMux()

	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/auth/register", handleRegister)
	mux.HandleFunc("/auth/login", handleLogin)
	mux.HandleFunc("/notes", withAuth(handleNotes))
	mux.HandleFunc("/notes/", withAuth(handleNote))

	addr := fmt.Sprintf(":%s", *port)
	fmt.Printf("🚀 Notes API running on http://localhost%s\n", addr)
	fmt.Println()
	fmt.Println("Endpoints:")
	fmt.Println("  POST   /auth/register  — create account")
	fmt.Println("  POST   /auth/login     — get token")
	fmt.Println("  GET    /notes          — list notes")
	fmt.Println("  POST   /notes          — create note")
	fmt.Println("  GET    /notes/:id      — get note")
	fmt.Println("  PUT    /notes/:id      — update note")
	fmt.Println("  DELETE /notes/:id      — delete note")
	fmt.Println("  GET    /health         — health check")
	fmt.Println()
	fmt.Println("Auth: Bearer token in Authorization header")

	log.Fatal(http.ListenAndServe(addr, mux))
}
