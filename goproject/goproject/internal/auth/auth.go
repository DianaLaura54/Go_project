package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidToken  = errors.New("invalid token")
	ErrExpiredToken  = errors.New("token expired")
	ErrUserNotFound  = errors.New("user not found")
	ErrWrongPassword = errors.New("wrong password")
	ErrUserExists    = errors.New("user already exists")
)

// Simple in-memory user store (replace with DB in production)
type UserStore struct {
	mu    sync.RWMutex
	users map[string]*User
}

type User struct {
	ID           string
	Username     string
	PasswordHash string // sha256 hex of salt+password
	Salt         string
	CreatedAt    time.Time
}

func NewUserStore() *UserStore {
	return &UserStore{users: make(map[string]*User)}
}

func generateSalt() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func hashPassword(password, salt string) string {
	h := sha256.New()
	h.Write([]byte(salt + password))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func (s *UserStore) Register(username, password string) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[username]; exists {
		return nil, ErrUserExists
	}

	salt, err := generateSalt()
	if err != nil {
		return nil, err
	}

	user := &User{
		ID:           generateID(),
		Username:     username,
		PasswordHash: hashPassword(password, salt),
		Salt:         salt,
		CreatedAt:    time.Now(),
	}
	s.users[username] = user
	return user, nil
}

func (s *UserStore) Login(username, password string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[username]
	if !exists {
		return nil, ErrUserNotFound
	}

	if hashPassword(password, user.Salt) != user.PasswordHash {
		return nil, ErrWrongPassword
	}
	return user, nil
}

// Minimal JWT-like token using HMAC-SHA256
type TokenManager struct {
	secret []byte
}

func NewTokenManager(secret string) *TokenManager {
	return &TokenManager{secret: []byte(secret)}
}

type Claims struct {
	UserID   string    `json:"uid"`
	Username string    `json:"usr"`
	Expires  time.Time `json:"exp"`
}

func (tm *TokenManager) CreateToken(user *User, ttl time.Duration) (string, error) {
	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		Expires:  time.Now().Add(ttl),
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	encoded := base64.RawURLEncoding.EncodeToString(payload)
	sig := tm.sign(encoded)
	return encoded + "." + sig, nil
}

func (tm *TokenManager) ValidateToken(token string) (*Claims, error) {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return nil, ErrInvalidToken
	}

	if tm.sign(parts[0]) != parts[1] {
		return nil, ErrInvalidToken
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, ErrInvalidToken
	}

	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	if time.Now().After(claims.Expires) {
		return nil, ErrExpiredToken
	}

	return &claims, nil
}

func (tm *TokenManager) sign(data string) string {
	mac := hmac.New(sha256.New, tm.secret)
	mac.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
