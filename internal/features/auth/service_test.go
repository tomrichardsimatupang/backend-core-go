package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type mockRepo struct {
	users  map[string]Credentials
	nextID int64
}

func newMockRepo() *mockRepo {
	return &mockRepo{users: map[string]Credentials{}, nextID: 1}
}

func (m *mockRepo) FindCredentialsByEmail(_ context.Context, email string) (*Credentials, error) {
	c, ok := m.users[email]
	if !ok {
		return nil, errCredentialsNotFound
	}
	return &c, nil
}

func (m *mockRepo) CreateUser(_ context.Context, _, email, hash string) (int64, error) {
	if _, ok := m.users[email]; ok {
		return 0, ErrEmailTaken
	}
	id := m.nextID
	m.nextID++
	m.users[email] = Credentials{ID: id, PasswordHash: hash}
	return id, nil
}

const testSecret = "test-secret-that-is-at-least-32-chars"

func newTestService() (Service, *mockRepo, TokenManager) {
	repo := newMockRepo()
	tokens := NewTokenManager(testSecret, time.Hour, "test")
	return NewService(repo, tokens), repo, tokens
}

func TestRegister_NormalizesEmailAndHashesPassword(t *testing.T) {
	svc, repo, tokens := newTestService()

	res, err := svc.Register(context.Background(), RegisterRequest{Name: " Budi ", Email: " Budi@Example.COM ", Password: "secret123"})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	creds, ok := repo.users["budi@example.com"]
	if !ok {
		t.Fatal("email should be stored lowercased and trimmed")
	}
	if bcrypt.CompareHashAndPassword([]byte(creds.PasswordHash), []byte("secret123")) != nil {
		t.Fatal("password should be stored as bcrypt hash")
	}
	userID, err := tokens.Parse(res.AccessToken)
	if err != nil || userID != creds.ID {
		t.Fatalf("token user id = %d, err = %v; want %d", userID, err, creds.ID)
	}
}

func TestRegister_EmailTaken(t *testing.T) {
	svc, _, _ := newTestService()
	req := RegisterRequest{Name: "Budi", Email: "budi@example.com", Password: "secret123"}
	if _, err := svc.Register(context.Background(), req); err != nil {
		t.Fatalf("first Register: %v", err)
	}
	if _, err := svc.Register(context.Background(), req); !errors.Is(err, ErrEmailTaken) {
		t.Fatalf("err = %v, want ErrEmailTaken", err)
	}
}

func TestLogin(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()
	if _, err := svc.Register(ctx, RegisterRequest{Name: "Budi", Email: "budi@example.com", Password: "secret123"}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	tests := []struct {
		name    string
		req     LoginRequest
		wantErr error
	}{
		{"success", LoginRequest{Email: "BUDI@example.com", Password: "secret123"}, nil},
		{"wrong password", LoginRequest{Email: "budi@example.com", Password: "wrong"}, ErrInvalidCredentials},
		{"unknown email", LoginRequest{Email: "nobody@example.com", Password: "secret123"}, ErrInvalidCredentials},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := svc.Login(ctx, tt.req)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && res.AccessToken == "" {
				t.Fatal("expected access token")
			}
		})
	}
}
