// Package service contains the business logic for the Gophermart loyalty system.
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"gophermart/internal/domain"
	"gophermart/internal/repository"
)

const (
	// TokenTTL is the lifetime of issued JWT tokens.
	TokenTTL = 24 * time.Hour

	bcryptCost = bcrypt.DefaultCost
)

// AuthService handles user registration, login, and JWT token management.
type AuthService struct {
	userRepo  repository.UserRepo
	jwtSecret []byte
}

// NewAuthService creates a new AuthService with the given repository and JWT secret.
func NewAuthService(userRepo repository.UserRepo, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: []byte(jwtSecret),
	}
}

// Register creates a new user and returns a signed JWT token.
// Returns domain.ErrUserAlreadyExists if the login is taken.
func (s *AuthService) Register(ctx context.Context, login, password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	user, err := s.userRepo.CreateUser(ctx, login, string(hash))
	if err != nil {
		return "", err
	}

	return s.issueToken(user.ID)
}

// Login verifies credentials and returns a signed JWT token.
// Returns domain.ErrInvalidCredentials on wrong login or password.
func (s *AuthService) Login(ctx context.Context, login, password string) (string, error) {
	user, err := s.userRepo.GetUserByLogin(ctx, login)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return "", domain.ErrInvalidCredentials
		}
		return "", fmt.Errorf("compare password: %w", err)
	}

	return s.issueToken(user.ID)
}

// ValidateToken parses and validates a JWT token string, returning the user ID as a string.
func (s *AuthService) ValidateToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	sub, err := claims.GetSubject()
	if err != nil {
		return "", fmt.Errorf("get subject: %w", err)
	}
	if sub == "" {
		return "", errors.New("empty subject in token")
	}
	return sub, nil
}

// issueToken creates a signed JWT for the given user ID.
func (s *AuthService) issueToken(userID int64) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   fmt.Sprintf("%d", userID),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenTTL)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
