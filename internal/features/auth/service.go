package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Register(ctx context.Context, req RegisterRequest) (*TokenResponse, error)
	Login(ctx context.Context, req LoginRequest) (*TokenResponse, error)
}

type service struct {
	repo   Repository
	tokens TokenManager
	// dummyHash dibandingkan saat email tidak ditemukan supaya waktu respons
	// login tidak membocorkan email mana yang terdaftar.
	dummyHash []byte
}

func NewService(repo Repository, tokens TokenManager) Service {
	dummyHash, _ := bcrypt.GenerateFromPassword([]byte("dummy-password-for-timing"), bcrypt.DefaultCost)
	return &service{repo: repo, tokens: tokens, dummyHash: dummyHash}
}

func (s *service) Register(ctx context.Context, req RegisterRequest) (*TokenResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	id, err := s.repo.CreateUser(ctx, strings.TrimSpace(req.Name), normalizeEmail(req.Email), string(hash))
	if err != nil {
		return nil, fmt.Errorf("register user: %w", err)
	}
	return s.issueToken(id)
}

func (s *service) Login(ctx context.Context, req LoginRequest) (*TokenResponse, error) {
	creds, err := s.repo.FindCredentialsByEmail(ctx, normalizeEmail(req.Email))
	if errors.Is(err, errCredentialsNotFound) {
		_ = bcrypt.CompareHashAndPassword(s.dummyHash, []byte(req.Password))
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("login: %w", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(creds.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	return s.issueToken(creds.ID)
}

func (s *service) issueToken(userID int64) (*TokenResponse, error) {
	token, expiresAt, err := s.tokens.Generate(userID)
	if err != nil {
		return nil, fmt.Errorf("issue token for user %d: %w", userID, err)
	}
	return &TokenResponse{AccessToken: token, TokenType: "Bearer", ExpiresAt: expiresAt}, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
