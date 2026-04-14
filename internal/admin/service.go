package admin

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const adminTokenTTL = 24 * time.Hour

// ErrInvalidAdminCredentials is returned when the admin login or password is incorrect.
var ErrInvalidAdminCredentials = errors.New("invalid admin credentials")

// ErrUnauthorized is returned when an admin token is missing or invalid.
var ErrUnauthorized = errors.New("unauthorized")

// Service handles admin authentication and JWT token management.
type Service struct {
	login     string
	password  string
	jwtSecret []byte
}

// NewService creates a new admin Service with the given credentials and JWT secret.
func NewService(login, password, jwtSecret string) *Service {
	return &Service{
		login:     login,
		password:  password,
		jwtSecret: []byte(jwtSecret),
	}
}

// Login verifies admin credentials and returns a signed JWT token with role="admin".
// Returns ErrInvalidAdminCredentials on wrong credentials.
func (s *Service) Login(login, password string) (string, error) {
	if login != s.login || password != s.password {
		return "", ErrInvalidAdminCredentials
	}
	return s.issueAdminToken()
}

// ValidateAdminToken parses and validates a JWT, confirming it carries role="admin".
// Returns ErrUnauthorized if the token is invalid or belongs to a regular user.
func (s *Service) ValidateAdminToken(tokenStr string) error {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return ErrUnauthorized
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return ErrUnauthorized
	}
	role, _ := claims["role"].(string)
	if role != "admin" {
		return ErrUnauthorized
	}
	return nil
}

// issueAdminToken creates a signed JWT with role="admin" claim.
func (s *Service) issueAdminToken() (string, error) {
	claims := jwt.MapClaims{
		"sub":  "admin",
		"role": "admin",
		"exp":  time.Now().Add(adminTokenTTL).Unix(),
		"iat":  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
