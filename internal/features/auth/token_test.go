package auth

import (
	"errors"
	"testing"
	"time"
)

func TestTokenManager_RoundTrip(t *testing.T) {
	m := NewTokenManager(testSecret, time.Hour, "test")
	token, _, err := m.Generate(42)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	id, err := m.Parse(token)
	if err != nil || id != 42 {
		t.Fatalf("Parse = %d, %v; want 42, nil", id, err)
	}
}

func TestTokenManager_Rejects(t *testing.T) {
	issued := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	m := &jwtManager{secret: []byte(testSecret), ttl: time.Hour, issuer: "test", now: func() time.Time { return issued }}
	token, _, err := m.Generate(42)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	tests := []struct {
		name  string
		m     TokenManager
		token string
	}{
		{"expired", &jwtManager{secret: []byte(testSecret), issuer: "test", now: func() time.Time { return issued.Add(2 * time.Hour) }}, token},
		{"wrong secret", &jwtManager{secret: []byte("another-secret-of-at-least-32-chars"), issuer: "test", now: m.now}, token},
		{"wrong issuer", &jwtManager{secret: []byte(testSecret), issuer: "other", now: m.now}, token},
		{"garbage", m, "not-a-token"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.m.Parse(tt.token); !errors.Is(err, ErrInvalidToken) {
				t.Fatalf("err = %v, want ErrInvalidToken", err)
			}
		})
	}
}
