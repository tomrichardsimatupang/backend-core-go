package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenManager interface {
	Generate(userID int64) (token string, expiresAt time.Time, err error)
	Parse(token string) (userID int64, err error)
}

// Claims hanya berisi user_id dan klaim standar. Role/permission TIDAK
// disimpan di token; otorisasi selalu dicek lewat rbac.
type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

type jwtManager struct {
	secret []byte
	ttl    time.Duration
	issuer string
	now    func() time.Time
}

func NewTokenManager(secret string, ttl time.Duration, issuer string) TokenManager {
	return &jwtManager{secret: []byte(secret), ttl: ttl, issuer: issuer, now: time.Now}
}

func (m *jwtManager) Generate(userID int64) (string, time.Time, error) {
	now := m.now()
	expiresAt := now.Add(m.ttl)
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}
	return token, expiresAt, nil
}

func (m *jwtManager) Parse(token string) (int64, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(token, claims,
		func(*jwt.Token) (any, error) { return m.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(m.issuer),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(m.now),
	)
	if err != nil || claims.UserID <= 0 {
		return 0, ErrInvalidToken
	}
	return claims.UserID, nil
}
