package services

import (
	"aswadwk/messaging-task-go/internal/config"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type JwtService interface {
	GenerateToken(tenantID string) (string, error)
	ValidateToken(token string) (*jwt.Token, error)
}

type jwtService struct {
}

// GenerateToken implements JwtService.
func (j *jwtService) GenerateToken(tenantID string) (string, error) {
	token, err := generateToken(tenantID)

	if err != nil {
		return "", err
	}

	return token, nil
}

// ValidateToken implements JwtService.
func (j *jwtService) ValidateToken(token string) (*jwt.Token, error) {
	decode, err := parseToken(token)

	if err != nil {
		return nil, err
	}

	return decode, nil
}

func NewJwtService() JwtService {
	return &jwtService{}
}

func generateToken(tenantID string) (string, error) {
	ttl := config.Cfg.JWTAccessTokenTTL

	if ttl == "" {
		panic("JWT_ACCESS_TOKEN_TTL is not set")
	}

	secretKey := config.Cfg.JWTSecret

	ttlDuration, err := time.ParseDuration(ttl)

	if err != nil {
		return "", err
	}

	claims := jwt.MapClaims{
		"iss":  "api-hub",
		"sub":  tenantID,
		"aud":  "api-hub",
		"exp":  time.Now().Add(ttlDuration).Unix(),
		"nbf":  time.Now().Unix(),
		"iat":  time.Now().Unix(),
		"role": "user",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(secretKey))

	if err != nil {
		return signedToken, err
	}

	return signedToken, nil
}

func parseToken(token string) (*jwt.Token, error) {
	secretKey := config.Cfg.JWTSecret

	tokenParse, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Token invalid or expired")
	}

	return tokenParse, nil
}
