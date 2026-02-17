package auth_test

import (
	"testing"
	"time"

	"goproject/internal/auth"
)

func TestRegisterAndLogin(t *testing.T) {
	store := auth.NewUserStore()

	user, err := store.Register("alice", "password123")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if user.Username != "alice" {
		t.Errorf("expected username alice, got %s", user.Username)
	}

	// duplicate
	_, err = store.Register("alice", "other")
	if err != auth.ErrUserExists {
		t.Errorf("expected ErrUserExists, got %v", err)
	}

	// login success
	loggedIn, err := store.Login("alice", "password123")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if loggedIn.ID != user.ID {
		t.Error("user IDs don't match")
	}

	// wrong password
	_, err = store.Login("alice", "wrong")
	if err != auth.ErrWrongPassword {
		t.Errorf("expected ErrWrongPassword, got %v", err)
	}

	// unknown user
	_, err = store.Login("bob", "pass")
	if err != auth.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestTokenCreateAndValidate(t *testing.T) {
	store := auth.NewUserStore()
	tm := auth.NewTokenManager("test-secret")

	user, _ := store.Register("bob", "pass")

	token, err := tm.CreateToken(user, time.Hour)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	claims, err := tm.ValidateToken(token)
	if err != nil {
		t.Fatalf("validate token: %v", err)
	}
	if claims.Username != "bob" {
		t.Errorf("expected bob, got %s", claims.Username)
	}

	// tampered token
	_, err = tm.ValidateToken(token + "x")
	if err != auth.ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken for tampered token, got %v", err)
	}

	// expired token
	expiredToken, _ := tm.CreateToken(user, -time.Second)
	_, err = tm.ValidateToken(expiredToken)
	if err != auth.ErrExpiredToken {
		t.Errorf("expected ErrExpiredToken, got %v", err)
	}
}
